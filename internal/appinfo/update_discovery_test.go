// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only
package appinfo

import (
	"os"
	"testing"
)

func TestDiscoveryClient_GetUpdatesFromRepo_LiveServer(t *testing.T) {
	if os.Getenv("SOLO_RUN_LIVE_UPDATE_TESTS") != "1" {
		t.Skip("skipping live update discovery test; set SOLO_RUN_LIVE_UPDATE_TESTS=1 to enable")
	}

	dc := InitDiscoveryCient()

	/*
	   if you want to check the behavior in case you have installed the latest version of the
	   application, just change the param name of GetUpdatesFromRepo with the name of the latest release.
	   Then it should go in the "response == nil" condition.
	*/
	response, err := dc.GetUpdatesFromRepo("0.1.0-rc1")
	if err != nil {
		t.Fatalf("GetUpdatesFromRepo failed: %v", err)
	}
	if response == nil {
		t.Fatal("GetUpdatesFromRepo returned nil response")
	}
	if response.Release == nil {
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

func TestIsCurrentVersionOlder(t *testing.T) {
	cases := []struct {
		name           string
		currentVersion string
		latestVersion  string
		want           bool
	}{
		{
			name:           "current older prerelease",
			currentVersion: "v1.4.0-rc.1",
			latestVersion:  "v1.4.0-rc.3",
			want:           true,
		},
		{
			name:           "current equal latest",
			currentVersion: "v1.4.0-rc.3",
			latestVersion:  "v1.4.0-rc.3",
			want:           false,
		},
		{
			name:           "current newer than latest",
			currentVersion: "v1.5.0-rc.1",
			latestVersion:  "v1.4.9-rc.8",
			want:           false,
		},
		{
			name:           "empty current means update available",
			currentVersion: "",
			latestVersion:  "v1.4.0-rc.3",
			want:           true,
		},
		{
			name:           "dev always updates to latest",
			currentVersion: "dev",
			latestVersion:  "v9.9.9-rc.1",
			want:           true,
		},
		{
			name:           "fallback to lexical compare",
			currentVersion: "build-001",
			latestVersion:  "build-002",
			want:           true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isCurrentVersionOlder(tc.currentVersion, tc.latestVersion)
			if got != tc.want {
				t.Fatalf("isCurrentVersionOlder(%q, %q) = %v, want %v", tc.currentVersion, tc.latestVersion, got, tc.want)
			}
		})
	}
}
