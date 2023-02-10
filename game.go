package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	gameFont font.Face
)

type Game struct {
	cochan      chan func() bool
	routines    []func() bool
	objects     []*Object
	target      *Object
	images      map[string]*ebiten.Image
	lockedInput bool
}

func (g *Game) Init() {
	bytes, err := os.ReadFile("runescape-npc-chat.ttf")
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

	g.loadMap(Maps[0])
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

	if !g.lockedInput {
		// TODO
		if pl := g.getObject("player"); pl != nil {
			if inpututil.IsKeyJustPressed(ebiten.KeyA) {
				pl.step(-1, 0)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyD) {
				pl.step(1, 0)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyW) {
				pl.step(0, -1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyS) {
				pl.step(0, 1)
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	if g.target != nil {
		x := float64(g.target.x * g.target.image.Bounds().Dx())
		y := float64(g.target.y * g.target.image.Bounds().Dy())
		x -= float64(screen.Bounds().Dx() / 2)
		y -= float64(screen.Bounds().Dy() / 2)
		opts.GeoM.Translate(float64(-x), float64(-y))
	}
	for _, o := range g.objects {
		o.Draw(screen, opts)
	}

	for _, o := range g.objects {
		if o.saying != "" {
			bounds := text.BoundString(gameFont, o.saying)
			x := float64(o.x*o.image.Bounds().Dx()) - float64(bounds.Dx()/2)
			y := float64(o.y * o.image.Bounds().Dy())
			x += opts.GeoM.Element(0, 2)
			y += opts.GeoM.Element(1, 2)
			for i := -1; i < 2; i += 2 {
				text.Draw(screen, o.saying, gameFont, int(x)+i, int(y), color.Black)
				for j := -1; j < 2; j += 2 {
					text.Draw(screen, o.saying, gameFont, int(x)+i, int(y)+j, color.Black)
					text.Draw(screen, o.saying, gameFont, int(x)+i, int(y), color.Black)
				}
			}
			text.Draw(screen, o.saying, gameFont, int(x), int(y), o.color)
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("%f", ebiten.ActualTPS()))
	return
}

func (g *Game) Layout(w, h int) (int, int) {
	return w / 2, h / 2
}

func (g *Game) submit(fnc func() bool) {
	select {
	case g.cochan <- fnc:
	default:
	}
}

func (g *Game) Delay(amount int) bool {
	ticks := 0
	done := make(chan bool)
	g.submit(func() bool {
		ticks++
		if ticks >= amount {
			done <- true
			return true
		}
		return false
	})
	return <-done
}

func (g *Game) NewObject(tag string, image string, color *color.RGBA) *Object {
	done := make(chan *Object)
	g.submit(func() bool {
		o := &Object{
			x:     -1,
			y:     -1,
			game:  g,
			tag:   tag,
			image: g.loadImage(image),
			color: color,
		}
		done <- o
		return true
	})
	return <-done
}

func (g *Game) getObject(tag string) *Object {
	for _, o := range g.objects {
		if o.tag == tag {
			return o
		}
	}
	return nil
}

func (g *Game) GetObject(tag string) *Object {
	done := make(chan *Object)
	g.submit(func() bool {
		done <- g.getObject(tag)
		return true
	})
	return <-done
}

func (g *Game) PlaceObject(o *Object, x, y int) *Object {
	done := make(chan bool)
	g.submit(func() bool {
		o.game = g
		o.x = x
		o.y = y
		g.objects = append(g.objects, o)
		done <- true
		return true
	})
	<-done
	return o
}

func (g *Game) checkCollision(o *Object, x, y int) (touch *Object) {
	for _, o2 := range g.objects {
		if o2.x == x && o2.y == y && !o2.noblock {
			return o2
		}
	}
	return nil
}

func (g *Game) FollowObject(o *Object) {
	done := make(chan bool)
	g.submit(func() bool {
		o.game = g
		g.target = o
		done <- true
		return true
	})
	<-done
}

func (g *Game) Exec(fnc func()) {
	done := make(chan bool)
	select {
	case g.cochan <- func() bool {
		fnc()
		done <- true
		return true
	}:
	default:
	}
	<-done
}

func (g *Game) Freeze() {
	done := make(chan bool)
	g.submit(func() bool {
		g.lockedInput = true
		done <- true
		return true
	})
	<-done
}

func (g *Game) Thaw() {
	done := make(chan bool)
	g.submit(func() bool {
		g.lockedInput = false
		done <- true
		return true
	})
	<-done
}

func (g *Game) Scene(fnc func()) {
	done := make(chan bool)
	g.submit(func() bool {
		g.lockedInput = true
		fnc()
		g.lockedInput = false
		done <- true
		return false
	})
	<-done
}

func (g *Game) loadImage(s string) *ebiten.Image {
	if img, ok := g.images[s]; ok {
		return img
	}
	img, _, err := ebitenutil.NewImageFromFile(s + ".png")
	if err != nil {
		log.Fatal(err)
	}
	g.images[s] = img
	return img
}

// Maps
func (g *Game) loadMap(m *Map) {
	//lm := LiveMap{}
	g.objects = make([]*Object, 0)

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
						obj.game = g
						obj.x = x
						obj.y = y
						g.objects = append(g.objects, obj)
						//lm.EnsureSize(y, x)
						//lm.PlaceObject(obj, x, y)
					}
				}
			}
		}

		if m.load != nil {
			go func() {
				g.Freeze()
				m.load(g)
				g.Thaw()
			}()
		}
		m.created = true
	} else {
		// ???
	}
}
