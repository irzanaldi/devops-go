# Bab 19: Performance & Load Testing — Dari Profiling Sampai Production-Ready

## Kenapa Bab Ini Penting?

Selamat. Kamu udah sampai di bab ke-19. Kalau kamu udah ngikutin dari Bab 1 sampai 18, berarti Task Manager API kamu udah:
- Jalan di cloud (AWS) dengan Docker container
- Ada CI/CD pipeline yang auto deploy
- Ada Nginx reverse proxy dengan HTTPS
- Ada monitoring dengan Prometheus + Grafana
- Ada Kubernetes orchestration
- Ada Terraform infrastructure as code
- Udah paham microservices architecture

**Tapi satu pertanyaan penting:** API kamu **secepat apa**? Bisa handle **berapa user**? Kalau besok tiba-tiba viral, **apa yang terjadi**?

```
Sebelum bab ini:
- API jalan, fitur lengkap, deploy lancar
- Tapi... ga tau response time P95 berapa
- Ga tau bisa handle berapa request/detik
- Ga tau ada memory leak atau ga
- Ga tau query mana yang lambat
- Ga tau bottleneck di mana
- Kalau ada masalah performance, "coba restart aja"

Setelah bab ini:
- Bisa profiling Go app (CPU, memory, goroutine) pakai pprof
- Bisa benchmarking function-level dengan testing.B
- Bisa load testing full API flow pakai k6
- Paham EXPLAIN ANALYZE buat query optimization
- Paham caching strategy yang proper
- Bisa optimize connection pooling, JSON, HTTP, memory
- Punya performance budget dan checklist
- Kalau ada masalah performance, tau PERSIS di mana bottleneck-nya
```

> **Analogi gampang:** Performance itu kayak **kesehatan tubuh**. Kamu bisa kelihatan sehat di luar (fitur jalan, ga error), tapi di dalam mungkin ada **kolesterol tinggi** (memory leak), **penyumbatan arteri** (slow query), atau **tekanan darah tinggi** (goroutine leak). Profiling itu medical check-up — kamu butuh data, bukan feeling. Load testing itu stress test jantung — kamu kasih beban tinggi dan lihat apa yang pecah duluan.

---

## 19.1 Kenapa Performance Penting

### Slow App = Lost Users = Lost Revenue

Ini bukan opini. Ini data:

```
Google (2006):
  100ms delay = 1% revenue drop
  Pada skala Google, 1% = miliaran dolar per tahun

Amazon (2012):
  100ms latency = 1% sales drop
  1 detik delay = $1.6 BILLION per tahun

Walmart (2012):
  1 detik improvement = 2% conversion increase
  100ms improvement = 1% incremental revenue

Akamai (2017):
  53% mobile users abandon page jika load > 3 detik
  2 detik delay = 103% bounce rate increase

Pinterest (2015):
  Reduce perceived wait time 40% = 15% increase signup
  SEO traffic naik 15%
```

### Analogi: Restoran

Bayangin kamu lagi lapar, masuk ke restoran:

```
Restoran A (Fast):
━━━━━━━━━━━━━━━━━
- Duduk → langsung dikasih menu (50ms)
- Pesan → "Siap, Pak!" (100ms)
- Makanan datang → 5 menit (reasonable wait)
- Happy customer, kasih review bintang 5
- Besok balik lagi, ajak temen

Restoran B (Slow):
━━━━━━━━━━━━━━━━━
- Duduk → nunggu 10 menit baru dikasih menu
- Pesan → "Tunggu ya, Pak..." (dan ilang)
- Makanan datang → 45 menit
- Bete, makan buru-buru, ga kasih review
- Ga pernah balik lagi

Restoran C (Crash Under Load):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Weekend, tiba-tiba rame
- Dapur ga bisa handle pesanan sebanyak itu
- Pesanan ketuker, makanan ga selesai-selesai
- Customer mulai marah-marah, banyak yang cabut
- Viral di Google Reviews: "JANGAN KE SINI"
```

API kamu itu restoran. User itu customer. Performance itu kualitas pelayanan.

### Performance as a Feature

```
"Performance is not an optimization. Performance is a FEATURE."
— Jeff Atwood (co-founder Stack Overflow)
```

Banyak developer nganggep performance itu hal yang diurus "nanti", setelah semua fitur selesai. **Salah besar.**

```
Typical Mindset (Wrong):
━━━━━━━━━━━━━━━━━━━━━━━
1. Build semua fitur
2. Deploy ke production
3. "Nanti kalau udah banyak user, kita optimize"
4. User datang, API lemot
5. User pergi
6. Ga ada user buat di-optimize
7. Startup mati

Correct Mindset:
━━━━━━━━━━━━━━━
1. Build fitur dengan performance awareness
2. Benchmark critical paths dari awal
3. Set performance budget
4. Load test sebelum deploy
5. Monitor performance di production
6. Optimize berdasarkan data, bukan feeling
```

Di NestJS/Laravel, kamu mungkin udah terbiasa: "ah, pake ORM aja, ga usah mikir query." Itu **ok buat mulai**, tapi di tahap tertentu kamu harus **mikir performance**. Go punya tools yang sangat powerful buat ini — jauh lebih baik daripada yang ada di PHP/Node.js ecosystem.

---

## 19.2 Performance Metrics — Apa yang Harus Diukur?

### The Golden Signals

Google SRE Book define **Four Golden Signals** yang harus selalu dimonitor:

```
┌─────────────────────────────────────────────────────────────┐
│                    FOUR GOLDEN SIGNALS                       │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │   LATENCY    │  │  TRAFFIC     │                        │
│  │              │  │              │                        │
│  │  Berapa lama │  │  Berapa      │                        │
│  │  request     │  │  request     │                        │
│  │  diproses?   │  │  per detik?  │                        │
│  │              │  │              │                        │
│  │  P50, P95,   │  │  RPS, RPM   │                        │
│  │  P99         │  │              │                        │
│  └──────────────┘  └──────────────┘                        │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │   ERRORS     │  │  SATURATION  │                        │
│  │              │  │              │                        │
│  │  Berapa %    │  │  Seberapa    │                        │
│  │  request     │  │  "penuh"     │                        │
│  │  yang gagal? │  │  system?     │                        │
│  │              │  │              │                        │
│  │  5xx rate,   │  │  CPU%, RAM%, │                        │
│  │  timeout     │  │  disk, conn  │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

### 1. Latency (Response Time)

Latency = waktu dari request masuk sampai response keluar.

**Jangan cuma pakai average!** Average itu bohong.

```
Contoh:
9 request: 10ms, 10ms, 10ms, 10ms, 10ms, 10ms, 10ms, 10ms, 10ms
1 request: 5000ms (timeout!)

Average = (9 × 10 + 5000) / 10 = 509ms

"Average 509ms? Hmm, agak lambat tapi masih ok."
PADAHAL 90% request cuma 10ms, dan 10% user kena 5 DETIK timeout!
```

**Pakai percentile (P50, P95, P99):**

```
P50 (Median):   50% request lebih cepat dari ini
                → Pengalaman "typical user"

P95:            95% request lebih cepat dari ini
                → Pengalaman "hampir semua user"
                → INI yang paling penting di-monitor

P99:            99% request lebih cepat dari ini
                → Pengalaman "worst case" (kecuali outlier extreme)

P99.9:          99.9% request lebih cepat dari ini
                → Untuk high-traffic services (Google scale)
```

Kenapa P95 paling penting?

```
Kalau P95 kamu 200ms, artinya:
- 95% user dapet response < 200ms ✅
- 5% user dapet response > 200ms (bisa 500ms, bisa 2s, bisa timeout)
- 5% dari 1 juta user = 50,000 user yang frustrated

Yang biasanya kena P99/P99.9: power users (yang paling sering pakai app kamu,
yang paling banyak data, yang paling penting buat revenue).
```

### 2. Throughput (RPS / Requests Per Second)

```
RPS = berapa request yang bisa di-handle per detik

Contoh capacity:
┌──────────────────┬───────────────┬───────────────┐
│ Skenario         │ Expected RPS  │ Kapasitas     │
├──────────────────┼───────────────┼───────────────┤
│ Blog pribadi     │ 1-10          │ 1 server OK   │
│ Startup app      │ 100-1,000     │ 2-3 server    │
│ E-commerce       │ 1,000-10,000  │ Load balanced  │
│ Social media     │ 10,000-100K   │ Distributed    │
│ Google/Meta      │ 100K-1M+      │ Global infra   │
└──────────────────┴───────────────┴───────────────┘
```

Task Manager API kita? Realistic target: **500-2,000 RPS** per instance.

### 3. Error Rate

```
Error Rate = jumlah error / total request × 100%

Target:
- Healthy: < 0.1% error rate
- Warning: 0.1% - 1%
- Critical: > 1%
- Outage: > 5%

HTTP Status Code yang dihitung "error":
- 5xx (500, 502, 503, 504) → Server error → SELALU dihitung
- 4xx → Tergantung:
  - 400 Bad Request → Client error, biasanya ga dihitung
  - 401 Unauthorized → Normal auth flow
  - 429 Too Many Requests → Rate limiting working as expected
  - 404 Not Found → Bisa normal, bisa bug
```

### 4. TTFB (Time To First Byte)

```
TTFB = waktu dari browser kirim request sampai byte pertama response diterima

Breakdown:
[DNS Lookup] → [TCP Connect] → [TLS Handshake] → [Server Processing] → [First Byte]
    ~20ms          ~30ms           ~50ms              ~100ms              = ~200ms TTFB

Yang bisa kamu kontrol: Server Processing time
Yang di luar kontrol: Network latency (DNS, TCP, TLS)
```

### 5. Apdex Score

Apdex = **Application Performance Index**. Cara simple buat nyatain "seberapa puas user dengan performance?"

```
Apdex = (Satisfied + Tolerating/2) / Total Samples

Dimana:
- Satisfied: response time < T (threshold, misal 200ms)
- Tolerating: response time antara T dan 4T (200ms - 800ms)
- Frustrated: response time > 4T (> 800ms)

Contoh:
100 requests dengan T = 200ms:
- 70 request < 200ms (Satisfied)
- 20 request 200ms-800ms (Tolerating)
- 10 request > 800ms (Frustrated)

Apdex = (70 + 20/2) / 100 = (70 + 10) / 100 = 0.80

Score interpretation:
- 1.00 - 0.94: Excellent
- 0.93 - 0.85: Good
- 0.84 - 0.70: Fair
- 0.69 - 0.50: Poor
- < 0.50: Unacceptable
```

### SLA / SLO / SLI

Tiga konsep yang sering ketuker:

```
SLI (Service Level Indicator):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
= Metric yang diukur
= "Apa yang kita ukur?"
Contoh:
- P95 latency dari GET /api/v1/tasks
- Error rate endpoint /api/v1/auth/login
- Uptime percentage

SLO (Service Level Objective):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
= Target internal yang mau dicapai
= "Target kita berapa?"
Contoh:
- P95 latency < 200ms
- Error rate < 0.1%
- Uptime > 99.9%

SLA (Service Level Agreement):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
= Kontrak dengan customer (ada konsekuensi kalau dilanggar)
= "Janji kita ke customer"
Contoh:
- Uptime 99.95% → kalau kurang, customer dapet credit
- P95 < 500ms → kalau lebih, ada penalty

Hubungannya:
SLI mengukur → SLO yang jadi target → SLA yang jadi kontrak
SLO biasanya lebih ketat dari SLA (buffer)

SLA: 99.9% uptime (8.76 jam downtime/tahun)
SLO: 99.95% uptime (4.38 jam downtime/tahun) ← lebih ketat!
SLI: Actual uptime bulan ini = 99.97% ← ngukur aktualnya
```

---

## 19.3 Profiling Go App — pprof

### Apa Itu Profiling?

Profiling = mengumpulkan data tentang **apa yang dilakukan program kamu saat runtime**. CPU profiling bilang function mana yang paling banyak pakai CPU time. Memory profiling bilang dimana memory dialokasikan. Goroutine profiling bilang ada berapa goroutine yang jalan dan apa yang mereka lakukan.

### Go pprof — Built-in Profiling Tool

Go punya profiling tool yang **built-in** dan **production-ready**. Ga perlu install apa-apa. Ga ada overhead signifikan kalau ga aktif. Ini salah satu hal yang bikin Go unggul dari Node.js dan PHP.

```
Go Profiling Types:
━━━━━━━━━━━━━━━━━━
┌──────────────┬─────────────────────────────────────────────┐
│ Profile Type │ Apa yang diukur                             │
├──────────────┼─────────────────────────────────────────────┤
│ CPU          │ Function mana yang pakai CPU paling banyak  │
│ Heap         │ Memory allocation yang masih aktif          │
│ Allocs       │ Semua memory allocation (termasuk yg freed) │
│ Goroutine    │ Stack trace semua goroutine yang running    │
│ Block        │ Dimana goroutine nge-block (waiting)        │
│ Mutex        │ Contention pada mutex (lock competition)    │
│ Threadcreate │ OS thread yang dibuat                       │
└──────────────┴─────────────────────────────────────────────┘
```

### Setup pprof di Task Manager API

Ada 2 cara pakai pprof:

**Cara 1: Import `net/http/pprof` (paling gampang)**

```go
// cmd/api/main.go — Tambahkan pprof endpoint
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof" // <-- Import ini aja, auto register ke DefaultServeMux
    "os"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/internal/config"
    "github.com/irzan-aldi/task-manager/internal/handler"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/internal/repository"
    "github.com/irzan-aldi/task-manager/internal/router"
    "github.com/irzan-aldi/task-manager/internal/service"
)

func main() {
    cfg := config.Load()
    db := config.ConnectDB(cfg)

    if err := db.AutoMigrate(&model.User{}, &model.Board{}, &model.Task{}); err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }

    redisClient := config.ConnectRedis(cfg)

    userRepo := repository.NewUserRepository(db)
    boardRepo := repository.NewBoardRepository(db)
    taskRepo := repository.NewTaskRepository(db)

    authService := service.NewAuthService(userRepo, cfg.JWTSecret)
    userService := service.NewUserService(userRepo)
    boardService := service.NewBoardService(boardRepo)
    taskService := service.NewTaskService(taskRepo, boardRepo)

    handlers := &router.Handlers{
        Auth:  handler.NewAuthHandler(authService),
        User:  handler.NewUserHandler(userService),
        Board: handler.NewBoardHandler(boardService),
        Task:  handler.NewTaskHandler(taskService),
    }

    if cfg.Env == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.Default()
    router.Setup(r, handlers, authService, redisClient)

    // === PPROF: Start pprof server on separate port ===
    // PENTING: Jangan expose ke public! Hanya internal/localhost
    go func() {
        pprofAddr := ":6060"
        log.Printf("pprof server starting on %s", pprofAddr)
        if err := http.ListenAndServe(pprofAddr, nil); err != nil {
            log.Printf("pprof server error: %v", err)
        }
    }()

    port := cfg.Port
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on port %s (env: %s)", port, cfg.Env)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
        os.Exit(1)
    }
}
```

**Cara 2: Gin pprof middleware (lebih terintegrasi)**

```go
// Pakai library github.com/gin-contrib/pprof
import "github.com/gin-contrib/pprof"

func main() {
    r := gin.Default()

    // Register pprof routes di /debug/pprof/
    // Hanya di development!
    if cfg.Env != "production" {
        pprof.Register(r)
    }

    router.Setup(r, handlers, authService, redisClient)
    r.Run(":8080")
}
```

**Rekomendasi: pakai Cara 1 (port terpisah)**. Kenapa? Karena di production kamu bisa expose port 8080 ke public, tapi port 6060 **hanya bisa diakses internal** (via SSH tunnel, VPN, atau Kubernetes port-forward). Lebih aman.

### Akses pprof

```bash
# Setelah app running, buka browser:
# http://localhost:6060/debug/pprof/

# Atau via command line:
# CPU profile (30 detik)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap (memory) profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Allocs profile
go tool pprof http://localhost:6060/debug/pprof/allocs

# Block profile
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex profile
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### pprof Security Warning

```
⚠️  JANGAN PERNAH expose pprof ke public internet!

pprof bisa:
- Melihat source code yang running
- Melihat goroutine stack (bisa ada sensitive data)
- Melihat memory contents
- Trigger CPU profiling (CPU overhead)
- Trigger full GC

Protection strategies:
1. Run pprof di port terpisah, firewall port itu
2. Hanya aktifkan di development environment
3. Kalau di production, akses via SSH tunnel:
   ssh -L 6060:localhost:6060 user@production-server
4. Di Kubernetes, pakai port-forward:
   kubectl port-forward pod/task-manager-xxx 6060:6060
```

---

## 19.4 CPU Profiling — Cari Function yang Lambat

### Cara CPU Profiling Bekerja

CPU profiling di Go bekerja dengan **sampling**:
1. Go runtime pause program setiap ~10ms (100x per detik)
2. Rekam stack trace saat itu
3. Akumulasi data: function mana yang paling sering muncul di sample
4. Function yang sering muncul = function yang paling banyak pakai CPU

```
Sampling Process:
━━━━━━━━━━━━━━━━━
Time: 0ms   10ms   20ms   30ms   40ms   50ms   60ms   70ms
      ↓     ↓      ↓      ↓      ↓      ↓      ↓      ↓
      A     A      B      A      C      A      A      B
      ↓     ↓      ↓      ↓      ↓      ↓      ↓      ↓
      main  main   main   main   main   main   main   main

Results:
- Function A: 5 samples (62.5%) → ini yang paling banyak pakai CPU
- Function B: 2 samples (25%)
- Function C: 1 sample (12.5%)
```

### Step-by-Step CPU Profiling Task Manager

**Step 1: Start app dan generate traffic**

```bash
# Terminal 1: Start app
cd /Users/irzan-aldi/Downloads/PROJECT/Devops\ Go
go run cmd/api/main.go

# Terminal 2: Generate some traffic (kita bakal bahas k6 nanti)
# Untuk sekarang, pakai simple loop
for i in $(seq 1 100); do
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@mail.com","password":"password123"}' > /dev/null
done
```

**Step 2: Capture CPU profile**

```bash
# Capture CPU profile selama 30 detik
# Sambil capture, generate traffic di terminal lain!
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

**Step 3: Analisis di interactive mode**

```bash
# Setelah capture selesai, kamu masuk interactive mode:
(pprof) top 20
Showing nodes accounting for 2.5s, 89.29% of 2.8s total
Dropped 42 nodes (cum <= 0.014s)
Showing top 20 nodes out of 78
      flat  flat%   sum%        cum   cum%
     0.52s 18.57% 18.57%      0.52s 18.57%  crypto/sha256.block
     0.31s 11.07% 29.64%      0.31s 11.07%  runtime.memmove
     0.28s 10.00% 39.64%      0.42s 15.00%  encoding/json.(*encodeState).marshal
     0.22s  7.86% 47.50%      0.22s  7.86%  syscall.syscall
     0.18s  6.43% 53.93%      0.82s 29.29%  golang.org/x/crypto/bcrypt.expensiveBlowfishSetup
     ...

# Penjelasan kolom:
# flat  = waktu yang dihabiskan DI function itu sendiri
# cum   = waktu yang dihabiskan DI function itu + semua function yang dipanggil
# flat% = persentase dari total time
# sum%  = kumulatif flat% dari atas

# Dari output di atas, bisa kita simpulkan:
# 1. bcrypt (password hashing) = 29.29% cum → expected, bcrypt memang berat
# 2. JSON encoding = 15% cum → bisa dioptimize!
# 3. memmove = 11.07% → banyak copy data, mungkin perlu optimize struct
```

**Step 4: Drill down ke specific function**

```bash
# Lihat detail satu function:
(pprof) list marshal
Total: 2.8s
     0.28s      0.42s (flat, cum) 15.00% of Total
         .          .     42:func (e *encodeState) marshal(v interface{}, opts encOpts) {
     0.08s      0.08s     43:    // reflection code here
         .      0.12s     44:    e.reflectValue(reflect.ValueOf(v), opts)
     0.20s      0.22s     45:    // encode fields
         .          .     46:}

# Atau lihat function di service kita:
(pprof) list TaskService
(pprof) list Login
```

**Step 5: Visual mode (RECOMMENDED)**

```bash
# Buka web UI — INI YANG PALING BERGUNA
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30

# Browser otomatis buka di http://localhost:8081
# Ada beberapa view:
# - Top: Daftar function berurutan berdasarkan CPU time
# - Graph: Call graph — siapa panggil siapa, berapa cost-nya
# - Flame Graph: Visualisasi stack — PALING BERGUNA
# - Source: Source code annotated dengan CPU time
# - Peek: Lihat callers + callees suatu function
```

### Membaca Flame Graph

```
Flame Graph (dari bawah ke atas):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

┌─────────────────────────────────────────────────────┐
│                     bcrypt.hash (29%)                │
├────────────────────────┬────────────────────────────┤
│    AuthService.Login   │   json.Marshal (15%)       │
│         (45%)          │                            │
├────────────────────────┴────────────────────────────┤
│              gin.HandleContext (80%)                  │
├─────────────────────────────────────────────────────┤
│                   main.main (100%)                   │
└─────────────────────────────────────────────────────┘

Cara baca:
1. Lebar bar = berapa banyak CPU time
2. Semakin lebar = semakin banyak pakai CPU
3. Stack dari bawah ke atas = call chain
4. Cari bar yang paling LEBAR di ATAS = bottleneck

Dari flame graph di atas:
- bcrypt.hash 29% → Expected (password hashing memang CPU-intensive)
- json.Marshal 15% → Bisa di-optimize! (pakai JSON library yang lebih cepat)
```

### Contoh Real: Finding a Hot Function

```go
// Misalnya profiling nunjukin function ini lambat:
func (s *TaskService) GetByBoardID(boardID uint, page, limit int) ([]model.Task, int64, error) {
    return s.repo.FindByBoardID(boardID, page, limit)
}

// Dan di repository:
func (r *TaskRepository) FindByBoardID(boardID uint, page, limit int) ([]model.Task, int64, error) {
    var tasks []model.Task
    var total int64

    query := r.db.Where("board_id = ?", boardID)
    query.Model(&model.Task{}).Count(&total)  // ← Query 1: COUNT

    offset := (page - 1) * limit
    err := query.Preload("Assignee").        // ← Query 2: SELECT tasks
        Offset(offset).Limit(limit).
        Order("created_at DESC").
        Find(&tasks).Error
    // Preload("Assignee")                   // ← Query 3: SELECT users WHERE id IN (...)

    return tasks, total, err
}

// Problem: 3 queries! Dan kalau ada 10 tasks, Preload bisa jadi N+1
```

**Fix:**

```go
// Optimized version:
func (r *TaskRepository) FindByBoardIDOptimized(boardID uint, page, limit int) ([]model.Task, int64, error) {
    var tasks []model.Task
    var total int64

    offset := (page - 1) * limit

    // Single query with JOIN instead of Preload
    err := r.db.
        Joins("LEFT JOIN users ON users.id = tasks.assignee_id").
        Where("tasks.board_id = ?", boardID).
        Select("tasks.*, users.name as assignee_name, users.email as assignee_email").
        Offset(offset).
        Limit(limit).
        Order("tasks.created_at DESC").
        Find(&tasks).Error
    if err != nil {
        return nil, 0, err
    }

    // Count in separate, efficient query
    r.db.Model(&model.Task{}).Where("board_id = ?", boardID).Count(&total)

    return tasks, total, err
}
```

### CPU Profiling dari Test / Benchmark

```bash
# Generate CPU profile dari benchmark (ga perlu running server)
go test -bench=BenchmarkGetTasks -cpuprofile=cpu.prof ./internal/service/

# Analisis
go tool pprof -http=:8081 cpu.prof

# Generate CPU profile dari test
go test -run=TestGetByBoardID -cpuprofile=cpu.prof ./internal/repository/
go tool pprof cpu.prof
```

---

## 19.5 Memory Profiling — Cari Memory Leak

### Heap vs Stack

Di Go, memory dialokasikan di 2 tempat:

```
Stack:
━━━━━━
- Cepat (pointer digeser aja)
- Otomatis di-cleanup saat function return
- Ukuran terbatas (~1MB per goroutine, bisa grow)
- Local variables yang ga "escape"

Heap:
━━━━━
- Lebih lambat (perlu GC / garbage collector)
- Dibersihkan oleh Garbage Collector
- Ukuran unlimited (sampai RAM habis)
- Variables yang "escape" dari function scope

Contoh:
func fastFunc() int {
    x := 42        // ← stack (ga escape, local variable)
    return x        // ← copy value ke caller
}

func slowFunc() *int {
    x := 42        // ← HEAP! (escape ke caller via pointer)
    return &x       // ← pointer ke x harus hidup setelah function return
}
```

### Memory Profile — Inuse vs Allocs

```
Ada 2 tipe memory profile:

Inuse (heap):
- Memory yang SEDANG dipakai saat ini
- Berguna buat: menemukan memory leak
- "Berapa memory yang lagi dipake sekarang?"

Allocs:
- SEMUA allocation sejak program start
- Berguna buat: menemukan function yang terlalu banyak alokasi
- "Function mana yang paling banyak bikin garbage?"

Analogi:
Inuse = berapa barang yang ada di rumah kamu SEKARANG
Allocs = berapa TOTAL barang yang pernah kamu beli (termasuk yang udah dibuang)
```

### Step-by-Step Memory Profiling

```bash
# Terminal 1: Start app
go run cmd/api/main.go

# Terminal 2: Generate traffic
for i in $(seq 1 1000); do
  curl -s http://localhost:8080/api/v1/boards \
    -H "Authorization: Bearer <token>" > /dev/null
done

# Terminal 3: Capture heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Interactive mode:
(pprof) top 10
Showing nodes accounting for 15.50MB, 92.81% of 16.70MB total
Dropped 25 nodes (cum <= 0.0835MB)
      flat  flat%   sum%        cum   cum%
    4.50MB 26.95% 26.95%     4.50MB 26.95%  encoding/json.(*decodeState).literalStore
    3.20MB 19.16% 46.11%     3.20MB 19.16%  reflect.mapassign
    2.10MB 12.57% 58.68%     6.80MB 40.72%  gorm.io/gorm.(*DB).Find
    1.80MB 10.78% 69.46%     1.80MB 10.78%  runtime.malg
    ...

# Allocs view:
(pprof) alloc_space
(pprof) top 10

# Visual:
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap
```

### Escape Analysis — Tau Mana yang Ke Heap

```bash
# Go compiler bisa bilang variable mana yang "escape" ke heap
go build -gcflags '-m' ./internal/service/ 2>&1 | head -30

# Output:
# ./service/task.go:30:9: &model.Task{...} escapes to heap
# ./service/task.go:44:2: moved to heap: task
# ./service/auth.go:41:9: &model.User{...} escapes to heap
# ./service/auth.go:109:9: &JWTClaims{...} escapes to heap

# Lebih verbose:
go build -gcflags '-m -m' ./internal/service/ 2>&1 | head -50
```

**Apa artinya?** Kalau variable "escapes to heap", berarti Go allocate di heap (butuh GC). Ini ga selalu bad, tapi kalau terlalu banyak di hot path, bisa jadi bottleneck.

### Common Memory Optimization Patterns

**Pattern 1: Reuse slice dengan sync.Pool**

```go
// BEFORE: Setiap request bikin slice baru
func (h *TaskHandler) GetByBoardID(c *gin.Context) {
    tasks, total, err := h.taskService.GetByBoardID(boardID, page, limit)
    // tasks = new []model.Task setiap request → heap allocation → GC pressure
    response.Success(c, http.StatusOK, "Board tasks", gin.H{
        "tasks": tasks,
        "total": total,
    })
}

// AFTER: Pakai sync.Pool buat reuse buffer
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func writeJSON(c *gin.Context, data interface{}) {
    buf := bufferPool.Get().(*bytes.Buffer)
    buf.Reset()
    defer bufferPool.Put(buf)

    encoder := json.NewEncoder(buf)
    if err := encoder.Encode(data); err != nil {
        c.JSON(500, gin.H{"error": "encoding failed"})
        return
    }

    c.Data(200, "application/json", buf.Bytes())
}
```

**Pattern 2: Pre-allocate slice**

```go
// BEFORE: Append grows slice dynamically → multiple reallocations
func processTaskIDs(tasks []model.Task) []uint {
    var ids []uint                    // len=0, cap=0
    for _, t := range tasks {
        ids = append(ids, t.ID)      // grow: 0→1→2→4→8→16→32→...
    }                                // banyak reallocation!
    return ids
}

// AFTER: Pre-allocate with known capacity
func processTaskIDs(tasks []model.Task) []uint {
    ids := make([]uint, 0, len(tasks))  // len=0, cap=len(tasks)
    for _, t := range tasks {
        ids = append(ids, t.ID)          // no reallocation!
    }
    return ids
}
```

**Pattern 3: Avoid unnecessary pointers**

```go
// BEFORE: Pointer causes heap allocation
func getDefaultPriority() *string {
    p := "medium"
    return &p        // escapes to heap!
}

// AFTER: Return value, not pointer (stays on stack)
func getDefaultPriority() string {
    return "medium"  // stack, no allocation
}
```

### Detecting Memory Leaks

Memory leak di Go biasanya bukan "lupa free memory" (GC handle itu), tapi **goroutine leak** atau **cache yang ga pernah di-evict**.

```go
// Memory leak pattern: unbounded cache
var taskCache = make(map[uint]*model.Task)

func cacheTask(task *model.Task) {
    taskCache[task.ID] = task  // NEVER DELETED! Grows forever!
}

// Fix: Pakai LRU cache dengan max size
import lru "github.com/hashicorp/golang-lru/v2"

var taskCache, _ = lru.New[uint, *model.Task](1000) // max 1000 entries

func cacheTask(task *model.Task) {
    taskCache.Add(task.ID, task) // Otomatis evict yang paling lama ga diakses
}
```

Cara detect memory leak di production:

```bash
# Capture heap profile at time T1
curl -o heap1.prof http://localhost:6060/debug/pprof/heap

# Tunggu beberapa menit/jam
sleep 300

# Capture heap profile at time T2
curl -o heap2.prof http://localhost:6060/debug/pprof/heap

# Compare! Lihat apa yang grow:
go tool pprof -base heap1.prof heap2.prof
(pprof) top
# Output menunjukkan DIFF: apa yang naik antara T1 dan T2
# Kalau ada function yang terus naik → itu sumber leak
```

---

## 19.6 Goroutine Profiling — Cari Goroutine Leak

### Apa Itu Goroutine Leak?

Goroutine leak = goroutine yang **ga pernah selesai**. Terus jalan selamanya, makan memory (~2-8KB per goroutine), dan ga bisa di-cleanup oleh GC.

```
Goroutine Leak Patterns:
━━━━━━━━━━━━━━━━━━━━━━━

Pattern 1: Channel yang ga pernah di-read
──────────────────────────────────────────
func leakyFunction() {
    ch := make(chan int)
    go func() {
        result := expensiveComputation()
        ch <- result  // ← BLOCK FOREVER kalau nobody reads ch!
    }()
    // Function return tanpa read dari ch
    // Goroutine stuck selamanya
}

Pattern 2: Channel yang ga pernah di-close
──────────────────────────────────────────
func leakyWorker() {
    ch := make(chan int)
    go func() {
        for val := range ch {  // ← BLOCK FOREVER kalau ch ga pernah di-close
            process(val)
        }
    }()
    // Lupa close(ch)
}

Pattern 3: Context ga di-cancel
────────────────────────────────
func leakyHTTPCall() {
    ctx := context.Background()  // ← no timeout, no cancel!
    resp, err := http.NewRequestWithContext(ctx, "GET", "https://slow-api.com", nil)
    // Kalau slow-api.com hang, goroutine ini hang selamanya
}

Pattern 4: Infinite loop tanpa exit condition
─────────────────────────────────────────────
func leakyPoller() {
    go func() {
        for {
            checkSomething()
            time.Sleep(time.Second)
            // Ga ada return, ga ada select case <-ctx.Done()
            // Goroutine ini jalan SELAMANYA
        }
    }()
}
```

### Detect Goroutine Leak

```bash
# Lihat semua goroutine yang running:
go tool pprof http://localhost:6060/debug/pprof/goroutine

(pprof) top 10
Showing nodes accounting for 156, 100% of 156 total
      flat  flat%   sum%        cum   cum%
       100 64.10% 64.10%        100 64.10%  runtime.gopark
        30 19.23% 83.33%         30 19.23%  main.leakyFunction.func1
        ...

# 100 goroutine parking = mungkin normal (waiting for I/O)
# 30 goroutine stuck di leakyFunction = LEAK!

# Lihat full stack trace:
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Output: full stack trace semua goroutine
# Cari yang suspicious — goroutine yang harusnya udah selesai tapi masih ada
```

### Monitoring goroutine count over time:

```go
// Tambahkan di main.go atau middleware
import "runtime"

// Goroutine count endpoint
r.GET("/debug/goroutines", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "goroutine_count": runtime.NumGoroutine(),
    })
})

// Atau expose sebagai Prometheus metric (di Bab 15 udah setup)
// Prometheus automatically collects go_goroutines metric
```

```bash
# Monitor goroutine count over time:
watch -n 5 'curl -s http://localhost:8080/debug/goroutines | jq .goroutine_count'

# Kalau angkanya terus NAIK tanpa turun → ada leak!
# Normal: naik saat traffic tinggi, turun saat traffic rendah
# Leak: naik terus menerus, ga pernah turun
```

### Fix Goroutine Leak Patterns

**Fix 1: Always use context with timeout**

```go
// BEFORE (leak-prone):
func callExternalAPI(url string) (*Response, error) {
    resp, err := http.Get(url)  // No timeout! Bisa hang forever
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    // ...
}

// AFTER (safe):
func callExternalAPI(ctx context.Context, url string) (*Response, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err  // Timeout → context canceled → goroutine cleaned up
    }
    defer resp.Body.Close()
    // ...
}
```

**Fix 2: Use errgroup for parallel work**

```go
import "golang.org/x/sync/errgroup"

// BEFORE (leak-prone):
func fetchBoardWithTasks(boardID uint) (*BoardDetail, error) {
    var board *model.Board
    var tasks []model.Task
    var taskCount int64

    // 3 goroutine, no coordination, no error handling
    go func() { board, _ = boardRepo.FindByID(boardID) }()
    go func() { tasks, _, _ = taskRepo.FindByBoardID(boardID, 1, 100) }()
    go func() { taskCount = countTasks(boardID) }()

    time.Sleep(time.Second) // "Wait" — terrible pattern!

    return &BoardDetail{Board: board, Tasks: tasks, Count: taskCount}, nil
}

// AFTER (safe):
func fetchBoardWithTasks(ctx context.Context, boardID uint) (*BoardDetail, error) {
    g, ctx := errgroup.WithContext(ctx)

    var board *model.Board
    var tasks []model.Task
    var taskCount int64

    g.Go(func() error {
        var err error
        board, err = boardRepo.FindByID(boardID)
        return err
    })

    g.Go(func() error {
        var err error
        tasks, _, err = taskRepo.FindByBoardID(boardID, 1, 100)
        return err
    })

    g.Go(func() error {
        var err error
        taskCount, err = countTasks(ctx, boardID)
        return err
    })

    // Wait for all goroutines to complete
    if err := g.Wait(); err != nil {
        return nil, err  // Kalau satu error, semua di-cancel via context
    }

    return &BoardDetail{Board: board, Tasks: tasks, Count: taskCount}, nil
}
```

**Fix 3: Always make goroutines cancellable**

```go
// Background worker yang bisa di-stop
func startTaskCleanupWorker(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(1 * time.Hour)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                log.Println("Task cleanup worker stopped")
                return  // ← Clean exit!
            case <-ticker.C:
                cleanupOldTasks()
            }
        }
    }()
}

// Di main.go:
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // ← Semua worker akan stop saat main exit

startTaskCleanupWorker(ctx)
startNotificationWorker(ctx)
```

---

## 19.7 Benchmarking in Go — testing.B

### Benchmark Basics

Di Go, benchmarking itu **built-in**. Ga perlu install library terpisah. File test yang sama, convention function beda.

```go
// internal/service/task_bench_test.go
package service

import (
    "testing"
)

// Naming convention: BenchmarkXxx(b *testing.B)
func BenchmarkProcessTaskTitle(b *testing.B) {
    title := "Complete the performance testing chapter"

    // b.N = jumlah iterasi, ditentukan otomatis oleh Go
    // Go akan run ulang dengan N yang makin besar sampai hasilnya stabil
    for i := 0; i < b.N; i++ {
        processTaskTitle(title)
    }
}

func processTaskTitle(title string) string {
    // Some processing
    return strings.ToUpper(strings.TrimSpace(title))
}
```

```bash
# Run benchmark
go test -bench=. ./internal/service/
# Output:
# BenchmarkProcessTaskTitle-8    12345678    87.3 ns/op

# Penjelasan:
# -8         = menggunakan 8 CPU cores (GOMAXPROCS)
# 12345678   = jumlah iterasi (b.N)
# 87.3 ns/op = 87.3 nanosecond per operation
```

### Benchmark Options

```bash
# Run specific benchmark
go test -bench=BenchmarkProcessTaskTitle ./internal/service/

# Run semua benchmark yang match pattern
go test -bench=BenchmarkTask ./internal/service/

# Run benchmark N kali (untuk consistency)
go test -bench=. -count=5 ./internal/service/

# Benchmark dengan memory allocation stats
go test -bench=. -benchmem ./internal/service/
# Output:
# BenchmarkProcessTaskTitle-8  12345678  87.3 ns/op  48 B/op  2 allocs/op
#                                                     ↑         ↑
#                                                  bytes per op  allocations per op

# Set waktu benchmark (default 1s)
go test -bench=. -benchtime=10s ./internal/service/

# Set specific iteration count
go test -bench=. -benchtime=1000x ./internal/service/

# CPU profile dari benchmark
go test -bench=. -cpuprofile=cpu.prof ./internal/service/

# Memory profile dari benchmark
go test -bench=. -memprofile=mem.prof ./internal/service/
```

### Benchmark Patterns untuk Task Manager

**Benchmark 1: JSON Marshaling**

```go
// internal/handler/bench_test.go
package handler

import (
    "encoding/json"
    "testing"
    "time"

    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/pkg/response"
)

func BenchmarkJSONMarshalTask(b *testing.B) {
    task := model.Task{
        ID:          1,
        Title:       "Complete performance testing",
        Description: "Write benchmarks for all critical paths",
        Status:      model.StatusInProgress,
        Priority:    model.PriorityHigh,
        BoardID:     1,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    b.ResetTimer() // Reset timer setelah setup
    for i := 0; i < b.N; i++ {
        json.Marshal(task)
    }
}

func BenchmarkJSONMarshalTaskList(b *testing.B) {
    tasks := make([]model.Task, 100) // 100 tasks
    for i := range tasks {
        tasks[i] = model.Task{
            ID:          uint(i + 1),
            Title:       "Task " + string(rune('A'+i%26)),
            Description: "Description for task",
            Status:      model.StatusTodo,
            Priority:    model.PriorityMedium,
            BoardID:     1,
            CreatedAt:   time.Now(),
            UpdatedAt:   time.Now(),
        }
    }

    resp := response.Response{
        Status:  200,
        Message: "Board tasks",
        Data:    tasks,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        json.Marshal(resp)
    }
}

func BenchmarkJSONMarshalResponse(b *testing.B) {
    resp := response.Response{
        Status:  200,
        Message: "Success",
        Data: map[string]interface{}{
            "id":    1,
            "title": "My Board",
        },
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        json.Marshal(resp)
    }
}
```

**Benchmark 2: Password Hashing**

```go
// internal/model/user_bench_test.go
package model

import "testing"

func BenchmarkHashPassword(b *testing.B) {
    for i := 0; i < b.N; i++ {
        user := &User{Password: "password123456"}
        user.HashPassword()
    }
}

func BenchmarkCheckPassword(b *testing.B) {
    user := &User{Password: "password123456"}
    user.HashPassword() // Hash once

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user.CheckPassword("password123456")
    }
}
```

**Benchmark 3: Sub-benchmarks**

```go
// internal/service/task_bench_test.go
func BenchmarkProcessTaskTitle(b *testing.B) {
    titles := map[string]string{
        "Short":    "Fix bug",
        "Medium":   "Implement user authentication with JWT tokens",
        "Long":     "Complete the entire performance testing chapter including profiling memory optimization and load testing with comprehensive examples and explanations",
        "Unicode":  "Selesaikan fitur pembayaran untuk toko online",
    }

    for name, title := range titles {
        b.Run(name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                processTaskTitle(title)
            }
        })
    }
}
```

```bash
# Output:
# BenchmarkProcessTaskTitle/Short-8     20000000   62.1 ns/op
# BenchmarkProcessTaskTitle/Medium-8    10000000   98.4 ns/op
# BenchmarkProcessTaskTitle/Long-8       5000000   187.2 ns/op
# BenchmarkProcessTaskTitle/Unicode-8   10000000   124.6 ns/op
```

### b.ReportAllocs() dan b.ResetTimer()

```go
func BenchmarkWithSetup(b *testing.B) {
    // Heavy setup — JANGAN dihitung dalam benchmark
    db := setupTestDB()
    defer db.Close()
    seedData(db, 1000)

    b.ResetTimer()  // ← Reset timer SETELAH setup selesai
    b.ReportAllocs() // ← Report memory allocations

    for i := 0; i < b.N; i++ {
        // Ini yang di-benchmark
        _, _, _ = taskRepo.FindByBoardID(1, 1, 10)
    }
}
```

### benchstat — Compare Benchmark Results

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run benchmark SEBELUM optimization, save ke file
go test -bench=BenchmarkJSONMarshal -benchmem -count=10 ./internal/handler/ > old.txt

# ... optimize code ...

# Run benchmark SETELAH optimization
go test -bench=BenchmarkJSONMarshal -benchmem -count=10 ./internal/handler/ > new.txt

# Compare!
benchstat old.txt new.txt

# Output:
# name               old time/op    new time/op    delta
# JSONMarshalTask-8    1.23µs ±2%     0.89µs ±1%   -27.64%  (p=0.000 n=10+10)
#
# name               old alloc/op   new alloc/op   delta
# JSONMarshalTask-8     384B ± 0%      256B ± 0%   -33.33%  (p=0.000 n=10+10)
#
# name               old allocs/op  new allocs/op  delta
# JSONMarshalTask-8     5.00 ± 0%      3.00 ± 0%   -40.00%  (p=0.000 n=10+10)

# Penjelasan:
# -27.64% time → 27% lebih cepat ✅
# -33.33% alloc → 33% less memory ✅
# -40.00% allocs → 40% fewer allocations ✅
# p=0.000 → statistically significant (bukan noise)
```

---

## 19.8 Load Testing dengan k6

### Apa Itu k6?

k6 itu **modern load testing tool** yang dibuat oleh Grafana Labs. Script-nya pakai **JavaScript** (ES6), tapi runtime-nya pakai **Go** — jadi performant banget.

```
Kenapa k6?
━━━━━━━━━━
✅ Script pakai JavaScript — kamu udah tau dari NestJS
✅ Runtime Go — ga ada Node.js overhead
✅ Built-in metrics — latency, RPS, error rate, dll
✅ Thresholds — pass/fail criteria
✅ Stages — ramp-up, steady state, ramp-down
✅ Checks — assertion di response
✅ Multiple protocols — HTTP, WebSocket, gRPC
✅ Integration — Grafana, Prometheus, InfluxDB
✅ CLI-based — cocok buat CI/CD
✅ Free & open source
```

### Install k6

```bash
# macOS
brew install k6

# Linux (Debian/Ubuntu)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D68
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Docker
docker run --rm -i grafana/k6 run - < script.js

# Verify
k6 version
```

### Basic k6 Script

```javascript
// loadtest/basic.js
import http from 'k6/http';
import { check, sleep } from 'k6';

// Configuration
export const options = {
  // 10 virtual users selama 30 detik
  vus: 10,
  duration: '30s',

  // Thresholds: kapan test dianggap FAIL
  thresholds: {
    http_req_duration: ['p(95)<200'],  // P95 harus < 200ms
    http_req_failed: ['rate<0.01'],    // Error rate < 1%
  },
};

// Main test function — dipanggil setiap VU setiap iterasi
export default function () {
  // Hit health endpoint
  const res = http.get('http://localhost:8080/health');

  // Checks — assertions
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 100ms': (r) => r.timings.duration < 100,
    'body contains status': (r) => r.body.includes('ok'),
  });

  // Simulate user think time
  sleep(1);
}
```

```bash
# Run
k6 run loadtest/basic.js

# Output:
#      ✓ status is 200
#      ✓ response time < 100ms
#      ✓ body contains status
#
#      checks.........................: 100.00% ✓ 900      ✗ 0
#      data_received..................: 180 kB  5.9 kB/s
#      data_sent......................: 81 kB   2.7 kB/s
#      http_req_blocked...............: avg=2.1ms   min=1µs    med=3µs    max=42ms   p(90)=5µs    p(95)=32ms
#      http_req_connecting............: avg=1.8ms   min=0s     med=0s     max=38ms   p(90)=0s     p(95)=28ms
#      http_req_duration..............: avg=5.2ms   min=1.1ms  med=3.8ms  max=42ms   p(90)=8.5ms  p(95)=12ms
#        { expected_response:true }...: avg=5.2ms   min=1.1ms  med=3.8ms  max=42ms   p(90)=8.5ms  p(95)=12ms
#      ✓ http_req_failed...............: 0.00%   ✓ 0         ✗ 300
#      http_req_receiving.............: avg=32µs    min=10µs   med=28µs   max=250µs  p(90)=52µs   p(95)=68µs
#      http_req_sending...............: avg=12µs    min=4µs    med=10µs   max=95µs   p(90)=18µs   p(95)=24µs
#      http_req_tls_handshaking.......: avg=0s      min=0s     med=0s     max=0s     p(90)=0s     p(95)=0s
#      http_req_waiting...............: avg=5.1ms   min=1.0ms  med=3.7ms  max=42ms   p(90)=8.4ms  p(95)=12ms
#      http_reqs......................: 300     9.97/s
#      iteration_duration.............: avg=1.005s  min=1.001s med=1.004s max=1.045s p(90)=1.009s p(95)=1.015s
#      iterations.....................: 300     9.97/s
#      vus............................: 10      min=10      max=10
#      vus_max........................: 10      min=10      max=10
```

### Full k6 Script — Task Manager API Flow

Ini script yang complete — register, login, create board, create task, get tasks. Real end-to-end flow.

```javascript
// loadtest/task-manager-full.js
import http from 'k6/http';
import { check, group, sleep, fail } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const createTaskDuration = new Trend('create_task_duration');
const getTasksDuration = new Trend('get_tasks_duration');
const taskCreated = new Counter('tasks_created');

// Configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 VUs
    { duration: '1m', target: 10 },    // Stay at 10 VUs for 1 minute
    { duration: '30s', target: 50 },   // Ramp up to 50 VUs
    { duration: '2m', target: 50 },    // Stay at 50 VUs for 2 minutes
    { duration: '30s', target: 100 },  // Ramp up to 100 VUs
    { duration: '1m', target: 100 },   // Stay at 100 VUs for 1 minute
    { duration: '30s', target: 0 },    // Ramp down to 0
  ],

  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],  // P95 < 500ms, P99 < 1s
    http_req_failed: ['rate<0.05'],                   // Error rate < 5%
    errors: ['rate<0.1'],                              // Custom error rate < 10%
    login_duration: ['p(95)<300'],                     // Login P95 < 300ms
    create_task_duration: ['p(95)<400'],               // Create task P95 < 400ms
    get_tasks_duration: ['p(95)<200'],                 // Get tasks P95 < 200ms
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Setup — runs once before the test
export function setup() {
  // Register a test user
  const uniqueId = Date.now();
  const registerPayload = JSON.stringify({
    name: `Load Test User ${uniqueId}`,
    email: `loadtest-${uniqueId}@test.com`,
    password: 'password123456',
  });

  const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, registerPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  if (registerRes.status !== 201) {
    console.log(`Register failed: ${registerRes.status} - ${registerRes.body}`);
  }

  return {
    email: `loadtest-${uniqueId}@test.com`,
    password: 'password123456',
  };
}

export default function (data) {
  let token = '';
  let boardId = 0;

  // === GROUP 1: Login ===
  group('01_Login', function () {
    const loginPayload = JSON.stringify({
      email: data.email,
      password: data.password,
    });

    const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'Login' },
    });

    loginDuration.add(loginRes.timings.duration);

    const loginCheck = check(loginRes, {
      'login status is 200': (r) => r.status === 200,
      'login has access_token': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.data && body.data.access_token;
        } catch (e) {
          return false;
        }
      },
    });

    if (!loginCheck) {
      errorRate.add(1);
      console.log(`Login failed: ${loginRes.status} - ${loginRes.body}`);
      return;
    }

    errorRate.add(0);
    const body = JSON.parse(loginRes.body);
    token = body.data.access_token;
  });

  if (!token) return;

  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  sleep(0.5); // Think time antara actions

  // === GROUP 2: Create Board ===
  group('02_CreateBoard', function () {
    const boardPayload = JSON.stringify({
      title: `Load Test Board ${Date.now()}-${__VU}`,
      description: 'Board created during load test',
    });

    const boardRes = http.post(`${BASE_URL}/api/v1/boards`, boardPayload, {
      headers: authHeaders,
      tags: { name: 'CreateBoard' },
    });

    const boardCheck = check(boardRes, {
      'create board status is 201': (r) => r.status === 201,
      'board has id': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.data && body.data.id;
        } catch (e) {
          return false;
        }
      },
    });

    if (boardCheck) {
      const body = JSON.parse(boardRes.body);
      boardId = body.data.id;
      errorRate.add(0);
    } else {
      errorRate.add(1);
      console.log(`Create board failed: ${boardRes.status}`);
    }
  });

  if (!boardId) return;

  sleep(0.3);

  // === GROUP 3: Create Tasks ===
  group('03_CreateTasks', function () {
    const priorities = ['low', 'medium', 'high'];

    for (let i = 0; i < 3; i++) {
      const taskPayload = JSON.stringify({
        title: `Task ${i + 1} - VU ${__VU} - ${Date.now()}`,
        description: `Task description for load testing iteration ${__ITER}`,
        priority: priorities[i % 3],
      });

      const taskRes = http.post(
        `${BASE_URL}/api/v1/boards/${boardId}/tasks`,
        taskPayload,
        {
          headers: authHeaders,
          tags: { name: 'CreateTask' },
        }
      );

      createTaskDuration.add(taskRes.timings.duration);

      const taskCheck = check(taskRes, {
        'create task status is 201': (r) => r.status === 201,
      });

      if (taskCheck) {
        taskCreated.add(1);
        errorRate.add(0);
      } else {
        errorRate.add(1);
      }

      sleep(0.1); // Small delay between task creation
    }
  });

  sleep(0.3);

  // === GROUP 4: Get Board Tasks ===
  group('04_GetBoardTasks', function () {
    const tasksRes = http.get(
      `${BASE_URL}/api/v1/boards/${boardId}/tasks?page=1&limit=10`,
      {
        headers: authHeaders,
        tags: { name: 'GetBoardTasks' },
      }
    );

    getTasksDuration.add(tasksRes.timings.duration);

    check(tasksRes, {
      'get tasks status is 200': (r) => r.status === 200,
      'response has tasks array': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.data && Array.isArray(body.data.tasks);
        } catch (e) {
          return false;
        }
      },
      'has correct number of tasks': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.data.tasks.length >= 1;
        } catch (e) {
          return false;
        }
      },
    });
  });

  sleep(0.3);

  // === GROUP 5: Get My Tasks ===
  group('05_GetMyTasks', function () {
    const myTasksRes = http.get(`${BASE_URL}/api/v1/tasks/my?page=1&limit=10`, {
      headers: authHeaders,
      tags: { name: 'GetMyTasks' },
    });

    check(myTasksRes, {
      'get my tasks status is 200': (r) => r.status === 200,
    });
  });

  sleep(0.5);

  // === GROUP 6: Get Boards ===
  group('06_GetMyBoards', function () {
    const boardsRes = http.get(`${BASE_URL}/api/v1/boards`, {
      headers: authHeaders,
      tags: { name: 'GetMyBoards' },
    });

    check(boardsRes, {
      'get boards status is 200': (r) => r.status === 200,
    });
  });

  sleep(1); // End of iteration think time
}

// Teardown — runs once after the test
export function teardown(data) {
  console.log(`Load test completed for user: ${data.email}`);
}
```

```bash
# Run the full load test
k6 run loadtest/task-manager-full.js

# Run dengan environment variable
k6 run --env BASE_URL=http://staging.example.com:8080 loadtest/task-manager-full.js

# Run dan output ke JSON
k6 run --out json=results.json loadtest/task-manager-full.js

# Run dan kirim ke InfluxDB (buat Grafana dashboard)
k6 run --out influxdb=http://localhost:8086/k6 loadtest/task-manager-full.js
```

### k6 dengan Stages (Ramp Up/Down Pattern)

```javascript
// loadtest/stages-example.js
export const options = {
  stages: [
    // Ramp-up: 0 → 50 VUs dalam 2 menit
    { duration: '2m', target: 50 },

    // Steady state: 50 VUs selama 5 menit
    { duration: '5m', target: 50 },

    // Ramp-down: 50 → 0 VUs dalam 1 menit
    { duration: '1m', target: 0 },
  ],
};

/*
  Visual:
        VUs
  50 ┤           ╭──────────────────╮
     │          ╱                    ╲
     │         ╱                      ╲
     │        ╱                        ╲
     │       ╱                          ╲
   0 ┤──────╱                            ╲──
     └──────┬──────────┬──────────────────┬────
            2m         7m                 8m
         ramp-up    steady state      ramp-down
*/
```

### k6 Thresholds — Pass/Fail Criteria

```javascript
export const options = {
  thresholds: {
    // HTTP request duration
    http_req_duration: [
      'p(95)<200',       // 95% request harus < 200ms
      'p(99)<500',       // 99% request harus < 500ms
      'avg<100',          // Average harus < 100ms
      'max<2000',         // Maximum harus < 2s
    ],

    // Error rate
    http_req_failed: [
      'rate<0.01',        // Less than 1% errors
    ],

    // Custom metrics
    login_duration: [
      'p(95)<500',        // Login P95 < 500ms (bcrypt is slow)
    ],

    // Checks pass rate
    checks: [
      'rate>0.95',        // 95% checks harus pass
    ],

    // Per-endpoint threshold (tagged)
    'http_req_duration{name:GetBoardTasks}': ['p(95)<200'],
    'http_req_duration{name:Login}': ['p(95)<500'],
    'http_req_duration{name:CreateTask}': ['p(95)<300'],
  },
};
```

---

## 19.9 Load Testing dengan wrk dan ab

### wrk — Quick HTTP Benchmark

wrk itu tool benchmarking yang super ringan dan cepat. Cocok buat quick benchmark satu endpoint.

```bash
# Install
brew install wrk    # macOS
apt install wrk     # Ubuntu

# Basic usage: 10 threads, 100 connections, 30 seconds
wrk -t10 -c100 -d30s http://localhost:8080/health

# Output:
# Running 30s test @ http://localhost:8080/health
#   10 threads and 100 connections
#   Thread Stats   Avg      Stdev     Max   +/- Stdev
#     Latency     3.42ms    1.28ms   42.15ms   89.23%
#     Req/Sec     2.95k   234.12     3.89k    72.33%
#   884521 requests in 30.00s, 142.23MB read
# Requests/sec:  29483.37
# Transfer/sec:      4.74MB
```

```bash
# wrk dengan custom headers (authenticated endpoint)
wrk -t4 -c50 -d30s \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/v1/boards

# wrk dengan POST body
wrk -t4 -c50 -d30s \
  -s post.lua \
  http://localhost:8080/api/v1/auth/login
```

```lua
-- post.lua (wrk Lua script untuk POST)
wrk.method = "POST"
wrk.body   = '{"email":"test@mail.com","password":"password123"}'
wrk.headers["Content-Type"] = "application/json"
```

### ab (Apache Benchmark) — Super Simple

```bash
# Install
# macOS: sudah ada by default
# Ubuntu: apt install apache2-utils

# 1000 total requests, 50 concurrent
ab -n 1000 -c 50 http://localhost:8080/health

# POST request
ab -n 1000 -c 50 \
  -p login.json \
  -T "application/json" \
  http://localhost:8080/api/v1/auth/login

# Output:
# Concurrency Level:      50
# Time taken for tests:   2.340 seconds
# Complete requests:      1000
# Failed requests:        0
# Requests per second:    427.35 [#/sec] (mean)
# Time per request:       117.002 [ms] (mean)
# Time per request:       2.340 [ms] (mean, across all concurrent requests)
#
# Percentage of the requests served within a certain time (ms)
#   50%     85
#   66%    102
#   75%    118
#   80%    125
#   90%    168
#   95%    220
#   98%    312
#   99%    445
#  100%    892 (longest request)
```

### Kapan Pakai Apa?

```
┌────────────┬─────────────────────────────────────────────────┐
│ Tool       │ Kapan Pakai                                     │
├────────────┼─────────────────────────────────────────────────┤
│ ab         │ Quick sanity check, 1 endpoint, simple GET/POST │
│            │ "API ini bisa handle berapa RPS?"                │
│            │                                                 │
│ wrk        │ High-performance benchmark, 1 endpoint          │
│            │ "Berapa MAX throughput endpoint ini?"            │
│            │ Lua scripting buat custom logic                  │
│            │                                                 │
│ k6         │ Full scenario testing, multi-endpoint flow      │
│            │ "Simulate 100 users doing register→login→CRUD"  │
│            │ Thresholds, checks, stages, CI/CD integration   │
│            │                                                 │
│ hey        │ Alternatif ab yang lebih modern (Go-based)      │
│            │ `go install github.com/rakyll/hey@latest`       │
└────────────┴─────────────────────────────────────────────────┘
```

---

## 19.10 Load Testing Scenarios — Smoke, Load, Stress, Spike, Soak

### 5 Tipe Load Test

```
┌──────────────────────────────────────────────────────────────────┐
│                    5 TYPES OF LOAD TESTS                         │
│                                                                  │
│  1. SMOKE TEST     → "API masih jalan ga?"                       │
│  2. LOAD TEST      → "Bisa handle traffic normal?"              │
│  3. STRESS TEST    → "Sampe mana batasnya?"                     │
│  4. SPIKE TEST     → "Gimana kalau tiba-tiba rame?"             │
│  5. SOAK TEST      → "Bisa jalan lama tanpa masalah?"           │
└──────────────────────────────────────────────────────────────────┘
```

### 1. Smoke Test — "Masih Jalan?"

```javascript
// loadtest/smoke.js
// Tujuan: Verify system bisa handle minimal load
// Durasi: 1-2 menit
// VUs: 1-5
export const options = {
  vus: 1,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(99)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

/*
  Visual:
  VUs
  1 ┤──────────────────────────
  0 ┤
    └─────────────────────────
    0                        1m
*/
```

### 2. Load Test — "Handle Traffic Normal?"

```javascript
// loadtest/load.js
// Tujuan: Test di expected traffic level
// Durasi: 5-10 menit
// VUs: Expected concurrent users
export const options = {
  stages: [
    { duration: '2m', target: 50 },    // Ramp up
    { duration: '5m', target: 50 },    // Steady at expected load
    { duration: '1m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

/*
  Visual:
  VUs
  50 ┤       ╭───────────────────╮
     │      ╱                     ╲
     │     ╱                       ╲
   0 ┤────╱                         ╲───
     └────┬──────┬──────────────────┬───
          2m     7m                 8m
*/
```

### 3. Stress Test — "Sampe Mana Batasnya?"

```javascript
// loadtest/stress.js
// Tujuan: Cari breaking point
// Durasi: 10-20 menit
// VUs: Naik terus sampai system mulai error
export const options = {
  stages: [
    { duration: '2m', target: 50 },    // Normal load
    { duration: '2m', target: 100 },   // Above normal
    { duration: '2m', target: 200 },   // High load
    { duration: '2m', target: 300 },   // Very high load
    { duration: '2m', target: 400 },   // Breaking point?
    { duration: '2m', target: 500 },   // Beyond breaking point
    { duration: '2m', target: 0 },     // Recovery
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],  // Lebih lenient — kita MAU lihat degradation
    http_req_failed: ['rate<0.5'],      // Allow higher error rate
  },
};

/*
  Visual:
  VUs
  500 ┤                           ╭──╮
  400 ┤                      ╭────╯  │
  300 ┤                 ╭────╯       │
  200 ┤            ╭────╯            │
  100 ┤       ╭────╯                 │
   50 ┤  ╭────╯                      ╲
    0 ┤──╯                            ╲──
      └──┬────┬────┬────┬────┬────┬────┬──
         2m   4m   6m   8m  10m  12m  14m

  Expected results:
  - 50-100 VUs: Normal, P95 < 200ms ✅
  - 200 VUs: Slight degradation, P95 ~300ms ⚠️
  - 300 VUs: Noticeable degradation, P95 ~800ms ⚠️
  - 400 VUs: Errors start appearing, P95 > 2s ❌
  - 500 VUs: Many errors, timeouts ❌
  - Recovery: System recovers back to normal ✅ (hopefully!)
*/
```

### 4. Spike Test — "Tiba-tiba Viral!"

```javascript
// loadtest/spike.js
// Tujuan: Simulate sudden traffic spike
// Real-world: Flash sale, viral tweet, breaking news
export const options = {
  stages: [
    { duration: '1m', target: 10 },    // Normal traffic
    { duration: '10s', target: 500 },  // SPIKE! 0 → 500 dalam 10 detik
    { duration: '2m', target: 500 },   // Stay at spike level
    { duration: '10s', target: 10 },   // Spike ends
    { duration: '2m', target: 10 },    // Recovery period
    { duration: '30s', target: 0 },    // Ramp down
  ],
};

/*
  Visual:
  VUs
  500 ┤         ╭─────────────────╮
      │         │                 │
      │         │                 │
      │         │                 │
   10 ┤─────────╯                 ╰───────────────
    0 ┤                                           ╲──
      └────┬────┬─────────────────┬───────────────┬──
           1m                     3m+10s          5m+50s

  Key questions:
  - Berapa lama system bisa handle spike?
  - Ada request yang dropped?
  - Recovery time berapa lama?
  - Ada data corruption ga?
*/
```

### 5. Soak Test — "Bisa Jalan Lama?"

```javascript
// loadtest/soak.js
// Tujuan: Detect problems that appear over time
// Memory leak, connection leak, DB connection exhaustion, disk full
// Durasi: 4-12 jam (biasanya overnight)
export const options = {
  stages: [
    { duration: '5m', target: 50 },     // Ramp up
    { duration: '4h', target: 50 },     // Sustained load selama 4 JAM
    { duration: '5m', target: 0 },      // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
  },
};

/*
  Visual:
  VUs
  50 ┤     ╭────────────────────────────────────────────╮
     │    ╱                                              ╲
     │   ╱                                                ╲
   0 ┤──╱                                                  ╲──
     └──┬──────────────────────────────────────────────────┬──
        5m                                                4h+10m

  What to watch:
  - Memory usage over time (should be flat, not climbing)
  - Response time over time (should be stable)
  - Error rate over time (should stay near 0%)
  - Goroutine count (should be stable)
  - DB connection count (should be stable)
  - Disk usage (logs growing?)
*/
```

### Run Semua Scenario di CI/CD

```yaml
# .github/workflows/load-test.yml
name: Load Tests

on:
  push:
    branches: [main]
  schedule:
    - cron: '0 2 * * 1'  # Soak test setiap Senin jam 2 pagi

jobs:
  smoke-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Start API
        run: docker compose up -d
      - name: Wait for API
        run: sleep 10
      - name: Run Smoke Test
        uses: grafana/k6-action@v0.3.1
        with:
          filename: loadtest/smoke.js

  load-test:
    runs-on: ubuntu-latest
    needs: smoke-test
    steps:
      - uses: actions/checkout@v4
      - name: Start API
        run: docker compose up -d
      - name: Wait for API
        run: sleep 10
      - name: Run Load Test
        uses: grafana/k6-action@v0.3.1
        with:
          filename: loadtest/load.js
```

---

## 19.11 Database Optimization

### The #1 Performance Killer: Slow Queries

Di hampir semua API, **database query** adalah bottleneck #1. Bukan CPU, bukan memory, bukan network — **database**.

```
Typical request breakdown:
━━━━━━━━━━━━━━━━━━━━━━━━
Total: 150ms
├── Network (client → server): 10ms
├── Gin routing + middleware: 2ms
├── JSON parsing: 1ms
├── Business logic: 3ms
├── DATABASE QUERY: 120ms  ← 80% of total time!
├── JSON encoding: 4ms
└── Network (server → client): 10ms

Database optimization = 80% dari performance improvement kamu.
```

### EXPLAIN ANALYZE — X-ray Query Kamu

```sql
-- SEBELUM optimize, SELALU jalanin EXPLAIN ANALYZE dulu

-- Contoh 1: Get tasks by board_id (endpoint GET /boards/:id/tasks)
EXPLAIN ANALYZE
SELECT * FROM tasks
WHERE board_id = 1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 10 OFFSET 0;

-- Output:
-- Limit  (cost=0.00..12.85 rows=10 width=312) (actual time=0.025..0.032 rows=10 loops=1)
--   ->  Seq Scan on tasks  (cost=0.00..25.00 rows=15 width=312) (actual time=0.023..0.029 rows=10 loops=1)
--         Filter: ((deleted_at IS NULL) AND (board_id = 1))
--         Rows Removed by Filter: 5
-- Planning Time: 0.085 ms
-- Execution Time: 0.052 ms

-- "Seq Scan" = SEQUENTIAL SCAN = baca SEMUA baris = LAMBAT di tabel besar!
```

```sql
-- Contoh 2: Query TANPA index — 100,000 rows
EXPLAIN ANALYZE
SELECT * FROM tasks
WHERE board_id = 42 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 10;

-- Output (SLOW):
-- Limit  (cost=2500.00..2500.05 rows=10 width=312) (actual time=45.232..45.238 rows=10 loops=1)
--   ->  Sort  (cost=2500.00..2512.50 rows=5000 width=312) (actual time=45.229..45.234 rows=10 loops=1)
--         Sort Key: created_at DESC
--         Sort Method: top-N heapsort  Memory: 30kB
--         ->  Seq Scan on tasks  (cost=0.00..2375.00 rows=5000 width=312)
--               (actual time=0.025..38.125 rows=5000 loops=1)
--               Filter: ((deleted_at IS NULL) AND (board_id = 42))
--               Rows Removed by Filter: 95000
-- Planning Time: 0.112 ms
-- Execution Time: 45.289 ms  ← 45ms! LAMBAT!
```

### Indexing Strategy

```sql
-- FIX: Buat index!

-- Index 1: board_id (most common filter)
CREATE INDEX idx_tasks_board_id ON tasks(board_id)
WHERE deleted_at IS NULL;  -- Partial index: hanya index row yang belum di-delete

-- Index 2: assignee_id (untuk GET /tasks/my)
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id)
WHERE deleted_at IS NULL;

-- Index 3: Composite index untuk query yang sering dipakai
-- (board_id + created_at untuk sorting)
CREATE INDEX idx_tasks_board_created ON tasks(board_id, created_at DESC)
WHERE deleted_at IS NULL;

-- Index 4: Status filter (kalau ada feature filter by status)
CREATE INDEX idx_tasks_status ON tasks(board_id, status)
WHERE deleted_at IS NULL;
```

```sql
-- SETELAH index, jalanin EXPLAIN ANALYZE lagi:
EXPLAIN ANALYZE
SELECT * FROM tasks
WHERE board_id = 42 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 10;

-- Output (FAST):
-- Limit  (cost=0.29..1.52 rows=10 width=312) (actual time=0.035..0.048 rows=10 loops=1)
--   ->  Index Scan using idx_tasks_board_created on tasks
--         (cost=0.29..61.53 rows=500 width=312) (actual time=0.033..0.044 rows=10 loops=1)
--         Index Cond: (board_id = 42)
-- Planning Time: 0.135 ms
-- Execution Time: 0.068 ms  ← 0.068ms! Dari 45ms ke 0.068ms = 661x LEBIH CEPAT!
```

### Index Types

```
┌─────────────────┬──────────────────────────────────────────────┐
│ Index Type      │ Kapan Pakai                                  │
├─────────────────┼──────────────────────────────────────────────┤
│ B-tree          │ Default. = < > BETWEEN, ORDER BY             │
│ (default)       │ Contoh: WHERE board_id = 1                   │
│                 │ Contoh: ORDER BY created_at DESC              │
│                 │                                              │
│ Hash            │ Hanya equality (=)                           │
│                 │ Slightly faster than B-tree for =            │
│                 │ Tapi ga bisa range query                     │
│                 │                                              │
│ GIN             │ Full-text search, JSONB, Array               │
│                 │ Contoh: WHERE tags @> '{"urgent"}'           │
│                 │ Contoh: WHERE to_tsvector(title) @@ query    │
│                 │                                              │
│ GiST            │ Geometric, range, nearest-neighbor           │
│                 │ Contoh: PostGIS location queries              │
│                 │                                              │
│ Partial Index   │ Index hanya subset rows                      │
│                 │ Contoh: WHERE deleted_at IS NULL              │
│                 │ Lebih kecil, lebih cepat                     │
│                 │                                              │
│ Covering Index  │ Index contains semua column yang dibutuhkan  │
│ (INCLUDE)       │ Query ga perlu ke table (index-only scan)    │
└─────────────────┴──────────────────────────────────────────────┘
```

### N+1 Query Problem

N+1 problem itu klasik. Kalau kamu dari Laravel (Eloquent) atau NestJS (TypeORM), ini pasti familiar:

```go
// N+1 PROBLEM dengan GORM:
func (r *TaskRepository) FindByBoardIDBad(boardID uint) ([]model.Task, error) {
    var tasks []model.Task
    r.db.Where("board_id = ?", boardID).Find(&tasks) // Query 1: SELECT * FROM tasks

    // Setiap task, GORM lazy-load assignee → 1 query per task
    // Kalau ada 100 tasks → 100 extra queries!
    // Total: 1 + 100 = 101 queries 💀
    for i := range tasks {
        r.db.First(&tasks[i].Assignee, tasks[i].AssigneeID) // Query 2,3,4,...101
    }
    return tasks, nil
}
```

**Fix: Preload (Eager Loading)**

```go
// FIX 1: GORM Preload (2 queries instead of N+1)
func (r *TaskRepository) FindByBoardIDGood(boardID uint, page, limit int) ([]model.Task, int64, error) {
    var tasks []model.Task
    var total int64

    offset := (page - 1) * limit

    // Query 1: SELECT * FROM tasks WHERE board_id = ? LIMIT 10
    // Query 2: SELECT * FROM users WHERE id IN (1,2,3,4,5) ← Preload batches!
    err := r.db.
        Where("board_id = ?", boardID).
        Preload("Assignee"). // ← Ini bikin GORM batch load assignees
        Offset(offset).
        Limit(limit).
        Order("created_at DESC").
        Find(&tasks).Error

    r.db.Model(&model.Task{}).Where("board_id = ?", boardID).Count(&total)

    return tasks, total, err
}

// FIX 2: Joins (1 query — fastest)
func (r *TaskRepository) FindByBoardIDFastest(boardID uint, page, limit int) ([]model.Task, int64, error) {
    var tasks []model.Task
    var total int64

    offset := (page - 1) * limit

    // Single query with JOIN
    err := r.db.
        Joins("Assignee"). // ← LEFT JOIN users ON ...
        Where("tasks.board_id = ?", boardID).
        Offset(offset).
        Limit(limit).
        Order("tasks.created_at DESC").
        Find(&tasks).Error

    r.db.Model(&model.Task{}).Where("board_id = ?", boardID).Count(&total)

    return tasks, total, err
}
```

### Slow Query Log

```sql
-- PostgreSQL: Enable slow query logging
-- postgresql.conf:
-- log_min_duration_statement = 100  -- Log queries yang > 100ms

-- Atau di runtime:
ALTER SYSTEM SET log_min_duration_statement = 100;
SELECT pg_reload_conf();

-- Cek slow queries:
-- tail -f /var/log/postgresql/postgresql-15-main.log | grep duration
```

```go
// GORM: Enable slow query logging
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             200 * time.Millisecond, // Log query > 200ms
            LogLevel:                  logger.Warn,
            IgnoreRecordNotFoundError: true,
            Colorful:                  true,
        },
    ),
})
```

### Connection Pooling — pgbouncer

```
Tanpa Connection Pooler:
━━━━━━━━━━━━━━━━━━━━━━━
App (100 goroutines) ──→ PostgreSQL (100 connections)
                         └→ Setiap connection = ~2-10MB RAM
                         └→ 100 connections = 200MB-1GB RAM!
                         └→ PostgreSQL max_connections default = 100

Dengan pgbouncer:
━━━━━━━━━━━━━━━━
App (100 goroutines) ──→ pgbouncer (100 client connections)
                              ↓
                         pgbouncer ──→ PostgreSQL (20 server connections)
                         └→ Reuse connections efficiently
                         └→ 80% fewer database connections
```

```ini
; /etc/pgbouncer/pgbouncer.ini
[databases]
task_manager = host=localhost port=5432 dbname=task_manager

[pgbouncer]
listen_port = 6432
listen_addr = 0.0.0.0
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction          ; Reuse connection after each transaction
default_pool_size = 20           ; 20 connections ke PostgreSQL
max_client_conn = 1000           ; Accept 1000 client connections
min_pool_size = 5                ; Minimum 5 connections always ready
reserve_pool_size = 5            ; Extra connections buat peak
server_idle_timeout = 600        ; Close idle server connections after 10 min
```

```go
// Di Go, connect ke pgbouncer instead of PostgreSQL directly:
dsn := "host=localhost port=6432 user=postgres password=secret dbname=task_manager sslmode=disable"
// Port 6432 = pgbouncer, bukan 5432 (PostgreSQL langsung)
```

---

## 19.12 Caching Strategy

### Cache Hierarchy

```
Request masuk → cek cache level by level:

┌──────────────────────────────────────────────────────┐
│  Level 1: In-Memory Cache (process-local)            │
│  ┌──────────────────────────────────────────┐        │
│  │  sync.Map / LRU cache                    │        │
│  │  Speed: ~10ns                            │        │
│  │  Size: limited by process RAM            │        │
│  │  Scope: single process only              │        │
│  │  Expiry: manual                          │        │
│  └──────────────┬───────────────────────────┘        │
│                 │ MISS                               │
│  Level 2: Redis Cache (shared)                       │
│  ┌──────────────▼───────────────────────────┐        │
│  │  Redis / Memcached                       │        │
│  │  Speed: ~0.5ms (network hop)             │        │
│  │  Size: dedicated Redis server RAM        │        │
│  │  Scope: shared across all instances      │        │
│  │  Expiry: TTL-based                       │        │
│  └──────────────┬───────────────────────────┘        │
│                 │ MISS                               │
│  Level 3: Database                                   │
│  ┌──────────────▼───────────────────────────┐        │
│  │  PostgreSQL                              │        │
│  │  Speed: ~5-50ms                          │        │
│  │  Size: disk-based (unlimited)            │        │
│  │  Scope: source of truth                  │        │
│  │  Always returns data (if exists)         │        │
│  └──────────────────────────────────────────┘        │
└──────────────────────────────────────────────────────┘
```

### Cache-Aside Pattern (Lazy Loading)

Ini pattern paling umum. Bab 8 udah bahas Redis, sekarang kita implementasi beneran:

```go
// internal/service/task_cached.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/internal/repository"
)

type CachedTaskService struct {
    repo      *repository.TaskRepository
    boardRepo *repository.BoardRepository
    redis     *redis.Client
    cacheTTL  time.Duration
}

func NewCachedTaskService(
    repo *repository.TaskRepository,
    boardRepo *repository.BoardRepository,
    redis *redis.Client,
) *CachedTaskService {
    return &CachedTaskService{
        repo:      repo,
        boardRepo: boardRepo,
        redis:     redis,
        cacheTTL:  5 * time.Minute,
    }
}

// Cache-aside pattern:
// 1. Cek cache dulu
// 2. Kalau ada → return dari cache (FAST)
// 3. Kalau ga ada → query DB → simpan ke cache → return
func (s *CachedTaskService) GetByBoardID(ctx context.Context, boardID uint, page, limit int) ([]model.Task, int64, error) {
    cacheKey := fmt.Sprintf("board:%d:tasks:p%d:l%d", boardID, page, limit)

    // Step 1: Try cache
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        // Cache HIT!
        var result struct {
            Tasks []model.Task `json:"tasks"`
            Total int64        `json:"total"`
        }
        if json.Unmarshal([]byte(cached), &result) == nil {
            return result.Tasks, result.Total, nil
        }
    }

    // Step 2: Cache MISS → query database
    tasks, total, err := s.repo.FindByBoardID(boardID, page, limit)
    if err != nil {
        return nil, 0, err
    }

    // Step 3: Store in cache
    cacheData, _ := json.Marshal(struct {
        Tasks []model.Task `json:"tasks"`
        Total int64        `json:"total"`
    }{tasks, total})

    s.redis.Set(ctx, cacheKey, cacheData, s.cacheTTL)

    return tasks, total, nil
}

// Invalidate cache saat data berubah (create, update, delete)
func (s *CachedTaskService) InvalidateBoardCache(ctx context.Context, boardID uint) {
    // Delete semua cache yang related ke board ini
    pattern := fmt.Sprintf("board:%d:tasks:*", boardID)

    // Scan dan delete (lebih aman daripada KEYS command)
    iter := s.redis.Scan(ctx, 0, pattern, 100).Iterator()
    for iter.Next(ctx) {
        s.redis.Del(ctx, iter.Val())
    }
}

// Create task + invalidate cache
func (s *CachedTaskService) Create(ctx context.Context, boardID uint, userID uint, req model.CreateTaskRequest) (*model.Task, error) {
    // ... business logic ...

    task := &model.Task{
        Title:       req.Title,
        Description: req.Description,
        Priority:    req.Priority,
        BoardID:     boardID,
        AssigneeID:  req.AssigneeID,
        DueDate:     req.DueDate,
        Status:      model.StatusTodo,
    }

    if err := s.repo.Create(task); err != nil {
        return nil, err
    }

    // Invalidate cache!
    s.InvalidateBoardCache(ctx, boardID)

    return task, nil
}
```

### TTL Strategy

```
TTL (Time To Live) Strategy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━

Data yang jarang berubah → TTL panjang:
- User profile: 1 jam
- Board metadata: 30 menit
- Static config: 24 jam

Data yang sering berubah → TTL pendek:
- Task list: 5 menit
- Dashboard counters: 1 menit
- Search results: 2 menit

Data yang real-time → Jangan cache:
- Auth tokens (udah punya expiry sendiri)
- Notification count
- Currently editing indicator

TTL tips:
- Terlalu pendek → cache ga berguna (banyak MISS)
- Terlalu panjang → data stale (user lihat data lama)
- Goldilocks: 5-15 menit buat most API responses
```

### Cache Invalidation Patterns

```
"There are only two hard things in Computer Science:
 cache invalidation and naming things."
— Phil Karlton

Pattern 1: TTL-based (simplest)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Set TTL saat cache write
- Cache otomatis expire
- Pro: Simple, ga mungkin stale forever
- Con: Data bisa stale sampai TTL habis

Pattern 2: Event-based invalidation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Create/Update/Delete → invalidate related cache
- Pro: Data selalu fresh
- Con: Complex, bisa lupa invalidate

Pattern 3: Write-through
━━━━━━━━━━━━━━━━━━━━━━━
- Write ke cache DAN database bersamaan
- Pro: Cache selalu fresh
- Con: Write jadi lebih lambat (2 writes)

Pattern 4: Write-behind (async)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Write ke cache dulu, async write ke database
- Pro: Write super fast
- Con: Data loss risk kalau cache crash sebelum write ke DB
```

### HTTP Caching Headers

```go
// middleware/cache.go
package middleware

import (
    "fmt"
    "crypto/md5"
    "github.com/gin-gonic/gin"
)

// ETag middleware — browser caching
func ETagMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // Hanya untuk GET responses yang success
        if c.Request.Method != "GET" || c.Writer.Status() >= 400 {
            return
        }

        // Generate ETag dari response body
        // Note: di production, generate dari data hash, bukan body
        body := c.Writer.Header().Get("X-Response-Body")
        if body != "" {
            etag := fmt.Sprintf(`"%x"`, md5.Sum([]byte(body)))
            c.Header("ETag", etag)

            // Check If-None-Match
            if c.GetHeader("If-None-Match") == etag {
                c.AbortWithStatus(304) // Not Modified
                return
            }
        }
    }
}

// Cache-Control header untuk static-ish endpoints
func CacheControl(maxAge int) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "GET" {
            c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
        } else {
            c.Header("Cache-Control", "no-store")
        }
        c.Next()
    }
}

// Usage di router:
// r.GET("/api/v1/boards", CacheControl(60), h.Board.GetMyBoards)  // Cache 60 detik
// r.POST("/api/v1/boards", h.Board.Create)                         // No cache for mutations
```

---

## 19.13 Connection Pooling

### Database Connection Pool

Di Bab 6, kita udah setup connection pool basic. Sekarang kita tuning beneran.

```go
// internal/config/database.go
func ConnectDB(cfg *Config) *gorm.DB {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalf("Failed to get database instance: %v", err)
    }

    // === CONNECTION POOL TUNING ===

    // MaxOpenConns: Jumlah MAXIMUM koneksi yang bisa dibuka ke database
    // Terlalu kecil → request antri, latency naik
    // Terlalu besar → database kewalahan, RAM habis
    // Rule of thumb: cores * 2 + effective_spindle_count
    // Untuk SSD: cores * 2 + 1 (biasanya 25-50 cukup)
    sqlDB.SetMaxOpenConns(25)

    // MaxIdleConns: Jumlah koneksi yang TETAP dibuka saat idle
    // Harus <= MaxOpenConns
    // Terlalu kecil → sering buka/tutup connection (overhead)
    // Terlalu besar → koneksi idle makan resource
    // Rule of thumb: MaxOpenConns / 2
    sqlDB.SetMaxIdleConns(10)

    // ConnMaxLifetime: Berapa lama koneksi boleh hidup
    // PostgreSQL/pgbouncer punya server-side timeout juga
    // Set ini lebih pendek dari server-side timeout
    sqlDB.SetConnMaxLifetime(30 * time.Minute)

    // ConnMaxIdleTime: Berapa lama koneksi idle sebelum di-close
    // Kalau traffic bursty, set lebih panjang (koneksi siap pakai)
    // Kalau traffic steady, set lebih pendek (hemat resource)
    sqlDB.SetConnMaxIdleTime(5 * time.Minute)

    return db
}
```

```
Connection Pool Visual:
━━━━━━━━━━━━━━━━━━━━━━

MaxOpenConns = 25, MaxIdleConns = 10

Idle state (no traffic):
┌──────────────────────────────────────────────┐
│  Pool: [C1][C2][C3][C4][C5][C6][C7][C8][C9][C10]  │  10 idle connections
│         ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓    ↓   │
│        idle idle idle idle idle idle idle idle idle idle│
└──────────────────────────────────────────────┘

Under load (20 concurrent requests):
┌──────────────────────────────────────────────┐
│  Pool: [C1][C2][C3]...[C20][ ][ ][ ][ ][ ]  │
│        busy busy busy  busy  ← 5 slots available │
│        20 active, 0 idle, 5 available             │
└──────────────────────────────────────────────┘

Over limit (30 concurrent requests, max=25):
┌──────────────────────────────────────────────┐
│  Pool: [C1][C2][C3]...[C25]                  │  25 active (ALL USED)
│  Queue: [R26][R27][R28][R29][R30]            │  5 requests WAITING!
│         ← These wait until a connection frees up │
└──────────────────────────────────────────────┘
```

### Redis Connection Pool

```go
// internal/config/redis.go
func ConnectRedis(cfg *Config) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
        Password: cfg.Redis.Password,
        DB:       0,

        // Connection pool settings
        PoolSize:     50,              // Max connections in pool
        MinIdleConns: 10,              // Keep 10 connections ready
        MaxIdleConns: 30,              // Max idle connections

        // Timeouts
        DialTimeout:  5 * time.Second, // Connection timeout
        ReadTimeout:  3 * time.Second, // Read timeout
        WriteTimeout: 3 * time.Second, // Write timeout
        PoolTimeout:  4 * time.Second, // Wait for pool connection

        // Connection lifetime
        ConnMaxLifetime: 30 * time.Minute,
        ConnMaxIdleTime: 5 * time.Minute,
    })

    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }

    return client
}
```

### HTTP Client Pool

```go
// Buat HTTP client yang reusable (JANGAN buat http.Client baru setiap request!)

// BEFORE (BAD):
func callExternalAPI() (*http.Response, error) {
    client := &http.Client{} // ← Bikin client baru setiap kali = ga reuse TCP connection!
    return client.Get("https://api.example.com/data")
}

// AFTER (GOOD):
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,              // Max total idle connections
        MaxIdleConnsPerHost: 10,               // Max idle connections per host
        MaxConnsPerHost:     50,               // Max total connections per host
        IdleConnTimeout:     90 * time.Second, // Idle connection timeout
        DisableCompression:  false,
        // TLS handshake timeout
        TLSHandshakeTimeout: 5 * time.Second,
    },
}

func callExternalAPI() (*http.Response, error) {
    return httpClient.Get("https://api.example.com/data")
}
```

---

## 19.14 Concurrency Optimization

### Parallel I/O — Fetch Multiple Resources

Salah satu kekuatan Go: goroutine buat parallel I/O. Di Node.js kamu pakai `Promise.all()`, di Go pakai goroutine + errgroup.

```go
// Contoh: Dashboard endpoint yang butuh data dari beberapa source
// GET /api/v1/dashboard

// BEFORE (Sequential — SLOW):
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
    userID := c.GetUint("userID")

    // Sequential: 50ms + 30ms + 40ms + 20ms = 140ms total
    boards, _ := h.boardService.GetMyBoards(userID)         // 50ms
    tasks, _ := h.taskService.GetMyTasks(userID, 1, 10)     // 30ms
    stats, _ := h.statsService.GetUserStats(userID)          // 40ms
    notifications, _ := h.notifService.GetRecent(userID, 5)  // 20ms

    c.JSON(200, gin.H{
        "boards":        boards,
        "tasks":         tasks,
        "stats":         stats,
        "notifications": notifications,
    })
}

// AFTER (Parallel — FAST):
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
    userID := c.GetUint("userID")
    ctx := c.Request.Context()

    g, ctx := errgroup.WithContext(ctx)

    var boards []model.Board
    var tasks []model.Task
    var stats *UserStats
    var notifications []Notification

    // Parallel: max(50ms, 30ms, 40ms, 20ms) = 50ms total! (3x faster)
    g.Go(func() error {
        var err error
        boards, err = h.boardService.GetMyBoards(userID)
        return err
    })

    g.Go(func() error {
        var err error
        tasks, _, err = h.taskService.GetMyTasks(userID, 1, 10)
        return err
    })

    g.Go(func() error {
        var err error
        stats, err = h.statsService.GetUserStats(userID)
        return err
    })

    g.Go(func() error {
        var err error
        notifications, err = h.notifService.GetRecent(userID, 5)
        return err
    })

    if err := g.Wait(); err != nil {
        response.Error(c, 500, "Failed to fetch dashboard data")
        return
    }

    response.Success(c, 200, "Dashboard", gin.H{
        "boards":        boards,
        "tasks":         tasks,
        "stats":         stats,
        "notifications": notifications,
    })
}
```

### Worker Pool Pattern

```go
// Worker pool: batasi jumlah goroutine yang jalan bersamaan
// Contoh: process batch tasks (email notification ke semua assignees)

func sendBatchNotifications(ctx context.Context, tasks []model.Task) error {
    const maxWorkers = 10 // Max 10 goroutine bersamaan

    // Semaphore pattern pakai buffered channel
    sem := make(chan struct{}, maxWorkers)
    g, ctx := errgroup.WithContext(ctx)

    for _, task := range tasks {
        task := task // Capture loop variable

        g.Go(func() error {
            // Acquire semaphore slot
            select {
            case sem <- struct{}{}: // Block kalau udah 10 goroutine
            case <-ctx.Done():
                return ctx.Err()
            }
            defer func() { <-sem }() // Release slot

            return sendNotification(ctx, task)
        })
    }

    return g.Wait()
}
```

### Fan-out / Fan-in Pattern

```go
// Fan-out: distribute work ke multiple goroutines
// Fan-in: collect results dari multiple goroutines

func searchTasksAcrossBoards(ctx context.Context, boardIDs []uint, query string) ([]model.Task, error) {
    // Fan-out: search di setiap board secara parallel
    resultCh := make(chan []model.Task, len(boardIDs))
    errCh := make(chan error, len(boardIDs))

    for _, boardID := range boardIDs {
        boardID := boardID
        go func() {
            tasks, err := searchTasksInBoard(ctx, boardID, query)
            if err != nil {
                errCh <- err
                return
            }
            resultCh <- tasks
        }()
    }

    // Fan-in: collect semua results
    var allTasks []model.Task
    for range boardIDs {
        select {
        case tasks := <-resultCh:
            allTasks = append(allTasks, tasks...)
        case err := <-errCh:
            return nil, err
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }

    return allTasks, nil
}
```

### Context Timeout

```go
// SELALU set timeout di setiap operation yang bisa lambat!

func (h *TaskHandler) GetByBoardID(c *gin.Context) {
    // Set timeout: kalau 5 detik ga selesai, cancel
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    boardID, _ := strconv.ParseUint(c.Param("boardId"), 10, 64)
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    // Pass context ke service dan repository
    tasks, total, err := h.taskService.GetByBoardIDWithContext(ctx, uint(boardID), page, limit)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            response.Error(c, http.StatusGatewayTimeout, "Request timeout")
            return
        }
        response.Error(c, http.StatusInternalServerError, "Failed to fetch tasks")
        return
    }

    response.Success(c, http.StatusOK, "Board tasks", gin.H{
        "tasks": tasks,
        "total": total,
    })
}

// Di repository, pass context ke GORM:
func (r *TaskRepository) FindByBoardIDWithContext(ctx context.Context, boardID uint, page, limit int) ([]model.Task, int64, error) {
    var tasks []model.Task
    var total int64

    offset := (page - 1) * limit

    err := r.db.WithContext(ctx). // ← Context propagation!
        Where("board_id = ?", boardID).
        Preload("Assignee").
        Offset(offset).
        Limit(limit).
        Order("created_at DESC").
        Find(&tasks).Error

    r.db.WithContext(ctx).Model(&model.Task{}).Where("board_id = ?", boardID).Count(&total)

    return tasks, total, err
}
```

---

## 19.15 JSON Performance

### encoding/json — Standard Library

Go standard library `encoding/json` itu **correct** dan **safe**, tapi **bukan yang paling cepat**. Kenapa? Karena pakai reflection.

```go
import "encoding/json"

// Standard library — pakai reflection buat baca struct tags
data, err := json.Marshal(task)       // Encode
err = json.Unmarshal(data, &task)     // Decode
```

### Alternatif JSON Libraries

```
┌────────────────────┬──────────────┬───────────────┬───────────────────────┐
│ Library            │ Speed vs std │ Compatibility │ Notes                 │
├────────────────────┼──────────────┼───────────────┼───────────────────────┤
│ encoding/json      │ 1x (baseline)│ ✅ 100%      │ Standard library      │
│ sonic (bytedance)  │ 2-5x faster  │ ✅ Drop-in   │ Pakai SIMD, JIT       │
│ go-json (goccy)    │ 2-3x faster  │ ✅ Drop-in   │ Zero-copy decode      │
│ easyjson           │ 3-5x faster  │ ⚠️ Code gen  │ Butuh generate code   │
│ json-iterator      │ 2-3x faster  │ ✅ Drop-in   │ Compatible API        │
│ segmentio encoding │ 2-4x faster  │ ⚠️ Partial   │ Slightly different API│
└────────────────────┴──────────────┴───────────────┴───────────────────────┘
```

### Benchmark Comparison

```go
// internal/handler/json_bench_test.go
package handler

import (
    "encoding/json"
    "testing"
    "time"

    "github.com/bytedance/sonic"
    gojson "github.com/goccy/go-json"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/pkg/response"
)

func createTestResponse() response.Response {
    tasks := make([]model.Task, 50)
    now := time.Now()
    for i := range tasks {
        tasks[i] = model.Task{
            ID:          uint(i + 1),
            Title:       "Performance testing task with a reasonably long title",
            Description: "This is a description that might contain some longer text to be more realistic",
            Status:      model.StatusInProgress,
            Priority:    model.PriorityHigh,
            BoardID:     1,
            CreatedAt:   now,
            UpdatedAt:   now,
        }
    }
    return response.Response{
        Status:  200,
        Message: "Board tasks",
        Data:    tasks,
    }
}

func BenchmarkJSONMarshal_StdLib(b *testing.B) {
    data := createTestResponse()
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        json.Marshal(data)
    }
}

func BenchmarkJSONMarshal_Sonic(b *testing.B) {
    data := createTestResponse()
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        sonic.Marshal(data)
    }
}

func BenchmarkJSONMarshal_GoJSON(b *testing.B) {
    data := createTestResponse()
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        gojson.Marshal(data)
    }
}

// Unmarshal benchmarks
func BenchmarkJSONUnmarshal_StdLib(b *testing.B) {
    data := createTestResponse()
    jsonBytes, _ := json.Marshal(data)
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        var result response.Response
        json.Unmarshal(jsonBytes, &result)
    }
}

func BenchmarkJSONUnmarshal_Sonic(b *testing.B) {
    data := createTestResponse()
    jsonBytes, _ := sonic.Marshal(data)
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        var result response.Response
        sonic.Unmarshal(jsonBytes, &result)
    }
}
```

```bash
# Run JSON benchmarks
go test -bench=BenchmarkJSON -benchmem ./internal/handler/

# Expected output (approximate):
# BenchmarkJSONMarshal_StdLib-8     50000    28234 ns/op    16384 B/op   120 allocs/op
# BenchmarkJSONMarshal_Sonic-8     180000     8245 ns/op     9216 B/op    12 allocs/op
# BenchmarkJSONMarshal_GoJSON-8    120000    12034 ns/op    12288 B/op    35 allocs/op

# Sonic = ~3.4x faster, ~44% less memory, ~90% fewer allocations!
```

### Kapan Ganti dari encoding/json?

```
Pakai encoding/json (default) kalau:
- API traffic rendah (< 100 RPS)
- JSON payload kecil
- Ga mau tambah dependency
- Compatibility paling penting

Ganti ke sonic/go-json kalau:
- High traffic (> 1000 RPS)
- JSON payload besar (list 100+ items)
- JSON marshal/unmarshal muncul di profiling
- Butuh tiap millisecond
- Gin udah pakai sonic by default! (cek go.mod kamu)
```

Fun fact: Gin framework udah include `bytedance/sonic` sebagai dependency (lihat `go.mod` project kamu). Gin otomatis pakai sonic kalau available.

---

## 19.16 HTTP Performance

### Keep-Alive

```
Tanpa Keep-Alive:
━━━━━━━━━━━━━━━━━
Request 1: [DNS → TCP → TLS → Request → Response → Close]  150ms
Request 2: [DNS → TCP → TLS → Request → Response → Close]  150ms
Request 3: [DNS → TCP → TLS → Request → Response → Close]  150ms
Total: 450ms

Dengan Keep-Alive:
━━━━━━━━━━━━━━━━━━
Request 1: [DNS → TCP → TLS → Request → Response]  150ms
Request 2: [                    Request → Response]   20ms  ← Reuse connection!
Request 3: [                    Request → Response]   20ms
Total: 190ms (2.4x faster!)
```

Go HTTP server dan Gin sudah **enable keep-alive by default**. Ga perlu setting apa-apa. Tapi pastikan proxy (Nginx) juga enable keep-alive ke upstream.

### Compression (gzip/brotli)

```go
// Gin gzip middleware
import "github.com/gin-contrib/gzip"

func main() {
    r := gin.Default()

    // Enable gzip compression
    // Level 1-9 (1=fastest, 9=smallest)
    // Recommended: gzip.DefaultCompression (level 6)
    r.Use(gzip.Gzip(gzip.DefaultCompression))

    // Atau dengan options:
    r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{
        "/health",              // Health check ga perlu compress
        "/debug/pprof/",        // pprof output ga perlu compress
    })))

    router.Setup(r, handlers, authService, redisClient)
    r.Run(":8080")
}
```

```
Compression impact:
━━━━━━━━━━━━━━━━━━
JSON response 50 tasks (uncompressed):  ~25KB
JSON response 50 tasks (gzip level 6):  ~3KB    (88% smaller!)
JSON response 50 tasks (brotli):        ~2.5KB  (90% smaller!)

Trade-off:
- CPU cost: ~0.5-2ms per response (biasanya worth it)
- Network savings: 80-90% bandwidth reduction
- TTFB: sedikit lebih lambat (compression time), tapi transfer lebih cepat
```

### Pagination Best Practices

**Offset-based Pagination** (yang udah kita implement):

```go
// Current implementation:
// GET /api/v1/boards/:boardId/tasks?page=1&limit=10
func (r *TaskRepository) FindByBoardID(boardID uint, page, limit int) ([]model.Task, int64, error) {
    offset := (page - 1) * limit
    err := r.db.Where("board_id = ?", boardID).
        Offset(offset).Limit(limit).
        Order("created_at DESC").
        Find(&tasks).Error
    // ...
}

// Problem: OFFSET is SLOW on large datasets!
// OFFSET 100000 = database harus scan 100,000 rows, terus buang 99,990
// Semakin besar offset, semakin lambat
```

**Cursor-based Pagination** (lebih scalable):

```go
// Cursor-based: GET /api/v1/boards/:boardId/tasks?cursor=2024-01-15T10:30:00Z&limit=10
func (r *TaskRepository) FindByBoardIDCursor(
    boardID uint,
    cursor time.Time,  // created_at of last item from previous page
    limit int,
) ([]model.Task, *time.Time, error) {
    var tasks []model.Task

    query := r.db.Where("board_id = ? AND deleted_at IS NULL", boardID)

    if !cursor.IsZero() {
        // WHERE created_at < cursor (older items)
        query = query.Where("created_at < ?", cursor)
    }

    err := query.
        Preload("Assignee").
        Limit(limit + 1). // Request 1 extra to know if there's more
        Order("created_at DESC").
        Find(&tasks).Error

    if err != nil {
        return nil, nil, err
    }

    // Check if there are more results
    var nextCursor *time.Time
    if len(tasks) > limit {
        tasks = tasks[:limit]                     // Remove extra item
        lastCreatedAt := tasks[limit-1].CreatedAt
        nextCursor = &lastCreatedAt               // This becomes the next cursor
    }

    return tasks, nextCursor, nil
}

// Response:
// {
//   "data": {
//     "tasks": [...],
//     "next_cursor": "2024-01-15T10:30:00Z",
//     "has_more": true
//   }
// }
```

```
Offset vs Cursor Pagination:
━━━━━━━━━━━━━━━━━━━━━━━━━━━

┌─────────────┬──────────────────────────┬────────────────────────────┐
│ Aspek       │ Offset                   │ Cursor                     │
├─────────────┼──────────────────────────┼────────────────────────────┤
│ Speed       │ Degradasi di page besar  │ Konsisten (pakai index)    │
│ Complexity  │ Simple                   │ Sedikit lebih complex      │
│ Jump to     │ Bisa (page=50)           │ Ga bisa (sequential only)  │
│ page        │                          │                            │
│ Real-time   │ Data bisa shift          │ Konsisten, ga ada data     │
│ data        │ (miss/duplicate)         │ yang terlewat              │
│ Use case    │ Admin panel, dashboard   │ Feed, infinite scroll      │
│ SQL         │ OFFSET N LIMIT M        │ WHERE id < cursor LIMIT M  │
│ Performance │ O(offset + limit)        │ O(limit) ← SELALU CEPAT   │
└─────────────┴──────────────────────────┴────────────────────────────┘
```

---

## 19.17 Memory Optimization

### Struct Field Ordering (Alignment)

Go compiler melakukan **memory alignment** pada struct fields. Urutan field mempengaruhi ukuran struct!

```go
// BAD ordering — 40 bytes (banyak padding)
type TaskBad struct {
    Active    bool      // 1 byte  + 7 bytes padding
    ID        uint64    // 8 bytes
    Priority  byte      // 1 byte  + 7 bytes padding
    BoardID   uint64    // 8 bytes
    Status    byte      // 1 byte  + 7 bytes padding
}
// Total: 1+7 + 8 + 1+7 + 8 + 1+7 = 40 bytes

// GOOD ordering — 24 bytes (minimal padding)
type TaskGood struct {
    ID        uint64    // 8 bytes
    BoardID   uint64    // 8 bytes
    Active    bool      // 1 byte
    Priority  byte      // 1 byte
    Status    byte      // 1 byte + 5 bytes padding
}
// Total: 8 + 8 + 1+1+1+5 = 24 bytes (40% SMALLER!)
```

```go
// Verify struct size:
fmt.Println(unsafe.Sizeof(TaskBad{}))   // 40
fmt.Println(unsafe.Sizeof(TaskGood{}))  // 24
```

**Rule of thumb:** Urutkan fields dari yang BESAR ke KECIL.

```bash
# Tool: fieldalignment — otomatis detect dan fix struct ordering
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

# Check semua structs
fieldalignment ./...

# Auto-fix
fieldalignment -fix ./...
```

### String vs []byte

```go
// Strings di Go itu immutable. Setiap "modifikasi" = copy baru.

// BAD: String concatenation in loop
func buildTaskListBad(tasks []model.Task) string {
    result := ""
    for _, t := range tasks {
        result += t.Title + "\n"   // ← Setiap += = new string allocation!
    }                                // 1000 tasks = 1000 allocations!
    return result
}

// GOOD: strings.Builder (atau bytes.Buffer)
func buildTaskListGood(tasks []model.Task) string {
    var sb strings.Builder
    sb.Grow(len(tasks) * 50)  // Pre-allocate estimated capacity
    for _, t := range tasks {
        sb.WriteString(t.Title)
        sb.WriteByte('\n')
    }
    return sb.String()  // 1 allocation!
}
```

### Slice Pre-allocation

```go
// BAD: Append without pre-allocation
func getTaskIDs(tasks []model.Task) []uint {
    var ids []uint  // len=0, cap=0
    for _, t := range tasks {
        ids = append(ids, t.ID)
        // Internally: cap grows 0→1→2→4→8→16→32→...
        // Each growth = copy all elements to new array!
    }
    return ids
}

// GOOD: Pre-allocate with make
func getTaskIDs(tasks []model.Task) []uint {
    ids := make([]uint, 0, len(tasks))  // len=0, cap=len(tasks)
    for _, t := range tasks {
        ids = append(ids, t.ID)  // No reallocation needed!
    }
    return ids
}

// BEST: Direct index (kalau ukuran pasti)
func getTaskIDs(tasks []model.Task) []uint {
    ids := make([]uint, len(tasks))  // len=len(tasks), cap=len(tasks)
    for i, t := range tasks {
        ids[i] = t.ID  // Direct write, no append overhead
    }
    return ids
}
```

### Benchmark Memory Optimization

```go
// internal/handler/memory_bench_test.go
package handler

import (
    "strings"
    "testing"
)

func BenchmarkStringConcat(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        result := ""
        for j := 0; j < 100; j++ {
            result += "task title here\n"
        }
        _ = result
    }
}

func BenchmarkStringBuilder(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        var sb strings.Builder
        sb.Grow(100 * 16)
        for j := 0; j < 100; j++ {
            sb.WriteString("task title here\n")
        }
        _ = sb.String()
    }
}

func BenchmarkSliceAppend(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        var ids []uint
        for j := 0; j < 1000; j++ {
            ids = append(ids, uint(j))
        }
        _ = ids
    }
}

func BenchmarkSlicePreAlloc(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        ids := make([]uint, 0, 1000)
        for j := 0; j < 1000; j++ {
            ids = append(ids, uint(j))
        }
        _ = ids
    }
}
```

```bash
# Run benchmarks
go test -bench=BenchmarkString -benchmem ./internal/handler/
go test -bench=BenchmarkSlice -benchmem ./internal/handler/

# Expected results:
# BenchmarkStringConcat-8     5000   312456 ns/op   530048 B/op   99 allocs/op
# BenchmarkStringBuilder-8  200000     8234 ns/op     3456 B/op    3 allocs/op
#                                                    (38x faster, 153x less memory!)
#
# BenchmarkSliceAppend-8     50000    28456 ns/op   41088 B/op   18 allocs/op
# BenchmarkSlicePreAlloc-8  200000     6123 ns/op    8192 B/op    1 allocs/op
#                                                    (4.6x faster, 5x less memory!)
```

---

## 19.18 Nginx Performance Tuning

Di Bab 14 kita udah setup Nginx. Sekarang kita focus ke performance tuning.

### Worker Configuration

```nginx
# /etc/nginx/nginx.conf

# Worker processes = jumlah CPU cores
# `auto` = otomatis detect
worker_processes auto;

# Worker connections = max simultaneous connections per worker
# Total connections = worker_processes × worker_connections
events {
    worker_connections 4096;    # Default 1024, naikkan!
    multi_accept on;            # Accept multiple connections at once
    use epoll;                  # Linux: pakai epoll (faster than select/poll)
}
```

### HTTP Performance Settings

```nginx
http {
    # === KEEPALIVE ===
    keepalive_timeout 65;           # Client keepalive timeout
    keepalive_requests 1000;        # Max requests per keepalive connection

    # Upstream keepalive (ke Go app)
    upstream task_manager {
        server 127.0.0.1:8080;
        keepalive 32;              # Keep 32 connections open to backend
        keepalive_requests 1000;
        keepalive_timeout 60s;
    }

    # === BUFFERS ===
    client_body_buffer_size 16k;    # Buffer for POST body
    client_header_buffer_size 1k;   # Buffer for headers
    client_max_body_size 10m;       # Max upload size
    large_client_header_buffers 4 8k;

    # Proxy buffers (response dari backend)
    proxy_buffer_size 128k;
    proxy_buffers 4 256k;
    proxy_busy_buffers_size 256k;

    # === COMPRESSION ===
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 4;             # 1-9, 4 is good balance
    gzip_min_length 256;           # Jangan compress response < 256 bytes
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml
        application/xml+rss
        application/x-javascript;

    # === TIMEOUTS ===
    proxy_connect_timeout 5s;       # Timeout connect ke backend
    proxy_send_timeout 30s;         # Timeout kirim request ke backend
    proxy_read_timeout 30s;         # Timeout baca response dari backend
    send_timeout 30s;               # Timeout kirim response ke client

    # === PROXY CACHE (optional) ===
    proxy_cache_path /var/cache/nginx
        levels=1:2
        keys_zone=api_cache:10m     # 10MB shared memory for keys
        max_size=1g                 # 1GB disk cache
        inactive=10m                # Remove cached items after 10 min inactive
        use_temp_path=off;

    server {
        listen 80;
        server_name api.taskmanager.com;

        location /api/ {
            proxy_pass http://task_manager;
            proxy_http_version 1.1;                   # Required for keepalive
            proxy_set_header Connection "";            # Required for keepalive
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

            # Cache GET requests for 60 seconds
            # proxy_cache api_cache;
            # proxy_cache_methods GET;
            # proxy_cache_valid 200 60s;
            # proxy_cache_valid 404 10s;
            # add_header X-Cache-Status $upstream_cache_status;
        }

        location /health {
            proxy_pass http://task_manager;
            proxy_http_version 1.1;
            proxy_set_header Connection "";

            # Cache health check aggressively
            proxy_cache api_cache;
            proxy_cache_valid 200 5s;
        }
    }
}
```

### Nginx Performance Checklist

```
┌──────────────────────────────────────────────────────────────┐
│  NGINX PERFORMANCE CHECKLIST                                  │
├──────────────────────────────────────────────────────────────┤
│  [x] worker_processes = auto (match CPU cores)               │
│  [x] worker_connections = 4096+                              │
│  [x] multi_accept on                                        │
│  [x] use epoll (Linux)                                      │
│  [x] keepalive_timeout 65                                   │
│  [x] keepalive ke upstream (keepalive 32)                   │
│  [x] gzip on (level 4-6)                                   │
│  [x] gzip_min_length 256                                    │
│  [x] Proper buffer sizes                                    │
│  [x] Timeout yang reasonable (5-30s)                        │
│  [x] proxy_http_version 1.1 (untuk keepalive upstream)      │
│  [ ] SSL session cache (kalau HTTPS)                        │
│  [ ] HTTP/2 enabled (kalau HTTPS)                           │
│  [ ] Rate limiting per endpoint                             │
│  [ ] Access log buffering (buffer=32k)                      │
└──────────────────────────────────────────────────────────────┘
```

---

## 19.19 Database Migration Performance

### The Problem: Locking

```
Standard migration pada tabel besar:

ALTER TABLE tasks ADD COLUMN priority_score INTEGER DEFAULT 0;

Apa yang terjadi di PostgreSQL:
1. Acquire ACCESS EXCLUSIVE lock pada tabel tasks
2. Scan SEMUA rows dan set default value
3. Rewrite table (bisa menit-jam di tabel besar)
4. Selama proses ini: SEMUA query ke tasks table BLOCKED!
5. API hang, user marah, HRD nanya "kenapa sistem down?"

Timeline:
00:00  Migration start, lock acquired
00:00  → GET /tasks BLOCKED ❌
00:01  → POST /tasks BLOCKED ❌
00:05  → Still migrating...
00:10  Migration done, lock released
00:10  → GET /tasks works again ✅
00:10  → 10 menit downtime 💀
```

### Safe Migration Patterns

**Pattern 1: ADD COLUMN without DEFAULT (instant)**

```sql
-- PostgreSQL 11+: ADD COLUMN with DEFAULT is fast (metadata-only change)
-- Tapi di PostgreSQL < 11, atau kalau ragu:

-- Step 1: Add column WITHOUT default (instant, no lock)
ALTER TABLE tasks ADD COLUMN priority_score INTEGER;

-- Step 2: Backfill in batches (no lock)
UPDATE tasks SET priority_score = 0
WHERE id BETWEEN 1 AND 10000;

UPDATE tasks SET priority_score = 0
WHERE id BETWEEN 10001 AND 20000;
-- ... repeat ...

-- Step 3: Add default for future rows
ALTER TABLE tasks ALTER COLUMN priority_score SET DEFAULT 0;

-- Step 4: Add NOT NULL constraint (kalau perlu)
ALTER TABLE tasks ADD CONSTRAINT tasks_priority_score_nn
  CHECK (priority_score IS NOT NULL) NOT VALID;  -- NOT VALID = ga check existing rows

-- Step 5: Validate constraint (takes shared lock, not exclusive)
ALTER TABLE tasks VALIDATE CONSTRAINT tasks_priority_score_nn;
```

**Pattern 2: CREATE INDEX CONCURRENTLY**

```sql
-- BAD: Blocks all writes!
CREATE INDEX idx_tasks_priority ON tasks(priority);

-- GOOD: Doesn't block writes (takes longer, tapi ga ada downtime)
CREATE INDEX CONCURRENTLY idx_tasks_priority ON tasks(priority);
```

**Pattern 3: Batch Processing di Go**

```go
// migrations/backfill_priority_score.go
func backfillPriorityScore(db *gorm.DB) error {
    batchSize := 1000
    var lastID uint = 0

    for {
        result := db.Exec(`
            UPDATE tasks
            SET priority_score = 0
            WHERE id > ? AND id <= ? AND priority_score IS NULL
        `, lastID, lastID+uint(batchSize))

        if result.RowsAffected == 0 {
            break // No more rows to update
        }

        lastID += uint(batchSize)

        // Log progress
        log.Printf("Backfilled up to ID %d, rows affected: %d", lastID, result.RowsAffected)

        // Sleep between batches to reduce load
        time.Sleep(100 * time.Millisecond)
    }

    return nil
}
```

### Migration Safety Checklist

```
Before running migration on production:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[x] Test migration on staging with production-size data
[x] Estimate lock duration (EXPLAIN ANALYZE ALTER TABLE ...)
[x] Schedule during low-traffic window
[x] CREATE INDEX CONCURRENTLY (never without CONCURRENTLY on production!)
[x] Backfill in batches, not single UPDATE
[x] Have rollback plan ready
[x] Monitor during migration (Grafana dashboard open)
[x] Alert team before migration
[x] Verify application handles NULL values for new columns
```

---

## 19.20 Performance Monitoring

### Prometheus Metrics for Performance

Di Bab 15 kita udah setup Prometheus. Sekarang kita tambahkan metrics yang specific ke performance.

```go
// internal/middleware/metrics.go
package middleware

import (
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP request duration histogram (P50, P95, P99)
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
            // Buckets: 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
        },
        []string{"method", "path", "status"},
    )

    // HTTP requests total counter
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    // Request size
    httpRequestSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_size_bytes",
            Help:    "HTTP request size in bytes",
            Buckets: prometheus.ExponentialBuckets(100, 10, 6), // 100B, 1KB, 10KB, 100KB, 1MB, 10MB
        },
        []string{"method", "path"},
    )

    // Response size
    httpResponseSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_response_size_bytes",
            Help:    "HTTP response size in bytes",
            Buckets: prometheus.ExponentialBuckets(100, 10, 6),
        },
        []string{"method", "path"},
    )

    // Database query duration
    dbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
        },
        []string{"operation", "table"},
    )

    // Cache hit/miss
    cacheHits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total cache hits",
        },
        []string{"cache_type"}, // "redis", "local"
    )

    cacheMisses = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total cache misses",
        },
        []string{"cache_type"},
    )
)

// PrometheusMiddleware records HTTP request metrics
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // Process request
        c.Next()

        // Record metrics
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        path := c.FullPath() // Use route pattern, not actual path (avoid cardinality explosion)
        if path == "" {
            path = "unknown"
        }

        httpRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
        httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
        httpResponseSize.WithLabelValues(c.Request.Method, path).Observe(float64(c.Writer.Size()))
        httpRequestSize.WithLabelValues(c.Request.Method, path).Observe(float64(c.Request.ContentLength))
    }
}
```

### Grafana Dashboard — Performance Panels

```
Grafana Dashboard: "Task Manager Performance"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Row 1: Request Overview
┌─────────────────┬─────────────────┬─────────────────┬─────────────────┐
│  Request Rate   │  Error Rate     │  P95 Latency    │  Apdex Score    │
│  1,234 req/s    │  0.05%          │  42ms           │  0.97           │
└─────────────────┴─────────────────┴─────────────────┴─────────────────┘

Row 2: Latency Distribution
┌───────────────────────────────────────────────────────────────────────┐
│  P50: 12ms    P90: 35ms    P95: 42ms    P99: 120ms    Max: 450ms   │
│  [Histogram chart showing latency distribution over time]           │
└───────────────────────────────────────────────────────────────────────┘

Row 3: Per-Endpoint Latency
┌───────────────────────────────────────────────────────────────────────┐
│  GET  /api/v1/boards/:id/tasks    P95: 38ms                         │
│  POST /api/v1/auth/login          P95: 180ms  (bcrypt = expected)   │
│  POST /api/v1/boards/:id/tasks    P95: 25ms                         │
│  GET  /api/v1/boards              P95: 15ms                         │
│  GET  /health                     P95: 2ms                          │
└───────────────────────────────────────────────────────────────────────┘
```

**PromQL queries untuk dashboard:**

```promql
# Request rate (requests per second)
rate(http_requests_total[5m])

# Error rate (percentage)
sum(rate(http_requests_total{status=~"5.."}[5m])) /
sum(rate(http_requests_total[5m])) * 100

# P50 latency
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))

# P95 latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# P99 latency
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# P95 latency per endpoint
histogram_quantile(0.95,
  sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path)
)

# Apdex score (T = 0.2s / 200ms)
(
  sum(rate(http_request_duration_seconds_bucket{le="0.2"}[5m])) +
  sum(rate(http_request_duration_seconds_bucket{le="0.8"}[5m]))
) / 2 /
sum(rate(http_request_duration_seconds_count[5m]))

# Cache hit ratio
sum(rate(cache_hits_total[5m])) /
(sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100

# DB query P95
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))
```

### Alerting on Performance

```yaml
# prometheus/alert_rules.yml
groups:
  - name: performance_alerts
    rules:
      # P95 latency too high
      - alert: HighP95Latency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "P95 latency above 500ms for 5 minutes"
          description: "P95 latency is {{ $value | printf \"%.2f\" }}s"

      # P99 latency critical
      - alert: CriticalP99Latency
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 2
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "P99 latency above 2s"

      # Error rate too high
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) /
          sum(rate(http_requests_total[5m])) > 0.01
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "Error rate above 1% for 3 minutes"

      # Goroutine leak detection
      - alert: GoroutineLeak
        expr: go_goroutines > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Goroutine count above 1000 for 10 minutes"
          description: "Current goroutine count: {{ $value }}"

      # DB query slow
      - alert: SlowDBQueries
        expr: histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m])) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database P95 query time above 100ms"
```

---

## 19.21 Performance Budget

### Apa Itu Performance Budget?

Performance budget = **batas performance yang ga boleh dilewatin**. Sama kayak budget keuangan — kalau kamu set budget Rp 5 juta/bulan, kamu pantau pengeluaran dan alarm kalau mau over.

### Set Performance Budget untuk Task Manager

```
TASK MANAGER API — PERFORMANCE BUDGET
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

API Response Time:
┌─────────────────────────────────────┬────────────┬────────────┐
│ Endpoint                           │ P95 Target │ P99 Target │
├─────────────────────────────────────┼────────────┼────────────┤
│ GET  /health                       │ < 10ms     │ < 50ms     │
│ POST /api/v1/auth/login            │ < 300ms    │ < 500ms    │
│ POST /api/v1/auth/register         │ < 400ms    │ < 600ms    │
│ GET  /api/v1/boards                │ < 100ms    │ < 200ms    │
│ GET  /api/v1/boards/:id            │ < 100ms    │ < 200ms    │
│ POST /api/v1/boards                │ < 100ms    │ < 200ms    │
│ GET  /api/v1/boards/:id/tasks      │ < 150ms    │ < 300ms    │
│ POST /api/v1/boards/:id/tasks      │ < 100ms    │ < 200ms    │
│ GET  /api/v1/tasks/my              │ < 150ms    │ < 300ms    │
│ PUT  /api/v1/tasks/:id             │ < 100ms    │ < 200ms    │
└─────────────────────────────────────┴────────────┴────────────┘

Note: Login/Register target lebih tinggi karena bcrypt = CPU intensive.
Itu expected dan acceptable.

Throughput:
┌─────────────────────────────────────┬────────────┐
│ Metric                             │ Target     │
├─────────────────────────────────────┼────────────┤
│ Minimum RPS (per instance)         │ > 500      │
│ Target RPS (per instance)          │ > 1,000    │
│ Maximum concurrent connections     │ > 500      │
└─────────────────────────────────────┴────────────┘

Reliability:
┌─────────────────────────────────────┬────────────┐
│ Metric                             │ Target     │
├─────────────────────────────────────┼────────────┤
│ Error rate                         │ < 0.1%     │
│ Uptime                             │ > 99.9%    │
│ TTFB (Time To First Byte)          │ < 200ms    │
│ Apdex score (T=200ms)              │ > 0.95     │
└─────────────────────────────────────┴────────────┘

Resource:
┌─────────────────────────────────────┬────────────┐
│ Metric                             │ Target     │
├─────────────────────────────────────┼────────────┤
│ Memory usage (per instance)        │ < 256MB    │
│ CPU usage (steady state)           │ < 50%      │
│ Goroutine count (steady state)     │ < 200      │
│ DB connection pool utilization     │ < 70%      │
│ Redis connection pool utilization  │ < 60%      │
└─────────────────────────────────────┴────────────┘
```

### Enforce Performance Budget di CI/CD

```yaml
# .github/workflows/performance-check.yml
name: Performance Budget Check

on:
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Run Benchmarks
        run: |
          go test -bench=. -benchmem -count=5 ./... > current.txt

      - name: Compare with Baseline
        run: |
          # Download baseline benchmarks from main branch
          # (stored as artifact or in repo)
          go install golang.org/x/perf/cmd/benchstat@latest
          benchstat baseline.txt current.txt > comparison.txt
          cat comparison.txt

          # Fail if any benchmark regressed more than 10%
          if grep -q '+[1-9][0-9]\.' comparison.txt; then
            echo "Performance regression detected!"
            exit 1
          fi

      - name: Load Test with k6
        run: |
          docker compose up -d
          sleep 10
          k6 run loadtest/smoke.js
          # k6 thresholds will fail the test if performance budget exceeded
```

---

## 19.22 Performance Checklist — Before Production

### Pre-Launch Performance Review

```
╔══════════════════════════════════════════════════════════════════╗
║              PRE-PRODUCTION PERFORMANCE CHECKLIST                ║
╠══════════════════════════════════════════════════════════════════╣
║                                                                  ║
║  DATABASE                                                        ║
║  [ ] EXPLAIN ANALYZE semua query di critical path                ║
║  [ ] Index untuk semua WHERE clause yang sering dipakai          ║
║  [ ] Ga ada N+1 query (Preload/Joins sudah benar)               ║
║  [ ] Connection pool di-tune (MaxOpenConns, MaxIdleConns)        ║
║  [ ] Slow query logging enabled (> 100ms)                        ║
║  [ ] Partial indexes WHERE deleted_at IS NULL                    ║
║  [ ] pg_stat_statements extension installed                      ║
║                                                                  ║
║  APPLICATION                                                     ║
║  [ ] pprof available (tapi ga exposed ke public)                 ║
║  [ ] Context timeout di semua DB dan HTTP calls                  ║
║  [ ] Ga ada goroutine leak (pprof goroutine check)               ║
║  [ ] JSON library di-benchmark (encoding/json vs sonic)          ║
║  [ ] Slice pre-allocation di hot paths                           ║
║  [ ] sync.Pool di hot paths (kalau perlu)                        ║
║  [ ] Struct field alignment checked                              ║
║  [ ] No string concatenation in loops                            ║
║  [ ] HTTP client reused (bukan bikin baru tiap request)          ║
║  [ ] Gin mode = release di production                            ║
║                                                                  ║
║  CACHING                                                         ║
║  [ ] Redis cache untuk endpoint yang sering di-hit               ║
║  [ ] TTL strategy defined per data type                          ║
║  [ ] Cache invalidation implemented                              ║
║  [ ] Cache warming strategy (kalau perlu)                        ║
║  [ ] HTTP caching headers (Cache-Control, ETag)                  ║
║                                                                  ║
║  INFRASTRUCTURE                                                  ║
║  [ ] Nginx performance tuned (workers, keepalive, gzip)          ║
║  [ ] Nginx keepalive ke upstream enabled                         ║
║  [ ] SSL session cache enabled (kalau HTTPS)                     ║
║  [ ] HTTP/2 enabled                                              ║
║  [ ] CDN untuk static assets (kalau ada)                         ║
║                                                                  ║
║  TESTING                                                         ║
║  [ ] Smoke test passed                                           ║
║  [ ] Load test passed (expected traffic)                         ║
║  [ ] Stress test done (tau breaking point)                       ║
║  [ ] Soak test done (no memory leak)                             ║
║  [ ] Performance budget defined dan enforced                     ║
║                                                                  ║
║  MONITORING                                                      ║
║  [ ] Prometheus metrics untuk request duration, error rate        ║
║  [ ] Grafana dashboard dengan P50/P95/P99                        ║
║  [ ] Alert rules untuk latency, error rate, goroutine count      ║
║  [ ] Baseline metrics recorded (before launch)                   ║
║  [ ] On-call rotation setup                                      ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
```

---

## 19.23 Tools Comparison

### Load Testing Tools

```
┌──────────────┬──────────┬────────────┬──────────────┬────────────┬────────────────┐
│ Feature      │ k6       │ Locust     │ Gatling      │ JMeter     │ Artillery      │
├──────────────┼──────────┼────────────┼──────────────┼────────────┼────────────────┤
│ Language     │ JS (Go   │ Python     │ Scala/Java   │ Java (XML) │ JS (Node.js)   │
│              │ runtime) │            │              │            │                │
│ Script       │ JS/TS    │ Python     │ Scala/Java   │ GUI+XML    │ YAML+JS        │
│ Performance  │ High     │ Medium     │ High         │ Medium     │ Medium         │
│ Learning     │ Easy     │ Easy       │ Medium       │ Hard       │ Easy           │
│ CI/CD        │ ✅ Easy  │ ✅ OK     │ ✅ OK       │ ⚠️ Hard   │ ✅ Easy       │
│ Cloud        │ Grafana  │ Locust.io  │ Gatling Ent. │ BlazeMeter │ Artillery.io   │
│              │ Cloud    │            │              │            │                │
│ Protocols    │ HTTP,WS, │ HTTP       │ HTTP,WS,     │ HTTP,FTP,  │ HTTP,WS,       │
│              │ gRPC     │            │ gRPC,MQTT    │ JDBC,LDAP  │ Socket.io      │
│ Real-time    │ CLI      │ Web UI     │ Reports      │ Listeners  │ CLI            │
│ Dashboard    │ output   │            │              │            │                │
│ License      │ AGPL-3.0 │ MIT        │ Apache 2.0   │ Apache 2.0 │ MPL-2.0        │
│ Best for     │ Dev+CI   │ Python     │ Enterprise   │ Legacy     │ Quick tests    │
│              │          │ teams      │              │            │                │
├──────────────┼──────────┼────────────┼──────────────┼────────────┼────────────────┤
│ wrk          │ ab       │ hey        │ vegeta       │ bombardier │ plow           │
├──────────────┼──────────┼────────────┼──────────────┼────────────┼────────────────┤
│ C (Lua)      │ C        │ Go         │ Go           │ Go         │ Go             │
│ Very fast    │ Simple   │ Modern ab  │ CLI/lib      │ Fast       │ Nice output    │
│ Single EP    │ Single   │ Single EP  │ Programmable │ Single EP  │ Realtime chart │
│ Lua scripts  │ No       │ No script  │ Stdin-based  │ No script  │ No script      │
│              │ scripting│            │              │            │                │
│ Quick bench  │ Quick    │ Quick test │ Load pattern │ Quick test │ Quick + visual │
│              │ test     │            │ testing      │            │                │
└──────────────┴──────────┴────────────┴──────────────┴────────────┴────────────────┘

Rekomendasi:
- Baru mulai → k6 (mudah, powerful, CI/CD friendly)
- Python team → Locust (familiar Python syntax)
- Enterprise/complex → Gatling
- Quick single endpoint → wrk atau hey
- Legacy project → JMeter (reluctantly)
```

### Profiling Tools

```
┌──────────────────┬────────────────────────────────────────────┐
│ Tool             │ Use Case                                   │
├──────────────────┼────────────────────────────────────────────┤
│ go tool pprof    │ CPU, memory, goroutine profiling (Go)      │
│ go tool trace    │ Goroutine scheduling, latency analysis     │
│ go test -bench   │ Function-level benchmarking                │
│ benchstat        │ Compare benchmark results                  │
│ fieldalignment   │ Check struct memory alignment              │
│ golangci-lint    │ Static analysis (includes perf linters)    │
│ gopls            │ IDE integration for analysis               │
│ Pyroscope        │ Continuous profiling (production)          │
│ Grafana Tempo    │ Distributed tracing                        │
│ Jaeger           │ Distributed tracing (alternative)          │
└──────────────────┴────────────────────────────────────────────┘
```

---

## 19.24 Cheat Sheet

### pprof Commands

```bash
# === CAPTURE PROFILES ===
# CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap (current memory usage)
go tool pprof http://localhost:6060/debug/pprof/heap

# Allocs (all allocations)
go tool pprof http://localhost:6060/debug/pprof/allocs

# Goroutine
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Block
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex
go tool pprof http://localhost:6060/debug/pprof/mutex

# === INTERACTIVE COMMANDS ===
(pprof) top 20          # Top 20 functions by CPU/memory
(pprof) top -cum 20     # Top 20 by cumulative time
(pprof) list FuncName   # Show annotated source code
(pprof) web             # Open call graph in browser
(pprof) weblist FuncName # Annotated source in browser
(pprof) disasm FuncName # Disassembly
(pprof) peek FuncName   # Show callers and callees
(pprof) tree            # Tree view of call graph

# === WEB UI ===
# Open flame graph in browser
go tool pprof -http=:8081 profile.prof

# Compare two profiles
go tool pprof -base old.prof new.prof

# === FROM TESTS ===
go test -cpuprofile=cpu.prof -bench=. ./...
go test -memprofile=mem.prof -bench=. ./...
go tool pprof -http=:8081 cpu.prof

# === SAVE/LOAD PROFILES ===
# Save
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30
curl -o heap.prof http://localhost:6060/debug/pprof/heap

# Load saved profile
go tool pprof cpu.prof
go tool pprof -http=:8081 heap.prof
```

### k6 Patterns

```javascript
// === BASIC STRUCTURE ===
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<200'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/health');
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(1);
}

// === AUTHENTICATED REQUEST ===
const token = getToken();
const params = {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
};
const res = http.get(url, params);

// === POST JSON ===
const payload = JSON.stringify({ title: 'New Task', priority: 'high' });
const res = http.post(url, payload, params);

// === STAGES ===
export const options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp up
    { duration: '5m', target: 50 },   // Steady
    { duration: '1m', target: 0 },    // Ramp down
  ],
};

// === CUSTOM METRICS ===
import { Trend, Rate, Counter } from 'k6/metrics';
const loginTime = new Trend('login_duration');
const errorRate = new Rate('errors');
const taskCount = new Counter('tasks_created');

loginTime.add(res.timings.duration);
errorRate.add(res.status !== 200);
taskCount.add(1);

// === PER-ENDPOINT THRESHOLD ===
export const options = {
  thresholds: {
    'http_req_duration{name:Login}': ['p(95)<500'],
    'http_req_duration{name:GetTasks}': ['p(95)<200'],
  },
};
const res = http.get(url, { tags: { name: 'GetTasks' } });
```

### EXPLAIN ANALYZE Patterns

```sql
-- === BASIC ===
EXPLAIN ANALYZE SELECT * FROM tasks WHERE board_id = 1;

-- === WHAT TO LOOK FOR ===
-- Seq Scan       → BAD (full table scan) → Add index!
-- Index Scan     → GOOD (using index)
-- Bitmap Scan    → OK (multiple index conditions)
-- Sort           → Check if can use index for ordering
-- Nested Loop    → Check join condition has index
-- Hash Join      → Normal for larger joins
-- actual rows=X  → Compare with estimated rows

-- === USEFUL QUERIES ===
-- Find slow queries (requires pg_stat_statements):
SELECT query, calls, mean_exec_time, total_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Find missing indexes:
SELECT schemaname, tablename, seq_scan, idx_scan,
       seq_scan - idx_scan AS too_many_seqscans
FROM pg_stat_user_tables
WHERE seq_scan - idx_scan > 0
ORDER BY too_many_seqscans DESC;

-- Find unused indexes:
SELECT indexname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
AND indexname NOT LIKE '%pkey%'
ORDER BY indexname;

-- Table/index sizes:
SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) AS index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Go Optimization Patterns

```go
// === PRE-ALLOCATE SLICES ===
// Bad
var items []int
for i := 0; i < n; i++ { items = append(items, i) }
// Good
items := make([]int, 0, n)
for i := 0; i < n; i++ { items = append(items, i) }

// === STRING BUILDER ===
// Bad
s := ""
for _, v := range items { s += v.String() }
// Good
var sb strings.Builder
sb.Grow(n * avgLen)
for _, v := range items { sb.WriteString(v.String()) }

// === SYNC.POOL ===
var pool = sync.Pool{
    New: func() interface{} { return new(bytes.Buffer) },
}
buf := pool.Get().(*bytes.Buffer)
buf.Reset()
defer pool.Put(buf)

// === AVOID POINTER FOR SMALL STRUCTS ===
// Bad (heap allocation)
func newStatus() *string { s := "active"; return &s }
// Good (stack)
func newStatus() string { return "active" }

// === PARALLEL I/O ===
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { /* fetch A */ return nil })
g.Go(func() error { /* fetch B */ return nil })
err := g.Wait()

// === CONTEXT TIMEOUT ===
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
db.WithContext(ctx).Find(&tasks)

// === STRUCT ALIGNMENT ===
// Order fields from largest to smallest
type Good struct {
    ID   uint64  // 8 bytes
    Name string  // 16 bytes (ptr+len)
    Age  uint8   // 1 byte
}
```

---

## 19.25 Checklist Akhir Bab 19

```
BAB 19: PERFORMANCE & LOAD TESTING — CHECKLIST
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

KONSEP:
[ ] Paham kenapa performance penting (revenue, UX, competitive advantage)
[ ] Paham performance metrics: latency (P50/P95/P99), throughput (RPS), error rate
[ ] Paham perbedaan average vs percentile
[ ] Paham SLA vs SLO vs SLI
[ ] Paham Apdex score
[ ] Paham TTFB (Time To First Byte)

PROFILING:
[ ] Setup pprof di Task Manager API (port terpisah)
[ ] Bisa capture CPU profile dan baca flame graph
[ ] Bisa capture memory profile (heap dan allocs)
[ ] Paham escape analysis (go build -gcflags '-m')
[ ] Bisa detect goroutine leak via pprof
[ ] Paham common goroutine leak patterns dan fix-nya
[ ] Tau cara compare 2 profile (base comparison)

BENCHMARKING:
[ ] Bisa menulis benchmark function (testing.B, b.N)
[ ] Paham b.ResetTimer, b.ReportAllocs, -benchmem
[ ] Bisa pakai sub-benchmarks (b.Run)
[ ] Bisa pakai benchstat untuk compare results
[ ] Sudah benchmark JSON marshal/unmarshal
[ ] Sudah benchmark password hashing

LOAD TESTING:
[ ] Install k6
[ ] Bisa tulis basic k6 script (VUs, duration, checks, thresholds)
[ ] Bisa tulis full scenario k6 script (register → login → CRUD)
[ ] Paham stages (ramp-up, steady, ramp-down)
[ ] Paham 5 tipe load test (smoke, load, stress, spike, soak)
[ ] Tau perbedaan k6 vs wrk vs ab dan kapan pakai mana
[ ] Bisa integrate k6 di CI/CD

DATABASE:
[ ] Bisa pakai EXPLAIN ANALYZE
[ ] Paham index types (B-tree, GIN, partial)
[ ] Sudah buat index yang proper untuk Task Manager queries
[ ] Paham N+1 problem dan fix dengan Preload/Joins
[ ] Paham connection pooling tuning
[ ] Tau cara enable slow query logging

CACHING:
[ ] Paham cache hierarchy (in-memory → Redis → DB)
[ ] Bisa implement cache-aside pattern
[ ] Paham TTL strategy
[ ] Paham cache invalidation patterns
[ ] Paham HTTP caching headers (ETag, Cache-Control)

OPTIMIZATION:
[ ] Paham connection pool tuning (DB, Redis, HTTP client)
[ ] Bisa pakai errgroup untuk parallel I/O
[ ] Paham worker pool dan semaphore pattern
[ ] Paham context timeout dan propagation
[ ] Tau alternatif JSON libraries dan kapan ganti
[ ] Paham HTTP keep-alive dan compression
[ ] Paham cursor-based vs offset pagination
[ ] Paham struct field alignment
[ ] Paham string builder vs concatenation
[ ] Paham slice pre-allocation

INFRASTRUCTURE:
[ ] Nginx performance tuning (workers, keepalive, gzip, buffers)
[ ] Paham safe database migration patterns
[ ] Paham CREATE INDEX CONCURRENTLY
[ ] Paham batch processing untuk large table migrations

MONITORING & BUDGET:
[ ] Prometheus metrics untuk request duration, error rate
[ ] Grafana dashboard dengan latency percentiles
[ ] Alert rules untuk P95 latency dan error rate
[ ] Performance budget defined (target per endpoint)
[ ] Performance budget enforced di CI/CD
[ ] Pre-production performance checklist completed
```

---

## Penutup

Bab ini banyak banget materinya. Dan itu disengaja — karena performance itu **bukan satu hal**, tapi **kombinasi banyak hal kecil** yang terakumulasi.

Kunci dari performance optimization:

1. **Measure first, optimize later.** Jangan nebak. Profiling > feeling.
2. **Optimize bottleneck, bukan semuanya.** 80/20 rule: 80% improvement datang dari 20% code.
3. **Benchmark before and after.** Kalau ga ada angka "sebelum", kamu ga bisa buktiin "sesudah" lebih baik.
4. **Performance is a feature, not an afterthought.** Bake it into your development process.
5. **Production is different from development.** Always load test with realistic data and traffic patterns.

Yang perlu kamu lakuin sekarang:
1. Setup pprof di Task Manager API
2. Run benchmark buat critical paths (login, get tasks, create task)
3. Tulis k6 script buat full flow testing
4. Buat index yang proper di database
5. Set performance budget dan monitor di Grafana

Di bab selanjutnya, kita akan masuk ke topik yang lebih advanced. Tapi fondasi performance ini akan selalu relevan — di setiap system yang kamu build, di setiap scale yang kamu hadapi.

> "Premature optimization is the root of all evil." — Donald Knuth
>
> Tapi dia juga bilang: "We should not pass up our opportunities in that critical 3%."
>
> Artinya: jangan optimize semuanya, tapi **jangan abaikan** yang penting.
