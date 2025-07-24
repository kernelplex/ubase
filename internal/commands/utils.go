package commands

import (
	"fmt"
	"os"
	"os/user"
)

func GetSystemUsername() string {
	if user, err := user.Current(); err == nil {
		return user.Username
	}
	return "unknown"
}

func GetAgent() string {
	return GetSystemUsername() + "@" + os.Getenv("HOSTNAME")
}

func maybeReadInput(prompt string, existing string) string {
	if existing != "" {
		return existing
	}

	for true {

		fmt.Print(prompt)
		var input string
		_, err := fmt.Scanln(&input)
		if err == nil {
			return input
		}
		fmt.Println("Invalid input: " + err.Error())
	}
	return ""
}

func maybeReadOptionalInput(prompt string, existing string) *string {
	if existing != "" {
		return &existing
	}

	for true {

		fmt.Print(prompt)
		var input string
		_, err := fmt.Scanln(&input)
		if input == "" && err == nil {
			return nil
		}
		if err == nil {
			return &input
		}
		fmt.Println("Invalid input: " + err.Error())
	}
	return nil
}
