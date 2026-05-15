package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"deploytools/internal/config"
)

type ExportedConfig struct {
	Servers     []config.Server       `json:"servers" yaml:"servers"`
	Projects    []config.Project      `json:"projects" yaml:"projects"`
	ProjectGroups []config.ProjectGroup `json:"project_groups" yaml:"project_groups"`
}

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration import/export",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "export [output-file]",
		Short: "Export configuration to file",
		Args:  cobra.ExactArgs(1),
		Run:   runExportConfig,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "import [input-file]",
		Short: "Import configuration from file",
		Args:  cobra.ExactArgs(1),
		Run:   runImportConfig,
	})

	cmd.PersistentFlags().String("format", "yaml", "Output format: yaml or json")

	return cmd
}

func runExportConfig(cmd *cobra.Command, args []string) {
	outputFile := args[0]
	format, _ := cmd.Flags().GetString("format")

	exported := ExportedConfig{
		Servers:       cfgManager.GetServers(),
		Projects:      cfgManager.GetProjects(),
		ProjectGroups: cfgManager.GetProjectGroups(),
	}

	var data []byte
	var err error

	switch strings.ToLower(format) {
	case "json":
		data, err = json.MarshalIndent(exported, "", "  ")
	case "yaml", "yml":
		data, err = yaml.Marshal(exported)
	default:
		fmt.Printf("Unsupported format: %s (use yaml or json)\n", format)
		return
	}

	if err != nil {
		fmt.Printf("Failed to marshal configuration: %v\n", err)
		return
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
		return
	}

	fmt.Printf("Configuration exported to %s successfully\n", outputFile)
	fmt.Printf("  Servers: %d\n", len(exported.Servers))
	fmt.Printf("  Projects: %d\n", len(exported.Projects))
	fmt.Printf("  Groups: %d\n", len(exported.ProjectGroups))
}

func runImportConfig(cmd *cobra.Command, args []string) {
	inputFile := args[0]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		return
	}

	var imported ExportedConfig
	if err := yaml.Unmarshal(data, &imported); err != nil {
		if err := json.Unmarshal(data, &imported); err != nil {
			fmt.Printf("Failed to parse configuration: %v\n", err)
			return
		}
	}

	fmt.Println("\n=== Import Preview ===")
	fmt.Printf("Servers to import: %d\n", len(imported.Servers))
	fmt.Printf("Projects to import: %d\n", len(imported.Projects))
	fmt.Printf("Project groups to import: %d\n", len(imported.ProjectGroups))

	fmt.Print("\nImport mode: [m]erge / [o]verwrite / [c]ancel: ")
	reader := bufio.NewReader(os.Stdin)
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(strings.ToLower(mode))

	switch mode {
	case "m", "merge":
		if err := mergeConfig(imported); err != nil {
			fmt.Printf("Merge failed: %v\n", err)
			return
		}
	case "o", "overwrite":
		fmt.Print("WARNING: This will delete all existing configuration. Are you sure? (y/N): ")
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Import cancelled")
			return
		}
		if err := overwriteConfig(imported); err != nil {
			fmt.Printf("Overwrite failed: %v\n", err)
			return
		}
	default:
		fmt.Println("Import cancelled")
		return
	}

	fmt.Println("\nConfiguration imported successfully!")
}

func mergeConfig(imported ExportedConfig) error {
	for _, server := range imported.Servers {
		existing := cfgManager.GetServer(server.ID)
		if existing == nil {
			if _, err := cfgManager.AddServer(server); err != nil {
				return err
			}
		}
	}

	for _, project := range imported.Projects {
		existing := cfgManager.GetProject(project.ID)
		if existing == nil {
			if _, err := cfgManager.AddProject(project); err != nil {
				return err
			}
		}
	}

	for _, group := range imported.ProjectGroups {
		existing := cfgManager.GetProjectGroup(group.ID)
		if existing == nil {
			if _, err := cfgManager.AddProjectGroup(group); err != nil {
				return err
			}
		}
	}

	return nil
}

func overwriteConfig(imported ExportedConfig) error {
	for _, server := range cfgManager.GetServers() {
		if err := cfgManager.DeleteServer(server.ID); err != nil {
			return err
		}
	}

	for _, project := range cfgManager.GetProjects() {
		if err := cfgManager.DeleteProject(project.ID); err != nil {
			return err
		}
	}

	for _, group := range cfgManager.GetProjectGroups() {
		if err := cfgManager.DeleteProjectGroup(group.ID); err != nil {
			return err
		}
	}

	for _, server := range imported.Servers {
		if _, err := cfgManager.AddServer(server); err != nil {
			return err
		}
	}

	for _, project := range imported.Projects {
		if _, err := cfgManager.AddProject(project); err != nil {
			return err
		}
	}

	for _, group := range imported.ProjectGroups {
		if _, err := cfgManager.AddProjectGroup(group); err != nil {
			return err
		}
	}

	return nil
}
