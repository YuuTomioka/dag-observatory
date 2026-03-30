package shared

import "strconv"

func ParseInt(raw string, def int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

func ClampExpires(v int) int {
	switch v {
	case 60, 300, 900:
		return v
	default:
		return 900
	}
}
