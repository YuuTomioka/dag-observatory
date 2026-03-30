package http

import (
	artifactsrepository "dag-observatory/dag-core/internal/application/artifacts/repository"
	ctraderusecase "dag-observatory/dag-core/internal/application/ctrader/usecase"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	dbbackupsrepository "dag-observatory/dag-core/internal/application/dbbackups/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/application/observability/port"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"
	miniostore "dag-observatory/dag-core/internal/infrastructure/storage/minio"
	artifactshandler "dag-observatory/dag-core/internal/interface/http/artifacts/handler"
	artifactsroute "dag-observatory/dag-core/internal/interface/http/artifacts/route"
	ctraderhandler "dag-observatory/dag-core/internal/interface/http/ctrader/handler"
	ctraderroute "dag-observatory/dag-core/internal/interface/http/ctrader/route"
	daghandler "dag-observatory/dag-core/internal/interface/http/dag/handler"
	dagroute "dag-observatory/dag-core/internal/interface/http/dag/route"
	dbbackupshandler "dag-observatory/dag-core/internal/interface/http/dbbackups/handler"
	dbbackupsroute "dag-observatory/dag-core/internal/interface/http/dbbackups/route"
	healthhandler "dag-observatory/dag-core/internal/interface/http/health/handler"
	healthroute "dag-observatory/dag-core/internal/interface/http/health/route"
	marketdatahandler "dag-observatory/dag-core/internal/interface/http/marketdata/handler"
	marketdataroute "dag-observatory/dag-core/internal/interface/http/marketdata/route"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

type Dependencies struct {
	IntentLog     port.IntentLog
	AppLog        *applog.Logger
	Tracer        trace.Tracer
	Metrics       *metrics.Instruments
	RunWorkflow   *usecase.RunWorkflow
	ArtifactsRepo artifactsrepository.Reader
	Presigner     *miniostore.Presigner
	DBBackupsRepo dbbackupsrepository.Reader

	CreateSymbol              *marketdatausecase.CreateSymbol
	GetSymbolByCode           *marketdatausecase.GetSymbolByCode
	ListSymbols               *marketdatausecase.ListSymbols
	UpsertTicks               *marketdatausecase.UpsertTicks
	GetLatestTickBySymbol     *marketdatausecase.GetLatestTickBySymbol
	ListTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
	BackfillTimeframeBars     *marketdatausecase.BackfillTimeframeBars
	PostCTraderTicks          *ctraderusecase.PostTicksUsecase
}

func RegisterRoutes(e *echo.Echo, d Dependencies) {
	hh := healthhandler.New()
	dh := daghandler.New(daghandler.Dependencies{
		AppLog:      d.AppLog,
		Tracer:      d.Tracer,
		Metrics:     d.Metrics,
		RunWorkflow: d.RunWorkflow,
	})
	ah := artifactshandler.New(artifactshandler.Dependencies{
		ArtifactsRepo: d.ArtifactsRepo,
		Presigner:     d.Presigner,
	})
	bh := dbbackupshandler.New(dbbackupshandler.Dependencies{
		DBBackupsRepo: d.DBBackupsRepo,
		Presigner:     d.Presigner,
	})
	mh := marketdatahandler.New(marketdatahandler.Dependencies{
		CreateSymbol:              d.CreateSymbol,
		GetSymbolByCode:           d.GetSymbolByCode,
		ListSymbols:               d.ListSymbols,
		UpsertTicks:               d.UpsertTicks,
		GetLatestTickBySymbol:     d.GetLatestTickBySymbol,
		ListTicksBySymbolAndRange: d.ListTicksBySymbolAndRange,
		BackfillTimeframeBars:     d.BackfillTimeframeBars,
	})
	var postTicksUsecase ctraderusecase.PostTicksUsecase
	if d.PostCTraderTicks != nil {
		postTicksUsecase = *d.PostCTraderTicks
	}
	ch := ctraderhandler.NewPostTicksHandler(postTicksUsecase)

	healthroute.Register(e, hh)
	dagroute.Register(e, dh)
	artifactsroute.Register(e, ah)
	dbbackupsroute.Register(e, bh)
	marketdataroute.Register(e, mh)
	ctraderroute.Register(e, ch)
}
