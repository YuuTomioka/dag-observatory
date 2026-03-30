package di

import (
	"fmt"

	artifactsrepository "dag-observatory/dag-core/internal/application/artifacts/repository"
	ctraderusecase "dag-observatory/dag-core/internal/application/ctrader/usecase"
	dbbackupsrepository "dag-observatory/dag-core/internal/application/dbbackups/repository"
	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	artifactsinfra "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/repository/artifacts"
	dbbackupsinfra "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/repository/dbbackups"
	miniostore "dag-observatory/dag-core/internal/infrastructure/storage/minio"
	httpif "dag-observatory/dag-core/internal/interface/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewEchoContainer(
	cfg Config,
	otelc *OTelContainer,
	dagRuntime *DAGRuntimeContainer,
	marketData *MarketDataContainer,
) (*echo.Echo, error) {
	e := echo.New()

	// OTel middleware (official instrumentation)
	e.Use(otelecho.Middleware(cfg.ServiceName))

	appLog := applog.New(cfg.AppLogLevel, cfg.AppLogOutput, otelc.AppLogger)

	if dagRuntime == nil {
		compiled, err := compileDefaultWorkflow(cfg)
		if err != nil {
			return nil, fmt.Errorf("dagruntime: compile default workflow: %w", err)
		}
		dagRuntime, err = NewDAGRuntimeContainer(cfg, otelc, compiled)
		if err != nil {
			return nil, err
		}
	}

	var artifactsRepo artifactsrepository.Reader
	var dbBackupsRepo dbbackupsrepository.Reader
	var presigner *miniostore.Presigner
	if cfg.TSDBURL != "" {
		client, err := tsdb.New(cfg.TSDBURL)
		if err != nil {
			return nil, err
		}
		artifactsRepo = artifactsinfra.NewRepository(client)
		dbBackupsRepo = dbbackupsinfra.NewRepository(client)
	}
	if cfg.MinIOEndpoint != "" {
		presigner = miniostore.NewPresigner(miniostore.Config{
			Endpoint:  cfg.MinIOEndpoint,
			AccessKey: cfg.MinIOAccessKey,
			SecretKey: cfg.MinIOSecretKey,
			Secure:    cfg.MinIOSecure == "true",
			Bucket:    cfg.MinIOBucket,
		})
	}

	var createSymbol *marketdatausecase.CreateSymbol
	var getSymbolByCode *marketdatausecase.GetSymbolByCode
	var listSymbols *marketdatausecase.ListSymbols
	var upsertTicks *marketdatausecase.UpsertTicks
	var getLatestTickBySymbol *marketdatausecase.GetLatestTickBySymbol
	var listTicksBySymbolAndRange *marketdatausecase.ListTicksBySymbolAndRange
	var postCTraderTicks *ctraderusecase.PostTicksUsecase
	if marketData != nil {
		createSymbol = marketData.CreateSymbol
		getSymbolByCode = marketData.GetSymbolByCode
		listSymbols = marketData.ListSymbols
		upsertTicks = marketData.UpsertTicks
		getLatestTickBySymbol = marketData.GetLatestTickBySymbol
		listTicksBySymbolAndRange = marketData.ListTicksBySymbolAndRange
		postCTraderTicks = &ctraderusecase.PostTicksUsecase{
			UnitOfWork: marketData.UnitOfWork,
		}
	}

	httpif.RegisterRoutes(e, httpif.Dependencies{
		IntentLog:                 otelc.IntentLog,
		AppLog:                    appLog,
		Tracer:                    otelc.Tracer,
		Metrics:                   otelc.Metrics,
		RunWorkflow:               dagRuntime.Usecase,
		ArtifactsRepo:             artifactsRepo,
		DBBackupsRepo:             dbBackupsRepo,
		Presigner:                 presigner,
		CreateSymbol:              createSymbol,
		GetSymbolByCode:           getSymbolByCode,
		ListSymbols:               listSymbols,
		UpsertTicks:               upsertTicks,
		GetLatestTickBySymbol:     getLatestTickBySymbol,
		ListTicksBySymbolAndRange: listTicksBySymbolAndRange,
		PostCTraderTicks:          postCTraderTicks,
	})

	return e, nil
}
