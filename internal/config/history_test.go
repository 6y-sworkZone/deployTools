package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDeployHistoryCRUD(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	history := DeployHistory{
		ProjectID: "project-1",
		ServerID:  "server-1",
		Status:    "success",
	}

	created, err := cm.AddDeployHistory(history)
	if err != nil {
		t.Fatalf("Failed to add history: %v", err)
	}

	if created.ID == "" {
		t.Error("History ID should not be empty")
	}

	allHistory := cm.GetAllDeployHistory()
	if len(allHistory) != 1 {
		t.Errorf("Expected 1 history record, got %d", len(allHistory))
	}

	retrieved := cm.GetDeployHistory(created.ID)
	if retrieved == nil {
		t.Fatal("Failed to retrieve history")
	}
	if retrieved.ProjectID != "project-1" {
		t.Errorf("Expected ProjectID 'project-1', got '%s'", retrieved.ProjectID)
	}

	updated := *retrieved
	updated.Status = "failed"
	updated.EndTime = time.Now()
	err = cm.UpdateDeployHistory(created.ID, updated)
	if err != nil {
		t.Fatalf("Failed to update history: %v", err)
	}

	retrieved = cm.GetDeployHistory(created.ID)
	if retrieved.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", retrieved.Status)
	}
}

func TestGetDeployHistoryByProject(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	_, _ = cm.AddDeployHistory(DeployHistory{
		ProjectID: "project-a",
		Status:    "success",
	})
	_, _ = cm.AddDeployHistory(DeployHistory{
		ProjectID: "project-a",
		Status:    "success",
	})
	_, _ = cm.AddDeployHistory(DeployHistory{
		ProjectID: "project-b",
		Status:    "success",
	})

	projectAHistory := cm.GetDeployHistoryByProject("project-a")
	if len(projectAHistory) != 2 {
		t.Errorf("Expected 2 history records for project-a, got %d", len(projectAHistory))
	}

	projectBHistory := cm.GetDeployHistoryByProject("project-b")
	if len(projectBHistory) != 1 {
		t.Errorf("Expected 1 history record for project-b, got %d", len(projectBHistory))
	}
}

func TestGetDeployHistoryByServer(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	_, _ = cm.AddDeployHistory(DeployHistory{
		ServerID: "server-1",
		Status:   "success",
	})
	_, _ = cm.AddDeployHistory(DeployHistory{
		ServerID: "server-2",
		Status:   "failed",
	})

	server1History := cm.GetDeployHistoryByServer("server-1")
	if len(server1History) != 1 {
		t.Errorf("Expected 1 history record for server-1, got %d", len(server1History))
	}
}

func TestGetDeployHistoryByTimeRange(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	now := time.Now()

	h1, _ := cm.AddDeployHistory(DeployHistory{Status: "success"})
	h1.StartTime = now.AddDate(0, 0, -5)
	cm.UpdateDeployHistory(h1.ID, h1)

	h2, _ := cm.AddDeployHistory(DeployHistory{Status: "success"})
	h2.StartTime = now.AddDate(0, 0, -2)
	cm.UpdateDeployHistory(h2.ID, h2)

	h3, _ := cm.AddDeployHistory(DeployHistory{Status: "success"})
	h3.StartTime = now.AddDate(0, 0, -1)
	cm.UpdateDeployHistory(h3.ID, h3)

	start := now.AddDate(0, 0, -3)
	end := now.AddDate(0, 0, 0)

	history := cm.GetDeployHistoryByTimeRange(start, end)
	if len(history) != 2 {
		t.Errorf("Expected 2 history records in time range, got %d", len(history))
	}
}

func TestDeployHistoryWithBackup(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	history := DeployHistory{
		ProjectID:   "project-1",
		ServerID:    "server-1",
		Status:      "success",
		BackupPath:  "/backup/path",
		RemotePath:  "/remote/path",
		ProjectName: "Test Project",
		IsRollback:  true,
	}

	created, err := cm.AddDeployHistory(history)
	if err != nil {
		t.Fatalf("Failed to add history: %v", err)
	}

	retrieved := cm.GetDeployHistory(created.ID)
	if retrieved.BackupPath != "/backup/path" {
		t.Errorf("Expected BackupPath '/backup/path', got '%s'", retrieved.BackupPath)
	}
	if retrieved.RemotePath != "/remote/path" {
		t.Errorf("Expected RemotePath '/remote/path', got '%s'", retrieved.RemotePath)
	}
	if retrieved.ProjectName != "Test Project" {
		t.Errorf("Expected ProjectName 'Test Project', got '%s'", retrieved.ProjectName)
	}
	if !retrieved.IsRollback {
		t.Error("Expected IsRollback to be true")
	}
}

func TestGetDeployHistoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	history := cm.GetDeployHistory("non-existent-id")
	if history != nil {
		t.Error("Expected nil for non-existent history")
	}
}
