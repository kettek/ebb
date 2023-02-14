package main

import (
	"fmt"
	"syscall/js"
)

func (g *Game) SystemInit() {
	fmt.Println("JS")
	hash := js.Global().Get("location").Get("hash")
	if !hash.IsUndefined() {
		m := hash.String()[1:]
		g.defaultMap = m
	}
}
