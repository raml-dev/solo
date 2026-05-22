// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"encoding/json"
	"maps"
	"slices"
	"solo/internal/configuration"
	"solo/internal/tools"
	"strings"
	"time"
)

type Request struct {
	Id                  string                                 `json:"id"`
	Name                string                                 `json:"name"`
	Url                 string                                 `json:"url"`
	Verb                string                                 `json:"verb"`
	Body                string                                 `json:"body"`
	BodyType            string                                 `json:"bodyType"`
	Headers             map[string]string                      `json:"headers"`
	Cookies             map[string]string                      `json:"cookies"`
	Params              map[string]string                      `json:"params,omitempty"`
	Auth                *AuthConfiguration                     `json:"auth,omitempty"`
	Settings            *configuration.RequestSettingsOverride `json:"settings,omitempty"`
	PreRequestScript    string                                 `json:"preRequestScript,omitempty"`
	PostResponseScript  string                                 `json:"postResponseScript,omitempty"`
	CreationTimestamp   time.Time                              `json:"creationTimestamp"`
	LastUpdateTimestamp time.Time                              `json:"lastUpdateTimestamp"`
}

func (req *Request) UnmarshalJSON(data []byte) error {
	type Alias Request
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(req),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if req.BodyType == "" {
		req.BodyType = "json"
	} else {
		req.BodyType = strings.ToLower(req.BodyType)
	}

	return nil
}

func (req Request) GetPlaceholders() []string {

	uniqueMap := make(map[string]struct{})

	add := func(text string) {
		matches := tools.ExtractPlaceholders(text)
		for _, m := range matches {
			uniqueMap[m] = struct{}{}
		}
	}

	add(req.Url)
	add(req.Body)

	for _, v := range req.Headers {
		add(v)
	}
	for _, v := range req.Cookies {
		add(v)
	}
	for _, v := range req.Params {
		add(v)
	}

	return slices.Collect(maps.Keys(uniqueMap))
}
