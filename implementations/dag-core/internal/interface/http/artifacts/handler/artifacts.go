package handler

import (
	"net/http"
	"strings"
	"time"

	"dag-observatory/dag-core/internal/interface/http/artifacts/response"
	"dag-observatory/dag-core/internal/interface/http/dto"
	httpshared "dag-observatory/dag-core/internal/interface/http/shared"

	"github.com/labstack/echo/v4"
)

const (
	defaultLimit  = 100
	defaultOffset = 0
)

// ListArtifactsByWorkflow
// @Summary List artifacts by workflow run
// @Description Returns artifacts associated with a workflow run id
// @Tags artifacts
// @Produce json
// @Param workflow_run_id path string true "Workflow run ID"
// @Param limit query int false "Max items (default 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} response.ListArtifactsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /workflow-runs/{workflow_run_id}/artifacts [get]
func (h *Handler) ListArtifactsByWorkflow(c echo.Context) error {
	if h.artifacts == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "artifacts repository not configured",
			Status: "error",
		})
	}
	workflowRunID := c.Param("workflow_run_id")
	if workflowRunID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "workflow_run_id is required",
			Status: "error",
		})
	}
	limit := httpshared.ParseInt(c.QueryParam("limit"), defaultLimit)
	offset := httpshared.ParseInt(c.QueryParam("offset"), defaultOffset)
	items, err := h.artifacts.ListByWorkflow(c.Request().Context(), workflowRunID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
	resp := response.ListArtifactsResponse{Items: make([]response.ArtifactItem, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, response.ArtifactItem{
			ArtifactID:    item.ArtifactID,
			WorkflowRunID: item.WorkflowRunID,
			TaskID:        item.TaskID,
			AttemptNo:     item.AttemptNo,
			Kind:          item.Kind,
			Filename:      item.Filename,
			ContentType:   item.ContentType,
			SizeBytes:     item.SizeBytes,
			Status:        item.Status,
			CreatedAt:     item.CreatedAt,
			Tags:          item.Tags,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// PresignArtifact
// @Summary Presign artifact download URL
// @Description Returns a time-limited URL for downloading an artifact
// @Tags artifacts
// @Produce json
// @Param artifact_id path string true "Artifact ID"
// @Param workflow_run_id query string true "Workflow run ID for access control"
// @Param expires query int false "Expiry seconds (allowed: 60,300,900; default 900)"
// @Success 200 {object} response.PresignResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 410 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /artifacts/{artifact_id}:presign-download [post]
func (h *Handler) PresignArtifact(c echo.Context) error {
	if h.artifacts == nil || h.presigner == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "artifacts presigner not configured",
			Status: "error",
		})
	}
	artifactID := c.Param("artifact_id")
	if artifactID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "artifact_id is required",
			Status: "error",
		})
	}
	workflowRunID := c.QueryParam("workflow_run_id")
	if workflowRunID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "workflow_run_id is required",
			Status: "error",
		})
	}
	artifact, err := h.artifacts.GetByID(c.Request().Context(), artifactID)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:  "artifact not found",
			Status: "error",
		})
	}
	if artifact.WorkflowRunID != workflowRunID {
		return c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:  "access denied",
			Status: "error",
		})
	}
	status := strings.ToLower(artifact.Status)
	if status == "expired" || status == "deleted" {
		return c.JSON(http.StatusGone, dto.ErrorResponse{
			Error:  "artifact is expired or deleted",
			Status: "error",
		})
	}
	expiresSec := httpshared.ParseInt(c.QueryParam("expires"), 900)
	expires := httpshared.ClampExpires(expiresSec)
	url, err := h.presigner.PresignGet(c.Request().Context(), artifact.ObjectKey, time.Duration(expires)*time.Second)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
	return c.JSON(http.StatusOK, response.PresignResponse{
		ArtifactID:       artifactID,
		URL:              url.String(),
		ExpiresInSeconds: int64(expires),
	})
}
