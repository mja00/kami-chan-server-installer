package utils

import (
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
