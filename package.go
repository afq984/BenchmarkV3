package main

type Package interface {
	SetUp(buildDir string) error
}
