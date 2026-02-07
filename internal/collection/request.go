package collection

import (
	"maps"
	"slices"
	"time"
	"yapla/internal/configuration"
	"yapla/internal/tools"
)

type Request struct {
	Id                  string                                 `json:"id"`
	Name                string                                 `json:"name"`
	Url                 string                                 `json:"url"`
	Verb                string                                 `json:"verb"`
	Body                string                                 `json:"body"`
	Headers             map[string]string                      `json:"headers"`
	Cookies             map[string]string                      `json:"cookies"`
	Settings            *configuration.RequestSettingsOverride `json:"settings,omitempty"`
	CreationTimestamp   time.Time                              `json:"creationTimestamp"`
	LastUpdateTimestamp time.Time                              `json:"lastUpdateTimestamp"`
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

	return slices.Collect(maps.Keys(uniqueMap))
}
