package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Object struct {
	area *Area
	//
	title        string
	tag          string
	x, y         int
	iterX, iterY float64
	saying       string
	image        *ebiten.Image
	noblock      bool
	color        *color.RGBA
	touch        func(o *Object, toucher *Object, act string) (shouldBlock bool)
	lastTouched  *Object
}

func (o *Object) Draw(screen *ebiten.Image, screenOpts *ebiten.DrawImageOptions) {
	opts := &ebiten.DrawImageOptions{}
	x := float64(o.x * o.image.Bounds().Dx())
	y := float64(o.y * o.image.Bounds().Dy())

	if o.iterX < x {
		o.iterX++
	} else if o.iterX > x {
		o.iterX--
	}
	if o.iterY < y {
		o.iterY++
	} else if o.iterY > y {
		o.iterY--
	}

	if o.color != nil {
		opts.ColorM.ScaleWithColor(*o.color)
	}
	opts.GeoM.Translate(o.iterX, o.iterY)
	opts.GeoM.Concat(screenOpts.GeoM)
	screen.DrawImage(o.image, opts)
}

func (o *Object) GoTo(x, y int) bool {
	done := make(chan bool)
	fnc := func() bool {
		tx := o.x
		ty := o.y
		if tx < x {
			tx++
		}
		if tx > x {
			tx--
		}
		if ty < y {
			ty++
		}
		if ty > y {
			ty--
		}
		if other := o.area.checkCollision(o, tx, ty, ""); other != nil {
			done <- false
			return true
		}
		o.x = tx
		o.y = ty
		if math.Abs(float64(o.x-x)) < 2 && math.Abs(float64(o.y-y)) < 2 {
			done <- true
			return true
		}
		return false
	}
	select {
	case o.area.cochan <- fnc:
	default:
	}
	return <-done
}

func (o *Object) Step(x, y int) bool {
	done := make(chan bool)
	o.area.submit(func() bool {
		o.step(x, y, "")
		done <- true
		return true
	})
	return <-done
}

func (o *Object) step(x, y int, act string) *Object {
	if other := o.area.checkCollision(o, o.x+x, o.y+y, act); other != nil {
		return other
	}
	o.x += x
	o.y += y
	return nil
}

func (o *Object) WalkTo(o2 *Object) bool {
	done := make(chan bool)
	steps := 0
	o.area.submit(func() bool {
		steps++
		if steps < 30 {
			return false
		}
		x := o.x
		y := o.y
		steps = 0
		if x < o2.x {
			x++
		} else if x > o2.x {
			x--
		}
		if y < o2.y {
			y++
		} else if y > o2.y {
			y--
		}

		if other := o.area.checkCollision(o, x, y, ""); other != nil {
			done <- false
			return true
		}
		o.x = x
		o.y = y

		if math.Abs(float64(o.x-o2.x)) < 2 && math.Abs(float64(o.y-o2.y)) < 2 {
			done <- true
			return true
		}
		return false
	})
	return <-done
}

func (o *Object) Say(s string) {
	done := make(chan bool)
	first := true
	ticks := 0
	o.area.submit(func() bool {
		if first {
			o.saying = s
			first = false
		}
		ticks++
		if ticks >= 20+len(s)*5 {
			o.saying = ""
			done <- true
			return true
		}
		return false
	})
	<-done
}

func (o *Object) SetImage(s string) {
	done := make(chan bool)
	o.area.submit(func() bool {
		o.image = o.area.game.loadImage(s)
		done <- true
		return true
	})
	<-done
}

func (o *Object) SetBlocking(b bool) {
	done := make(chan bool)
	o.area.submit(func() bool {
		o.noblock = !b
		done <- true
		return true
	})
	<-done
}
