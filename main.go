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
	screen      tcell.Screen
	ball        Ball
	p1          Player
	p2          Player
	targetScore int
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
		screen: screen,
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
				game.p2.Paddle.MoveUp(0)
			} else if event.Key() == tcell.KeyDown {
				game.p2.Paddle.MoveDown(height)
			} else if event.Rune() == 'w' {
				game.p1.Paddle.MoveUp(0)
			} else if event.Rune() == 's' {
				game.p1.Paddle.MoveDown(height)
			}
		}
	}
}

func (g *Game) Init() {

	width, _ := g.screen.Size()

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

	g.ball = ball
	g.p1 = p1
	g.p2 = p2
	g.targetScore = 9
}

func (g *Game) Loop() {
	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	g.screen.SetStyle(defStyle)

	paddleStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorWhite)

	width, height := g.screen.Size()

	for {
		g.screen.Clear()

		// Check game over
		if g.CheckGameOver() {
			drawSprite(g.screen, (width/2)-4, (height / 2), (width/2)+5, (height / 2), defStyle, "Game Over")
			drawSprite(g.screen, (width/2)-(len(g.WinnerString())/2), (height/2)+2, (width/2)+(len(g.WinnerString())+1/2), (height/2)+2, defStyle, g.WinnerString())
			g.screen.Show()
			return
		}

		// Update the ball position
		g.ball.Update()
		g.ball.CheckBoundingBox(width, height)

		drawSprite(g.screen, g.ball.Sprite.X, g.ball.Sprite.Y, g.ball.Sprite.X+g.ball.Sprite.Width, g.ball.Sprite.Y+g.ball.Sprite.Height, defStyle, g.ball.Draw())
		drawSprite(g.screen, g.p1.Paddle.Sprite.X, g.p1.Paddle.Sprite.Y, g.p1.Paddle.Sprite.X+g.p1.Paddle.Sprite.Width, g.p1.Paddle.Sprite.Y+g.p1.Paddle.Sprite.Height, paddleStyle, g.p1.Paddle.Draw())
		drawSprite(g.screen, g.p2.Paddle.Sprite.X, g.p2.Paddle.Sprite.Y, g.p2.Paddle.Sprite.X+g.p2.Paddle.Sprite.Width, g.p2.Paddle.Sprite.Y+g.p2.Paddle.Sprite.Height, paddleStyle, g.p2.Paddle.Draw())

		// Scores
		drawSprite(g.screen, 10, 1, 1, 1, defStyle, strconv.Itoa(g.p1.Score))
		drawSprite(g.screen, width-10, 1, 1, 1, defStyle, strconv.Itoa(g.p2.Score))

		// Game Title
		drawSprite(g.screen, (width/2)-3, 1, (width/2)-3+6, 1, defStyle, "GoPong")

		// Check for collisions with the Paddles
		if checkCollision(g.ball.Sprite, g.p1.Paddle.Sprite) {
			g.ball.reverseX()
			g.ball.reverseY()
		}

		if checkCollision(g.ball.Sprite, g.p2.Paddle.Sprite) {
			g.ball.reverseX()
			g.ball.reverseY()
		}

		// Check to see if the ball passes left or right edge of the screen
		if g.ball.Sprite.X <= 0 {
			g.p2.Score++
			g.ball.Reset(width/2, height/2, -1, 1)
		}

		if g.ball.Sprite.X >= width {
			g.p1.Score++
			g.ball.Reset(width/2, height/2, 1, 1)
		}

		time.Sleep(40 * time.Millisecond)
		g.screen.Show()
	}
}

func (g *Game) CheckGameOver() bool {
	if g.p1.Score == g.targetScore || g.p2.Score == g.targetScore {
		return true
	}
	return false
}

func (g *Game) WinnerString() string {
	if g.p1.Score > g.p2.Score {
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
