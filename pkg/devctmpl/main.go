package devctmpl

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/otiai10/copy"
)

// TemplateOption represents a configurable option in the template
type TemplateOption struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Proposals   []string `json:"proposals,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`
}

// DevContainerTemplate represents the structure of devcontainer-template.json
type DevContainerTemplate struct {
	ID               string                    `json:"id"`
	Version          string                    `json:"version"`
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	DocumentationURL string                    `json:"documentationURL,omitempty"`
	LicenseURL       string                    `json:"licenseURL,omitempty"`
	Options          map[string]TemplateOption `json:"options,omitempty"`
	Platforms        []string                  `json:"platforms,omitempty"`
	Publisher        string                    `json:"publisher,omitempty"`
	Keywords         []string                  `json:"keywords,omitempty"`
	OptionalPaths    []string                  `json:"optionalPaths,omitempty"`
}

type Config struct {
	TmpRootDir string
	KeepTmpDir bool
	OmitPaths  []string
}

// NewConfig creates a new Config with default values
func NewConfig() Config {
	return Config{
		KeepTmpDir: false,
		OmitPaths:  []string{},
	}
}

func GenerateTemplate(source string, target string, options map[string]string) error {
	return GenerateTemplateWithConfig(source, target, options, Config{})
}

func GenerateTemplateWithConfig(source string, target string, options map[string]string, cfg Config) error {
	// Prepare source directory
	source, cleanup, err := prepareSource(source, cfg.TmpRootDir)
	if err != nil {
		return fmt.Errorf("failed to prepare source: %w", err)
	}

	if !cfg.KeepTmpDir {
		defer cleanup()
	}

	template, err := loadTemplate(source)
	if err != nil {
		return err
	}

	// If template has no options defined but options were provided
	if template.Options == nil && len(options) > 0 {
		return fmt.Errorf("template has no options defined, but got options: %v", options)
	}

	// If options were provided but template has no options defined
	if err := checkOptions(template, options); err != nil {
		return err
	}
	tmpDir, err := copyTemplateToTemp(source, template, cfg.TmpRootDir, cfg.OmitPaths)
	if err != nil {
		return err
	}

	// Add default values for options not provided
	for optName, optDef := range template.Options {
		if _, exists := options[optName]; !exists && optDef.Default != "" {
			options[optName] = optDef.Default
		}
	}

	if err := replaceTemplateOptions(tmpDir, options); err != nil {
		return fmt.Errorf("failed to replace template options: %w", err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy processed template to target directory
	if err := copy.Copy(tmpDir, target); err != nil {
		return fmt.Errorf("failed to copy template to target directory: %w", err)
	}

	return nil
}

func GenerateFromEmbedWithConfig(source embed.FS, target string, options map[string]string, cfg Config) error {
	tmpDir, err := getTmpDir(cfg.TmpRootDir, "devcontainer-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	entries, err := source.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read embedded source: %w", err)
	}

	for _, entry := range entries {
		srcPath := entry.Name()
		dstPath := filepath.Join(tmpDir, srcPath)

		if entry.IsDir() {
			if err := copy.Copy(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", srcPath, err)
			}
		} else {
			data, err := source.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", srcPath, err)
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", dstPath, err)
			}
		}
	}

	return GenerateTemplateWithConfig(tmpDir, target, options, cfg)
}

func getTmpDir(tmpRootDir string, pattern string) (string, error) {
	// Create temporary directory
	return os.MkdirTemp(tmpRootDir, pattern)
}

// CopyTemplateToTemp copies the template files to a temporary directory
func copyTemplateToTemp(sourceDir string, template *DevContainerTemplate, tmpRootDir string, omitPaths []string) (string, error) {
	tmpDir, err := getTmpDir(tmpRootDir, "devcontainer-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Copy .devcontainer folder
	devcontainerSrc := filepath.Join(sourceDir, ".devcontainer")
	devcontainerDst := filepath.Join(tmpDir, ".devcontainer")

	if err := copy.Copy(devcontainerSrc, devcontainerDst); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to copy .devcontainer folder: %w", err)
	}

	// Copy optional paths
	sourceDirFs := os.DirFS(sourceDir)
	for _, pattern := range template.OptionalPaths {
		matches, err := fs.Glob(sourceDirFs, pattern)
		if err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
		}

		for _, match := range matches {
			absMatch := filepath.Join(sourceDir, match)
			relPath, err := filepath.Rel(sourceDir, absMatch)
			if err != nil {
				os.RemoveAll(tmpDir)
				return "", fmt.Errorf("failed to get relative path: %w", err)
			}

			dst := filepath.Join(tmpDir, relPath)
			if err := copy.Copy(absMatch, dst); err != nil {
				os.RemoveAll(tmpDir)
				return "", fmt.Errorf("failed to copy '%s': %w", relPath, err)
			}
		}
	}

	// Remove folders and files in tmpDir that match omitPaths globs
	tmpDirFs := os.DirFS(tmpDir)
	for _, pattern := range omitPaths {
		matches, err := fs.Glob(tmpDirFs, pattern)
		if err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
		}

		for _, match := range matches {
			if err := os.RemoveAll(filepath.Join(tmpDir, match)); err != nil {
				os.RemoveAll(tmpDir)
				return "", fmt.Errorf("failed to remove '%s': %w", match, err)
			}
		}
	}

	return tmpDir, nil
}

// ReplaceTemplateOptions walks through all files in the directory and replaces
// template variables of the form ${templateOption:key} with their corresponding values
func replaceTemplateOptions(dir string, options map[string]string) error {
	// Compile regex for finding template variables
	varRegex := regexp.MustCompile(`\${templateOption:([^}]+)}`)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Check if file contains any template variables
		if !varRegex.Match(content) {
			return nil
		}

		// Replace all template variables
		newContent := varRegex.ReplaceAllFunc(content, func(match []byte) []byte {
			// Extract key from ${templateOption:key}
			key := varRegex.FindSubmatch(match)[1]

			// Get value from options map
			if value, exists := options[string(key)]; exists {
				return []byte(value)
			}
			// If no value found, leave original template variable
			return match
		})

		// Write modified content back to file
		if err := os.WriteFile(path, newContent, info.Mode()); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}

		return nil
	})
}

func checkOptions(template *DevContainerTemplate, options map[string]string) error {
	names := make([]string, 0, len(template.Options))
	if options == nil {
		options = make(map[string]string)
	}
	for name := range template.Options {
		names = append(names, name)
	}
	sort.Strings(names)

	for optName := range options {
		if _, exists := template.Options[optName]; !exists {
			return fmt.Errorf("option '%s' is not defined in template (available options: %v)",
				optName,
				names,
			)
		}
	}
	return nil
}

func parseTemplate(content []byte) (*DevContainerTemplate, error) {
	var template DevContainerTemplate
	if err := json.Unmarshal(content, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}
	return &template, nil
}

func loadTemplate(dir string) (*DevContainerTemplate, error) {
	content, err := os.ReadFile(filepath.Join(dir, "devcontainer-template.json"))
	if err != nil {
		return nil, fmt.Errorf("error reading devcontainer-template.json in directory %s: %v", dir, err)
	}

	// Parse the template first
	template, err := parseTemplate(content)
	if err != nil {
		return nil, fmt.Errorf("error parsing template in directory %s: %v", dir, err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".devcontainer", "devcontainer.json")); os.IsNotExist(err) {
		return template, fmt.Errorf(".devcontainer/devcontainer.json file does not exist in directory %s", dir)
	}

	return template, nil
}

// PrepareSource downloads/copies the source to a temporary directory
func prepareSource(source string, tmpDirRoot string) (string, func(), error) {
	tmpDir, err := getTmpDir(tmpDirRoot, "devcontainer-source-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// For local directories, use copy instead of go-getter
	if info, err := os.Stat(source); err == nil && info.IsDir() {
		if err := copy.Copy(source, tmpDir); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("failed to copy local directory: %w", err)
		}
		return tmpDir, cleanup, nil
	}

	// Check if it's an OCI reference
	if isOCIRepository(source) {
		if err := pullOCITemplate(source, tmpDir); err != nil {
			cleanup()
			return "", nil, err
		}
		return tmpDir, cleanup, nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Expand . and .. if source starts with file://
	source = strings.TrimPrefix(source, "file://")
	// Handle other sources using go-getter
	client := &getter.Client{
		Src:  source,
		Dst:  tmpDir,
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
		Options: []getter.ClientOption{
			getter.WithProgress(nil),
		},
	}

	if err := client.Get(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to get source: %w", err)
	}

	// Find the actual template directory
	templateDir, err := findTemplateDir(tmpDir)
	if err != nil {
		cleanup()
		return "", nil, err
	}

	return templateDir, cleanup, nil
}

func findTemplateDir(dir string) (string, error) {
	// Check current directory
	if _, err := os.Stat(filepath.Join(dir, "devcontainer-template.json")); err == nil {
		return dir, nil
	}

	// Check immediate subdirectories
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			path := filepath.Join(dir, entry.Name())
			if _, err := os.Stat(filepath.Join(path, "devcontainer-template.json")); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("devcontainer-template.json not found in %s or its subdirectories", dir)
}
