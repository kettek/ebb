//go:build !js

package main

import "flag"

func (g *Game) SystemInit() {
	m := flag.String("map", "start", "default starting map")
	flag.Parse()

	g.defaultMap = *m
}
