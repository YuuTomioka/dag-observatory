package dto

type DagRunRequest struct {
	Symbol string `json:"symbol"`
	Mode   string `json:"mode"`
}
