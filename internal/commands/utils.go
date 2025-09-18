package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	"golang.org/x/term"
)

func GetSystemUsername() string {
	if user, err := user.Current(); err == nil {
		return user.Username
	}
	return "unknown"
}

func GetAgent() string {
	return "cli:" + GetSystemUsername() + "@" + os.Getenv("HOSTNAME")
}

func readLine() (string, error) {
	// Use buffered reader to read a line from standard input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n') // Reads until a newline
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	// Trim the spaces and newline characters
	input = strings.TrimSpace(input)
	return input, nil
}

func maybeReadInput(prompt string, existing string) string {
	if existing != "" {
		return existing
	}

	for true {

		fmt.Print(prompt)
		input, err := readLine()
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
		input, err := readLine()
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

// parseKeyValuePairs parses a comma-separated list of key=value pairs into a map.
// Example: "a=1,b=2" -> {"a":"1","b":"2"}
func parseKeyValuePairs(input string) (map[string]string, error) {
    result := make(map[string]string)
    if strings.TrimSpace(input) == "" {
        return result, nil
    }
    pairs := strings.Split(input, ",")
    for _, p := range pairs {
        p = strings.TrimSpace(p)
        if p == "" {
            continue
        }
        kv := strings.SplitN(p, "=", 2)
        if len(kv) != 2 {
            return nil, fmt.Errorf("invalid key=value pair: %q", p)
        }
        key := strings.TrimSpace(kv[0])
        val := strings.TrimSpace(kv[1])
        if key == "" {
            return nil, fmt.Errorf("empty key in pair: %q", p)
        }
        result[key] = val
    }
    return result, nil
}

// promptKeyValuePairs interactively reads key/value pairs until key is empty.
func promptKeyValuePairs() map[string]string {
    settings := make(map[string]string)
    for {
        key := maybeReadInput("Setting key (empty to finish): ", "")
        if key == "" {
            if len(settings) == 0 {
                fmt.Println("At least one setting is required")
                continue
            }
            break
        }
        value := maybeReadInput("Value: ", "")
        settings[key] = value
    }
    return settings
}

// parseCSVKeys parses a comma-separated list of keys into a slice.
func parseCSVKeys(input string) []string {
    if strings.TrimSpace(input) == "" {
        return []string{}
    }
    items := strings.Split(input, ",")
    out := make([]string, 0, len(items))
    for _, it := range items {
        v := strings.TrimSpace(it)
        if v != "" {
            out = append(out, v)
        }
    }
    return out
}

// promptKeys interactively reads keys until empty input.
func promptKeys() []string {
    keys := make([]string, 0, 1)
    for {
        key := maybeReadInput("Setting key to remove (empty to finish): ", "")
        if key == "" {
            if len(keys) == 0 {
                fmt.Println("At least one key is required")
                continue
            }
            break
        }
        keys = append(keys, key)
    }
    return keys
}
