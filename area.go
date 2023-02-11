package main

import (
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type Area struct {
	game            *Game
	mappe           *Map
	cochan          chan func() bool
	routines        []func() bool
	objects         []*Object
	traveledObjects map[string][2]int
	target          *Object
	created         bool
	lockedInput     bool
}

func (a *Area) Update() error {
	for done := false; !done; {
		select {
		case routine := <-a.cochan:
			a.routines = append(a.routines, routine)
		default:
			done = true
		}
	}
	routines := a.routines[:0]
	for _, r := range a.routines {
		if !r() {
			routines = append(routines, r)
		}
	}
	a.routines = routines

	return nil
}

func (a *Area) sortObjects() {
	sort.SliceStable(a.objects, func(i, j int) bool {
		return a.objects[i].z < a.objects[j].z
	})
}

func (a *Area) Draw(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	if a.target != nil {
		x := a.target.iterX
		y := a.target.iterY
		x -= float64(screen.Bounds().Dx() / 2)
		y -= float64(screen.Bounds().Dy() / 2)
		opts.GeoM.Translate(float64(-x), float64(-y))
	}
	for _, o := range a.objects {
		o.Draw(screen, opts)
	}

	for _, o := range a.objects {
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

	return
}

func (a *Area) submit(fnc func() bool) {
	select {
	case a.cochan <- fnc:
	default:
	}
}

func (a *Area) Delay(amount int) bool {
	ticks := 0
	done := make(chan bool)
	a.submit(func() bool {
		ticks++
		if ticks >= amount {
			done <- true
			return true
		}
		return false
	})
	return <-done
}

func (a *Area) NewObject(tag string, image string, color *color.RGBA) *Object {
	done := make(chan *Object)
	a.submit(func() bool {
		o := &Object{
			x:     -1,
			y:     -1,
			area:  a,
			tag:   tag,
			image: a.game.loadImage(image),
			color: color,
		}
		done <- o
		return true
	})
	return <-done
}

func (a *Area) getObject(tag string) *Object {
	for _, o := range a.objects {
		if o.tag == tag {
			return o
		}
	}
	return nil
}

func (a *Area) GetObject(tag string) *Object {
	done := make(chan *Object)
	a.submit(func() bool {
		done <- a.getObject(tag)
		return true
	})
	return <-done
}

func (a *Area) RemoveObject(tag string) *Object {
	done := make(chan *Object)
	a.submit(func() bool {
		o := a.getObject(tag)
		done <- o
		return true
	})
	return <-done
}

func (a *Area) removeObject(o *Object) *Object {
	for i, o2 := range a.objects {
		if o == o2 {
			a.objects = append(a.objects[:i], a.objects[i+1:]...)
			return o
		}
	}
	return nil
}

func (a *Area) PlaceObject(o *Object, x, y int) *Object {
	done := make(chan bool)
	a.submit(func() bool {
		o.area = a
		o.x = x
		o.y = y
		o.iterX = float64(o.x * o.image.Bounds().Dx())
		o.iterY = float64(o.y * o.image.Bounds().Dy())
		a.objects = append(a.objects, o)
		a.sortObjects()
		done <- true
		return true
	})
	<-done
	return o
}

func (a *Area) checkCollision(o *Object, x, y int, act string) (touch *Object) {
	for _, o2 := range a.objects {
		if o2.x == x && o2.y == y {
			blocked := !o2.noblock
			if o2.touch != nil {
				blocked = o2.touch(o2, o, act)
			}
			o.lastTouched = o2
			if blocked {
				return o2
			}
		}
	}
	return nil
}

func (a *Area) FollowObject(o *Object) {
	done := make(chan bool)
	a.submit(func() bool {
		a.target = o
		done <- true
		return true
	})
	<-done
}

func (a *Area) Exec(fnc func()) {
	done := make(chan bool)
	select {
	case a.cochan <- func() bool {
		fnc()
		done <- true
		return true
	}:
	default:
	}
	<-done
}

func (a *Area) Freeze() {
	done := make(chan bool)
	a.submit(func() bool {
		a.lockedInput = true
		done <- true
		return true
	})
	<-done
}

func (a *Area) Thaw() {
	done := make(chan bool)
	a.submit(func() bool {
		a.lockedInput = false
		done <- true
		return true
	})
	<-done
}

func (a *Area) Scene(fnc func()) {
	done := make(chan bool)
	a.submit(func() bool {
		a.lockedInput = true
		fnc()
		a.lockedInput = false
		done <- true
		return false
	})
	<-done
}

func (a *Area) Travel(s string, o *Object) {
	a.game.LoadArea(s, o)
}

func (a *Area) PreviousObjectPosition(s string) (x, y int, ok bool) {
	xy, ok := a.traveledObjects[s]
	return xy[0], xy[1], ok
}
