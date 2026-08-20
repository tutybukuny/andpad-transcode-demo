package utils

import (
	"fmt"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func PasswordPrompt() (string, error) {
	fmt.Print("Enter password: ")
	bytePassword, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", fmt.Errorf("ReadPassword: %w", err)
	}

	return string(bytePassword), nil
}
