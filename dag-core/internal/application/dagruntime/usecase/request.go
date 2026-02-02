package usecase

type RunWorkflowRequest struct {
	Symbol string
	Mode   string
	RunID  string
}

type RunWorkflowResult struct {
	RunID       string
	Symbol      string
	Mode        string
	TaskID      string
	TaskName    string
	Attempt     int
	EnqueueMode bool
}
