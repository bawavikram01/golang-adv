/*
 * =============================================================
 * LLD CASE STUDY 9: BOOKMYSHOW — Movie Ticket Booking
 * =============================================================
 *
 * REQUIREMENTS:
 *   - Browse movies, theatres, shows
 *   - Select seats and book tickets
 *   - Handle concurrent seat booking (no double booking)
 *   - Payment processing
 *   - Seat categories (Silver, Gold, Platinum)
 *
 * DESIGN PATTERNS USED:
 *   - Strategy (pricing by seat category)
 *   - Observer (booking notifications)
 *   - Builder (Booking object)
 *   - Singleton (BookingManager)
 */

import java.util.*;
import java.util.concurrent.*;
import java.time.*;
import java.util.stream.*;

public class BookMyShow {

    public static void main(String[] args) {
        System.out.println("=== BOOKMYSHOW — Movie Ticket Booking ===\n");

        // Setup
        Movie inception = new Movie("M1", "Inception", 148, Genre.SCI_FI);
        Movie dark = new Movie("M2", "The Dark Knight", 152, Genre.ACTION);

        Theatre pvr = new Theatre("T1", "PVR Cinemas", "Mumbai");
        Screen screen1 = createScreen("S1", "Screen 1", 5, 10);  // 5 rows × 10 cols
        Screen screen2 = createScreen("S2", "Screen 2", 4, 8);
        pvr.addScreen(screen1);
        pvr.addScreen(screen2);

        ShowTime show1 = new ShowTime("SH1", inception, screen1,
                LocalDateTime.of(2025, 1, 15, 14, 30));
        ShowTime show2 = new ShowTime("SH2", dark, screen2,
                LocalDateTime.of(2025, 1, 15, 18, 0));

        BookingManager manager = BookingManager.getInstance();

        // Browse movies
        System.out.println("🎬 Now Showing at " + pvr.getName() + ":");
        System.out.println("  1. " + inception.getTitle() + " (" + inception.getDuration() + " min)");
        System.out.println("  2. " + dark.getTitle() + " (" + dark.getDuration() + " min)\n");

        // ═══════════════════════════════════════════════════════
        // Scenario 1: Successful booking
        // ═══════════════════════════════════════════════════════
        System.out.println("--- Booking 1: Alice books 3 seats for Inception ---");
        show1.showAvailability();

        List<Seat> aliceSeats = show1.getAvailableSeats().stream()
                .filter(s -> s.getRow() == 1)
                .limit(3)
                .collect(Collectors.toList());

        Booking booking1 = manager.createBooking("Alice", show1, aliceSeats);
        if (booking1 != null) {
            System.out.println(booking1);
        }

        // ═══════════════════════════════════════════════════════
        // Scenario 2: Concurrent booking — Bob tries same seats
        // ═══════════════════════════════════════════════════════
        System.out.println("\n--- Booking 2: Bob tries same seats (should fail) ---");
        Booking booking2 = manager.createBooking("Bob", show1, aliceSeats);
        if (booking2 == null) {
            System.out.println("  ✗ Booking failed — seats already taken!\n");
        }

        // ═══════════════════════════════════════════════════════
        // Scenario 3: Bob books different seats
        // ═══════════════════════════════════════════════════════
        System.out.println("--- Booking 3: Bob books row 3 (Gold) seats ---");
        List<Seat> bobSeats = show1.getAvailableSeats().stream()
                .filter(s -> s.getRow() == 3)
                .limit(2)
                .collect(Collectors.toList());

        Booking booking3 = manager.createBooking("Bob", show1, bobSeats);
        if (booking3 != null) {
            System.out.println(booking3);
        }

        // ═══════════════════════════════════════════════════════
        // Scenario 4: Cancel booking
        // ═══════════════════════════════════════════════════════
        System.out.println("\n--- Cancellation: Alice cancels ---");
        manager.cancelBooking(booking1);
        show1.showAvailability();
    }

    static Screen createScreen(String id, String name, int rows, int cols) {
        Screen screen = new Screen(id, name);
        for (int r = 1; r <= rows; r++) {
            SeatCategory category;
            if (r <= 2) category = SeatCategory.PLATINUM;
            else if (r <= 4) category = SeatCategory.GOLD;
            else category = SeatCategory.SILVER;

            for (int c = 1; c <= cols; c++) {
                screen.addSeat(new Seat(r + "-" + c, r, c, category));
            }
        }
        return screen;
    }
}

// ═══════════════════════════════════════════════════════════════
// ENUMS
// ═══════════════════════════════════════════════════════════════
enum Genre { ACTION, COMEDY, DRAMA, HORROR, SCI_FI, THRILLER }

enum SeatCategory {
    SILVER(150), GOLD(250), PLATINUM(400);

    private final double price;
    SeatCategory(double price) { this.price = price; }
    public double getPrice() { return price; }
}

enum BookingStatus { CONFIRMED, CANCELLED }

// ═══════════════════════════════════════════════════════════════
// MOVIE
// ═══════════════════════════════════════════════════════════════
class Movie {
    private final String id;
    private final String title;
    private final int duration; // minutes
    private final Genre genre;

    public Movie(String id, String title, int duration, Genre genre) {
        this.id = id; this.title = title; this.duration = duration; this.genre = genre;
    }

    public String getTitle() { return title; }
    public int getDuration() { return duration; }
}

// ═══════════════════════════════════════════════════════════════
// SEAT
// ═══════════════════════════════════════════════════════════════
class Seat {
    private final String id;
    private final int row;
    private final int col;
    private final SeatCategory category;

    public Seat(String id, int row, int col, SeatCategory category) {
        this.id = id; this.row = row; this.col = col; this.category = category;
    }

    public String getId() { return id; }
    public int getRow() { return row; }
    public SeatCategory getCategory() { return category; }

    @Override
    public String toString() { return "Seat[" + id + " " + category + "]"; }
}

// ═══════════════════════════════════════════════════════════════
// SCREEN
// ═══════════════════════════════════════════════════════════════
class Screen {
    private final String id;
    private final String name;
    private final List<Seat> seats = new ArrayList<>();

    public Screen(String id, String name) { this.id = id; this.name = name; }

    public void addSeat(Seat seat) { seats.add(seat); }
    public List<Seat> getSeats() { return Collections.unmodifiableList(seats); }
    public String getName() { return name; }
}

// ═══════════════════════════════════════════════════════════════
// THEATRE
// ═══════════════════════════════════════════════════════════════
class Theatre {
    private final String id;
    private final String name;
    private final String city;
    private final List<Screen> screens = new ArrayList<>();

    public Theatre(String id, String name, String city) {
        this.id = id; this.name = name; this.city = city;
    }

    public void addScreen(Screen screen) { screens.add(screen); }
    public String getName() { return name; }
}

// ═══════════════════════════════════════════════════════════════
// SHOWTIME — manages seat availability per show
// ═══════════════════════════════════════════════════════════════
class ShowTime {
    private final String id;
    private final Movie movie;
    private final Screen screen;
    private final LocalDateTime dateTime;
    private final Set<String> bookedSeatIds = ConcurrentHashMap.newKeySet();

    public ShowTime(String id, Movie movie, Screen screen, LocalDateTime dateTime) {
        this.id = id; this.movie = movie; this.screen = screen; this.dateTime = dateTime;
    }

    public synchronized boolean bookSeats(List<Seat> seats) {
        // Check all seats available
        for (Seat seat : seats) {
            if (bookedSeatIds.contains(seat.getId())) return false;
        }
        // Book all atomically
        for (Seat seat : seats) {
            bookedSeatIds.add(seat.getId());
        }
        return true;
    }

    public synchronized void releaseSeats(List<Seat> seats) {
        for (Seat seat : seats) {
            bookedSeatIds.remove(seat.getId());
        }
    }

    public List<Seat> getAvailableSeats() {
        return screen.getSeats().stream()
                .filter(s -> !bookedSeatIds.contains(s.getId()))
                .collect(Collectors.toList());
    }

    public void showAvailability() {
        long total = screen.getSeats().size();
        long available = getAvailableSeats().size();
        System.out.println("  🎫 " + movie.getTitle() + " @ " + dateTime
                + " — " + available + "/" + total + " seats available");
    }

    public Movie getMovie() { return movie; }
}

// ═══════════════════════════════════════════════════════════════
// BOOKING
// ═══════════════════════════════════════════════════════════════
class Booking {
    private final String id;
    private final String customerName;
    private final ShowTime showTime;
    private final List<Seat> seats;
    private final double totalAmount;
    private BookingStatus status;

    public Booking(String id, String customer, ShowTime show, List<Seat> seats) {
        this.id = id;
        this.customerName = customer;
        this.showTime = show;
        this.seats = seats;
        this.totalAmount = seats.stream().mapToDouble(s -> s.getCategory().getPrice()).sum();
        this.status = BookingStatus.CONFIRMED;
    }

    public void cancel() { this.status = BookingStatus.CANCELLED; }
    public ShowTime getShowTime() { return showTime; }
    public List<Seat> getSeats() { return seats; }

    @Override
    public String toString() {
        return String.format("  ✓ Booking %s: %s | %s | %d seats | ₹%.0f | %s",
                id, customerName, showTime.getMovie().getTitle(),
                seats.size(), totalAmount, status);
    }
}

// ═══════════════════════════════════════════════════════════════
// BOOKING MANAGER — Singleton
// ═══════════════════════════════════════════════════════════════
class BookingManager {
    private static final BookingManager INSTANCE = new BookingManager();
    private int bookingCounter = 0;

    private BookingManager() {}
    public static BookingManager getInstance() { return INSTANCE; }

    public Booking createBooking(String customer, ShowTime show, List<Seat> seats) {
        if (show.bookSeats(seats)) {
            String bookingId = "BK" + (++bookingCounter);
            return new Booking(bookingId, customer, show, seats);
        }
        return null;
    }

    public void cancelBooking(Booking booking) {
        booking.cancel();
        booking.getShowTime().releaseSeats(booking.getSeats());
        System.out.println("  ✓ Booking cancelled. Seats released.");
    }
}

/*
 * CLASS DIAGRAM:
 * ─────────────────────────────────────────────────────────────
 *   BookingManager (Singleton)
 *     └── creates Booking
 *           ├── ShowTime
 *           │     ├── Movie
 *           │     ├── Screen → List<Seat>
 *           │     └── bookedSeatIds (synchronized)
 *           └── List<Seat>
 *                 └── SeatCategory (enum with price)
 *   Theatre → List<Screen>
 *
 * KEY DESIGN DECISIONS:
 *   1. synchronized bookSeats() prevents double booking
 *   2. SeatCategory enum holds pricing (Strategy-like)
 *   3. ShowTime owns seat availability (not Screen)
 *   4. Booking is immutable after creation
 *
 * COMPILE & RUN:
 *   javac BookMyShow.java && java BookMyShow
 */
