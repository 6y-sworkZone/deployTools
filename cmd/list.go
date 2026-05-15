package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		Long:  `List servers, projects, and project groups.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "servers",
		Short: "List all servers",
		Run:   runListServers,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "projects",
		Short: "List all projects",
		Run:   runListProjects,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "groups",
		Short: "List all project groups",
		Run:   runListGroups,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "List all resources",
		Run:   runListAll,
	})

	return cmd
}

func runListServers(cmd *cobra.Command, args []string) {
	servers := cfgManager.GetServers()
	if len(servers) == 0 {
		fmt.Println("No servers configured")
		return
	}

	fmt.Printf("\nServers (%d):\n", len(servers))
	printServersTable(servers)
	fmt.Println()
}

func runListProjects(cmd *cobra.Command, args []string) {
	projects := cfgManager.GetProjects()
	if len(projects) == 0 {
		fmt.Println("No projects configured")
		return
	}

	fmt.Printf("\nProjects (%d):\n", len(projects))
	printProjectsTable(projects)
	fmt.Println()
}

func runListGroups(cmd *cobra.Command, args []string) {
	groups := cfgManager.GetProjectGroups()
	if len(groups) == 0 {
		fmt.Println("No project groups configured")
		return
	}

	fmt.Printf("\nProject Groups (%d):\n", len(groups))
	printGroupsTable(groups)
	fmt.Println()
}

func runListAll(cmd *cobra.Command, args []string) {
	runListServers(cmd, args)
	runListProjects(cmd, args)
	runListGroups(cmd, args)
}
