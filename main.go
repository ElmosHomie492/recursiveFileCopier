package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <source_directory> <destination_directory>")
		return
	}

	config, err := getConfiguration(os.Args)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	err = copyFiles(config)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func copyFiles(config *Config) error {
	if err := os.MkdirAll(config.targetDirectory, os.ModePerm); err != nil {
		return err
	}

	err := filepath.WalkDir(config.sourceDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if config.isDirectoryBlacklisted(path) {
			return filepath.SkipDir
		}

		if !d.IsDir() && config.isTargetFileType(path) {
			destPath := filepath.Join(config.targetDirectory, d.Name())

			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				return err
			}

			fmt.Printf("SOURCE: %s\n", path)
			fmt.Printf("Copied: %s\n", destPath)
		}

		return nil
	})

	return err
}

func (c *Config) isTargetFileType(path string) bool {
	extension := strings.Replace(filepath.Ext(path), ".", "", -1)
	for _, targetExt := range c.fileTypes {
		if extension == targetExt {
			return true
		}
	}

	return false
}

type Config struct {
	sourceDirectory        string
	targetDirectory        string
	fileTypes              []string
	blacklistedDirectories []string
}

func getConfiguration(args []string) (*Config, error) {
	if len(args) < 3 {
		return nil, errors.New("ERROR: Not enough arguments\n\nUsage: go run copyImages.go <source_directory> <destination_directory>")
	}

	var config Config

	config.sourceDirectory = args[1]
	config.targetDirectory = args[2]

	err := config.setFileTypes()
	if err != nil {
		return nil, err
	}

	err = config.setBlacklistedDirectories()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) setFileTypes() error {
	file, err := os.Open("./fileTypes.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		c.fileTypes = append(c.fileTypes, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (c *Config) setBlacklistedDirectories() error {
	file, err := os.Open("./blacklistedDirectories.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		c.blacklistedDirectories = append(c.blacklistedDirectories, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (c *Config) isDirectoryBlacklisted(path string) bool {
	for _, blacklistedDirectory := range c.blacklistedDirectories {
		if strings.Contains(path, blacklistedDirectory) {
			return true
		}
	}

	return false
}
