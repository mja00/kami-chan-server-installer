package cfg

import (
	"github.com/goccy/go-json"
	"os"
)

// This'll handle our config file that'll store information about the server for our use

type Config struct {
	MinecraftVersion string `json:"minecraft_version"`
	PaperBuild       string `json:"paper_build"`
	LastPaperBuild   string `json:"last_paper_build"`
}

func NewConfig() *Config {
	return &Config{
		MinecraftVersion: "",
		PaperBuild:       "",
		LastPaperBuild:   "",
	}
}

func (c *Config) Save(path string) error {
	// If our file doesn't exist, create it
	var file *os.File
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create the file
		file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	} else {
		// Open the file
		file, err = os.OpenFile(path, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	// Write to the file
	err := json.NewEncoder(file).Encode(c)
	if err != nil {
		return err
	}
	return nil

}

func (c *Config) Load(path string) error {
	// Read the config from a file
	var file *os.File
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Just make the file
		file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	// Read the file
	err = json.NewDecoder(file).Decode(c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) GetPaperBuild() string {
	return c.PaperBuild
}

func (c *Config) SetPaperBuild(build string) {
	c.PaperBuild = build
}

func (c *Config) GetLastPaperBuild() string {
	return c.LastPaperBuild
}

func (c *Config) SetLastPaperBuild(build string) {
	c.LastPaperBuild = build
}

func (c *Config) SetMinecraftVersion(version string) {
	c.MinecraftVersion = version
}

func (c *Config) GetMinecraftVersion() string {
	return c.MinecraftVersion
}
