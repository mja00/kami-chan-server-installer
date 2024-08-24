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
	"runtime"
	"strings"
)

func PrintOSWarnings() {
	// Print nothing, linux is good :)
	log.Println("Good job for using Linux!")
}

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

func DownloadJava(version int, cliCtx *cli.Context) (string, error) {
	if !cliCtx.Bool("install-java-please") {
		log.Println("\n\nWe won't actually download Java, as we want you to use 'apt-get' to install it.")
		log.Println("Don't worry! We'll walk you through it!\n\n")
		return "", nil
	}
	arch := GetArch()
	javaURL := fmt.Sprintf("https://corretto.aws/downloads/latest/amazon-corretto-%d-%s-linux-jdk.deb", version, arch)

	if _, err := os.Stat("./temp"); os.IsNotExist(err) {
		err := os.Mkdir("./temp", 0755)
		if err != nil {
			return "", err
		}
	}
	out, err := os.Create(fmt.Sprintf("./temp/java-%d-%s.deb", version, arch))
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

	return fmt.Sprintf("./temp/java-%d-%s.deb", version, arch), nil
}

func InstallJava(javaPath string, cliCtx *cli.Context) error {
	if !cliCtx.Bool("install-java-please") {
		log.Println("When the script exits, go ahead and run the commands found in the Paper guide here: https://docs.papermc.io/misc/java-install#ubuntudebian")
		log.Println("If you're running CentOS, RHEL, Fedora, openSUSE, SLES, or any other RPM-based Linux distribution use: https://docs.papermc.io/misc/java-install#rpm-based")
		return nil
	}
	debug := cliCtx.Bool("debug")
	// The user really wants us to install Java for them. We need to ensure we're root. Otherwise we cannot
	// run id command and grab the user ID
	userCommand := exec.Command("id", "-u")
	output, err := userCommand.CombinedOutput()
	if err != nil {
		return err
	}
	userId := strings.TrimSpace(string(output))

	if userId != "0" {
		return fmt.Errorf("you must be root to install Java automatically")
	}
	// We're root, so lets download the deb file and install it
	// Install: dpkg -i ./temp/java-21-x64.deb
	cmd := exec.Command("dpkg", "-i", javaPath)
	if debug {
		// Just print the command we'd run
		log.Println(cmd.String())
		return nil
	}
	// We want to run the command and in real time print the output
	return RunCommandAndPipeOutput(cmd)
}
