package appinfo

import (
	"testing"
)

func TestDiscoveryClient_GetUpdatesFromRepo_LiveServer(t *testing.T) {
	dc := InitDiscoveryCient()

	response, err := dc.GetUpdatesFromRepo()
	if err != nil {
		t.Fatalf("GetUpdatesFromRepo failed: %v", err)
	}
	if response == nil {
		t.Fatal("GetUpdatesFromRepo returned nil response")
	}
	if len(response.Releases) == 0 {
		t.Fatal("GetUpdatesFromRepo returned no releases")
	}

	s, err := dc.DownloadAssets(response, "")
	if err != nil {
		t.Fatalf("DownloadAssets failed: %v", err)
	}
	if s == "" {
		t.Fatal("DownloadAssets returned empty release body")
	}
}
