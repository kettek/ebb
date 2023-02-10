package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kettek/go-multipath/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed data/*
var embedFS embed.FS
var (
	gameFont font.Face
)

type Game struct {
	fs          multipath.FS
	images      map[string]*ebiten.Image
	areas       map[string]*Area
	currentArea *Area
}

func (g *Game) Init() {
	g.fs.InsertFS(os.DirFS("data"), multipath.FirstPriority)
	sub, err := fs.Sub(embedFS, "data")
	if err != nil {
		log.Fatal(err)
	}
	g.fs.InsertFS(sub, multipath.LastPriority)

	bytes, err := g.fs.ReadFile("runescape-npc-chat.ttf")
	if err != nil {
		log.Fatal(err)
	}
	// font
	tt, err := opentype.Parse(bytes)
	if err != nil {
		//log.Fatal(err)
	}

	const dpi = 72
	gameFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    32,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	g.areas["start"] = g.loadMap(Maps["start"])
	g.currentArea = g.areas["start"]
}

func (g *Game) Update() error {
	if g.currentArea != nil {
		if err := g.currentArea.Update(); err != nil {
			panic(err)
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.currentArea != nil {
		g.currentArea.Draw(screen)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%f", ebiten.ActualTPS()))
	return
}

func (g *Game) Layout(w, h int) (int, int) {
	return w / 2, h / 2
}

func (g *Game) loadImage(s string) *ebiten.Image {
	if img, ok := g.images[s]; ok {
		return img
	}
	img, _, err := ebitenutil.NewImageFromFileSystem(g.fs, s+".png")
	if err != nil {
		log.Fatal(err)
	}
	g.images[s] = img
	return img
}

// Maps
func (g *Game) loadMap(m *Map) *Area {
	area := &Area{
		game:   g,
		cochan: make(chan func() bool, 10),
	}

	if !m.created {
		lines := strings.Split(m.tiles, "\n")[1:]
		for y, line := range lines {
			for x, r := range line {
				ctor, ok := m.things[r]
				if !ok {
					ctor, ok = GlobalThings[r]
				}
				if ok {
					obj := ctor(g)
					if obj != nil {
						obj.area = area
						obj.x = x
						obj.y = y
						area.objects = append(area.objects, obj)
					}
				}
			}
		}

		if m.load != nil {
			go func() {
				area.Freeze()
				m.load(g, area)
				area.Thaw()
			}()
		}
		m.created = true
	} else {
		// ???
	}
	return area
}
