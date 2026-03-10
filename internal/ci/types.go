package ci

import "time"

type Pipeline struct {
	ID        string    `json:"id"`
	Number    int       `json:"number"`
	State     string    `json:"state"`
	VCS       VCSInfo   `json:"vcs"`
	CreatedAt time.Time `json:"created_at"`
}

type VCSInfo struct {
	Branch   string `json:"branch"`
	Revision string `json:"revision"`
}

type PipelineResponse struct {
	Items    []Pipeline `json:"items"`
	NextPage string     `json:"next_page_token"`
}

type Workflow struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type WorkflowResponse struct {
	Items    []Workflow `json:"items"`
	NextPage string     `json:"next_page_token"`
}

type Job struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	JobNumber int        `json:"job_number"`
	StartedAt *time.Time `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at"`
}

type JobResponse struct {
	Items    []Job  `json:"items"`
	NextPage string `json:"next_page_token"`
}

type Artifact struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

type ArtifactResponse struct {
	Items    []Artifact `json:"items"`
	NextPage string     `json:"next_page_token"`
}

// Duration returns the job duration, or zero if timestamps are missing.
func (j Job) Duration() time.Duration {
	if j.StartedAt == nil || j.StoppedAt == nil {
		return 0
	}
	return j.StoppedAt.Sub(*j.StartedAt)
}
