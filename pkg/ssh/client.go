package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"deploytools/internal/config"
)

type Client struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	server     config.Server
}

func NewClient(server config.Server) (*Client, error) {
	var auth ssh.AuthMethod
	var err error

	if server.AuthType == config.AuthTypeKey {
		auth, err = getKeyAuth(server.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load SSH key: %w", err)
		}
	} else {
		auth = ssh.Password(server.Password)
	}

	sshConfig := &ssh.ClientConfig{
		User:            server.Username,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", server.IP, server.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &Client{
		sshClient:  sshClient,
		sftpClient: sftpClient,
		server:     server,
	}, nil
}

func getKeyAuth(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}

func (c *Client) TestConnection() error {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.Run("echo test")
}

func (c *Client) Close() error {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		return c.sshClient.Close()
	}
	return nil
}

func (c *Client) UploadFile(localPath, remotePath string, progressFn func(int64, int64)) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	localInfo, err := localFile.Stat()
	if err != nil {
		return err
	}

	remoteDir := filepath.Dir(remotePath)
	if err := c.MkdirAll(remoteDir); err != nil {
		return err
	}

	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	totalSize := localInfo.Size()
	var uploaded int64
	buf := make([]byte, 32*1024)

	for {
		n, err := localFile.Read(buf)
		if n > 0 {
			nw, err := remoteFile.Write(buf[:n])
			if err != nil {
				return err
			}
			uploaded += int64(nw)
			if progressFn != nil {
				progressFn(uploaded, totalSize)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DownloadFile(remotePath, localPath string, progressFn func(int64, int64)) error {
	remoteFile, err := c.sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	remoteInfo, err := remoteFile.Stat()
	if err != nil {
		return err
	}

	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	totalSize := remoteInfo.Size()
	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		n, err := remoteFile.Read(buf)
		if n > 0 {
			nw, err := localFile.Write(buf[:n])
			if err != nil {
				return err
			}
			downloaded += int64(nw)
			if progressFn != nil {
				progressFn(downloaded, totalSize)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) MkdirAll(remotePath string) error {
	return c.sftpClient.MkdirAll(remotePath)
}

func (c *Client) FileExists(remotePath string) bool {
	_, err := c.sftpClient.Stat(remotePath)
	return err == nil
}

func (c *Client) GetFileMD5(remotePath string) (string, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(fmt.Sprintf("md5sum %s 2>/dev/null || md5 %s 2>/dev/null", remotePath, remotePath))
	if err != nil {
		return "", err
	}

	parts := strings.Fields(string(output))
	if len(parts) > 0 {
		return parts[0], nil
	}
	return "", fmt.Errorf("could not get MD5 hash")
}

func (c *Client) Remove(remotePath string) error {
	return c.sftpClient.Remove(remotePath)
}

func (c *Client) RemoveDirectory(remotePath string) error {
	walker := c.sftpClient.Walk(remotePath)
	var files []string
	var dirs []string

	for walker.Step() {
		if err := walker.Err(); err != nil {
			continue
		}
		path := walker.Path()
		if walker.Stat().IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
	}

	for _, file := range files {
		if err := c.sftpClient.Remove(file); err != nil {
			return err
		}
	}

	for i := len(dirs) - 1; i >= 0; i-- {
		if err := c.sftpClient.RemoveDirectory(dirs[i]); err != nil {
			return err
		}
	}

	return c.sftpClient.RemoveDirectory(remotePath)
}

func (c *Client) RunCommand(cmd string) (string, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

func (c *Client) ListFiles(remotePath string) ([]os.FileInfo, error) {
	return c.sftpClient.ReadDir(remotePath)
}
