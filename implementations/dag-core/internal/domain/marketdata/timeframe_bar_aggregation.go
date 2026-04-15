package marketdata

import (
	"fmt"
	"time"
)

// BackfillRange expands [from, to) into the bar-open range that covers all
// windows overlapping the requested tick range.
func (c TimeframeCode) BackfillRange(from, to UTCTime) (UTCTime, UTCTime, error) {
	if from.IsZero() || to.IsZero() {
		return UTCTime{}, UTCTime{}, fmt.Errorf("marketdata timeframe backfill range: from/to are required")
	}
	if !from.Before(to) {
		return UTCTime{}, UTCTime{}, fmt.Errorf("marketdata timeframe backfill range: from must be before to")
	}
	open, _ := c.Window(from)
	lastInstant := to.Add(-time.Nanosecond)
	_, close := c.Window(lastInstant)
	return open, close, nil
}

func AggregateTimeframeBars(
	timeframeCode TimeframeCode,
	symbolID SymbolID,
	ticks []Tick,
) []TimeframeBar {
	if len(ticks) == 0 {
		return nil
	}

	bars := make([]TimeframeBar, 0)
	var current *TimeframeBar

	for _, tick := range ticks {
		openTime, closeTime := timeframeCode.Window(tick.Time)
		mid := tick.Bid.Add(tick.Ask).DivInt(2)
		if current == nil || !current.Opentime.Equal(openTime) {
			if current != nil {
				bars = append(bars, *current)
			}
			current = &TimeframeBar{
				TimeframeCode: timeframeCode,
				SymbolID:      symbolID,
				OHLCV: OHLCV{
					Opentime:  openTime,
					Closetime: closeTime,
					Open:      mid,
					High:      tick.Ask,
					Hightime:  tick.Time,
					Low:       tick.Bid,
					Lowtime:   tick.Time,
					Close:     mid,
					Volume:    1,
				},
				Source: TimeframeBarSourceTickBidAskMid,
			}
			continue
		}

		if tick.Ask.Gt(current.High) || tick.Ask.Eq(current.High) {
			current.High = tick.Ask
			current.Hightime = tick.Time
		}
		if tick.Bid.Lt(current.Low) || tick.Bid.Eq(current.Low) {
			current.Low = tick.Bid
			current.Lowtime = tick.Time
		}
		current.Close = mid
		current.Volume = current.Volume.Add(1)
	}

	if current != nil {
		bars = append(bars, *current)
	}
	return bars
}
