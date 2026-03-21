// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package host

type Host struct {
	Id        string            `json:"id"`
	Name      string            `json:"name"`
	TlsConfig TLSConfig         `json:"tlsConfig"`
	Cookies   map[string]string `json:"cookies"`
}

type TLSConfig struct {
	Enabled                   bool   `json:"enabled"`
	PublicCertificateFilePath string `json:"publicCertificateFilePath"` // .crt or .pem
	PrivateKeyFilePath        string `json:"privateKeyFilePath"`        // .key
	CaCertificateFilePath     string `json:"caCertificateFilePath"`     // For mTLS or self-signed certificate
	InsecureSkipVerify        bool   `json:"insecureSkipVerify"`
}
