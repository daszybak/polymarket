// Package httpclient is a http utils package.
package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func GetResource[T any](client *http.Client, baseURL, endpoint string, expectedStatusCodes []int) (T, error) {
	var zero T
	body, err := requestJSON(client, http.MethodGet, baseURL+endpoint, expectedStatusCodes, nil)
	if err != nil {
		return zero, err
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return zero, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

func PostResource[T any](client *http.Client, baseURL, endpoint string, data any, expectedStatusCodes []int) (T, error) {
	var zero T
	var reqBody io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return zero, fmt.Errorf("marshaling data for %s: %w", endpoint, err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	body, err := requestJSON(client, http.MethodPost, baseURL+endpoint, expectedStatusCodes, reqBody)
	if err != nil {
		return zero, err
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return zero, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

func requestJSON(client *http.Client, method, url string, expectedStatusCodes []int, reqBody io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating %s request for %s: %w", method, url, err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making %s request to %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading %s response from %s: %w", method, url, err)
	}

	if !slices.Contains(expectedStatusCodes, resp.StatusCode) {
		rendered := renderStatusCodes(expectedStatusCodes)
		return nil, fmt.Errorf("expected %s, got %d from %s %s: %s", rendered, resp.StatusCode, method, url, strings.TrimSpace(string(body)))
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	return body, nil
}

func renderStatusCodes(codes []int) string {
	codeReprs := make([]string, len(codes))
	for i, code := range codes {
		codeReprs[i] = strconv.Itoa(code)
	}
	return strings.Join(codeReprs, "/")
}
