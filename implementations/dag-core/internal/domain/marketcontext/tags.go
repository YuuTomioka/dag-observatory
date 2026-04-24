package marketcontext

var signalTags = map[SignalID][]string{
	SignalTokyoLondonTransition: {
		"transition",
		"tokyo",
		"london",
	},
	SignalLondonNewYorkOverlap: {
		"overlap",
		"london",
		"newyork",
		"high_liquidity",
	},
	SignalNewYorkCloseTransition: {
		"transition",
		"newyork",
		"close",
		"unwind",
	},
}
