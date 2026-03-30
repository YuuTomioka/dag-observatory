package dto

type ErrorResponse struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	RunID  string `json:"run_id,omitempty"`
}
