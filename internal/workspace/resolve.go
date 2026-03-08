package workspace

import (
	"errors"
	"os"
	"path/filepath"
)

var errNoValidRoot = errors.New("no valid augury-node root found")

func ResolveRoot(flagPath, configPath, cwd string) (string, error) {
	if flagPath != "" {
		abs, err := absFrom(flagPath, cwd)
		if err != nil {
			return "", err
		}
		if err := ValidateRoot(abs); err != nil {
			return "", err
		}
		return abs, nil
	}
	if configPath != "" {
		abs, err := absFrom(configPath, cwd)
		if err != nil {
			return "", err
		}
		if err := ValidateRoot(abs); err != nil {
			return "", err
		}
		return abs, nil
	}
	root, err := findAncestorRoot(cwd)
	if err != nil {
		return "", err
	}
	return root, nil
}

func absFrom(p, cwd string) (string, error) {
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	return filepath.Abs(filepath.Join(cwd, p))
}

func findAncestorRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if err := ValidateRoot(dir); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errNoValidRoot
		}
		dir = parent
	}
}

func ValidateRoot(root string) error {
	required := []string{
		"scripts/devices",
		"scripts/lib",
		"pkg",
	}
	for _, rel := range required {
		p := filepath.Join(root, rel)
		info, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				return errNoValidRoot
			}
			return err
		}
		if !info.IsDir() {
			return errNoValidRoot
		}
	}
	return nil
}
