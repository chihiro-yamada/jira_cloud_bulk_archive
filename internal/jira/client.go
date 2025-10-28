package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Client represents a JIRA API client
type Client struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *http.Client
}

// Issue represents a JIRA issue
type Issue struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

// IssueFields represents fields in a JIRA issue
type IssueFields struct {
	Summary string `json:"summary"`
}

// SearchResult represents the result of a JQL search
type SearchResult struct {
	Issues        []Issue `json:"issues"`
	Total         int     `json:"total"`
	NextPageToken string  `json:"nextPageToken,omitempty"`
}

// NewClient creates a new JIRA API client
func NewClient(baseURL, email, apiToken string) *Client {
	return &Client{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchIssues searches for issues using JQL with the new search/jql endpoint
func (c *Client) SearchIssues(jql, nextPageToken string, maxResults int) (*SearchResult, error) {
	endpoint := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)

	params := url.Values{}
	params.Add("jql", jql)
	params.Add("maxResults", fmt.Sprintf("%d", maxResults))
	params.Add("fields", "summary")

	if nextPageToken != "" {
		params.Add("nextPageToken", nextPageToken)
	}

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	log.Printf("fullURL: %s\n", fullURL)

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ArchiveRequest represents the request body for bulk archiving
type ArchiveRequest struct {
	IssueIdsOrKeys []string `json:"issueIdsOrKeys"`
}

// ArchiveResponse represents the response from bulk archive API
type ArchiveResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
}

// ArchiveIssues archives multiple issues in a single API call
func (c *Client) ArchiveIssues(issueKeys []string) (*ArchiveResponse, error) {
	endpoint := fmt.Sprintf("%s/rest/api/3/issue/archive", c.baseURL)

	requestBody := ArchiveRequest{
		IssueIdsOrKeys: issueKeys,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, _ := io.ReadAll(resp.Body)

	// Archive API returns 200 or 204 on success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response if there's a body
	var archiveResp ArchiveResponse
	if len(body) > 0 {
		if err := json.Unmarshal(body, &archiveResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return &archiveResp, nil
}

// GetAllIssuesByLabel retrieves all issues with a specific label in a project
func (c *Client) GetAllIssuesByLabel(projectKey, label string) ([]Issue, error) {
	jql := fmt.Sprintf("project = %s AND labels = %s", projectKey, label)

	var allIssues []Issue
	nextPageToken := ""
	maxResults := 100 // JIRA's recommended batch size

	for {
		result, err := c.SearchIssues(jql, nextPageToken, maxResults)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, result.Issues...)

		// Check if there are more pages
		if result.NextPageToken == "" {
			break
		}

		nextPageToken = result.NextPageToken
	}

	return allIssues, nil
}
