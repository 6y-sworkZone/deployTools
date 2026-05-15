package deploy

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"deploytools/internal/config"
)

type MockSSHClient struct {
	TestConnectionCalled    bool
	UploadFileCalled        bool
	DownloadFileCalled      bool
	MkdirAllCalled          bool
	FileExistsCalled        bool
	FileExistsResult        bool
	GetFileMD5Called        bool
	GetFileMD5Result        string
	RemoveCalled            bool
	RemoveDirectoryCalled   bool
	RunCommandCalled        bool
	RunCommandResult        string
	ListFilesCalled         bool
	CloseCalled             bool
	UploadShouldFail        bool
	FileExistsShouldChange  bool
	FilesExistCounter       int
}

func (m *MockSSHClient) TestConnection() error {
	m.TestConnectionCalled = true
	return nil
}

func (m *MockSSHClient) UploadFile(localPath, remotePath string, progressFn func(int64, int64)) error {
	m.UploadFileCalled = true
	if m.UploadShouldFail {
		return io.ErrShortWrite
	}
	return nil
}

func (m *MockSSHClient) DownloadFile(remotePath, localPath string, progressFn func(int64, int64)) error {
	m.DownloadFileCalled = true
	return nil
}

func (m *MockSSHClient) MkdirAll(remotePath string) error {
	m.MkdirAllCalled = true
	return nil
}

func (m *MockSSHClient) FileExists(remotePath string) bool {
	m.FileExistsCalled = true
	if m.FileExistsShouldChange {
		m.FilesExistCounter++
		if m.FilesExistCounter == 1 {
			return true
		}
		if m.FilesExistCounter == 2 {
			return false
		}
	}
	return m.FileExistsResult
}

func (m *MockSSHClient) GetFileMD5(remotePath string) (string, error) {
	m.GetFileMD5Called = true
	return m.GetFileMD5Result, nil
}

func (m *MockSSHClient) Remove(remotePath string) error {
	m.RemoveCalled = true
	return nil
}

func (m *MockSSHClient) RemoveDirectory(remotePath string) error {
	m.RemoveDirectoryCalled = true
	return nil
}

func (m *MockSSHClient) RunCommand(cmd string) (string, error) {
	m.RunCommandCalled = true
	return m.RunCommandResult, nil
}

func (m *MockSSHClient) ListFiles(remotePath string) ([]os.FileInfo, error) {
	m.ListFilesCalled = true
	return []os.FileInfo{}, nil
}

func (m *MockSSHClient) Close() error {
	m.CloseCalled = true
	return nil
}

func TestNewDeployer(t *testing.T) {
	project := config.Project{
		Name:       "test-project",
		LocalPath:  "/local/path",
		RemotePath: "/remote/path",
		UpdateMode: config.UpdateModeDiff,
	}

	server := config.Server{
		Name: "test-server",
		IP:   "192.168.1.1",
	}

	cm := &config.ConfigManager{}

	deployer := NewDeployer(project, server, cm, true)
	if deployer == nil {
		t.Fatal("Deployer should not be nil")
	}
}

func TestSetProgressCallback(t *testing.T) {
	project := config.Project{Name: "test"}
	server := config.Server{Name: "test"}
	cm := &config.ConfigManager{}

	deployer := NewDeployer(project, server, cm, true)

	callback := func(current, total int, message string) {
		// Just a placeholder
	}

	deployer.SetProgressCallback(callback)

	if deployer.progressCb == nil {
		t.Error("Progress callback should be set")
	}
}

func TestFileInfoStruct(t *testing.T) {
	fi := FileInfo{
		RelativePath: "test/file.txt",
		LocalPath:    "/local/test/file.txt",
		RemotePath:   "/remote/test/file.txt",
		Size:         1024,
		NeedsUpdate:  true,
		Uploaded:     false,
	}

	if fi.RelativePath != "test/file.txt" {
		t.Error("RelativePath not set correctly")
	}
	if fi.Size != 1024 {
		t.Error("Size not set correctly")
	}
	if !fi.NeedsUpdate {
		t.Error("NeedsUpdate should be true")
	}
}

func TestDeployStatusConstants(t *testing.T) {
	if StatusPending != "pending" {
		t.Error("StatusPending constant incorrect")
	}
	if StatusRunning != "running" {
		t.Error("StatusRunning constant incorrect")
	}
	if StatusSuccess != "success" {
		t.Error("StatusSuccess constant incorrect")
	}
	if StatusFailed != "failed" {
		t.Error("StatusFailed constant incorrect")
	}
	if StatusCancelled != "cancelled" {
		t.Error("StatusCancelled constant incorrect")
	}
}

func TestDeployerGetFiles(t *testing.T) {
	project := config.Project{Name: "test"}
	server := config.Server{Name: "test"}
	cm := &config.ConfigManager{}

	deployer := NewDeployer(project, server, cm, true)

	files := deployer.GetFiles()
	if files == nil {
		t.Error("GetFiles should return a slice, even if empty")
	}
	if len(files) != 0 {
		t.Error("New deployer should have zero files")
	}
}

func TestProgressCallbackInvocation(t *testing.T) {
	project := config.Project{Name: "test"}
	server := config.Server{Name: "test"}
	cm := &config.ConfigManager{}

	deployer := NewDeployer(project, server, cm, true)

	lastMessage := ""
	messageCount := 0

	callback := func(current, total int, message string) {
		lastMessage = message
		messageCount++
	}

	deployer.SetProgressCallback(callback)

	deployer.reportProgress(50, 100, "Test progress")

	if messageCount != 1 {
		t.Errorf("Expected 1 callback invocation, got %d", messageCount)
	}
	if lastMessage != "Test progress" {
		t.Errorf("Expected message 'Test progress', got '%s'", lastMessage)
	}
}

func TestBackupEnabledFlag(t *testing.T) {
	project := config.Project{Name: "test"}
	server := config.Server{Name: "test"}
	cm := &config.ConfigManager{}

	deployerWithBackup := NewDeployer(project, server, cm, true)
	if !deployerWithBackup.backupEnabled {
		t.Error("Backup should be enabled")
	}

	deployerWithoutBackup := NewDeployer(project, server, cm, false)
	if deployerWithoutBackup.backupEnabled {
		t.Error("Backup should be disabled")
	}
}

func TestProjectAndServerStorage(t *testing.T) {
	project := config.Project{
		Name:       "my-project",
		LocalPath:  "/tmp/local",
		RemotePath: "/var/www",
		UpdateMode: config.UpdateModeReplaceDir,
	}

	server := config.Server{
		Name: "my-server",
		IP:   "10.0.0.1",
		Port: 22,
	}

	cm := &config.ConfigManager{}

	deployer := NewDeployer(project, server, cm, true)

	if deployer.project.Name != "my-project" {
		t.Error("Project name not stored correctly")
	}
	if deployer.project.LocalPath != "/tmp/local" {
		t.Error("Local path not stored correctly")
	}
	if deployer.project.RemotePath != "/var/www" {
		t.Error("Remote path not stored correctly")
	}
	if deployer.project.UpdateMode != config.UpdateModeReplaceDir {
		t.Error("Update mode not stored correctly")
	}

	if deployer.server.Name != "my-server" {
		t.Error("Server name not stored correctly")
	}
	if deployer.server.IP != "10.0.0.1" {
		t.Error("Server IP not stored correctly")
	}
}

func TestEmptyLocalPathScanning(t *testing.T) {
	tempDir := t.TempDir()

	project := config.Project{
		Name:       "test-project",
		LocalPath:  tempDir,
		RemotePath: "/remote/path",
		UpdateMode: config.UpdateModeDiff,
	}

	server := config.Server{Name: "test-server"}

	tempConfig := filepath.Join(tempDir, "config.yaml")
	cm, _ := config.NewConfigManager(tempConfig)

	deployer := NewDeployer(project, server, cm, false)

	files := deployer.GetFiles()
	if len(files) != 0 {
		t.Errorf("Expected 0 files for empty directory, got %d", len(files))
	}
}

func TestExcludePatterns(t *testing.T) {
	tempDir := t.TempDir()

	os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test.log"), []byte("log content"), 0644)

	project := config.Project{
		Name:            "test-project",
		LocalPath:       tempDir,
		RemotePath:      "/remote/path",
		UpdateMode:      config.UpdateModeDiff,
		ExcludePatterns: []string{"*.log"},
	}

	server := config.Server{Name: "test-server"}
	tempConfig := filepath.Join(tempDir, "config.yaml")
	cm, _ := config.NewConfigManager(tempConfig)

	deployer := NewDeployer(project, server, cm, false)

	files := deployer.GetFiles()
	for _, f := range files {
		if f.RelativePath == "test.log" {
			t.Error("test.log should have been excluded")
		}
	}
}

func TestPostScriptConfig(t *testing.T) {
	project := config.Project{
		Name:       "test-project",
		PostScript: "echo 'deploy complete'",
	}

	server := config.Server{Name: "test-server"}
	cm := &config.ConfigManager{}

	deployer := NewDeployer(project, server, cm, false)

	if deployer.project.PostScript != "echo 'deploy complete'" {
		t.Error("Post script not stored correctly")
	}
}
