package utils

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

func GetArch() string {
	// Check GOARCH
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	default:
		return "x64"
	}
}

func DownloadJava(version int, _ *cli.Context) (string, error) {
	arch := GetArch()
	javaURL := fmt.Sprintf("https://corretto.aws/downloads/latest/amazon-corretto-%d-%s-macos-jdk.pkg", version, arch)
	// Make sure the temp directory exists
	if _, err := os.Stat("./temp"); os.IsNotExist(err) {
		err := os.MkdirAll("./temp", 0755)
		if err != nil {
			return "", err
		}
	}
	// Download this file to the ./temp directory
	out, err := os.Create(fmt.Sprintf("./temp/java-%d-%s.pkg", version, arch))
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

	return fmt.Sprintf("./temp/java-%d-%s.pkg", version, arch), nil
}

func InstallJava(javaPath string, cliCtx *cli.Context) error {
	debug := cliCtx.Bool("debug")
	// If we're in debug, don't actually install Java, just print what we'd do
	// For install we're running: installer -pkg ./temp/java-21-x64.pkg -target CurrentUserHomeDirectory
	cmd := exec.Command("installer", "-pkg", javaPath, "-target", "CurrentUserHomeDirectory")
	if debug {
		// Just print the command we'd run
		log.Println(cmd.String())
		return nil
	}
	// We want to run the command and in real time print the output
	return RunCommandAndPipeOutput(cmd)
}

func PrintOSWarnings() {
	// Just let them know macOS isn't the _best_ OS for running a Minecraft server
	color.Set(color.FgYellow)
	log.Println("Warning: macOS is not the best OS for running a Minecraft server. You may experience issues with performance or stability.")
	color.Unset()
}

func WriteStartScript(path string, ramAmount int, cliCtx *cli.Context) error {
	// Write our start.sh file
	startScript := fmt.Sprintf(`#!/usr/bin/env sh

java -Xms%dM -Xmx%dM -XX:+AlwaysPreTouch -XX:+DisableExplicitGC -XX:+ParallelRefProcEnabled -XX:+PerfDisableSharedMem -XX:+UnlockExperimentalVMOptions -XX:+UseG1GC -XX:G1HeapRegionSize=8M -XX:G1HeapWastePercent=5 -XX:G1MaxNewSizePercent=40 -XX:G1MixedGCCountTarget=4 -XX:G1MixedGCLiveThresholdPercent=90 -XX:G1NewSizePercent=30 -XX:G1RSetUpdatingPauseTimePercent=5 -XX:G1ReservePercent=20 -XX:InitiatingHeapOccupancyPercent=15 -XX:MaxGCPauseMillis=200 -XX:MaxTenuringThreshold=1 -XX:SurvivorRatio=32 -Dusing.aikars.flags=https://mcflags.emc.gs -Daikars.new.flags=true -jar paper.jar nogui`, ramAmount, ramAmount)
	err := os.WriteFile(fmt.Sprintf("%s.sh", path), []byte(startScript), 0755)
	if err != nil {
		return err
	}
	// Make the file executable
	return os.Chmod(fmt.Sprintf("%s.sh", path), 0755)
}

func GetStartScript(path string) string {
	return fmt.Sprintf("%s.sh", path)
}
