package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"deploytools/internal/config"
	"deploytools/pkg/utils"
)

func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage project configurations",
		Long:  `Add, edit, delete, and list project configurations.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a new project",
		Run:   runAddProject,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "edit [id]",
		Short: "Edit an existing project",
		Args:  cobra.ExactArgs(1),
		Run:   runEditProject,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		Run:   runDeleteProject,
	})

	return cmd
}

func runAddProject(cmd *cobra.Command, args []string) {
	project := promptProjectInfo(config.Project{
		UpdateMode: config.UpdateModeDiff,
	})

	if err := utils.ValidateLocalPath(project.LocalPath); err != nil {
		fmt.Printf("Invalid local path: %v\n", err)
		return
	}

	if err := utils.ValidateRemotePath(project.RemotePath); err != nil {
		fmt.Printf("Invalid remote path: %v\n", err)
		return
	}

	project, err := cfgManager.AddProject(project)
	if err != nil {
		fmt.Printf("Failed to add project: %v\n", err)
		return
	}

	fmt.Printf("Project '%s' added successfully (ID: %s)\n", project.Name, project.ID)
}

func runEditProject(cmd *cobra.Command, args []string) {
	id := args[0]
	project := cfgManager.GetProject(id)
	if project == nil {
		fmt.Printf("Project not found: %s\n", id)
		return
	}

	updated := promptProjectInfo(*project)
	updated.ID = id

	if err := cfgManager.UpdateProject(id, updated); err != nil {
		fmt.Printf("Failed to update project: %v\n", err)
		return
	}

	fmt.Printf("Project '%s' updated successfully\n", updated.Name)
}

func runDeleteProject(cmd *cobra.Command, args []string) {
	id := args[0]
	project := cfgManager.GetProject(id)
	if project == nil {
		fmt.Printf("Project not found: %s\n", id)
		return
	}

	fmt.Printf("Are you sure you want to delete project '%s'? (y/N): ", project.Name)
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Delete cancelled")
		return
	}

	if err := cfgManager.DeleteProject(id); err != nil {
		fmt.Printf("Failed to delete project: %v\n", err)
		return
	}

	fmt.Printf("Project '%s' deleted successfully\n", project.Name)
}

func promptProjectInfo(defaults config.Project) config.Project {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Project Name [%s]: ", defaults.Name)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaults.Name
	}

	fmt.Printf("Local Path [%s]: ", defaults.LocalPath)
	localPath, _ := reader.ReadString('\n')
	localPath = strings.TrimSpace(localPath)
	if localPath == "" {
		localPath = defaults.LocalPath
	}

	servers := cfgManager.GetServers()
	if len(servers) > 0 {
		fmt.Println("\nAvailable servers:")
		for i, s := range servers {
			fmt.Printf("  %d. %s (%s)\n", i+1, s.Name, s.IP)
		}
	}
	fmt.Printf("Server ID [%s]: ", defaults.ServerID)
	serverID, _ := reader.ReadString('\n')
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		serverID = defaults.ServerID
	}

	fmt.Printf("Remote Path [%s]: ", defaults.RemotePath)
	remotePath, _ := reader.ReadString('\n')
	remotePath = strings.TrimSpace(remotePath)
	if remotePath == "" {
		remotePath = defaults.RemotePath
	}

	defaultMode := defaults.UpdateMode
	if defaultMode == "" {
		defaultMode = config.UpdateModeDiff
	}
	fmt.Printf("Update Mode (diff/replace_file/replace_dir) [%s]: ", defaultMode)
	modeStr, _ := reader.ReadString('\n')
	modeStr = strings.TrimSpace(strings.ToLower(modeStr))
	var updateMode config.UpdateMode
	switch modeStr {
	case "replace_file":
		updateMode = config.UpdateModeReplaceFile
	case "replace_dir":
		updateMode = config.UpdateModeReplaceDir
	default:
		updateMode = config.UpdateModeDiff
	}

	fmt.Printf("Post-deploy Script [%s]: ", defaults.PostScript)
	postScript, _ := reader.ReadString('\n')
	postScript = strings.TrimSpace(postScript)
	if postScript == "" {
		postScript = defaults.PostScript
	}

	fmt.Printf("Exclude Patterns (comma-separated): ")
	excludeStr, _ := reader.ReadString('\n')
	excludeStr = strings.TrimSpace(excludeStr)
	var excludePatterns []string
	if excludeStr != "" {
		for _, p := range strings.Split(excludeStr, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				excludePatterns = append(excludePatterns, p)
			}
		}
	}
	if len(excludePatterns) == 0 {
		excludePatterns = defaults.ExcludePatterns
	}

	return config.Project{
		Name:            name,
		LocalPath:       localPath,
		ServerID:        serverID,
		RemotePath:      remotePath,
		UpdateMode:      updateMode,
		PostScript:      postScript,
		ExcludePatterns: excludePatterns,
	}
}

func printProjectsTable(projects []config.Project) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tLOCAL PATH\tSERVER ID\tUPDATE MODE")
	for _, p := range projects {
		serverName := p.ServerID
		if server := cfgManager.GetServer(p.ServerID); server != nil {
			serverName = server.Name
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			utils.TruncateString(p.ID, 8),
			p.Name,
			utils.TruncateString(p.LocalPath, 30),
			serverName,
			p.UpdateMode,
		)
	}
	w.Flush()
}
