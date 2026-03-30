package route

import (
	dbbackupshandler "dag-observatory/dag-core/internal/interface/http/dbbackups/handler"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo, h *dbbackupshandler.Handler) {
	e.GET("/db-backups", h.ListDBBackups)
	e.POST("/db-backups/:backup_id\\:presign-download", h.PresignDBBackup)
}
