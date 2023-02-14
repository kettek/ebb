package main

import (
	"syscall/js"
)

func (g *Game) SystemInit() {
	hash := js.Global().Get("location").Get("hash").String()
	if hash != "" {
		g.defaultMap = hash[1:]
	}
}
