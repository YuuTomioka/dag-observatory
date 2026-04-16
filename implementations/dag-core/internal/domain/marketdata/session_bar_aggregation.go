package marketdata

func AggregateSessionBar(
	session Session,
	sessionDate SessionDate,
	symbolID SymbolID,
	ticks []Tick,
) (SessionBar, bool, error) {
	openTime, closeTime, err := session.WindowForDate(sessionDate)
	if err != nil {
		return SessionBar{}, false, err
	}

	var current *SessionBar
	for _, tick := range ticks {
		if tick.Time.Before(openTime) || !tick.Time.Before(closeTime) {
			continue
		}

		mid := tick.Bid.Add(tick.Ask).DivInt(2)
		if current == nil {
			current = &SessionBar{
				SessionCode: session.Code,
				SessionDate: sessionDate,
				SymbolID:    symbolID,
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

	if current == nil {
		return SessionBar{}, false, nil
	}
	return *current, true, nil
}
