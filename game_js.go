package main

import (
	"syscall/js"
)

func (g *Game) SystemInit() {
	hash := js.Global().Get("location").Get("hash")
	if !hash.IsUndefined() {
		m := hash.String()[1:]
		g.defaultMap = m
	}
}
