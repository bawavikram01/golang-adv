/*
 * ============================================================
 *  CHAPTER 39: DATA STRUCTURES IN JAVA
 * ============================================================
 *  Implementing classic data structures from scratch gives you
 *  deep understanding that makes you a god-level programmer.
 *
 *  COVERED:
 *    1. Linked List (Singly + Doubly)
 *    2. Stack
 *    3. Queue
 *    4. Binary Search Tree (BST)
 *    5. Hash Map (from scratch)
 *    6. Heap / Priority Queue
 *    7. Graph (adjacency list)
 *    8. Trie (prefix tree)
 * ============================================================
 */

import java.util.*;

public class Chapter39_DataStructures {

    // ========================================================
    // 1. SINGLY LINKED LIST
    // ========================================================
    static class SinglyLinkedList<T> {
        private static class Node<T> {
            T data;
            Node<T> next;
            Node(T data) { this.data = data; }
        }

        private Node<T> head;
        private int size;

        void addFirst(T data) {
            Node<T> node = new Node<>(data);
            node.next = head;
            head = node;
            size++;
        }

        void addLast(T data) {
            Node<T> node = new Node<>(data);
            if (head == null) { head = node; }
            else {
                Node<T> curr = head;
                while (curr.next != null) curr = curr.next;
                curr.next = node;
            }
            size++;
        }

        T removeFirst() {
            if (head == null) throw new NoSuchElementException();
            T data = head.data;
            head = head.next;
            size--;
            return data;
        }

        void reverse() {
            Node<T> prev = null, curr = head, next;
            while (curr != null) {
                next = curr.next;
                curr.next = prev;
                prev = curr;
                curr = next;
            }
            head = prev;
        }

        int size() { return size; }

        @Override
        public String toString() {
            StringBuilder sb = new StringBuilder("[");
            Node<T> curr = head;
            while (curr != null) {
                sb.append(curr.data);
                if (curr.next != null) sb.append(" → ");
                curr = curr.next;
            }
            return sb.append("]").toString();
        }
    }

    // ========================================================
    // 2. STACK (LIFO)
    // ========================================================
    static class MyStack<T> {
        private Object[] data;
        private int top = -1;

        @SuppressWarnings("unchecked")
        MyStack(int capacity) { data = new Object[capacity]; }

        void push(T item) {
            if (top == data.length - 1) throw new StackOverflowError("Stack full");
            data[++top] = item;
        }

        @SuppressWarnings("unchecked")
        T pop() {
            if (top == -1) throw new EmptyStackException();
            T item = (T) data[top];
            data[top--] = null;
            return item;
        }

        @SuppressWarnings("unchecked")
        T peek() {
            if (top == -1) throw new EmptyStackException();
            return (T) data[top];
        }

        boolean isEmpty() { return top == -1; }
        int size() { return top + 1; }
    }

    // ========================================================
    // 3. QUEUE (FIFO) — Circular Array
    // ========================================================
    static class MyQueue<T> {
        private Object[] data;
        private int front, rear, size;

        @SuppressWarnings("unchecked")
        MyQueue(int capacity) { data = new Object[capacity]; }

        void enqueue(T item) {
            if (size == data.length) throw new IllegalStateException("Queue full");
            data[rear] = item;
            rear = (rear + 1) % data.length;
            size++;
        }

        @SuppressWarnings("unchecked")
        T dequeue() {
            if (size == 0) throw new NoSuchElementException();
            T item = (T) data[front];
            data[front] = null;
            front = (front + 1) % data.length;
            size--;
            return item;
        }

        boolean isEmpty() { return size == 0; }
        int size() { return size; }
    }

    // ========================================================
    // 4. BINARY SEARCH TREE
    // ========================================================
    static class BST {
        private static class Node {
            int val;
            Node left, right;
            Node(int val) { this.val = val; }
        }

        private Node root;

        void insert(int val) { root = insertRec(root, val); }

        private Node insertRec(Node node, int val) {
            if (node == null) return new Node(val);
            if (val < node.val) node.left = insertRec(node.left, val);
            else if (val > node.val) node.right = insertRec(node.right, val);
            return node;
        }

        boolean search(int val) { return searchRec(root, val); }

        private boolean searchRec(Node node, int val) {
            if (node == null) return false;
            if (val == node.val) return true;
            return val < node.val ? searchRec(node.left, val) : searchRec(node.right, val);
        }

        // In-order traversal → sorted output
        List<Integer> inOrder() {
            List<Integer> result = new ArrayList<>();
            inOrderRec(root, result);
            return result;
        }

        private void inOrderRec(Node node, List<Integer> result) {
            if (node == null) return;
            inOrderRec(node.left, result);
            result.add(node.val);
            inOrderRec(node.right, result);
        }

        // Level-order (BFS)
        List<Integer> levelOrder() {
            List<Integer> result = new ArrayList<>();
            if (root == null) return result;
            Queue<Node> queue = new LinkedList<>();
            queue.add(root);
            while (!queue.isEmpty()) {
                Node node = queue.poll();
                result.add(node.val);
                if (node.left != null) queue.add(node.left);
                if (node.right != null) queue.add(node.right);
            }
            return result;
        }

        int height() { return heightRec(root); }

        private int heightRec(Node node) {
            if (node == null) return -1;
            return 1 + Math.max(heightRec(node.left), heightRec(node.right));
        }
    }

    // ========================================================
    // 5. HASH MAP (simplified)
    // ========================================================
    static class MyHashMap<K, V> {
        private static class Entry<K, V> {
            K key;
            V value;
            Entry<K, V> next;  // chaining for collisions
            Entry(K key, V value) { this.key = key; this.value = value; }
        }

        @SuppressWarnings("unchecked")
        private Entry<K, V>[] buckets = new Entry[16];
        private int size;

        private int getBucket(K key) {
            return Math.abs(key.hashCode() % buckets.length);
        }

        void put(K key, V value) {
            int idx = getBucket(key);
            Entry<K, V> curr = buckets[idx];
            while (curr != null) {
                if (curr.key.equals(key)) { curr.value = value; return; }
                curr = curr.next;
            }
            Entry<K, V> entry = new Entry<>(key, value);
            entry.next = buckets[idx];
            buckets[idx] = entry;
            size++;
        }

        V get(K key) {
            int idx = getBucket(key);
            Entry<K, V> curr = buckets[idx];
            while (curr != null) {
                if (curr.key.equals(key)) return curr.value;
                curr = curr.next;
            }
            return null;
        }

        boolean containsKey(K key) { return get(key) != null; }
        int size() { return size; }
    }

    // ========================================================
    // 6. MIN HEAP
    // ========================================================
    static class MinHeap {
        private int[] data;
        private int size;

        MinHeap(int capacity) { data = new int[capacity]; }

        void insert(int val) {
            data[size] = val;
            siftUp(size);
            size++;
        }

        int extractMin() {
            int min = data[0];
            data[0] = data[--size];
            siftDown(0);
            return min;
        }

        int peek() { return data[0]; }

        private void siftUp(int i) {
            while (i > 0) {
                int parent = (i - 1) / 2;
                if (data[i] < data[parent]) {
                    swap(i, parent);
                    i = parent;
                } else break;
            }
        }

        private void siftDown(int i) {
            while (2 * i + 1 < size) {
                int child = 2 * i + 1;
                if (child + 1 < size && data[child + 1] < data[child]) child++;
                if (data[i] > data[child]) {
                    swap(i, child);
                    i = child;
                } else break;
            }
        }

        private void swap(int a, int b) {
            int tmp = data[a]; data[a] = data[b]; data[b] = tmp;
        }

        int size() { return size; }
    }

    // ========================================================
    // 7. GRAPH (adjacency list)
    // ========================================================
    static class Graph {
        private Map<String, List<String>> adjList = new LinkedHashMap<>();

        void addVertex(String v) { adjList.putIfAbsent(v, new ArrayList<>()); }

        void addEdge(String from, String to) {
            addVertex(from);
            addVertex(to);
            adjList.get(from).add(to);
            adjList.get(to).add(from);  // undirected
        }

        // BFS
        List<String> bfs(String start) {
            List<String> visited = new ArrayList<>();
            Queue<String> queue = new LinkedList<>();
            Set<String> seen = new HashSet<>();

            queue.add(start);
            seen.add(start);

            while (!queue.isEmpty()) {
                String node = queue.poll();
                visited.add(node);
                for (String neighbor : adjList.getOrDefault(node, List.of())) {
                    if (seen.add(neighbor)) queue.add(neighbor);
                }
            }
            return visited;
        }

        // DFS
        List<String> dfs(String start) {
            List<String> visited = new ArrayList<>();
            dfsRec(start, new HashSet<>(), visited);
            return visited;
        }

        private void dfsRec(String node, Set<String> seen, List<String> visited) {
            if (!seen.add(node)) return;
            visited.add(node);
            for (String neighbor : adjList.getOrDefault(node, List.of())) {
                dfsRec(neighbor, seen, visited);
            }
        }
    }

    // ========================================================
    // 8. TRIE (Prefix Tree)
    // ========================================================
    static class Trie {
        private static class TrieNode {
            TrieNode[] children = new TrieNode[26];
            boolean isEnd;
        }

        private TrieNode root = new TrieNode();

        void insert(String word) {
            TrieNode node = root;
            for (char c : word.toCharArray()) {
                int idx = c - 'a';
                if (node.children[idx] == null) node.children[idx] = new TrieNode();
                node = node.children[idx];
            }
            node.isEnd = true;
        }

        boolean search(String word) {
            TrieNode node = findNode(word);
            return node != null && node.isEnd;
        }

        boolean startsWith(String prefix) {
            return findNode(prefix) != null;
        }

        private TrieNode findNode(String s) {
            TrieNode node = root;
            for (char c : s.toCharArray()) {
                int idx = c - 'a';
                if (node.children[idx] == null) return null;
                node = node.children[idx];
            }
            return node;
        }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        // --- 1. Linked List ---
        System.out.println("=== LINKED LIST ===\n");
        SinglyLinkedList<Integer> list = new SinglyLinkedList<>();
        list.addLast(1); list.addLast(2); list.addLast(3); list.addFirst(0);
        System.out.println("  " + list);
        list.reverse();
        System.out.println("  Reversed: " + list);
        System.out.println("  Removed: " + list.removeFirst());
        System.out.println("  After remove: " + list);

        // --- 2. Stack ---
        System.out.println("\n=== STACK ===\n");
        MyStack<String> stack = new MyStack<>(10);
        stack.push("A"); stack.push("B"); stack.push("C");
        System.out.println("  Pop: " + stack.pop());
        System.out.println("  Peek: " + stack.peek());
        System.out.println("  Size: " + stack.size());

        // --- 3. Queue ---
        System.out.println("\n=== QUEUE ===\n");
        MyQueue<String> queue = new MyQueue<>(10);
        queue.enqueue("First"); queue.enqueue("Second"); queue.enqueue("Third");
        System.out.println("  Dequeue: " + queue.dequeue());
        System.out.println("  Dequeue: " + queue.dequeue());
        System.out.println("  Size: " + queue.size());

        // --- 4. BST ---
        System.out.println("\n=== BINARY SEARCH TREE ===\n");
        BST bst = new BST();
        int[] vals = {8, 3, 10, 1, 6, 14, 4, 7, 13};
        for (int v : vals) bst.insert(v);
        System.out.println("  In-order (sorted): " + bst.inOrder());
        System.out.println("  Level-order (BFS): " + bst.levelOrder());
        System.out.println("  Height: " + bst.height());
        System.out.println("  Search 6: " + bst.search(6));
        System.out.println("  Search 5: " + bst.search(5));

        // --- 5. HashMap ---
        System.out.println("\n=== CUSTOM HASH MAP ===\n");
        MyHashMap<String, Integer> map = new MyHashMap<>();
        map.put("apple", 1); map.put("banana", 2); map.put("cherry", 3);
        System.out.println("  apple: " + map.get("apple"));
        System.out.println("  banana: " + map.get("banana"));
        map.put("apple", 10);  // update
        System.out.println("  apple (updated): " + map.get("apple"));
        System.out.println("  contains 'grape': " + map.containsKey("grape"));

        // --- 6. Min Heap ---
        System.out.println("\n=== MIN HEAP ===\n");
        MinHeap heap = new MinHeap(20);
        heap.insert(5); heap.insert(3); heap.insert(8);
        heap.insert(1); heap.insert(7);
        System.out.println("  Min: " + heap.peek());
        System.out.println("  Extract: " + heap.extractMin());
        System.out.println("  Extract: " + heap.extractMin());
        System.out.println("  Extract: " + heap.extractMin());

        // --- 7. Graph ---
        System.out.println("\n=== GRAPH ===\n");
        Graph graph = new Graph();
        graph.addEdge("A", "B");
        graph.addEdge("A", "C");
        graph.addEdge("B", "D");
        graph.addEdge("C", "D");
        graph.addEdge("D", "E");

        System.out.println("  BFS from A: " + graph.bfs("A"));
        System.out.println("  DFS from A: " + graph.dfs("A"));

        // --- 8. Trie ---
        System.out.println("\n=== TRIE ===\n");
        Trie trie = new Trie();
        trie.insert("apple");
        trie.insert("app");
        trie.insert("application");
        trie.insert("banana");

        System.out.println("  search 'apple': " + trie.search("apple"));
        System.out.println("  search 'app': " + trie.search("app"));
        System.out.println("  search 'ap': " + trie.search("ap"));
        System.out.println("  startsWith 'ap': " + trie.startsWith("ap"));
        System.out.println("  startsWith 'ban': " + trie.startsWith("ban"));
        System.out.println("  startsWith 'cat': " + trie.startsWith("cat"));

        // --- Complexity Summary ---
        System.out.println("\n=== TIME COMPLEXITY SUMMARY ===");
        System.out.println("  Array:       Access O(1)   Search O(n)   Insert O(n)   Delete O(n)");
        System.out.println("  LinkedList:  Access O(n)   Search O(n)   Insert O(1)*  Delete O(1)*");
        System.out.println("  Stack:       Push O(1)     Pop O(1)      Peek O(1)");
        System.out.println("  Queue:       Enqueue O(1)  Dequeue O(1)  Peek O(1)");
        System.out.println("  BST:         Search O(log n)  Insert O(log n)  *Worst O(n)");
        System.out.println("  HashMap:     Get O(1)      Put O(1)      *Worst O(n)");
        System.out.println("  Heap:        Insert O(log n)  Extract O(log n)  Peek O(1)");
        System.out.println("  Trie:        Insert O(m)   Search O(m)   m = word length");

        System.out.println("\n✓ Data Structures Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Implement a doubly linked list with addLast, removeLast, get(index).
 * 2. Implement BST deletion (3 cases: leaf, one child, two children).
 * 3. Implement a hash map with resizing (load factor > 0.75 → double capacity).
 * 4. Add shortest path (Dijkstra) to the Graph class.
 *
 * NEXT: Chapter 40 — Algorithms in Java
 */
