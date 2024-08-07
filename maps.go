package main

import (
	"image/color"
	"math"
	"math/rand"
)

type ThingCreatorFunc func(g *Game) *Object
type ThingCreatorFuncs map[rune]ThingCreatorFunc

type Map struct {
	title   string
	enter   func(a, previousArea *Area, triggering *Object, first bool)
	leave   func(a, previousArea *Area, triggering *Object)
	loaded  func(g *Game, a *Area)
	tiles   string
	things  ThingCreatorFuncs
	created bool
}

var GlobalThings = ThingCreatorFuncs{
	'@': func(g *Game) *Object {
		return &Object{
			Tag:   "player",
			Image: "character",
			Color: &color.RGBA{R: 255, G: 255, B: 0, A: 255},
			Z:     1,
		}
	},
	'#': func(g *Game) *Object {
		return &Object{
			Image: "woodwall",
			Color: &color.RGBA{R: 165, G: 42, B: 42, A: 255},
		}
	},
	'.': func(g *Game) *Object {
		return &Object{
			Image:   "grass",
			NoBlock: true,
		}
	},
	'*': func(g *Game) *Object {
		return &Object{
			Image: "tree",
		}
	},
	'/': func(g *Game) *Object {
		return &Object{
			Image:   "tree-hideable",
			NoBlock: true,
			Z:       10,
		}
	},
	'+': func(g *Game) *Object {
		return &Object{
			Tag:   "east door",
			Image: "door",
			Color: &color.RGBA{R: 145, G: 22, B: 22, A: 255},
			Touch: func(o *Object, toucher *Object, act string) (blocked bool) {
				if act == "interact" {
					o.NoBlock = !o.NoBlock
					if o.NoBlock {
						go o.SetImage("door-open")
					} else {
						go o.SetImage("door")
						go o.Say("*click*")
					}
					return true
				}
				if toucher.lastTouched != o && !o.NoBlock {
					go o.Say("*thump*")
				} else if toucher.lastTouched == o && !o.NoBlock {
					go o.SetImage("door-open")
					o.NoBlock = true
					return true
				}

				return !o.NoBlock
			},
		}
	},
	'T': func(g *Game) *Object {
		table := "table"
		if rand.Intn(2) == 1 {
			table = "table-food"
		}
		return &Object{
			Image: table,
			Color: &color.RGBA{R: 145, G: 22, B: 22, A: 255},
			Touch: func(o *Object, toucher *Object, act string) (blocked bool) {
				if o.image == g.loadImage("table-food") {
					if act == "" && toucher.lastTouched != o {
						go toucher.Say("food!")
						return true
					}
					if act == "interact" || toucher.lastTouched == o {
						go toucher.Say("*snarf*")
						o.image = g.loadImage("table")
					}
				}

				return true
			},
		}
	},
	'h': func(g *Game) *Object {
		return &Object{
			Image:   "chair-right",
			Color:   &color.RGBA{R: 145, G: 22, B: 22, A: 255},
			NoBlock: true,
		}
	},
	'n': func(g *Game) *Object {
		return &Object{
			Image:   "chair-left",
			Color:   &color.RGBA{R: 145, G: 22, B: 22, A: 255},
			NoBlock: true,
		}
	},
	'w': func(g *Game) *Object {
		return &Object{
			Image: "woodwallwindow",
			Color: &color.RGBA{R: 165, G: 42, B: 42, A: 255},
		}
	},
	'~': func(g *Game) *Object {
		return &Object{
			Image:   "water",
			Color:   &color.RGBA{R: 0, G: 64, B: 255, A: 255},
			NoBlock: true,
		}
	},
}

var Maps map[string]*Map = make(map[string]*Map)

func init() {
	Maps["start"] = &Map{
		title: "a start",
		tiles: `
   ##########**
   #        #.****
   #  1  hTnw.///***
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
					Tag:   "npc",
					Image: "character",
					Color: &color.RGBA{R: 255, G: 255, B: 255, A: 255},
					Z:     1,
				}
			},
			'E': func(g *Game) *Object {
				return &Object{
					Tag:   "east exit",
					Image: "exit",
					Color: &color.RGBA{R: 255, G: 255, B: 255, A: 255},
					Touch: func(o, toucher *Object, act string) (shouldBlock bool) {
						go o.area.Travel("east woods", toucher)
						return true
					},
				}
			},
		},
		enter: func(a *Area, prev *Area, triggering *Object, first bool) {
			player := triggering
			if player == nil {
				player = a.Object("player")
			}
			triggering = player
			if first {
				a.game.ControlObject(player)
				npc := a.Object("npc")
				door := a.Object("east door")
				a.FollowObject(player)
				a.Delay(60)
				npc.Say("hey, come here!")
				player.WalkTo(npc)
				a.Delay(20)
				npc.Say("have you heard of the high elves?")
				player.Say("no")
				npc.Say("me neither")
				// if it sucks... hit da bricks!!
				a.Freeze()
				npc2 := a.NewObject("npc 2", "character", &color.RGBA{R: 255, G: 0, B: 255, A: 255})
				door.SetImage("door-open")
				door.SetBlocking(false)
				a.PlaceObject(npc2, door.x, door.y)
				a.FollowObject(npc2)
				door.Say("*bang*")
				a.Delay(30)
				npc2.Step(-1, 0)
				a.Delay(30)
				door.SetImage("door")
				door.SetBlocking(true)
				a.Delay(30)
				npc2.Say("...greetings")
				a.Delay(10)
				npc2.WalkTo(npc)
				a.Delay(20)
				npc2.Say("I have heard of the high elves")
				//
				a.FollowObject(player)
				a.Thaw()
				a.Delay(300)
				npc.Say("They're a devious bunch")
				npc2.Say("You don't know the half of it")
			} else {
				x, y, ok := a.PreviousObjectPosition(triggering.Tag)
				if !ok {
					door := a.Object("east exit")
					x = door.x - 1
					y = door.y
				}
				a.Exec(func() {
					prev.removeObject(triggering)
				})
				a.PlaceObject(triggering, x, y)
			}
		},
	}
	Maps["east woods"] = &Map{
		title: "the east woods",
		tiles: `
****************
***********
****/////
*////*///
*/******
**/***       .,
*//**  / .  .,~,
<........ * ,~v~~,
*//**   //* ,~~~,
*/****  . ..,,~~,
*/*////*   ..,~,
*****/***    ,,
**********/   ,
*/
`,
		things: ThingCreatorFuncs{
			'<': func(g *Game) *Object {
				return &Object{
					Tag:   "west exit",
					Image: "exit",
					Color: &color.RGBA{R: 255, G: 255, B: 255, A: 255},
					Touch: func(o, toucher *Object, act string) (shouldBlock bool) {
						go o.area.Travel("start", toucher)
						return true
					},
				}
			},
			'v': func(g *Game) *Object {
				return &Object{
					Image: "whirlpool",
					Color: &color.RGBA{R: 64, G: 128, B: 255, A: 255},
					Touch: func(o, toucher *Object, act string) (shouldBlock bool) {
						go o.area.Travel("pool", toucher)
						return true
					},
				}
			},
			',': func(g *Game) *Object {
				opts := []string{
					"*shplut*",
					"*splort*",
				}
				return &Object{
					Image: "grass",
					Color: &color.RGBA{R: 64, G: 196, B: 255, A: 255},
					Touch: func(o, toucher *Object, act string) (shouldBlock bool) {
						sfx := opts[rand.Intn(len(opts))]
						go o.Say(sfx)
						return false
					},
					NoBlock: true,
				}
			},
		},
		enter: func(a *Area, prev *Area, triggering *Object, first bool) {
			x, y, ok := a.PreviousObjectPosition(triggering.Tag)
			if !ok {
				door := a.Object("west exit")
				x = door.x + 1
				y = door.y
			}
			a.Exec(func() {
				prev.removeObject(triggering)
			})
			a.PlaceObject(triggering, x, y)
		},
	}
	Maps["pool"] = &Map{
		title: "pool of whirling",
		tiles: `
 #########
 #~~~~~~~#
#~~~~~~~~~#
#~~~~~~^~~~#
 #~~~~~~~~##
 #~~~~~~~~~~#
 #~~~#####~~#
  #~#     #~~#
 ##~###### ##
#~~~~~~~f~#
 ##~~#####
   #~#
    #
`,
		things: ThingCreatorFuncs{
			'^': func(g *Game) *Object {
				return &Object{
					Tag:   "up exit",
					Image: "exit",
					Color: &color.RGBA{R: 64, G: 128, B: 255, A: 255},
					Touch: func(o, toucher *Object, act string) (shouldBlock bool) {
						go o.area.Travel("east woods", toucher)
						return true
					},
				}
			},
			'#': func(g *Game) *Object {
				return &Object{
					Image: "groundwall",
					Color: &color.RGBA{R: 96, G: 60, B: 12, A: 255},
				}
			},
			'f': func(g *Game) *Object {
				return &Object{
					Image: "froge",
					Color: &color.RGBA{R: 64, G: 255, B: 160, A: 255},
					Touch: func(o, toucher *Object, act string) (shouldBlock bool) {
						go o.Say("*ribbt*")
						return true
					},
				}
			},
		},
		enter: func(a *Area, prev *Area, triggering *Object, first bool) {
			player := triggering
			a.FollowObject(player)
			door := a.Object("up exit")
			a.Exec(func() {
				prev.removeObject(player)
			})
			a.PlaceObject(player, door.x-1, door.y)
		},
	}
	Maps["klb"] = &Map{
		title: "klb",
		tiles: `
      ####    ####
     #    #  #    #
    #  k   ##  b   #
    #              #
     #            #
      #          #
       #   p    #
        #      #
         #    #
          #  #
           ##
`,
		things: ThingCreatorFuncs{
			'#': func(g *Game) *Object {
				return &Object{
					Image: "heart",
					Color: &color.RGBA{R: 255, G: 105, B: 180, A: 255},
				}
			},
			'k': func(g *Game) *Object {
				return &Object{
					Tag:    "kit",
					Image:  "kit",
					Mirror: true,
					Color:  &color.RGBA{R: 204, G: 85, B: 0, A: 255},
				}
			},
			'b': func(g *Game) *Object {
				return &Object{
					Tag:   "birb",
					Image: "birb",
					Color: &color.RGBA{R: 249, G: 246, B: 238, A: 255},
				}
			},
			'p': func(g *Game) *Object {
				return &Object{
					Tag:     "point",
					Image:   "empty",
					NoBlock: true,
				}
			},
		},
		enter: func(a *Area, prev *Area, triggering *Object, first bool) {
			point := a.Object("point")
			a.FollowObject(point)
			a.Delay(60)
			birb := a.Object("birb")
			kit := a.Object("kit")
			birb.WalkTo(point)
			kit.WalkTo(point)
			kit.Step(1, 0)
			a.Delay(60)
			kit.Say("*kees*")
			o2 := a.NewObject("heart", "sprouts", &color.RGBA{R: 255, G: 0, B: 0, A: 255})
			a.PlaceObject(o2, kit.x, kit.y-1)
			a.Delay(60)
			birb.Say("*smoch*")
			o3 := a.NewObject("heart", "sprouts", &color.RGBA{R: 255, G: 0, B: 0, A: 255})
			a.PlaceObject(o3, birb.x, birb.y-1)
			a.Delay(30)

			t := 0.0
			r := 0.0
			for i := 0; i < 200; i++ {
				x := r * math.Cos(t)
				y := r * math.Sin(t)
				c := &color.RGBA{R: 255, G: 0, B: 0, A: 255}
				if i%2 == 0 {
					c.G = 255
					c.B = 255
				}
				o := a.NewObject("heart", "heart", c)
				a.PlaceObject(o, point.x+int(x), point.y+int(y))
				t += 0.3
				r += 0.3
			}
		},
	}
}
