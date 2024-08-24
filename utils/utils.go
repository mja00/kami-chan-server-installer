package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type JavaVersion struct {
	Version string
	Major   int
	Minor   int
}

func GetJavaVersion() (JavaVersion, error) {
	// Just run java -version and parse out the version
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return JavaVersion{}, err
	}
	// On modern Java (which we care about), the output will follow:
	// openjdk version "21.0.3" 2024-04-16 LTS
	// OpenJDK Runtime Environment Temurin-21.0.3+9 (build 21.0.3+9-LTS)
	// OpenJDK 64-Bit Server VM Temurin-21.0.3+9 (build 21.0.3+9-LTS, mixed mode)
	// All we care about is the first line, which will be the version
	// Essentially just grab the string in the quotes
	version := strings.Split(string(output), "\"")[1]
	versionSplit := strings.Split(version, ".")
	major, err := strconv.Atoi(versionSplit[0])
	if err != nil {
		return JavaVersion{}, err
	}
	minor, err := strconv.Atoi(versionSplit[1])
	if err != nil {
		return JavaVersion{}, err
	}
	return JavaVersion{
		Version: version,
		Major:   major,
		Minor:   minor,
	}, nil
}

func MCVersionToJavaMajor(mcVersion string) (int, error) {
	// 1.8 -> 8+
	// 1.17 -> 16+
	// 1.18 -> 17+
	// 1.20.5 -> 21+
	// Just return the major version for the given MC version
	versionSplit := strings.Split(mcVersion, ".")
	mcMajor, err := strconv.Atoi(versionSplit[1])
	if err != nil {
		return 0, err
	}
	// 1.8 -> 1.16
	if mcMajor >= 8 && mcMajor < 17 {
		return 8, nil
	} else if mcMajor >= 17 && mcMajor < 18 {
		return 16, nil
	} else if mcMajor >= 18 && mcMajor < 20 {
		return 17, nil
	} else if mcMajor >= 20 {
		return 21, nil
	}
	return 0, fmt.Errorf("invalid MC version: %s", mcVersion)
}

func GetSha256Hash(filePath string) (string, error) {
	// Stream the file, that way we don't have to load the whole thing into memory
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func GetServerFolder(path string, cCtx *cli.Context) string {
	serverFolder := cCtx.String("server-dir")
	if _, err := os.Stat(serverFolder); os.IsNotExist(err) {
		err := os.Mkdir(serverFolder, 0755)
		if err != nil {
			return ""
		}
	}
	return fmt.Sprintf("%s/%s", serverFolder, path)
}

func RunCommandAndPipeOutput(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("[STDOUT] %s\n", scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[STDERR] %s\n", scanner.Text())
		}
	}()
	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
