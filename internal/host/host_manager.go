package host

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type HostManager struct {
	configs    map[string]Host
	clientPool map[string]*http.Transport
	mu         sync.RWMutex
}

func NewHostManager() *HostManager {
	return &HostManager{
		configs:    make(map[string]Host),
		clientPool: make(map[string]*http.Transport)}
}

func (hm *HostManager) UpsertHost(config Host) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.configs[config.Name] = config

	delete(hm.clientPool, config.Name)
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

func (hm *HostManager) DeleteHost(hostname string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.configs, hostname)
	delete(hm.clientPool, hostname)
}

func (hm *HostManager) GetClientForUrl(resolvedUrl string) (*http.Client, error) {
	parsed, err := url.Parse(resolvedUrl)
	if err != nil {
		return nil, err
	}
	hostname := parsed.Host

	hm.mu.RLock()
	transport, exists := hm.clientPool[hostname]
	hm.mu.RUnlock()

	if exists {
		return &http.Client{Transport: transport}, nil
	}

	return hm.createNewClient(hostname)
}

func (hm *HostManager) createNewClient(hostname string) (*http.Client, error) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// double-check locking
	if t, ok := hm.clientPool[hostname]; ok {
		return &http.Client{Transport: t}, nil
	}

	host := hm.configs[hostname]

	transport := &http.Transport{
		TLSClientConfig:    buildTLS(host.TlsConfig),
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	hm.clientPool[hostname] = transport
	return &http.Client{Transport: transport}, nil
}

func buildTLS(config TLSConfig) *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	if config.PublicCertificateFilePath != "" && config.PrivateKeyFilePath != "" {
		cert, err := tls.LoadX509KeyPair(config.PublicCertificateFilePath, config.PrivateKeyFilePath)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	if config.CaCertificateFilePath != "" {
		caCert, err := os.ReadFile(config.CaCertificateFilePath)
		if err == nil {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); ok {
				tlsConfig.RootCAs = caCertPool
			}
		}
	}

	tlsConfig.MinVersion = tls.VersionTLS12
	tlsConfig.CurvePreferences = []tls.CurveID{tls.CurveP256, tls.X25519}

	return tlsConfig
}
