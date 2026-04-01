// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package host

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	fs "solo/internal/tools"
	"testing"
)

func TestUpsertHost(t *testing.T) {
	tests := []struct {
		name          string
		host          Host
		setupExisting bool
		expectError   bool
	}{
		{
			name: "Create new host",
			host: Host{
				Id:   "1",
				Name: "api.example.com",
				TlsConfig: TLSConfig{
					Enabled:            false,
					InsecureSkipVerify: false,
				},
			},
			setupExisting: false,
			expectError:   false,
		},
		{
			name: "Update existing host",
			host: Host{
				Id:   "1",
				Name: "api.example.com",
				TlsConfig: TLSConfig{
					Enabled:            true,
					InsecureSkipVerify: true,
				},
			},
			setupExisting: true,
			expectError:   false,
		},
		{
			name: "Create host with TLS config",
			host: Host{
				Id:   "2",
				Name: "secure.example.com",
				TlsConfig: TLSConfig{
					Enabled:                   true,
					PublicCertificateFilePath: "/path/to/cert.pem",
					PrivateKeyFilePath:        "/path/to/key.pem",
					CaCertificateFilePath:     "/path/to/ca.pem",
					InsecureSkipVerify:        false,
				},
			},
			setupExisting: false,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm := setupTestHostManager(t)
			defer cleanupTestDir(hm.config)

			if tt.setupExisting {
				if err := hm.UpsertHost(Host{Id: "1", Name: "api.example.com"}); err != nil {
					t.Fatalf("Failed to setup existing host: %v", err)
				}
			}

			err := hm.UpsertHost(tt.host)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify host was persisted
				data, readErr := os.ReadFile(buildHostFileName(hm.config))
				if readErr != nil {
					t.Errorf("Failed to read hosts file: %v", readErr)
				}

				var hosts []Host
				if unmarshalErr := json.Unmarshal(data, &hosts); unmarshalErr != nil {
					t.Errorf("Failed to unmarshal hosts: %v", unmarshalErr)
				}

				found := false
				for _, h := range hosts {
					if h.Name == tt.host.Name {
						found = true
						if h.Id != tt.host.Id {
							t.Errorf("Expected host ID %s, got %s", tt.host.Id, h.Id)
						}
						if h.TlsConfig.Enabled != tt.host.TlsConfig.Enabled {
							t.Errorf("Expected TLS enabled %v, got %v", tt.host.TlsConfig.Enabled, h.TlsConfig.Enabled)
						}
						break
					}
				}

				if !found {
					t.Errorf("Host %s not found in persisted data", tt.host.Name)
				}
			}
		})
	}
}

func TestGetAllHosts(t *testing.T) {
	tests := []struct {
		name          string
		setupHosts    []Host
		expectedCount int
	}{
		{
			name:          "Get all hosts when empty",
			setupHosts:    []Host{},
			expectedCount: 0,
		},
		{
			name: "Get all hosts with one host",
			setupHosts: []Host{
				{Id: "1", Name: "api.example.com"},
			},
			expectedCount: 1,
		},
		{
			name: "Get all hosts with multiple hosts",
			setupHosts: []Host{
				{Id: "1", Name: "api.example.com"},
				{Id: "2", Name: "api2.example.com"},
				{Id: "3", Name: "api3.example.com"},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm := setupTestHostManager(t)
			defer cleanupTestDir(hm.config)

			for _, host := range tt.setupHosts {
				if err := hm.UpsertHost(host); err != nil {
					t.Fatalf("Failed to setup host: %v", err)
				}
			}

			hosts := hm.GetAllHosts()

			if len(hosts) != tt.expectedCount {
				t.Errorf("Expected %d hosts, got %d", tt.expectedCount, len(hosts))
			}

			// Verify hosts are correctly retrieved
			for _, setupHost := range tt.setupHosts {
				found := false
				for _, retrievedHost := range hosts {
					if retrievedHost.Name == setupHost.Name {
						found = true
						if retrievedHost.Id != setupHost.Id {
							t.Errorf("Expected host ID %s, got %s", setupHost.Id, retrievedHost.Id)
						}
						break
					}
				}
				if !found {
					t.Errorf("Host %s not found in retrieved hosts", setupHost.Name)
				}
			}
		})
	}
}

func TestDeleteHost(t *testing.T) {
	tests := []struct {
		name          string
		hostname      string
		setupHost     bool
		expectError   bool
		expectedCount int
	}{
		{
			name:          "Delete existing host",
			hostname:      "api.example.com",
			setupHost:     true,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "Delete non-existent host",
			hostname:      "non-existent.com",
			setupHost:     false,
			expectError:   false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm := setupTestHostManager(t)
			defer cleanupTestDir(hm.config)

			if tt.setupHost {
				host := Host{Id: "1", Name: tt.hostname}
				if err := hm.UpsertHost(host); err != nil {
					t.Fatalf("Failed to setup host: %v", err)
				}
			}

			err := hm.DeleteHost(tt.hostname)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify host was deleted from file
				hosts := hm.GetAllHosts()
				if len(hosts) != tt.expectedCount {
					t.Errorf("Expected %d hosts after deletion, got %d", tt.expectedCount, len(hosts))
				}

				for _, h := range hosts {
					if h.Name == tt.hostname {
						t.Errorf("Host %s should have been deleted", tt.hostname)
					}
				}
			}
		})
	}
}

func TestLoadHosts(t *testing.T) {
	tests := []struct {
		name       string
		setupHosts []Host
	}{
		{
			name:       "Load empty hosts file",
			setupHosts: []Host{},
		},
		{
			name: "Load hosts from file",
			setupHosts: []Host{
				{Id: "1", Name: "api.example.com"},
				{Id: "2", Name: "api2.example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm := setupTestHostManager(t)
			defer cleanupTestDir(hm.config)

			// Setup hosts
			for _, host := range tt.setupHosts {
				hm.UpsertHost(host)
			}

			// Create new manager to test loading
			hm2 := &HostManager{
				config:     hm.config,
				configs:    make(map[string]Host),
				clientPool: make(map[string]*http.Transport),
			}
			hm2.loadHosts()

			// Verify all hosts were loaded
			if len(hm2.configs) != len(tt.setupHosts) {
				t.Errorf("Expected %d hosts loaded, got %d", len(tt.setupHosts), len(hm2.configs))
			}

			for _, setupHost := range tt.setupHosts {
				if loaded, ok := hm2.configs[setupHost.Name]; !ok {
					t.Errorf("Host %s not loaded", setupHost.Name)
				} else if loaded.Id != setupHost.Id {
					t.Errorf("Expected host ID %s, got %s", setupHost.Id, loaded.Id)
				}
			}
		})
	}
}

func TestSaveHosts(t *testing.T) {
	tests := []struct {
		name  string
		hosts []Host
	}{
		{
			name:  "Save empty hosts",
			hosts: []Host{},
		},
		{
			name: "Save multiple hosts",
			hosts: []Host{
				{Id: "1", Name: "api.example.com"},
				{Id: "2", Name: "api2.example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm := setupTestHostManager(t)
			defer cleanupTestDir(hm.config)

			// Add hosts to manager
			for _, host := range tt.hosts {
				hm.configs[host.Name] = host
			}

			// Save hosts
			err := hm.saveHosts()
			if err != nil {
				t.Errorf("Unexpected error saving hosts: %v", err)
			}

			// Verify file was created and contains correct data
			data, readErr := os.ReadFile(buildHostFileName(hm.config))
			if readErr != nil {
				t.Errorf("Failed to read hosts file: %v", readErr)
			}

			var savedHosts []Host
			if unmarshalErr := json.Unmarshal(data, &savedHosts); unmarshalErr != nil {
				t.Errorf("Failed to unmarshal hosts: %v", unmarshalErr)
			}

			if len(savedHosts) != len(tt.hosts) {
				t.Errorf("Expected %d hosts saved, got %d", len(tt.hosts), len(savedHosts))
			}
		})
	}
}

func TestPersistenceAcrossOperations(t *testing.T) {
	hm := setupTestHostManager(t)
	defer cleanupTestDir(hm.config)

	// Add first host
	host1 := Host{Id: "1", Name: "api.example.com"}
	if err := hm.UpsertHost(host1); err != nil {
		t.Fatalf("Failed to upsert host1: %v", err)
	}

	// Add second host
	host2 := Host{Id: "2", Name: "api2.example.com"}
	if err := hm.UpsertHost(host2); err != nil {
		t.Fatalf("Failed to upsert host2: %v", err)
	}

	// Verify both hosts are present
	hosts := hm.GetAllHosts()
	if len(hosts) != 2 {
		t.Errorf("Expected 2 hosts, got %d", len(hosts))
	}

	// Delete first host
	if err := hm.DeleteHost(host1.Name); err != nil {
		t.Fatalf("Failed to delete host1: %v", err)
	}

	// Verify only one host remains
	hosts = hm.GetAllHosts()
	if len(hosts) != 1 {
		t.Errorf("Expected 1 host after deletion, got %d", len(hosts))
	}
	if hosts[0].Name != host2.Name {
		t.Errorf("Expected remaining host to be %s, got %s", host2.Name, hosts[0].Name)
	}

	// Create new manager and verify persistence
	hm2 := &HostManager{
		config:     hm.config,
		configs:    make(map[string]Host),
		clientPool: make(map[string]*http.Transport),
	}
	hm2.loadHosts()

	hosts2 := hm2.GetAllHosts()
	if len(hosts2) != 1 {
		t.Errorf("Expected 1 host after reload, got %d", len(hosts2))
	}
	if hosts2[0].Name != host2.Name {
		t.Errorf("Expected loaded host to be %s, got %s", host2.Name, hosts2[0].Name)
	}
}

func TestGetClientForURL_HostPortPrecedence(t *testing.T) {
	hm := setupTestHostManager(t)
	defer cleanupTestDir(hm.config)

	hm.configs["localhost"] = Host{
		Id:   "host-only",
		Name: "localhost",
		TlsConfig: TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: false,
		},
	}
	hm.configs["localhost:8443"] = Host{
		Id:   "host-port",
		Name: "localhost:8443",
		TlsConfig: TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
		},
	}

	client, err := hm.GetClientForUrl("https://localhost:8443/health")
	if err != nil {
		t.Fatalf("GetClientForUrl failed: %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.Transport)
	}
	if transport.TLSClientConfig == nil {
		t.Fatalf("expected TLSClientConfig to be set")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("expected host:port TLS config to win (InsecureSkipVerify=true)")
	}
}

func TestGetClientForURL_FallsBackToHostnameWhenHostPortNotConfigured(t *testing.T) {
	hm := setupTestHostManager(t)
	defer cleanupTestDir(hm.config)

	hm.configs["localhost"] = Host{
		Id:   "host-only",
		Name: "localhost",
		TlsConfig: TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
		},
	}

	client, err := hm.GetClientForUrl("https://localhost:9443/health")
	if err != nil {
		t.Fatalf("GetClientForUrl failed: %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.Transport)
	}
	if transport.TLSClientConfig == nil {
		t.Fatalf("expected TLSClientConfig to be set")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("expected hostname fallback TLS config to be applied (InsecureSkipVerify=true)")
	}
}

func TestGetHost_HostPortThenHostnameFallback(t *testing.T) {
	hm := setupTestHostManager(t)
	defer cleanupTestDir(hm.config)

	hm.configs["localhost"] = Host{Id: "host-only", Name: "localhost"}
	hm.configs["localhost:8443"] = Host{Id: "host-port", Name: "localhost:8443"}

	h, ok := hm.GetHost("localhost:8443")
	if !ok {
		t.Fatalf("expected exact host:port match")
	}
	if h.Id != "host-port" {
		t.Fatalf("expected host-port config, got %q", h.Id)
	}

	h, ok = hm.GetHost("localhost:9443")
	if !ok {
		t.Fatalf("expected hostname fallback match")
	}
	if h.Id != "host-only" {
		t.Fatalf("expected hostname fallback config, got %q", h.Id)
	}
}

// Helper functions
func setupTestHostManager(t *testing.T) *HostManager {
	tmpDir := filepath.Join(os.TempDir(), "solo-test-host-"+t.Name())
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return &HostManager{
		config:     tmpDir,
		configs:    make(map[string]Host),
		clientPool: make(map[string]*http.Transport),
	}
}

func cleanupTestDir(path string) {
	os.RemoveAll(path)
}

func buildHostFileName(configPath string) string {
	return filepath.Join(configPath, fs.CONFIG_HOST_FILENAME)
}
