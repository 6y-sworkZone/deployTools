package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"deploytools/pkg/utils"
)

func NewLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Manage log files",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all log files",
		Run:   runListLogs,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show [filename]",
		Short: "Show log file content",
		Args:  cobra.ExactArgs(1),
		Run:   runShowLog,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup old log files",
		Run:   runCleanupLogs,
	})

	cmd.PersistentFlags().Int("keep-days", 30, "Keep logs for this many days")

	return cmd
}

func runListLogs(cmd *cobra.Command, args []string) {
	logDir := utils.DefaultLogger.GetLogDir()

	files, err := filepath.Glob(filepath.Join(logDir, "deploy_*.log"))
	if err != nil {
		fmt.Printf("Failed to list log files: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No log files found")
		return
	}

	sort.Strings(files)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILENAME\tSIZE\tMODIFIED")

	var totalSize int64
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		totalSize += info.Size()
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			filepath.Base(file),
			utils.FormatFileSize(info.Size()),
			info.ModTime().Format("2006-01-02 15:04"),
		)
	}
	w.Flush()

	fmt.Printf("\nTotal: %d files, %s\n", len(files), utils.FormatFileSize(totalSize))
}

func runShowLog(cmd *cobra.Command, args []string) {
	filename := args[0]
	logDir := utils.DefaultLogger.GetLogDir()
	logPath := filepath.Join(logDir, filename)

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Printf("Log file not found: %s\n", filename)
		return
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Printf("Failed to read log file: %v\n", err)
		return
	}

	fmt.Printf("\n=== Log: %s ===\n\n", filename)
	fmt.Println(string(content))
}

func runCleanupLogs(cmd *cobra.Command, args []string) {
	keepDays, _ := cmd.Flags().GetInt("keep-days")

	fmt.Printf("This will delete log files older than %d days.\n", keepDays)
	fmt.Print("Continue? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Cleanup cancelled")
		return
	}

	deletedCount, err := utils.DefaultLogger.CleanupOldLogs(keepDays)
	if err != nil {
		fmt.Printf("Cleanup failed: %v\n", err)
		return
	}

	fmt.Printf("Deleted %d old log file(s)\n", deletedCount)
}
