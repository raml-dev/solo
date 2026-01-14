package host

type Host struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	TlsConfig TLSConfig `json:"tlsConfig"`
}

type TLSConfig struct {
	Enabled                   bool   `json:"enabled"`
	PublicCertificateFilePath string `json:"publicCertificateFilePath"` // .crt or .pem
	PrivateKeyFilePath        string `json:"privateKeyFilePath"`        // .key
	CaCertificateFilePath     string `json:"caCertificateFilePath"`     // For mTLS or self-signed certificate
	InsecureSkipVerify        bool   `json:"insecureSkipVerify"`
}
