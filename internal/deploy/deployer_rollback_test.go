package deploy

import (
	"path/filepath"
	"testing"

	"deploytools/internal/config"
)

func TestDeployerRollback(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := config.NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	project := config.Project{
		Name:       "test-project",
		LocalPath:  tempDir,
		RemotePath: "/test/path",
	}

	server := config.Server{
		Name: "test-server",
		IP:   "127.0.0.1",
	}

	deployer := NewDeployer(project, server, cm, true)
	if deployer == nil {
		t.Fatal("Deployer should not be nil")
	}

	if deployer.project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", deployer.project.Name)
	}

	if deployer.server.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", deployer.server.Name)
	}

	if deployer.backupEnabled != true {
		t.Error("Expected backupEnabled to be true")
	}
}

func TestProgressCallback(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := config.NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	project := config.Project{Name: "test"}
	server := config.Server{Name: "test"}

	deployer := NewDeployer(project, server, cm, true)

	callbackCalled := false
	lastMessage := ""

	callback := func(current, total int, message string) {
		callbackCalled = true
		lastMessage = message
	}

	deployer.SetProgressCallback(callback)

	deployer.reportProgress(50, 100, "Test progress")

	if !callbackCalled {
		t.Error("Progress callback should have been called")
	}

	if lastMessage != "Test progress" {
		t.Errorf("Expected message 'Test progress', got '%s'", lastMessage)
	}
}

func TestBackupPath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := config.NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	project := config.Project{Name: "test"}
	server := config.Server{Name: "test"}

	deployer := NewDeployer(project, server, cm, true)

	if deployer.backupPath != "" {
		t.Error("Backup path should be empty initially")
	}
}
