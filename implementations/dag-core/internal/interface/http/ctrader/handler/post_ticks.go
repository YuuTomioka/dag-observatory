package handler

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	usecasepkg "dag-observatory/dag-core/internal/application/ctrader/usecase"
	marketdatarepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/interface/http/ctrader/request"
	"dag-observatory/dag-core/internal/interface/http/ctrader/response"
	"dag-observatory/dag-core/internal/interface/http/dto"
)

type PostTicksHandler struct {
	usecase usecasepkg.PostTicksUsecase
}

func NewPostTicksHandler(usecase usecasepkg.PostTicksUsecase) *PostTicksHandler {
	return &PostTicksHandler{
		usecase: usecase,
	}
}

func (h *PostTicksHandler) PostTicks(c echo.Context) error {
	var req request.PostTicksRequest
	if err := decodePostTicksRequest(c, &req); err != nil {
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			return c.JSON(httpErr.Code, dto.ErrorResponse{
				Error:  "unsupported content encoding",
				Status: "error",
			})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	ticks := make([]usecasepkg.TickInput, 0, len(req.Ticks))
	for i, item := range req.Ticks {
		parsedTime, err := time.Parse(time.RFC3339Nano, item.Time)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:  "ticks[" + strconvItoa(i) + "].time must be RFC3339Nano",
				Status: "error",
			})
		}
		ticks = append(ticks, usecasepkg.TickInput{
			Time: marketdata.NewUTCTime(parsedTime),
			Bid:  int64(item.Bid),
			Ask:  int64(item.Ask),
		})
	}

	result, err := h.usecase.Execute(c.Request().Context(), usecasepkg.PostTicksRequest{
		Symbol:     req.Symbol,
		PriceScale: req.PriceScale,
		Ticks:      ticks,
		RequestID:  req.RequestID,
		Day:        req.Day,
		BatchSeq:   req.BatchSeq,
	})
	if err != nil {
		return writePostTicksError(c, err)
	}

	return c.JSON(http.StatusOK, response.PostTicksResponse{
		Status:    "ok",
		Symbol:    result.Symbol,
		SymbolID:  result.SymbolID,
		Upserted:  result.Upserted,
		RequestID: req.RequestID,
		BatchSeq:  req.BatchSeq,
	})
}

func decodePostTicksRequest(c echo.Context, req *request.PostTicksRequest) error {
	encoding := strings.TrimSpace(strings.ToLower(c.Request().Header.Get(echo.HeaderContentEncoding)))
	switch {
	case encoding == "":
		return json.NewDecoder(c.Request().Body).Decode(req)
	case strings.Contains(encoding, "gzip"):
		reader, err := gzip.NewReader(c.Request().Body)
		if err != nil {
			return err
		}
		defer reader.Close()
		return json.NewDecoder(reader).Decode(req)
	default:
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "unsupported content encoding")
	}
}

func writePostTicksError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, marketdatarepository.ErrInvalidArgument):
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error(), Status: "error"})
	case errors.Is(err, marketdatarepository.ErrNotFound):
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not found", Status: "error"})
	case errors.Is(err, marketdatarepository.ErrConflict):
		return c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error(), Status: "error"})
	case errors.Is(err, marketdatarepository.ErrNotConfigured):
		return c.JSON(http.StatusNotImplemented, dto.ErrorResponse{Error: "marketdata usecases not configured", Status: "error"})
	default:
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error(), Status: "error"})
	}
}

func strconvItoa(n int) string {
	return strconv.Itoa(n)
}
