package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"deploytools/internal/config"
	"deploytools/pkg/utils"
)

func NewHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Manage deployment history",
		Long:  "View and filter deployment history records.",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List deployment history",
		Run:   runListHistory,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show [id]",
		Short: "Show detailed information for a specific deployment",
		Args:  cobra.ExactArgs(1),
		Run:   runShowHistory,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "logs [id]",
		Short: "Show log content for a specific deployment",
		Args:  cobra.ExactArgs(1),
		Run:   runShowHistoryLogs,
	})

	cmd.PersistentFlags().String("project", "", "Filter by project ID")
	cmd.PersistentFlags().String("server", "", "Filter by server ID")
	cmd.PersistentFlags().String("start", "", "Start time filter (YYYY-MM-DD)")
	cmd.PersistentFlags().String("end", "", "End time filter (YYYY-MM-DD)")
	cmd.PersistentFlags().Int("limit", 50, "Limit number of results")

	return cmd
}

func runListHistory(cmd *cobra.Command, args []string) {
	projectID, _ := cmd.Flags().GetString("project")
	serverID, _ := cmd.Flags().GetString("server")
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	limit, _ := cmd.Flags().GetInt("limit")

	var history []config.DeployHistory

	if projectID != "" {
		history = cfgManager.GetDeployHistoryByProject(projectID)
	} else if serverID != "" {
		history = cfgManager.GetDeployHistoryByServer(serverID)
	} else if startStr != "" || endStr != "" {
		var start, end time.Time
		var err error

		if startStr != "" {
			start, err = time.Parse("2006-01-02", startStr)
			if err != nil {
				fmt.Printf("Invalid start time format: %v\n", err)
				return
			}
		} else {
			start = time.Time{}
		}

		if endStr != "" {
			end, err = time.Parse("2006-01-02", endStr)
			if err != nil {
				fmt.Printf("Invalid end time format: %v\n", err)
				return
			}
			end = end.Add(24 * time.Hour)
		} else {
			end = time.Now()
		}

		history = cfgManager.GetDeployHistoryByTimeRange(start, end)
	} else {
		history = cfgManager.GetAllDeployHistory()
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].StartTime.After(history[j].StartTime)
	})

	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	if len(history) == 0 {
		fmt.Println("No deployment history found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPROJECT\tSTATUS\tSTART TIME\tDURATION\tROLLBACK")
	for _, h := range history {
		duration := "-"
		if !h.EndTime.IsZero() {
			duration = utils.FormatDuration(h.EndTime.Sub(h.StartTime))
		}
		rollback := "No"
		if h.IsRollback {
			rollback = "Yes"
		}
		projectName := h.ProjectName
		if projectName == "" {
			projectName = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			utils.TruncateString(h.ID, 8),
			utils.TruncateString(projectName, 15),
			h.Status,
			h.StartTime.Format("2006-01-02 15:04"),
			duration,
			rollback,
		)
	}
	w.Flush()
}

func runShowHistory(cmd *cobra.Command, args []string) {
	id := args[0]
	history := cfgManager.GetDeployHistory(id)
	if history == nil {
		fmt.Printf("Deployment history not found: %s\n", id)
		return
	}

	fmt.Println("\n=== Deployment Details ===")
	fmt.Printf("ID:              %s\n", history.ID)
	fmt.Printf("Project:         %s\n", history.ProjectName)
	fmt.Printf("Project ID:      %s\n", history.ProjectID)
	fmt.Printf("Server ID:       %s\n", history.ServerID)
	fmt.Printf("Status:          %s\n", history.Status)
	fmt.Printf("Start Time:      %s\n", history.StartTime.Format("2006-01-02 15:04:05"))
	if !history.EndTime.IsZero() {
		fmt.Printf("End Time:        %s\n", history.EndTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Duration:        %s\n", utils.FormatDuration(history.EndTime.Sub(history.StartTime)))
	}
	fmt.Printf("Is Rollback:     %v\n", history.IsRollback)
	if history.BackupPath != "" {
		fmt.Printf("Backup Path:     %s\n", history.BackupPath)
	}
	if history.RemotePath != "" {
		fmt.Printf("Remote Path:     %s\n", history.RemotePath)
	}
	if history.LogFile != "" {
		fmt.Printf("Log File:        %s\n", history.LogFile)
	}
	if history.ErrorMsg != "" {
		fmt.Printf("Error Message:   %s\n", history.ErrorMsg)
	}
	fmt.Println()
}

func runShowHistoryLogs(cmd *cobra.Command, args []string) {
	id := args[0]
	history := cfgManager.GetDeployHistory(id)
	if history == nil {
		fmt.Printf("Deployment history not found: %s\n", id)
		return
	}

	if history.LogFile == "" {
		fmt.Println("No log file associated with this deployment")
		return
	}

	if _, err := os.Stat(history.LogFile); os.IsNotExist(err) {
		fmt.Printf("Log file not found: %s\n", history.LogFile)
		return
	}

	content, err := os.ReadFile(history.LogFile)
	if err != nil {
		fmt.Printf("Failed to read log file: %v\n", err)
		return
	}

	fmt.Printf("\n=== Log Content: %s ===\n\n", history.LogFile)
	fmt.Println(string(content))
}
