# Bab 5: REST API dengan Gin Framework

## Gambaran Besar

Di Bab 3 kita udah setup project dan paham folder structure. Sekarang kita masuk ke inti backend development: **bikin REST API pakai Gin**.

Kalau kamu dari NestJS, ini equivalent-nya bikin Controller, pasang decorator, setup routing, dan return response. Bedanya: di Go, semuanya **explicit** — ga ada magic, ga ada decorator, ga ada auto-inject. Tapi justru itu yang bikin kamu 100% paham apa yang terjadi.

Di bab ini, semua contoh kode diambil langsung dari project Task Manager API kita. Bukan kode dummy — ini kode yang beneran jalan.

---

## 1. Gin Framework — Kenapa Gin?

### Apa itu Gin?

Gin adalah HTTP web framework untuk Go yang fokus di **performance** dan **minimalism**. Kalau NestJS itu "batteries included" (bawa semuanya), Gin itu cuma kasih kamu router + middleware chain + JSON binding. Sisanya kamu pilih sendiri.

### Perbandingan Framework Go

| | Gin | Echo | Fiber |
|---|-----|------|-------|
| GitHub Stars | ~80k | ~30k | ~35k |
| Basis | `net/http` | `net/http` | **fasthttp** |
| Performance | Sangat cepat | Cepat | Paling cepat (benchmark) |
| Ecosystem | **Paling besar** | Besar | Growing |
| Middleware | Banyak banget | Banyak | Banyak |
| API Style | `c *gin.Context` | `c echo.Context` | `c *fiber.Ctx` |
| Kompatibilitas stdlib | **Penuh** | Penuh | Terbatas (fasthttp) |
| Learning resource | **Paling banyak** | Banyak | Lumayan |

### Kenapa kita pilih Gin?

1. **Komunitas terbesar** — kalau stuck, gampang cari jawaban
2. **Kompatibel penuh dengan `net/http`** — semua library Go standard bisa dipake tanpa adapter
3. **Battle-tested** — dipake di production oleh banyak company
4. **Simple API** — ga over-engineered, sesuai filosofi Go
5. **Fiber** cepat tapi berbasis `fasthttp`, bukan `net/http`. Ini berarti banyak library Go yang ga langsung kompatibel

### Perbandingan dengan NestJS

| Aspek | NestJS | Gin |
|-------|--------|-----|
| Framework type | Full framework (opinionated) | Micro framework (un-opinionated) |
| Routing | Decorator-based (`@Get()`) | Function-based (`r.GET()`) |
| DI | Built-in container | Manual |
| Validation | class-validator + decorator | Binding tags di struct |
| Middleware | `@UseGuards()`, `@UseInterceptors()` | `r.Use(middleware)` |
| Module system | `@Module()` | Package-based |

**Intinya:** NestJS = Rails-like (convention over configuration). Gin = Express-like (minimal, kamu yang compose).

---

## 2. Hello World — Server Pertama

Sebelum bahas project kita, ini Gin server paling minimal:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.Default() // bikin engine dengan Logger + Recovery middleware

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "pong",
        })
    })

    r.Run(":8080") // listen di port 8080
}
```

**NestJS equivalent:**
```typescript
@Controller()
export class AppController {
  @Get('/ping')
  ping() {
    return { message: 'pong' };
  }
}
```

Perbedaan mendasar:
- NestJS: decorator `@Get('/ping')` di atas method class
- Gin: langsung register `r.GET("/ping", handlerFunc)` — route dan handler di satu tempat

`gin.H` itu shortcut untuk `map[string]interface{}`. Mirip object literal di JS.

`gin.Default()` vs `gin.New()`:
- `gin.Default()` = `gin.New()` + Logger middleware + Recovery middleware (panic handler)
- Di development pake `Default()`, di production bisa pake `New()` dan pasang middleware sendiri

---

## 3. Routing

### 3.1 Basic Route

Gin support semua HTTP method:

```go
r.GET("/resource", handler)
r.POST("/resource", handler)
r.PUT("/resource/:id", handler)
r.DELETE("/resource/:id", handler)
r.PATCH("/resource/:id", handler)
```

### 3.2 Route Group

Route group = cara kita mengelompokkan route yang punya prefix atau middleware yang sama.

Ini **langsung dari project kita** (`internal/router/router.go`):

```go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/internal/handler"
    "github.com/irzan-aldi/task-manager/internal/middleware"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/internal/service"
    "github.com/redis/go-redis/v9"
    "time"
)

type Handlers struct {
    Auth  *handler.AuthHandler
    User  *handler.UserHandler
    Board *handler.BoardHandler
    Task  *handler.TaskHandler
}

func Setup(r *gin.Engine, h *Handlers, authService *service.AuthService, redisClient *redis.Client) {
    // Health check (public)
    r.GET("/health", handler.HealthCheck)

    // Rate limiter: 100 requests per minute per IP
    r.Use(middleware.RateLimiter(redisClient, 100, time.Minute))

    v1 := r.Group("/api/v1")
    {
        // Public routes
        auth := v1.Group("/auth")
        {
            auth.POST("/register", h.Auth.Register)
            auth.POST("/login", h.Auth.Login)
            auth.POST("/refresh", h.Auth.RefreshToken)
        }

        // Protected routes
        protected := v1.Group("")
        protected.Use(middleware.AuthMiddleware(authService))
        {
            // Auth
            protected.GET("/auth/me", h.Auth.Me)

            // Users
            protected.GET("/users/profile", h.User.GetProfile)
            protected.PUT("/users/profile", h.User.Update)

            // Users — admin only
            admin := protected.Group("")
            admin.Use(middleware.RoleMiddleware(model.RoleAdmin))
            {
                admin.GET("/users", h.User.GetAll)
                admin.DELETE("/users/:id", h.User.Delete)
            }

            // Boards
            protected.POST("/boards", h.Board.Create)
            protected.GET("/boards", h.Board.GetMyBoards)
            protected.GET("/boards/:id", h.Board.GetByID)
            protected.PUT("/boards/:id", h.Board.Update)
            protected.DELETE("/boards/:id", h.Board.Delete)

            // Tasks
            protected.POST("/boards/:boardId/tasks", h.Task.Create)
            protected.GET("/boards/:boardId/tasks", h.Task.GetByBoardID)
            protected.GET("/tasks/my", h.Task.GetMyTasks)
            protected.GET("/tasks/:id", h.Task.GetByID)
            protected.PUT("/tasks/:id", h.Task.Update)
            protected.DELETE("/tasks/:id", h.Task.Delete)
        }
    }
}
```

> File: `internal/router/router.go`

**Breakdown:**

1. **`r.GET("/health", ...)`** — route public di root, ga masuk group manapun
2. **`r.Use(middleware.RateLimiter(...))`** — middleware global, semua request kena
3. **`v1 := r.Group("/api/v1")`** — group dengan prefix `/api/v1`
4. **`auth := v1.Group("/auth")`** — nested group, jadinya `/api/v1/auth/register`, etc.
5. **`protected := v1.Group("")`** — group tanpa prefix tambahan, tapi pakai auth middleware
6. **`admin := protected.Group("")`** — nested lagi, nambah role middleware

**Visualisasi hierarki middleware:**

```
Request
 ├── /health → HealthCheck (no middleware)
 │
 └── /api/v1/* → RateLimiter
      ├── /auth/register → public
      ├── /auth/login → public
      │
      └── /* → AuthMiddleware (JWT check)
           ├── /boards → CRUD (authenticated user)
           ├── /tasks → CRUD (authenticated user)
           │
           └── /users → RoleMiddleware(admin)
                ├── GET /users → admin only
                └── DELETE /users/:id → admin only
```

**NestJS equivalent:**
```typescript
// NestJS pakai module + decorator
@Module({
  controllers: [BoardController],
})
export class BoardModule {}

// Route prefix via decorator
@Controller('api/v1/boards')
@UseGuards(AuthGuard)
export class BoardController {
  @Get()
  getAll() { ... }
}
```

Di Gin, kamu bisa "lihat" semua route di satu file. Di NestJS, route tersebar di banyak file Controller + Module. Ini trade-off: NestJS lebih modular, Gin lebih overview-friendly.

### 3.3 URL Params

URL params = bagian dinamis di URL path. Di Gin pake `:namaParam`.

```go
// Register route dengan param
protected.GET("/boards/:id", h.Board.GetByID)
protected.POST("/boards/:boardId/tasks", h.Task.Create)
```

Ambil nilainya di handler:

```go
func (h *BoardHandler) GetByID(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid board ID")
        return
    }
    // id sekarang bertipe uint64
}
```

**Penting:** `c.Param()` selalu return `string`. Kamu harus convert sendiri ke tipe yang dibutuhkan. Di NestJS, Pipe (`ParseIntPipe`) yang handle ini. Di Go, kamu handle manual — lebih verbose tapi lebih kontrol.

**NestJS equivalent:**
```typescript
@Get(':id')
getById(@Param('id', ParseIntPipe) id: number) {
  return this.boardService.getById(id);
}
```

### 3.4 Query Params

Query params = parameter di URL setelah `?`. Biasa dipake untuk filtering, pagination, sorting.

Dari `handler/board.go` — method `GetMyBoards`:

```go
func (h *BoardHandler) GetMyBoards(c *gin.Context) {
    userID := c.GetUint("userID")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }

    boards, total, err := h.boardService.GetByOwnerID(userID, page, limit)
    // ...
}
```

Method query yang tersedia:

```go
c.Query("page")              // return "" kalau ga ada
c.DefaultQuery("page", "1")  // return "1" kalau ga ada
c.QueryArray("tags")         // return []string untuk ?tags=a&tags=b
c.QueryMap("filter")         // return map[string]string untuk ?filter[name]=x
```

**NestJS equivalent:**
```typescript
@Get()
getMyBoards(
  @Query('page') page: string = '1',
  @Query('limit') limit: string = '10',
) { ... }

// Atau pakai DTO
@Get()
getMyBoards(@Query() query: PaginationDto) { ... }
```

---

## 4. Request Handling

### 4.1 Bind JSON (Request Body)

Ini yang paling sering kamu pake. Client kirim JSON body, kita parse ke struct.

Pertama, define struct request-nya (dari `internal/model/board.go`):

```go
type CreateBoardRequest struct {
    Title       string `json:"title" binding:"required,min=1,max=200"`
    Description string `json:"description" binding:"max=500"`
}

type UpdateBoardRequest struct {
    Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
    Description *string `json:"description" binding:"omitempty,max=500"`
}
```

Terus di handler, bind-nya:

```go
func (h *BoardHandler) Create(c *gin.Context) {
    userID := c.GetUint("userID")
    var req model.CreateBoardRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    board, err := h.boardService.Create(userID, req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, http.StatusCreated, "Board created", board)
}
```

> File: `internal/handler/board.go`

**`ShouldBindJSON` vs `BindJSON`:**
- `ShouldBindJSON` — return error, kamu handle sendiri. **Selalu pake ini.**
- `BindJSON` — auto-respond 400 kalau error. Masalahnya, format response-nya ga bisa kamu kontrol.

**Pattern Update dengan Pointer:**

Perhatikan `UpdateBoardRequest` pake pointer `*string`. Kenapa?

```go
type UpdateBoardRequest struct {
    Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
    Description *string `json:"description" binding:"omitempty,max=500"`
}
```

Karena kita mau bisa bedain antara:
- Client ga kirim field `title` → `Title` = `nil` → jangan update
- Client kirim `"title": ""` → `Title` = pointer ke empty string → update ke kosong

Kalau pake `string` biasa, dua-duanya jadi `""` dan ga bisa dibedain. Ini pattern umum untuk PATCH/PUT partial update.

### 4.2 Binding Tags (Validation)

Gin pake library `go-playground/validator` di balik layar. Validasi ditulis sebagai struct tag `binding:"..."`.

Ini contoh lengkap dari model-model di project kita:

```go
// Register — semua field required
type RegisterRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// Login — email + password required
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// Create Task — title required, priority harus salah satu dari enum
type CreateTaskRequest struct {
    Title       string     `json:"title" binding:"required,min=1,max=200"`
    Description string     `json:"description"`
    Priority    Priority   `json:"priority" binding:"omitempty,oneof=low medium high"`
    AssigneeID  *uint      `json:"assignee_id"`
    DueDate     *time.Time `json:"due_date"`
}

// Update Task — semua optional, tapi kalau diisi harus valid
type UpdateTaskRequest struct {
    Title       *string     `json:"title" binding:"omitempty,min=1,max=200"`
    Description *string     `json:"description"`
    Status      *TaskStatus `json:"status" binding:"omitempty,oneof=todo in_progress done"`
    Priority    *Priority   `json:"priority" binding:"omitempty,oneof=low medium high"`
    AssigneeID  *uint       `json:"assignee_id"`
    DueDate     *time.Time  `json:"due_date"`
}
```

> File: `internal/model/user.go`, `internal/model/task.go`

**Binding tags yang sering dipake:**

| Tag | Fungsi | Contoh |
|-----|--------|--------|
| `required` | Wajib diisi | `binding:"required"` |
| `email` | Format email valid | `binding:"required,email"` |
| `min` | Minimal length/value | `binding:"min=2"` |
| `max` | Maksimal length/value | `binding:"max=200"` |
| `oneof` | Harus salah satu dari nilai | `binding:"oneof=low medium high"` |
| `omitempty` | Skip validasi kalau kosong | `binding:"omitempty,min=1"` |
| `len` | Exact length | `binding:"len=6"` |
| `numeric` | Harus numeric string | `binding:"numeric"` |
| `url` | Format URL valid | `binding:"url"` |
| `uuid` | Format UUID valid | `binding:"uuid"` |

**NestJS equivalent:**
```typescript
// NestJS pakai class-validator decorator
export class CreateBoardDto {
  @IsString()
  @MinLength(1)
  @MaxLength(200)
  title: string;

  @IsOptional()
  @MaxLength(500)
  description?: string;
}
```

Konsepnya sama — validasi di level DTO/request struct. Bedanya, NestJS pakai decorator di atas property, Go pakai struct tag di samping field. Efeknya sama.

### 4.3 Bind Form Data & File Upload

Selain JSON, Gin juga bisa bind form data:

```go
// Bind form-urlencoded
type LoginForm struct {
    Username string `form:"username" binding:"required"`
    Password string `form:"password" binding:"required"`
}

func Login(c *gin.Context) {
    var form LoginForm
    if err := c.ShouldBind(&form); err != nil { // auto-detect Content-Type
        // handle error
    }
}

// Upload file
func Upload(c *gin.Context) {
    file, err := c.FormFile("avatar")
    if err != nil {
        // handle error
    }
    c.SaveUploadedFile(file, "/uploads/" + file.Filename)
}
```

`ShouldBind` (tanpa JSON) otomatis detect Content-Type: kalau `application/json` → bind JSON, kalau `multipart/form-data` → bind form.

### 4.4 Ambil Data dari Context

Di project kita, setelah middleware auth berhasil, data user disimpan di `gin.Context`:

```go
// Di middleware (set data)
c.Set("userID", claims.UserID)
c.Set("email", claims.Email)
c.Set("role", claims.Role)

// Di handler (get data)
userID := c.GetUint("userID")  // langsung ke uint
email, _ := c.Get("email")     // return interface{}, harus type assert
role, _ := c.Get("role")       // return interface{}, harus type assert
```

Ini mirip `request.user` di NestJS setelah AuthGuard jalan. Bedanya, di Go kamu set dan get secara manual — bukan dari decorator `@Req()`.

---

## 5. Response

### 5.1 JSON Response Dasar

Cara paling basic kirim JSON response di Gin:

```go
c.JSON(http.StatusOK, gin.H{
    "status": "ok",
})
```

Dari project kita — `handler/health.go` (handler paling simple):

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
    })
}

func Ping(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "pong",
    })
}
```

> File: `internal/handler/health.go`

Ini handler paling sederhana — ga ada dependency, ga ada struct, langsung function biasa. Cocok buat health check endpoint.

**NestJS equivalent:**
```typescript
@Controller()
export class HealthController {
  @Get('health')
  check() {
    return { status: 'ok' };
  }
}
```

### 5.2 Standard Response Format

Di production, kamu butuh format response yang **konsisten**. Jangan sampe endpoint A return `{ data: ... }` dan endpoint B return `{ result: ... }`.

Ini standard response package dari project kita (`pkg/response/json.go`):

```go
package response

import "github.com/gin-gonic/gin"

type Response struct {
    Status  int         `json:"status"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, status int, message string, data interface{}) {
    c.JSON(status, Response{
        Status:  status,
        Message: message,
        Data:    data,
    })
}

func Error(c *gin.Context, status int, message string) {
    c.JSON(status, Response{
        Status:  status,
        Message: message,
    })
}
```

> File: `pkg/response/json.go`

**Kenapa dipisah ke `pkg/response`?**
- Semua handler pake format yang sama
- Kalau mau ubah format (misalnya tambah field `timestamp`), ubah di satu tempat
- `pkg/` artinya bisa di-reuse di project lain

**Cara pakainya di handler:**

```go
// Sukses — kirim data
response.Success(c, http.StatusCreated, "Board created", board)

// Sukses — tanpa data
response.Success(c, http.StatusOK, "Board deleted", nil)

// Error
response.Error(c, http.StatusBadRequest, "Invalid board ID")
response.Error(c, http.StatusUnauthorized, "Invalid or expired token")
```

**Contoh output JSON:**

```json
// Success
{
  "status": 201,
  "message": "Board created",
  "data": {
    "id": 1,
    "title": "Sprint 1",
    "description": "First sprint board",
    "owner_id": 5,
    "created_at": "2026-04-28T10:00:00Z"
  }
}

// Error
{
  "status": 400,
  "message": "Invalid board ID"
}
```

Perhatikan: kalau `data` nil, field-nya ga muncul di JSON karena tag `omitempty`. Clean.

**NestJS equivalent:**
```typescript
// Di NestJS biasanya bikin interceptor untuk wrap response
@Injectable()
export class TransformInterceptor implements NestInterceptor {
  intercept(context, next) {
    return next.handle().pipe(
      map(data => ({
        status: context.switchToHttp().getResponse().statusCode,
        message: 'Success',
        data,
      }))
    );
  }
}
```

Di NestJS kamu perlu interceptor (cukup complex). Di Go, cukup helper function biasa. Simpel.

### 5.3 HTTP Status Codes yang Sering Dipake

Go punya constant untuk semua HTTP status code di package `net/http`:

```go
http.StatusOK                  // 200 — berhasil
http.StatusCreated             // 201 — resource berhasil dibuat
http.StatusBadRequest          // 400 — input salah
http.StatusUnauthorized        // 401 — belum login / token expired
http.StatusForbidden           // 403 — ga punya akses
http.StatusNotFound            // 404 — resource ga ketemu
http.StatusConflict            // 409 — duplicate (email udah terdaftar)
http.StatusTooManyRequests     // 429 — rate limited
http.StatusInternalServerError // 500 — server error
```

Di project kita, handler pake status code yang tepat sesuai konteks:

```go
// 201 — baru dibuat
response.Success(c, http.StatusCreated, "Board created", board)

// 200 — berhasil fetch / update / delete
response.Success(c, http.StatusOK, "Board detail", board)

// 400 — input jelek
response.Error(c, http.StatusBadRequest, err.Error())

// 404 — ga ketemu
response.Error(c, http.StatusNotFound, err.Error())

// 403 — ga boleh (bukan punya dia)
response.Error(c, http.StatusForbidden, err.Error())
```

---

## 6. Middleware

### 6.1 Konsep Middleware di Gin

Middleware di Gin = function yang jalan **sebelum** (dan/atau **sesudah**) handler utama. Dipake buat logging, auth, rate limiting, CORS, dll.

```
Request → Middleware 1 → Middleware 2 → Handler → Middleware 2 (after) → Middleware 1 (after) → Response
```

Setiap middleware bisa:
- **`c.Next()`** — lanjut ke middleware/handler berikutnya
- **`c.Abort()`** — stop chain, ga lanjut ke handler

Signature middleware Gin:

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // === SEBELUM handler ===
        // Logic sebelum handler jalan (auth check, logging start, dll)

        c.Next() // lanjut ke handler

        // === SESUDAH handler ===
        // Logic setelah handler selesai (logging end, dll)
    }
}
```

### 6.2 Auth Middleware (dari Project)

Ini middleware auth yang kita pake beneran:

```go
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Error(c, http.StatusUnauthorized, "Authorization header required")
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Error(c, http.StatusUnauthorized, "Invalid authorization format. Use: Bearer <token>")
            c.Abort()
            return
        }

        claims, err := authService.ValidateToken(parts[1])
        if err != nil {
            response.Error(c, http.StatusUnauthorized, "Invalid or expired token")
            c.Abort()
            return
        }

        // Set user info in context
        c.Set("userID", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

> File: `internal/middleware/auth.go`

**Flow:**
1. Ambil header `Authorization`
2. Kalau kosong → `c.Abort()` + return 401. Handler **ga akan jalan**.
3. Parse format `Bearer <token>`
4. Validate token via `authService`
5. Kalau valid → simpan data user di context (`c.Set`) → `c.Next()` → lanjut ke handler

**`c.Abort()` vs just `return`:**
- `return` aja **ga cukup**. Kamu harus panggil `c.Abort()` untuk kasih tau Gin supaya ga execute middleware/handler berikutnya di chain.
- `c.Abort()` + `return` = pattern standard.

### 6.3 Role Middleware (Authorization)

Setelah auth middleware verify user, role middleware check apakah user punya role yang tepat:

```go
func RoleMiddleware(roles ...model.Role) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("role")
        if !exists {
            response.Error(c, http.StatusUnauthorized, "User role not found")
            c.Abort()
            return
        }

        role := userRole.(model.Role)
        for _, r := range roles {
            if role == r {
                c.Next()
                return
            }
        }

        response.Error(c, http.StatusForbidden, "Insufficient permissions")
        c.Abort()
    }
}
```

> File: `internal/middleware/auth.go`

Dipake di router:

```go
admin := protected.Group("")
admin.Use(middleware.RoleMiddleware(model.RoleAdmin))
{
    admin.GET("/users", h.User.GetAll)
    admin.DELETE("/users/:id", h.User.Delete)
}
```

Jadi route `/api/v1/users` harus lewat: RateLimiter -> AuthMiddleware -> RoleMiddleware(admin) -> Handler. Tiga layer middleware.

### 6.4 Rate Limiter Middleware

```go
func RateLimiter(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        ip := c.ClientIP()
        key := fmt.Sprintf("rate_limit:%s", ip)

        count, err := redisClient.Incr(ctx, key).Result()
        if err != nil {
            // Redis down — allow request
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

> File: `internal/middleware/ratelimit.go`

Perhatikan design yang bagus: kalau Redis down, request tetap diloloskan (`c.Next()`). Ini prinsip **fail open** — lebih baik rate limiter ga jalan daripada semua request ditolak gara-gara Redis mati.

### 6.5 Cara Pasang Middleware

```go
// Global — semua route kena
r.Use(middleware.RateLimiter(redisClient, 100, time.Minute))

// Per group — hanya route dalam group ini
protected := v1.Group("")
protected.Use(middleware.AuthMiddleware(authService))

// Per route — hanya route ini
r.GET("/admin/stats", middleware.AuthMiddleware(authService), adminStatsHandler)
```

**NestJS equivalent:**

```typescript
// Global
app.useGlobalGuards(new RateLimitGuard());

// Per controller
@UseGuards(AuthGuard)
@Controller('boards')
export class BoardController {}

// Per route
@UseGuards(RolesGuard)
@Get()
getAll() {}
```

| NestJS | Gin |
|--------|-----|
| `@UseGuards(AuthGuard)` | `group.Use(middleware.AuthMiddleware(...))` |
| `@UseInterceptors(LoggingInterceptor)` | `r.Use(loggingMiddleware)` |
| `@UsePipes(ValidationPipe)` | Built-in via `ShouldBind` |
| Guard return boolean | Middleware call `c.Next()` or `c.Abort()` |

---

## 7. Handler Pattern — Struct-based dengan Dependency Injection

### 7.1 Simple Handler (Function)

Untuk endpoint yang ga butuh dependency (health check, ping), function biasa cukup:

```go
func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
    })
}
```

> File: `internal/handler/health.go`

### 7.2 Struct-based Handler (Dependency Injection)

Untuk handler yang butuh akses ke service layer, kita pake struct + method:

```go
type BoardHandler struct {
    boardService *service.BoardService
}

func NewBoardHandler(boardService *service.BoardService) *BoardHandler {
    return &BoardHandler{boardService: boardService}
}

func (h *BoardHandler) Create(c *gin.Context) {
    // h.boardService sudah tersedia
}

func (h *BoardHandler) GetByID(c *gin.Context) {
    // h.boardService sudah tersedia
}
```

> File: `internal/handler/board.go`

**Ini pattern DI manual di Go:**
1. **Struct** = container yang hold dependencies
2. **Constructor `NewXxxHandler()`** = terima dependencies sebagai parameter
3. **Method receiver `(h *BoardHandler)`** = setiap method bisa akses `h.boardService`

### 7.3 Full CRUD Handler (Board)

Ini handler lengkap untuk Board CRUD — langsung dari project:

```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/internal/service"
    "github.com/irzan-aldi/task-manager/pkg/response"
)

type BoardHandler struct {
    boardService *service.BoardService
}

func NewBoardHandler(boardService *service.BoardService) *BoardHandler {
    return &BoardHandler{boardService: boardService}
}

func (h *BoardHandler) Create(c *gin.Context) {
    userID := c.GetUint("userID")
    var req model.CreateBoardRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    board, err := h.boardService.Create(userID, req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, http.StatusCreated, "Board created", board)
}

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

func (h *BoardHandler) GetMyBoards(c *gin.Context) {
    userID := c.GetUint("userID")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }

    boards, total, err := h.boardService.GetByOwnerID(userID, page, limit)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to fetch boards")
        return
    }

    response.Success(c, http.StatusOK, "My boards", gin.H{
        "boards": boards,
        "total":  total,
        "page":   page,
        "limit":  limit,
    })
}

func (h *BoardHandler) Update(c *gin.Context) {
    userID := c.GetUint("userID")
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid board ID")
        return
    }

    var req model.UpdateBoardRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    board, err := h.boardService.Update(uint(id), userID, req)
    if err != nil {
        response.Error(c, http.StatusForbidden, err.Error())
        return
    }

    response.Success(c, http.StatusOK, "Board updated", board)
}

func (h *BoardHandler) Delete(c *gin.Context) {
    userID := c.GetUint("userID")
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid board ID")
        return
    }

    if err := h.boardService.Delete(uint(id), userID); err != nil {
        response.Error(c, http.StatusForbidden, err.Error())
        return
    }

    response.Success(c, http.StatusOK, "Board deleted", nil)
}
```

> File: `internal/handler/board.go`

**Pattern yang konsisten di setiap handler method:**

```
1. Ambil data dari context (userID, dll)
2. Parse URL params / query params
3. Bind request body (kalau POST/PUT)
4. Panggil service layer
5. Handle error → response.Error()
6. Success → response.Success()
```

Ini **sangat mirip** dengan pattern NestJS Controller:

```typescript
@Controller('boards')
export class BoardController {
  constructor(private boardService: BoardService) {} // Step: DI

  @Post()
  async create(@Body() dto: CreateBoardDto, @Req() req) { // Step: bind body
    return this.boardService.create(req.user.id, dto);      // Step: call service
  }
}
```

Bedanya cuma:
- NestJS: decorator handle routing + binding + validation
- Go: kamu tulis manual, tapi flow-nya identik

### 7.4 Wiring di main.go

Semua dependency di-wire di `cmd/api/main.go`:

```go
func main() {
    cfg := config.Load()
    db := config.ConnectDB(cfg)
    redisClient := config.ConnectRedis(cfg)

    // Repositories
    userRepo := repository.NewUserRepository(db)
    boardRepo := repository.NewBoardRepository(db)
    taskRepo := repository.NewTaskRepository(db)

    // Services
    authService := service.NewAuthService(userRepo, cfg.JWTSecret)
    userService := service.NewUserService(userRepo)
    boardService := service.NewBoardService(boardRepo)
    taskService := service.NewTaskService(taskRepo, boardRepo)

    // Handlers
    handlers := &router.Handlers{
        Auth:  handler.NewAuthHandler(authService),
        User:  handler.NewUserHandler(userService),
        Board: handler.NewBoardHandler(boardService),
        Task:  handler.NewTaskHandler(taskService),
    }

    // Gin Engine
    if cfg.Env == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
    r := gin.Default()
    router.Setup(r, handlers, authService, redisClient)

    // Start
    port := cfg.Port
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
```

> File: `cmd/api/main.go`

**Dependency graph-nya:**

```
Database ─────┬── UserRepo ───┬── AuthService ──── AuthHandler
              │               └── UserService ──── UserHandler
              ├── BoardRepo ──── BoardService ──── BoardHandler
              └── TaskRepo ─┬── TaskService ───── TaskHandler
                            │
              BoardRepo ────┘
```

Semua dependency explicit. Kamu bisa trace dari handler ke database cuma dengan baca kode, tanpa framework magic.

**NestJS equivalent:**
```typescript
// NestJS auto-resolve dependency graph via @Module() + @Injectable()
@Module({
  imports: [TypeOrmModule.forFeature([Board])],
  controllers: [BoardController],
  providers: [BoardService, BoardRepository],
})
export class BoardModule {}
```

Di NestJS, dependency resolution otomatis. Di Go, kamu wire manual di `main.go`. Trade-off: NestJS lebih convenient, Go lebih transparent.

---

## 8. Mapping Konsep: NestJS vs Gin (Rangkuman)

Ini tabel mapping lengkap. Bookmark halaman ini.

| NestJS Concept | Gin/Go Equivalent | Contoh di Project |
|---|---|---|
| `@Controller('boards')` | `r.Group("/boards")` | `router.go` |
| `@Get()`, `@Post()` | `r.GET()`, `r.POST()` | `router.go` |
| `@Param('id')` | `c.Param("id")` | `board.go` GetByID |
| `@Query('page')` | `c.Query("page")` / `c.DefaultQuery()` | `board.go` GetMyBoards |
| `@Body() dto` | `c.ShouldBindJSON(&req)` | `board.go` Create |
| `@Req() req.user` | `c.GetUint("userID")` | `board.go` Create |
| Class DTO + class-validator | Struct + binding tags | `model/board.go` |
| `@Injectable() Service` | Struct + `NewXxxService()` | `service/board.go` |
| `constructor(private svc)` | `BoardHandler{boardService}` | `handler/board.go` |
| `@UseGuards(AuthGuard)` | `group.Use(AuthMiddleware())` | `router.go` |
| `@UseGuards(RolesGuard)` | `group.Use(RoleMiddleware())` | `router.go` |
| `@UseInterceptors()` | Custom middleware | `middleware/` |
| `@Module()` | Package + manual wiring | `main.go` |
| Global pipe (validation) | `ShouldBindJSON` per handler | `handler/*.go` |
| `throw HttpException` | `response.Error() + return` | `handler/*.go` |
| Interceptor (transform) | `pkg/response` helper | `pkg/response/json.go` |
| NestJS Middleware | Gin Middleware (`gin.HandlerFunc`) | `middleware/*.go` |

### Mental Model Shift

```
NestJS Brain:
  "Dimana decorator-nya? Dimana module-nya? DI container auto-resolve kan?"

Go Brain:
  "Function terima parameter, return result. Error di-return, bukan di-throw.
   Dependency di-inject lewat constructor. Semua explicit."
```

Kalau kamu udah biasa di NestJS, perasaan pertama belajar Go itu:
- "Kok banyak banget boilerplate?"
- "Kok harus nulis `if err != nil` terus?"

Tapi setelah beberapa minggu:
- "Gue bisa trace bug tanpa debugger — tinggal baca kode"
- "Ga ada hidden magic. Kalau ada bug, pasti di kode gue sendiri"

---

## 9. Tips Practical

### 9.1 Gin Mode

```go
// Development — ada log warna-warni, detailed error
gin.SetMode(gin.DebugMode) // default

// Production — minimal log, no debug info
gin.SetMode(gin.ReleaseMode)

// Testing
gin.SetMode(gin.TestMode)
```

Atau via environment variable:
```bash
export GIN_MODE=release
```

### 9.2 Gin Recovery Middleware

`gin.Default()` sudah include Recovery middleware. Ini catch panic dan return 500 instead of crashing server:

```go
r := gin.Default() // includes Logger + Recovery

// Kalau mau custom:
r := gin.New()
r.Use(gin.Logger())
r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
    response.Error(c, http.StatusInternalServerError, "Internal server error")
}))
```

### 9.3 Inline Struct untuk One-off Request

Kalau request body-nya cuma dipake di satu handler, ga perlu bikin model terpisah:

```go
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req struct {
        RefreshToken string `json:"refresh_token" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }
    // use req.RefreshToken
}
```

> File: `internal/handler/auth.go` — RefreshToken method

Ini pattern yang dipake di project kita. Kalau struct-nya kecil dan cuma dipake sekali, inline aja. Ga perlu file model terpisah.

### 9.4 Pattern Pagination yang Reusable

Di project kita, pattern pagination ini muncul di banyak handler:

```go
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

if page < 1 {
    page = 1
}
if limit < 1 || limit > 100 {
    limit = 10
}
```

Kalau mau DRY, bisa extract ke helper:

```go
// pkg/pagination/pagination.go
func ParsePagination(c *gin.Context) (page, limit int) {
    page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }
    return page, limit
}

// Pake di handler:
page, limit := pagination.ParsePagination(c)
```

---

## 10. File-file Project yang Dibahas di Bab Ini

Quick reference ke semua file yang relevan:

| File | Fungsi |
|------|--------|
| `cmd/api/main.go` | Entry point, wiring semua dependency |
| `internal/router/router.go` | Route registration + middleware setup |
| `internal/handler/health.go` | Simple handler (function-based) |
| `internal/handler/board.go` | CRUD handler (struct-based, DI) |
| `internal/handler/auth.go` | Auth handler (register, login, refresh) |
| `internal/handler/task.go` | Task CRUD handler |
| `internal/handler/user.go` | User handler (profile, admin) |
| `internal/middleware/auth.go` | JWT auth + role middleware |
| `internal/middleware/ratelimit.go` | Redis-based rate limiter |
| `internal/model/board.go` | Board model + request structs (binding tags) |
| `internal/model/user.go` | User model + auth request structs |
| `internal/model/task.go` | Task model + request structs (enum validation) |
| `pkg/response/json.go` | Standard JSON response helper |

---

## Checklist Bab Ini

- [ ] Paham kenapa Gin dipilih (vs Echo, Fiber)
- [ ] Paham routing: basic route, group, nested group
- [ ] Paham URL params (`c.Param`) dan query params (`c.DefaultQuery`)
- [ ] Bisa bind JSON request ke struct (`ShouldBindJSON`)
- [ ] Paham binding tags untuk validasi (`required`, `min`, `max`, `oneof`, `omitempty`)
- [ ] Paham perbedaan pointer vs non-pointer field di update request
- [ ] Paham standard response format (`pkg/response`)
- [ ] Paham middleware: `c.Next()` vs `c.Abort()`
- [ ] Bisa baca dan nulis auth middleware
- [ ] Paham struct-based handler pattern (manual DI)
- [ ] Paham wiring dependency di `main.go`
- [ ] Bisa mapping mental dari NestJS concept ke Gin equivalent
- [ ] Coba bikin CRUD handler baru (misal: `CommentHandler`)

---

> **Next:** Bab 6 — Database dengan PostgreSQL + GORM. Kita akan bahas migration, model relationship, dan query pattern.