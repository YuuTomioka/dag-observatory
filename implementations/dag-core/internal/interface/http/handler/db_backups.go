package handler

import (
	"net/http"
	"strings"
	"time"

	"dag-observatory/dag-core/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
)

// ListDBBackups
// @Summary List DB backups
// @Description Returns DB backup metadata filtered by env
// @Tags db-backups
// @Produce json
// @Param env query string true "Environment"
// @Param limit query int false "Max items (default 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.ListDBBackupsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /db-backups [get]
func (h *Handlers) ListDBBackups(c echo.Context) error {
	if h.dbBackups == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "db_backups repository not configured",
			Status: "error",
		})
	}
	env := c.QueryParam("env")
	if env == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "env is required",
			Status: "error",
		})
	}
	limit := parseInt(c.QueryParam("limit"), defaultLimit)
	offset := parseInt(c.QueryParam("offset"), defaultOffset)
	items, err := h.dbBackups.ListByEnv(c.Request().Context(), env, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
	resp := dto.ListDBBackupsResponse{Items: make([]dto.DBBackupItem, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, dto.DBBackupItem{
			BackupID:       item.BackupID,
			Env:            item.Env,
			BackupType:     item.BackupType,
			DBName:         item.DBName,
			SchemaVersion:  item.SchemaVersion,
			SizeBytes:      item.SizeBytes,
			ChecksumSHA256: item.ChecksumSHA256,
			CreatedAt:      item.CreatedAt,
			StartedAt:      item.StartedAt,
			FinishedAt:     item.FinishedAt,
			Status:         item.Status,
			Meta:           item.Meta,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// PresignDBBackup
// @Summary Presign DB backup download URL
// @Description Returns a time-limited URL for downloading a DB backup object
// @Tags db-backups
// @Produce json
// @Param backup_id path string true "Backup ID"
// @Param expires query int false "Expiry seconds (allowed: 60,300,900; default 900)"
// @Success 200 {object} dto.PresignDBBackupResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 410 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 501 {object} dto.ErrorResponse
// @Router /db-backups/{backup_id}:presign-download [post]
func (h *Handlers) PresignDBBackup(c echo.Context) error {
	if h.dbBackups == nil || h.presigner == nil {
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
			Error:  "db_backups presigner not configured",
			Status: "error",
		})
	}
	backupID := c.Param("backup_id")
	if backupID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:  "backup_id is required",
			Status: "error",
		})
	}
	backup, err := h.dbBackups.GetByID(c.Request().Context(), backupID)
	if err != nil {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:  "backup not found",
			Status: "error",
		})
	}
	status := strings.ToLower(backup.Status)
	if status == "expired" || status == "deleted" {
		return c.JSON(http.StatusGone, dto.ErrorResponse{
			Error:  "backup is expired or deleted",
			Status: "error",
		})
	}
	expiresSec := parseInt(c.QueryParam("expires"), 900)
	expires := clampExpires(expiresSec)
	url, err := h.presigner.PresignGetWithBucket(
		c.Request().Context(),
		backup.Bucket,
		backup.ObjectKey,
		time.Duration(expires)*time.Second,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:  err.Error(),
			Status: "error",
		})
	}
	return c.JSON(http.StatusOK, dto.PresignDBBackupResponse{
		BackupID:         backupID,
		URL:              url.String(),
		ExpiresInSeconds: int64(expires),
	})
}
