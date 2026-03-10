package ci

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: "https://circleci.com/api/v2",
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Circle-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *Client) LatestPipeline(slug, branch string) (*Pipeline, error) {
	path := fmt.Sprintf("/project/%s/pipeline?branch=%s",
		url.PathEscape(slug), url.QueryEscape(branch))
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var resp PipelineResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no pipelines found for branch %q", branch)
	}
	return &resp.Items[0], nil
}

func (c *Client) Workflows(pipelineID string) ([]Workflow, error) {
	path := fmt.Sprintf("/pipeline/%s/workflow", pipelineID)
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var resp WorkflowResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) Jobs(workflowID string) ([]Job, error) {
	path := fmt.Sprintf("/workflow/%s/job", workflowID)
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var resp JobResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) Artifacts(slug string, jobNumber int) ([]Artifact, error) {
	path := fmt.Sprintf("/project/%s/%d/artifacts",
		url.PathEscape(slug), jobNumber)
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var resp ArtifactResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) DownloadArtifact(artifactURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", artifactURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Circle-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
