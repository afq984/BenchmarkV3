package main

import "context"

type Package interface {
	SetUp(ctx context.Context, buildDir string) error
}
