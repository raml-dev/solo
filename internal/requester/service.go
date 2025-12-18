package requester

import (
	"bytes"
	"fmt"
	"io"
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

type ResponseData struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration"`
}

func (s *Service) ExecuteRequest(method, url, body string, headers map[string]any, cookies map[string]any) (*ResponseData, error) {
	resp, err := s.Execute(method, url, body, headers, cookies)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Lettura della risposta
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Conversione headers
	respHeaders := make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		respHeaders[k] = v[0]
	}

	return &ResponseData{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(bodyBytes),
		Duration:   0, // TODO: time tracking
	}, nil
}
