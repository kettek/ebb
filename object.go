package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Object struct {
	game *Game
	//
	title   string
	tag     string
	x, y    int
	saying  string
	image   *ebiten.Image
	noblock bool
	color   *color.RGBA
}

func (o *Object) Draw(screen *ebiten.Image, screenOpts *ebiten.DrawImageOptions) {
	opts := &ebiten.DrawImageOptions{}
	x := float64(o.x * o.image.Bounds().Dx())
	y := float64(o.y * o.image.Bounds().Dy())
	if o.color != nil {
		opts.ColorM.ScaleWithColor(*o.color)
	}
	opts.GeoM.Translate(x, y)
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
		if other := o.game.checkCollision(o, tx, ty); other != nil {
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
	case o.game.cochan <- fnc:
	default:
	}
	return <-done
}

func (o *Object) Step(x, y int) bool {
	done := make(chan bool)
	o.game.submit(func() bool {
		o.step(x, y)
		done <- true
		return true
	})
	return <-done
}

func (o *Object) step(x, y int) *Object {
	if other := o.game.checkCollision(o, o.x+x, o.y+y); other != nil {
		return other
	}
	o.x += x
	o.y += y
	return nil
}

func (o *Object) WalkTo(o2 *Object) bool {
	done := make(chan bool)
	steps := 0
	o.game.submit(func() bool {
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

		if other := o.game.checkCollision(o, x, y); other != nil {
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
	o.game.submit(func() bool {
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
	o.game.submit(func() bool {
		o.image = o.game.loadImage(s)
		done <- true
		return true
	})
	<-done
}

func (o *Object) SetBlocking(b bool) {
	done := make(chan bool)
	o.game.submit(func() bool {
		o.noblock = !b
		done <- true
		return true
	})
	<-done
}
