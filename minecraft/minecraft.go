package minecraft

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func PropertyPrompt(property string, defaultValue string) string {
	// Y/N prompt
	ynPrompt := promptui.Prompt{
		Label:     fmt.Sprintf("Would you like to set the %s?", cases.Title(language.English).String(property)),
		IsConfirm: true,
	}
	result, err := ynPrompt.Run()
	if err != nil {
		return defaultValue
	}
	if result == "y" {
		// Prompt for the value
		prompt := promptui.Prompt{
			Label:   fmt.Sprintf("%s", cases.Title(language.English).String(property)),
			Default: defaultValue,
		}
		value, err := prompt.Run()
		if err != nil {
			return defaultValue
		}
		return value
	}
	return defaultValue
}

func ConfirmPrompt(property string) bool {
	// Y/N prompt
	ynPrompt := promptui.Prompt{
		Label:     fmt.Sprintf("Would you like to set the %s?", cases.Title(language.English).String(property)),
		IsConfirm: true,
	}
	result, err := ynPrompt.Run()
	if err != nil {
		return false
	}
	return result == "y"
}
