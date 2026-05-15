package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type ConfigManager struct {
	config *AppConfig
	viper  *viper.Viper
}

func NewConfigManager(configPath string) (*ConfigManager, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home dir: %w", err)
		}
		configPath = filepath.Join(homeDir, ".deploytools", "config.yaml")
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	cm := &ConfigManager{
		viper: v,
		config: &AppConfig{
			ConfigPath: configPath,
			Backup: BackupConfig{
				Enabled:      true,
				RemotePath:   "/tmp/backup",
				KeepVersions: 5,
			},
		},
	}

	if err := cm.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := cm.Save(); err != nil {
			return nil, err
		}
	}

	return cm, nil
}

func (cm *ConfigManager) Load() error {
	if err := cm.viper.ReadInConfig(); err != nil {
		return err
	}

	if err := cm.viper.Unmarshal(cm.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cm.config.ConfigPath = cm.viper.ConfigFileUsed()
	return nil
}

func (cm *ConfigManager) Save() error {
	cm.viper.Set("servers", cm.config.Servers)
	cm.viper.Set("projects", cm.config.Projects)
	cm.viper.Set("project_groups", cm.config.ProjectGroups)
	cm.viper.Set("history", cm.config.History)
	cm.viper.Set("backup", cm.config.Backup)

	return cm.viper.WriteConfig()
}

func (cm *ConfigManager) GetConfig() *AppConfig {
	return cm.config
}

func (cm *ConfigManager) AddServer(server Server) (Server, error) {
	server.ID = uuid.New().String()
	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()
	cm.config.Servers = append(cm.config.Servers, server)
	return server, cm.Save()
}

func (cm *ConfigManager) UpdateServer(id string, server Server) error {
	for i, s := range cm.config.Servers {
		if s.ID == id {
			server.ID = id
			server.CreatedAt = s.CreatedAt
			server.UpdatedAt = time.Now()
			cm.config.Servers[i] = server
			return cm.Save()
		}
	}
	return fmt.Errorf("server not found: %s", id)
}

func (cm *ConfigManager) DeleteServer(id string) error {
	for i, s := range cm.config.Servers {
		if s.ID == id {
			cm.config.Servers = append(cm.config.Servers[:i], cm.config.Servers[i+1:]...)
			return cm.Save()
		}
	}
	return fmt.Errorf("server not found: %s", id)
}

func (cm *ConfigManager) GetServer(id string) *Server {
	for _, s := range cm.config.Servers {
		if s.ID == id {
			return &s
		}
	}
	return nil
}

func (cm *ConfigManager) GetServers() []Server {
	return cm.config.Servers
}

func (cm *ConfigManager) AddProject(project Project) (Project, error) {
	project.ID = uuid.New().String()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	if project.UpdateMode == "" {
		project.UpdateMode = UpdateModeDiff
	}
	cm.config.Projects = append(cm.config.Projects, project)
	return project, cm.Save()
}

func (cm *ConfigManager) UpdateProject(id string, project Project) error {
	for i, p := range cm.config.Projects {
		if p.ID == id {
			project.ID = id
			project.CreatedAt = p.CreatedAt
			project.UpdatedAt = time.Now()
			cm.config.Projects[i] = project
			return cm.Save()
		}
	}
	return fmt.Errorf("project not found: %s", id)
}

func (cm *ConfigManager) DeleteProject(id string) error {
	for i, p := range cm.config.Projects {
		if p.ID == id {
			cm.config.Projects = append(cm.config.Projects[:i], cm.config.Projects[i+1:]...)
			for gi, g := range cm.config.ProjectGroups {
				newIDs := []string{}
				for _, pid := range g.ProjectIDs {
					if pid != id {
						newIDs = append(newIDs, pid)
					}
				}
				cm.config.ProjectGroups[gi].ProjectIDs = newIDs
			}
			return cm.Save()
		}
	}
	return fmt.Errorf("project not found: %s", id)
}

func (cm *ConfigManager) GetProject(id string) *Project {
	for _, p := range cm.config.Projects {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

func (cm *ConfigManager) GetProjects() []Project {
	return cm.config.Projects
}

func (cm *ConfigManager) AddProjectGroup(group ProjectGroup) (ProjectGroup, error) {
	group.ID = uuid.New().String()
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	cm.config.ProjectGroups = append(cm.config.ProjectGroups, group)
	return group, cm.Save()
}

func (cm *ConfigManager) UpdateProjectGroup(id string, group ProjectGroup) error {
	for i, g := range cm.config.ProjectGroups {
		if g.ID == id {
			group.ID = id
			group.CreatedAt = g.CreatedAt
			group.UpdatedAt = time.Now()
			cm.config.ProjectGroups[i] = group
			return cm.Save()
		}
	}
	return fmt.Errorf("project group not found: %s", id)
}

func (cm *ConfigManager) DeleteProjectGroup(id string) error {
	for i, g := range cm.config.ProjectGroups {
		if g.ID == id {
			cm.config.ProjectGroups = append(cm.config.ProjectGroups[:i], cm.config.ProjectGroups[i+1:]...)
			return cm.Save()
		}
	}
	return fmt.Errorf("project group not found: %s", id)
}

func (cm *ConfigManager) GetProjectGroup(id string) *ProjectGroup {
	for _, g := range cm.config.ProjectGroups {
		if g.ID == id {
			return &g
		}
	}
	return nil
}

func (cm *ConfigManager) GetProjectGroups() []ProjectGroup {
	return cm.config.ProjectGroups
}

func (cm *ConfigManager) AddDeployHistory(history DeployHistory) (DeployHistory, error) {
	history.ID = uuid.New().String()
	history.StartTime = time.Now()
	cm.config.History = append(cm.config.History, history)
	return history, cm.Save()
}

func (cm *ConfigManager) UpdateDeployHistory(id string, history DeployHistory) error {
	for i, h := range cm.config.History {
		if h.ID == id {
			history.ID = id
			cm.config.History[i] = history
			return cm.Save()
		}
	}
	return fmt.Errorf("history not found: %s", id)
}
