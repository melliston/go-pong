package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Game struct {
	Screen      tcell.Screen
	Ball        Ball
	P1          Player
	P2          Player
	TargetScore int
}

type Player struct {
	Paddle Paddle
	Score  int
}

type Sprite struct {
	X      int
	Y      int
	Width  int
	Height int
}

type Ball struct {
	Sprite Sprite
	XSpeed int
	YSpeed int
}

type Paddle struct {
	Sprite Sprite
	YSpeed int
}

func main() {
	// Setup the screen.
	screen, err := tcell.NewScreen()

	if err != nil {
		log.Fatalf("%+v", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	// Setup the game
	game := Game{
		Screen: screen,
	}
	game.Init()

	// Run the game Loop in a goroutine so the input is non-blocking
	go game.Loop()

	_, height := screen.Size()

	for {
		switch event := screen.PollEvent().(type) {
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
				screen.Fini()
				os.Exit(0)
			} else if event.Key() == tcell.KeyUp {
				game.P2.Paddle.MoveUp(0)
			} else if event.Key() == tcell.KeyDown {
				game.P2.Paddle.MoveDown(height)
			} else if event.Rune() == 'w' {
				game.P1.Paddle.MoveUp(0)
			} else if event.Rune() == 's' {
				game.P1.Paddle.MoveDown(height)
			}
		}
	}
}

func (g *Game) Init() {

	width, _ := g.Screen.Size()

	p1 := Player{
		Score: 0,
		Paddle: Paddle{
			Sprite: Sprite{
				Width:  1,
				Height: 6,
				Y:      3,
				X:      5,
			},
			YSpeed: 3,
		},
	}

	p2 := Player{
		Score: 0,
		Paddle: Paddle{
			Sprite: Sprite{
				Width:  1,
				Height: 6,
				Y:      3,
				X:      width - 5,
			},
			YSpeed: 3,
		},
	}

	ball := Ball{
		Sprite: Sprite{
			X:      5,
			Y:      1,
			Width:  1,
			Height: 1,
		},
		XSpeed: 1,
		YSpeed: 1,
	}

	g.Ball = ball
	g.P1 = p1
	g.P2 = p2
	g.TargetScore = 9
}

func (g *Game) Loop() {
	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	g.Screen.SetStyle(defStyle)

	paddleStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorWhite)

	width, height := g.Screen.Size()

	for {
		g.Screen.Clear()

		// Check game over
		if g.CheckGameOver() {
			drawSprite(g.Screen, (width/2)-4, (height / 2), (width/2)+5, (height / 2), defStyle, "Game Over")
			drawSprite(g.Screen, (width/2)-(len(g.WinnerString())/2), (height/2)+2, (width/2)+(len(g.WinnerString())+1/2), (height/2)+2, defStyle, g.WinnerString())
			g.Screen.Show()
			return
		}

		// Update the ball position
		g.Ball.Update()
		g.Ball.CheckBoundingBox(width, height)

		drawSprite(g.Screen, g.Ball.Sprite.X, g.Ball.Sprite.Y, g.Ball.Sprite.X+g.Ball.Sprite.Width, g.Ball.Sprite.Y+g.Ball.Sprite.Height, defStyle, g.Ball.Draw())
		drawSprite(g.Screen, g.P1.Paddle.Sprite.X, g.P1.Paddle.Sprite.Y, g.P1.Paddle.Sprite.X+g.P1.Paddle.Sprite.Width, g.P1.Paddle.Sprite.Y+g.P1.Paddle.Sprite.Height, paddleStyle, g.P1.Paddle.Draw())
		drawSprite(g.Screen, g.P2.Paddle.Sprite.X, g.P2.Paddle.Sprite.Y, g.P2.Paddle.Sprite.X+g.P2.Paddle.Sprite.Width, g.P2.Paddle.Sprite.Y+g.P2.Paddle.Sprite.Height, paddleStyle, g.P2.Paddle.Draw())

		// Scores
		drawSprite(g.Screen, 10, 1, 1, 1, defStyle, strconv.Itoa(g.P1.Score))
		drawSprite(g.Screen, width-10, 1, 1, 1, defStyle, strconv.Itoa(g.P2.Score))

		// Game Title
		drawSprite(g.Screen, (width/2)-3, 1, (width/2)-3+6, 1, defStyle, "GoPong")

		// Check for collisions with the Paddles
		if checkCollision(g.Ball.Sprite, g.P1.Paddle.Sprite) {
			g.Ball.reverseX()
			g.Ball.reverseY()
		}

		if checkCollision(g.Ball.Sprite, g.P2.Paddle.Sprite) {
			g.Ball.reverseX()
			g.Ball.reverseY()
		}

		// Check to see if the ball passes left or right edge of the screen
		if g.Ball.Sprite.X <= 0 {
			g.P2.Score++
			g.Ball.Reset(width/2, height/2, -1, 1)
		}

		if g.Ball.Sprite.X >= width {
			g.P1.Score++
			g.Ball.Reset(width/2, height/2, 1, 1)
		}

		time.Sleep(40 * time.Millisecond)
		g.Screen.Show()
	}
}

func (g *Game) CheckGameOver() bool {
	if g.P1.Score == g.TargetScore || g.P2.Score == g.TargetScore {
		return true
	}
	return false
}

func (g *Game) WinnerString() string {
	if g.P1.Score > g.P2.Score {
		return "P1 Wins"
	}
	return "P2 Wins"
}

// drawSprite Draws a string easily - taken from the tcell getting started tutorial.
func drawSprite(screen tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, sprite string) {
	drawY := y1
	drawX := x1

	for _, r := range sprite {
		screen.SetContent(drawX, drawY, rune(r), nil, style)
		drawX++
		// Check see if we need to carry on to the next line if there is still more to be drawn.
		if drawX >= x2 {
			drawY++
			drawX = x1
		}
		if drawY > y2 {
			break
		}
	}
}

// CheckCollision Checks to see if there is a collision between two sprites
func checkCollision(sprite1, sprite2 Sprite) bool {
	if sprite1.X >= sprite2.X &&
		sprite1.X <= sprite2.X+sprite2.Width &&
		sprite1.Y >= sprite2.Y &&
		sprite1.Y <= sprite2.Y+sprite2.Height {
		return true
	}

	return false
}

func (b *Ball) Draw() string {
	return "\u25a2"
}

func (b *Ball) Update() {
	b.Sprite.X += b.XSpeed
	b.Sprite.Y += b.YSpeed
}

func (b *Ball) Reset(x, y, xSpeed, ySpeed int) {
	b.Sprite.X = x
	b.Sprite.Y = y
	b.XSpeed = xSpeed
	b.YSpeed = ySpeed
}

func (b *Ball) reverseX() {
	b.XSpeed *= -1
}

func (b *Ball) reverseY() {
	b.YSpeed *= -1
}

func (b *Ball) CheckBoundingBox(maxWidth int, maxHeight int) {
	if b.Sprite.Y <= 0 || b.Sprite.Y > maxHeight-1 {
		b.reverseY()
	}
}

func (p *Paddle) Draw() string {
	return strings.Repeat(" ", p.Sprite.Height)
}

// MoveUp Moves the Paddle Up,
// minHeight is passed so that it will allow for the game area to be different to the screen size.
func (p *Paddle) MoveUp(minHeight int) {
	if p.Sprite.Y > minHeight {
		p.Sprite.Y -= p.YSpeed
	}
}

// MoveDown Moves the Paddle Down
func (p *Paddle) MoveDown(maxHeight int) {
	if p.Sprite.Y < maxHeight-p.Sprite.Height {
		p.Sprite.Y += p.YSpeed
	}
}
