package main

import (
	"image/color"
)

type ThingCreatorFunc func(g *Game) *Object
type ThingCreatorFuncs map[rune]ThingCreatorFunc

type Map struct {
	title   string
	load    func(g *Game)
	loaded  func(g *Game)
	tiles   string
	things  ThingCreatorFuncs
	created bool
}

var GlobalThings = ThingCreatorFuncs{
	'@': func(g *Game) *Object {
		return &Object{
			tag:   "player",
			image: g.loadImage("character"),
			color: &color.RGBA{R: 255, G: 255, B: 0, A: 255},
		}
	},
	'#': func(g *Game) *Object {
		return &Object{
			image: g.loadImage("woodwall"),
			color: &color.RGBA{R: 165, G: 42, B: 42, A: 255},
		}
	},
	'.': func(g *Game) *Object {
		return &Object{
			image:   g.loadImage("grass"),
			noblock: true,
		}
	},
	'*': func(g *Game) *Object {
		return &Object{
			image: g.loadImage("tree"),
		}
	},
	'/': func(g *Game) *Object {
		return &Object{
			image:   g.loadImage("tree-hideable"),
			noblock: true,
		}
	},
	'+': func(g *Game) *Object {
		return &Object{
			tag:   "east door",
			image: g.loadImage("door"),
			color: &color.RGBA{R: 145, G: 22, B: 22, A: 255},
		}
	},
	'E': func(g *Game) *Object {
		return &Object{
			tag:   "east exit",
			image: g.loadImage("exit"),
			color: &color.RGBA{R: 255, G: 255, B: 255, A: 255},
		}
	},
	'T': func(g *Game) *Object {
		return &Object{
			image: g.loadImage("table"),
			color: &color.RGBA{R: 145, G: 22, B: 22, A: 255},
		}
	},
	'h': func(g *Game) *Object {
		return &Object{
			image:   g.loadImage("chair-right"),
			color:   &color.RGBA{R: 145, G: 22, B: 22, A: 255},
			noblock: true,
		}
	},
	'n': func(g *Game) *Object {
		return &Object{
			image:   g.loadImage("chair-left"),
			color:   &color.RGBA{R: 145, G: 22, B: 22, A: 255},
			noblock: true,
		}
	},
	'w': func(g *Game) *Object {
		return &Object{
			image: g.loadImage("woodwallwindow"),
			color: &color.RGBA{R: 165, G: 42, B: 42, A: 255},
		}
	},
}

var Maps = map[int]*Map{
	0: {
		title: "a start",
		tiles: `
   ##########**
   #        #.****
   #    1   #.///***
 *.#        #..///***
...whTn     +.......E
...#        #...////*
*..#     hTnw..///***
...whTn     #///***
 *.#     @  #****
   ######w###
      /...../
       */./
         *
		`,
		things: ThingCreatorFuncs{
			'1': func(g *Game) *Object {
				return &Object{
					tag:   "npc",
					image: g.loadImage("character"),
					color: &color.RGBA{R: 255, G: 255, B: 255, A: 255},
				}
			},
		},
		load: func(g *Game) {
			npc := g.GetObject("npc")
			player := g.GetObject("player")
			door := g.GetObject("east door")
			g.FollowObject(player)
			g.Delay(60)
			npc.Say("hey, come here!")
			player.WalkTo(npc)
			g.Delay(20)
			npc.Say("have you heard of the high elves?")
			player.Say("no")
			npc.Say("me neither")
			npc2 := g.NewObject("npc 2", "character", &color.RGBA{R: 255, G: 0, B: 255, A: 255})
			door.SetImage("door-open")
			door.SetBlocking(false)
			g.PlaceObject(npc2, door.x, door.y)
			g.FollowObject(npc2)
			door.Say("*bang*")
			g.Delay(30)
			npc2.Step(-1, 0)
			g.Delay(30)
			door.SetImage("door")
			door.SetBlocking(true)
			g.Delay(30)
			npc2.Say("...greetings")
			g.Delay(10)
			npc2.WalkTo(npc)
			g.Delay(20)
			npc2.Say("I have heard of the high elves")
			//
			g.FollowObject(player)
			g.Thaw()
			g.Delay(300)
			npc.Say("They're a devious bunch")
			npc2.Say("You don't know the half of it")
		},
	},
}
