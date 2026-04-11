package handler

import (
	"net/http"

	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/interface/http/dag/response"
	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

// ListRuns
// @Summary List workflow runs
// @Description Returns run-level read model items. Optional `partition` narrows results.
// @Tags dag
// @Produce json
// @Param partition query string false "Runtime partition filter"
// @Success 200 {object} response.ListRunsResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /runs [get]
func (h *Handler) ListRuns(c echo.Context) error {
	if h.listRuns == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "run read API not configured",
			Status: "error",
		})
	}
	items, err := h.listRuns.Execute(c.Request().Context(), usecase.ListRunsRequest{
		Partition: c.QueryParam("partition"),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
	return c.JSON(http.StatusOK, response.ListRunsResponse{Items: items})
}

// GetRun
// @Summary Get run detail
// @Description Returns one run-level read model by run id.
// @Tags dag
// @Produce json
// @Param run_id path string true "Run ID"
// @Success 200 {object} response.RunDetailResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /runs/{run_id} [get]
func (h *Handler) GetRun(c echo.Context) error {
	if h.getRun == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "run read API not configured",
			Status: "error",
		})
	}
	runID := c.Param("run_id")
	if runID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "run_id is required",
			Status: "error",
		})
	}
	item, ok, err := h.getRun.Execute(c.Request().Context(), usecase.GetRunRequest{RunID: runID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
	if !ok {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:  "run not found",
			Status: "error",
		})
	}
	return c.JSON(http.StatusOK, response.RunDetailResponse{Run: item})
}

// ListRunSteps
// @Summary List run steps
// @Description Returns node execution steps for one run.
// @Tags dag
// @Produce json
// @Param run_id path string true "Run ID"
// @Success 200 {object} response.ListRunStepsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /runs/{run_id}/steps [get]
func (h *Handler) ListRunSteps(c echo.Context) error {
	if h.listRunSteps == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "run read API not configured",
			Status: "error",
		})
	}
	runID := c.Param("run_id")
	if runID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "run_id is required",
			Status: "error",
		})
	}
	items, err := h.listRunSteps.Execute(c.Request().Context(), usecase.ListRunStepsRequest{
		RunID: runID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
	return c.JSON(http.StatusOK, response.ListRunStepsResponse{
		RunID: runID,
		Items: items,
	})
}

// GetRunNode
// @Summary Get run node execution detail
// @Description Returns one node execution record by run id and execution id (sequence number).
// @Tags dag
// @Produce json
// @Param run_id path string true "Run ID"
// @Param execution_id path string true "Execution ID (sequence number)"
// @Success 200 {object} response.RunNodeDetailResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /runs/{run_id}/nodes/{execution_id} [get]
func (h *Handler) GetRunNode(c echo.Context) error {
	if h.getRunNode == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "run read API not configured",
			Status: "error",
		})
	}
	runID := c.Param("run_id")
	if runID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "run_id is required",
			Status: "error",
		})
	}
	executionID := c.Param("execution_id")
	if executionID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "execution_id is required",
			Status: "error",
		})
	}
	item, ok, err := h.getRunNode.Execute(c.Request().Context(), usecase.GetRunNodeRequest{
		RunID:       runID,
		ExecutionID: executionID,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
	if !ok {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:  "run node not found",
			Status: "error",
		})
	}
	return c.JSON(http.StatusOK, response.RunNodeDetailResponse{
		RunID:       runID,
		ExecutionID: executionID,
		Node:        item,
	})
}
