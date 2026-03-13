package host

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
	fs "yapla/internal/tools"
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
	delete(hm.clientPool, config.Name)

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
	delete(hm.clientPool, hostname)

	return hm.saveHosts()
}

func (hm *HostManager) GetClientForUrl(resolvedUrl string) (*http.Client, error) {
	parsed, err := url.Parse(resolvedUrl)
	if err != nil {
		slog.Error("Failed to parse URL", "url", resolvedUrl, "error", err)
		return nil, err
	}
	hostname := parsed.Host

	hm.mu.RLock()
	transport, exists := hm.clientPool[hostname]
	hm.mu.RUnlock()

	if exists {
		slog.Debug("Reusing pooled HTTP client", "hostname", hostname)
		return &http.Client{Transport: transport}, nil
	}

	return hm.createNewClient(hostname)
}

func (hm *HostManager) createNewClient(hostname string) (*http.Client, error) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// double-check locking
	if t, ok := hm.clientPool[hostname]; ok {
		slog.Debug("Reusing pooled HTTP client (after lock)", "hostname", hostname)
		return &http.Client{Transport: t}, nil
	}

	host := hm.configs[hostname]

	transport := &http.Transport{
		TLSClientConfig:    buildTLS(host.TlsConfig, hostname),
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	hm.clientPool[hostname] = transport
	slog.Info("Created new HTTP client", "hostname", hostname)
	return &http.Client{Transport: transport}, nil
}

func buildTLS(config TLSConfig, hostname string) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	hasCerts := false
	if config.PublicCertificateFilePath != "" && config.PrivateKeyFilePath != "" {
		cert, err := tls.LoadX509KeyPair(config.PublicCertificateFilePath, config.PrivateKeyFilePath)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
			hasCerts = true
		} else {
			slog.Warn("Failed to load TLS key pair",
				"hostname", hostname,
				"cert_path", config.PublicCertificateFilePath,
				"key_path", config.PrivateKeyFilePath,
				"error", err)
		}
	}

	if config.CaCertificateFilePath != "" {
		caCert, err := os.ReadFile(config.CaCertificateFilePath)
		if err == nil {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); ok {
				tlsConfig.RootCAs = caCertPool
			}
		} else {
			slog.Warn("Failed to load CA certificate",
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
