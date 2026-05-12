/*
 * =============================================================
 * LLD CASE STUDY 4: TIC-TAC-TOE
 * =============================================================
 *
 * REQUIREMENTS:
 *   - 2 players, N×N board (default 3×3)
 *   - Players alternate turns placing X or O
 *   - Win by filling a row, column, or diagonal
 *   - Detect win, draw, and invalid moves
 *
 * DESIGN PATTERNS USED:
 *   - Strategy (win-checking, player input)
 *   - State (game state: PLAYING, WIN, DRAW)
 *
 * This is a VERY common LLD interview question.
 */

import java.util.*;

public class TicTacToe {

    public static void main(String[] args) {

        Player p1 = new Player("Alice", 'X');
        Player p2 = new Player("Bob", 'O');

        Game game = new Game(3, p1, p2);

        // Simulate a game
        game.makeMove(0, 0);  // X
        game.makeMove(1, 1);  // O
        game.makeMove(0, 1);  // X
        game.makeMove(2, 2);  // O
        game.makeMove(0, 2);  // X wins (top row)

        System.out.println("\n=== GAME 2 (Draw) ===");

        Game game2 = new Game(3, p1, p2);
        game2.makeMove(0, 0);  // X
        game2.makeMove(0, 1);  // O
        game2.makeMove(0, 2);  // X
        game2.makeMove(1, 0);  // O
        game2.makeMove(1, 1);  // X
        game2.makeMove(2, 2);  // O
        game2.makeMove(1, 2);  // X
        game2.makeMove(2, 0);  // O — but let's see...
        game2.makeMove(2, 1);  // X — draw if no winner
    }
}

// ═══════════════════════════════════════════════════════════════
// ENUMS
// ═══════════════════════════════════════════════════════════════
enum GameStatus {
    PLAYING, WIN, DRAW
}

// ═══════════════════════════════════════════════════════════════
// PLAYER
// ═══════════════════════════════════════════════════════════════
class Player {
    private final String name;
    private final char symbol;

    public Player(String name, char symbol) {
        this.name = name;
        this.symbol = symbol;
    }

    public String getName() { return name; }
    public char getSymbol() { return symbol; }

    @Override
    public String toString() { return name + "(" + symbol + ")"; }
}

// ═══════════════════════════════════════════════════════════════
// BOARD
// ═══════════════════════════════════════════════════════════════
class Board {
    private final int size;
    private final char[][] grid;
    private int movesCount;

    public Board(int size) {
        this.size = size;
        this.grid = new char[size][size];
        this.movesCount = 0;
        for (char[] row : grid) Arrays.fill(row, ' ');
    }

    public boolean placeMove(int row, int col, char symbol) {
        if (row < 0 || row >= size || col < 0 || col >= size) {
            System.out.println("  ✗ Invalid position!");
            return false;
        }
        if (grid[row][col] != ' ') {
            System.out.println("  ✗ Position already taken!");
            return false;
        }
        grid[row][col] = symbol;
        movesCount++;
        return true;
    }

    public boolean isFull() {
        return movesCount == size * size;
    }

    public boolean checkWin(int row, int col, char symbol) {
        // Check row
        boolean rowWin = true;
        for (int c = 0; c < size; c++) {
            if (grid[row][c] != symbol) { rowWin = false; break; }
        }
        if (rowWin) return true;

        // Check column
        boolean colWin = true;
        for (int r = 0; r < size; r++) {
            if (grid[r][col] != symbol) { colWin = false; break; }
        }
        if (colWin) return true;

        // Check main diagonal
        if (row == col) {
            boolean diagWin = true;
            for (int i = 0; i < size; i++) {
                if (grid[i][i] != symbol) { diagWin = false; break; }
            }
            if (diagWin) return true;
        }

        // Check anti-diagonal
        if (row + col == size - 1) {
            boolean antiDiagWin = true;
            for (int i = 0; i < size; i++) {
                if (grid[i][size - 1 - i] != symbol) { antiDiagWin = false; break; }
            }
            if (antiDiagWin) return true;
        }

        return false;
    }

    public void display() {
        System.out.println("  ┌───┬───┬───┐");
        for (int r = 0; r < size; r++) {
            System.out.print("  │");
            for (int c = 0; c < size; c++) {
                System.out.print(" " + grid[r][c] + " │");
            }
            System.out.println();
            if (r < size - 1) System.out.println("  ├───┼───┼───┤");
        }
        System.out.println("  └───┴───┴───┘");
    }
}

// ═══════════════════════════════════════════════════════════════
// GAME — Orchestrates everything
// ═══════════════════════════════════════════════════════════════
class Game {
    private final Board board;
    private final Player[] players;
    private int currentPlayerIndex;
    private GameStatus status;

    public Game(int size, Player p1, Player p2) {
        this.board = new Board(size);
        this.players = new Player[]{p1, p2};
        this.currentPlayerIndex = 0;
        this.status = GameStatus.PLAYING;
        System.out.println("=== TIC-TAC-TOE: " + p1 + " vs " + p2 + " ===");
    }

    public void makeMove(int row, int col) {
        if (status != GameStatus.PLAYING) {
            System.out.println("  Game is already over!");
            return;
        }

        Player current = players[currentPlayerIndex];
        System.out.println("  " + current + " plays at (" + row + "," + col + ")");

        if (!board.placeMove(row, col, current.getSymbol())) {
            return;  // invalid move, try again (same player)
        }

        board.display();

        if (board.checkWin(row, col, current.getSymbol())) {
            status = GameStatus.WIN;
            System.out.println("  🏆 " + current.getName() + " WINS!");
            return;
        }

        if (board.isFull()) {
            status = GameStatus.DRAW;
            System.out.println("  🤝 It's a DRAW!");
            return;
        }

        // Switch player
        currentPlayerIndex = 1 - currentPlayerIndex;
    }

    public GameStatus getStatus() { return status; }
}

/*
 * CLASS DIAGRAM:
 * ─────────────────────────────────────────────────────────────
 *   Game
 *     ├── Board
 *     │     └── char[][] grid
 *     ├── Player[] (2 players)
 *     │     ├── name, symbol
 *     └── GameStatus (enum)
 *
 * INTERVIEW TIPS:
 * ✦ Start by clarifying: board size? 2 players or more?
 * ✦ Separate Board logic from Game orchestration (SRP)
 * ✦ Win checking is O(N) per move (check row+col+diags)
 * ✦ For O(1) win checking: maintain row/col/diag sum counters
 * ✦ Extensible: add AI player, larger boards, undo
 *
 * COMPILE & RUN:
 *   javac TicTacToe.java && java TicTacToe
 */
