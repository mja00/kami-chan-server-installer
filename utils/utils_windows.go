package utils

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func PrintOSWarnings() {
	// Just let them know macOS isn't the _best_ OS for running a Minecraft server
	log.Println("Warning: Windows is not the best OS for running a Minecraft server. You may experience issues with performance or stability.")
}

func GetArch() string {
	// Literally don't support anything other than x64 for now
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	default:
		// Fatally die
		panic("Unsupported architecture")
	}
}

func DownloadJava(version int, _ *cli.Context) (string, error) {
	arch := GetArch()
	javaURL := fmt.Sprintf("https://corretto.aws/downloads/latest/amazon-corretto-%d-x64-windows-jdk.msi", version)

	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		err := os.MkdirAll("temp", 0755)
		if err != nil {
			return "", err
		}
	}
	out, err := os.Create(filepath.Join("temp", fmt.Sprintf("java-%d-%s.msi", version, arch)))
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(javaURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading Java",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return "", err
	}

	return filepath.Join("temp", fmt.Sprintf("java-%d-%s.msi", version, arch)), nil
}

func InstallJava(javaPath string, cliCtx *cli.Context) error {
	debug := cliCtx.Bool("debug")
	// If we're in debug, don't actually install Java, just print what we'd do
	// Install: msiexec /i ./temp/java-21-x64.msi /quit /qn /norestart /log ./temp/java-install.log
	cmd := exec.Command("msiexec", "/i", javaPath, "/quiet", "/qn", "/norestart", "/log", filepath.Join("temp", "java-install.log"))
	if debug {
		// Just print the command we'd run
		log.Println(cmd.String())
		return nil
	}
	return RunCommandAndPipeOutput(cmd)
}

func WriteStartScript(path string, ramAmount int, cliCtx *cli.Context) error {
	startScript := fmt.Sprintf(`@echo off

java -Xms%s6M -Xmx%sM -XX:+AlwaysPreTouch -XX:+DisableExplicitGC -XX:+ParallelRefProcEnabled -XX:+PerfDisableSharedMem -XX:+UnlockExperimentalVMOptions -XX:+UseG1GC -XX:G1HeapRegionSize=8M -XX:G1HeapWastePercent=5 -XX:G1MaxNewSizePercent=40 -XX:G1MixedGCCountTarget=4 -XX:G1MixedGCLiveThresholdPercent=90 -XX:G1NewSizePercent=30 -XX:G1RSetUpdatingPauseTimePercent=5 -XX:G1ReservePercent=20 -XX:InitiatingHeapOccupancyPercent=15 -XX:MaxGCPauseMillis=200 -XX:MaxTenuringThreshold=1 -XX:SurvivorRatio=32 -Dusing.aikars.flags=https://mcflags.emc.gs -Daikars.new.flags=true -jar server.jar nogui

pause`, ramAmount, ramAmount)
	err := os.WriteFile(fmt.Sprintf("%s.bat", path), []byte(startScript), 0755)
	if err != nil {
		return err
	}
	return nil
}
