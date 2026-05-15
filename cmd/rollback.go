package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"deploytools/internal/deploy"
)

func NewRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback [history-id]",
		Short: "Rollback to a previous deployment version",
		Args:  cobra.ExactArgs(1),
		Run:   runRollback,
	}

	return cmd
}

func runRollback(cmd *cobra.Command, args []string) {
	historyID := args[0]
	history := cfgManager.GetDeployHistory(historyID)
	if history == nil {
		fmt.Printf("Deployment history not found: %s\n", historyID)
		return
	}

	if history.BackupPath == "" {
		fmt.Println("No backup path available for this deployment")
		return
	}

	project := cfgManager.GetProject(history.ProjectID)
	if project == nil {
		fmt.Printf("Project not found: %s\n", history.ProjectID)
		return
	}

	server := cfgManager.GetServer(project.ServerID)
	if server == nil {
		fmt.Printf("Server not found: %s\n", project.ServerID)
		return
	}

	fmt.Println("\n=== Rollback Preview ===")
	fmt.Printf("Deployment ID:   %s\n", history.ID)
	fmt.Printf("Project:         %s\n", history.ProjectName)
	fmt.Printf("Server:          %s\n", server.Name)
	fmt.Printf("Backup Path:     %s\n", history.BackupPath)
	fmt.Printf("Target Path:     %s\n", history.RemotePath)
	fmt.Printf("Deployed At:     %s\n", history.StartTime.Format("2006-01-02 15:04:05"))

	fmt.Print("\nWARNING: This will overwrite the current deployment with the backup version.\n")
	fmt.Print("Continue with rollback? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Rollback cancelled")
		return
	}

	fmt.Println("\nStarting rollback...")

	deployer := deploy.NewDeployer(*project, *server, cfgManager, false)
	deployer.SetProgressCallback(func(current, total int, message string) {
		percent := current * 100 / total
		fmt.Printf("\r[%d%%] %s", percent, message)
	})

	if err := deployer.Rollback(history.BackupPath); err != nil {
		fmt.Printf("\n\nRollback failed: %v\n", err)
		return
	}

	fmt.Println("\n\nRollback completed successfully!")
}
