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
	cochan      chan func() bool
	routines    []func() bool
}

func (g *Game) Init() {
	g.cochan = make(chan func() bool, 10)

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

	g.loadArea("start", nil)
}

func (g *Game) Update() error {
	for done := false; !done; {
		select {
		case routine := <-g.cochan:
			g.routines = append(g.routines, routine)
		default:
			done = true
		}
	}
	routines := g.routines[:0]
	for _, r := range g.routines {
		if !r() {
			routines = append(routines, r)
		}
	}
	g.routines = routines

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

func (g *Game) LoadArea(s string, o *Object) *Area {
	done := make(chan *Area)
	select {
	case g.cochan <- func() bool {
		done <- g.loadArea(s, o)
		return true
	}:
	default:
	}
	return <-done
}

func (g *Game) loadArea(s string, o *Object) *Area {
	area := g.areas[s]
	if area == nil {
		area = &Area{
			game:   g,
			cochan: make(chan func() bool, 10),
		}
	}

	m := Maps[s]
	if m == nil {
		panic("no map")
	}

	area.mappe = m

	if !area.created {
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
	}

	area.sortObjects()

	go func(area *Area, prev *Area, first bool, triggering *Object) {
		if prev != nil && prev.mappe.leave != nil {
			prev.mappe.leave(prev, area, triggering)
		}
		if area.mappe.enter != nil {
			area.mappe.enter(area, prev, triggering, first)
		}
		// ... send area chan with new area...
	}(area, g.currentArea, !area.created, o)

	area.created = true
	g.areas[s] = area
	g.currentArea = area

	return area
}
