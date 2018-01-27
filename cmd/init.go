package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/devlocker/tychus/tychus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Tychus with a configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		// Get the current working directory
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("Cant get working dir")
			return
		}

		// Generate a configuration based on project.
		c, err := detectLangauge(wd)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Conig already exist? Ask to overwrite if so.
		if _, err = os.Stat(configFile); !os.IsNotExist(err) {
			fmt.Println("Config file already exists")
			reader := bufio.NewReader(os.Stdin)

			for {
				fmt.Print("Replace configuration file? [y/n]: ")

				response, err := reader.ReadString('\n')
				if err != nil {
					log.Fatal(err)
				}

				response = strings.ToLower(strings.TrimSpace(response))
				if response == "y" || response == "yes" {
					break
				} else if response == "n" || response == "no" {
					return
				}
			}
		}

		// Write config file to disk.
		fmt.Printf("Creating %v\n", configFile)
		c.Write(configFile)
	},
}

// Naive project language detection. Scans the top level files to determine
// project type. Return a configuration with sensible defaults.
func detectLangauge(dir string) (*tychus.Configuration, error) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		return nil, err
	}

	c := &tychus.Configuration{
		Build: tychus.BuildConfig{
			BuildCommand: "go build -i",
			BinName:      "tychus-bin",
			Enabled:      true,
			TargetPath:   "tmp/",
		},
		Watch: tychus.WatchConfig{
			Extensions: []string{".go"},
			Ignore:     []string{"node_modules", "tmp", "log", "vendor"},
		},
		Proxy: tychus.ProxyConfig{
			Enabled:   true,
			AppPort:   3000,
			ProxyPort: 4000,
			Timeout:   10,
		},
	}

	// Go Project?
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext == ".go" {
			// Already configured for Go
			return c, nil
		}
	}

	// Ruby Project?
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if f.Name() == "Gemfile" || ext == ".rb" {
			c.Build.Enabled = false
			c.Build.BuildCommand = ""
			c.Build.BinName = ""
			c.Watch.Extensions = []string{".rb"}
			return c, nil
		}
	}

	// Python Project?
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext == ".py" {
			c.Build.Enabled = false
			c.Build.BuildCommand = ""
			c.Build.BinName = ""
			c.Watch.Extensions = []string{".py"}
			return c, nil
		}
	}

	// Rust Project?
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if f.Name() == "Cargo.toml" || ext == ".rs" {
			c.Build.BuildCommand = "rustc main.rs"
			c.Watch.Extensions = []string{".rs"}
			return c, nil
		}
	}

	// JS Project?
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if f.Name() == "package.json" || ext == ".js" {
			c.Build.Enabled = false
			c.Build.BuildCommand = ""
			c.Build.BinName = ""
			c.Watch.Extensions = []string{".js"}
			return c, nil
		}
	}

	// Something else, return default
	return c, nil
}
