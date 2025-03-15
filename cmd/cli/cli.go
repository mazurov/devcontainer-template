package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mazurov/devcontainer-template/pkg/devctmpl"
	"github.com/spf13/cobra"
)

func main() {
	var (
		workspaceFolder string
		templateID      string
		templateArgs    string
	)

	cmd := &cobra.Command{
		Use:   "devcontainer-template",
		Short: "Apply devcontainer templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse template arguments
			options := make(map[string]string)
			if templateArgs != "" {
				if err := json.Unmarshal([]byte(templateArgs), &options); err != nil {
					return fmt.Errorf("invalid template arguments JSON: %w", err)
				}
			}

			// Call GenerateTemplate
			if err := devctmpl.GenerateTemplate(templateID, workspaceFolder, options); err != nil {
				return fmt.Errorf("failed to generate template: %w", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&workspaceFolder, "workspace-folder", "w", "", "Target workspace folder")
	cmd.Flags().StringVarP(&templateID, "template-id", "t", "", "Source template directory")
	cmd.Flags().StringVarP(&templateArgs, "template-args", "a", "", "Template arguments as JSON string")

	// Mark required flags
	cmd.MarkFlagRequired("workspace-folder")
	cmd.MarkFlagRequired("template-id")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
