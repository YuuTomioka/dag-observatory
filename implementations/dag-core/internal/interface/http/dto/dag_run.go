package dto

type DagRunRequest struct {
	Symbol string    `json:"symbol"`
	Mode   string    `json:"mode"`
	Bars   []float64 `json:"bars,omitempty"`
}

type DagRunResponse struct {
	Status string `json:"status"`
	Symbol string `json:"symbol"`
	RunID  string `json:"run_id"`
}

type ErrorResponse struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	RunID  string `json:"run_id,omitempty"`
}
