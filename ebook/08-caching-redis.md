# Bab 8: Caching dengan Redis

## Kenapa Bab Ini Penting?

Kalo kamu pernah ngerasain API yang lambat padahal query-nya "cuma" SELECT biasa — itu karena setiap request selalu ke database. Redis hadir untuk menyelesaikan ini. Di bab ini kita bakal belajar gimana project Task Manager kita pakai Redis buat caching, rate limiting, session storage, dan pub/sub.

> **Analogi gampang:** Database = gudang barang. Redis = meja kasir. Kamu ga mungkin jalan ke gudang setiap kali ada pembeli. Barang yang sering diminta, taruh di meja kasir.

---

## 8.1 Redis Overview — Apa Itu Redis?

Redis = **RE**mote **DI**ctionary **S**erver.

Singkatnya: Redis itu **in-memory data store**. Data disimpan di RAM, bukan di disk. Makanya super cepat.

### Angka bicara:

| Operation | PostgreSQL | Redis |
|-----------|-----------|-------|
| Read latency | 1-10ms | **0.1-0.5ms** |
| Throughput | ~10K ops/sec | **100K+ ops/sec** |
| Storage | Disk (persisten) | RAM (+ opsional disk) |

### Use Cases Redis:

| Use Case | Contoh |
|----------|--------|
| **Caching** | Cache hasil query GET /boards, GET /tasks |
| **Rate Limiting** | Batasi 100 request/menit per IP |
| **Session Storage** | Simpan session user login |
| **Pub/Sub** | Real-time notification, chat |
| **Queue** | Job queue sederhana |
| **Leaderboard** | Sorted set untuk ranking |
| **Distributed Lock** | Cegah race condition antar server |

### Install Redis

```bash
# macOS
brew install redis
brew services start redis

# Verify
redis-cli ping
# Output: PONG

# Docker (recommended buat development)
docker run -d --name redis -p 6379:6379 redis:alpine
```

---

## 8.2 Redis Data Types

Redis bukan cuma key-value biasa. Dia punya beberapa tipe data yang powerful.

### 1. String — Tipe paling basic

```bash
# Set value
SET user:1:name "Irzan Aldi"

# Get value
GET user:1:name
# "Irzan Aldi"

# Set dengan expiry (TTL) — 300 detik
SET board:5:cache '{"id":5,"title":"Sprint 1"}' EX 300

# Increment (berguna buat counter / rate limiting)
INCR page_views
INCR page_views
GET page_views
# "2"

# Set hanya kalau belum ada (berguna buat distributed lock)
SETNX lock:process_job "locked"
```

### 2. Hash — Object / Map

```bash
# Simpan user sebagai hash
HSET user:1 name "Irzan" email "irzan@mail.com" role "admin"

# Get satu field
HGET user:1 name
# "Irzan"

# Get semua field
HGETALL user:1
# 1) "name"
# 2) "Irzan"
# 3) "email"
# 4) "irzan@mail.com"
# 5) "role"
# 6) "admin"

# Update satu field aja
HSET user:1 name "Irzan Aldi"
```

**Kapan pakai Hash vs String?**
- Kalau data punya banyak field dan kamu sering update per-field -> **Hash**
- Kalau data selalu di-read/write secara utuh (JSON blob) -> **String**

### 3. List — Array / Queue

```bash
# Push ke kiri (depan)
LPUSH notifications:user:1 "Task assigned to you"
LPUSH notifications:user:1 "Board updated"

# Pop dari kanan (FIFO queue)
RPOP notifications:user:1
# "Task assigned to you"

# Ambil range
LRANGE notifications:user:1 0 -1
# Semua item

# Cek panjang
LLEN notifications:user:1
```

**Use case:** Notification queue, recent activity log.

### 4. Set — Unique collection (tanpa duplikat)

```bash
# Tambah member ke set
SADD board:1:members 1 2 3 5
SADD board:2:members 2 3 7

# Cek member
SISMEMBER board:1:members 5
# (integer) 1 = true

# Semua member
SMEMBERS board:1:members
# 1, 2, 3, 5

# Intersection — member yang ada di kedua board
SINTER board:1:members board:2:members
# 2, 3
```

**Use case:** Tag system, online users, mutual friends.

### 5. Sorted Set — Set + score (otomatis diurutkan)

```bash
# Leaderboard: user dan jumlah task completed
ZADD leaderboard 15 "user:1" 23 "user:2" 8 "user:3"

# Top 3 (score tertinggi)
ZREVRANGE leaderboard 0 2 WITHSCORES
# 1) "user:2" -> 23
# 2) "user:1" -> 15
# 3) "user:3" -> 8

# Rank user tertentu
ZREVRANK leaderboard "user:1"
# (integer) 1 — ranking ke-2 (0-based)

# Increment score
ZINCRBY leaderboard 5 "user:3"
```

**Use case:** Leaderboard, priority queue, time-series data.

---

## 8.3 Redis di Go — Library `go-redis`

Project kita pakai library `github.com/redis/go-redis/v9` — library Redis paling populer di Go.

### Install

```bash
go get github.com/redis/go-redis/v9
```

### Config Struct

Pertama, lihat config kita di `internal/config/config.go`:

```go
// internal/config/config.go

type Config struct {
    Env       string
    Port      string
    DB        DatabaseConfig
    Redis     RedisConfig
    JWTSecret string
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
}

func Load() *Config {
    return &Config{
        // ... config lainnya ...
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
        },
    }
}
```

Environment variables di `.env`:

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

### Connection Setup

Ini file `internal/config/redis.go` dari project kita:

```go
// internal/config/redis.go

package config

import (
    "context"
    "log"

    "github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *Config) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
        Password: cfg.Redis.Password,
        DB:       0,
    })

    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        log.Printf("Warning: Redis not connected: %v (caching disabled)", err)
        return client
    }

    log.Println("Redis connected successfully")
    return client
}
```

**Beberapa hal penting di sini:**

1. **`redis.Options`** — mirip kayak buat database, kita set address, password, dan DB number (Redis punya 16 DB, default 0).

2. **`Ping`** — test koneksi ke Redis. Kalau gagal, kita cuma log warning (ga fatal). Ini design choice: app tetap jalan walau Redis mati, cuma caching disabled.

3. **Graceful degradation** — kita return client walau gagal connect. Nanti di middleware/handler, kita cek error dari Redis dan fallback ke database langsung.

### Wiring di main.go

```go
// cmd/api/main.go

func main() {
    cfg := config.Load()
    db := config.ConnectDB(cfg)

    // ... auto migrate ...

    // Connect Redis
    redisClient := config.ConnectRedis(cfg)

    // ... setup repo, service, handler ...

    r := gin.Default()
    router.Setup(r, handlers, authService, redisClient)

    r.Run(":" + cfg.Port)
}
```

### go-redis Basic Operations di Go

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

func main() {
    ctx := context.Background()
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // ========== STRING ==========
    // Set dengan TTL 5 menit
    rdb.Set(ctx, "greeting", "Halo dari Go!", 5*time.Minute)

    // Get
    val, err := rdb.Get(ctx, "greeting").Result()
    if err == redis.Nil {
        fmt.Println("Key tidak ditemukan")
    } else if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println(val) // "Halo dari Go!"
    }

    // ========== HASH ==========
    rdb.HSet(ctx, "user:1", map[string]interface{}{
        "name":  "Irzan",
        "email": "irzan@mail.com",
        "role":  "admin",
    })

    name, _ := rdb.HGet(ctx, "user:1", "name").Result()
    fmt.Println(name) // "Irzan"

    allFields, _ := rdb.HGetAll(ctx, "user:1").Result()
    fmt.Println(allFields) // map[name:Irzan email:irzan@mail.com role:admin]

    // ========== LIST ==========
    rdb.LPush(ctx, "queue:emails", "email1", "email2", "email3")

    email, _ := rdb.RPop(ctx, "queue:emails").Result()
    fmt.Println(email) // "email1" (FIFO)

    // ========== SET ==========
    rdb.SAdd(ctx, "online_users", "user:1", "user:2", "user:3")

    isMember, _ := rdb.SIsMember(ctx, "online_users", "user:1").Result()
    fmt.Println(isMember) // true

    // ========== SORTED SET ==========
    rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "user:1"})
    rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 250, Member: "user:2"})

    // Top 10
    topUsers, _ := rdb.ZRevRangeWithScores(ctx, "leaderboard", 0, 9).Result()
    for _, z := range topUsers {
        fmt.Printf("%s: %.0f points\n", z.Member, z.Score)
    }

    // ========== TTL & DELETE ==========
    rdb.Expire(ctx, "user:1", 1*time.Hour)    // Set TTL
    ttl, _ := rdb.TTL(ctx, "user:1").Result() // Check TTL
    fmt.Println(ttl)

    rdb.Del(ctx, "greeting") // Delete key
}
```

> **Pattern penting:** Di go-redis, setiap command return `*redis.XxxCmd`. Kamu bisa chain `.Result()` buat dapet value + error, atau `.Err()` kalau cuma mau cek error.

---

## 8.4 Caching Patterns

Ada 3 pattern utama untuk caching. Pilih berdasarkan use case.

### Pattern 1: Cache-Aside (Lazy Loading)

**Paling umum dipakai.** App yang ngatur kapan read/write ke cache.

```
Client Request
    ↓
Cek Redis → HIT? → Return dari cache
    ↓ MISS
Query Database
    ↓
Simpan ke Redis (dengan TTL)
    ↓
Return ke client
```

**Kelebihan:**
- Hanya data yang diminta yang masuk cache (hemat memory)
- Kalau Redis mati, app tetap jalan (fallback ke DB)

**Kekurangan:**
- Request pertama selalu lambat (cold start / cache miss)
- Data bisa stale kalau ga di-invalidate

**Kapan pakai:** Hampir semua GET endpoint. GET /boards/:id, GET /tasks/:id.

### Pattern 2: Write-Through

Setiap kali nulis ke DB, langsung tulis juga ke cache.

```
Client Write Request
    ↓
Tulis ke Database
    ↓
Tulis ke Redis (sync)
    ↓
Return response
```

**Kelebihan:**
- Cache selalu up-to-date
- Ga ada stale data

**Kekurangan:**
- Write jadi lebih lambat (2x write: DB + Redis)
- Banyak data di cache yang mungkin ga pernah dibaca

**Kapan pakai:** Data yang sering dibaca setelah ditulis. Profile user, settings.

### Pattern 3: Write-Behind (Write-Back)

Tulis ke cache dulu, nanti async tulis ke DB.

```
Client Write Request
    ↓
Tulis ke Redis
    ↓
Return response (cepat!)
    ↓
Background worker tulis ke DB (async)
```

**Kelebihan:**
- Write sangat cepat
- Bisa batch write ke DB

**Kekurangan:**
- Risk data loss kalau Redis crash sebelum sync ke DB
- Complexity tinggi

**Kapan pakai:** Analytics, page views, activity log — data yang boleh hilang sedikit.

### Perbandingan

| Pattern | Read Speed | Write Speed | Consistency | Complexity |
|---------|-----------|-------------|-------------|------------|
| Cache-Aside | Cepat (setelah warm) | Normal | Eventual | Rendah |
| Write-Through | Cepat | Lambat | Strong | Sedang |
| Write-Behind | Cepat | **Sangat cepat** | Weak | **Tinggi** |

> **Rekomendasi untuk project ini:** Pakai **Cache-Aside** untuk GET endpoints dan **manual invalidation** saat update/delete. Simple, effective, dan robust.

---

## 8.5 Implementing Cache-Aside di Project Kita

Sekarang kita implementasiin cache-aside pattern buat endpoint GET board dan GET tasks.

### Helper: Cache Service

Pertama, buat helper generic buat operasi cache:

```go
// internal/cache/cache.go

package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type CacheService struct {
    redis *redis.Client
    ctx   context.Context
}

func NewCacheService(redisClient *redis.Client) *CacheService {
    return &CacheService{
        redis: redisClient,
        ctx:   context.Background(),
    }
}

// Get — ambil dari cache, return true kalau HIT
func (c *CacheService) Get(key string, dest interface{}) (bool, error) {
    val, err := c.redis.Get(c.ctx, key).Result()
    if err == redis.Nil {
        return false, nil // MISS — bukan error
    }
    if err != nil {
        return false, err // Redis error
    }

    // Unmarshal JSON ke struct tujuan
    if err := json.Unmarshal([]byte(val), dest); err != nil {
        return false, err
    }

    return true, nil // HIT
}

// Set — simpan ke cache dengan TTL
func (c *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.redis.Set(c.ctx, key, data, ttl).Err()
}

// Delete — hapus key dari cache
func (c *CacheService) Delete(keys ...string) error {
    return c.redis.Del(c.ctx, keys...).Err()
}

// DeletePattern — hapus semua key yang match pattern
// Contoh: DeletePattern("board:*") hapus semua cache board
func (c *CacheService) DeletePattern(pattern string) error {
    iter := c.redis.Scan(c.ctx, 0, pattern, 0).Iterator()
    for iter.Next(c.ctx) {
        if err := c.redis.Del(c.ctx, iter.Val()).Err(); err != nil {
            return err
        }
    }
    return iter.Err()
}

// Key builder helpers
func BoardKey(id uint) string {
    return fmt.Sprintf("board:%d", id)
}

func BoardsByOwnerKey(ownerID uint, page, limit int) string {
    return fmt.Sprintf("boards:owner:%d:page:%d:limit:%d", ownerID, page, limit)
}

func TaskKey(id uint) string {
    return fmt.Sprintf("task:%d", id)
}

func TasksByBoardKey(boardID uint, page, limit int) string {
    return fmt.Sprintf("tasks:board:%d:page:%d:limit:%d", boardID, page, limit)
}
```

### Caching di Board Service

Sekarang kita upgrade `BoardService` buat pakai cache:

```go
// internal/service/board.go (versi dengan cache)

package service

import (
    "errors"
    "time"

    "github.com/irzan-aldi/task-manager/internal/cache"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/internal/repository"
)

const boardCacheTTL = 5 * time.Minute

type BoardService struct {
    repo  *repository.BoardRepository
    cache *cache.CacheService
}

func NewBoardService(repo *repository.BoardRepository, cache *cache.CacheService) *BoardService {
    return &BoardService{repo: repo, cache: cache}
}

func (s *BoardService) GetByID(id uint) (*model.Board, error) {
    // STEP 1: Cek cache
    var board model.Board
    key := cache.BoardKey(id)

    hit, err := s.cache.Get(key, &board)
    if err != nil {
        // Redis error — log aja, lanjut ke DB
        // log.Printf("Cache error: %v", err)
    }
    if hit {
        return &board, nil // Cache HIT! Langsung return
    }

    // STEP 2: Cache MISS — query database
    dbBoard, err := s.repo.FindByID(id)
    if err != nil {
        return nil, errors.New("board not found")
    }

    // STEP 3: Simpan ke cache
    _ = s.cache.Set(key, dbBoard, boardCacheTTL)

    return dbBoard, nil
}

func (s *BoardService) GetByOwnerID(ownerID uint, page, limit int) ([]model.Board, int64, error) {
    // Cek cache
    type boardsResult struct {
        Boards []model.Board `json:"boards"`
        Total  int64         `json:"total"`
    }

    var cached boardsResult
    key := cache.BoardsByOwnerKey(ownerID, page, limit)

    hit, _ := s.cache.Get(key, &cached)
    if hit {
        return cached.Boards, cached.Total, nil
    }

    // Cache MISS
    boards, total, err := s.repo.FindByOwnerID(ownerID, page, limit)
    if err != nil {
        return nil, 0, err
    }

    // Simpan ke cache
    _ = s.cache.Set(key, boardsResult{
        Boards: boards,
        Total:  total,
    }, boardCacheTTL)

    return boards, total, nil
}

func (s *BoardService) Create(ownerID uint, req model.CreateBoardRequest) (*model.Board, error) {
    board := &model.Board{
        Title:       req.Title,
        Description: req.Description,
        OwnerID:     ownerID,
    }

    if err := s.repo.Create(board); err != nil {
        return nil, errors.New("failed to create board")
    }

    // Invalidate list cache (karena ada board baru)
    _ = s.cache.DeletePattern("boards:owner:*")

    return board, nil
}

func (s *BoardService) Update(id, ownerID uint, req model.UpdateBoardRequest) (*model.Board, error) {
    board, err := s.repo.FindByID(id)
    if err != nil {
        return nil, errors.New("board not found")
    }

    if board.OwnerID != ownerID {
        return nil, errors.New("not authorized to update this board")
    }

    if req.Title != nil {
        board.Title = *req.Title
    }
    if req.Description != nil {
        board.Description = *req.Description
    }

    if err := s.repo.Update(board); err != nil {
        return nil, errors.New("failed to update board")
    }

    // Invalidate cache untuk board ini dan list cache
    _ = s.cache.Delete(cache.BoardKey(id))
    _ = s.cache.DeletePattern("boards:owner:*")

    return board, nil
}

func (s *BoardService) Delete(id, ownerID uint) error {
    board, err := s.repo.FindByID(id)
    if err != nil {
        return errors.New("board not found")
    }

    if board.OwnerID != ownerID {
        return errors.New("not authorized to delete this board")
    }

    if err := s.repo.Delete(id); err != nil {
        return err
    }

    // Invalidate cache
    _ = s.cache.Delete(cache.BoardKey(id))
    _ = s.cache.DeletePattern("boards:owner:*")
    _ = s.cache.DeletePattern("tasks:board:" + fmt.Sprintf("%d", id) + ":*")

    return nil
}
```

### Flow Lengkap — Visual

```
GET /api/v1/boards/5
    ↓
Handler: BoardHandler.GetByID()
    ↓
Service: BoardService.GetByID(5)
    ↓
cache.Get("board:5", &board)
    ↓
┌─── HIT ───────────────────┐
│ Return board dari Redis    │ ← 0.1ms
│ (skip database query)      │
└────────────────────────────┘
    ↓ MISS
repo.FindByID(5)             ← Query PostgreSQL (1-10ms)
    ↓
cache.Set("board:5", board, 5m)  ← Simpan ke Redis
    ↓
Return board ke client
```

---

## 8.6 Cache Invalidation

> "There are only two hard things in Computer Science: cache invalidation and naming things." — Phil Karlton

### Strategy 1: TTL (Time To Live)

Cara paling simple — set expiry time. Data otomatis hilang setelah waktu tertentu.

```go
// Cache selama 5 menit
rdb.Set(ctx, "board:5", boardJSON, 5*time.Minute)

// Cek sisa TTL
ttl, _ := rdb.TTL(ctx, "board:5").Result()
fmt.Println(ttl) // 4m32s (misalnya)
```

**Pilih TTL berdasarkan seberapa sering data berubah:**

| Data | TTL Recommended | Alasan |
|------|----------------|--------|
| Board detail | 5 menit | Jarang diupdate |
| Task list | 2 menit | Sering berubah |
| User profile | 10 menit | Hampir ga pernah berubah |
| Dashboard stats | 30 detik | Harus agak fresh |

### Strategy 2: Manual Invalidation

Hapus cache secara explicit saat data berubah. Ini yang kita pakai di atas.

```go
// Saat UPDATE board
func (s *BoardService) Update(id, ownerID uint, req model.UpdateBoardRequest) (*model.Board, error) {
    // ... update di database ...

    // Hapus cache spesifik
    _ = s.cache.Delete(cache.BoardKey(id))

    // Hapus list cache yang mungkin terpengaruh
    _ = s.cache.DeletePattern("boards:owner:*")

    return board, nil
}

// Saat DELETE board
func (s *BoardService) Delete(id, ownerID uint) error {
    // ... delete di database ...

    // Hapus semua cache terkait
    _ = s.cache.Delete(cache.BoardKey(id))
    _ = s.cache.DeletePattern("boards:owner:*")

    return nil
}
```

### Strategy 3: TTL + Manual (Recommended)

Kombinasi keduanya:
- **TTL** sebagai safety net — data pasti expire walau kita lupa invalidate
- **Manual invalidation** untuk instant consistency saat write

```go
// Set dengan TTL
_ = s.cache.Set(key, board, 5*time.Minute) // Safety net: max 5 menit stale

// Tapi juga invalidate saat update
_ = s.cache.Delete(cache.BoardKey(id)) // Instant invalidation
```

### Hati-hati: Cache Stampede

Bayangkan 1000 user request `/boards/5` bersamaan, dan cache baru expired. Semua 1000 request bakal miss cache dan query DB bareng-bareng. Ini namanya **cache stampede**.

Solusi sederhana — **singleflight**:

```go
import "golang.org/x/sync/singleflight"

var group singleflight.Group

func (s *BoardService) GetByID(id uint) (*model.Board, error) {
    key := cache.BoardKey(id)

    // Cek cache
    var board model.Board
    hit, _ := s.cache.Get(key, &board)
    if hit {
        return &board, nil
    }

    // Singleflight: walaupun 1000 goroutine panggil ini,
    // cuma 1 yang beneran query DB. Sisanya nunggu hasil yang sama.
    result, err, _ := group.Do(key, func() (interface{}, error) {
        dbBoard, err := s.repo.FindByID(id)
        if err != nil {
            return nil, err
        }
        _ = s.cache.Set(key, dbBoard, 5*time.Minute)
        return dbBoard, nil
    })

    if err != nil {
        return nil, err
    }
    return result.(*model.Board), nil
}
```

---

## 8.7 Rate Limiting dengan Redis

Rate limiting penting banget buat:
- Cegah abuse / brute force
- Protect server dari overload
- Fair usage policy

### Implementasi di Project Kita

Ini file `internal/middleware/ratelimit.go` — sudah ada di project:

```go
// internal/middleware/ratelimit.go

package middleware

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/pkg/response"
    "github.com/redis/go-redis/v9"
)

func RateLimiter(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        ip := c.ClientIP()
        key := fmt.Sprintf("rate_limit:%s", ip)

        count, err := redisClient.Incr(ctx, key).Result()
        if err != nil {
            // Redis down — allow request (graceful degradation)
            c.Next()
            return
        }

        if count == 1 {
            redisClient.Expire(ctx, key, window)
        }

        if count > int64(limit) {
            response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded. Try again later.")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### Cara Kerjanya — Step by Step

```
Request dari IP 192.168.1.1
    ↓
key = "rate_limit:192.168.1.1"
    ↓
INCR key → count = 1 (request pertama)
    ↓
count == 1? → EXPIRE key 60s (set window 1 menit)
    ↓
count (1) > limit (100)? → NO → c.Next() ✅
    ↓
... 99 request lagi ...
    ↓
INCR key → count = 101
    ↓
count (101) > limit (100)? → YES → 429 Too Many Requests ❌
    ↓
... setelah 60 detik, key expire, counter reset ke 0 ...
```

### Dipasang di Router

```go
// internal/router/router.go

func Setup(r *gin.Engine, h *Handlers, authService *service.AuthService, redisClient *redis.Client) {
    r.GET("/health", handler.HealthCheck)

    // Rate limiter: 100 requests per minute per IP
    r.Use(middleware.RateLimiter(redisClient, 100, time.Minute))

    v1 := r.Group("/api/v1")
    {
        // ... routes ...
    }
}
```

### Rate Limiter yang Lebih Advanced: Per-User + Per-Endpoint

```go
// Rate limit berbeda per endpoint
func RateLimiterAdvanced(redisClient *redis.Client, limit int, window time.Duration, prefix string) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()

        // Prefer user ID (kalau authenticated), fallback ke IP
        identifier := c.ClientIP()
        if userID, exists := c.Get("userID"); exists {
            identifier = fmt.Sprintf("user:%d", userID.(uint))
        }

        key := fmt.Sprintf("rate_limit:%s:%s", prefix, identifier)

        count, err := redisClient.Incr(ctx, key).Result()
        if err != nil {
            c.Next()
            return
        }

        if count == 1 {
            redisClient.Expire(ctx, key, window)
        }

        // Tambahin header biar client tau sisa limit
        remaining := int64(limit) - count
        if remaining < 0 {
            remaining = 0
        }
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
        c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

        if count > int64(limit) {
            response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded. Try again later.")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

Contoh pakai:

```go
// Global: 100 req/menit
r.Use(middleware.RateLimiterAdvanced(redisClient, 100, time.Minute, "global"))

// Login: ketat, 5 req/menit (anti brute force)
auth.POST("/login",
    middleware.RateLimiterAdvanced(redisClient, 5, time.Minute, "login"),
    h.Auth.Login,
)

// API: 30 req/menit per user
protected.Use(middleware.RateLimiterAdvanced(redisClient, 30, time.Minute, "api"))
```

### Sliding Window Rate Limiter

Implementasi di atas pakai **fixed window** — counter reset setiap window selesai. Kelemahannya: kalau user kirim 100 request di detik ke-59, lalu 100 lagi di detik ke-61, total 200 request dalam 2 detik!

**Sliding window** lebih akurat pakai Sorted Set:

```go
func SlidingWindowRateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        key := fmt.Sprintf("sliding_rate:%s", c.ClientIP())
        now := time.Now()

        pipe := rdb.Pipeline()

        // Hapus entry yang sudah di luar window
        pipe.ZRemRangeByScore(ctx, key, "0",
            fmt.Sprintf("%d", now.Add(-window).UnixNano()))

        // Tambah request sekarang
        pipe.ZAdd(ctx, key, redis.Z{
            Score:  float64(now.UnixNano()),
            Member: fmt.Sprintf("%d", now.UnixNano()),
        })

        // Hitung total request dalam window
        countCmd := pipe.ZCard(ctx, key)

        // Set expiry key
        pipe.Expire(ctx, key, window)

        _, err := pipe.Exec(ctx)
        if err != nil {
            c.Next()
            return
        }

        if countCmd.Val() > int64(limit) {
            response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded.")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## 8.8 Redis Pub/Sub — Real-time Features

Pub/Sub = Publish/Subscribe. Satu service publish message, semua subscriber yang dengerin dapet message itu.

### Konsep

```
Publisher ──publish──→ Channel "task_updates"
                            ↓
                      ┌─────┴─────┐
                      ↓           ↓
               Subscriber 1   Subscriber 2
               (WebSocket)    (Notification)
```

### Basic Pub/Sub di Go

**Publisher:**

```go
func publishTaskUpdate(rdb *redis.Client, boardID uint, message string) {
    ctx := context.Background()
    channel := fmt.Sprintf("board:%d:updates", boardID)

    payload, _ := json.Marshal(map[string]interface{}{
        "event":   "task_updated",
        "message": message,
        "time":    time.Now().Unix(),
    })

    rdb.Publish(ctx, channel, payload)
}
```

**Subscriber:**

```go
func subscribeToBoard(rdb *redis.Client, boardID uint) {
    ctx := context.Background()
    channel := fmt.Sprintf("board:%d:updates", boardID)

    sub := rdb.Subscribe(ctx, channel)
    defer sub.Close()

    // Channel buat terima message
    ch := sub.Channel()

    for msg := range ch {
        fmt.Printf("Board %d update: %s\n", boardID, msg.Payload)
        // Kirim ke WebSocket client, push notification, dll
    }
}
```

### Contoh: Notify saat task di-update

```go
func (s *TaskService) Update(id, userID uint, req model.UpdateTaskRequest) (*model.Task, error) {
    // ... update logic ...

    if err := s.repo.Update(task); err != nil {
        return nil, errors.New("failed to update task")
    }

    // Publish event ke Redis Pub/Sub
    go publishTaskUpdate(s.redis, task.BoardID, fmt.Sprintf(
        "Task '%s' updated by user %d", task.Title, userID,
    ))

    return task, nil
}
```

### Kapan pakai Pub/Sub vs Message Queue?

| | Pub/Sub | Message Queue (Asynq/RabbitMQ) |
|---|---------|-------------------------------|
| Delivery | Fire and forget | Guaranteed delivery |
| Persistence | Tidak | Ya |
| Consumer offline | Message hilang | Message nunggu |
| Use case | Real-time notification | Background job processing |

> **Pub/Sub cocok untuk:** real-time updates yang boleh miss (dashboard live update, typing indicator). Kalau butuh guaranteed delivery, pakai message queue (Bab 9).

---

## 8.9 Session Storage dengan Redis

JWT itu stateless — server ga nyimpan state. Tapi kadang kamu butuh:
- **Blacklist token** saat logout
- **Store session data** yang terlalu besar buat JWT
- **Revoke token** kalau user ganti password

### Token Blacklist

```go
// internal/cache/session.go

package cache

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type SessionStore struct {
    redis *redis.Client
    ctx   context.Context
}

func NewSessionStore(rdb *redis.Client) *SessionStore {
    return &SessionStore{
        redis: rdb,
        ctx:   context.Background(),
    }
}

// BlacklistToken — simpan token yang sudah di-revoke
func (s *SessionStore) BlacklistToken(tokenString string, expiry time.Duration) error {
    key := fmt.Sprintf("blacklist:%s", tokenString)
    return s.redis.Set(s.ctx, key, "revoked", expiry).Err()
}

// IsBlacklisted — cek apakah token sudah di-blacklist
func (s *SessionStore) IsBlacklisted(tokenString string) bool {
    key := fmt.Sprintf("blacklist:%s", tokenString)
    _, err := s.redis.Get(s.ctx, key).Result()
    return err == nil // kalau key ada = blacklisted
}

// StoreSession — simpan session data
func (s *SessionStore) StoreSession(userID uint, data map[string]string, ttl time.Duration) error {
    key := fmt.Sprintf("session:%d", userID)
    pipe := s.redis.Pipeline()
    for field, value := range data {
        pipe.HSet(s.ctx, key, field, value)
    }
    pipe.Expire(s.ctx, key, ttl)
    _, err := pipe.Exec(s.ctx)
    return err
}

// GetSession — ambil session data
func (s *SessionStore) GetSession(userID uint) (map[string]string, error) {
    key := fmt.Sprintf("session:%d", userID)
    return s.redis.HGetAll(s.ctx, key).Result()
}

// DestroySession — hapus session (logout)
func (s *SessionStore) DestroySession(userID uint) error {
    key := fmt.Sprintf("session:%d", userID)
    return s.redis.Del(s.ctx, key).Err()
}
```

### Pakai di Auth Middleware

```go
func AuthMiddleware(authService *service.AuthService, sessionStore *cache.SessionStore) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... extract token dari header ...

        // Cek blacklist sebelum validate
        if sessionStore.IsBlacklisted(tokenString) {
            response.Error(c, http.StatusUnauthorized, "Token has been revoked")
            c.Abort()
            return
        }

        claims, err := authService.ValidateToken(tokenString)
        if err != nil {
            response.Error(c, http.StatusUnauthorized, "Invalid or expired token")
            c.Abort()
            return
        }

        c.Set("userID", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

### Logout Handler

```go
func (h *AuthHandler) Logout(c *gin.Context) {
    tokenString := extractToken(c) // ambil dari header

    // Blacklist token selama sisa TTL-nya
    // Kalau token expire 15 menit, blacklist juga 15 menit
    _ = h.sessionStore.BlacklistToken(tokenString, 15*time.Minute)

    response.Success(c, http.StatusOK, "Logged out successfully", nil)
}
```

---

## 8.10 Full Code: Caching GET /boards/:id — End to End

Ini contoh lengkap gimana caching terintegrasi dari handler sampai repository.

### Handler (ga berubah — caching di service layer)

```go
// internal/handler/board.go — handler TETAP sama, ga perlu tau soal cache

func (h *BoardHandler) GetByID(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid board ID")
        return
    }

    board, err := h.boardService.GetByID(uint(id))
    if err != nil {
        response.Error(c, http.StatusNotFound, err.Error())
        return
    }

    response.Success(c, http.StatusOK, "Board detail", board)
}
```

### Service (cache logic di sini)

```go
// internal/service/board.go — cache logic ada di service layer

func (s *BoardService) GetByID(id uint) (*model.Board, error) {
    // 1. Cek cache
    var board model.Board
    key := cache.BoardKey(id)

    hit, _ := s.cache.Get(key, &board)
    if hit {
        return &board, nil // Cache HIT
    }

    // 2. Cache MISS — query DB
    dbBoard, err := s.repo.FindByID(id)
    if err != nil {
        return nil, errors.New("board not found")
    }

    // 3. Store ke cache
    _ = s.cache.Set(key, dbBoard, boardCacheTTL)

    return dbBoard, nil
}
```

### Repository (ga berubah — repository pure database)

```go
// internal/repository/board.go — tetap pure database, ga tau soal cache

func (r *BoardRepository) FindByID(id uint) (*model.Board, error) {
    var board model.Board
    err := r.db.Preload("Owner").Preload("Tasks").First(&board, id).Error
    if err != nil {
        return nil, err
    }
    return &board, nil
}
```

### Wiring di main.go

```go
// cmd/api/main.go

func main() {
    cfg := config.Load()
    db := config.ConnectDB(cfg)

    // ... auto migrate ...

    redisClient := config.ConnectRedis(cfg)
    cacheService := cache.NewCacheService(redisClient)

    // Repositories
    boardRepo := repository.NewBoardRepository(db)
    taskRepo := repository.NewTaskRepository(db)

    // Services — inject cache
    boardService := service.NewBoardService(boardRepo, cacheService)
    taskService := service.NewTaskService(taskRepo, boardRepo, cacheService)

    // ... handlers, router ...
}
```

### Kenapa Cache Logic di Service Layer?

```
Handler  → Parse request, validate input, return response
Service  → Business logic + CACHING ← di sini
Repository → Pure database query
```

- **Handler** ga perlu tau data dari cache atau DB — dia cuma terima result
- **Repository** tetap bersih — cuma query database
- **Service** yang decide: "cek cache dulu, kalau miss baru ke repo"

Ini bikin caching bisa di-toggle tanpa ubah handler atau repository.

---

## 8.11 Redis CLI Commands Cheat Sheet

Buat debugging dan monitoring, kamu harus familiar sama Redis CLI.

### Connection

```bash
# Connect ke Redis
redis-cli

# Connect ke host/port tertentu
redis-cli -h 192.168.1.100 -p 6380

# Connect dengan password
redis-cli -a yourpassword

# Ping — test koneksi
redis-cli ping
# PONG
```

### Key Operations

```bash
# Lihat semua key (JANGAN pakai di production!)
KEYS *

# Lihat key dengan pattern
KEYS board:*
KEYS rate_limit:*

# Cek key ada atau tidak
EXISTS board:5

# Hapus key
DEL board:5

# Hapus banyak key sekaligus
DEL board:1 board:2 board:3

# Cek tipe data key
TYPE board:5

# Set TTL
EXPIRE board:5 300

# Cek sisa TTL
TTL board:5
# (integer) 287 — sisa 287 detik
# (integer) -1  — no expiry (permanent)
# (integer) -2  — key ga ada

# Hapus TTL (bikin permanent)
PERSIST board:5

# Scan — alternative KEYS yang production-safe
SCAN 0 MATCH board:* COUNT 100
```

### String Commands

```bash
SET mykey "hello"
GET mykey
INCR counter
DECR counter
INCRBY counter 10
APPEND mykey " world"     # "hello world"
STRLEN mykey              # 11
MSET key1 "a" key2 "b"   # Set multiple
MGET key1 key2            # Get multiple
SETNX mykey "val"         # Set only if Not eXists
```

### Hash Commands

```bash
HSET user:1 name "Irzan" email "a@b.com"
HGET user:1 name
HGETALL user:1
HDEL user:1 email
HEXISTS user:1 name       # 1 atau 0
HLEN user:1               # jumlah field
HKEYS user:1              # semua field names
HVALS user:1              # semua values
```

### List Commands

```bash
LPUSH mylist "a" "b"      # Push ke depan
RPUSH mylist "c"           # Push ke belakang
LPOP mylist                # Pop dari depan
RPOP mylist                # Pop dari belakang
LRANGE mylist 0 -1         # Semua element
LLEN mylist                # Panjang list
LINDEX mylist 0            # Element di index 0
```

### Set Commands

```bash
SADD myset "a" "b" "c"
SMEMBERS myset
SISMEMBER myset "a"       # 1 (true)
SREM myset "a"            # Remove
SCARD myset                # Count
SINTER set1 set2           # Intersection
SUNION set1 set2           # Union
SDIFF set1 set2            # Difference
```

### Sorted Set Commands

```bash
ZADD leaderboard 100 "user:1" 200 "user:2"
ZRANGE leaderboard 0 -1 WITHSCORES     # Low to high
ZREVRANGE leaderboard 0 -1 WITHSCORES  # High to low
ZSCORE leaderboard "user:1"             # Score of member
ZRANK leaderboard "user:1"              # Rank (0-based)
ZREM leaderboard "user:1"               # Remove
ZCARD leaderboard                        # Count
ZINCRBY leaderboard 50 "user:1"         # Increment score
```

### Utility

```bash
# Database size (jumlah keys)
DBSIZE

# Flush current DB (HATI-HATI!)
FLUSHDB

# Flush ALL database (SANGAT HATI-HATI!)
FLUSHALL

# Info server
INFO

# Info memory aja
INFO memory

# Rename key
RENAME oldkey newkey

# Select database (0-15)
SELECT 1
```

---

## 8.12 Monitoring Redis

### INFO Command

```bash
redis-cli INFO
```

Output paling penting:

```
# Memory
used_memory_human:1.23M        ← Total memory dipakai
used_memory_peak_human:2.45M   ← Peak memory

# Stats
total_connections_received:1234 ← Total koneksi sejak start
total_commands_processed:56789  ← Total commands
keyspace_hits:45000             ← Cache HIT
keyspace_misses:5000            ← Cache MISS

# Keyspace
db0:keys=150,expires=130        ← 150 keys, 130 punya TTL
```

### Hit Rate Calculation

```bash
# Hitung cache hit rate
redis-cli INFO stats | grep keyspace

# Hit rate = hits / (hits + misses) * 100
# 45000 / (45000 + 5000) * 100 = 90% — bagus!
```

Target hit rate yang bagus: **> 80%**. Kalau di bawah itu, kemungkinan:
- TTL terlalu pendek
- Cache key design kurang tepat
- Data terlalu banyak variasi (contoh: cache per-page tanpa limit)

### MONITOR Command

Real-time log semua command yang masuk ke Redis. **Jangan pakai di production** (berat!), tapi sangat berguna buat debugging.

```bash
redis-cli MONITOR
```

Output:

```
1714300000.123456 [0 127.0.0.1:54321] "GET" "board:5"
1714300000.234567 [0 127.0.0.1:54321] "SET" "board:5" "{\"id\":5,...}" "EX" "300"
1714300000.345678 [0 127.0.0.1:54322] "INCR" "rate_limit:192.168.1.1"
```

Kamu bisa lihat semua operasi Redis real-time — super useful buat verify cache-mu jalan.

### SLOWLOG — Cari command yang lambat

```bash
# Lihat 10 command paling lambat
redis-cli SLOWLOG GET 10

# Reset slowlog
redis-cli SLOWLOG RESET

# Config threshold (default 10ms)
redis-cli CONFIG SET slowlog-log-slower-than 5000
```

### Memory Analysis

```bash
# Memory usage per key
redis-cli MEMORY USAGE board:5
# (integer) 256 — 256 bytes

# Top keys by memory (pakai redis-cli --bigkeys)
redis-cli --bigkeys
```

Output `--bigkeys`:

```
# Scanning the entire keyspace to find biggest keys

[00.00%] Biggest string found so far '"board:5"' with 1234 bytes
[00.00%] Biggest hash   found so far '"session:user:1"' with 5 fields
[00.00%] Biggest list   found so far '"notifications:user:1"' with 25 items

-------- summary -------
Biggest   string found '"board:5"' has 1234 bytes
Biggest     hash found '"session:user:1"' has 5 fields
```

### Monitoring Script Sederhana

```bash
#!/bin/bash
# redis-monitor.sh — jalanin berkala pakai cron

HITS=$(redis-cli INFO stats | grep keyspace_hits | cut -d: -f2 | tr -d '\r')
MISSES=$(redis-cli INFO stats | grep keyspace_misses | cut -d: -f2 | tr -d '\r')
MEMORY=$(redis-cli INFO memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
KEYS=$(redis-cli DBSIZE | awk '{print $2}')

echo "$(date) | Keys: $KEYS | Memory: $MEMORY | Hits: $HITS | Misses: $MISSES"
```

### Health Check di Go

```go
// internal/handler/health.go — tambah Redis health check

func HealthCheck(c *gin.Context) {
    status := gin.H{"status": "ok"}

    // Cek Redis
    ctx := context.Background()
    if err := redisClient.Ping(ctx).Err(); err != nil {
        status["redis"] = "disconnected"
    } else {
        info, _ := redisClient.Info(ctx, "memory", "stats").Result()
        status["redis"] = "connected"
        // Parse info kalau mau detail
        _ = info
    }

    c.JSON(http.StatusOK, status)
}
```

---

## Best Practices & Tips

### 1. Key Naming Convention

Pakai format yang konsisten dan descriptive:

```
# Format: <entity>:<id>
board:5
task:123

# Format: <entity>:<filter>:<value>:<pagination>
boards:owner:1:page:1:limit:10
tasks:board:5:page:1:limit:10

# Format: <feature>:<identifier>
rate_limit:192.168.1.1
session:user:1
blacklist:<token_hash>
```

### 2. Jangan Cache Semuanya

```
CACHE:
✅ Data yang sering dibaca, jarang ditulis (board detail, user profile)
✅ Data yang expensive untuk generate (aggregation, report)
✅ Data yang bisa stale sedikit (dashboard stats)

JANGAN CACHE:
❌ Data yang selalu berubah (real-time stock, live score)
❌ Data yang sangat personal dan sensitif (tanpa encryption)
❌ Data yang jarang diakses (old archived data)
```

### 3. Serialization: JSON vs MessagePack

```go
// JSON — readable, tapi agak besar
data, _ := json.Marshal(board) // ~200 bytes

// MessagePack — lebih kecil, lebih cepat
// go get github.com/vmihailenco/msgpack/v5
data, _ := msgpack.Marshal(board) // ~120 bytes (40% lebih kecil)
```

Untuk project kecil-menengah, JSON cukup. Kalau data besar dan throughput tinggi, pertimbangkan MessagePack.

### 4. Connection Pooling

go-redis sudah handle connection pool secara otomatis. Tapi kamu bisa tune:

```go
client := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,              // Max connections
    MinIdleConns: 5,               // Min idle connections
    DialTimeout:  5 * time.Second, // Connection timeout
    ReadTimeout:  3 * time.Second, // Read timeout
    WriteTimeout: 3 * time.Second, // Write timeout
})
```

### 5. Graceful Shutdown

```go
// cmd/api/main.go

func main() {
    redisClient := config.ConnectRedis(cfg)
    defer redisClient.Close() // Tutup koneksi saat app shutdown

    // ... rest of setup ...
}
```

---

## Mapping NestJS ↔ Go Redis

| NestJS (ioredis/cache-manager) | Go (go-redis) |
|-------------------------------|---------------|
| `@CacheKey('board:5')` decorator | Manual: `cache.Get("board:5", &board)` |
| `@CacheTTL(300)` | `cache.Set(key, val, 5*time.Minute)` |
| `CacheInterceptor` | Middleware atau logic di service |
| `this.redis.get(key)` | `rdb.Get(ctx, key).Result()` |
| `this.redis.set(key, val, 'EX', 300)` | `rdb.Set(ctx, key, val, 300*time.Second)` |
| `@InjectRedis()` | Manual inject `*redis.Client` |
| Bull Queue + Redis | Asynq (Bab 9) |

Di NestJS, caching banyak pake decorator dan interceptor (magic). Di Go, semuanya explicit. Lebih verbose, tapi kamu selalu tau persis apa yang terjadi.

---

## Checklist Bab Ini

- [ ] Paham apa itu Redis dan kapan pakainya
- [ ] Paham 5 tipe data Redis (string, hash, list, set, sorted set)
- [ ] Redis terinstall dan bisa `redis-cli ping`
- [ ] Paham file `internal/config/redis.go` — connection setup
- [ ] Paham file `internal/middleware/ratelimit.go` — rate limiting
- [ ] Paham 3 caching pattern (cache-aside, write-through, write-behind)
- [ ] Bisa implementasi cache-aside di service layer
- [ ] Paham cache invalidation (TTL + manual)
- [ ] Paham Redis pub/sub untuk real-time features
- [ ] Paham session storage dan token blacklist
- [ ] Familiar sama Redis CLI commands
- [ ] Bisa monitoring Redis (INFO, MONITOR, SLOWLOG)
- [ ] Baca redis.io/docs — self-study
