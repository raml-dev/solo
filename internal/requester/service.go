package requester

import (
	"bytes"
	"fmt"
	"net/http"
	"yapla/internal/configuration"
)

type Service struct {
	httpClient *http.Client
}

func NewService(config *configuration.Configuration) *Service {
	return &Service{httpClient: http.DefaultClient}
}

func (s *Service) Execute(method, url string, body string, headers map[string]any, cookies map[string]any) (*http.Response, error) {

	fmt.Print(body)
	request, err := http.NewRequest(method, url, bytes.NewReader([]byte(body)))

	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		v, ok := v.(string)

		if !ok {
			return nil, fmt.Errorf("header %s  with value %v is not parsable as string", k, v)
		} else {
			request.Header.Add(k, v)
		}

	}

	for k, v := range cookies {
		v, ok := v.(string)

		if !ok {
			return nil, fmt.Errorf("cookie %s  with value %v is not parsable as string", k, v)
		} else {
			request.AddCookie(&http.Cookie{Name: k, Value: v})
		}

	}

	fmt.Printf("here")

	return s.httpClient.Do(request)

}
