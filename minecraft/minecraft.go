package minecraft

import (
	"github.com/spf13/viper"
	"os"
)

func ReadServerProperties(filePath string) error {
	// Create the file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create the file
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	viper.SetConfigFile(filePath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

func WriteServerProperties(filePath string) error {
	// Write back out the file from viper
	return viper.WriteConfigAs(filePath)
}
