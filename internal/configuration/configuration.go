package configuration

import "solo/internal/theme"

type Configuration struct {
	General      GeneralSettings `json:"general"`
	Request      RequestSettings `json:"request"`
	CustomThemes []theme.Theme   `json:"customThemes"`
}

type GeneralSettings struct {
	ActiveTheme         string `json:"activeTheme"`
	ThemeMode           string `json:"themeMode,omitempty"`  // "manual" | "sync"
	DayTheme            string `json:"dayTheme,omitempty"`   // tema usato in sync light
	NightTheme          string `json:"nightTheme,omitempty"` // tema usato in sync dark
	CheckForUpdates     bool   `json:"checkForUpdates"`
	DebugMode           bool   `json:"debugMode"`
	SelectedEnvironment string `json:"selectedEnvironment,omitempty"`
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
