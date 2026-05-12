/*
 * ============================================================
 *  CHAPTER 26: DATE & TIME API (java.time — Java 8+)
 * ============================================================
 *  The modern Date/Time API — immutable, thread-safe, easy to use.
 *  Replaces legacy Date, Calendar, SimpleDateFormat.
 *
 *  Key classes:
 *  LocalDate     — date only (2024-12-25)
 *  LocalTime     — time only (14:30:00)
 *  LocalDateTime — date + time
 *  ZonedDateTime — date + time + timezone
 *  Instant       — timestamp (epoch seconds)
 *  Duration      — time-based amount (hours, minutes, seconds)
 *  Period        — date-based amount (years, months, days)
 *  DateTimeFormatter — formatting and parsing
 * ============================================================
 */

import java.time.*;
import java.time.format.DateTimeFormatter;
import java.time.temporal.ChronoUnit;

public class Chapter26_DateTime {

    public static void main(String[] args) {

        // --- 1. LocalDate ---
        System.out.println("=== LOCAL DATE ===\n");
        LocalDate today = LocalDate.now();
        LocalDate birthday = LocalDate.of(1995, 6, 15);
        LocalDate parsed = LocalDate.parse("2024-12-25");

        System.out.println("Today: " + today);
        System.out.println("Birthday: " + birthday);
        System.out.println("Year: " + today.getYear() + ", Month: " + today.getMonthValue()
                + ", Day: " + today.getDayOfMonth());
        System.out.println("Day of week: " + today.getDayOfWeek());
        System.out.println("Day of year: " + today.getDayOfYear());
        System.out.println("Leap year: " + today.isLeapYear());

        // Manipulation (returns NEW object — immutable!)
        System.out.println("Plus 7 days: " + today.plusDays(7));
        System.out.println("Minus 1 month: " + today.minusMonths(1));
        System.out.println("Plus 1 year: " + today.plusYears(1));

        // Comparison
        System.out.println("Before birthday? " + today.isBefore(birthday));
        System.out.println("After birthday? " + today.isAfter(birthday));

        // --- 2. LocalTime ---
        System.out.println("\n=== LOCAL TIME ===\n");
        LocalTime now = LocalTime.now();
        LocalTime lunch = LocalTime.of(12, 30, 0);
        LocalTime parsedTime = LocalTime.parse("14:30:00");

        System.out.println("Now: " + now);
        System.out.println("Lunch: " + lunch);
        System.out.println("Hour: " + now.getHour() + ", Minute: " + now.getMinute());
        System.out.println("Plus 2 hours: " + now.plusHours(2));

        // --- 3. LocalDateTime ---
        System.out.println("\n=== LOCAL DATETIME ===\n");
        LocalDateTime dateTime = LocalDateTime.now();
        LocalDateTime specific = LocalDateTime.of(2024, 12, 25, 10, 30, 0);

        System.out.println("Now: " + dateTime);
        System.out.println("Christmas: " + specific);
        System.out.println("Date part: " + dateTime.toLocalDate());
        System.out.println("Time part: " + dateTime.toLocalTime());

        // --- 4. ZonedDateTime ---
        System.out.println("\n=== ZONED DATETIME ===\n");
        ZonedDateTime zoned = ZonedDateTime.now();
        ZonedDateTime tokyo = ZonedDateTime.now(ZoneId.of("Asia/Tokyo"));
        ZonedDateTime ny = ZonedDateTime.now(ZoneId.of("America/New_York"));

        System.out.println("Local: " + zoned);
        System.out.println("Tokyo: " + tokyo);
        System.out.println("New York: " + ny);

        // --- 5. Instant (Epoch timestamp) ---
        System.out.println("\n=== INSTANT ===\n");
        Instant timestamp = Instant.now();
        System.out.println("Instant: " + timestamp);
        System.out.println("Epoch seconds: " + timestamp.getEpochSecond());
        System.out.println("Epoch millis: " + timestamp.toEpochMilli());

        // --- 6. Duration & Period ---
        System.out.println("\n=== DURATION & PERIOD ===\n");

        // Duration: time-based (hours, minutes, seconds)
        Duration duration = Duration.between(lunch, now);
        System.out.println("Since lunch: " + duration);
        System.out.println("In minutes: " + duration.toMinutes());

        Duration twoHours = Duration.ofHours(2);
        System.out.println("2 hours: " + twoHours);

        // Period: date-based (years, months, days)
        Period age = Period.between(birthday, today);
        System.out.println("Age: " + age.getYears() + " years, " + age.getMonths()
                + " months, " + age.getDays() + " days");

        long daysBetween = ChronoUnit.DAYS.between(birthday, today);
        System.out.println("Total days alive: " + daysBetween);

        // --- 7. DateTimeFormatter ---
        System.out.println("\n=== FORMATTING ===\n");

        DateTimeFormatter fmt1 = DateTimeFormatter.ofPattern("dd/MM/yyyy");
        DateTimeFormatter fmt2 = DateTimeFormatter.ofPattern("EEEE, MMMM dd, yyyy");
        DateTimeFormatter fmt3 = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

        System.out.println("dd/MM/yyyy: " + today.format(fmt1));
        System.out.println("Full: " + today.format(fmt2));
        System.out.println("DateTime: " + dateTime.format(fmt3));

        // Parsing with formatter
        LocalDate parsedDate = LocalDate.parse("25/12/2024", fmt1);
        System.out.println("Parsed: " + parsedDate);

        System.out.println("\nALL date/time objects are IMMUTABLE and THREAD-SAFE!");
    }
}

/*
 * EXERCISES:
 * 1. Calculate how many days until your next birthday.
 * 2. Find what day of the week January 1, 2000 was.
 * 3. List all Fridays in the current month.
 * 4. Calculate working days between two dates (exclude weekends).
 *
 * NEXT: Chapter 27 — Regular Expressions
 */
