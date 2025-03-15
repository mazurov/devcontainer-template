package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mazurov/devcontainer-template/internal/logger"
	"github.com/mazurov/devcontainer-template/pkg/devctmpl"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	var (
		workspaceFolder string
		templateID      string
		templateArgs    string
		logLevel        string
		tmpDir          string
		keepTmpDir      bool
	)

	cmd := &cobra.Command{
		Use:   "devcontainer-template",
		Short: "Apply devcontainer templates",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Setup logging before running any command
			if err := logger.SetLevel(logLevel); err != nil {
				return fmt.Errorf("invalid log level: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.GetLogger()
			// Parse template arguments
			options := make(map[string]string)
			if templateArgs != "" {
				log.Debug("Parsing template arguments")
				if err := json.Unmarshal([]byte(templateArgs), &options); err != nil {
					return fmt.Errorf("invalid template arguments JSON: %w", err)
				}
			}

			log.WithFields(logrus.Fields{
				"templateID": templateID,
				"workspace":  workspaceFolder,
				"options":    options,
			}).Info("Generating template")

			// Call GenerateTemplate
			config := devctmpl.NewConfig()
			config.TmpRootDir = tmpDir
			config.KeepTmpDir = keepTmpDir
			if err := devctmpl.GenerateTemplateWithConfig(templateID, workspaceFolder, options, config); err != nil {
				return fmt.Errorf("failed to generate template: %w", err)
			}

			log.Info("Template generated successfully")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&workspaceFolder, "workspace-folder", "w", "", "Target workspace folder")
	cmd.Flags().StringVarP(&templateID, "template-id", "t", "", "Source template directory")
	cmd.Flags().StringVarP(&templateArgs, "template-args", "a", "", "Template arguments as JSON string")
	cmd.Flags().StringVarP(&tmpDir, "tmp-dir", "", "", "Directory to use for temporary files. If not provided, the system default will be used.")
	cmd.Flags().BoolVarP(&keepTmpDir, "keep-tmp-dir", "", false, "Keep temporary directory after execution")

	cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	// Mark required flags
	cmd.MarkFlagRequired("workspace-folder")
	cmd.MarkFlagRequired("template-id")

	if err := cmd.Execute(); err != nil {
		logger.GetLogger().Error(err)
		os.Exit(1)
	}
}
