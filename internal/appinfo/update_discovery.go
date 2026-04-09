package appinfo

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v76/github"
	"github.com/google/uuid"
	"github.com/matstech/aegis-go/client"
)

type DiscoveryClient struct {
	client   *http.Client
	endpoint string
	ghClient *github.Client
}

type GitHubResponse struct {
	Releases []GitHubRelease
}

type GitHubRelease struct {
	Assets     []Asset   `json:"assets"`
	Body       string    `json:"string"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Name       string    `json:"name"`
	TagName    string    `json:"tag_name"`
	PreRelease bool      `json:"prerelease"`
}

type Asset struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
	Url   string `json:"browser_download_url"`
}

type discoveryConfig struct {
	Endpoint string `json:"endpoint"`
	Kid      string `json:"kid"`
	Secret   string `json:"secret"`
}

//go:embed update_config.json
var embeddedUpdateConfig []byte

var (
	discoveryConfigOnce sync.Once
	discoveryConfigData discoveryConfig
	discoveryConfigErr  error
)

func InitDiscoveryCient() *DiscoveryClient {
	cfg, err := loadDiscoveryConfig()
	if err != nil {
		slog.Error("Failed to load embedded update discovery configuration", "error", err)
	}

	httpClient := &http.Client{
		Transport: client.NewTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			fmt.Println(curlForRequest(req))
			return http.DefaultTransport.RoundTrip(req)
		}), client.Config{
			Kid:    cfg.Kid,
			Secret: cfg.Secret,
		}),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return newDiscoveryClient(cfg.Endpoint, httpClient)
}

func loadDiscoveryConfig() (discoveryConfig, error) {
	discoveryConfigOnce.Do(func() {
		if len(embeddedUpdateConfig) == 0 {
			discoveryConfigErr = errors.New("embedded update configuration is empty")
			return
		}

		var cfg discoveryConfig
		if err := json.Unmarshal(embeddedUpdateConfig, &cfg); err != nil {
			discoveryConfigErr = fmt.Errorf("invalid embedded update configuration: %w", err)
			return
		}
		if strings.TrimSpace(cfg.Endpoint) == "" {
			discoveryConfigErr = errors.New("update endpoint is required")
			return
		}
		if strings.TrimSpace(cfg.Kid) == "" {
			discoveryConfigErr = errors.New("update kid is required")
			return
		}
		if strings.TrimSpace(cfg.Secret) == "" {
			discoveryConfigErr = errors.New("update secret is required")
			return
		}

		discoveryConfigData = cfg
	})

	return discoveryConfigData, discoveryConfigErr
}

func newDiscoveryClient(endpoint string, httpClient *http.Client) *DiscoveryClient {
	ghClient := github.NewClient(httpClient)
	parsedEndpoint, err := url.Parse(endpoint)
	if err == nil {
		ghClient.BaseURL = &url.URL{
			Scheme: parsedEndpoint.Scheme,
			Host:   parsedEndpoint.Host,
			Path:   "/",
		}
	}

	return &DiscoveryClient{
		client:   httpClient,
		endpoint: endpoint,
		ghClient: ghClient,
	}
}

func (dc *DiscoveryClient) newGetUpdatesRequest() (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, dc.endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-CorrelationId", uuid.NewString())

	return req, nil
}

func (dc *DiscoveryClient) newGetAssetsRequest(endpoint string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-CorrelationId", uuid.NewString())

	if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
		req.Header.Set("Accept", "application/x-apple-diskimage")
	} else {
		req.Header.Set("Accept", "application/octet-stream")
	}

	return req, nil
}

func curlForRequest(req *http.Request) string {
	var parts []string
	parts = append(parts, "curl")
	parts = append(parts, "-X", shellQuote(req.Method))

	for key, values := range req.Header {
		for _, value := range values {
			parts = append(parts, "-H", shellQuote(key+": "+value))
		}
	}

	parts = append(parts, shellQuote(req.URL.String()))
	return strings.Join(parts, " ")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func (dc *DiscoveryClient) GetUpdatesFromRepo() (*GitHubResponse, error) {
	owner, repo, perPage, err := parseRepoEndpoint(dc.endpoint)
	if err != nil {
		return nil, err
	}

	options := &github.ListOptions{PerPage: perPage}
	releases, _, err := dc.ghClient.Repositories.ListReleases(context.Background(), owner, repo, options)
	if err != nil {
		return nil, err
	}

	result := make([]GitHubRelease, 0, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}

		assets := make([]Asset, 0, len(release.Assets))
		for _, asset := range release.Assets {
			assets = append(assets, Asset{
				ID:    asset.GetID(),
				Name:  asset.GetName(),
				State: asset.GetState(),
				Url:   asset.GetBrowserDownloadURL(),
			})
		}

		result = append(result, GitHubRelease{
			Assets:     assets,
			Body:       release.GetBody(),
			CreatedAt:  release.GetCreatedAt().Time,
			UpdatedAt:  release.GetPublishedAt().Time,
			Name:       release.GetName(),
			TagName:    release.GetTagName(),
			PreRelease: release.GetPrerelease(),
		})
	}

	return &GitHubResponse{Releases: result}, nil
}

func (dc *DiscoveryClient) DownloadAssets(info *GitHubResponse, currentVersion string) (string, error) {
	return dc.downloadAssets(info, currentVersion, "")
}

func (dc *DiscoveryClient) DownloadAssetsToPath(info *GitHubResponse, currentVersion, destinationPath string) (string, error) {
	if strings.TrimSpace(destinationPath) == "" {
		return "", errors.New("destination path is required")
	}

	return dc.downloadAssets(info, currentVersion, destinationPath)
}

func (dc *DiscoveryClient) downloadAssets(info *GitHubResponse, currentVersion, destinationPath string) (string, error) {
	_ = currentVersion // Version comparison will be implemented in a dedicated step.
	if info == nil || len(info.Releases) == 0 {
		return "", errors.New("no releases available")
	}

	target := selectTargetRelease(info.Releases)
	if target == nil {
		return "", errors.New("no stable release with downloadable assets")
	}

	asset := selectAsset(target.Assets)
	if asset == nil {
		return "", fmt.Errorf("no compatible asset found for runtime %s/%s (%d-bit)", runtime.GOOS, runtime.GOARCH, strconv.IntSize)
	}

	owner, repo, _, err := parseRepoEndpoint(dc.endpoint)
	if err != nil {
		return "", err
	}

	assetURL, err := buildProxyAssetURL(dc.endpoint, owner, repo, asset.ID)
	if err != nil {
		return "", err
	}
	req, err := dc.newGetAssetsRequest(assetURL)
	if err != nil {
		return "", err
	}
	resp, err := dc.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	outputPath := asset.Name
	if destinationPath != "" {
		outputPath = destinationPath
	}

	fileout, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer fileout.Close()

	if resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest {
		location := resp.Header.Get("Location")
		if location == "" {
			return "", fmt.Errorf("asset download redirected by proxy without location: %s", resp.Status)
		}

		redirectReq, redirectErr := dc.newGetAssetsRequest(location)
		if redirectErr != nil {
			return "", redirectErr
		}
		redirectResp, redirectDoErr := dc.client.Do(redirectReq)
		if redirectDoErr != nil {
			return "", redirectDoErr
		}
		defer redirectResp.Body.Close()
		if redirectResp.StatusCode < http.StatusOK || redirectResp.StatusCode >= http.StatusMultipleChoices {
			body, _ := io.ReadAll(redirectResp.Body)
			return "", fmt.Errorf("asset redirected download failed: %s - %s", redirectResp.Status, string(body))
		}

		_, copyErr := io.Copy(fileout, redirectResp.Body)
		if copyErr != nil {
			return "", copyErr
		}
		return target.Body, nil
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("asset download failed: %s - %s", resp.Status, string(body))
	}
	_, copyErr := io.Copy(fileout, resp.Body)
	if copyErr != nil {
		return "", copyErr
	}

	return target.Body, nil
}

func (dc *DiscoveryClient) doRequest(req *http.Request) string {
	resp, err := dc.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("status:", resp.Status)
	fmt.Println("response:", string(body))

	return string(body)
}

func parseRepoEndpoint(endpoint string) (owner string, repo string, perPage int, err error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", "", 0, err
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[0] != "repos" {
		return "", "", 0, fmt.Errorf("invalid release endpoint path: %s", parsed.Path)
	}

	perPage = 10
	if value := parsed.Query().Get("per_page"); value != "" {
		parsedPerPage, convErr := strconv.Atoi(value)
		if convErr != nil || parsedPerPage <= 0 {
			return "", "", 0, fmt.Errorf("invalid per_page value: %s", value)
		}
		perPage = parsedPerPage
	}

	return parts[1], parts[2], perPage, nil
}

func buildProxyAssetURL(endpoint string, owner string, repo string, assetID int64) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	parsed.Path = fmt.Sprintf("/repos/%s/%s/releases/assets/%d", owner, repo, assetID)
	parsed.RawQuery = ""
	return parsed.String(), nil
}

func selectTargetRelease(releases []GitHubRelease) *GitHubRelease {
	for i := range releases {
		release := &releases[i]
		if release.PreRelease || len(release.Assets) == 0 {
			continue
		}
		return release
	}
	for i := range releases {
		release := &releases[i]
		if len(release.Assets) == 0 {
			continue
		}
		return release
	}
	return nil
}

func selectAsset(assets []Asset) *Asset {
	osName := strings.ToLower(runtime.GOOS)
	arch := strings.ToLower(runtime.GOARCH)

	bestScore := -1
	var best *Asset
	for _, asset := range assets {
		if asset.ID == 0 {
			continue
		}

		name := strings.ToLower(asset.Name)
		if name == "" {
			name = strings.ToLower(asset.Url)
		}

		score := 0
		if strings.Contains(name, osName) {
			score += 4
		}
		if strings.Contains(name, arch) {
			score += 5
		}
		if is32BitRuntime() && strings.Contains(name, "32") {
			score += 2
		}
		if !is32BitRuntime() && strings.Contains(name, "64") {
			score += 2
		}
		if runtime.GOOS == "darwin" && strings.Contains(name, ".dmg") {
			score += 2
		}
		if runtime.GOOS == "windows" && strings.Contains(name, ".exe") {
			score += 2
		}
		if runtime.GOOS == "linux" && strings.Contains(name, ".appimage") {
			score += 2
		}

		if score > bestScore {
			bestScore = score
			chosen := asset
			best = &chosen
		}
	}

	return best
}

func is32BitRuntime() bool {
	return strconv.IntSize == 32
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
