package host

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	fs "solo/internal/tools"
	"sync"
	"time"
)

type HostManager struct {
	config     string
	configs    map[string]Host
	clientPool map[string]*http.Transport
	mu         sync.RWMutex
}

func NewHostManager() *HostManager {
	configDir, err := fs.GetMainConfig(fs.CONFIG_HOST_DIR)
	if err != nil {
		return nil
	}

	hm := &HostManager{
		config:     configDir,
		configs:    make(map[string]Host),
		clientPool: make(map[string]*http.Transport)}

	hm.loadHosts()
	return hm
}

func (hm *HostManager) UpsertHost(config Host) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.configs[config.Name] = config

	// Clear all pooled clients for this hostname (across all ports)
	for key := range hm.clientPool {
		h, _, err := net.SplitHostPort(key)
		if err != nil {
			// key doesn't have a port, it's just the hostname
			if key == config.Name {
				delete(hm.clientPool, key)
			}
		} else {
			// key has a port, compare only the host part
			if h == config.Name {
				delete(hm.clientPool, key)
			}
		}
	}

	return hm.saveHosts()
}

func (hm *HostManager) GetHost(hostname string) (Host, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	h, ok := hm.configs[hostname]
	return h, ok
}

func (hm *HostManager) GetAllHosts() []Host {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	hosts := make([]Host, 0, len(hm.configs))
	for _, v := range hm.configs {
		hosts = append(hosts, v)
	}
	return hosts
}

func (hm *HostManager) loadHosts() {
	data, err := fs.ReadConfigFile(hm.config, fs.CONFIG_HOST_FILENAME)
	if err != nil {
		return
	}

	var hosts []Host
	if err := json.Unmarshal(data, &hosts); err != nil {
		return
	}

	for _, host := range hosts {
		hm.configs[host.Name] = host
	}
}

func (hm *HostManager) saveHosts() error {
	hosts := make([]Host, 0, len(hm.configs))
	for _, v := range hm.configs {
		hosts = append(hosts, v)
	}

	data, err := json.Marshal(hosts)
	if err != nil {
		return err
	}

	if err := fs.UpdateConfigFile(hm.config, fs.CONFIG_HOST_FILENAME, data); err != nil {
		return fs.CreateConfigFile(hm.config, fs.CONFIG_HOST_FILENAME, data)
	}

	return nil
}

func (hm *HostManager) DeleteHost(hostname string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.configs, hostname)

	for key := range hm.clientPool {
		h, _, err := net.SplitHostPort(key)
		if err != nil {
			if key == hostname {
				delete(hm.clientPool, key)
			}
		} else {
			if h == hostname {
				delete(hm.clientPool, key)
			}
		}
	}

	return hm.saveHosts()
}

func (hm *HostManager) GetClientForUrl(resolvedUrl string) (*http.Client, error) {
	parsed, err := url.Parse(resolvedUrl)
	if err != nil {
		slog.Error("Failed to parse URL", "url", resolvedUrl, "error", err)
		return nil, err
	}

	// Separate hostname from port for configuration lookup
	hostname := parsed.Hostname()

	hm.mu.RLock()
	// We still use the full host (with port) for the client pool to avoid sharing
	// connections between different ports, but we use hostname for config.
	transport, exists := hm.clientPool[parsed.Host]
	hm.mu.RUnlock()

	if exists {
		slog.Debug("Reusing pooled HTTP client", "host", parsed.Host)
		return &http.Client{Transport: transport}, nil
	}

	return hm.createNewClient(parsed.Host, hostname)
}

func (hm *HostManager) createNewClient(fullHost string, hostname string) (*http.Client, error) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// double-check locking
	if t, ok := hm.clientPool[fullHost]; ok {
		slog.Debug("Reusing pooled HTTP client (after lock)", "host", fullHost)
		return &http.Client{Transport: t}, nil
	}

	host, ok := hm.configs[hostname]
	if !ok {
		slog.Debug("No custom host configuration found", "hostname", hostname)
	} else {
		slog.Info("Custom host configuration found", "hostname", hostname, "tls_enabled", host.TlsConfig.Enabled)
	}

	transport := &http.Transport{
		TLSClientConfig:    buildTLS(host.TlsConfig, hostname),
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	hm.clientPool[fullHost] = transport
	slog.Info("Created new HTTP client", "host", fullHost, "hostname", hostname)
	return &http.Client{Transport: transport}, nil
}

func buildTLS(config TLSConfig, hostname string) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	if !config.Enabled {
		slog.Debug("TLS not enabled for host", "hostname", hostname)
		return tlsConfig
	}

	hasCerts := false
	if config.PublicCertificateFilePath != "" && config.PrivateKeyFilePath != "" {
		cert, err := tls.LoadX509KeyPair(config.PublicCertificateFilePath, config.PrivateKeyFilePath)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
			hasCerts = true
			slog.Info("Loaded client certificate for mTLS", "hostname", hostname, "cert", config.PublicCertificateFilePath)
		} else {
			slog.Error("Failed to load client certificate key pair",
				"hostname", hostname,
				"cert_path", config.PublicCertificateFilePath,
				"key_path", config.PrivateKeyFilePath,
				"error", err)
		}
	} else if (config.PublicCertificateFilePath != "" && config.PrivateKeyFilePath == "") ||
		(config.PublicCertificateFilePath == "" && config.PrivateKeyFilePath != "") {
		slog.Warn("mTLS configuration incomplete: both certificate and key are required", "hostname", hostname)
	}

	if config.CaCertificateFilePath != "" {
		caCert, err := os.ReadFile(config.CaCertificateFilePath)
		if err == nil {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); ok {
				tlsConfig.RootCAs = caCertPool
				slog.Info("Loaded CA certificate", "hostname", hostname, "ca", config.CaCertificateFilePath)
			} else {
				slog.Error("Failed to append CA certificate to pool", "hostname", hostname, "ca", config.CaCertificateFilePath)
			}
		} else {
			slog.Error("Failed to read CA certificate file",
				"hostname", hostname,
				"ca_path", config.CaCertificateFilePath,
				"error", err)
		}
	}

	tlsConfig.MinVersion = tls.VersionTLS12
	tlsConfig.CurvePreferences = []tls.CurveID{tls.CurveP256, tls.X25519}

	slog.Debug("TLS config built",
		"hostname", hostname,
		"insecure_skip_verify", config.InsecureSkipVerify,
		"has_client_certs", hasCerts,
		"has_ca_cert", config.CaCertificateFilePath != "")

	return tlsConfig
}
