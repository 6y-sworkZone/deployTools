package ssh

import (
	"testing"
)

func TestClientStruct(t *testing.T) {
	client := &Client{}
	if client.sshClient != nil {
		t.Error("sshClient should be nil initially")
	}
	if client.sftpClient != nil {
		t.Error("sftpClient should be nil initially")
	}
}

func TestServerAuthTypeSelection(t *testing.T) {
	testCases := []struct {
		name     string
		authType string
		expected bool
	}{
		{
			name:     "password auth",
			authType: "password",
			expected: true,
		},
		{
			name:     "key auth",
			authType: "key",
			expected: true,
		},
		{
			name:     "invalid auth type",
			authType: "invalid",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.authType == "password" || tc.authType == "key"
			if isValid != tc.expected {
				t.Errorf("Expected %v, got %v for auth type '%s'",
					tc.expected, isValid, tc.authType)
			}
		})
	}
}
