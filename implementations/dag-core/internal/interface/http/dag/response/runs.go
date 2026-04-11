package response

import "dag-observatory/dag-core/internal/application/dagruntime/usecase"

type ListRunsResponse struct {
	Items []usecase.RunView `json:"items"`
}

type RunDetailResponse struct {
	Run usecase.RunView `json:"run"`
}

type ListRunStepsResponse struct {
	RunID string                `json:"run_id"`
	Items []usecase.RunStepView `json:"items"`
}

type RunNodeDetailResponse struct {
	RunID       string              `json:"run_id"`
	ExecutionID string              `json:"execution_id"`
	Node        usecase.RunStepView `json:"node"`
}
