package common

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func ClearScreen() {
	os.Stdout.Write([]byte{0x1B, 0x5B, 0x33, 0x3B, 0x4A, 0x1B, 0x5B, 0x48, 0x1B, 0x5B, 0x32, 0x4A})
}

func Scan(placeholder string) (string, error) {
	fmt.Print(placeholder)
	return bufio.NewReader(os.Stdin).ReadString('\n')
}

func IsPathExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func ToAbsToSlash(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	path = filepath.ToSlash(path)
	if !IsPathExists(path) {
		return "", errors.New("initDir: path " + path + " doesn't exist.")
	}
	return path, err
}
