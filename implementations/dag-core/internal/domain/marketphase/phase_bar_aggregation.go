package marketphase

import (
	"fmt"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

func AggregatePhaseBar(
	phase ResolvedPhase,
	symbolID marketdata.SymbolID,
	ticks []marketdata.Tick,
) (PhaseBar, bool, error) {
	if phase.Category != PhaseCategorySingle {
		return PhaseBar{}, false, fmt.Errorf("marketphase aggregate phase bar: single phase is required")
	}
	if !phase.Active {
		return PhaseBar{}, false, nil
	}

	var current *PhaseBar
	openTime := marketdata.NewUTCTime(phase.UTCStart)
	closeTime := marketdata.NewUTCTime(phase.UTCEnd)
	for _, tick := range ticks {
		if tick.Time.Before(openTime) || !tick.Time.Before(closeTime) {
			continue
		}

		mid := tick.Bid.Add(tick.Ask).DivInt(2)
		if current == nil {
			current = &PhaseBar{
				PhaseID:  phase.ID,
				Market:   phase.Market,
				Timezone: phase.Timezone,
				SymbolID: symbolID,
				OHLCV: marketdata.OHLCV{
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
				Source: PhaseBarSourceTickBidAskMid,
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

	if current == nil {
		return PhaseBar{}, false, nil
	}
	return *current, true, nil
}
