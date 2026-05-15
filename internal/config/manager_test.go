package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewConfigManager(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer os.Remove(configPath)

	if cm == nil {
		t.Fatal("ConfigManager is nil")
	}

	if cm.GetConfig() == nil {
		t.Fatal("Config is nil")
	}
}

func TestServerCRUD(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer os.Remove(configPath)

	server := Server{
		Name:     "test-server",
		IP:       "192.168.1.1",
		Port:     22,
		Username: "testuser",
		AuthType: AuthTypePassword,
		Password: "testpass",
	}

	created, err := cm.AddServer(server)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	if created.ID == "" {
		t.Error("Server ID is empty")
	}
	if created.Name != server.Name {
		t.Errorf("Expected name %s, got %s", server.Name, created.Name)
	}
	if created.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}

	servers := cm.GetServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	retrieved := cm.GetServer(created.ID)
	if retrieved == nil {
		t.Fatal("Failed to retrieve server")
	}
	if retrieved.Name != server.Name {
		t.Errorf("Expected name %s, got %s", server.Name, retrieved.Name)
	}

	updatedServer := *retrieved
	updatedServer.Name = "updated-server"
	updatedServer.Port = 2222

	err = cm.UpdateServer(created.ID, updatedServer)
	if err != nil {
		t.Fatalf("Failed to update server: %v", err)
	}

	retrieved = cm.GetServer(created.ID)
	if retrieved.Name != "updated-server" {
		t.Errorf("Expected name 'updated-server', got %s", retrieved.Name)
	}
	if retrieved.Port != 2222 {
		t.Errorf("Expected port 2222, got %d", retrieved.Port)
	}

	err = cm.DeleteServer(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete server: %v", err)
	}

	servers = cm.GetServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers after delete, got %d", len(servers))
	}

	retrieved = cm.GetServer(created.ID)
	if retrieved != nil {
		t.Error("Server should be deleted")
	}
}

func TestProjectCRUD(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer os.Remove(configPath)

	project := Project{
		Name:        "test-project",
		LocalPath:   "/local/path",
		ServerID:    "server-123",
		RemotePath:  "/remote/path",
		UpdateMode:  UpdateModeDiff,
		PostScript:  "echo done",
	}

	created, err := cm.AddProject(project)
	if err != nil {
		t.Fatalf("Failed to add project: %v", err)
	}

	if created.ID == "" {
		t.Error("Project ID is empty")
	}
	if created.Name != project.Name {
		t.Errorf("Expected name %s, got %s", project.Name, created.Name)
	}

	projects := cm.GetProjects()
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}

	retrieved := cm.GetProject(created.ID)
	if retrieved == nil {
		t.Fatal("Failed to retrieve project")
	}

	updatedProject := *retrieved
	updatedProject.Name = "updated-project"
	updatedProject.UpdateMode = UpdateModeReplaceDir

	err = cm.UpdateProject(created.ID, updatedProject)
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	retrieved = cm.GetProject(created.ID)
	if retrieved.Name != "updated-project" {
		t.Errorf("Expected name 'updated-project', got %s", retrieved.Name)
	}
	if retrieved.UpdateMode != UpdateModeReplaceDir {
		t.Errorf("Expected UpdateModeReplaceDir, got %s", retrieved.UpdateMode)
	}

	err = cm.DeleteProject(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	projects = cm.GetProjects()
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects after delete, got %d", len(projects))
	}
}

func TestProjectGroupCRUD(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer os.Remove(configPath)

	project1, _ := cm.AddProject(Project{Name: "project1"})
	project2, _ := cm.AddProject(Project{Name: "project2"})

	group := ProjectGroup{
		Name:        "test-group",
		Description: "Test description",
		ProjectIDs:  []string{project1.ID, project2.ID},
	}

	created, err := cm.AddProjectGroup(group)
	if err != nil {
		t.Fatalf("Failed to add project group: %v", err)
	}

	if created.ID == "" {
		t.Error("Group ID is empty")
	}
	if len(created.ProjectIDs) != 2 {
		t.Errorf("Expected 2 project IDs, got %d", len(created.ProjectIDs))
	}

	groups := cm.GetProjectGroups()
	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}

	retrieved := cm.GetProjectGroup(created.ID)
	if retrieved == nil {
		t.Fatal("Failed to retrieve group")
	}

	updatedGroup := *retrieved
	updatedGroup.Name = "updated-group"
	updatedGroup.ProjectIDs = []string{project1.ID}

	err = cm.UpdateProjectGroup(created.ID, updatedGroup)
	if err != nil {
		t.Fatalf("Failed to update group: %v", err)
	}

	retrieved = cm.GetProjectGroup(created.ID)
	if retrieved.Name != "updated-group" {
		t.Errorf("Expected name 'updated-group', got %s", retrieved.Name)
	}
	if len(retrieved.ProjectIDs) != 1 {
		t.Errorf("Expected 1 project ID, got %d", len(retrieved.ProjectIDs))
	}

	err = cm.DeleteProjectGroup(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete group: %v", err)
	}

	groups = cm.GetProjectGroups()
	if len(groups) != 0 {
		t.Errorf("Expected 0 groups after delete, got %d", len(groups))
	}
}

func TestDeployHistory(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer os.Remove(configPath)

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
		t.Error("History ID is empty")
	}
	if created.StartTime.IsZero() {
		t.Error("StartTime is zero")
	}

	if len(cm.GetConfig().History) != 1 {
		t.Errorf("Expected 1 history record, got %d", len(cm.GetConfig().History))
	}

	updated := created
	updated.EndTime = time.Now()
	updated.LogFile = "/path/to/log"

	err = cm.UpdateDeployHistory(created.ID, updated)
	if err != nil {
		t.Fatalf("Failed to update history: %v", err)
	}
}

func TestDeleteProjectRemovesFromGroups(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer os.Remove(configPath)

	project1, _ := cm.AddProject(Project{Name: "project1"})
	project2, _ := cm.AddProject(Project{Name: "project2"})

	group := ProjectGroup{
		Name:       "test-group",
		ProjectIDs: []string{project1.ID, project2.ID},
	}

	createdGroup, _ := cm.AddProjectGroup(group)

	err = cm.DeleteProject(project1.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	groupAfterDelete := cm.GetProjectGroup(createdGroup.ID)
	if len(groupAfterDelete.ProjectIDs) != 1 {
		t.Errorf("Expected 1 project ID in group after delete, got %d", len(groupAfterDelete.ProjectIDs))
	}
	if groupAfterDelete.ProjectIDs[0] != project2.ID {
		t.Error("Wrong project ID remaining in group")
	}
}

func TestServerNotFound(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer os.Remove(configPath)

	err = cm.UpdateServer("non-existent", Server{})
	if err == nil {
		t.Error("Expected error for updating non-existent server")
	}

	err = cm.DeleteServer("non-existent")
	if err == nil {
		t.Error("Expected error for deleting non-existent server")
	}

	server := cm.GetServer("non-existent")
	if server != nil {
		t.Error("Expected nil for non-existent server")
	}
}

func TestDefaultUpdateMode(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer os.Remove(configPath)

	project := Project{
		Name: "test-project",
	}

	created, err := cm.AddProject(project)
	if err != nil {
		t.Fatalf("Failed to add project: %v", err)
	}

	if created.UpdateMode != UpdateModeDiff {
		t.Errorf("Expected default UpdateModeDiff, got %s", created.UpdateMode)
	}
}
