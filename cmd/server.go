package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"deploytools/internal/config"
	"deploytools/pkg/ssh"
	"deploytools/pkg/utils"
)

func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Manage server configurations",
		Long:  `Add, edit, delete, and list server configurations.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a new server",
		Run:   runAddServer,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "edit [id]",
		Short: "Edit an existing server",
		Args:  cobra.ExactArgs(1),
		Run:   runEditServer,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a server",
		Args:  cobra.ExactArgs(1),
		Run:   runDeleteServer,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "test [id]",
		Short: "Test server connection",
		Args:  cobra.ExactArgs(1),
		Run:   runTestServer,
	})

	return cmd
}

func runAddServer(cmd *cobra.Command, args []string) {
	server := promptServerInfo(config.Server{
		Port: 22,
	})

	server, err := cfgManager.AddServer(server)
	if err != nil {
		fmt.Printf("Failed to add server: %v\n", err)
		return
	}

	fmt.Printf("Server '%s' added successfully (ID: %s)\n", server.Name, server.ID)
}

func runEditServer(cmd *cobra.Command, args []string) {
	id := args[0]
	server := cfgManager.GetServer(id)
	if server == nil {
		fmt.Printf("Server not found: %s\n", id)
		return
	}

	updated := promptServerInfo(*server)
	updated.ID = id

	if err := cfgManager.UpdateServer(id, updated); err != nil {
		fmt.Printf("Failed to update server: %v\n", err)
		return
	}

	fmt.Printf("Server '%s' updated successfully\n", updated.Name)
}

func runDeleteServer(cmd *cobra.Command, args []string) {
	id := args[0]
	server := cfgManager.GetServer(id)
	if server == nil {
		fmt.Printf("Server not found: %s\n", id)
		return
	}

	fmt.Printf("Are you sure you want to delete server '%s'? (y/N): ", server.Name)
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Delete cancelled")
		return
	}

	if err := cfgManager.DeleteServer(id); err != nil {
		fmt.Printf("Failed to delete server: %v\n", err)
		return
	}

	fmt.Printf("Server '%s' deleted successfully\n", server.Name)
}

func runTestServer(cmd *cobra.Command, args []string) {
	id := args[0]
	server := cfgManager.GetServer(id)
	if server == nil {
		fmt.Printf("Server not found: %s\n", id)
		return
	}

	fmt.Printf("Testing connection to server '%s'...\n", server.Name)

	client, err := ssh.NewClient(*server)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		return
	}
	defer client.Close()

	if err := client.TestConnection(); err != nil {
		fmt.Printf("Connection test failed: %v\n", err)
		return
	}

	fmt.Println("Connection successful!")
}

func promptServerInfo(defaults config.Server) config.Server {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Server Name [%s]: ", defaults.Name)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaults.Name
	}

	fmt.Printf("IP Address [%s]: ", defaults.IP)
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = defaults.IP
	}

	fmt.Printf("Port [%d]: ", defaults.Port)
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	port := defaults.Port
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	fmt.Printf("Username [%s]: ", defaults.Username)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = defaults.Username
	}

	authType := defaults.AuthType
	if authType == "" {
		authType = config.AuthTypePassword
	}
	fmt.Printf("Authentication type (password/key) [%s]: ", authType)
	authTypeStr, _ := reader.ReadString('\n')
	authTypeStr = strings.TrimSpace(strings.ToLower(authTypeStr))
	if authTypeStr == "key" {
		authType = config.AuthTypeKey
	} else if authTypeStr == "password" {
		authType = config.AuthTypePassword
	}

	var password, keyPath string
	if authType == config.AuthTypeKey {
		defaultKey := defaults.KeyPath
		if defaultKey == "" {
			homeDir, _ := os.UserHomeDir()
			defaultKey = fmt.Sprintf("%s/.ssh/id_rsa", homeDir)
		}
		fmt.Printf("Private Key Path [%s]: ", defaultKey)
		keyPath, _ = reader.ReadString('\n')
		keyPath = strings.TrimSpace(keyPath)
		if keyPath == "" {
			keyPath = defaultKey
		}
	} else {
		fmt.Printf("Password [%s]: ", strings.Repeat("*", len(defaults.Password)))
		password, _ = reader.ReadString('\n')
		password = strings.TrimSpace(password)
		if password == "" {
			password = defaults.Password
		}
	}

	return config.Server{
		Name:     name,
		IP:       ip,
		Port:     port,
		Username: username,
		AuthType: authType,
		Password: password,
		KeyPath:  keyPath,
	}
}

func printServersTable(servers []config.Server) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tIP\tPORT\tUSERNAME\tAUTH TYPE")
	for _, s := range servers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			utils.TruncateString(s.ID, 8),
			s.Name,
			s.IP,
			s.Port,
			s.Username,
			s.AuthType,
		)
	}
	w.Flush()
}
