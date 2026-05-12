/*
 * =============================================================
 * LLD CASE STUDY 3: LIBRARY MANAGEMENT SYSTEM
 * =============================================================
 *
 * REQUIREMENTS:
 *   - Add/remove books
 *   - Search by title, author, ISBN
 *   - Issue/return books
 *   - Track members and their borrowed books
 *   - Fine calculation for overdue books
 *   - Book reservation
 *
 * DESIGN PATTERNS USED:
 *   - Strategy (search, fine calculation)
 *   - Observer (notify when reserved book is returned)
 *   - Singleton (library catalog)
 *   - Builder (book construction)
 */

import java.time.LocalDate;
import java.util.*;
import java.util.stream.Collectors;

public class LibraryManagementSystem {

    public static void main(String[] args) {

        Library library = Library.getInstance();

        // Add books
        System.out.println("=== ADDING BOOKS ===");
        Book b1 = new Book.Builder("978-0-13-468599-1", "Clean Code")
                .author("Robert C. Martin")
                .genre("Software Engineering")
                .copies(3)
                .build();

        Book b2 = new Book.Builder("978-0-201-63361-0", "Design Patterns")
                .author("Gang of Four")
                .genre("Software Engineering")
                .copies(2)
                .build();

        Book b3 = new Book.Builder("978-0-596-00712-6", "Head First Design Patterns")
                .author("Eric Freeman")
                .genre("Software Engineering")
                .copies(1)
                .build();

        library.addBook(b1);
        library.addBook(b2);
        library.addBook(b3);

        // Register members
        System.out.println("\n=== REGISTERING MEMBERS ===");
        Member alice = new Member("M001", "Alice");
        Member bob = new Member("M002", "Bob");
        library.registerMember(alice);
        library.registerMember(bob);

        // Search books
        System.out.println("\n=== SEARCHING ===");
        List<Book> results = library.searchByTitle("Clean");
        System.out.println("  Found by title 'Clean': " + results);

        results = library.searchByAuthor("Gang");
        System.out.println("  Found by author 'Gang': " + results);

        // Issue books
        System.out.println("\n=== ISSUING BOOKS ===");
        library.issueBook("978-0-13-468599-1", alice);
        library.issueBook("978-0-13-468599-1", bob);
        library.issueBook("978-0-596-00712-6", alice);

        // Display status
        library.displayBookStatus();

        // Try to issue last copy (already all taken if copies=1, alice has it)
        System.out.println("\n=== TRYING TO ISSUE UNAVAILABLE BOOK ===");
        library.issueBook("978-0-596-00712-6", bob);  // should fail or reserve

        // Return books
        System.out.println("\n=== RETURNING BOOKS ===");
        library.returnBook("978-0-13-468599-1", alice);
        library.returnBook("978-0-596-00712-6", alice);

        library.displayBookStatus();

        // Display member info
        System.out.println("\n=== MEMBER INFO ===");
        alice.displayInfo();
        bob.displayInfo();
    }
}

// ═══════════════════════════════════════════════════════════════
// BOOK (with Builder)
// ═══════════════════════════════════════════════════════════════
class Book {
    private final String isbn;
    private final String title;
    private final String author;
    private final String genre;
    private int totalCopies;
    private int availableCopies;
    private final List<BookReservationObserver> observers = new ArrayList<>();

    private Book(Builder builder) {
        this.isbn = builder.isbn;
        this.title = builder.title;
        this.author = builder.author;
        this.genre = builder.genre;
        this.totalCopies = builder.copies;
        this.availableCopies = builder.copies;
    }

    public boolean isAvailable() { return availableCopies > 0; }

    public void checkOut() {
        if (availableCopies > 0) availableCopies--;
    }

    public void checkIn() {
        if (availableCopies < totalCopies) {
            availableCopies++;
            notifyObservers();
        }
    }

    public void addObserver(BookReservationObserver obs) { observers.add(obs); }
    public void removeObserver(BookReservationObserver obs) { observers.remove(obs); }

    private void notifyObservers() {
        for (BookReservationObserver obs : observers) {
            obs.onBookAvailable(this);
        }
        observers.clear();
    }

    public String getIsbn() { return isbn; }
    public String getTitle() { return title; }
    public String getAuthor() { return author; }
    public int getAvailableCopies() { return availableCopies; }
    public int getTotalCopies() { return totalCopies; }

    @Override
    public String toString() { return "\"" + title + "\" by " + author; }

    public static class Builder {
        private final String isbn;
        private final String title;
        private String author = "Unknown";
        private String genre = "General";
        private int copies = 1;

        public Builder(String isbn, String title) {
            this.isbn = isbn;
            this.title = title;
        }

        public Builder author(String a)   { this.author = a; return this; }
        public Builder genre(String g)    { this.genre = g; return this; }
        public Builder copies(int c)      { this.copies = c; return this; }
        public Book build() { return new Book(this); }
    }
}

// ═══════════════════════════════════════════════════════════════
// OBSERVER for book reservations
// ═══════════════════════════════════════════════════════════════
interface BookReservationObserver {
    void onBookAvailable(Book book);
}

// ═══════════════════════════════════════════════════════════════
// MEMBER
// ═══════════════════════════════════════════════════════════════
class Member implements BookReservationObserver {
    private final String memberId;
    private final String name;
    private final List<LoanRecord> activeLoans = new ArrayList<>();
    private final List<LoanRecord> loanHistory = new ArrayList<>();
    private static final int MAX_BOOKS = 5;

    public Member(String memberId, String name) {
        this.memberId = memberId;
        this.name = name;
    }

    public boolean canBorrow() { return activeLoans.size() < MAX_BOOKS; }

    public void addLoan(LoanRecord record) { activeLoans.add(record); }

    public LoanRecord returnBook(String isbn) {
        LoanRecord record = activeLoans.stream()
                .filter(r -> r.getBook().getIsbn().equals(isbn))
                .findFirst()
                .orElse(null);
        if (record != null) {
            record.setReturnDate(LocalDate.now());
            activeLoans.remove(record);
            loanHistory.add(record);
        }
        return record;
    }

    @Override
    public void onBookAvailable(Book book) {
        System.out.println("  🔔 " + name + ": \"" + book.getTitle() + "\" is now available!");
    }

    public void displayInfo() {
        System.out.println("  Member: " + name + " (" + memberId + ")");
        System.out.println("    Active loans: " + activeLoans.size());
        for (LoanRecord r : activeLoans) {
            System.out.println("      - " + r.getBook().getTitle() + " (due: " + r.getDueDate() + ")");
        }
        System.out.println("    History: " + loanHistory.size() + " books returned");
    }

    public String getName() { return name; }
    public String getMemberId() { return memberId; }
}

// ═══════════════════════════════════════════════════════════════
// LOAN RECORD
// ═══════════════════════════════════════════════════════════════
class LoanRecord {
    private final Book book;
    private final Member member;
    private final LocalDate issueDate;
    private final LocalDate dueDate;
    private LocalDate returnDate;

    public LoanRecord(Book book, Member member) {
        this.book = book;
        this.member = member;
        this.issueDate = LocalDate.now();
        this.dueDate = issueDate.plusDays(14);  // 2-week loan period
    }

    public boolean isOverdue() {
        LocalDate checkDate = returnDate != null ? returnDate : LocalDate.now();
        return checkDate.isAfter(dueDate);
    }

    public void setReturnDate(LocalDate date) { this.returnDate = date; }
    public Book getBook() { return book; }
    public LocalDate getDueDate() { return dueDate; }
}

// ═══════════════════════════════════════════════════════════════
// LIBRARY (Singleton)
// ═══════════════════════════════════════════════════════════════
class Library {
    private static Library instance;
    private final Map<String, Book> catalog = new HashMap<>();  // isbn → book
    private final Map<String, Member> members = new HashMap<>();  // id → member

    private Library() {}

    public static Library getInstance() {
        if (instance == null) instance = new Library();
        return instance;
    }

    public void addBook(Book book) {
        catalog.put(book.getIsbn(), book);
        System.out.println("  ✓ Added: " + book + " (" + book.getTotalCopies() + " copies)");
    }

    public void registerMember(Member member) {
        members.put(member.getMemberId(), member);
        System.out.println("  ✓ Registered: " + member.getName());
    }

    public List<Book> searchByTitle(String keyword) {
        return catalog.values().stream()
                .filter(b -> b.getTitle().toLowerCase().contains(keyword.toLowerCase()))
                .collect(Collectors.toList());
    }

    public List<Book> searchByAuthor(String keyword) {
        return catalog.values().stream()
                .filter(b -> b.getAuthor().toLowerCase().contains(keyword.toLowerCase()))
                .collect(Collectors.toList());
    }

    public boolean issueBook(String isbn, Member member) {
        Book book = catalog.get(isbn);
        if (book == null) {
            System.out.println("  ✗ Book not found: " + isbn);
            return false;
        }
        if (!member.canBorrow()) {
            System.out.println("  ✗ " + member.getName() + " has reached borrowing limit!");
            return false;
        }
        if (!book.isAvailable()) {
            System.out.println("  ✗ " + book.getTitle() + " unavailable. Reserving for " + member.getName());
            book.addObserver(member);  // Observer: notify when returned
            return false;
        }

        book.checkOut();
        LoanRecord record = new LoanRecord(book, member);
        member.addLoan(record);
        System.out.println("  ✓ Issued: \"" + book.getTitle() + "\" to " + member.getName()
                + " (due: " + record.getDueDate() + ")");
        return true;
    }

    public void returnBook(String isbn, Member member) {
        LoanRecord record = member.returnBook(isbn);
        if (record != null) {
            record.getBook().checkIn();  // triggers observer notification
            System.out.println("  ✓ Returned: \"" + record.getBook().getTitle() + "\" by " + member.getName());
        }
    }

    public void displayBookStatus() {
        System.out.println("  ┌─── CATALOG STATUS ─────────────────────────┐");
        for (Book book : catalog.values()) {
            System.out.printf("  │ %-35s %d/%d available%n",
                    book.getTitle(), book.getAvailableCopies(), book.getTotalCopies());
        }
        System.out.println("  └─────────────────────────────────────────────┘");
    }
}

/*
 * CLASS DIAGRAM:
 * ─────────────────────────────────────────────────────────────
 *   Library (Singleton)
 *     ├── Map<ISBN, Book>
 *     │     ├── title, author, availableCopies
 *     │     └── List<BookReservationObserver>
 *     └── Map<MemberId, Member>
 *           ├── List<LoanRecord> activeLoans
 *           └── implements BookReservationObserver
 *
 *   LoanRecord → Book + Member + dates
 *
 * PATTERNS USED:
 *   ✦ Singleton — Library
 *   ✦ Builder — Book construction
 *   ✦ Observer — reservation notifications
 *   ✦ Strategy-ready — search and fine calculation
 *
 * COMPILE & RUN:
 *   javac LibraryManagementSystem.java && java LibraryManagementSystem
 */
