package config

import (
	"time"
)

type UpdateMode string

const (
	UpdateModeDiff       UpdateMode = "diff"
	UpdateModeReplaceFile UpdateMode = "replace_file"
	UpdateModeReplaceDir UpdateMode = "replace_dir"
)

type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeKey      AuthType = "key"
)

type Server struct {
	ID         string    `json:"id" yaml:"id"`
	Name       string    `json:"name" yaml:"name"`
	IP         string    `json:"ip" yaml:"ip"`
	Port       int       `json:"port" yaml:"port"`
	Username   string    `json:"username" yaml:"username"`
	AuthType   AuthType  `json:"auth_type" yaml:"auth_type"`
	Password   string    `json:"password,omitempty" yaml:"password,omitempty"`
	KeyPath    string    `json:"key_path,omitempty" yaml:"key_path,omitempty"`
	CreatedAt  time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" yaml:"updated_at"`
}

type Project struct {
	ID              string     `json:"id" yaml:"id"`
	Name            string     `json:"name" yaml:"name"`
	LocalPath       string     `json:"local_path" yaml:"local_path"`
	ServerID        string     `json:"server_id" yaml:"server_id"`
	RemotePath      string     `json:"remote_path" yaml:"remote_path"`
	UpdateMode      UpdateMode `json:"update_mode" yaml:"update_mode"`
	PostScript      string     `json:"post_script,omitempty" yaml:"post_script,omitempty"`
	ExcludePatterns []string   `json:"exclude_patterns,omitempty" yaml:"exclude_patterns,omitempty"`
	CreatedAt       time.Time  `json:"created_at" yaml:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" yaml:"updated_at"`
}

type ProjectGroup struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	ProjectIDs  []string  `json:"project_ids" yaml:"project_ids"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
}

type DeployHistory struct {
	ID         string    `json:"id" yaml:"id"`
	GroupID    string    `json:"group_id,omitempty" yaml:"group_id,omitempty"`
	ProjectID  string    `json:"project_id,omitempty" yaml:"project_id,omitempty"`
	ServerID   string    `json:"server_id" yaml:"server_id"`
	Status     string    `json:"status" yaml:"status"`
	StartTime  time.Time `json:"start_time" yaml:"start_time"`
	EndTime    time.Time `json:"end_time,omitempty" yaml:"end_time,omitempty"`
	LogFile    string    `json:"log_file,omitempty" yaml:"log_file,omitempty"`
	ErrorMsg   string    `json:"error_msg,omitempty" yaml:"error_msg,omitempty"`
}

type BackupConfig struct {
	Enabled      bool   `json:"enabled" yaml:"enabled"`
	RemotePath   string `json:"remote_path" yaml:"remote_path"`
	KeepVersions int    `json:"keep_versions" yaml:"keep_versions"`
}

type AppConfig struct {
	Servers      []Server       `json:"servers" yaml:"servers"`
	Projects     []Project      `json:"projects" yaml:"projects"`
	ProjectGroups []ProjectGroup `json:"project_groups" yaml:"project_groups"`
	History      []DeployHistory `json:"history" yaml:"history"`
	Backup       BackupConfig    `json:"backup" yaml:"backup"`
	ConfigPath   string         `json:"-" yaml:"-"`
}
