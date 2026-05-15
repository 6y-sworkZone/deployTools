package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"deploytools/internal/config"
)

func TestNewClientConnectionFailure(t *testing.T) {
	server := config.Server{
		Name:     "test-server",
		IP:       "127.0.0.1",
		Port:     1,
		Username: "testuser",
		AuthType: config.AuthTypePassword,
		Password: "testpass",
	}

	_, err := NewClient(server)
	if err == nil {
		t.Error("Expected error for connection to invalid port")
	}
}

func TestGetKeyAuth(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "id_rsa")

	invalidKey := []byte("not a valid key")
	err := os.WriteFile(keyPath, invalidKey, 0600)
	if err != nil {
		t.Fatalf("Failed to write test key: %v", err)
	}

	_, err = getKeyAuth(keyPath)
	if err == nil {
		t.Error("Expected error for invalid key file")
	}
}

func TestGetKeyAuthNonExistentFile(t *testing.T) {
	_, err := getKeyAuth("/nonexistent/key")
	if err == nil {
		t.Error("Expected error for non-existent key file")
	}
}

func TestServerAuthTypePassword(t *testing.T) {
	server := config.Server{
		Name:     "test-server",
		IP:       "192.168.1.1",
		Port:     22,
		Username: "testuser",
		AuthType: config.AuthTypePassword,
		Password: "testpass",
	}

	if server.AuthType != config.AuthTypePassword {
		t.Errorf("Expected AuthTypePassword, got %s", server.AuthType)
	}
}

func TestServerAuthTypeKey(t *testing.T) {
	server := config.Server{
		Name:     "test-server",
		IP:       "192.168.1.1",
		Port:     22,
		Username: "testuser",
		AuthType: config.AuthTypeKey,
		KeyPath:  "/home/user/.ssh/id_rsa",
	}

	if server.AuthType != config.AuthTypeKey {
		t.Errorf("Expected AuthTypeKey, got %s", server.AuthType)
	}
}

func TestClientCloseWithoutConnection(t *testing.T) {
	client := &Client{}
	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error when closing without connection, got: %v", err)
	}
}

func TestClientStructInitialization(t *testing.T) {
	client := &Client{}
	if client.sshClient != nil {
		t.Error("sshClient should be nil initially")
	}
	if client.sftpClient != nil {
		t.Error("sftpClient should be nil initially")
	}
}

func TestPasswordAuthSelection(t *testing.T) {
	server := config.Server{
		Name:     "test-server",
		IP:       "192.168.1.100",
		Port:     22,
		Username: "deploy",
		AuthType: config.AuthTypePassword,
		Password: "securepassword",
	}

	if server.AuthType == config.AuthTypePassword {
		if server.Password == "" {
			t.Error("Password should not be empty for password auth type")
		}
	} else {
		t.Error("AuthType should be password")
	}
}

func TestKeyAuthSelection(t *testing.T) {
	server := config.Server{
		Name:     "test-server",
		IP:       "192.168.1.100",
		Port:     22,
		Username: "deploy",
		AuthType: config.AuthTypeKey,
		KeyPath:  "/home/deploy/.ssh/id_rsa",
	}

	if server.AuthType == config.AuthTypeKey {
		if server.KeyPath == "" {
			t.Error("KeyPath should not be empty for key auth type")
		}
	} else {
		t.Error("AuthType should be key")
	}
}

func TestNewClientMissingKeyPath(t *testing.T) {
	server := config.Server{
		Name:     "test-server",
		IP:       "127.0.0.1",
		Port:     22,
		Username: "testuser",
		AuthType: config.AuthTypeKey,
		KeyPath:  "/nonexistent/key",
	}

	_, err := NewClient(server)
	if err == nil {
		t.Error("Expected error for non-existent key path")
	}
}

func TestServerConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		server   config.Server
		authType config.AuthType
		hasCreds bool
	}{
		{
			name: "password auth",
			server: config.Server{
				AuthType: config.AuthTypePassword,
				Password: "test",
			},
			authType: config.AuthTypePassword,
			hasCreds: true,
		},
		{
			name: "key auth",
			server: config.Server{
				AuthType: config.AuthTypeKey,
				KeyPath:  "/path/to/key",
			},
			authType: config.AuthTypeKey,
			hasCreds: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.server.AuthType != tt.authType {
				t.Errorf("Expected auth type %s, got %s", tt.authType, tt.server.AuthType)
			}

			hasCreds := false
			if tt.server.AuthType == config.AuthTypePassword && tt.server.Password != "" {
				hasCreds = true
			}
			if tt.server.AuthType == config.AuthTypeKey && tt.server.KeyPath != "" {
				hasCreds = true
			}

			if hasCreds != tt.hasCreds {
				t.Errorf("Expected hasCreds %v, got %v", tt.hasCreds, hasCreds)
			}
		})
	}
}
