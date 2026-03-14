package importer

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	models "yapla/internal/collection"
)

type PostmanImporter struct{}

func NewPostmanImporter() *PostmanImporter {
	return &PostmanImporter{}
}

func (p *PostmanImporter) Import(path string) (*models.Collection, error) {
	slog.Info("Importing Postman collection", "path", path)
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere il file: %w", err)
	}

	var pc postmanCollection
	if err := json.Unmarshal(fileData, &pc); err != nil {
		return nil, fmt.Errorf("errore nel parsing della collection Postman: %w", err)
	}

	slog.Debug("Postman collection unmarshaled", "name", pc.Info.Name, "items_count", len(pc.Item))

	now := time.Now()
	coll := &models.Collection{
		Id:                  generateUUID(),
		Name:                pc.Info.Name,
		Requests:            []models.Request{},
		CreationTimestamp:   now,
		LastUpdateTimestamp: now,
	}

	processItems(pc.Item, "", &coll.Requests)

	slog.Info("Postman import completed", "requests_found", len(coll.Requests))
	return coll, nil
}

// Recursive function to flatten folders
func processItems(items []postmanItem, folderPath string, dest *[]models.Request) {
	for _, item := range items {
		// 1. Process as folder if it has children
		if len(item.Item) > 0 {
			newPath := item.Name + " / "
			if folderPath != "" {
				newPath = folderPath + item.Name + " / "
			}
			processItems(item.Item, newPath, dest)
			// Note: We don't skip the rest of the loop because some exporters 
			// might put a request and an item array in the same object.
		}

		// 2. Process as request if it has request data
		if item.Request != nil {
			reqName := item.Name
			if folderPath != "" {
				reqName = folderPath + item.Name
			}

			slog.Debug("Processing Postman request", "name", reqName)

			req := models.Request{
				Id:                  generateUUID(),
				Name:                reqName,
				Url:                 item.Request.URL.Raw,
				Verb:                item.Request.Method,
				Headers:             make(map[string]string),
				Cookies:             make(map[string]string),
				CreationTimestamp:   time.Now(),
				LastUpdateTimestamp: time.Now(),
			}

			// Convert Headers from an array of structs to a map
			for _, h := range item.Request.Header {
				req.Headers[h.Key] = h.Value
			}

			// Parse the Body and determine BodyType
			if item.Request.Body != nil {
				req.Body = item.Request.Body.Raw

				bodyLang := item.Request.Body.Options.Raw.Language
				if bodyLang != "" {
					req.BodyType = bodyLang
				} else {
					// Fallback: check the Content-Type header
					contentType := ""
					for k, v := range req.Headers {
						if strings.ToLower(k) == "content-type" {
							contentType = strings.ToLower(v)
							break
						}
					}

					if strings.Contains(contentType, "json") {
						req.BodyType = "json"
					} else if strings.Contains(contentType, "xml") {
						req.BodyType = "xml"
					} else if strings.Contains(contentType, "html") {
						req.BodyType = "html"
					} else if contentType != "" {
						req.BodyType = "text"
					}
				}
			}

			*dest = append(*dest, req)
		}
	}
}

// UUID v4 generator using only crypto/rand
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Safe fallback if rand.Read fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	// Set version to 4
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant to RFC4122
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// --- Internal models for partial parsing of Postman v2.1 JSON ---

type postmanCollection struct {
	Info postmanInfo   `json:"info"`
	Item []postmanItem `json:"item"`
}

type postmanInfo struct {
	Name string `json:"name"`
}

type postmanItem struct {
	Name    string          `json:"name"`
	Item    []postmanItem   `json:"item"` // For nested folders
	Request *postmanRequest `json:"request"`
}

type postmanRequest struct {
	Method string          `json:"method"`
	URL    postmanURL      `json:"url"`
	Header []postmanHeader `json:"header"`
	Body   *postmanBody    `json:"body"`
}

type postmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type postmanBody struct {
	Mode    string `json:"mode"`
	Raw     string `json:"raw"`
	Options struct {
		Raw struct {
			Language string `json:"language"`
		} `json:"raw"`
	} `json:"options"`
}

// postmanURL handles URLs that can be either a string or an object
type postmanURL struct {
	Raw string `json:"raw"`
}

// Custom Unmarshaler to support Postman's dual URL representation (string or object)
func (u *postmanURL) UnmarshalJSON(data []byte) error {
	// First try to deserialize it as a string
	var raw string
	if err := json.Unmarshal(data, &raw); err == nil {
		u.Raw = raw
		return nil
	}

	// Otherwise try as an object
	var obj struct {
		Raw      string   `json:"raw"`
		Protocol string   `json:"protocol"`
		Host     []string `json:"host"`
		Path     []string `json:"path"`
		Query    []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"query"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		if obj.Raw != "" {
			u.Raw = obj.Raw
		} else {
			// Ricostruisce l'URL se il campo "raw" non è presente
			urlStr := ""
			if obj.Protocol != "" {
				urlStr += obj.Protocol + "://"
			}
			if len(obj.Host) > 0 {
				urlStr += strings.Join(obj.Host, ".")
			}
			if len(obj.Path) > 0 {
				urlStr += "/" + strings.Join(obj.Path, "/")
			}
			
			// Aggiungiamo le query string
			if len(obj.Query) > 0 {
				qs := []string{}
				for _, q := range obj.Query {
					if q.Value != "" {
						qs = append(qs, q.Key+"="+q.Value)
					} else {
						qs = append(qs, q.Key)
					}
				}
				urlStr += "?" + strings.Join(qs, "&")
			}
			
			u.Raw = urlStr
		}
		return nil
	}

	return nil
}
