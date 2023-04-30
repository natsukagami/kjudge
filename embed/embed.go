package embed

import "embed"

//go:embed assets/* templates/*
var Content embed.FS
