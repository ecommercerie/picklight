package poller

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Poll(url, jsonPath string, tlsSkipVerify bool) (int, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("poller: HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("poller: HTTP %d", resp.StatusCode)
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("poller: JSON decode error: %w", err)
	}

	value, err := extractPath(data, jsonPath)
	if err != nil {
		return 0, err
	}

	return toInt(value)
}

func extractPath(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("poller: path %q: expected object, got %T", path, current)
		}
		val, exists := m[part]
		if !exists {
			return nil, fmt.Errorf("poller: path %q: key %q not found", path, part)
		}
		current = val
	}
	return current, nil
}

func toInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case string:
		i, err := strconv.Atoi(n)
		if err != nil {
			return 0, fmt.Errorf("poller: cannot convert %q to int", n)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("poller: unexpected type %T for value", v)
	}
}
