package game

import (
	"context"
	"time"

	termbox "github.com/nsf/termbox-go"
)

// PlayerChar defines how should each player entries rendered
var PlayerChar = []string{"\\//\\", "/\\\\/"}
var piece = '█'

// Game defines the state of the game and arena details
type Game struct {
	State         [][]int
	CurrentPlayer int
	Width         int
	Height        int
	offsetX       int
	offsetY       int
	Winner        int
	players       []string
	wonState      [][]int
	ctx           context.Context
	Cancel        context.CancelFunc
}

func getEmptyState(height, width int) [][]int {
	state := [][]int{}
	for i := 0; i < height; i++ {
		rowState := make([]int, width)
		state = append(state, rowState)
	}
	return state
}

// NewGame return a new instance of game
func NewGame(width, height int) *Game {

	ctx, cancel := context.WithCancel(context.Background())

	game := &Game{
		Width:         width,
		Height:        height,
		State:         getEmptyState(height, width),
		CurrentPlayer: 1,
		ctx:           ctx,
		Cancel:        cancel,
	}
	game.getOffset()
	return game
}

func (g *Game) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for x := 0; x < g.Width; x++ {
		termbox.SetCell(g.offsetX+x*3, g.offsetY-2, rune(48+x), termbox.ColorYellow, termbox.ColorDefault)
	}
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			g.setContent(x, y, true)
		}
	}
	termbox.Flush()
}

func (g *Game) getOffset() {
	sw, sh := termbox.Size()
	g.offsetX = (sw - g.Width*3) / 2
	g.offsetY = (sh - g.Height*2) / 2
}

func (g *Game) setContent(x, y int, show bool) {
	for i := 0; i < 2; i++ {
		for j := 0; j < 1; j++ {
			ch, fore, bg := getplayerDisplayPropsLarge(g.State[y][x], i, j)
			if show {
				termbox.SetCell((g.offsetX + x*3 + i), (g.offsetY + y*2 + j), ch, fore, bg)
			} else {
				termbox.SetCell((g.offsetX + x*3 + i), (g.offsetY + y*2 + j), ch, fore, termbox.ColorDefault)
			}
		}
	}

	// ch, fore, bg := getplayerDisplayProps(g.State[y][x])
	// termbox.SetCell((g.offsetX + x*3), (g.offsetY + y*2), ch, fore, bg)
}
func getplayerDisplayProps(player int) (rune, termbox.Attribute, termbox.Attribute) {
	if player == 1 {
		return piece, termbox.ColorRed, termbox.ColorDefault
	}
	if player == 2 {
		return piece, termbox.ColorBlue, termbox.ColorDefault
	}
	return piece, termbox.ColorDefault, termbox.ColorBlack
}
func getplayerDisplayPropsLarge(player, col, row int) (rune, termbox.Attribute, termbox.Attribute) {
	if player == 1 {
		return piece, termbox.ColorRed, termbox.ColorDefault
	}
	if player == 2 {
		return piece, termbox.ColorBlue, termbox.ColorDefault
	}
	return ' ', termbox.ColorDefault, termbox.ColorBlack
}

func (g *Game) addEntry(col, player int) {
	if g.Winner != 0 {
		return
	}
	if col < 0 || col > g.Width-1 {
		return
	}
	if g.State[0][col] != 0 {
		return
	}
	i := 0
	for ; i < g.Height && g.State[i][col] == 0; i++ {
		g.State[i][col] = player
		if i > 0 {
			g.State[i-1][col] = 0
		}
		g.Draw()
		time.Sleep(30 * time.Millisecond)
	}
	g.togglePlayer()
	if g.isWon(col, i-1, player) {
		g.declareWinner()
	}
}

func (g *Game) isWon(col, row, player int) bool {
	// check vertically
	wonState := make([][]int, 0)
	count := 0
	for i := row; i < g.Width && g.State[i][col] == player; i++ {
		count++
		wonState = append(wonState, []int{i, col})
	}
	if count >= 4 {
		g.Winner = player
		g.wonState = wonState
		return true
	}
	// check horizontally left
	wonState = make([][]int, 0)
	count = 0
	i := col
	for ; i > 0 && g.State[row][i] == player; i-- {
	}
	i++
	// check horizontally right
	for ; i < g.Width && g.State[row][i] == player; i++ {
		wonState = append(wonState, []int{row, i})
		count++
	}
	if count >= 4 {
		g.Winner = player
		g.wonState = wonState
		return true
	}

	// check positive diagonal
	wonState = make([][]int, 0)
	i = 0
	count = 0
	for ; col-i > 0 && row-i > 0 && g.State[row-i][col-i] == player; i++ {
	}
	i--

	for ; col-i < g.Width && row-i < g.Height && g.State[row-i][col-i] == player; i-- {
		wonState = append(wonState, []int{row - i, col - i})
		count++
	}
	if count >= 4 {
		g.wonState = wonState
		g.Winner = player
		return true
	}

	// check negative diagonal
	wonState = make([][]int, 0)
	i = 0
	count = 0
	for ; col+i < g.Width && row-i > 0 && g.State[row-i][col+i] == player; i++ {
	}
	i--
	for ; col+i > 0 && row-i < g.Height && g.State[row-i][col+i] == player; i-- {
		wonState = append(wonState, []int{row - i, col + i})
		count++
	}
	if count >= 4 {
		g.wonState = wonState
		g.Winner = player
		return true
	}
	return false
}

func (g *Game) declareWinner() {
	go func(g *Game) {
		show := true
	inf_loop:
		for {
			select {
			case <-g.ctx.Done():
				g = NewGame(g.Width, g.Height)
				g.Draw()
				break inf_loop
			default:
			}
			for i := 0; i < 4; i++ {
				g.setContent(g.wonState[i][1], g.wonState[i][0], show)
			}
			g.renderText("% won the game", show)
			show = !show
			termbox.Flush()
			time.Sleep(time.Millisecond * 500)
		}
	}(g)
}

func (g *Game) renderText(text string, show bool) {
	x, y := g.offsetX+8, g.offsetY+9

	for i := 0; i < len(text); i++ {
		if !show {
			termbox.SetCell(x, y+i, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
			termbox.SetCell(x+i, y, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
			continue
		}
		if text[i] == byte('%') {
			ch, fore, bg := getplayerDisplayProps(g.Winner)
			termbox.SetCell(x, y+i, ch, fore, bg)
		} else {
			termbox.SetCell(x+i, y, rune(text[i]), termbox.ColorWhite, termbox.ColorDefault)
		}
	}
	termbox.Flush()
}

func (g *Game) togglePlayer() {
	if g.CurrentPlayer == 1 {
		g.CurrentPlayer = 2
		return
	}
	g.CurrentPlayer = 1
}

func (g *Game) Input(col int) {
	g.addEntry(col, g.CurrentPlayer)
}

func (g *Game) generateSplashContent() {
	g.State[1][5] = 1
	g.State[2][4] = 1
	g.State[2][5] = 2
	g.State[3][4] = 1
	g.State[3][5] = 2
	g.State[4][3] = 2
	g.State[4][5] = 1
	g.State[5][3] = 2
	g.State[5][5] = 1
	g.State[6][2] = 2
	g.State[6][3] = 1
	g.State[6][4] = 2
	g.State[6][5] = 1
	g.State[6][6] = 2
	g.State[6][7] = 1
	g.State[7][5] = 2
	g.State[8][5] = 2

	g.Draw()
	g.Winner = 2
}

func (g *Game) SplashScreen() {
	go func(g *Game) {
		show := true
	inf_loop:
		for {
			select {
			case <-g.ctx.Done():
				g = NewGame(g.Width, g.Height)
				g.Draw()
				break inf_loop
			default:
			}
			if show {
				g.generateSplashContent()
			} else {
				g.State = getEmptyState(g.Height, g.Width)
			}
			g.Draw()
			show = !show
			time.Sleep(time.Millisecond * 500)
		}
	}(g)
}
