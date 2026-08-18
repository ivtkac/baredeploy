package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func readPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("stdin is not a terminal")
	}

	pwd, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(pwd), nil
}
