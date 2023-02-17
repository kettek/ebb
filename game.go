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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	fs               multipath.FS
	images           map[string]*ebiten.Image
	areas            map[string]*Area
	currentArea      *Area
	activeAreas      []*Area
	controlledObject *Object
	cochan           chan func() bool
	routines         []func() bool
	defaultMap       string
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

	g.SystemInit()

	if _, ok := Maps[g.defaultMap]; !ok {
		g.defaultMap = "start"
	}

	g.loadArea(g.defaultMap, nil)
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

	for _, a := range g.activeAreas {
		if err := a.Update(); err != nil {
			panic(err)
		}
	}
	// FIXME: We need to tie the concept of input to a specific object and directly interface with it regardless of current area.
	if g.controlledObject != nil && g.controlledObject.area != nil {
		a := g.controlledObject.area
		if !a.lockedInput {
			// TODO
			pl := g.controlledObject
			act := ""
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				act = "interact"
			}

			if inpututil.IsKeyJustPressed(ebiten.KeyA) {
				pl.step(-1, 0, act)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyD) {
				pl.step(1, 0, act)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyW) {
				pl.step(0, -1, act)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyS) {
				pl.step(0, 1, act)
			}
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

func (g *Game) ActivateArea(a *Area) {
	done := make(chan *Area)
	select {
	case g.cochan <- func() bool {
		for _, a2 := range g.activeAreas {
			if a2 == a {
				done <- a
				return true
			}
		}
		g.activeAreas = append(g.activeAreas, a)
		done <- a
		return true
	}:
	default:
	}
	<-done
}

func (g *Game) DeactivateArea(a *Area) {
	done := make(chan *Area)
	select {
	case g.cochan <- func() bool {
		for i, a2 := range g.activeAreas {
			if a2 == a {
				g.activeAreas = append(g.activeAreas[:i], g.activeAreas[i+1:]...)
				break
			}
		}
		done <- a
		return true
	}:
	default:
	}
	<-done
}

func (g *Game) loadArea(s string, o *Object) *Area {
	area := g.areas[s]
	if area == nil {
		area = &Area{
			game:            g,
			cochan:          make(chan func() bool, 10),
			traveledObjects: make(map[string][2]int),
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
						obj.image = g.loadImage(obj.Image)
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
		go g.DeactivateArea(prev)
		if prev != nil && triggering != nil {
			prev.traveledObjects[triggering.Tag] = [2]int{triggering.x, triggering.y}
		}
		go g.ActivateArea(area)
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

func (g *Game) ControlObject(o *Object) {
	done := make(chan bool)
	select {
	case g.cochan <- func() bool {
		g.controlledObject = o
		done <- true
		return true
	}:
	default:
	}
	<-done
}
