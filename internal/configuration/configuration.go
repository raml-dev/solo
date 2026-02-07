package configuration

type Configuration struct {
	General GeneralSettings `json:"general"`
	Request RequestSettings `json:"request"`
}

type GeneralSettings struct {
	Theme           string `json:"theme"`
	CheckForUpdates bool   `json:"checkForUpdates"`
	// Path overrides could be added here in the future
}

type RequestSettings struct {
	TimeoutSeconds   int    `json:"timeoutSeconds"`
	FollowRedirects  bool   `json:"followRedirects"`
	MaxRedirects     int    `json:"maxRedirects"`
	ValidateSSL      bool   `json:"validateSSL"` // false equals InsecureSkipVerify = true
	DefaultUserAgent string `json:"defaultUserAgent"`
	ProxyURL         string `json:"proxyUrl,omitempty"`
}

// RequestSettingsOverride mirrors RequestSettings but uses pointers to allow nil (unset) values.
// This enables cascading configuration: if a field is nil, the global value is used.
type RequestSettingsOverride struct {
	TimeoutSeconds   *int    `json:"timeoutSeconds,omitempty"`
	FollowRedirects  *bool   `json:"followRedirects,omitempty"`
	MaxRedirects     *int    `json:"maxRedirects,omitempty"`
	ValidateSSL      *bool   `json:"validateSSL,omitempty"`
	DefaultUserAgent *string `json:"defaultUserAgent,omitempty"`
	ProxyURL         *string `json:"proxyUrl,omitempty"`
}
