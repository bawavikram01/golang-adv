/*
 * =============================================================
 * BEHAVIORAL PATTERN 7: ITERATOR
 * =============================================================
 *
 * INTENT: Provide a way to access elements of a collection
 *         sequentially WITHOUT exposing its internal structure.
 *
 * ANALOGY: A Spotify playlist — you press Next/Previous without
 *          knowing if songs are stored in an array, linked list,
 *          or database. The iterator handles traversal.
 *
 * USE WHEN:
 *   - You want to traverse a collection without exposing internals
 *   - You need multiple traversal strategies (forward, reverse, filtered)
 *   - Different collections need a uniform traversal interface
 *
 * REAL EXAMPLES: java.util.Iterator, for-each loop,
 *                java.util.stream, database cursors
 */

import java.util.*;

public class IteratorPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Custom Collection with Iterator
        // ═══════════════════════════════════════════════════════
        System.out.println("=== CUSTOM COLLECTION ITERATOR ===");

        Playlist playlist = new Playlist("My Favorites");
        playlist.addSong(new Song("Bohemian Rhapsody", "Queen", 354));
        playlist.addSong(new Song("Hotel California", "Eagles", 391));
        playlist.addSong(new Song("Stairway to Heaven", "Led Zeppelin", 482));
        playlist.addSong(new Song("Imagine", "John Lennon", 183));
        playlist.addSong(new Song("Hey Jude", "The Beatles", 431));

        // Forward iteration
        System.out.println("Forward:");
        SongIterator it = playlist.createIterator();
        while (it.hasNext()) {
            System.out.println("  ▶ " + it.next());
        }

        // Reverse iteration — different iterator, same collection!
        System.out.println("\nReverse:");
        SongIterator reverseIt = playlist.createReverseIterator();
        while (reverseIt.hasNext()) {
            System.out.println("  ◀ " + reverseIt.next());
        }

        // Filtered iteration — only songs over 5 minutes
        System.out.println("\nSongs > 5 minutes:");
        SongIterator longSongs = playlist.createFilteredIterator(s -> s.getDuration() > 300);
        while (longSongs.hasNext()) {
            System.out.println("  🎵 " + longSongs.next());
        }

        // ═══════════════════════════════════════════════════════
        // Java's built-in Iterable (for-each loop support)
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== JAVA ITERABLE (for-each) ===");

        NumberRange range = new NumberRange(1, 10);
        System.out.print("  Range 1-10: ");
        for (int n : range) {  // for-each works because we implement Iterable!
            System.out.print(n + " ");
        }
        System.out.println();

        // ═══════════════════════════════════════════════════════
        // Tree iterator — traverse tree structure  
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== TREE ITERATOR (BFS vs DFS) ===");

        TreeNode root = new TreeNode("CEO");
        TreeNode vpEng = new TreeNode("VP Eng");
        TreeNode vpSales = new TreeNode("VP Sales");
        root.addChild(vpEng);
        root.addChild(vpSales);
        vpEng.addChild(new TreeNode("Dev Lead"));
        vpEng.addChild(new TreeNode("QA Lead"));
        vpSales.addChild(new TreeNode("Sales Rep"));

        System.out.print("  BFS: ");
        TreeIterator bfs = new BfsTreeIterator(root);
        while (bfs.hasNext()) {
            System.out.print(bfs.next().getValue() + " → ");
        }
        System.out.println("end");

        System.out.print("  DFS: ");
        TreeIterator dfs = new DfsTreeIterator(root);
        while (dfs.hasNext()) {
            System.out.print(dfs.next().getValue() + " → ");
        }
        System.out.println("end");
    }
}

// ═══════════════════════════════════════════════════════════════
// CUSTOM ITERATOR
// ═══════════════════════════════════════════════════════════════
class Song {
    private String title;
    private String artist;
    private int duration;  // seconds

    public Song(String title, String artist, int duration) {
        this.title = title;
        this.artist = artist;
        this.duration = duration;
    }

    public int getDuration() { return duration; }

    @Override
    public String toString() {
        return title + " — " + artist + " (" + duration / 60 + ":" + String.format("%02d", duration % 60) + ")";
    }
}

// Iterator interface
interface SongIterator {
    boolean hasNext();
    Song next();
}

// Aggregate (collection)
class Playlist {
    private String name;
    private List<Song> songs = new ArrayList<>();

    public Playlist(String name) { this.name = name; }

    public void addSong(Song song) { songs.add(song); }

    // Factory methods for different iterators
    public SongIterator createIterator() {
        return new ForwardIterator(songs);
    }

    public SongIterator createReverseIterator() {
        return new ReverseIterator(songs);
    }

    public SongIterator createFilteredIterator(java.util.function.Predicate<Song> filter) {
        return new FilteredIterator(songs, filter);
    }
}

// Forward iterator
class ForwardIterator implements SongIterator {
    private List<Song> songs;
    private int position = 0;

    public ForwardIterator(List<Song> songs) { this.songs = songs; }

    @Override public boolean hasNext() { return position < songs.size(); }
    @Override public Song next()       { return songs.get(position++); }
}

// Reverse iterator — same collection, different traversal!
class ReverseIterator implements SongIterator {
    private List<Song> songs;
    private int position;

    public ReverseIterator(List<Song> songs) {
        this.songs = songs;
        this.position = songs.size() - 1;
    }

    @Override public boolean hasNext() { return position >= 0; }
    @Override public Song next()       { return songs.get(position--); }
}

// Filtered iterator
class FilteredIterator implements SongIterator {
    private List<Song> filtered;
    private int position = 0;

    public FilteredIterator(List<Song> songs, java.util.function.Predicate<Song> pred) {
        this.filtered = new ArrayList<>();
        for (Song s : songs) {
            if (pred.test(s)) filtered.add(s);
        }
    }

    @Override public boolean hasNext() { return position < filtered.size(); }
    @Override public Song next()       { return filtered.get(position++); }
}

// ═══════════════════════════════════════════════════════════════
// JAVA ITERABLE — enables for-each loop
// ═══════════════════════════════════════════════════════════════
class NumberRange implements Iterable<Integer> {
    private int start, end;

    public NumberRange(int start, int end) {
        this.start = start;
        this.end = end;
    }

    @Override
    public Iterator<Integer> iterator() {
        return new Iterator<>() {
            private int current = start;

            @Override public boolean hasNext() { return current <= end; }
            @Override public Integer next()    { return current++; }
        };
    }
}

// ═══════════════════════════════════════════════════════════════
// TREE ITERATOR — BFS and DFS
// ═══════════════════════════════════════════════════════════════
class TreeNode {
    private String value;
    private List<TreeNode> children = new ArrayList<>();

    public TreeNode(String value) { this.value = value; }

    public void addChild(TreeNode child) { children.add(child); }
    public String getValue() { return value; }
    public List<TreeNode> getChildren() { return children; }
}

interface TreeIterator {
    boolean hasNext();
    TreeNode next();
}

class BfsTreeIterator implements TreeIterator {
    private Queue<TreeNode> queue = new LinkedList<>();

    public BfsTreeIterator(TreeNode root) {
        if (root != null) queue.add(root);
    }

    @Override public boolean hasNext() { return !queue.isEmpty(); }

    @Override
    public TreeNode next() {
        TreeNode node = queue.poll();
        queue.addAll(node.getChildren());
        return node;
    }
}

class DfsTreeIterator implements TreeIterator {
    private Deque<TreeNode> stack = new ArrayDeque<>();

    public DfsTreeIterator(TreeNode root) {
        if (root != null) stack.push(root);
    }

    @Override public boolean hasNext() { return !stack.isEmpty(); }

    @Override
    public TreeNode next() {
        TreeNode node = stack.pop();
        List<TreeNode> children = node.getChildren();
        for (int i = children.size() - 1; i >= 0; i--) {
            stack.push(children.get(i));
        }
        return node;
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Iterator decouples traversal logic from the collection.
 * ✦ Same collection can have MULTIPLE iterators (forward, reverse, filtered).
 * ✦ In Java: implement Iterable<T> + Iterator<T> for for-each support.
 * ✦ Tree traversal (BFS/DFS) is a classic iterator application.
 *
 * ✦ Java streams are essentially lazy iterators with transformations.
 *
 * COMPILE & RUN:
 *   javac IteratorPattern.java && java IteratorPattern
 */
