package di

import "testing"

func TestNewMarketDataContainerWithoutTSDBURL(t *testing.T) {
	t.Parallel()

	container, err := NewMarketDataContainer(Config{})
	if err != nil {
		t.Fatalf("new marketdata container: %v", err)
	}
	if container != nil {
		t.Fatalf("expected nil container when TSDB_URL is empty")
	}
}
