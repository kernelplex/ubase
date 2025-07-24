package commands

import (
	"fmt"
	"os"
	"os/user"

	"golang.org/x/term"
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

func maybeReadPasswordInput(prompt string, value string, required bool) (*string, error) {
	if value != "" {
		return &value, nil
	}
	for {
		// First password entry
		fmt.Print(prompt)
		password, err := readPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}

		// Handle empty password
		if password == "" {
			if !required {
				return nil, nil
			}
			fmt.Println("Password is required")
			continue
		}

		// Password confirmation
		fmt.Print("Confirm password: ")
		confirm, err := readPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to read password confirmation: %w", err)
		}

		if password != confirm {
			fmt.Println("Passwords do not match")
			continue
		}

		return &password, nil
	}
}

func maybeReadInt64Input(prompt string, existing int64) int64 {
	if existing != 0 {
		return existing
	}

	for {
		fmt.Print(prompt)
		var input int64
		_, err := fmt.Scanln(&input)
		if err == nil {
			return input
		}
		fmt.Println("Invalid input: please enter a valid number")
	}
}

func readPassword() (string, error) {
	// For Unix-like systems
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	term := term.NewTerminal(os.Stdin, "")
	return term.ReadPassword("")
}
