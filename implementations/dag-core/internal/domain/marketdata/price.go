package marketdata

type Price int64

// Raw は Price の内部表現（整数値）を返す。
func (p Price) Raw() int64 {
	return int64(p)
}

// NewPriceFromRaw は内部表現の整数値から Price を生成する。
func NewPriceFromRaw(raw int64) Price {
	return Price(raw)
}

// Neg は符号を反転した Price を返す。
func (p Price) Neg() Price {
	return Price(-int64(p))
}

// Abs は絶対値を返す。
func (p Price) Abs() Price {
	if p < 0 {
		return -p
	}
	return p
}

// Add は 2 つの Price を加算した結果を返す。
func (p Price) Add(q Price) Price {
	return Price(int64(p) + int64(q))
}

// Sub は 2 つの Price の差分（p - q）を返す。
func (p Price) Sub(q Price) Price {
	return Price(int64(p) - int64(q))
}

// MulInt は Price を整数倍率で乗算した結果を返す。
func (p Price) MulInt(n int64) Price {
	return Price(int64(p) * n)
}

// DivInt は Price を整数で除算した結果を返す。n が 0 の場合は panic する。
func (p Price) DivInt(n int64) Price {
	if n == 0 {
		panic("shared.Price.DivInt: division by zero")
	}
	return Price(int64(p) / n)
}

// Cmp は 2 つの Price を比較し、p<q なら -1、p>q なら 1、等しい場合は 0 を返す。
func (p Price) Cmp(q Price) int {
	switch {
	case p < q:
		return -1
	case p > q:
		return 1
	default:
		return 0
	}
}

// Eq は 2 つの Price が等しいかを返す。
func (p Price) Eq(q Price) bool { return p == q }

// Ne は 2 つの Price が等しくないかを返す。
func (p Price) Ne(q Price) bool { return p != q }

// Lt は p < q かを返す。
func (p Price) Lt(q Price) bool { return p < q }

// Le は p <= q かを返す。
func (p Price) Le(q Price) bool { return p <= q }

// Gt は p > q かを返す。
func (p Price) Gt(q Price) bool { return p > q }

// Ge は p >= q かを返す。
func (p Price) Ge(q Price) bool { return p >= q }
