package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"deploytools/internal/config"
	"deploytools/internal/deploy"
)

var (
	deployYes bool
)

func NewDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy projects or groups",
		Long:  `Deploy individual projects or entire project groups to remote servers.`,
	}

	cmd.PersistentFlags().BoolVarP(&deployYes, "yes", "y", false, "Skip confirmation prompt")

	cmd.AddCommand(&cobra.Command{
		Use:   "project [id]",
		Short: "Deploy a single project",
		Args:  cobra.ExactArgs(1),
		Run:   runDeployProject,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "group [id]",
		Short: "Deploy all projects in a group",
		Args:  cobra.ExactArgs(1),
		Run:   runDeployGroup,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "interactive",
		Short: "Interactive deployment mode",
		Run:   runDeployInteractive,
	})

	return cmd
}

func runDeployProject(cmd *cobra.Command, args []string) {
	id := args[0]
	project := cfgManager.GetProject(id)
	if project == nil {
		fmt.Printf("Project not found: %s\n", id)
		return
	}

	server := cfgManager.GetServer(project.ServerID)
	if server == nil {
		fmt.Printf("Server not found for project: %s\n", project.ServerID)
		return
	}

	if !deployYes {
		fmt.Println("\nDeployment Preview:")
		fmt.Printf("  Project:  %s\n", project.Name)
		fmt.Printf("  Server:   %s (%s)\n", server.Name, server.IP)
		fmt.Printf("  From:     %s\n", project.LocalPath)
		fmt.Printf("  To:       %s\n", project.RemotePath)
		fmt.Printf("  Mode:     %s\n", project.UpdateMode)
		fmt.Printf("  Backup:   %v\n", !noBackup)
		fmt.Print("\nContinue? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		if confirm != "y" && confirm != "yes" {
			fmt.Println("Deployment cancelled")
			return
		}
	}

	fmt.Printf("\nStarting deployment of '%s'...\n", project.Name)

	deployer := deploy.NewDeployer(*project, *server, cfgManager, !noBackup)
	deployer.SetProgressCallback(func(current, total int, message string) {
		percent := current * 100 / total
		fmt.Printf("\r[%d%%] %s", percent, message)
	})

	startTime := time.Now()
	if err := deployer.Deploy(); err != nil {
		fmt.Printf("\n\nDeployment failed: %v\n", err)
		return
	}

	duration := time.Since(startTime)
	fmt.Printf("\n\nDeployment completed successfully in %v\n", duration.Round(time.Second))
}

func runDeployGroup(cmd *cobra.Command, args []string) {
	id := args[0]
	group := cfgManager.GetProjectGroup(id)
	if group == nil {
		fmt.Printf("Group not found: %s\n", id)
		return
	}

	var projects []config.Project
	var servers []config.Server
	for _, pid := range group.ProjectIDs {
		if p := cfgManager.GetProject(pid); p != nil {
			if s := cfgManager.GetServer(p.ServerID); s != nil {
				projects = append(projects, *p)
				servers = append(servers, *s)
			}
		}
	}

	if len(projects) == 0 {
		fmt.Println("No valid projects found in this group")
		return
	}

	if !deployYes {
		fmt.Println("\nGroup Deployment Preview:")
		fmt.Printf("  Group:    %s\n", group.Name)
		fmt.Printf("  Projects: %d\n", len(projects))
		fmt.Printf("  Backup:   %v\n", !noBackup)
		fmt.Println("\nProjects to deploy:")
		for i, p := range projects {
			fmt.Printf("  %d. %s -> %s\n", i+1, p.Name, servers[i].Name)
		}
		fmt.Print("\nContinue? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		if confirm != "y" && confirm != "yes" {
			fmt.Println("Deployment cancelled")
			return
		}
	}

	fmt.Printf("\nStarting group deployment of '%s' (%d projects)...\n", group.Name, len(projects))

	startTime := time.Now()
	successCount := 0

	for i, p := range projects {
		fmt.Printf("\n[%d/%d] Deploying '%s'...\n", i+1, len(projects), p.Name)

		deployer := deploy.NewDeployer(p, servers[i], cfgManager, !noBackup)
		deployer.SetProgressCallback(func(current, total int, message string) {
			percent := current * 100 / total
			fmt.Printf("\r  [%d%%] %s", percent, message)
		})

		if err := deployer.Deploy(); err != nil {
			fmt.Printf("\n  Error: %v\n", err)
		} else {
			fmt.Println("\n  Completed successfully")
			successCount++
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n\nGroup deployment completed: %d/%d succeeded in %v\n",
		successCount, len(projects), duration.Round(time.Second))
}

func runDeployInteractive(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	groups := cfgManager.GetProjectGroups()
	projects := cfgManager.GetProjects()

	fmt.Println("\n=== Interactive Deployment ===")
	fmt.Println("\n1. Deploy a project group")
	fmt.Println("2. Deploy a single project")
	fmt.Print("\nChoose option (1/2): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "1" {
		if len(groups) == 0 {
			fmt.Println("No project groups available. Create one first with 'group add'")
			return
		}

		fmt.Println("\nAvailable groups:")
		for i, g := range groups {
			fmt.Printf("  %d. %s (%d projects)\n", i+1, g.Name, len(g.ProjectIDs))
		}

		fmt.Print("\nEnter group number: ")
		numStr, _ := reader.ReadString('\n')
		numStr = strings.TrimSpace(numStr)
		var num int
		if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil || num < 1 || num > len(groups) {
			fmt.Println("Invalid selection")
			return
		}

		runDeployGroup(cmd, []string{groups[num-1].ID})
	} else if choice == "2" {
		if len(projects) == 0 {
			fmt.Println("No projects available. Create one first with 'project add'")
			return
		}

		fmt.Println("\nAvailable projects:")
		for i, p := range projects {
			serverName := p.ServerID
			if s := cfgManager.GetServer(p.ServerID); s != nil {
				serverName = s.Name
			}
			fmt.Printf("  %d. %s -> %s\n", i+1, p.Name, serverName)
		}

		fmt.Print("\nEnter project number: ")
		numStr, _ := reader.ReadString('\n')
		numStr = strings.TrimSpace(numStr)
		var num int
		if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil || num < 1 || num > len(projects) {
			fmt.Println("Invalid selection")
			return
		}

		runDeployProject(cmd, []string{projects[num-1].ID})
	} else {
		fmt.Println("Invalid option")
	}
}
