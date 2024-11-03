package main

import (
	"fmt"
	"io"
	"net/http"
)

func getRequest(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to send get request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to complete the request:\n\tStatus code: %d\n\tURL: %s\n\tResponse: %s", resp.StatusCode, url, body)
	}

	return body, nil
}
