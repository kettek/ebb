package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game := &Game{
		images: make(map[string]*ebiten.Image),
		areas:  make(map[string]*Area),
	}
	game.Init()

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("ebb")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
