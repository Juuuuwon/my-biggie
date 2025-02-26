package main

import (
	"embed"
)

//go:embed static/*
var staticContent embed.FS
