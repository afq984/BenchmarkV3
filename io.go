package main

import (
	"context"
	"io"
)

type reader struct {
	ctx context.Context
	r   io.Reader
}

func readerContext(ctx context.Context, r io.Reader) io.Reader {
	return reader{ctx, r}
}

func (r reader) Read(p []byte) (int, error) {
	err := r.ctx.Err()
	if err != nil {
		return 0, err
	}
	n, err := r.r.Read(p)
	if err != nil {
		return n, err
	}
	return n, r.ctx.Err()
}
