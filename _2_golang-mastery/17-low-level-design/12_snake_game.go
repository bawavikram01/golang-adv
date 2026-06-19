package lowleveldesign

import (
	"fmt"
	"math/rand"
	"time"
)

// =============================================================================
// LLD PROBLEM: SNAKE GAME
// =============================================================================
// Design the classic Snake game.
//
// Requirements:
// 1. Grid-based board
// 2. Snake moves in 4 directions
// 3. Food appears randomly
// 4. Snake grows when eating food
// 5. Game over on wall collision or self-collision
// 6. Score tracking
// 7. Increasing difficulty (speed)
//
// Design Patterns:
// - State Pattern: Game states (running, paused, game over)
// - Observer Pattern: Score updates, game events
// - Command Pattern: Direction changes as commands
// - Strategy Pattern: Food placement algorithms

// =============================================================================
// DOMAIN MODELS
// =============================================================================

type Point struct {
	X, Y int
}

func (p Point) Equals(other Point) bool {
	return p.X == other.X && p.Y == other.Y
}

type SnakeDirection int

const (
	Up SnakeDirection = iota
	Down
	Left
	Right
)

func (d SnakeDirection) Opposite() SnakeDirection {
	switch d {
	case Up:
		return Down
	case Down:
		return Up
	case Left:
		return Right
	case Right:
		return Left
	}
	return Up
}

type GameState int

const (
	GameRunning GameState = iota
	GamePaused
	GameOver
)

func (s GameState) String() string {
	return [...]string{"RUNNING", "PAUSED", "GAME_OVER"}[s]
}

// =============================================================================
// SNAKE
// =============================================================================

type Snake struct {
	Body      []Point
	Direction SnakeDirection
	Growing   bool
}

func NewSnake(start Point) *Snake {
	return &Snake{
		Body:      []Point{start, {start.X - 1, start.Y}, {start.X - 2, start.Y}},
		Direction: Right,
	}
}

func (s *Snake) Head() Point {
	return s.Body[0]
}

func (s *Snake) Move() {
	head := s.Head()
	var newHead Point

	switch s.Direction {
	case Up:
		newHead = Point{head.X, head.Y - 1}
	case Down:
		newHead = Point{head.X, head.Y + 1}
	case Left:
		newHead = Point{head.X - 1, head.Y}
	case Right:
		newHead = Point{head.X + 1, head.Y}
	}

	// Add new head
	s.Body = append([]Point{newHead}, s.Body...)

	// Remove tail unless growing
	if !s.Growing {
		s.Body = s.Body[:len(s.Body)-1]
	}
	s.Growing = false
}

func (s *Snake) Grow() {
	s.Growing = true
}

func (s *Snake) SetDirection(dir SnakeDirection) {
	// Prevent 180-degree turns (can't go directly backwards)
	if dir.Opposite() == s.Direction {
		return
	}
	s.Direction = dir
}

func (s *Snake) CollidesWithSelf() bool {
	head := s.Head()
	for _, part := range s.Body[1:] {
		if head.Equals(part) {
			return true
		}
	}
	return false
}

func (s *Snake) OccupiesPoint(p Point) bool {
	for _, part := range s.Body {
		if part.Equals(p) {
			return true
		}
	}
	return false
}

func (s *Snake) Length() int {
	return len(s.Body)
}

// =============================================================================
// FOOD PLACEMENT STRATEGY
// =============================================================================

type FoodPlacer interface {
	Place(width, height int, snake *Snake) Point
}

// Random placement — avoids snake body
type RandomFoodPlacer struct{}

func (f *RandomFoodPlacer) Place(width, height int, snake *Snake) Point {
	for {
		p := Point{
			X: rand.Intn(width),
			Y: rand.Intn(height),
		}
		if !snake.OccupiesPoint(p) {
			return p
		}
	}
}

// =============================================================================
// GAME BOARD
// =============================================================================

type Board struct {
	Width  int
	Height int
	Snake  *Snake
	Food   Point
}

func NewBoard(width, height int) *Board {
	return &Board{
		Width:  width,
		Height: height,
	}
}

func (b *Board) IsWallCollision(p Point) bool {
	return p.X < 0 || p.X >= b.Width || p.Y < 0 || p.Y >= b.Height
}

// Render returns a string representation of the board
func (b *Board) Render() string {
	grid := make([][]rune, b.Height)
	for y := 0; y < b.Height; y++ {
		grid[y] = make([]rune, b.Width)
		for x := 0; x < b.Width; x++ {
			grid[y][x] = '·'
		}
	}

	// Draw food
	if b.Food.X >= 0 && b.Food.X < b.Width && b.Food.Y >= 0 && b.Food.Y < b.Height {
		grid[b.Food.Y][b.Food.X] = '★'
	}

	// Draw snake body
	for i, p := range b.Snake.Body {
		if p.X >= 0 && p.X < b.Width && p.Y >= 0 && p.Y < b.Height {
			if i == 0 {
				grid[p.Y][p.X] = '▣' // head
			} else {
				grid[p.Y][p.X] = '■' // body
			}
		}
	}

	// Build string
	result := "┌" + repeatRune('─', b.Width) + "┐\n"
	for _, row := range grid {
		result += "│" + string(row) + "│\n"
	}
	result += "└" + repeatRune('─', b.Width) + "┘"
	return result
}

func repeatRune(r rune, n int) string {
	runes := make([]rune, n)
	for i := range runes {
		runes[i] = r
	}
	return string(runes)
}

// =============================================================================
// GAME EVENT SYSTEM (Observer Pattern)
// =============================================================================

type GameEventType int

const (
	EventFoodEaten GameEventType = iota
	EventCollision
	EventDirectionChange
	EventScoreUpdate
	EventGameOver
)

type GameEvent struct {
	Type    GameEventType
	Payload interface{}
}

type GameEventListener func(event GameEvent)

// =============================================================================
// GAME ENGINE
// =============================================================================

type SnakeGame struct {
	Board      *Board
	State      GameState
	Score      int
	Level      int
	FoodPlacer FoodPlacer
	Listeners  []GameEventListener
	TickRate   time.Duration
	moves      []SnakeDirection // command queue
}

func NewSnakeGame(width, height int) *SnakeGame {
	board := NewBoard(width, height)
	startPos := Point{X: width / 2, Y: height / 2}
	board.Snake = NewSnake(startPos)

	game := &SnakeGame{
		Board:      board,
		State:      GameRunning,
		Score:      0,
		Level:      1,
		FoodPlacer: &RandomFoodPlacer{},
		TickRate:   200 * time.Millisecond,
	}

	// Place initial food
	board.Food = game.FoodPlacer.Place(width, height, board.Snake)
	return game
}

func (g *SnakeGame) AddListener(listener GameEventListener) {
	g.Listeners = append(g.Listeners, listener)
}

func (g *SnakeGame) emit(event GameEvent) {
	for _, listener := range g.Listeners {
		listener(event)
	}
}

// Input: Queue a direction change (Command Pattern)
func (g *SnakeGame) Input(dir SnakeDirection) {
	if g.State != GameRunning {
		return
	}
	g.moves = append(g.moves, dir)
}

func (g *SnakeGame) Pause() {
	if g.State == GameRunning {
		g.State = GamePaused
	}
}

func (g *SnakeGame) Resume() {
	if g.State == GamePaused {
		g.State = GameRunning
	}
}

// Tick: One game step
func (g *SnakeGame) Tick() bool {
	if g.State != GameRunning {
		return g.State != GameOver
	}

	// Process queued direction change (only first one per tick)
	if len(g.moves) > 0 {
		g.Board.Snake.SetDirection(g.moves[0])
		g.moves = g.moves[1:]
		g.emit(GameEvent{Type: EventDirectionChange, Payload: g.Board.Snake.Direction})
	}

	// Move snake
	g.Board.Snake.Move()

	// Check wall collision
	head := g.Board.Snake.Head()
	if g.Board.IsWallCollision(head) {
		g.State = GameOver
		g.emit(GameEvent{Type: EventCollision, Payload: "wall"})
		g.emit(GameEvent{Type: EventGameOver, Payload: g.Score})
		return false
	}

	// Check self collision
	if g.Board.Snake.CollidesWithSelf() {
		g.State = GameOver
		g.emit(GameEvent{Type: EventCollision, Payload: "self"})
		g.emit(GameEvent{Type: EventGameOver, Payload: g.Score})
		return false
	}

	// Check food
	if head.Equals(g.Board.Food) {
		g.Board.Snake.Grow()
		g.Score += 10 * g.Level
		g.emit(GameEvent{Type: EventFoodEaten, Payload: g.Score})

		// Level up every 50 points
		newLevel := (g.Score / 50) + 1
		if newLevel > g.Level {
			g.Level = newLevel
			g.TickRate = g.TickRate * 9 / 10 // 10% faster
		}

		// Place new food
		g.Board.Food = g.FoodPlacer.Place(g.Board.Width, g.Board.Height, g.Board.Snake)
		g.emit(GameEvent{Type: EventScoreUpdate, Payload: g.Score})
	}

	return true
}

// Status returns game state info
func (g *SnakeGame) Status() string {
	return fmt.Sprintf("Score: %d | Level: %d | Length: %d | State: %s",
		g.Score, g.Level, g.Board.Snake.Length(), g.State)
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================

func ExampleSnakeGame() {
	game := NewSnakeGame(20, 15)

	// Add event listener
	game.AddListener(func(event GameEvent) {
		switch event.Type {
		case EventFoodEaten:
			fmt.Printf("  🍎 Food eaten! Score: %v\n", event.Payload)
		case EventGameOver:
			fmt.Printf("  💀 Game Over! Final score: %v\n", event.Payload)
		case EventCollision:
			fmt.Printf("  💥 Collision with: %v\n", event.Payload)
		}
	})

	// Simulate a game
	fmt.Println("=== Snake Game ===")
	fmt.Println(game.Board.Render())
	fmt.Println(game.Status())

	// Simulate moves
	directions := []SnakeDirection{Right, Right, Down, Down, Right, Right, Up}
	for _, dir := range directions {
		game.Input(dir)
		game.Tick()
	}

	fmt.Println("\nAfter moves:")
	fmt.Println(game.Board.Render())
	fmt.Println(game.Status())
}
