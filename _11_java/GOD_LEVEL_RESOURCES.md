# God-Level Java Resources

## Books — Tier 1 (Must Read)

| Book | Author | Why |
|------|--------|-----|
| *Effective Java* (3rd Ed) | Joshua Bloch | **THE bible**. 90 best-practice items. Read after Ch 1-30. |
| *Java Concurrency in Practice* | Brian Goetz | Deepest concurrency book ever written. Covers JMM, atomics, everything. |
| *Java Performance* (2nd Ed) | Scott Oaks | JIT, GC tuning, JFR, benchmarking — matches Ch 48, 53, 60, 62. |
| *Java Puzzlers* | Joshua Bloch & Neal Gafter | Tricky edge cases that expose deep language understanding. |

## Books — Tier 2 (Go Deeper)

| Book | Author | Why |
|------|--------|-----|
| *JVM Performance Engineering* | Monica Beckwith | Modern GC, ZGC, Shenandoah, JFR deep dives. |
| *Java: The Complete Reference* (12th Ed) | Herbert Schildt | Exhaustive language reference, every corner covered. |
| *Optimizing Java* | Benjamin Evans | JVM internals, hardware sympathy, real optimization. |
| *Java Generics and Collections* | Maurice Naftalin | After Ch 19 & 47 — the full generics story. |

---

## Read the Source Code

This is what separates gods from mortals:

- `java.util.HashMap` — treeification, hash spreading, resize
- `java.util.concurrent.ConcurrentHashMap` — lock striping, CAS
- `java.lang.String` — compact strings, intern, hash caching
- `java.util.concurrent.ForkJoinPool` — work-stealing
- `java.lang.ThreadLocal` — thread-local storage implementation
- `java.util.stream.ReferencePipeline` — how streams chain

Browse at: **https://github.com/openjdk/jdk**

---

## Online Resources

| Resource | URL | Focus |
|----------|-----|-------|
| Baeldung | baeldung.com | Practical examples for every Java topic |
| JEPs (JDK Enhancement Proposals) | openjdk.org/jeps | Every Java feature's design rationale |
| Inside the JVM (Article Series) | shipilev.net | Aleksey Shipilëv — JVM performance god |
| Java Specialists Newsletter | javaspecialists.eu | Deep-dive weekly puzzles by Heinz Kabutz |
| Vlad Mihalcea's Blog | vladmihalcea.com | JPA/Hibernate + Java performance |

---

## YouTube Channels

| Channel | Why |
|---------|-----|
| **Venkat Subramaniam** | Best Java/FP talks. Watch *"Let's Get Lazy"* and *"Functional Programming"* |
| **Heinz Kabutz** | Java Specialists — extreme deep dives |
| **Devoxx / JFokus** | Conference talks by JDK engineers |
| **Java (official)** | Inside Java podcast + dev talks |
| **Defog Tech** | Concurrency, JVM internals visualized |

---

## Practice

| Platform | What |
|----------|------|
| **LeetCode** (Java) | Algorithms — use Java collections/streams |
| **Exercism** (Java track) | Idiomatic Java problem solving |
| **Contribute to OpenJDK** | Fix real bugs in the JDK itself |
| **Build from scratch** | Write your own HashMap, ThreadPool, DI container, ORM |

---

## The God-Level Path (in order)

1. Finish all 62 chapters *(you're here)*
2. Read *Effective Java* cover to cover
3. Read *Java Concurrency in Practice*
4. Read HashMap & ConcurrentHashMap source code
5. Read *Java Performance* + profile a real app with JFR
6. Contribute a patch to an open-source Java project
7. Read JEPs for Records, Sealed Classes, Virtual Threads — understand *why*, not just *what*
