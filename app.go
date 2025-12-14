package main

import (
	"context"
	"io"
	"yapla/internal/requester"
)

// App struct
type App struct {
	ctx     context.Context
	service *requester.Service
}

type RequestOptions struct {
	Method  string         `json:"method"`
	URL     string         `json:"url"`
	Headers map[string]any `json:"headers"`
	Body    string         `json:"body"`
}

type ResponseData struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{service: requester.NewService(nil)}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Execute(options RequestOptions) (ResponseData, error) {
	result := ResponseData{}

	resp, err := a.service.Execute(options.Method, options.URL, options.Body, options.Headers, nil)

	if err != nil {
		return ResponseData{}, err
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// 6. Lettura della Risposta
	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return result, err
	}

	result.Body = string(bodyBytes)

	respHeaders := make(map[string]string, len(resp.Header))

	for k, v := range resp.Header {
		respHeaders[k] = v[0]
	}

	result.Headers = respHeaders

	return result, nil
}
