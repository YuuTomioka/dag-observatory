package response

type ResultReflectionRunResponse struct {
	Status     string `json:"status"`
	Entrypoint string `json:"entrypoint"`
	Partition  string `json:"partition"`
	RunID      string `json:"run_id"`
	Symbol     string `json:"symbol,omitempty"`
}
