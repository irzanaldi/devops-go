# Bab 4: Go Syntax & Pattern untuk Backend Dev

> Ini bab paling panjang dan paling penting. Semua fundamental Go ada di sini — dari variable sampai goroutine. Setiap konsep dibandingin langsung sama JavaScript/TypeScript yang udah kamu kuasai. Anggep bab ini kayak "kamus" yang bakal sering kamu buka ulang.

---

## 1. Variable & Types

### 1.1 Deklarasi Variable

**JavaScript/TypeScript:**
```javascript
let name = "Irzan";         // bisa di-reassign
const age = 25;              // ga bisa di-reassign
var legacy = "jangan pake";  // function-scoped, avoid
```

**Go:**
```go
// Cara 1: var (explicit type)
var name string = "Irzan"
var age int = 25

// Cara 2: var (type inference — Go tebak sendiri)
var name = "Irzan"    // Go tau ini string
var age = 25          // Go tau ini int

// Cara 3: short declaration := (paling sering dipake)
name := "Irzan"       // otomatis string
age := 25             // otomatis int

// Cara 4: multiple declaration
var (
    host     string = "localhost"
    port     int    = 8080
    debug    bool   = true
)
```

### 1.2 Const

**JavaScript:**
```javascript
const API_URL = "https://api.example.com";
const MAX_RETRY = 3;
```

**Go:**
```go
const ApiURL = "https://api.example.com"
const MaxRetry = 3

// Grouped const
const (
    StatusPending  = "pending"
    StatusActive   = "active"
    StatusDone     = "done"
)

// iota — auto-increment (ga ada di JS)
const (
    RoleUser  = iota // 0
    RoleAdmin        // 1
    RoleSuper        // 2
)
```

### 1.3 Basic Types

| Go Type | JS/TS Equivalent | Contoh | Zero Value |
|---------|------------------|--------|------------|
| `string` | `string` | `"hello"` | `""` |
| `int` | `number` | `42` | `0` |
| `int64` | `number` (BigInt) | `9999999999` | `0` |
| `float64` | `number` | `3.14` | `0` |
| `bool` | `boolean` | `true` | `false` |
| `byte` | — | `'A'` | `0` |
| `rune` | — | `'Z'` (Unicode) | `0` |

### 1.4 Zero Values

Ini konsep yang **ga ada di JavaScript**. Di Go, semua variable pasti punya nilai default (zero value) — ga ada `undefined` atau `null`.

**JavaScript:**
```javascript
let name;          // undefined
let obj;           // undefined
console.log(name); // undefined — sumber bug tersembunyi
```

**Go:**
```go
var name string    // "" (string kosong)
var age int        // 0
var price float64  // 0
var active bool    // false
var data []int     // nil (slice kosong)
var m map[string]int // nil

fmt.Println(name)  // "" — selalu predictable
fmt.Println(age)   // 0 — ga pernah undefined
```

> **Gotcha:** Di Go ga ada `undefined` dan ga ada `null` untuk basic type. Zero value bikin program lebih predictable. Tapi hati-hati: `nil` map atau `nil` pointer bisa panic kalau diakses langsung.

### 1.5 Type Conversion

**JavaScript (implicit, sering bikin bug):**
```javascript
"5" + 3    // "53" — string concat, bukan addition!
"5" - 3    // 2 — suddenly jadi number
true + 1   // 2 — what??
```

**Go (explicit, selalu aman):**
```go
// Go TIDAK pernah convert type otomatis
var i int = 42
var f float64 = float64(i)  // harus explicit
var u uint = uint(f)        // harus explicit

// String ke int
num, err := strconv.Atoi("42")  // Atoi = Ascii to Integer
if err != nil {
    log.Fatal("bukan angka!")
}

// Int ke string
str := strconv.Itoa(42)  // "42"
```

> **Gotcha:** Go ga akan compile kalau kamu coba assign `int` ke `float64` tanpa conversion. Ini ketat, tapi mencegah bug kayak `"5" + 3 = "53"` di JavaScript.

---

## 2. String

### 2.1 String Concatenation

**JavaScript:**
```javascript
const name = "Irzan";
const greeting = "Hello, " + name + "!";
```

**Go:**
```go
name := "Irzan"
greeting := "Hello, " + name + "!"

// Untuk banyak concatenation, pakai strings.Builder (lebih efisien)
var builder strings.Builder
builder.WriteString("Hello, ")
builder.WriteString(name)
builder.WriteString("!")
result := builder.String()
```

### 2.2 Template Literal vs fmt.Sprintf

**JavaScript (template literal):**
```javascript
const name = "Irzan";
const age = 25;
const msg = `Halo ${name}, umur kamu ${age} tahun`;
```

**Go (fmt.Sprintf):**
```go
name := "Irzan"
age := 25
msg := fmt.Sprintf("Halo %s, umur kamu %d tahun", name, age)
// %s = string, %d = integer, %f = float, %v = any value, %+v = struct detail
```

**Format specifiers yang sering dipake:**

| Specifier | Fungsi | Contoh |
|-----------|--------|--------|
| `%s` | string | `fmt.Sprintf("%s", "hello")` |
| `%d` | integer | `fmt.Sprintf("%d", 42)` |
| `%f` | float | `fmt.Sprintf("%.2f", 3.14)` → `"3.14"` |
| `%v` | value apapun | `fmt.Sprintf("%v", someStruct)` |
| `%+v` | struct + field name | `fmt.Sprintf("%+v", user)` |
| `%T` | tipe data | `fmt.Sprintf("%T", 42)` → `"int"` |
| `%t` | boolean | `fmt.Sprintf("%t", true)` → `"true"` |

### 2.3 Multiline String

**JavaScript:**
```javascript
const query = `
  SELECT *
  FROM users
  WHERE active = true
`;
```

**Go:**
```go
// Backtick = raw string literal (kayak template literal tapi tanpa interpolation)
query := `
  SELECT *
  FROM users
  WHERE active = true
`
// Perhatian: backtick di Go ga support interpolation (ga ada ${var})
// Kalau butuh variable, pakai fmt.Sprintf
query := fmt.Sprintf(`
  SELECT *
  FROM users
  WHERE id = %d
`, userID)
```

### 2.4 Package strings

**JavaScript:**
```javascript
"hello".toUpperCase();            // "HELLO"
"HELLO".toLowerCase();            // "hello"
"hello world".includes("world");  // true
"hello world".split(" ");         // ["hello", "world"]
"  hello  ".trim();               // "hello"
"hello world".replace("world", "Go"); // "hello Go"
"ha".repeat(3);                   // "hahaha"
```

**Go:**
```go
import "strings"

strings.ToUpper("hello")              // "HELLO"
strings.ToLower("HELLO")              // "hello"
strings.Contains("hello world", "world") // true
strings.Split("hello world", " ")     // []string{"hello", "world"}
strings.TrimSpace("  hello  ")        // "hello"
strings.Replace("hello world", "world", "Go", 1) // "hello Go"
strings.Repeat("ha", 3)               // "hahaha"
strings.HasPrefix("hello", "hel")     // true
strings.HasSuffix("hello", "llo")     // true
strings.Join([]string{"a", "b"}, ",") // "a,b"
strings.Count("hello", "l")           // 2
```

### 2.5 Package strconv (String Conversion)

```go
import "strconv"

// String → Int
n, err := strconv.Atoi("42")          // n = 42
// Int → String
s := strconv.Itoa(42)                 // s = "42"
// String → Float
f, err := strconv.ParseFloat("3.14", 64) // f = 3.14
// Float → String
s := strconv.FormatFloat(3.14, 'f', 2, 64) // s = "3.14"
// String → Bool
b, err := strconv.ParseBool("true")   // b = true
```

> **Gotcha:** Di JavaScript, `parseInt("42abc")` return `42` (silent truncation). Di Go, `strconv.Atoi("42abc")` return **error**. Go ga mau nebak-nebak.

---

## 3. Array & Slice

### 3.1 Array (Fixed Size)

**JavaScript:**
```javascript
// JS ga punya fixed-size array
const arr = [1, 2, 3]; // ini dynamic
```

**Go:**
```go
// Array: ukuran FIXED, ditentukan saat compile
var scores [3]int           // [0, 0, 0] — zero value
scores[0] = 90
scores[1] = 85
scores[2] = 92

names := [3]string{"Alice", "Bob", "Charlie"}

// Ukuran = bagian dari type!
var a [3]int
var b [5]int
// a = b  // COMPILE ERROR! [3]int dan [5]int itu type berbeda
```

> **Gotcha:** Array di Go **jarang dipake** secara langsung. Yang sering dipake itu **Slice**. Array cuma jadi underlying structure di belakang slice.

### 3.2 Slice (Dynamic — yang sering dipake)

**JavaScript:**
```javascript
const fruits = ["apple", "banana"];
fruits.push("cherry");         // add to end
fruits.length;                 // 3
const sliced = fruits.slice(0, 2); // ["apple", "banana"]
```

**Go:**
```go
// Cara 1: literal
fruits := []string{"apple", "banana"} // perhatikan: TANPA angka di []

// Cara 2: make — bikin slice dengan length dan capacity
scores := make([]int, 5)     // len=5, cap=5, isi [0,0,0,0,0]
scores := make([]int, 0, 10) // len=0, cap=10, isi [] (kosong tapi udah alokasi 10 slot)

// Cara 3: dari array
arr := [5]int{1, 2, 3, 4, 5}
slice := arr[1:4] // [2, 3, 4] — index 1 sampai 3 (4 exclusive)
```

### 3.3 Operasi Slice

```go
fruits := []string{"apple", "banana"}

// Append (kayak push di JS)
fruits = append(fruits, "cherry")
fruits = append(fruits, "date", "elderberry") // bisa banyak sekaligus

// Length
len(fruits) // 5

// Capacity (total alokasi memori — ga ada di JS)
cap(fruits) // bervariasi, Go manage otomatis

// Slicing
sub := fruits[1:3]  // ["banana", "cherry"]
sub := fruits[:2]   // ["apple", "banana"] — dari awal sampai index 1
sub := fruits[2:]   // ["cherry", "date", "elderberry"] — dari index 2 sampai akhir

// Loop
for i, fruit := range fruits {
    fmt.Printf("Index %d: %s\n", i, fruit)
}

// Cek apakah slice kosong
if len(fruits) == 0 {
    fmt.Println("slice kosong")
}
```

### 3.4 Common Slice Operations

**JavaScript:**
```javascript
// Filter
const adults = users.filter(u => u.age >= 18);
// Map
const names = users.map(u => u.name);
// Find
const found = users.find(u => u.id === 5);
// Includes
const exists = arr.includes("apple");
```

**Go (manual — ga ada built-in filter/map):**
```go
// Filter
var adults []User
for _, u := range users {
    if u.Age >= 18 {
        adults = append(adults, u)
    }
}

// Map (transform)
var names []string
for _, u := range users {
    names = append(names, u.Name)
}

// Find
var found *User
for _, u := range users {
    if u.ID == 5 {
        found = &u
        break
    }
}

// Includes / Contains
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
// Sejak Go 1.21, bisa pakai slices.Contains()
import "slices"
exists := slices.Contains(fruits, "apple")
```

> **Gotcha:** Go **sengaja** ga punya `.filter()`, `.map()`, `.reduce()` bawaan kayak JS. Filosofinya: explicit loop lebih mudah dibaca dan di-debug. Tapi sejak Go 1.21+, package `slices` dan `maps` nambah banyak helper.

### 3.5 Slice Internal: Length vs Capacity

Ini penting buat performance. JavaScript sembunyiin ini, Go ekspos.

```
Slice: [apple] [banana] [cherry] [     ] [     ]
        ^                         ^
        |--- length = 3 ---------|
        |--- capacity = 5 -------------------|
```

```go
s := make([]int, 3, 5)
fmt.Println(len(s)) // 3 — jumlah elemen saat ini
fmt.Println(cap(s)) // 5 — total slot yang sudah dialokasi

// Kalau append melebihi capacity, Go alokasi array baru (2x lipat)
// Ini kayak ArrayList di Java — Go handle otomatis, tapi kalau tau ukurannya,
// lebih efisien pakai make([]T, 0, expectedSize)
```

> **Tips Performance:** Kalau kamu tau kira-kira slice bakal isi 1000 item, bikin `make([]Item, 0, 1000)`. Ini mencegah Go realokasi berkali-kali.

---

## 4. Map

### 4.1 Deklarasi Map

**JavaScript:**
```javascript
// Object (paling common)
const user = { name: "Irzan", age: 25 };

// Map (proper key-value)
const map = new Map();
map.set("name", "Irzan");
map.get("name"); // "Irzan"
```

**Go:**
```go
// Cara 1: literal
user := map[string]string{
    "name":  "Irzan",
    "email": "irzan@example.com",
}

// Cara 2: make
scores := make(map[string]int)
scores["math"] = 90
scores["english"] = 85

// Map dengan tipe kompleks
users := map[int]User{
    1: {Name: "Irzan", Age: 25},
    2: {Name: "Budi", Age: 30},
}
```

### 4.2 CRUD Operations

**JavaScript:**
```javascript
const m = {};
m["name"] = "Irzan";        // Create / Update
console.log(m["name"]);     // Read — "Irzan"
delete m["name"];            // Delete
console.log(m["age"]);      // undefined (key ga ada)
```

**Go:**
```go
m := make(map[string]string)

// Create / Update
m["name"] = "Irzan"

// Read
name := m["name"]    // "Irzan"
age := m["age"]      // "" (zero value, bukan error!)

// Delete
delete(m, "name")

// Iterate
for key, value := range m {
    fmt.Printf("%s: %s\n", key, value)
}
```

### 4.3 Comma Ok Pattern (Cek Key Exists)

Ini pattern unik Go yang ga ada di JS.

**JavaScript:**
```javascript
const obj = { name: "Irzan" };
if (obj.age !== undefined) {
    // key exists
}
// Masalah: gimana bedain "key ga ada" vs "key ada tapi value-nya undefined"?
```

**Go:**
```go
m := map[string]int{
    "math":    90,
    "english": 0, // sengaja 0
}

// Cara salah — ga bisa bedain "key ga ada" vs "value-nya 0"
score := m["science"] // 0 — tapi key "science" ga ada!
score := m["english"] // 0 — tapi key "english" ADA, value-nya emang 0

// Cara bener — comma ok pattern
score, ok := m["math"]
if ok {
    fmt.Println("Math score:", score) // 90
}

score, ok := m["science"]
if !ok {
    fmt.Println("Science score belum ada")
}

// Idiom Go: cek langsung di if
if score, ok := m["math"]; ok {
    fmt.Println("Ada, nilainya:", score)
} else {
    fmt.Println("Ga ada")
}
```

> **Gotcha:** Jangan pernah cek key existence pakai `if m["key"] == 0` — karena value-nya emang bisa 0. Selalu pakai comma ok pattern: `val, ok := m["key"]`.

### 4.4 Map dan Concurrency

```go
// Map di Go TIDAK thread-safe!
// Kalau diakses dari multiple goroutine → panic: concurrent map read and map write

// Solusi 1: sync.Mutex
var mu sync.Mutex
mu.Lock()
m["key"] = "value"
mu.Unlock()

// Solusi 2: sync.Map (built-in thread-safe map)
var sm sync.Map
sm.Store("key", "value")
val, ok := sm.Load("key")
```

---

## 5. Struct

### 5.1 Define Struct

**TypeScript (class):**
```typescript
class User {
  id: number;
  name: string;
  email: string;
  active: boolean;

  constructor(id: number, name: string, email: string) {
    this.id = id;
    this.name = name;
    this.email = email;
    this.active = true;
  }
}

const user = new User(1, "Irzan", "irzan@example.com");
```

**Go (struct — bukan class!):**
```go
type User struct {
    ID     int
    Name   string
    Email  string
    Active bool
}

// Bikin instance
user := User{
    ID:    1,
    Name:  "Irzan",
    Email: "irzan@example.com",
    Active: true,
}

// Atau tanpa field name (urutan harus sesuai, ga recommended)
user := User{1, "Irzan", "irzan@example.com", true}

// Akses field
fmt.Println(user.Name) // "Irzan"

// Update field
user.Name = "Irzan Aldi"
```

### 5.2 Struct Tags (untuk JSON, DB, Validation)

```go
type User struct {
    ID        int       `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" binding:"required"`
    Email     string    `json:"email" binding:"required,email"`
    Password  string    `json:"-"` // ga masuk JSON response
    CreatedAt time.Time `json:"created_at"`
}

// json:"-" artinya field ini di-skip waktu marshal ke JSON
// binding:"required" dipake Gin buat validasi input
```

**TypeScript equivalent:**
```typescript
class User {
  @Column()
  id: number;

  @Column()
  @IsNotEmpty()
  name: string;

  @Column()
  @IsEmail()
  email: string;

  @Exclude() // class-transformer
  password: string;
}
```

### 5.3 Nested Struct

**TypeScript:**
```typescript
class Address {
  street: string;
  city: string;
}

class User {
  name: string;
  address: Address;
}
```

**Go:**
```go
type Address struct {
    Street string `json:"street"`
    City   string `json:"city"`
}

type User struct {
    Name    string  `json:"name"`
    Address Address `json:"address"` // nested struct
}

user := User{
    Name: "Irzan",
    Address: Address{
        Street: "Jl. Sudirman",
        City:   "Jakarta",
    },
}
fmt.Println(user.Address.City) // "Jakarta"
```

### 5.4 Struct Method (Value vs Pointer Receiver)

**TypeScript:**
```typescript
class User {
  name: string;
  age: number;

  greet(): string {
    return `Hello, ${this.name}!`;
  }

  setAge(newAge: number): void {
    this.age = newAge; // langsung mutate
  }
}
```

**Go:**
```go
type User struct {
    Name string
    Age  int
}

// Value receiver — method dapet COPY dari struct
// Ga bisa modify struct asli
func (u User) Greet() string {
    return fmt.Sprintf("Hello, %s!", u.Name)
}

// Pointer receiver — method dapet POINTER ke struct asli
// BISA modify struct asli
func (u *User) SetAge(newAge int) {
    u.Age = newAge // ini modify struct yang asli
}

// Pakai
user := User{Name: "Irzan", Age: 25}
fmt.Println(user.Greet()) // "Hello, Irzan!"

user.SetAge(26)
fmt.Println(user.Age) // 26 — berubah!
```

**Kapan pakai Value vs Pointer Receiver?**

| Situasi | Pakai | Alasan |
|---------|-------|--------|
| Method perlu modify struct | **Pointer `*T`** | Biar perubahan ke-apply |
| Struct besar (banyak field) | **Pointer `*T`** | Ga copy-copy data besar |
| Method cuma baca, struct kecil | **Value `T`** | Lebih aman, no side effect |
| Konsistensi | **Pointer `*T`** | Kalau 1 method pointer, semua pointer |

> **Gotcha:** Kalau kamu dari JS, semua object itu pass-by-reference. Di Go, struct itu **pass-by-value** (di-copy). Makanya ada pointer receiver. Ini sering bikin bingung di awal.

### 5.5 Constructor Pattern (NewXxx)

Go ga punya keyword `constructor` atau `new`. Tapi ada konvensi:

**TypeScript:**
```typescript
class UserService {
  private repo: UserRepository;

  constructor(repo: UserRepository) {
    this.repo = repo;
  }
}

const service = new UserService(repo);
```

**Go:**
```go
type UserService struct {
    repo UserRepository
}

// Konvensi: function NewXxx yang return pointer ke struct
func NewUserService(repo UserRepository) *UserService {
    return &UserService{
        repo: repo,
    }
}

// Pakai
service := NewUserService(repo)
```

Ini pattern yang bakal kamu lihat di **semua** project Go. `NewXxx()` = constructor.

---

## 6. Interface

### 6.1 Implicit Interface

Ini salah satu perbedaan paling signifikan antara Go dan TypeScript.

**TypeScript (explicit implements):**
```typescript
interface Logger {
  Log(msg: string): void;
}

// HARUS tulis "implements Logger"
class ConsoleLogger implements Logger {
  Log(msg: string): void {
    console.log(msg);
  }
}
```

**Go (implicit — ga perlu tulis "implements"):**
```go
type Logger interface {
    Log(msg string)
}

// Ga ada "implements Logger" — cukup punya method yang sama
type ConsoleLogger struct{}

func (c ConsoleLogger) Log(msg string) {
    fmt.Println(msg)
}

// ConsoleLogger otomatis satisfy interface Logger
// karena punya method Log(msg string)

// Pakai
var logger Logger = ConsoleLogger{} // works!
logger.Log("hello")
```

Ini namanya **duck typing**: "Kalau dia jalan kayak bebek dan bersuara kayak bebek, ya dia bebek."

### 6.2 Interface untuk Dependency Injection

Ini pattern paling penting di Go backend:

```go
// Define interface di consumer, bukan di provider
type TaskRepository interface {
    FindByID(id int) (*Task, error)
    Create(task *Task) error
    Update(task *Task) error
    Delete(id int) error
}

// Service depend on interface, bukan concrete type
type TaskService struct {
    repo TaskRepository // interface, bukan *PostgresTaskRepo
}

func NewTaskService(repo TaskRepository) *TaskService {
    return &TaskService{repo: repo}
}

// Implementation 1: PostgreSQL
type PostgresTaskRepo struct {
    db *gorm.DB
}

func (r *PostgresTaskRepo) FindByID(id int) (*Task, error) {
    var task Task
    err := r.db.First(&task, id).Error
    return &task, err
}
// ... implement semua method

// Implementation 2: Mock (untuk testing)
type MockTaskRepo struct {
    tasks map[int]*Task
}

func (r *MockTaskRepo) FindByID(id int) (*Task, error) {
    task, ok := r.tasks[id]
    if !ok {
        return nil, errors.New("not found")
    }
    return task, nil
}
// ... implement semua method

// Pake di production
service := NewTaskService(&PostgresTaskRepo{db: db})

// Pake di testing
service := NewTaskService(&MockTaskRepo{tasks: testData})
```

### 6.3 Empty Interface (any)

**TypeScript:**
```typescript
function printAnything(val: any): void {
  console.log(val);
}
```

**Go:**
```go
// Sebelum Go 1.18
func printAnything(val interface{}) {
    fmt.Println(val)
}

// Go 1.18+ (alias, lebih clean)
func printAnything(val any) {
    fmt.Println(val)
}

// `any` dan `interface{}` itu sama persis
// Tipe apapun satisfy interface kosong
printAnything(42)
printAnything("hello")
printAnything(User{Name: "Irzan"})
```

### 6.4 Type Assertion

**TypeScript:**
```typescript
function process(val: any) {
  const str = val as string;     // type assertion (unsafe)
  const str = <string>val;       // alternative syntax
}
```

**Go:**
```go
func process(val any) {
    // Type assertion (bisa panic kalau salah!)
    str := val.(string) // kalau val bukan string → PANIC

    // Safe type assertion — pakai comma ok pattern
    str, ok := val.(string)
    if ok {
        fmt.Println("Ini string:", str)
    } else {
        fmt.Println("Bukan string")
    }
}
```

### 6.5 Type Switch

**TypeScript:**
```typescript
function describe(val: any) {
  if (typeof val === "string") {
    console.log("String:", val);
  } else if (typeof val === "number") {
    console.log("Number:", val);
  }
}
```

**Go:**
```go
func describe(val any) {
    switch v := val.(type) {
    case string:
        fmt.Println("String:", v)
    case int:
        fmt.Println("Integer:", v)
    case bool:
        fmt.Println("Boolean:", v)
    case User:
        fmt.Println("User:", v.Name)
    default:
        fmt.Printf("Unknown type: %T\n", v)
    }
}
```

> **Gotcha:** Interface di Go itu **implicit**. Artinya kamu ga bakal liat error "Class X doesn't implement interface Y" saat compile, kecuali kamu coba assign ke variable interface type. Tips: tambahin compile-time check: `var _ Logger = (*ConsoleLogger)(nil)`.

---

## 7. Function

### 7.1 Basic Function

**JavaScript:**
```javascript
function add(a, b) {
  return a + b;
}

// Arrow function
const add = (a, b) => a + b;
```

**Go:**
```go
func add(a int, b int) int {
    return a + b
}

// Shorthand kalau tipe sama
func add(a, b int) int {
    return a + b
}
```

### 7.2 Multiple Return Values

Ini **fitur killer Go** yang bikin error handling jadi natural.

**JavaScript (workaround pakai object/array):**
```javascript
function divide(a, b) {
  if (b === 0) {
    return { result: 0, error: "division by zero" };
  }
  return { result: a / b, error: null };
}

const { result, error } = divide(10, 0);
```

**Go (native support):**
```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Pakai
result, err := divide(10, 3)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result) // 3.333...
```

### 7.3 Named Return Values

```go
// Named return — jarang dipake tapi perlu tau
func divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = errors.New("division by zero")
        return // "naked return" — return result dan err
    }
    result = a / b
    return
}
```

> **Tips:** Named return berguna untuk dokumentasi (pembaca tau return value-nya apa), tapi naked return bisa bikin confusing di function panjang. Pakai dengan bijak.

### 7.4 Variadic Function

**JavaScript:**
```javascript
function sum(...numbers) {
  return numbers.reduce((a, b) => a + b, 0);
}
sum(1, 2, 3, 4); // 10
```

**Go:**
```go
func sum(numbers ...int) int {
    total := 0
    for _, n := range numbers {
        total += n
    }
    return total
}

sum(1, 2, 3, 4) // 10

// Spread slice ke variadic
nums := []int{1, 2, 3, 4}
sum(nums...) // 10 — tiga titik di belakang, bukan di depan kayak JS
```

### 7.5 First-Class Function

**JavaScript:**
```javascript
const greet = (name) => `Hello, ${name}!`;

function apply(fn, value) {
  return fn(value);
}

apply(greet, "Irzan"); // "Hello, Irzan!"
```

**Go:**
```go
// Function sebagai variable
greet := func(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}

// Function sebagai parameter
func apply(fn func(string) string, value string) string {
    return fn(value)
}

apply(greet, "Irzan") // "Hello, Irzan!"

// Function type alias (biar ga ribet nulis func(...) berulang)
type Transformer func(string) string

func apply(fn Transformer, value string) string {
    return fn(value)
}
```

### 7.6 Closure

**JavaScript:**
```javascript
function makeCounter() {
  let count = 0;
  return () => {
    count++;
    return count;
  };
}

const counter = makeCounter();
counter(); // 1
counter(); // 2
counter(); // 3
```

**Go:**
```go
func makeCounter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

counter := makeCounter()
fmt.Println(counter()) // 1
fmt.Println(counter()) // 2
fmt.Println(counter()) // 3
```

**Real-world example — middleware di Gin:**
```go
// Closure dipake buat inject dependency ke middleware
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // authService "tertangkap" oleh closure
        token := c.GetHeader("Authorization")
        user, err := authService.ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        c.Set("user", user)
        c.Next()
    }
}
```

---

## 8. Error Handling

Ini **perbedaan paling fundamental** antara Go dan JS/TS. Di Go, error bukan exception — error itu **value biasa** yang di-return.

### 8.1 Basic Error Handling

**JavaScript (try/catch/throw):**
```javascript
function findUser(id) {
  const user = db.find(id);
  if (!user) {
    throw new Error("user not found");
  }
  return user;
}

try {
  const user = findUser(5);
  console.log(user);
} catch (err) {
  console.error(err.message);
}
```

**Go (return error):**
```go
func findUser(id int) (*User, error) {
    user, err := db.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("find user %d: %w", id, err)
    }
    if user == nil {
        return nil, errors.New("user not found")
    }
    return user, nil
}

// Pakai
user, err := findUser(5)
if err != nil {
    log.Printf("Error: %v", err)
    return
}
fmt.Println(user.Name)
```

### 8.2 Kenapa `if err != nil` dan Bukan try/catch?

1. **Explicit** — kamu selalu tau mana yang bisa error dan handle di tempat
2. **No hidden control flow** — `throw` bisa loncat ke `catch` yang entah di mana. `return err` jelas alurnya
3. **Ga bisa "lupa" handle** — kalau kamu ignore return value `err`, linter bakal ngomel
4. **Performance** — ga ada overhead exception stack trace

Ya, ini bikin kode lebih verbose. Tapi ini **design decision** — Go pilih explicitness over convenience.

### 8.3 Error Wrapping (fmt.Errorf + %w)

```go
// Tanpa wrapping — info hilang
func getTask(id int) (*Task, error) {
    task, err := repo.FindByID(id)
    if err != nil {
        return nil, errors.New("task not found") // original error hilang!
    }
    return task, nil
}

// Dengan wrapping — chainable context
func getTask(id int) (*Task, error) {
    task, err := repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("get task %d: %w", id, err)
        // %w = wrap error. Original error masih bisa diakses
    }
    return task, nil
}

// Hasilnya: "get task 5: record not found"
// Error-nya kayak stack trace, tapi manual dan explicit
```

### 8.4 errors.Is dan errors.As

```go
import "errors"

// errors.Is — cek apakah error (atau wrapped error) sama dengan target
var ErrNotFound = errors.New("not found")

func getTask(id int) (*Task, error) {
    task, err := repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("get task: %w", ErrNotFound)
    }
    return task, nil
}

// Di handler
task, err := getTask(5)
if errors.Is(err, ErrNotFound) {
    c.JSON(404, gin.H{"error": "task not found"})
    return
}

// errors.As — extract specific error type
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// Cek dan extract
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Field %s: %s\n", valErr.Field, valErr.Message)
}
```

**Comparison table:**

| JS/TS | Go |
|-------|-----|
| `throw new Error("msg")` | `return errors.New("msg")` |
| `throw new CustomError()` | `return &CustomError{}` |
| `try { ... } catch(e) { ... }` | `if err != nil { ... }` |
| `instanceof CustomError` | `errors.As(err, &target)` |
| `err.message` | `err.Error()` |
| Error bubbling (auto) | Explicit `return err` |

### 8.5 Custom Error Type

```go
// Custom error harus implement interface error (punya method Error() string)
type AppError struct {
    Code    int
    Message string
    Err     error // original error
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap biar support errors.Is / errors.As
func (e *AppError) Unwrap() error {
    return e.Err
}

// Helper function
func NewNotFound(msg string) *AppError {
    return &AppError{Code: 404, Message: msg}
}

func NewBadRequest(msg string) *AppError {
    return &AppError{Code: 400, Message: msg}
}

// Pakai
func (s *TaskService) GetByID(id int) (*Task, error) {
    task, err := s.repo.FindByID(id)
    if err != nil {
        return nil, &AppError{
            Code:    404,
            Message: "task not found",
            Err:     err,
        }
    }
    return task, nil
}
```

> **Gotcha:** Di JavaScript, kamu bisa `throw` di mana aja dan `catch` di level atas. Di Go, kalau kamu ga handle `err`, error-nya **hilang** (silently ignored). Linter kayak `errcheck` bisa bantu detect ini. Biasain: setiap kali fungsi return `error`, **selalu handle**.

---

## 9. Pointer

JavaScript ga punya pointer. Ini konsep yang perlu dipahami dari nol.

### 9.1 Apa itu Pointer?

Pointer = variable yang **isinya alamat memori** dari variable lain.

```
Analogi sederhana:

Variable biasa = kamu pegang buku fisiknya langsung
Pointer        = kamu pegang catatan "buku ada di rak 3, baris 2"

Kalau kamu fotocopy buku (pass by value), perubahan di copy ga ngaruh ke aslinya.
Kalau kamu kasih catatan alamat (pass by pointer), siapapun bisa ubah buku aslinya.
```

### 9.2 & (Address-of) dan * (Dereference)

```go
name := "Irzan"          // variable biasa, value = "Irzan"
ptr := &name              // & = ambil alamat. ptr isinya alamat memori name
fmt.Println(ptr)          // 0xc0000b2000 (alamat memori, beda tiap run)
fmt.Println(*ptr)         // "Irzan" — * = dereference, ambil nilai di alamat itu

*ptr = "Budi"             // ubah nilai di alamat itu
fmt.Println(name)         // "Budi" — name juga berubah, karena satu alamat!
```

**Diagram:**

```
Tanpa pointer:                   Dengan pointer:

┌──────────┐                    ┌──────────┐
│ name     │                    │ name     │
│ "Irzan"  │                    │ "Irzan"  │ ← nilai asli
└──────────┘                    └──────┬───┘
                                      │ alamat: 0xc000
┌──────────┐                    ┌─────┴────┐
│ copy     │                    │ ptr      │
│ "Irzan"  │ ← copy terpisah   │ 0xc000   │ ← nunjuk ke name
└──────────┘                    └──────────┘
```

### 9.3 Pass by Value vs Pass by Pointer

**JavaScript (object selalu by reference):**
```javascript
function changeName(user) {
  user.name = "Budi"; // langsung mutate object asli!
}

const user = { name: "Irzan" };
changeName(user);
console.log(user.name); // "Budi" — berubah!
```

**Go (struct = by value, harus explicit pakai pointer):**
```go
// Pass by value — dapat COPY, ga bisa ubah aslinya
func changeName(user User) {
    user.Name = "Budi" // cuma ubah copy, aslinya ga berubah
}

user := User{Name: "Irzan"}
changeName(user)
fmt.Println(user.Name) // "Irzan" — ga berubah!

// Pass by pointer — dapat alamat, BISA ubah aslinya
func changeName(user *User) {
    user.Name = "Budi" // ubah aslinya lewat pointer
}

user := User{Name: "Irzan"}
changeName(&user) // kirim alamat pakai &
fmt.Println(user.Name) // "Budi" — berubah!
```

### 9.4 Pointer di Real Code

```go
// Pattern paling common: NewXxx return pointer
func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
    // & = ambil alamat dari struct baru
    // Return pointer biar ga copy-copy struct besar
}

// Function yang mungkin return nil (kayak null di JS)
func findUser(id int) (*User, error) {
    // Return pointer ke User
    // Kalau ga ketemu, return nil
    if id == 0 {
        return nil, errors.New("invalid id")
    }
    user := User{ID: id, Name: "Irzan"}
    return &user, nil
}

// Pakai
user, err := findUser(5)
if err != nil {
    log.Fatal(err)
}
// user itu *User (pointer), tapi akses field-nya sama aja:
fmt.Println(user.Name) // Go otomatis dereference — ga perlu (*user).Name
```

### 9.5 Nil Pointer

```go
var ptr *User // nil pointer — belum nunjuk ke mana-mana

// PANIC! Nil pointer dereference
// fmt.Println(ptr.Name) // runtime error: invalid memory address

// Selalu cek nil dulu
if ptr != nil {
    fmt.Println(ptr.Name)
}
```

**Kapan pointer bisa nil?**
- Variable pointer yang belum di-assign
- Return value fungsi yang return `nil`
- Field struct yang bertipe pointer dan ga diisi

> **Gotcha:** "Nil pointer dereference" itu panic paling common di Go — mirip `Cannot read property 'x' of null` di JavaScript. Selalu cek nil sebelum akses pointer.

### 9.6 Kapan Pakai Pointer?

| Situasi | Pakai Pointer? | Alasan |
|---------|---------------|--------|
| Struct besar (banyak field) | Ya | Hemat memori, ga copy |
| Function perlu modify input | Ya | Biar perubahan ke-apply |
| Bisa return "nothing" (null) | Ya | Return `nil` |
| Struct kecil, cuma dibaca | Tidak | Value lebih aman, no side effect |
| Basic type (int, string) | Biasanya tidak | Copy murah |

---

## 10. Control Flow

### 10.1 If/Else

**JavaScript:**
```javascript
if (age >= 18) {
  console.log("dewasa");
} else if (age >= 13) {
  console.log("remaja");
} else {
  console.log("anak-anak");
}
```

**Go:**
```go
if age >= 18 {
    fmt.Println("dewasa")
} else if age >= 13 {
    fmt.Println("remaja")
} else {
    fmt.Println("anak-anak")
}
// Perhatikan: TANPA kurung () di kondisi. Kurung kurawal {} WAJIB.
```

**If dengan init statement (idiom Go):**
```go
// Deklarasi variable di dalam if — scope-nya cuma di block itu
if user, err := findUser(5); err != nil {
    fmt.Println("Error:", err)
} else {
    fmt.Println("User:", user.Name)
}
// user dan err ga bisa diakses di sini
```

> **Gotcha:** Di Go, kurung kurawal `{` **harus** di baris yang sama dengan `if`. Beda baris = compile error. Ini karena Go auto-insert semicolon.
>
> ```go
> // COMPILE ERROR!
> if age >= 18
> {
>     fmt.Println("dewasa")
> }
>
> // BENAR
> if age >= 18 {
>     fmt.Println("dewasa")
> }
> ```

### 10.2 Switch

**JavaScript:**
```javascript
switch (status) {
  case "active":
    console.log("User aktif");
    break; // harus break!
  case "inactive":
    console.log("User nonaktif");
    break;
  default:
    console.log("Status unknown");
}
```

**Go:**
```go
switch status {
case "active":
    fmt.Println("User aktif")
    // ga perlu break! Go otomatis break di setiap case
case "inactive":
    fmt.Println("User nonaktif")
case "pending", "review": // multiple values di satu case
    fmt.Println("Masih diproses")
default:
    fmt.Println("Status unknown")
}
```

**Switch tanpa expression (pengganti if-else chain):**
```go
switch {
case score >= 90:
    grade = "A"
case score >= 80:
    grade = "B"
case score >= 70:
    grade = "C"
default:
    grade = "F"
}
```

**Fallthrough (jarang dipake):**
```go
switch n {
case 1:
    fmt.Println("satu")
    fallthrough // lanjut ke case berikutnya (kebalikan dari JS yang perlu break)
case 2:
    fmt.Println("dua")
}
// Input 1 → print "satu" DAN "dua"
```

> **Gotcha:** Di JS, lupa `break` bikin fall-through (bug). Di Go, **default-nya auto break**. Kalau mau fall-through, harus explicit tulis `fallthrough`. Terbalik dari JS!

### 10.3 For Loop (Satu-satunya Loop di Go)

Go cuma punya `for`. Ga ada `while`, ga ada `do-while`.

**JavaScript:**
```javascript
// Classic for
for (let i = 0; i < 10; i++) { ... }

// While
while (condition) { ... }

// For...of (iterate array)
for (const item of array) { ... }

// For...in (iterate object keys)
for (const key in object) { ... }

// forEach
array.forEach((item, index) => { ... });
```

**Go:**
```go
// Classic for (kayak JS)
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// While-style (for tanpa init dan post)
for condition {
    // sama kayak while
}

// Infinite loop
for {
    // sama kayak while(true)
    // break untuk keluar
}

// For range — iterate slice
fruits := []string{"apple", "banana", "cherry"}
for index, value := range fruits {
    fmt.Printf("%d: %s\n", index, value)
}

// Skip index pakai _
for _, value := range fruits {
    fmt.Println(value)
}

// Cuma index
for index := range fruits {
    fmt.Println(index)
}

// For range — iterate map
scores := map[string]int{"math": 90, "english": 85}
for key, value := range scores {
    fmt.Printf("%s: %d\n", key, value)
}

// For range — iterate string (per rune/character)
for i, ch := range "Hello" {
    fmt.Printf("Index %d: %c\n", i, ch)
}
```

### 10.4 Break, Continue, Label

```go
// Break — keluar dari loop
for i := 0; i < 10; i++ {
    if i == 5 {
        break
    }
    fmt.Println(i) // 0, 1, 2, 3, 4
}

// Continue — skip iterasi ini, lanjut ke berikutnya
for i := 0; i < 10; i++ {
    if i%2 == 0 {
        continue // skip genap
    }
    fmt.Println(i) // 1, 3, 5, 7, 9
}

// Label — break/continue dari nested loop (ga ada di JS)
outer:
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            if i == 1 && j == 1 {
                break outer // keluar dari loop LUAR, bukan cuma loop dalam
            }
            fmt.Printf("(%d, %d) ", i, j)
        }
    }
// Output: (0, 0) (0, 1) (0, 2) (1, 0)
```

> **Gotcha:** JS punya `for...in` dan `for...of` yang sering bikin bingung. Go cuma punya `for range` dan perilakunya konsisten untuk slice, map, string, dan channel.

---

## 11. Goroutine & Channel

Ini **superpower** Go. Concurrency di Go jauh lebih simple daripada async/await di JS.

### 11.1 Goroutine Basics

**JavaScript (async/await):**
```javascript
async function fetchUser() {
  const res = await fetch("/api/user");
  return res.json();
}

// Jalankan concurrent
const [user, tasks] = await Promise.all([
  fetchUser(),
  fetchTasks(),
]);
```

**Go (goroutine):**
```go
// Keyword `go` = spawn goroutine (lightweight thread)
go fetchUser()   // jalankan di background
go fetchTasks()  // jalankan di background

// Tapi... gimana cara dapet hasilnya? → Channel!
```

### 11.2 Channel (Unbuffered)

Channel = pipa untuk kirim data antar goroutine.

```go
// Bikin channel
ch := make(chan string) // channel yang kirim/terima string

// Goroutine 1: kirim data ke channel
go func() {
    result := "Hello from goroutine!"
    ch <- result  // kirim ke channel (blocking sampai ada yang terima)
}()

// Main goroutine: terima data dari channel
msg := <-ch  // terima dari channel (blocking sampai ada yang kirim)
fmt.Println(msg) // "Hello from goroutine!"
```

**Diagram:**
```
Goroutine 1                  Channel              Main Goroutine
┌──────────┐               ┌────────┐            ┌──────────┐
│ kirim →  │ ──ch <- data──│        │──data = <-ch│ ← terima │
└──────────┘               └────────┘            └──────────┘
```

### 11.3 Channel (Buffered)

```go
// Unbuffered: kirim BLOCKING sampai ada yang terima
ch := make(chan int)

// Buffered: bisa kirim tanpa blocking sampai buffer penuh
ch := make(chan int, 3) // buffer size 3

ch <- 1  // ga blocking (slot 1/3)
ch <- 2  // ga blocking (slot 2/3)
ch <- 3  // ga blocking (slot 3/3)
// ch <- 4  // BLOCKING! buffer penuh, tunggu ada yang terima

val := <-ch // terima: 1 (FIFO)
```

### 11.4 Real Example: Concurrent API Calls

**JavaScript:**
```javascript
async function getDashboardData() {
  const [user, tasks, notifications] = await Promise.all([
    fetchUser(),
    fetchTasks(),
    fetchNotifications(),
  ]);
  return { user, tasks, notifications };
}
```

**Go:**
```go
type DashboardData struct {
    User          *User
    Tasks         []Task
    Notifications []Notification
}

func getDashboardData(userID int) (*DashboardData, error) {
    var (
        user          *User
        tasks         []Task
        notifications []Notification
        userErr       error
        taskErr       error
        notifErr      error
    )

    // WaitGroup = "tunggu sampai semua goroutine selesai"
    var wg sync.WaitGroup
    wg.Add(3) // kita spawn 3 goroutine

    go func() {
        defer wg.Done() // signal: goroutine ini selesai
        user, userErr = fetchUser(userID)
    }()

    go func() {
        defer wg.Done()
        tasks, taskErr = fetchTasks(userID)
    }()

    go func() {
        defer wg.Done()
        notifications, notifErr = fetchNotifications(userID)
    }()

    wg.Wait() // blocking sampai semua Done()

    // Cek error
    if userErr != nil {
        return nil, fmt.Errorf("fetch user: %w", userErr)
    }
    if taskErr != nil {
        return nil, fmt.Errorf("fetch tasks: %w", taskErr)
    }
    if notifErr != nil {
        return nil, fmt.Errorf("fetch notifications: %w", notifErr)
    }

    return &DashboardData{
        User:          user,
        Tasks:         tasks,
        Notifications: notifications,
    }, nil
}
```

### 11.5 Select (Multiplexing Channel)

Select = `switch` untuk channel. Tunggu dari banyak channel sekaligus.

```go
// Kayak Promise.race() di JS

select {
case msg := <-ch1:
    fmt.Println("Dari ch1:", msg)
case msg := <-ch2:
    fmt.Println("Dari ch2:", msg)
case <-time.After(5 * time.Second):
    fmt.Println("Timeout! Ga ada yang kirim dalam 5 detik")
}
```

**Real example — graceful shutdown:**
```go
func main() {
    srv := &http.Server{Addr: ":8080"}

    go func() {
        if err := srv.ListenAndServe(); err != nil {
            log.Println("Server stopped:", err)
        }
    }()

    // Tunggu signal interrupt (Ctrl+C)
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit // blocking sampai dapat signal
    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    log.Println("Server exited cleanly")
}
```

### 11.6 sync.Mutex (Mutual Exclusion)

Kalau banyak goroutine akses data yang sama → data race.

**JavaScript (ga ada masalah ini karena single-threaded):**
```javascript
// JS single-threaded, jadi ga perlu mutex
let counter = 0;
// Even async, JS ga punya true parallelism
```

**Go (multi-threaded, perlu mutex):**
```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()         // kunci — goroutine lain tunggu
    c.count++
    c.mu.Unlock()       // buka kunci
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock() // defer unlock — dijamin unlock walau panic
    return c.count
}

// Pakai
counter := &SafeCounter{}
var wg sync.WaitGroup

for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        counter.Increment()
    }()
}

wg.Wait()
fmt.Println(counter.Value()) // 1000 — selalu tepat
```

**Comparison Table:**

| JavaScript | Go |
|-----------|-----|
| `async function` | `go func()` |
| `await promise` | `<-channel` |
| `Promise.all([...])` | `sync.WaitGroup` |
| `Promise.race([...])` | `select { ... }` |
| Single-threaded (event loop) | Multi-threaded (goroutine) |
| Ga perlu mutex | Perlu `sync.Mutex` untuk shared state |
| `setTimeout` | `time.After()` / `time.Sleep()` |

> **Gotcha:** Goroutine itu **sangat ringan** (2KB stack vs ~1MB thread OS). Kamu bisa spawn ratusan ribu goroutine. Tapi jangan lupa sinkronisasi — data race di Go itu silent bug yang susah di-debug. Pakai `go run -race .` untuk detect.

---

## 12. Package & Module

### 12.1 Go Module (go.mod)

**JavaScript (package.json):**
```json
{
  "name": "task-manager",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0"
  }
}
```

**Go (go.mod):**
```
module github.com/irzan-aldi/task-manager

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    gorm.io/gorm v1.25.5
)
```

| npm | Go |
|-----|-----|
| `npm init` | `go mod init module-name` |
| `npm install express` | `go get github.com/gin-gonic/gin` |
| `npm install` (dari lock file) | `go mod download` |
| `npm prune` / remove unused | `go mod tidy` |
| `package.json` | `go.mod` |
| `package-lock.json` | `go.sum` |
| `node_modules/` | `$GOPATH/pkg/mod/` (global cache) |

> **Tips:** `go mod tidy` itu command ajaib. Dia add dependency yang kurang DAN hapus yang ga kepake. Jalanin ini setiap habis tambah/hapus import.

### 12.2 Import

**JavaScript:**
```javascript
// ES Module
import express from 'express';
import { Router } from 'express';

// CommonJS
const express = require('express');
```

**Go:**
```go
import "fmt"                                   // standard library
import "github.com/gin-gonic/gin"              // third-party
import "github.com/irzan-aldi/task-manager/internal/handler" // local package

// Grouped import (convention)
import (
    // Standard library
    "fmt"
    "net/http"
    "time"

    // Third-party (kosongkan 1 baris sebagai separator)
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // Internal/local
    "github.com/irzan-aldi/task-manager/internal/handler"
    "github.com/irzan-aldi/task-manager/internal/service"
)
```

### 12.3 Exported vs Unexported (Huruf Besar!)

Ini rule paling unik di Go. Ga ada keyword `public`, `private`, `export`. Cuma **huruf besar vs huruf kecil**.

**JavaScript:**
```javascript
// Explicit export
export class UserService { ... }
export function createUser() { ... }

// Private (convention pakai underscore atau #)
class UserService {
  #password;         // truly private (ES2022)
  _internalData;     // convention private
}
```

**Go:**
```go
package service

// EXPORTED (huruf besar) — bisa diakses dari package lain
type UserService struct { ... }   // ✅ public
func NewUserService() { ... }    // ✅ public
func (s *UserService) GetByID() { ... } // ✅ public

// UNEXPORTED (huruf kecil) — cuma bisa diakses di package sendiri
type userRepository struct { ... } // ❌ private
func validateInput() { ... }      // ❌ private
func (s *UserService) hashPassword() { ... } // ❌ private
```

**Struct field juga kena aturan ini:**
```go
type User struct {
    ID       int    `json:"id"`       // Exported — bisa diakses dari luar
    Name     string `json:"name"`     // Exported
    password string                    // unexported — ga bisa diakses dari luar package
}

// Di package lain:
// user.ID = 1       // ✅ works
// user.Name = "Irzan" // ✅ works
// user.password = "x" // ❌ COMPILE ERROR: unexported field
```

> **Gotcha:** JSON marshal/unmarshal cuma bisa akses **Exported field** (huruf besar). Kalau field huruf kecil, ga bakal masuk JSON output. Ini sering bikin bingung di awal.

### 12.4 Internal Package

```
project/
├── internal/           ← package di sini CUMA bisa diakses oleh project ini
│   ├── handler/
│   ├── service/
│   └── repository/
├── pkg/                ← package di sini bisa diakses siapapun (public)
│   └── response/
└── cmd/
    └── api/main.go
```

Go **enforce** folder `internal/` di level compiler. Package luar ga bisa import apapun di dalamnya. Ini ga cuma convention — ini **language rule**.

### 12.5 init() Function

`init()` dipanggil otomatis SEBELUM `main()`. Bisa ada di file mana aja.

```go
package config

import "os"

var DatabaseURL string

func init() {
    // Dipanggil otomatis sebelum main()
    DatabaseURL = os.Getenv("DATABASE_URL")
    if DatabaseURL == "" {
        DatabaseURL = "postgres://localhost:5432/taskdb"
    }
}
```

**Urutan eksekusi:**
```
1. Import semua package
2. Jalankan init() di setiap package (urutan import)
3. Jalankan init() di package main
4. Jalankan main()
```

> **Tips:** Jangan terlalu banyak pakai `init()`. Ini bikin flow program susah dilacak. Untuk project real, lebih baik explicit initialization di `main()`.

---

## 13. Testing

### 13.1 Basics

**JavaScript (Jest):**
```javascript
// user.test.js
describe("UserService", () => {
  test("should create user", () => {
    const user = createUser("Irzan", "irzan@test.com");
    expect(user.name).toBe("Irzan");
    expect(user.email).toBe("irzan@test.com");
  });
});
```

**Go:**
```go
// user_test.go — HARUS diakhiri _test.go
package service

import "testing"

func TestCreateUser(t *testing.T) { // HARUS diawali Test + parameter *testing.T
    user := CreateUser("Irzan", "irzan@test.com")

    if user.Name != "Irzan" {
        t.Errorf("expected Name 'Irzan', got '%s'", user.Name)
    }
    if user.Email != "irzan@test.com" {
        t.Errorf("expected Email 'irzan@test.com', got '%s'", user.Email)
    }
}
```

**Jalankan test:**
```bash
go test ./...           # test semua package
go test ./internal/service/  # test package tertentu
go test -v ./...        # verbose — tampilkan nama test
go test -run TestCreateUser  # jalankan test tertentu
```

### 13.2 Table-Driven Tests (Best Practice Go)

Ini pattern testing paling populer di Go. Kayak `test.each` di Jest.

**Jest (parameterized):**
```javascript
test.each([
  [1, 1, 2],
  [2, 3, 5],
  [0, 0, 0],
])("add(%i, %i) = %i", (a, b, expected) => {
  expect(add(a, b)).toBe(expected);
});
```

**Go (table-driven):**
```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 1, 1, 2},
        {"with zero", 0, 5, 5},
        {"negative numbers", -1, -1, -2},
        {"mixed", -1, 3, 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### 13.3 Testify (Assertion Library)

Go standard library ga punya assertion kayak `expect().toBe()`. Testify nambahin itu.

```bash
go get github.com/stretchr/testify
```

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserService(t *testing.T) {
    // assert — test lanjut walau gagal
    assert.Equal(t, "Irzan", user.Name)
    assert.NotNil(t, user)
    assert.True(t, user.Active)
    assert.Len(t, users, 3)
    assert.Contains(t, "hello world", "world")
    assert.NoError(t, err)
    assert.Error(t, err)
    assert.Empty(t, slice)

    // require — test STOP kalau gagal (kayak assert tapi fatal)
    require.NoError(t, err) // kalau error, langsung stop test ini
    // kode di bawah ga jalan kalau require gagal
}
```

**Comparison Jest vs Go+Testify:**

| Jest | Go + Testify |
|------|-------------|
| `expect(x).toBe(y)` | `assert.Equal(t, y, x)` |
| `expect(x).not.toBe(y)` | `assert.NotEqual(t, y, x)` |
| `expect(x).toBeNull()` | `assert.Nil(t, x)` |
| `expect(x).toBeTruthy()` | `assert.True(t, x)` |
| `expect(x).toHaveLength(3)` | `assert.Len(t, x, 3)` |
| `expect(fn).toThrow()` | `assert.Panics(t, fn)` |
| `expect(x).toContain(y)` | `assert.Contains(t, x, y)` |

### 13.4 Benchmark

Go punya built-in benchmarking. Jest ga punya ini.

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ { // b.N ditentukan otomatis oleh Go
        Add(1, 2)
    }
}
```

```bash
go test -bench=. -benchmem
# Output:
# BenchmarkAdd-8    1000000000    0.25 ns/op    0 B/op    0 allocs/op
```

### 13.5 Test Coverage

```bash
go test -cover ./...                     # tampilkan persentase coverage
go test -coverprofile=coverage.out ./... # generate coverage file
go tool cover -html=coverage.out         # buka visualisasi di browser
```

| Jest | Go |
|------|-----|
| `jest --coverage` | `go test -cover ./...` |
| Coverage di `coverage/` folder | `go tool cover -html=coverage.out` |
| Istanbul/V8 reporter | Built-in HTML reporter |

> **Gotcha:** Nama file test HARUS diakhiri `_test.go`. Nama function test HARUS diawali `Test` (capital T) dan parameter pertama `*testing.T`. Ini ga cuma convention — Go ga bakal jalanin kalau salah format.

---

## 14. Common Patterns

### 14.1 Comma Ok Pattern

Udah kita bahas di beberapa section. Ini rangkumannya:

```go
// Map: cek key exists
value, ok := myMap["key"]

// Type assertion: cek tipe
str, ok := val.(string)

// Channel: cek channel masih open
msg, ok := <-ch
if !ok {
    fmt.Println("Channel closed")
}
```

Pattern ini **di mana-mana** di Go. Kalau liat `x, ok :=`, itu pasti comma ok pattern.

### 14.2 Defer

`defer` = "jalankan ini nanti, tepat sebelum function return." Perfect untuk cleanup.

**JavaScript:**
```javascript
async function readFile() {
  const file = await fs.open("data.txt");
  try {
    const data = await file.read();
    return data;
  } finally {
    await file.close(); // cleanup di finally
  }
}
```

**Go:**
```go
func readFile() ([]byte, error) {
    file, err := os.Open("data.txt")
    if err != nil {
        return nil, err
    }
    defer file.Close() // dijamin dipanggil sebelum function return

    data, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }
    return data, nil
    // file.Close() dipanggil otomatis di sini
}
```

**Defer itu LIFO (Last In, First Out):**
```go
func example() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
    // Output: 3, 2, 1
}
```

**Common use cases:**
```go
// Close file
defer file.Close()

// Close database connection
defer db.Close()

// Unlock mutex
mu.Lock()
defer mu.Unlock()

// Close HTTP response body
resp, err := http.Get(url)
if err != nil { return err }
defer resp.Body.Close()

// Measure execution time
func timer(name string) func() {
    start := time.Now()
    return func() {
        fmt.Printf("%s took %v\n", name, time.Since(start))
    }
}

defer timer("fetchData")() // parenthesis penting!
```

### 14.3 Panic & Recover

`panic` ≈ `throw` di JS, tapi **cuma dipake untuk error yang benar-benar fatal**.

```go
// PANIC — program crash (JANGAN pake untuk error handling biasa!)
func mustConnect(url string) *sql.DB {
    db, err := sql.Open("postgres", url)
    if err != nil {
        panic("cannot connect to database: " + err.Error())
    }
    return db
}

// RECOVER — tangkap panic (kayak catch)
func safeHandler() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from panic:", r)
        }
    }()

    panic("something went wrong") // panic terjadi di sini

    // kode di bawah ga pernah jalan
    fmt.Println("after panic")
}
```

**Kapan pakai panic?**
- Program ga bisa jalan tanpa resource (database connection gagal saat startup)
- Bug di kode (harusnya ga mungkin terjadi)
- **JANGAN** pakai buat error handling biasa — pakai `return error`

**Gin punya Recovery middleware yang auto-recover dari panic:**
```go
r := gin.Default() // ini udah include Recovery() middleware
// Kalau handler panic, Gin tangkep dan return 500 — server ga crash
```

### 14.4 Blank Identifier `_`

`_` = "Aku tau ada value di sini, tapi aku ga butuh."

```go
// Ignore index saat iterate
for _, value := range items {
    fmt.Println(value)
}

// Ignore return value
_, err := fmt.Println("hello") // ga butuh jumlah byte yang ditulis

// Ignore error (HATI-HATI — biasanya bad practice)
result, _ := strconv.Atoi("42") // ignore error — yakin ga bakal error

// Import package hanya untuk side effect (init function)
import _ "github.com/lib/pq" // import postgres driver, kita ga panggil langsung
```

> **Gotcha:** Di Go, kalau kamu deklarasi variable tapi ga dipake → **compile error**. Ini strict. `_` jadi solusinya. Di JS, unused variable cuma warning.

### 14.5 Type Embedding (Composition over Inheritance)

Go ga punya inheritance (`extends`). Sebagai gantinya: **embedding** (composition).

**TypeScript (inheritance):**
```typescript
class BaseModel {
  id: number;
  createdAt: Date;
  updatedAt: Date;
}

class User extends BaseModel {
  name: string;
  email: string;
}
// user.id ← dari BaseModel
```

**Go (embedding):**
```go
type BaseModel struct {
    ID        int       `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
    BaseModel        // embedded — semua field BaseModel "naik" ke User
    Name      string `json:"name"`
    Email     string `json:"email"`
}

user := User{
    BaseModel: BaseModel{
        ID:        1,
        CreatedAt: time.Now(),
    },
    Name:  "Irzan",
    Email: "irzan@example.com",
}

// Akses langsung — seolah-olah field User sendiri
fmt.Println(user.ID)        // 1 — dari BaseModel
fmt.Println(user.Name)      // "Irzan" — dari User
fmt.Println(user.CreatedAt) // 2026-04-29... — dari BaseModel
```

**Embed interface (powerful pattern):**
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Compose interface dari interface lain
type ReadWriter interface {
    Reader
    Writer
}
// ReadWriter punya method Read DAN Write
```

> **Tips:** Go lebih suka **composition over inheritance**. Daripada deep class hierarchy (BaseModel → User → AdminUser), Go prefer embedding dan interface composition. Lebih flat, lebih flexible.

---

## 15. Cheat Sheet: JavaScript/TypeScript → Go

Quick reference buat sehari-hari. Bookmark halaman ini.

### Variables & Types

| JavaScript/TypeScript | Go |
|----------------------|-----|
| `let x = 5` | `x := 5` |
| `const x = 5` | `const x = 5` |
| `let x: number` | `var x int` |
| `let x: string = ""` | `var x string` (zero value `""`) |
| `undefined` | Ga ada — pake zero value |
| `null` | `nil` (hanya pointer, slice, map, interface, channel) |
| `typeof x` | `fmt.Sprintf("%T", x)` atau `reflect.TypeOf(x)` |
| `Number("42")` | `strconv.Atoi("42")` |
| `String(42)` | `strconv.Itoa(42)` |
| Template literal `` `${x}` `` | `fmt.Sprintf("%v", x)` |

### Data Structures

| JavaScript/TypeScript | Go |
|----------------------|-----|
| `const arr = [1, 2, 3]` | `arr := []int{1, 2, 3}` |
| `arr.push(4)` | `arr = append(arr, 4)` |
| `arr.length` | `len(arr)` |
| `arr.slice(1, 3)` | `arr[1:3]` |
| `arr.includes(x)` | `slices.Contains(arr, x)` (Go 1.21+) |
| `arr.filter(fn)` | Manual `for` loop |
| `arr.map(fn)` | Manual `for` loop |
| `arr.forEach(fn)` | `for _, v := range arr` |
| `const obj = { k: v }` | `m := map[string]T{"k": v}` |
| `obj.key` atau `obj["key"]` | `m["key"]` |
| `delete obj.key` | `delete(m, "key")` |
| `"key" in obj` | `_, ok := m["key"]` |
| `Object.keys(obj)` | `for k := range m` (collect manually) |
| `new Map()` | `make(map[K]V)` |

### Functions & Control Flow

| JavaScript/TypeScript | Go |
|----------------------|-----|
| `function f(a: number): number` | `func f(a int) int` |
| `(a) => a * 2` | `func(a int) int { return a * 2 }` |
| `(...args)` spread | `args...` (setelah slice) |
| `try { } catch { }` | `if err != nil { }` |
| `throw new Error("msg")` | `return errors.New("msg")` |
| `if (cond) { }` | `if cond { }` (tanpa parenthesis) |
| `switch/case + break` | `switch/case` (auto break) |
| `for (let i = 0; ...)` | `for i := 0; ...` |
| `for (const x of arr)` | `for _, x := range arr` |
| `while (cond)` | `for cond { }` |

### OOP & Structure

| JavaScript/TypeScript | Go |
|----------------------|-----|
| `class User { }` | `type User struct { }` |
| `new User()` | `User{}` atau `&User{}` |
| `constructor()` | `func NewUser() *User` |
| `this.field` | `u.Field` (receiver) |
| `class A extends B` | Type embedding `struct { B; ... }` |
| `implements Interface` | Implicit (otomatis kalau method cocok) |
| `interface Logger { }` (TS) | `type Logger interface { }` |
| `private field` | `unexportedField` (huruf kecil) |
| `public field` | `ExportedField` (huruf besar) |
| `export function` | Huruf besar: `func PublicFunc()` |

### Async & Concurrency

| JavaScript/TypeScript | Go |
|----------------------|-----|
| `async/await` | Goroutine + channel |
| `Promise.all([...])` | `sync.WaitGroup` |
| `Promise.race([...])` | `select { }` |
| `setTimeout(fn, ms)` | `time.AfterFunc(d, fn)` |
| `setInterval(fn, ms)` | `ticker := time.NewTicker(d)` |
| Event loop (single thread) | Goroutine (multi-thread) |

### Ecosystem

| npm/Node.js | Go |
|-------------|-----|
| `npm init` | `go mod init` |
| `npm install pkg` | `go get pkg` |
| `npm run test` | `go test ./...` |
| `npm run build` | `go build` |
| `npm start` | `go run .` |
| `package.json` | `go.mod` |
| `package-lock.json` | `go.sum` |
| `node_modules/` | `$GOPATH/pkg/mod/` (global, shared) |
| `.env` + `dotenv` | `os.Getenv()` + `.env` loader |
| `jest` | `go test` + `testify` |
| `eslint` | `golangci-lint` |
| `prettier` | `gofmt` (built-in, wajib) |

---

## 16. Checklist — Self Assessment

Selesai baca bab ini, cek pemahaman kamu:

### Fundamentals
- [ ] Paham perbedaan `var`, `:=`, dan `const` di Go
- [ ] Paham zero values dan kenapa Go ga punya `undefined`
- [ ] Bisa convert antar tipe (int ↔ string ↔ float)
- [ ] Paham string formatting pakai `fmt.Sprintf`

### Data Structures
- [ ] Paham perbedaan array (fixed) vs slice (dynamic)
- [ ] Bisa pakai `append`, `len`, `cap`, slicing
- [ ] Bisa deklarasi, CRUD, dan iterate map
- [ ] Paham comma ok pattern di map

### Struct & Interface
- [ ] Bisa define struct dengan tags (json, binding)
- [ ] Paham value receiver vs pointer receiver
- [ ] Bisa bikin constructor `NewXxx()`
- [ ] Paham implicit interface — ga perlu `implements`
- [ ] Bisa pakai type assertion dan type switch

### Functions & Error Handling
- [ ] Bisa return multiple values
- [ ] Paham `if err != nil` pattern
- [ ] Bisa wrap error pakai `fmt.Errorf` + `%w`
- [ ] Bisa bikin custom error type

### Pointer
- [ ] Paham `&` (address-of) dan `*` (dereference)
- [ ] Paham kenapa Go butuh pointer (pass by value default)
- [ ] Bisa hindari nil pointer dereference

### Control Flow
- [ ] Paham `for` sebagai satu-satunya loop
- [ ] Bisa pakai `for range` untuk slice, map, string
- [ ] Paham switch auto-break di Go

### Concurrency
- [ ] Paham goroutine dan cara spawn pakai `go`
- [ ] Paham channel (unbuffered vs buffered)
- [ ] Bisa pakai `sync.WaitGroup` untuk tunggu goroutine
- [ ] Paham kapan butuh `sync.Mutex`

### Package & Testing
- [ ] Paham `go.mod` dan `go mod tidy`
- [ ] Paham exported (huruf besar) vs unexported (huruf kecil)
- [ ] Bisa tulis test dengan `_test.go`
- [ ] Bisa tulis table-driven test

### Patterns
- [ ] Paham `defer` dan use case-nya
- [ ] Paham `panic/recover` dan kapan (tidak) memakainya
- [ ] Paham blank identifier `_`
- [ ] Paham type embedding (composition over inheritance)

---

> **Next up:** Di Bab 5, kita langsung praktek semua syntax ini dengan bikin REST API pakai Gin Framework. Semua konsep di bab ini bakal dipake secara nyata di project Task Manager API.
