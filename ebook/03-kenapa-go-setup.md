# Bab 3: Kenapa Go + Setup Project

## Kenapa Go?

Go adalah **bahasa #1 di dunia DevOps**. Docker, Kubernetes, Terraform, Prometheus — semua ditulis pakai Go.

### Go vs NestJS

| | NestJS | Go |
|---|--------|-----|
| Demand DevOps/Backend | Tinggi | **Sangat tinggi** |
| Learning curve | Familiar | Beda paradigma |
| Performance | Baik | **Sangat cepat** |
| Concurrency | Event loop | **Goroutine (native)** |
| Binary output | Butuh Node.js | **Single binary** |

### Alasan pilih Go untuk project ini:
1. Go + DevOps = combo langka, bikin CV standout
2. Belajar bahasa baru = buktiin adaptif
3. Di Kubernetes/Docker internals, bisa baca source code langsung
4. Single binary deployment — ga perlu install runtime

---

## Mindset Shift dari NestJS/TypeScript

| NestJS/TypeScript | Go |
|-------------------|-----|
| Class | **Struct + method** |
| Decorator | **Middleware explicit** |
| DI Container (auto) | **Manual dependency injection** |
| ORM heavy (TypeORM) | **GORM atau sqlx (raw SQL common)** |
| throw Exception | **Return error (`if err != nil`)** |
| `implements Interface` | **Implicit interface** |
| `Promise.all()` | **Goroutine + channel** |
| `package.json` | **`go.mod`** |
| `npm install` | **`go get` / `go mod tidy`** |

---

## Go Tech Stack untuk Project Ini

| Kebutuhan | NestJS Stack | Go Stack |
|-----------|-------------|----------|
| Framework | NestJS | **Gin** |
| ORM | TypeORM/Prisma | **GORM** atau **sqlx** |
| Queue | Bull | **Asynq** |
| Auth | Passport JWT | **golang-jwt** |
| Testing | Jest | **Go testing + testify** |
| API Docs | Swagger | **swag** |
| Logger | Winston | **zerolog** / **slog** |

---

## Setup Project

### 1. Install Go
```bash
# macOS
brew install go

# Verify
go version
```

### 2. Init Project
```bash
mkdir "Devops Go" && cd "Devops Go"
go mod init github.com/irzan-aldi/task-manager
```

`go mod init` = `npm init`. Bikin file `go.mod` yang fungsinya sama kayak `package.json`.

### 3. Folder Structure

```
task-manager/
├── cmd/api/main.go           # Entry point
├── internal/                  # Private code
│   ├── config/               # Environment config
│   ├── handler/              # = Controller
│   ├── middleware/            # Auth, logging, rate limit
│   ├── model/                # = Entity/DTO
│   ├── repository/           # Database queries
│   ├── service/              # Business logic
│   └── router/               # Route definitions
├── pkg/                       # Public utilities
│   ├── response/             # JSON response helper
│   └── validator/            # Input validation
├── migrations/                # SQL migrations
├── Makefile                   # Shortcut commands
├── go.mod                     # = package.json
└── go.sum                     # = package-lock.json
```

Kenapa struktur ini?
- `cmd/` → bisa punya multiple entry point (api, worker, cli)
- `internal/` → Go enforce: package lain ga bisa import folder `internal`
- `pkg/` → reusable utilities, bisa di-import project lain

### 4. Install Framework
```bash
go get github.com/gin-gonic/gin
go mod tidy
```

### 5. Mapping NestJS Flow → Go Flow

```
Request masuk
    ↓
Router (Gin) — match route
    ↓
Middleware — auth check, logging
    ↓
Handler — parse request, validate input
    ↓
Service — business logic
    ↓
Repository — query database
    ↓
Response — return JSON
```

Sama persis kayak NestJS. Cuma tanpa decorator dan DI container.

---

## Core Pattern

### Dependency Injection (Manual)

**NestJS (auto-inject):**
```typescript
@Injectable()
export class TaskService {
  constructor(private taskRepo: TaskRepository) {}
}
```

**Go (manual, tapi simple):**
```go
type TaskService struct {
    repo *TaskRepository
}

func NewTaskService(repo *TaskRepository) *TaskService {
    return &TaskService{repo: repo}
}
```

Wiring di `main.go`:
```go
db := connectDB()
taskRepo := repository.NewTaskRepository(db)
taskService := service.NewTaskService(taskRepo)
taskHandler := handler.NewTaskHandler(taskService)
```

Lebih verbose, tapi selalu tau dependency flow. Ga ada magic.

### Error Handling

**NestJS:** throw exception, framework catch
**Go:** return error, caller handle

```go
func (s *TaskService) GetByID(id int) (*Task, error) {
    task, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("find task %d: %w", id, err)
    }
    return task, nil
}
```

`if err != nil` — bakal sering ketik ini. Ini explicit error handling, bukan boilerplate.

### Interface (Implicit)

```go
// Define interface
type TaskRepository interface {
    FindByID(id int) (*Task, error)
    Create(task *Task) error
}

// Implement — ga perlu tulis "implements"
type taskRepoImpl struct {
    db *sql.DB
}

func (r *taskRepoImpl) FindByID(id int) (*Task, error) { ... }
func (r *taskRepoImpl) Create(task *Task) error { ... }
// otomatis satisfy TaskRepository interface
```

### Concurrency

```go
// NestJS: Promise.all([fetchUsers(), fetchTasks()])
// Go: goroutine + channel

ch := make(chan Result, 2)
go func() { ch <- fetchUsers() }()
go func() { ch <- fetchTasks() }()

for i := 0; i < 2; i++ {
    result := <-ch
    // process
}
```

`go func()` = spawn goroutine. Sangat ringan — bisa spawn ribuan.

---

## Makefile

```makefile
run:     go run cmd/api/main.go
build:   go build -o bin/api cmd/api/main.go
test:    go test ./... -v
fmt:     go fmt ./...
deps:    go mod tidy
```

Shortcut biar ga hafal command panjang. `make run` = jalanin server.

---

## Resource Belajar Go (Gratis)

1. **Go Tour** — `go.dev/tour` — interactive, 2-3 jam
2. **Go by Example** — `gobyexample.com` — pattern-based
3. **Effective Go** — `go.dev/doc/effective_go` — best practices

---

## Checklist Bab Ini
- [x] Paham kenapa Go cocok untuk DevOps + Backend
- [x] Paham perbedaan paradigma Go vs NestJS
- [x] Project ter-setup dan bisa compile
- [x] Paham folder structure dan alasannya
- [ ] Baca Go Tour (go.dev/tour) — self-study
- [ ] Baca Go by Example — self-study
