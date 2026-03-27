package tsdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	Pool *pgxpool.Pool
}

func New(url string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("tsdb: TSDB_URL is required")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, err
	}
	return &Client{Pool: pool}, nil
}

func (c *Client) Close() {
	if c == nil || c.Pool == nil {
		return
	}
	c.Pool.Close()
}
