package response

type RunResponse struct {
	Status string `json:"status"`
	Symbol string `json:"symbol"`
	RunID  string `json:"run_id"`
}
