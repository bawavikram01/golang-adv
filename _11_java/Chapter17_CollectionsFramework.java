/*
 * ============================================================
 *  CHAPTER 17: COLLECTIONS FRAMEWORK
 * ============================================================
 *
 *  The Collections Framework provides data structures and algorithms.
 *
 *  HIERARCHY:
 *  ──────────
 *  Iterable
 *  └── Collection
 *      ├── List   (ordered, duplicates allowed)
 *      │   ├── ArrayList    — dynamic array, fast random access
 *      │   ├── LinkedList   — doubly linked list, fast insert/delete
 *      │   └── Vector       — synchronized ArrayList (legacy)
 *      ├── Set    (no duplicates)
 *      │   ├── HashSet      — unordered, O(1) operations
 *      │   ├── LinkedHashSet— insertion order preserved
 *      │   └── TreeSet      — sorted, O(log n)
 *      └── Queue  (FIFO)
 *          ├── LinkedList   — also implements Queue
 *          ├── PriorityQueue— elements ordered by priority
 *          └── Deque        — double-ended queue
 *              └── ArrayDeque — fast stack/queue implementation
 *
 *  Map (NOT part of Collection interface)
 *  ├── HashMap      — unordered, O(1), allows null key
 *  ├── LinkedHashMap— insertion order preserved
 *  ├── TreeMap      — sorted by keys, O(log n)
 *  └── Hashtable    — synchronized (legacy)
 *
 *  CHOOSING THE RIGHT COLLECTION:
 *  ┌─────────────────────┬──────────────────────────────────┐
 *  │ Need                │ Use                              │
 *  ├─────────────────────┼──────────────────────────────────┤
 *  │ Indexed access      │ ArrayList                        │
 *  │ Frequent insert/del │ LinkedList                       │
 *  │ No duplicates       │ HashSet                          │
 *  │ Sorted unique items │ TreeSet                          │
 *  │ Key-value pairs     │ HashMap                          │
 *  │ Sorted key-value    │ TreeMap                          │
 *  │ FIFO queue          │ ArrayDeque / LinkedList           │
 *  │ Stack (LIFO)        │ ArrayDeque (not Stack class)     │
 *  │ Priority ordering   │ PriorityQueue                    │
 *  └─────────────────────┴──────────────────────────────────┘
 *
 * ============================================================
 */

import java.util.*;

public class Chapter17_CollectionsFramework {

    public static void main(String[] args) {

        // =====================================================
        //  1. LIST — ArrayList
        // =====================================================

        System.out.println("=== ARRAYLIST ===\n");

        // ArrayList: resizable array, fast get(index), slow insert in middle
        List<String> fruits = new ArrayList<>();

        // Add elements
        fruits.add("Apple");
        fruits.add("Banana");
        fruits.add("Cherry");
        fruits.add("Apple"); // duplicates allowed!
        fruits.add(1, "Avocado"); // insert at index 1

        System.out.println("Fruits: " + fruits);
        System.out.println("Size: " + fruits.size());
        System.out.println("Get(0): " + fruits.get(0));
        System.out.println("Contains 'Banana': " + fruits.contains("Banana"));
        System.out.println("IndexOf 'Apple': " + fruits.indexOf("Apple"));
        System.out.println("LastIndexOf 'Apple': " + fruits.lastIndexOf("Apple"));

        // Modify
        fruits.set(2, "Blueberry"); // replace at index 2
        System.out.println("After set(2): " + fruits);

        // Remove
        fruits.remove("Apple");      // removes first occurrence
        fruits.remove(0);            // removes by index
        System.out.println("After removes: " + fruits);

        // Iterate
        System.out.println("\n--- Iterating ---");
        // for-each
        for (String f : fruits) System.out.print(f + " ");
        System.out.println();

        // Iterator
        Iterator<String> it = fruits.iterator();
        while (it.hasNext()) System.out.print(it.next() + " ");
        System.out.println();

        // Sublist
        List<String> sub = fruits.subList(0, 2);
        System.out.println("Sublist(0,2): " + sub);

        // Sort
        Collections.sort(fruits);
        System.out.println("Sorted: " + fruits);

        // Create from existing
        List<String> copy = new ArrayList<>(fruits);
        List<String> fixed = List.of("X", "Y", "Z"); // immutable list (Java 9+)
        List<String> fromArray = Arrays.asList("A", "B", "C"); // fixed-size list

        // =====================================================
        //  2. LIST — LinkedList
        // =====================================================

        System.out.println("\n=== LINKEDLIST ===\n");

        LinkedList<Integer> linked = new LinkedList<>();
        linked.add(10);
        linked.add(20);
        linked.add(30);
        linked.addFirst(5);   // add to beginning
        linked.addLast(40);   // add to end

        System.out.println("LinkedList: " + linked);
        System.out.println("First: " + linked.getFirst());
        System.out.println("Last: " + linked.getLast());
        System.out.println("Peek: " + linked.peek());     // like getFirst but doesn't throw
        System.out.println("Poll: " + linked.poll());      // removes and returns first
        System.out.println("After poll: " + linked);

        // =====================================================
        //  3. SET — HashSet
        // =====================================================

        System.out.println("\n=== HASHSET ===\n");

        Set<String> uniqueFruits = new HashSet<>();
        uniqueFruits.add("Apple");
        uniqueFruits.add("Banana");
        uniqueFruits.add("Cherry");
        uniqueFruits.add("Apple");    // duplicate — NOT added!
        uniqueFruits.add("Banana");   // duplicate — NOT added!

        System.out.println("HashSet: " + uniqueFruits);
        System.out.println("Size: " + uniqueFruits.size());      // 3
        System.out.println("Contains 'Apple': " + uniqueFruits.contains("Apple"));

        uniqueFruits.remove("Banana");
        System.out.println("After remove: " + uniqueFruits);

        // NOTE: HashSet has NO guaranteed order!

        // =====================================================
        //  4. SET — LinkedHashSet (maintains insertion order)
        // =====================================================

        System.out.println("\n=== LINKEDHASHSET ===\n");

        Set<String> orderedSet = new LinkedHashSet<>();
        orderedSet.add("Cherry");
        orderedSet.add("Apple");
        orderedSet.add("Banana");
        orderedSet.add("Apple");  // duplicate ignored
        System.out.println("LinkedHashSet: " + orderedSet); // maintains insertion order

        // =====================================================
        //  5. SET — TreeSet (sorted)
        // =====================================================

        System.out.println("\n=== TREESET ===\n");

        Set<Integer> sortedSet = new TreeSet<>();
        sortedSet.add(50);
        sortedSet.add(10);
        sortedSet.add(30);
        sortedSet.add(20);
        sortedSet.add(40);
        sortedSet.add(10); // duplicate ignored
        System.out.println("TreeSet (sorted): " + sortedSet);

        TreeSet<Integer> treeSet = (TreeSet<Integer>) sortedSet;
        System.out.println("First: " + treeSet.first());
        System.out.println("Last: " + treeSet.last());
        System.out.println("HeadSet(<30): " + treeSet.headSet(30));   // elements < 30
        System.out.println("TailSet(>=30): " + treeSet.tailSet(30));   // elements >= 30

        // =====================================================
        //  6. MAP — HashMap
        // =====================================================

        System.out.println("\n=== HASHMAP ===\n");

        Map<String, Integer> ages = new HashMap<>();
        ages.put("Alice", 25);
        ages.put("Bob", 30);
        ages.put("Charlie", 35);
        ages.put("Alice", 26);  // overwrites previous value!

        System.out.println("Map: " + ages);
        System.out.println("Size: " + ages.size());
        System.out.println("Get Alice: " + ages.get("Alice")); // 26
        System.out.println("Contains key 'Bob': " + ages.containsKey("Bob"));
        System.out.println("Contains value 30: " + ages.containsValue(30));

        // getOrDefault
        System.out.println("Get Dave (default): " + ages.getOrDefault("Dave", 0));

        // putIfAbsent — only adds if key doesn't exist
        ages.putIfAbsent("Alice", 99); // won't change, Alice exists
        ages.putIfAbsent("Dave", 28);  // adds Dave
        System.out.println("After putIfAbsent: " + ages);

        // Remove
        ages.remove("Dave");
        System.out.println("After remove: " + ages);

        // Iterate a Map
        System.out.println("\n--- Iterating Map ---");
        // Method 1: entrySet
        for (Map.Entry<String, Integer> entry : ages.entrySet()) {
            System.out.println("  " + entry.getKey() + " → " + entry.getValue());
        }
        // Method 2: keySet
        for (String key : ages.keySet()) {
            System.out.println("  Key: " + key);
        }
        // Method 3: values
        for (int value : ages.values()) {
            System.out.println("  Value: " + value);
        }
        // Method 4: forEach (Java 8+)
        ages.forEach((key, value) -> System.out.println("  " + key + "=" + value));

        // =====================================================
        //  7. MAP — TreeMap (sorted by keys)
        // =====================================================

        System.out.println("\n=== TREEMAP ===\n");

        TreeMap<String, Integer> sortedMap = new TreeMap<>(ages);
        System.out.println("TreeMap (sorted): " + sortedMap);
        System.out.println("FirstKey: " + sortedMap.firstKey());
        System.out.println("LastKey: " + sortedMap.lastKey());

        // =====================================================
        //  8. QUEUE & DEQUE
        // =====================================================

        System.out.println("\n=== QUEUE ===\n");

        // Queue: FIFO (First In, First Out)
        Queue<String> queue = new LinkedList<>();
        queue.offer("First");    // add to tail
        queue.offer("Second");
        queue.offer("Third");
        System.out.println("Queue: " + queue);
        System.out.println("Peek: " + queue.peek());   // view head without removing
        System.out.println("Poll: " + queue.poll());    // remove and return head
        System.out.println("After poll: " + queue);

        // Deque: Double-ended queue (can be used as Stack or Queue)
        System.out.println("\n=== DEQUE (Stack) ===\n");

        Deque<String> stack = new ArrayDeque<>();
        stack.push("First");    // push to top
        stack.push("Second");
        stack.push("Third");
        System.out.println("Stack: " + stack);
        System.out.println("Peek: " + stack.peek());
        System.out.println("Pop: " + stack.pop());
        System.out.println("After pop: " + stack);

        // PriorityQueue
        System.out.println("\n=== PRIORITY QUEUE ===\n");

        PriorityQueue<Integer> pq = new PriorityQueue<>(); // min-heap
        pq.offer(30);
        pq.offer(10);
        pq.offer(50);
        pq.offer(20);
        System.out.print("PriorityQueue (dequeue order): ");
        while (!pq.isEmpty()) {
            System.out.print(pq.poll() + " "); // comes out sorted!
        }
        System.out.println();

        // =====================================================
        //  9. COLLECTIONS UTILITY CLASS
        // =====================================================

        System.out.println("\n=== COLLECTIONS UTILITY ===\n");

        List<Integer> nums = new ArrayList<>(Arrays.asList(5, 2, 8, 1, 9, 3));

        Collections.sort(nums);
        System.out.println("Sorted: " + nums);

        Collections.reverse(nums);
        System.out.println("Reversed: " + nums);

        Collections.shuffle(nums);
        System.out.println("Shuffled: " + nums);

        System.out.println("Min: " + Collections.min(nums));
        System.out.println("Max: " + Collections.max(nums));
        System.out.println("Frequency of 5: " + Collections.frequency(nums, 5));

        Collections.fill(nums, 0);
        System.out.println("Filled with 0: " + nums);

        // Immutable collections
        List<String> immutable = Collections.unmodifiableList(fruits);
        // immutable.add("test"); // UnsupportedOperationException!
        System.out.println("Unmodifiable list: " + immutable);

        // Synchronized collections (for thread safety)
        List<String> syncList = Collections.synchronizedList(new ArrayList<>());

        // =====================================================
        //  10. PRACTICAL: Word Frequency Counter
        // =====================================================

        System.out.println("\n=== PRACTICAL: WORD FREQUENCY ===\n");

        String text = "the quick brown fox jumps over the lazy dog the fox";
        Map<String, Integer> freq = new HashMap<>();
        for (String word : text.split(" ")) {
            freq.put(word, freq.getOrDefault(word, 0) + 1);
        }
        System.out.println("Word frequencies: " + freq);

        // Sort by frequency
        List<Map.Entry<String, Integer>> entries = new ArrayList<>(freq.entrySet());
        entries.sort((e1, e2) -> e2.getValue() - e1.getValue());
        System.out.println("Sorted by frequency:");
        for (Map.Entry<String, Integer> entry : entries) {
            System.out.println("  " + entry.getKey() + ": " + entry.getValue());
        }
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Remove all duplicates from an ArrayList (preserve order).
 *
 *  2. Find the intersection and union of two sets.
 *
 *  3. Count character frequency in a string using a HashMap.
 *
 *  4. Implement an LRU cache using LinkedHashMap.
 *
 *  5. Group a list of strings by their length using a Map.
 *
 *  6. Implement a stack using ArrayDeque and solve:
 *     - Balanced parentheses check: "({[]})" → true
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 18 — Collections Deep Dive
 * ============================================================
 */
