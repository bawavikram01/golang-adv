# 1.1 — Relational Model Theory

> The relational model is the **mathematical foundation** underneath every SQL database.
> If you don't understand this, you're building on sand.

---

## 1. What Is the Relational Model?

In 1970, **Edgar F. Codd** (an IBM researcher) published a paper that changed computing forever:
*"A Relational Model of Data for Large Shared Data Banks"*

Before Codd, databases were **navigational** — you had to write code to traverse pointers and physical links between records (hierarchical model, network model). It was like walking through a maze.

Codd said: **forget the physical structure. Data is just math. It's sets of tuples.**

### The Core Idea

A database is a collection of **relations** (tables).
Each relation is a **set of tuples** (rows).
Each tuple contains **attribute values** (columns).
Operations on relations produce **new relations** (closure property).

That's it. Everything else in SQL flows from this.

---

## 2. Formal Definitions

### Relation

A relation R is defined over a **relation schema** R(A₁, A₂, ..., Aₙ) where:
- R is the relation name
- A₁, A₂, ..., Aₙ are **attributes** (column names)
- Each attribute Aᵢ has a **domain** dom(Aᵢ) — the set of allowed values

A relation instance r(R) is a **set of n-tuples**:

```
r(R) ⊆ dom(A₁) × dom(A₂) × ... × dom(Aₙ)
```

### What This Means in Plain English

```
Schema:   Student(id, name, age, gpa)
Domains:  id ∈ integers, name ∈ strings, age ∈ {1..150}, gpa ∈ {0.0..4.0}

Instance (the actual data):
┌────┬────────┬─────┬─────┐
│ id │ name   │ age │ gpa │
├────┼────────┼─────┼─────┤
│  1 │ Alice  │  20 │ 3.8 │
│  2 │ Bob    │  22 │ 3.5 │
│  3 │ Carol  │  21 │ 3.9 │
└────┴────────┴─────┴─────┘
```

### Key Properties of Relations

1. **No duplicate tuples** — a relation is a SET. {(1, Alice), (1, Alice)} is NOT a valid relation. (SQL tables violate this unless you have a primary key — SQL is not pure relational algebra!)

2. **Tuples are unordered** — there's no "first row" or "last row". The set {(1, Alice), (2, Bob)} = {(2, Bob), (1, Alice)}.

3. **Attributes are unordered** — columns have no inherent order. You refer to them by NAME, not position.

4. **Attribute values are atomic** — no nested tables, no arrays, no structs within a cell. This is **First Normal Form (1NF)**. (Modern SQL relaxes this with JSON, arrays, etc.)

---

## 3. Keys

### Superkey
Any set of attributes that **uniquely identifies** a tuple.
For Student(id, name, age, gpa):
- {id} is a superkey
- {id, name} is a superkey (adding attributes to a superkey gives another superkey)
- {name, age} might be a superkey if no two students share name+age

### Candidate Key
A **minimal superkey** — remove any attribute and it's no longer a superkey.
- {id} is a candidate key (can't remove anything)
- {id, name} is NOT a candidate key ({id} alone suffices)

### Primary Key
The candidate key you **choose** to be THE identifier. Convention: one per table.

### Foreign Key
An attribute (or set) in one relation that **references** the primary key of another relation.

```
Student(id, name, department_id)
Department(id, name)

Student.department_id → Department.id   (foreign key)
```

This is how you express **relationships** in the relational model.

### How to Find Candidate Keys (Algorithm)

Given a set of functional dependencies F over relation R(A₁, ..., Aₙ):

1. Find attributes that appear ONLY on the LEFT side of FDs → these MUST be in every key
2. Find attributes that appear ONLY on the RIGHT side of FDs → these are NEVER in any key
3. Find attributes that appear on NEITHER side → these MUST be in every key
4. Start with the must-have attributes, compute their closure. If it covers all attributes, that's a candidate key.
5. If not, try adding remaining attributes one at a time and check closure.

---

## 4. Relational Algebra

Relational algebra is a **procedural** query language — you specify HOW to get the result using a sequence of operations.

Every operation takes one or two relations as input and produces a new relation as output (**closure property**). This is why you can chain operations.

### Unary Operations (one input relation)

#### σ — Selection (sigma)
Filters ROWS based on a condition (like SQL WHERE).

```
σ_age>21(Student)

Means: give me all students where age > 21

SQL equivalent: SELECT * FROM Student WHERE age > 21
```

#### π — Projection (pi)
Selects COLUMNS and removes duplicates.

```
π_name,gpa(Student)

Means: give me just the name and gpa columns

SQL equivalent: SELECT DISTINCT name, gpa FROM Student
```

Note: projection removes duplicates because the result must be a valid relation (a set).

#### ρ — Rename (rho)
Renames a relation or its attributes.

```
ρ_S(A,B,C)(R)

Means: rename relation R to S, and its attributes to A, B, C
```

### Binary Operations (two input relations)

#### ∪ — Union
Combines tuples from two relations. Both must have the **same schema** (union-compatible).

```
StudentA ∪ StudentB

All students from either A or B (duplicates removed).

SQL: SELECT * FROM A UNION SELECT * FROM B
```

#### ∩ — Intersection
Tuples that exist in BOTH relations.

```
StudentA ∩ StudentB

SQL: SELECT * FROM A INTERSECT SELECT * FROM B
```

#### − — Set Difference
Tuples in the first relation but NOT in the second.

```
StudentA − StudentB

SQL: SELECT * FROM A EXCEPT SELECT * FROM B
```

#### × — Cartesian Product (Cross Product)
Every tuple from R paired with every tuple from S. If R has m tuples and S has n tuples, R × S has m × n tuples.

```
Student × Department

If 3 students × 5 departments = 15 rows

SQL: SELECT * FROM Student CROSS JOIN Department
(or: SELECT * FROM Student, Department)
```

This is almost never useful alone — you always follow with a selection.

### The Join Family

#### ⋈ — Natural Join
Combines two relations on ALL shared attribute names, keeping only matching tuples.

```
Student ⋈ Enrollment

Matches on any columns with the same name in both tables.
Equivalent to: σ_condition(Student × Enrollment) then remove duplicate columns.
```

#### ⋈_θ — Theta Join
A cross product followed by a selection on condition θ.

```
Student ⋈_(Student.dept_id = Department.id) Department

SQL: SELECT * FROM Student JOIN Department ON Student.dept_id = Department.id
```

#### Equijoin
A theta join where θ uses only equality (=). This is the most common join.

#### Left/Right/Full Outer Join
- Left outer: keep all tuples from left relation, NULL-pad if no match on right
- Right outer: keep all tuples from right, NULL-pad left
- Full outer: keep all tuples from both, NULL-pad missing sides

#### Semijoin (⋉)
Returns tuples from R that have a matching tuple in S, but only keeps R's attributes.

```
Student ⋉ Enrollment

Give me students who have at least one enrollment.

SQL: SELECT Student.* FROM Student WHERE EXISTS (SELECT 1 FROM Enrollment WHERE Enrollment.student_id = Student.id)
```

#### Antijoin (▷)
Returns tuples from R that have NO matching tuple in S.

```
Student ▷ Enrollment

Students with zero enrollments.

SQL: SELECT * FROM Student WHERE NOT EXISTS (SELECT 1 FROM Enrollment WHERE ...)
```

### Division (÷)
The most complex operation. "Give me tuples in R that are associated with ALL tuples in S."

```
StudentCourse ÷ RequiredCourse

Students who have taken ALL required courses.

SQL (tricky):
SELECT student_id FROM StudentCourse
WHERE course_id IN (SELECT id FROM RequiredCourse)
GROUP BY student_id
HAVING COUNT(DISTINCT course_id) = (SELECT COUNT(*) FROM RequiredCourse)
```

### Aggregate Operations (Extended Relational Algebra)

```
γ_department, AVG(gpa)(Student)

Group by department, compute average GPA.

SQL: SELECT department, AVG(gpa) FROM Student GROUP BY department
```

---

## 5. Relational Calculus

Relational calculus is **declarative** — you specify WHAT you want, not HOW to get it.

### Tuple Relational Calculus (TRC)

```
{ t | P(t) }

"The set of all tuples t such that predicate P(t) is true"
```

Example: Find students with GPA > 3.5
```
{ t | Student(t) ∧ t.gpa > 3.5 }
```

Example: Find names of students in CS department
```
{ t.name | Student(t) ∧ t.department = 'CS' }
```

Example: Find students enrolled in ALL courses (universal quantification)
```
{ t | Student(t) ∧ ∀c(Course(c) → ∃e(Enrollment(e) ∧ e.student_id = t.id ∧ e.course_id = c.id)) }
```

### Domain Relational Calculus (DRC)

Uses domain variables (individual values) instead of tuple variables.

```
{ <n> | ∃i,a,g (Student(i, n, a, g) ∧ g > 3.5) }

"Give me names where there exists an id, age, and gpa such that a Student tuple with those values has gpa > 3.5"
```

### Why This Matters

**Codd's theorem**: Relational algebra and (safe) relational calculus are **equivalent in expressive power**.

SQL is based on tuple relational calculus. The query optimizer translates your declarative SQL into an algebra execution plan.

---

## 6. Functional Dependencies (FDs)

A functional dependency X → Y means: if two tuples agree on attributes X, they MUST agree on attributes Y.

```
StudentID → Name, Age, GPA

"If two rows have the same StudentID, they must have the same Name, Age, and GPA"
```

This is a **constraint** on the data, not a query.

### Armstrong's Axioms

These are **sound and complete** inference rules for FDs:

1. **Reflexivity**: If Y ⊆ X, then X → Y (trivial FD)
2. **Augmentation**: If X → Y, then XZ → YZ
3. **Transitivity**: If X → Y and Y → Z, then X → Z

Derived rules (provable from the axioms):
4. **Union**: If X → Y and X → Z, then X → YZ
5. **Decomposition**: If X → YZ, then X → Y and X → Z
6. **Pseudotransitivity**: If X → Y and WY → Z, then WX → Z

### Closure of Attributes (X⁺)

The closure of a set of attributes X under a set of FDs F is the set of ALL attributes functionally determined by X.

**Algorithm to compute X⁺:**
```
Input: X (set of attributes), F (set of FDs)

result = X
repeat:
    for each FD (A → B) in F:
        if A ⊆ result:
            result = result ∪ B
until result doesn't change

return result
```

**Example:**
```
R(A, B, C, D, E)
F = { A → B, B → C, CD → E, A → D }

Compute {A}⁺:
- Start: {A}
- A → B:  {A, B}
- B → C:  {A, B, C}
- A → D:  {A, B, C, D}
- CD → E: {A, B, C, D, E}  ← All attributes!

Since {A}⁺ = {A,B,C,D,E} = all attributes, A is a candidate key.
```

### Canonical Cover (Minimal Cover)

A minimal set of FDs equivalent to F, with:
- No redundant FDs
- No extraneous attributes on left or right sides
- Each FD has a single attribute on the right

This is important for normalization (BCNF decomposition).

---

## 7. Codd's 12 Rules (Actually 13, Numbered 0-12)

These define what a RDBMS should support. No system fully satisfies all of them.

| # | Rule | Meaning |
|---|------|---------|
| 0 | Foundation Rule | Must use relational model to manage data |
| 1 | Information Rule | All data represented as values in tables |
| 2 | Guaranteed Access | Every value accessible by table+column+primary key |
| 3 | Systematic NULL Treatment | NULLs for missing/inapplicable data |
| 4 | Active Online Catalog | Database structure stored in same relational tables |
| 5 | Comprehensive Data Sublanguage | At least one language (SQL) for all DB operations |
| 6 | View Updating Rule | All theoretically updatable views must be updatable |
| 7 | High-Level Insert/Update/Delete | Set-at-a-time operations, not just row-at-a-time |
| 8 | Physical Data Independence | Apps unaffected by storage/access method changes |
| 9 | Logical Data Independence | Apps unaffected by table restructuring |
| 10 | Integrity Independence | Constraints defined in catalog, not in applications |
| 11 | Distribution Independence | Apps work same whether data is distributed or not |
| 12 | Nonsubversion Rule | Can't bypass integrity constraints using low-level access |

---

## 8. Practice Problems

### Problem 1: Find Candidate Keys
```
R(A, B, C, D, E)
FDs: AB → C, C → D, D → E, E → A

Find all candidate keys.
```

<details>
<summary>Solution</summary>

- B appears only on the left side → B must be in every key
- Compute {B}⁺ = {B} → not all attributes
- Try {AB}⁺: AB → C → D → E → A. {AB}⁺ = {A,B,C,D,E} ✓ → AB is a candidate key
- Try {BC}⁺: C → D → E → A. {BC}⁺ = {A,B,C,D,E} ✓ → BC is a candidate key
- Try {BD}⁺: D → E → A, AB → C. {BD}⁺ = {A,B,C,D,E} ✓ → BD is a candidate key
- Try {BE}⁺: E → A, AB → C, C → D. {BE}⁺ = {A,B,C,D,E} ✓ → BE is a candidate key

Candidate keys: **AB, BC, BD, BE**
</details>

### Problem 2: Express in Relational Algebra
Given:
- Employee(id, name, dept_id, salary)
- Department(id, name, location)

"Find names of employees in the 'Engineering' department earning more than 100000"

<details>
<summary>Solution</summary>

```
π_Employee.name(
  σ_(Department.name='Engineering' ∧ Employee.salary>100000)(
    Employee ⋈_(Employee.dept_id = Department.id) Department
  )
)
```

SQL:
```sql
SELECT e.name
FROM Employee e
JOIN Department d ON e.dept_id = d.id
WHERE d.name = 'Engineering' AND e.salary > 100000;
```
</details>

### Problem 3: Tuple Relational Calculus
"Find students who are enrolled in every course offered by the CS department"

<details>
<summary>Solution</summary>

```
{ s.name | Student(s) ∧
  ∀c( (Course(c) ∧ c.department = 'CS') →
    ∃e(Enrollment(e) ∧ e.student_id = s.id ∧ e.course_id = c.id)
  )
}
```

"For every CS course, there exists an enrollment linking this student to that course."
</details>

---

## Key Takeaways

1. **Relations are sets of tuples** — always think mathematically
2. **Relational algebra = procedural** (HOW), **Relational calculus = declarative** (WHAT) — they're equivalent
3. **Functional dependencies** are constraints that drive normalization
4. **Closure computation** is the algorithm that unlocks key finding and normalization
5. SQL is an imperfect implementation of relational calculus — it allows duplicates, NULLs break logic, etc.

---

Next: [02-sql-beginner.md](02-sql-beginner.md) →
