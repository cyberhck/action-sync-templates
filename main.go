package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	DefaultConfigPath = ".github/.template-sync.json"
	TemplateRepoDir   = "template-repo"
)

type Config struct {
	Files []string `json:"files"`
}

func copyFileOrDir(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			relPath, _ := filepath.Rel(src, path)
			targetPath := filepath.Join(dest, relPath)

			if info.IsDir() {
				return os.MkdirAll(targetPath, 0755)
			} else {
				return copyFile(path, targetPath)
			}
		})
	}

	// Single file
	return copyFile(src, dest)
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

func main() {
	configPath := DefaultConfigPath
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Files) == 0 {
		fmt.Println("::notice::No files to sync.")
		return
	}

	for _, file := range cfg.Files {
		src := filepath.Join(TemplateRepoDir, file)
		dest := file

		err := copyFileOrDir(src, dest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error copying %s: %v\n", file, err)
		} else {
			fmt.Printf("::notice::Copied: %s\n", file)
		}
	}
}
