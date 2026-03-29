package marketdata

type Volume int64

// Add は出来高を加算して新しい Volume を返す。
func (v Volume) Add(u Volume) Volume {
	return Volume(int64(v) + int64(u))
}
