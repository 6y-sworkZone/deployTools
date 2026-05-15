package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"deploytools/internal/config"
	"deploytools/pkg/ssh"
	"deploytools/pkg/utils"
)

type DeployStatus string

const (
	StatusPending   DeployStatus = "pending"
	StatusRunning   DeployStatus = "running"
	StatusSuccess   DeployStatus = "success"
	StatusFailed    DeployStatus = "failed"
	StatusCancelled DeployStatus = "cancelled"
)

type FileInfo struct {
	RelativePath string
	LocalPath    string
	RemotePath   string
	Size         int64
	NeedsUpdate  bool
	Uploaded     bool
}

type ProgressCallback func(current, total int, message string)

type Deployer struct {
	project        config.Project
	server         config.Server
	client         *ssh.Client
	files          []FileInfo
	historyID      string
	backupPath     string
	configManager  *config.ConfigManager
	progressCb     ProgressCallback
	backupEnabled  bool
}

func NewDeployer(project config.Project, server config.Server, cm *config.ConfigManager, backupEnabled bool) *Deployer {
	return &Deployer{
		project:       project,
		server:        server,
		configManager: cm,
		backupEnabled: backupEnabled,
		files:         []FileInfo{},
	}
}

func (d *Deployer) SetProgressCallback(cb ProgressCallback) {
	d.progressCb = cb
}

func (d *Deployer) reportProgress(current, total int, message string) {
	if d.progressCb != nil {
		d.progressCb(current, total, message)
	}
}

func (d *Deployer) Deploy() error {
	history := config.DeployHistory{
		ProjectID: d.project.ID,
		ServerID:  d.server.ID,
		Status:    string(StatusRunning),
	}
	history, err := d.configManager.AddDeployHistory(history)
	if err != nil {
		utils.Warn("Failed to create deploy history: %v", err)
	}
	d.historyID = history.ID

	startTime := time.Now()
	utils.Info("Starting deployment of project '%s' to server '%s'", d.project.Name, d.server.Name)

	if err := d.connect(); err != nil {
		d.updateHistory(StatusFailed, err.Error())
		return err
	}
	defer d.client.Close()

	if d.backupEnabled {
		d.reportProgress(0, 100, "Creating backup...")
		if err := d.createBackup(); err != nil {
			utils.Warn("Backup failed: %v", err)
		}
	}

	d.reportProgress(0, 100, "Scanning files...")
	if err := d.scanFiles(); err != nil {
		d.updateHistory(StatusFailed, err.Error())
		return err
	}

	if err := d.executeDeploy(); err != nil {
		d.updateHistory(StatusFailed, err.Error())
		return err
	}

	if d.project.PostScript != "" {
		d.reportProgress(90, 100, "Running post-deploy script...")
		if err := d.runPostScript(); err != nil {
			utils.Warn("Post-deploy script failed: %v", err)
		}
	}

	duration := time.Since(startTime)
	utils.Info("Deployment completed successfully in %v", duration)
	d.updateHistory(StatusSuccess, "")
	return nil
}

func (d *Deployer) connect() error {
	d.reportProgress(5, 100, "Connecting to server...")
	client, err := ssh.NewClient(d.server)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	d.client = client
	return nil
}

func (d *Deployer) createBackup() error {
	d.backupPath = fmt.Sprintf("%s/backup_%s", d.configManager.GetConfig().Backup.RemotePath, time.Now().Format("20060102_150405"))
	if err := d.client.MkdirAll(d.backupPath); err != nil {
		return err
	}
	_, err := d.client.RunCommand(fmt.Sprintf("cp -r %s/* %s/ 2>/dev/null || true", d.project.RemotePath, d.backupPath))
	return err
}

func (d *Deployer) scanFiles() error {
	localPath := d.project.LocalPath
	excludeMatcher := utils.GetExcludeMatcher(d.project.ExcludePatterns)

	d.files = nil

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		if excludeMatcher(path) {
			return nil
		}

		remotePath := filepath.Join(d.project.RemotePath, relPath)
		remotePath = strings.ReplaceAll(remotePath, "\\", "/")

		fileInfo := FileInfo{
			RelativePath: relPath,
			LocalPath:    path,
			RemotePath:   remotePath,
			Size:         info.Size(),
			NeedsUpdate:  true,
		}

		if d.project.UpdateMode == config.UpdateModeDiff {
			if d.client.FileExists(remotePath) {
				localMD5, err := utils.GetFileMD5(path)
				if err == nil {
					remoteMD5, err := d.client.GetFileMD5(remotePath)
					if err == nil && localMD5 == remoteMD5 {
						fileInfo.NeedsUpdate = false
					}
				}
			}
		}

		d.files = append(d.files, fileInfo)
		return nil
	})

	return err
}

func (d *Deployer) executeDeploy() error {
	switch d.project.UpdateMode {
	case config.UpdateModeReplaceDir:
		return d.deployReplaceDir()
	case config.UpdateModeReplaceFile:
		return d.deployReplaceFile()
	default:
		return d.deployDiff()
	}
}

func (d *Deployer) deployDiff() error {
	totalFiles := len(d.files)
	updatedFiles := 0

	for i, file := range d.files {
		if !file.NeedsUpdate {
			continue
		}

		progress := 10 + (i * 80 / totalFiles)
		d.reportProgress(progress, 100, fmt.Sprintf("Uploading: %s", utils.TruncateString(file.RelativePath, 40)))

		if err := d.client.UploadFile(file.LocalPath, file.RemotePath, nil); err != nil {
			utils.Error("Failed to upload %s: %v", file.RelativePath, err)
			return err
		}

		d.files[i].Uploaded = true
		updatedFiles++
	}

	d.reportProgress(90, 100, fmt.Sprintf("Updated %d/%d files", updatedFiles, totalFiles))
	utils.Info("Diff deployment: updated %d/%d files", updatedFiles, totalFiles)
	return nil
}

func (d *Deployer) deployReplaceFile() error {
	totalFiles := len(d.files)

	for i, file := range d.files {
		progress := 10 + (i * 80 / totalFiles)
		d.reportProgress(progress, 100, fmt.Sprintf("Uploading: %s", utils.TruncateString(file.RelativePath, 40)))

		if d.client.FileExists(file.RemotePath) {
			if err := d.client.Remove(file.RemotePath); err != nil {
				utils.Warn("Failed to remove existing file: %v", err)
			}
		}

		if err := d.client.UploadFile(file.LocalPath, file.RemotePath, nil); err != nil {
			utils.Error("Failed to upload %s: %v", file.RelativePath, err)
			return err
		}

		d.files[i].Uploaded = true
	}

	d.reportProgress(90, 100, fmt.Sprintf("Replaced %d files", totalFiles))
	utils.Info("Replace file deployment: uploaded %d files", totalFiles)
	return nil
}

func (d *Deployer) deployReplaceDir() error {
	d.reportProgress(10, 100, "Removing remote directory...")

	if d.client.FileExists(d.project.RemotePath) {
		if err := d.client.RemoveDirectory(d.project.RemotePath); err != nil {
			utils.Warn("Failed to remove remote directory: %v", err)
		}
	}

	totalFiles := len(d.files)
	for i, file := range d.files {
		progress := 20 + (i * 70 / totalFiles)
		d.reportProgress(progress, 100, fmt.Sprintf("Uploading: %s", utils.TruncateString(file.RelativePath, 40)))

		if err := d.client.UploadFile(file.LocalPath, file.RemotePath, nil); err != nil {
			utils.Error("Failed to upload %s: %v", file.RelativePath, err)
			return err
		}

		d.files[i].Uploaded = true
	}

	d.reportProgress(90, 100, fmt.Sprintf("Uploaded %d files", totalFiles))
	utils.Info("Replace directory deployment: uploaded %d files", totalFiles)
	return nil
}

func (d *Deployer) runPostScript() error {
	output, err := d.client.RunCommand(d.project.PostScript)
	if err != nil {
		utils.Error("Post-script output: %s", output)
		return err
	}
	utils.Debug("Post-script output: %s", output)
	return nil
}

func (d *Deployer) updateHistory(status DeployStatus, errorMsg string) {
	if d.historyID == "" {
		return
	}

	history := config.DeployHistory{
		EndTime:    time.Now(),
		Status:     string(status),
		ErrorMsg:   errorMsg,
		LogFile:    utils.DefaultLogger.GetLogFilePath(),
		BackupPath: d.backupPath,
		RemotePath: d.project.RemotePath,
		ProjectName: d.project.Name,
	}

	if err := d.configManager.UpdateDeployHistory(d.historyID, history); err != nil {
		utils.Warn("Failed to update deploy history: %v", err)
	}
}

func (d *Deployer) GetFiles() []FileInfo {
	return d.files
}

func (d *Deployer) Rollback(backupPath string) error {
	history := config.DeployHistory{
		ProjectID: d.project.ID,
		ServerID:  d.server.ID,
		Status:    string(StatusRunning),
		IsRollback: true,
	}
	history, err := d.configManager.AddDeployHistory(history)
	if err != nil {
		utils.Warn("Failed to create rollback history: %v", err)
	}
	d.historyID = history.ID

	startTime := time.Now()
	utils.Info("Starting rollback of project '%s' from backup '%s'", d.project.Name, backupPath)

	if err := d.connect(); err != nil {
		d.updateHistory(StatusFailed, err.Error())
		return err
	}
	defer d.client.Close()

	d.reportProgress(10, 100, "Removing current deployment...")
	if err := d.client.RemoveDirectory(d.project.RemotePath); err != nil {
		d.updateHistory(StatusFailed, err.Error())
		return fmt.Errorf("failed to remove current deployment: %w", err)
	}

	d.reportProgress(40, 100, "Restoring from backup...")
	if err := d.client.MkdirAll(d.project.RemotePath); err != nil {
		d.updateHistory(StatusFailed, err.Error())
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	_, err = d.client.RunCommand(fmt.Sprintf("cp -r %s/* %s/", backupPath, d.project.RemotePath))
	if err != nil {
		d.updateHistory(StatusFailed, err.Error())
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	if d.project.PostScript != "" {
		d.reportProgress(90, 100, "Running post-deploy script...")
		if err := d.runPostScript(); err != nil {
			utils.Warn("Post-deploy script failed: %v", err)
		}
	}

	duration := time.Since(startTime)
	utils.Info("Rollback completed successfully in %v", duration)
	d.backupPath = backupPath
	d.updateHistory(StatusSuccess, "")
	return nil
}
