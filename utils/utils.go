package utils

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
		// Check to see if the error is a command not found error
		if strings.Contains(err.Error(), "executable file not found") {
			return JavaVersion{
				Version: "unknown",
				Major:   0,
				Minor:   0,
			}, nil
		}
		return JavaVersion{
			Version: "unknown",
			Major:   0,
			Minor:   0,
		}, err
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
		err := os.MkdirAll(serverFolder, 0755)
		if err != nil {
			return ""
		}
	}
	return filepath.Join(serverFolder, path)
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
			log.Printf("[STDOUT] %s\n", scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[STDERR] %s\n", scanner.Text())
		}
	}()
	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func RunCommandAndPipeAllSTD(cmd *exec.Cmd, noPrefix bool) error {
	// We need to pipe buth STDOUT and STDERR to the terminal but also STDIN to the command we're running
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if !noPrefix {
				log.Printf("[STDOUT] %s\n", scanner.Text())
			} else {
				fmt.Println(scanner.Text())
			}
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if !noPrefix {
				log.Printf("[STDERR] %s\n", scanner.Text())
			} else {
				fmt.Println(scanner.Text())
			}
		}
	}()
	// Now we need to pipe STDIN to the command we're running
	go func() {
		// Read from this program's STDIN
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			// Write
			_, err := io.WriteString(stdin, scanner.Text()+"\n")
			if err != nil {
				return
			}
		}
	}()
	// Now we need to wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func ExtractTarGz(gzipStream io.Reader, extractPath string) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	// Make sure the extract path exists
	if _, err := os.Stat(extractPath); os.IsNotExist(err) {
		err := os.MkdirAll(extractPath, 0755)
		if err != nil {
			log.Fatalf("ExtractTarGz: MkdirAll() failed: %s", err.Error())
		}
	}

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		target := filepath.Join(extractPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				log.Fatalf("ExtractTarGz: MkdirAll() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Fatalf(
				"ExtractTarGz: unknown type: %b in %s",
				header.Typeflag,
				header.Name)
		}

	}
}
