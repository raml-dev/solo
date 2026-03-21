package importer

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"solo/internal/collection"
	"strings"
	"time"
)

// CurlImporter parses a cURL command string and converts it to a collection.Request.
type CurlImporter struct{}

func NewCurlImporter() *CurlImporter {
	return &CurlImporter{}
}

// ParseRequest parses a cURL command string and returns a collection.Request.
// The returned request has a new unique ID and timestamps set to now.
func (c *CurlImporter) ParseRequest(curlString string) (collection.Request, error) {
	tokens, err := tokenizeCurl(curlString)
	if err != nil {
		return collection.Request{}, err
	}
	if len(tokens) == 0 || !strings.EqualFold(tokens[0], "curl") {
		return collection.Request{}, fmt.Errorf("not a curl command")
	}

	var (
		rawURL     string
		verb       string
		body       string
		authHeader string
		headers    = make(map[string]string)
		cookies    = make(map[string]string)
	)

	args := tokens[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Helper: consume next token as value
		next := func() (string, bool) {
			if i+1 < len(args) {
				i++
				return args[i], true
			}
			return "", false
		}

		switch arg {
		case "-X", "--request":
			if v, ok := next(); ok {
				verb = strings.ToUpper(v)
			}

		case "-H", "--header":
			if v, ok := next(); ok {
				if name, val, ok2 := parseHeader(v); ok2 {
					headers[name] = val
				}
			}

		case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii":
			if v, ok := next(); ok {
				body = v
			}

		case "-u", "--user":
			if v, ok := next(); ok {
				encoded := base64.StdEncoding.EncodeToString([]byte(v))
				authHeader = "Basic " + encoded
			}

		case "-b", "--cookie":
			if v, ok := next(); ok {
				for name, val := range parseCookies(v) {
					cookies[name] = val
				}
			}

		case "--url":
			if v, ok := next(); ok {
				rawURL = v
			}

		default:
			if strings.HasPrefix(arg, "-") {
				// Unknown flag — log and skip.
				// Boolean flags (no argument) must not consume the next token.
				if curlBoolFlag[arg] || i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
					slog.Debug("curl import: unsupported flag skipped", "flag", arg)
				} else {
					slog.Debug("curl import: unsupported flag skipped", "flag", arg, "value", args[i+1])
					i++
				}
			} else {
				// Positional argument — treat as URL
				rawURL = arg
			}
		}
	}

	if rawURL == "" {
		return collection.Request{}, fmt.Errorf("no URL found in curl command")
	}

	// Infer method
	if verb == "" {
		if body != "" {
			verb = "POST"
		} else {
			verb = "GET"
		}
	}

	// Authorization from -u (only if -H didn't already set it)
	if authHeader != "" {
		if _, exists := headers["Authorization"]; !exists {
			headers["Authorization"] = authHeader
		}
	}

	// Body type detection
	bodyType := ""
	if body != "" {
		trimmed := strings.TrimSpace(body)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			bodyType = "json"
		} else {
			bodyType = "text"
		}
	}

	// Build request name from URL path
	name := buildCurlRequestName(verb, rawURL)

	now := time.Now()
	req := collection.Request{
		Id:                  generateUUID(),
		Name:                name,
		Url:                 rawURL,
		Verb:                verb,
		Body:                body,
		BodyType:            bodyType,
		Headers:             headers,
		Cookies:             cookies,
		CreationTimestamp:   now,
		LastUpdateTimestamp: now,
	}

	return req, nil
}

// curlBoolFlag lists known cURL flags that take no argument.
var curlBoolFlag = map[string]bool{
	"-s": true, "--silent": true,
	"-v": true, "--verbose": true,
	"-L": true, "--location": true,
	"-k": true, "--insecure": true,
	"-i": true, "--include": true,
	"-I": true, "--head": true,
	"-g": true, "--globoff": true,
	"-G": true, "--get": true,
	"-n": true, "--netrc": true,
	"-N": true, "--no-buffer": true,
	"--compressed":   true,
	"--http1.0":      true,
	"--http1.1":      true,
	"--http2":        true,
	"--http3":        true,
	"--ipv4":         true,
	"--ipv6":         true,
	"--no-keepalive": true,
	"--fail":         true, "-f": true,
	"--silent-show-error": true, "-S": true,
}

// buildCurlRequestName returns "<VERB> <path>" using the path component of the URL,
// or "<VERB> <rawURL>" if the URL cannot be parsed.
func buildCurlRequestName(verb, rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Path != "" {
		return verb + " " + u.Path
	}
	return verb + " " + rawURL
}

// parseHeader splits a raw "Name: Value" header string.
// Returns (name, value, true) on success, ("", "", false) on failure.
func parseHeader(raw string) (name, value string, ok bool) {
	idx := strings.Index(raw, ":")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(raw[:idx]), strings.TrimSpace(raw[idx+1:]), true
}

// parseCookies splits a "name=value; name2=value2" cookie string into a map.
func parseCookies(raw string) map[string]string {
	result := make(map[string]string)
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.Index(part, "=")
		if idx < 0 {
			continue
		}
		name := strings.TrimSpace(part[:idx])
		val := strings.TrimSpace(part[idx+1:])
		if name != "" {
			result[name] = val
		}
	}
	return result
}

// tokenizeCurl splits a cURL command string into tokens, respecting single and
// double quotes and handling line-continuation sequences (\<newline>).
func tokenizeCurl(s string) ([]string, error) {
	// Normalise line continuations: backslash + optional spaces + newline → single space
	s = normalizeLineContinuations(s)

	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	i := 0
	runes := []rune(s)

	for i < len(runes) {
		ch := runes[i]

		switch {
		case inSingle:
			if ch == '\'' {
				inSingle = false
			} else {
				current.WriteRune(ch)
			}
			i++

		case inDouble:
			if ch == '"' {
				inDouble = false
				i++
			} else if ch == '\\' && i+1 < len(runes) {
				next := runes[i+1]
				switch next {
				case '"', '\\', '$', '`':
					current.WriteRune(next)
				default:
					current.WriteRune(ch)
					current.WriteRune(next)
				}
				i += 2
			} else {
				current.WriteRune(ch)
				i++
			}

		case ch == '\'':
			inSingle = true
			i++

		case ch == '"':
			inDouble = true
			i++

		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			i++

		default:
			current.WriteRune(ch)
			i++
		}
	}

	if inSingle {
		return nil, fmt.Errorf("unterminated single quote in curl command")
	}
	if inDouble {
		return nil, fmt.Errorf("unterminated double quote in curl command")
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens, nil
}

// normalizeLineContinuations replaces "\ <optional whitespace> <newline>" sequences
// with a single space so that multi-line cURL commands are parsed as one line.
func normalizeLineContinuations(s string) string {
	var b strings.Builder
	runes := []rune(s)
	i := 0
	for i < len(runes) {
		if runes[i] == '\\' && i+1 < len(runes) {
			// Consume optional trailing spaces/tabs after backslash
			j := i + 1
			for j < len(runes) && (runes[j] == ' ' || runes[j] == '\t') {
				j++
			}
			if j < len(runes) && (runes[j] == '\n' || runes[j] == '\r') {
				// Skip past the newline (handle \r\n too)
				if runes[j] == '\r' && j+1 < len(runes) && runes[j+1] == '\n' {
					j++
				}
				b.WriteRune(' ')
				i = j + 1
				continue
			}
		}
		b.WriteRune(runes[i])
		i++
	}
	return b.String()
}
