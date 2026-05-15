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

func NewGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage project groups",
		Long:  `Create, edit, delete, and list project groups for batch deployment.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a new project group",
		Run:   runAddGroup,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "edit [id]",
		Short: "Edit an existing project group",
		Args:  cobra.ExactArgs(1),
		Run:   runEditGroup,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a project group",
		Args:  cobra.ExactArgs(1),
		Run:   runDeleteGroup,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show [id]",
		Short: "Show group details",
		Args:  cobra.ExactArgs(1),
		Run:   runShowGroup,
	})

	return cmd
}

func runAddGroup(cmd *cobra.Command, args []string) {
	group := promptGroupInfo(config.ProjectGroup{})

	group, err := cfgManager.AddProjectGroup(group)
	if err != nil {
		fmt.Printf("Failed to add group: %v\n", err)
		return
	}

	fmt.Printf("Group '%s' added successfully (ID: %s)\n", group.Name, group.ID)
}

func runEditGroup(cmd *cobra.Command, args []string) {
	id := args[0]
	group := cfgManager.GetProjectGroup(id)
	if group == nil {
		fmt.Printf("Group not found: %s\n", id)
		return
	}

	updated := promptGroupInfo(*group)
	updated.ID = id

	if err := cfgManager.UpdateProjectGroup(id, updated); err != nil {
		fmt.Printf("Failed to update group: %v\n", err)
		return
	}

	fmt.Printf("Group '%s' updated successfully\n", updated.Name)
}

func runDeleteGroup(cmd *cobra.Command, args []string) {
	id := args[0]
	group := cfgManager.GetProjectGroup(id)
	if group == nil {
		fmt.Printf("Group not found: %s\n", id)
		return
	}

	fmt.Printf("Are you sure you want to delete group '%s'? (y/N): ", group.Name)
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Delete cancelled")
		return
	}

	if err := cfgManager.DeleteProjectGroup(id); err != nil {
		fmt.Printf("Failed to delete group: %v\n", err)
		return
	}

	fmt.Printf("Group '%s' deleted successfully\n", group.Name)
}

func runShowGroup(cmd *cobra.Command, args []string) {
	id := args[0]
	group := cfgManager.GetProjectGroup(id)
	if group == nil {
		fmt.Printf("Group not found: %s\n", id)
		return
	}

	fmt.Printf("\nGroup: %s\n", group.Name)
	if group.Description != "" {
		fmt.Printf("Description: %s\n", group.Description)
	}
	fmt.Printf("Created: %s\n", group.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nProjects in this group (%d):\n", len(group.ProjectIDs))

	if len(group.ProjectIDs) > 0 {
		var projects []config.Project
		for _, pid := range group.ProjectIDs {
			if p := cfgManager.GetProject(pid); p != nil {
				projects = append(projects, *p)
			}
		}
		if len(projects) > 0 {
			printProjectsTable(projects)
		} else {
			fmt.Println("  (No valid projects found)")
		}
	}
}

func promptGroupInfo(defaults config.ProjectGroup) config.ProjectGroup {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Group Name [%s]: ", defaults.Name)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaults.Name
	}

	fmt.Printf("Description [%s]: ", defaults.Description)
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)
	if desc == "" {
		desc = defaults.Description
	}

	projects := cfgManager.GetProjects()
	if len(projects) > 0 {
		fmt.Println("\nAvailable projects:")
		for i, p := range projects {
			fmt.Printf("  %d. %s\n", i+1, p.Name)
		}
	}

	var defaultIDs string
	if len(defaults.ProjectIDs) > 0 {
		defaultIDs = strings.Join(defaults.ProjectIDs, ",")
	}
	fmt.Printf("Project IDs (comma-separated) [%s]: ", defaultIDs)
	idsStr, _ := reader.ReadString('\n')
	idsStr = strings.TrimSpace(idsStr)
	if idsStr == "" {
		idsStr = defaultIDs
	}

	var projectIDs []string
	if idsStr != "" {
		for _, id := range strings.Split(idsStr, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				projectIDs = append(projectIDs, id)
			}
		}
	}

	return config.ProjectGroup{
		Name:        name,
		Description: desc,
		ProjectIDs:  projectIDs,
	}
}

func printGroupsTable(groups []config.ProjectGroup) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPROJECTS\tDESCRIPTION")
	for _, g := range groups {
		desc := utils.TruncateString(g.Description, 30)
		if desc == "" {
			desc = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			utils.TruncateString(g.ID, 8),
			g.Name,
			len(g.ProjectIDs),
			desc,
		)
	}
	w.Flush()
}
