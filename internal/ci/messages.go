package ci

type PipelineLoadedMsg struct {
	Pipeline *Pipeline
	Slug     string
}

type JobsLoadedMsg struct {
	Jobs []Job
}

type LogDownloadedMsg struct {
	JobName string
	Path    string
}

type CIErrorMsg struct {
	Err error
}
