package usecase

type RunWorkflowRequest struct {
	Symbol string
	Mode   string
	Bars   []float64
	RunID  string
}

type RunWorkflowResult struct {
	RunID                string
	Symbol               string
	Mode                 string
	TaskID               string
	TaskName             string
	Attempt              int
	EnqueueMode          bool
	MessagingDestination string
}
