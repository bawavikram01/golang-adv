/*
 * =============================================================
 * LLD CASE STUDY 8: SNAKE AND LADDER
 * =============================================================
 *
 * REQUIREMENTS:
 *   - N×N board with snakes and ladders
 *   - Multiple players take turns
 *   - Dice roll determines movement
 *   - Snake: move down | Ladder: move up
 *   - First player to reach/cross last cell wins
 *
 * DESIGN PATTERNS USED:
 *   - Strategy (different dice types)
 *   - Template Method (game loop)
 *   - Observer (announce events)
 */

import java.util.*;

public class SnakeAndLadder {

    public static void main(String[] args) {
        System.out.println("=== SNAKE AND LADDER ===\n");

        // Build board
        Board board = new Board(100);
        // Snakes: head → tail (go DOWN)
        board.addSnake(16, 6);
        board.addSnake(47, 26);
        board.addSnake(49, 11);
        board.addSnake(56, 53);
        board.addSnake(62, 19);
        board.addSnake(64, 60);
        board.addSnake(87, 24);
        board.addSnake(93, 73);
        board.addSnake(95, 75);
        board.addSnake(98, 78);
        // Ladders: bottom → top (go UP)
        board.addLadder(1, 38);
        board.addLadder(4, 14);
        board.addLadder(9, 31);
        board.addLadder(21, 42);
        board.addLadder(28, 84);
        board.addLadder(36, 44);
        board.addLadder(51, 67);
        board.addLadder(71, 91);
        board.addLadder(80, 100);

        // Set up game
        Dice dice = new StandardDice(1);   // 1 die
        Game game = new Game(board, dice);
        game.addPlayer("Alice");
        game.addPlayer("Bob");
        game.addPlayer("Charlie");

        game.play();
    }
}

// ═══════════════════════════════════════════════════════════════
// BOARD
// ═══════════════════════════════════════════════════════════════
class Board {
    private final int size;
    private final Map<Integer, Integer> snakes = new HashMap<>();   // head → tail
    private final Map<Integer, Integer> ladders = new HashMap<>();  // bottom → top

    public Board(int size) { this.size = size; }

    public void addSnake(int head, int tail) {
        if (head <= tail) throw new IllegalArgumentException("Snake head must be > tail");
        snakes.put(head, tail);
    }

    public void addLadder(int bottom, int top) {
        if (bottom >= top) throw new IllegalArgumentException("Ladder bottom must be < top");
        ladders.put(bottom, top);
    }

    public int getFinalPosition(int position) {
        if (snakes.containsKey(position)) {
            System.out.println("    🐍 Snake! " + position + " → " + snakes.get(position));
            return snakes.get(position);
        }
        if (ladders.containsKey(position)) {
            System.out.println("    🪜 Ladder! " + position + " → " + ladders.get(position));
            return ladders.get(position);
        }
        return position;
    }

    public int getSize() { return size; }
}

// ═══════════════════════════════════════════════════════════════
// DICE — Strategy Pattern
// ═══════════════════════════════════════════════════════════════
interface Dice {
    int roll();
}

class StandardDice implements Dice {
    private final int numDice;
    private final Random random = new Random();

    public StandardDice(int numDice) { this.numDice = numDice; }

    @Override
    public int roll() {
        int total = 0;
        for (int i = 0; i < numDice; i++) {
            total += random.nextInt(6) + 1;
        }
        return total;
    }
}

// ═══════════════════════════════════════════════════════════════
// PLAYER
// ═══════════════════════════════════════════════════════════════
class Player {
    private final String name;
    private int position = 0;

    public Player(String name) { this.name = name; }

    public String getName() { return name; }
    public int getPosition() { return position; }
    public void setPosition(int pos) { this.position = pos; }
}

// ═══════════════════════════════════════════════════════════════
// GAME — Orchestrates everything
// ═══════════════════════════════════════════════════════════════
class Game {
    private final Board board;
    private final Dice dice;
    private final Queue<Player> players = new LinkedList<>();

    public Game(Board board, Dice dice) {
        this.board = board;
        this.dice = dice;
    }

    public void addPlayer(String name) {
        players.add(new Player(name));
    }

    public void play() {
        int turn = 0;
        while (true) {
            Player current = players.poll();
            turn++;

            int diceValue = dice.roll();
            int oldPos = current.getPosition();
            int newPos = oldPos + diceValue;

            System.out.printf("Turn %d: %s rolls %d (%d → %d)%n",
                    turn, current.getName(), diceValue, oldPos, newPos);

            if (newPos > board.getSize()) {
                System.out.println("    Exceeds board! Stay at " + oldPos);
                players.add(current);
                continue;
            }

            newPos = board.getFinalPosition(newPos);
            current.setPosition(newPos);

            if (newPos == board.getSize()) {
                System.out.println("\n  🏆 " + current.getName()
                        + " WINS in " + turn + " turns!\n");
                printFinalStandings(current);
                return;
            }

            players.add(current);
        }
    }

    private void printFinalStandings(Player winner) {
        System.out.println("Final Positions:");
        System.out.println("  1st: " + winner.getName() + " (position " + winner.getPosition() + ")");
        int rank = 2;
        for (Player p : players) {
            System.out.println("  " + rank + ": " + p.getName() + " (position " + p.getPosition() + ")");
            rank++;
        }
    }
}

/*
 * CLASS DIAGRAM:
 * ─────────────────────────────────────────────────────────────
 *   Game (orchestrator)
 *     ├── Board
 *     │     ├── snakes: Map<head, tail>
 *     │     └── ladders: Map<bottom, top>
 *     ├── Dice (Strategy)
 *     │     └── StandardDice
 *     └── Player (Queue for turn order)
 *
 * EXTENSIBILITY:
 *   - CrookedDice: always rolls 6
 *   - PowerUps on certain cells
 *   - Multiple board shapes
 *   - Network multiplayer via Observer
 *
 * COMPILE & RUN:
 *   javac SnakeAndLadder.java && java SnakeAndLadder
 */
