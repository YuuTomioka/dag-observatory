package usecase

import (
	"context"
	"os"
	"testing"
	"time"

	marketdatausecase "dag-observatory/dag-core/internal/application/marketdata/usecase"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/infrastructure/persistence/tsdb"
	marketdatarepo "dag-observatory/dag-core/internal/infrastructure/persistence/tsdb/repository/marketdata"
)

func TestPostTicksUsecasePersistsToTSDB(t *testing.T) {
	tsdbURL := os.Getenv("TSDB_URL")
	if tsdbURL == "" {
		t.Skip("TSDB_URL is required for TSDB integration test")
	}

	client, err := tsdb.New(tsdbURL)
	if err != nil {
		t.Fatalf("tsdb.New: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	uow := marketdatarepo.NewUnitOfWork(client)
	createSymbol := &marketdatausecase.CreateSymbol{UnitOfWork: uow}
	postTicks := &PostTicksUsecase{UnitOfWork: uow}
	listTicks := &marketdatausecase.ListTicksBySymbolAndRange{UnitOfWork: uow}

	code := "ITCTRADERPOSTTICKS"
	_, _ = client.Pool.Exec(ctx, `DELETE FROM tick WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
	_, _ = client.Pool.Exec(ctx, `DELETE FROM symbol WHERE code = $1`, code)
	defer func() {
		_, _ = client.Pool.Exec(ctx, `DELETE FROM tick WHERE symbol_id IN (SELECT id FROM symbol WHERE code = $1)`, code)
		_, _ = client.Pool.Exec(ctx, `DELETE FROM symbol WHERE code = $1`, code)
	}()

	symbol, err := createSymbol.Execute(ctx, marketdatausecase.CreateSymbolRequest{
		Code:        code,
		Base:        "ITC",
		Quote:       "JPY",
		PriceScale:  5,
		TickSizeRaw: 1,
		PipSizeRaw:  10,
	})
	if err != nil {
		t.Fatalf("create symbol: %v", err)
	}

	day := "2026-03-29"
	result, err := postTicks.Execute(ctx, PostTicksRequest{
		Symbol:     code,
		PriceScale: 5,
		Day:        day,
		BatchSeq:   0,
		Ticks: []TickInput{
			{Time: marketdata.MustParseUTCTime("2026-03-29T00:00:00.1234567Z"), Bid: 123456, Ask: 123460},
			{Time: marketdata.MustParseUTCTime("2026-03-29T00:00:01.1234567Z"), Bid: 123457, Ask: 123461},
		},
	})
	if err != nil {
		t.Fatalf("post ticks: %v", err)
	}
	if result.Upserted != 2 {
		t.Fatalf("expected 2 upserted ticks, got %d", result.Upserted)
	}

	items, err := listTicks.Execute(
		ctx,
		symbol.ID,
		marketdata.NewUTCTime(time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC)),
		marketdata.NewUTCTime(time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)),
	)
	if err != nil {
		t.Fatalf("list ticks: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 persisted ticks, got %d", len(items))
	}
	if items[0].Bid.Raw() != 123456 || items[0].Ask.Raw() != 123460 {
		t.Fatalf("unexpected first tick values: bid=%d ask=%d", items[0].Bid.Raw(), items[0].Ask.Raw())
	}
}
