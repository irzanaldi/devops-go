# Bab 9: Message Queue — Async Processing

## Kenapa Bab Ini Penting?

Di bab sebelumnya kita udah pakai Redis buat caching dan rate limiting — semuanya **synchronous**. Request masuk, proses, response keluar. Tapi di real-world, banyak hal yang **ga perlu diselesaikan sekarang juga**. Kirim email? Ga perlu bikin user nunggu. Generate report PDF? Bisa di background. Resize gambar? Biarkan worker yang kerjain.

Kalau kamu dari NestJS, kamu pasti kenal **Bull** — job queue yang populer banget. Nah, di Go, kita punya **Asynq** — library yang konsepnya mirip Bull, sama-sama Redis-based, tapi dengan performa Go.

Di bab ini, kita bakal integrasikan message queue ke project Task Manager API kita. Setiap user register, kita kirim welcome email. Setiap task di-assign, kita kirim notification. Semua async — user ga perlu nunggu.

> **Analogi gampang:** Waktu kamu pesan kopi di kafe, kamu ga berdiri di depan barista nunggu sampe kopinya jadi. Kamu dikasih **nomor antrian**, terus kamu duduk. Barista kerjain pesanan satu-satu. Kamu dipanggil waktu kopi udah ready. Itu message queue.

---

## 9.1 Apa Itu Message Queue?

### Konsep Dasar

Message queue itu **antrian pesan** antara dua pihak:
- **Producer** — yang bikin dan kirim pesan (task/job) ke antrian
- **Consumer/Worker** — yang ambil dan proses pesan dari antrian

```
Producer ──enqueue──→ [ Queue ] ──dequeue──→ Consumer/Worker
                     (antrian)
```

### Real-World Analogy: Antrian Kasir

Bayangin kamu di supermarket:

```
Pelanggan (Producer)     Antrian (Queue)     Kasir (Consumer)
     │                       │                     │
     ├─ ambil barang ──────→ │                     │
     │                       ├─ pelanggan 1 ─────→ proses belanjaan
     ├─ ambil barang ──────→ │                     │
     │                       ├─ pelanggan 2 ─────→ (nunggu giliran)
     ├─ ambil barang ──────→ │                     │
     │                       ├─ pelanggan 3         │
```

- **Pelanggan** ga perlu nunggu kasir selesai — mereka cuma taruh barang di antrian
- **Kasir** proses satu-satu, dengan kecepatan dia sendiri
- Kalau **antrian penuh**, bisa buka kasir baru (scale consumer)
- Kalau kasir **lagi error** (istirahat), antrian tetap ada — ga hilang

### Synchronous vs Asynchronous

**Tanpa Message Queue (Synchronous):**

```
User Register
    ↓
    ├── 1. Validate input      (5ms)
    ├── 2. Hash password       (100ms)
    ├── 3. Save to DB          (50ms)
    ├── 4. Send welcome email  (2000ms) ← USER NUNGGU INI!
    ├── 5. Send Slack notif    (500ms)  ← DAN INI!
    └── Response ke user       (Total: ~2655ms)
```

User nunggu 2.6 detik cuma buat register. Padahal email dan Slack notif **bukan hal yang user perlu tau hasilnya sekarang**.

**Dengan Message Queue (Asynchronous):**

```
User Register
    ↓
    ├── 1. Validate input      (5ms)
    ├── 2. Hash password       (100ms)
    ├── 3. Save to DB          (50ms)
    ├── 4. Enqueue "send email"    (2ms) ← Cuma naruh ke antrian!
    ├── 5. Enqueue "send slack"    (2ms) ← Cuma naruh ke antrian!
    └── Response ke user       (Total: ~159ms)

    ... Meanwhile, di background ...
    Worker 1: Ambil job "send email" → kirim email (2000ms) ✅
    Worker 2: Ambil job "send slack" → kirim notif (500ms) ✅
```

Response time turun dari **2.6 detik ke 159ms**. Drastis!

### Perbandingan dengan NestJS

Di NestJS, kamu mungkin udah pernah pakai:

```typescript
// NestJS — tanpa queue (synchronous)
@Post('register')
async register(@Body() dto: RegisterDto) {
  const user = await this.authService.register(dto);
  await this.mailService.sendWelcomeEmail(user); // Blocking!
  await this.slackService.notify(user);           // Blocking!
  return user;
}

// NestJS — dengan Bull queue (async)
@Post('register')
async register(@Body() dto: RegisterDto) {
  const user = await this.authService.register(dto);
  await this.emailQueue.add('welcome-email', { userId: user.id }); // Non-blocking!
  await this.notifQueue.add('slack-notify', { userId: user.id });  // Non-blocking!
  return user;
}
```

Di Go, konsepnya sama persis. Cuma syntax dan tooling yang beda.

---

## 9.2 Kapan Pakai Message Queue?

### Use Cases yang COCOK

| Use Case | Contoh di Project Kita | Kenapa Async? |
|----------|----------------------|---------------|
| **Email notification** | Welcome email setelah register | SMTP lambat, 1-3 detik |
| **Push notification** | Notif saat task di-assign | Ga perlu real-time instant |
| **Image processing** | Resize avatar user | CPU-intensive, bisa lama |
| **Report generation** | Export board ke PDF/CSV | Query berat, bisa 10+ detik |
| **Webhook delivery** | Notify external service | External service bisa slow/down |
| **Data sync** | Sync ke analytics/search | Ga perlu real-time |
| **Scheduled tasks** | Daily digest email | Perlu dijalankan di waktu tertentu |
| **Retry logic** | Retry failed payment | Butuh automatic retry dengan backoff |

### Kapan TIDAK Perlu Queue

Jangan over-engineering. Ga semua hal butuh queue.

```
JANGAN PAKAI QUEUE:
- Simple CRUD yang cepat (< 100ms)
- Response yang user butuh sekarang (validate email, check balance)
- Data yang harus konsisten real-time (update stock saat checkout)
- Proses yang ga pernah fail dan cepat
- Project masih MVP dan user cuma 10 orang

PAKAI QUEUE KALAU:
- Proses lambat (> 500ms) dan user ga perlu nunggu hasilnya
- Ada dependency ke external service (email, SMS, payment gateway)
- Butuh retry otomatis kalau gagal
- Butuh scheduled/delayed execution
- Ada spike traffic yang perlu di-smooth-kan
```

### Rule of Thumb

> Kalau user **ga perlu tau hasilnya sekarang**, dan prosesnya **bisa gagal tanpa merusak flow utama** — pakai queue.

---

## 9.3 Pilihan Message Queue

Ada banyak pilihan message queue. Ini 3 yang paling populer:

### Comparison Table

| | **RabbitMQ** | **Apache Kafka** | **Asynq (Redis-based)** |
|---|-------------|-----------------|------------------------|
| **Tipe** | Traditional message broker | Distributed event streaming | Task queue (Redis) |
| **Protocol** | AMQP | Custom binary | Redis protocol |
| **Use case utama** | Task queue, pub/sub | Event streaming, log aggregation | Background job processing |
| **Ordering** | Per-queue FIFO | Per-partition ordered | Per-queue FIFO |
| **Persistence** | Disk | Disk (append-only log) | Redis (memory + optional disk) |
| **Throughput** | ~50K msg/sec | **1M+ msg/sec** | ~100K task/sec |
| **Complexity** | Medium | **High** | **Low** |
| **Setup** | Erlang runtime, management UI | JVM, ZooKeeper/KRaft | Redis aja (udah ada!) |
| **Learning curve** | Medium | Steep | **Very low** |
| **Consumer group** | Ya | Ya | Ya (queue priority) |
| **Delayed message** | Plugin | Workaround | **Built-in** |
| **Retry** | Manual | Manual | **Built-in** |
| **Web UI** | Built-in management | Kafka UI (third-party) | **Asynqmon** |
| **Go library** | `amqp091-go` | `confluent-kafka-go` | `hibiken/asynq` |
| **NestJS equivalent** | `@nestjs/rabbitmq` | `@nestjs/kafka` | **Bull (`@nestjs/bull`)** |

### Kapan Pakai Apa?

```
RabbitMQ:
  → Butuh complex routing (topic, fanout, headers)
  → Multi-language ecosystem
  → Enterprise-grade message broker

Kafka:
  → Event streaming (jutaan event/detik)
  → Data pipeline (ETL)
  → Event sourcing / CQRS
  → Log aggregation

Asynq (Redis):
  → Background job processing (email, notification, report)
  → Delayed/scheduled tasks
  → Simple queue yang butuh retry
  → Project yang udah pakai Redis
```

### Kenapa Kita Pilih Asynq?

1. **Redis udah ada** — project kita udah pakai Redis buat caching (Bab 8). Ga perlu install infrastruktur baru.

2. **Mirip Bull di NestJS** — kalau kamu udah familiar sama Bull, Asynq itu hampir identik secara konsep. Redis-based, task queue, built-in retry, delayed task, web UI.

3. **Simple banget** — setup cuma beberapa baris kode. Ga perlu belajar protocol baru (AMQP, Kafka protocol).

4. **Built-in features** — retry, delay, timeout, unique task, priority queue, scheduled task, dead letter queue. Semua udah ada.

5. **Battle-tested** — dipakai banyak company di production.

6. **Perfect buat project kita** — kita butuh kirim email dan notification di background. Ga butuh event streaming jutaan per detik.

> **Overkill alert:** Kalau kamu cuma butuh kirim email dan notification di background, pakai Kafka itu kayak beli truk kontainer buat belanja ke warung. Asynq cukup banget.

---

## 9.4 Asynq Overview — Architecture

### Architecture

```
                         ┌──────────────────────────────────┐
                         │            Redis                  │
                         │                                   │
┌─────────────┐          │  ┌──────────┐   ┌──────────────┐ │         ┌─────────────────┐
│   Client     │──enqueue→│  │  Queue    │   │  Scheduled   │ │         │   Server/Worker  │
│  (Producer)  │          │  │  (ready)  │   │  (delayed)   │ │──poll──→│   (Consumer)     │
│              │          │  └──────────┘   └──────────────┘ │         │                  │
│ - API Handler│          │  ┌──────────┐   ┌──────────────┐ │         │ - HandleFunc()   │
│ - Service    │          │  │  Retry    │   │  Dead Letter  │ │         │ - Process task   │
│ - Scheduler  │          │  │  Queue    │   │  Queue (DLQ)  │ │         │ - Return error   │
└─────────────┘          │  └──────────┘   └──────────────┘ │         │   → auto retry   │
                         └──────────────────────────────────┘         └─────────────────┘
```

**Flow:**

1. **Client** bikin task dan enqueue ke Redis
2. Task masuk ke **queue** (ready) atau **scheduled** (kalau ada delay)
3. **Server/Worker** poll task dari queue
4. Worker **process** task — kalau sukses, task dihapus. Kalau gagal, masuk **retry queue**
5. Kalau udah max retry dan masih gagal, task masuk **Dead Letter Queue (DLQ)**

### Perbandingan dengan NestJS Bull

| Konsep | Bull (NestJS) | Asynq (Go) |
|--------|--------------|------------|
| **Queue backend** | Redis | Redis |
| **Enqueue** | `queue.add('job-name', data)` | `client.Enqueue(task)` |
| **Process** | `@Process('job-name')` | `mux.HandleFunc("task:type", handler)` |
| **Job/Task** | `Job` object | `asynq.Task` |
| **Delay** | `queue.add('job', data, { delay: 5000 })` | `asynq.ProcessIn(5 * time.Second)` |
| **Retry** | `{ attempts: 3 }` | `asynq.MaxRetry(3)` |
| **Queue priority** | `{ priority: 1 }` | Queue-level priority config |
| **Timeout** | `{ timeout: 30000 }` | `asynq.Timeout(30 * time.Second)` |
| **Unique** | `{ jobId: 'unique-id' }` | `asynq.Unique(1 * time.Hour)` |
| **Scheduled** | `queue.add('job', data, { repeat: { cron: '...' } })` | `scheduler.Register(cron, task)` |
| **Web UI** | Bull Board / Arena | Asynqmon |
| **Graceful shutdown** | `queue.close()` | `server.Shutdown()` |

Kalau kamu perhatiin, mapping-nya hampir 1:1. Yang beda cuma syntax dan cara Go handle sesuatu (explicit, no decorator).

---

## 9.5 Setup Asynq

### Install Package

```bash
go get github.com/hibiken/asynq
```

### Project Structure

Kita bakal tambah beberapa file baru ke project:

```
task-manager/
├── cmd/
│   ├── api/
│   │   └── main.go          ← API server (existing)
│   └── worker/
│       └── main.go          ← Worker process (BARU)
├── internal/
│   ├── worker/
│   │   ├── client.go        ← Asynq client (producer) (BARU)
│   │   ├── server.go        ← Asynq server (consumer) (BARU)
│   │   ├── tasks.go         ← Task type definitions (BARU)
│   │   ├── handler_email.go ← Email task handler (BARU)
│   │   └── handler_notif.go ← Notification task handler (BARU)
│   ├── config/
│   ├── handler/
│   ├── model/
│   ├── repository/
│   ├── service/
│   └── ...
```

**Kenapa worker dipisah dari API server?**

Worker jalan sebagai **process terpisah** dari API server. Ini penting karena:
- Worker punya lifecycle sendiri (bisa di-scale independen)
- Kalau worker crash, API tetap jalan
- Bisa deploy worker di server yang beda (atau lebih banyak instance)

Di NestJS Bull, biasanya worker dan API jalan di satu process. Di Go, best practice-nya **dipisah**.

```
# Terminal 1 — API server
go run cmd/api/main.go

# Terminal 2 — Worker
go run cmd/worker/main.go
```

### Basic Config — Redis Connection

Asynq butuh Redis connection. Karena kita udah punya Redis config, kita reuse:

```go
// internal/worker/client.go

package worker

import (
	"github.com/hibiken/asynq"
)

// NewAsynqClient bikin Asynq client untuk enqueue task
func NewAsynqClient(redisAddr, redisPassword string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})
}
```

```go
// internal/worker/server.go

package worker

import (
	"log"

	"github.com/hibiken/asynq"
)

// NewAsynqServer bikin Asynq server untuk process task
func NewAsynqServer(redisAddr, redisPassword string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			// Jumlah concurrent worker (goroutine)
			Concurrency: 10,

			// Queue priority — angka lebih tinggi = prioritas lebih tinggi
			Queues: map[string]int{
				"critical": 6, // 60% processing power
				"default":  3, // 30% processing power
				"low":      1, // 10% processing power
			},

			// Error handler — dipanggil setiap ada task yang gagal
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("[Asynq Error] task=%s, error=%v", task.Type(), err)
			}),
		},
	)
}
```

**Penjelasan Config:**

| Config | Deskripsi |
|--------|-----------|
| `Concurrency: 10` | Maks 10 task diproses bersamaan (10 goroutine) |
| `Queues` | Map queue name ke priority weight. `critical:6` berarti 60% dari capacity buat queue "critical" |
| `ErrorHandler` | Callback yang dipanggil setiap ada error. Bagus buat logging/alerting |

**NestJS Bull equivalent:**

```typescript
// NestJS Bull — queue config
@Module({
  imports: [
    BullModule.registerQueue({
      name: 'email',
      redis: { host: 'localhost', port: 6379 },
      defaultJobOptions: {
        attempts: 3,
        backoff: { type: 'exponential', delay: 1000 },
      },
    }),
  ],
})
export class AppModule {}
```

---

## 9.6 Task Definition

### Define Task Types

Setiap "job" di Asynq disebut **Task**. Task punya:
- **Type** — nama unik (string), mirip job name di Bull
- **Payload** — data yang dikirim ke worker (JSON)

```go
// internal/worker/tasks.go

package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// ==================== Task Type Constants ====================
// Naming convention: <domain>:<action>
// Mirip Bull job name: "email:send-welcome", "notification:task-assigned"

const (
	TaskSendWelcomeEmail    = "email:send-welcome"
	TaskSendTaskAssignedNotif = "notification:task-assigned"
	TaskGenerateReport      = "report:generate"
	TaskResizeAvatar        = "image:resize-avatar"
)

// ==================== Payload Structs ====================

// EmailWelcomePayload — data yang dikirim ke worker untuk kirim welcome email
type EmailWelcomePayload struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// TaskAssignedPayload — data untuk notifikasi task assigned
type TaskAssignedPayload struct {
	TaskID     uint   `json:"task_id"`
	TaskTitle  string `json:"task_title"`
	AssigneeID uint   `json:"assignee_id"`
	AssignerID uint   `json:"assigner_id"`
	BoardID    uint   `json:"board_id"`
}

// ReportPayload — data untuk generate report
type ReportPayload struct {
	BoardID   uint   `json:"board_id"`
	UserID    uint   `json:"user_id"`
	Format    string `json:"format"` // "pdf" atau "csv"
}

// ==================== Task Constructors ====================
// Setiap task punya constructor yang return *asynq.Task

// NewSendWelcomeEmailTask bikin task untuk kirim welcome email
func NewSendWelcomeEmailTask(payload EmailWelcomePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskSendWelcomeEmail, data), nil
}

// NewTaskAssignedNotifTask bikin task untuk kirim notification
func NewTaskAssignedNotifTask(payload TaskAssignedPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskSendTaskAssignedNotif, data), nil
}

// NewGenerateReportTask bikin task untuk generate report
func NewGenerateReportTask(payload ReportPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskGenerateReport, data), nil
}
```

### Perbandingan dengan NestJS Bull

**NestJS Bull — Job Definition:**

```typescript
// Di NestJS, biasanya job definition itu cuma string + interface
interface WelcomeEmailJob {
  userId: number;
  email: string;
  name: string;
}

// Enqueue
await this.emailQueue.add('send-welcome', {
  userId: user.id,
  email: user.email,
  name: user.name,
} as WelcomeEmailJob);
```

**Go Asynq — Task Definition:**

```go
// Di Go, kita bikin struct + constructor (lebih explicit, type-safe)
payload := worker.EmailWelcomePayload{
    UserID: user.ID,
    Email:  user.Email,
    Name:   user.Name,
}
task, err := worker.NewSendWelcomeEmailTask(payload)
```

Bedanya:
- NestJS: payload bisa object apa aja (loosely typed). Type safety opsional via interface
- Go: payload harus di-marshal ke JSON. Struct bikin type safety guaranteed

### Pattern: `NewTask()` Constructor

Kenapa kita bikin constructor function (`NewSendWelcomeEmailTask()`) daripada langsung `asynq.NewTask()`?

```go
// Tanpa constructor — error-prone, payload bisa salah
task := asynq.NewTask("email:send-welcome", []byte(`{"user_id": 1}`))

// Dengan constructor — type-safe, ga bisa salah field
task, err := worker.NewSendWelcomeEmailTask(worker.EmailWelcomePayload{
    UserID: 1,
    Email:  "irzan@mail.com",
    Name:   "Irzan",
})
```

Constructor memastikan:
1. Task type string ga typo (pake constant)
2. Payload structure konsisten (pake struct)
3. JSON marshaling di-handle di satu tempat

---

## 9.7 Producer — Enqueue Task dari Service Layer

### Inject Asynq Client ke Service

Sekarang kita perlu enqueue task dari service layer. Caranya: inject `*asynq.Client` ke service yang butuh.

```go
// internal/service/auth.go — UPDATED: tambah Asynq client

package service

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hibiken/asynq"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/repository"
	"github.com/irzan-aldi/task-manager/internal/worker"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtSecret   string
	asynqClient *asynq.Client // TAMBAH: Asynq client
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, asynqClient *asynq.Client) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		asynqClient: asynqClient,
	}
}
```

### Enqueue Task saat User Register

```go
// internal/service/auth.go — method Register UPDATED

func (s *AuthService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
	// Check if email already exists
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     model.RoleMember,
	}

	if err := user.HashPassword(); err != nil {
		return nil, errors.New("failed to hash password")
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// ========== ENQUEUE ASYNC TASK ==========
	// Kirim welcome email di background — user ga perlu nunggu
	task, err := worker.NewSendWelcomeEmailTask(worker.EmailWelcomePayload{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
	})
	if err != nil {
		// Log error tapi JANGAN return error ke user
		// Registration tetap sukses walau enqueue gagal
		log.Printf("Failed to create welcome email task: %v", err)
	} else {
		// Enqueue dengan options:
		// - MaxRetry(3): coba 3x kalau gagal
		// - Queue("default"): masuk queue "default"
		// - ProcessIn(10 * time.Second): delay 10 detik (biar DB commit dulu)
		info, err := s.asynqClient.Enqueue(task,
			asynq.MaxRetry(3),
			asynq.Queue("default"),
			asynq.ProcessIn(10*time.Second),
		)
		if err != nil {
			log.Printf("Failed to enqueue welcome email task: %v", err)
		} else {
			log.Printf("Enqueued task: id=%s queue=%s", info.ID, info.Queue)
		}
	}
	// ========================================

	resp := user.ToResponse()
	return &resp, nil
}
```

**Poin penting:**
1. **Jangan return error ke user kalau enqueue gagal.** Registration udah sukses — user udah ke-create di DB. Gagal kirim email bukan alasan buat rollback registration.
2. **ProcessIn(10 * time.Second)** — delay 10 detik. Ini safety net buat mastiin database transaction udah committed sebelum worker mulai proses.
3. **MaxRetry(3)** — kalau SMTP server down, Asynq otomatis retry 3 kali dengan exponential backoff.

### Enqueue Task saat Task Assigned

```go
// internal/service/task.go — UPDATED: tambah notification saat task assigned

package service

import (
	"errors"
	"log"

	"github.com/hibiken/asynq"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/repository"
	"github.com/irzan-aldi/task-manager/internal/worker"
)

type TaskService struct {
	repo        *repository.TaskRepository
	boardRepo   *repository.BoardRepository
	asynqClient *asynq.Client // TAMBAH: Asynq client
}

func NewTaskService(repo *repository.TaskRepository, boardRepo *repository.BoardRepository, asynqClient *asynq.Client) *TaskService {
	return &TaskService{
		repo:        repo,
		boardRepo:   boardRepo,
		asynqClient: asynqClient,
	}
}

func (s *TaskService) Create(boardID uint, userID uint, req model.CreateTaskRequest) (*model.Task, error) {
	// ... existing validation logic ...

	board, err := s.boardRepo.FindByID(boardID)
	if err != nil {
		return nil, errors.New("board not found")
	}

	if board.OwnerID != userID {
		return nil, errors.New("not authorized to add tasks to this board")
	}

	task := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		BoardID:     boardID,
		AssigneeID:  req.AssigneeID,
		DueDate:     req.DueDate,
		Status:      model.StatusTodo,
	}

	if task.Priority == "" {
		task.Priority = model.PriorityMedium
	}

	if err := s.repo.Create(task); err != nil {
		return nil, errors.New("failed to create task")
	}

	// ========== ENQUEUE NOTIFICATION ==========
	// Kalau task di-assign ke seseorang, kirim notification
	if req.AssigneeID != nil {
		notifTask, err := worker.NewTaskAssignedNotifTask(worker.TaskAssignedPayload{
			TaskID:     task.ID,
			TaskTitle:  task.Title,
			AssigneeID: *req.AssigneeID,
			AssignerID: userID,
			BoardID:    boardID,
		})
		if err != nil {
			log.Printf("Failed to create notification task: %v", err)
		} else {
			info, err := s.asynqClient.Enqueue(notifTask,
				asynq.MaxRetry(3),
				asynq.Queue("default"),
			)
			if err != nil {
				log.Printf("Failed to enqueue notification task: %v", err)
			} else {
				log.Printf("Enqueued notification: id=%s queue=%s", info.ID, info.Queue)
			}
		}
	}
	// ==========================================

	return task, nil
}
```

### Perbandingan dengan NestJS Bull — Producer

**NestJS Bull:**

```typescript
// NestJS — inject queue dan add job

@Injectable()
export class AuthService {
  constructor(
    @InjectQueue('email') private emailQueue: Queue,
    private userRepo: UserRepository,
  ) {}

  async register(dto: RegisterDto) {
    const user = await this.userRepo.create(dto);

    // Enqueue job
    await this.emailQueue.add('send-welcome', {
      userId: user.id,
      email: user.email,
      name: user.name,
    }, {
      delay: 10000,       // 10 detik delay
      attempts: 3,        // 3x retry
      backoff: {
        type: 'exponential',
        delay: 1000,
      },
    });

    return user;
  }
}
```

**Go Asynq:**

```go
// Go — inject *asynq.Client dan enqueue task

func (s *AuthService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
    user := // ... create user ...

    task, _ := worker.NewSendWelcomeEmailTask(worker.EmailWelcomePayload{
        UserID: user.ID,
        Email:  user.Email,
        Name:   user.Name,
    })
    s.asynqClient.Enqueue(task,
        asynq.ProcessIn(10*time.Second), // 10 detik delay
        asynq.MaxRetry(3),               // 3x retry
    )

    return &resp, nil
}
```

Hampir identik. Bedanya:
- NestJS: inject via `@InjectQueue()` decorator, options di object `{}`
- Go: inject via constructor, options sebagai variadic args

---

## 9.8 Consumer/Worker — Process Task

### Task Handler

Setiap task type butuh handler function. Ini yang beneran **ngerjain** job-nya.

```go
// internal/worker/handler_email.go

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

// HandleSendWelcomeEmail memproses task kirim welcome email
func HandleSendWelcomeEmail(ctx context.Context, t *asynq.Task) error {
	// 1. Unmarshal payload
	var payload EmailWelcomePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		// Return error → task akan di-retry
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Worker] Processing welcome email for user %d (%s)", payload.UserID, payload.Email)

	// 2. Kirim email (di sini kamu pakai SMTP client, SendGrid, dll)
	err := sendEmail(payload.Email, payload.Name)
	if err != nil {
		// Return error → Asynq akan retry sesuai MaxRetry
		return fmt.Errorf("failed to send welcome email to %s: %w", payload.Email, err)
	}

	// 3. Log success
	log.Printf("[Worker] Welcome email sent to %s", payload.Email)

	// Return nil = task sukses, dihapus dari queue
	return nil
}

// sendEmail — placeholder, ganti dengan real email service
func sendEmail(to, name string) error {
	// Di production, pakai:
	// - net/smtp (standard library)
	// - github.com/jordan-wright/email (simple wrapper)
	// - SendGrid API
	// - AWS SES
	// - Mailgun

	// Simulasi kirim email
	log.Printf("[Email] Sending welcome email to %s (name: %s)", to, name)
	// time.Sleep(2 * time.Second) // Simulasi latency SMTP

	// Uncomment untuk simulasi error (test retry):
	// return errors.New("SMTP connection refused")

	log.Printf("[Email] Email sent successfully to %s", to)
	return nil
}
```

```go
// internal/worker/handler_notif.go

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

// HandleTaskAssignedNotif memproses notification saat task di-assign
func HandleTaskAssignedNotif(ctx context.Context, t *asynq.Task) error {
	// 1. Unmarshal payload
	var payload TaskAssignedPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Worker] Processing task-assigned notification: task=%d, assignee=%d",
		payload.TaskID, payload.AssigneeID)

	// 2. Kirim notification
	// Di production, ini bisa:
	// - Push notification (FCM/APNs)
	// - In-app notification (simpan ke DB notification table)
	// - Email notification
	// - Slack/Discord webhook

	err := sendNotification(payload)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	log.Printf("[Worker] Notification sent for task %d to user %d", payload.TaskID, payload.AssigneeID)
	return nil
}

func sendNotification(payload TaskAssignedPayload) error {
	// Placeholder — ganti dengan real notification service
	log.Printf("[Notification] User %d assigned task '%s' (ID: %d) by user %d",
		payload.AssigneeID, payload.TaskTitle, payload.TaskID, payload.AssignerID)
	return nil
}
```

### Handler Pattern Explained

```go
func HandleSomething(ctx context.Context, t *asynq.Task) error {
    // ctx — context untuk cancellation, timeout
    // t   — task object, punya Type() dan Payload()

    // Unmarshal payload
    var payload SomePayload
    json.Unmarshal(t.Payload(), &payload)

    // Do work...

    // return nil    → task SUCCESS, dihapus dari queue
    // return error  → task FAILED, masuk retry queue
    // return asynq.SkipRetry → task FAILED, SKIP retry, langsung ke DLQ
}
```

### Kapan Return Error vs SkipRetry

```go
func HandleSendEmail(ctx context.Context, t *asynq.Task) error {
    var payload EmailWelcomePayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        // Payload rusak — retry juga ga akan fix
        // Skip retry, langsung ke Dead Letter Queue
        return fmt.Errorf("invalid payload: %w: %w", err, asynq.SkipRetry)
    }

    // Kirim email
    err := sendEmail(payload.Email, payload.Name)
    if err != nil {
        // SMTP error — bisa temporary, worth retry
        return fmt.Errorf("smtp error: %w", err)
    }

    return nil
}
```

Rule:
- **Return error** → retry (SMTP down, network timeout, temporary error)
- **Wrap with `asynq.SkipRetry`** → skip retry, masuk DLQ (invalid payload, user deleted, permanent error)

### Setup Worker — Register Handlers

```go
// internal/worker/server.go — LENGKAP

package worker

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
)

func NewAsynqServer(redisAddr, redisPassword string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("[Asynq Error] task=%s, error=%v", task.Type(), err)
			}),
		},
	)
}

// NewServeMux bikin mux dan register semua task handlers
func NewServeMux() *asynq.ServeMux {
	mux := asynq.NewServeMux()

	// Register handlers — mirip HTTP router tapi untuk task
	mux.HandleFunc(TaskSendWelcomeEmail, HandleSendWelcomeEmail)
	mux.HandleFunc(TaskSendTaskAssignedNotif, HandleTaskAssignedNotif)
	// mux.HandleFunc(TaskGenerateReport, HandleGenerateReport)
	// mux.HandleFunc(TaskResizeAvatar, HandleResizeAvatar)

	return mux
}
```

**Perhatiin pattern-nya:** `mux.HandleFunc()` mirip banget sama HTTP router di Gin (`r.GET()`, `r.POST()`). Bedanya ini router untuk **task processing**, bukan HTTP request.

### Perbandingan dengan NestJS Bull — Consumer

**NestJS Bull:**

```typescript
// NestJS — processor pakai decorator

@Processor('email')
export class EmailProcessor {
  @Process('send-welcome')
  async handleSendWelcome(job: Job<WelcomeEmailJob>) {
    const { userId, email, name } = job.data;
    console.log(`Sending welcome email to ${email}`);
    await this.mailService.sendWelcome(email, name);
  }

  @OnQueueFailed()
  onFailed(job: Job, error: Error) {
    console.error(`Job ${job.id} failed: ${error.message}`);
  }
}
```

**Go Asynq:**

```go
// Go — handler sebagai function, register di mux

func HandleSendWelcomeEmail(ctx context.Context, t *asynq.Task) error {
    var payload EmailWelcomePayload
    json.Unmarshal(t.Payload(), &payload)
    log.Printf("Sending welcome email to %s", payload.Email)
    return sendEmail(payload.Email, payload.Name)
}

// Register
mux.HandleFunc("email:send-welcome", HandleSendWelcomeEmail)
```

Perbedaan mendasar:
- NestJS: decorator `@Processor`, `@Process()`, `@OnQueueFailed()` — magic
- Go: plain function + explicit registration — no magic, everything visible

---

## 9.9 Task Options — Fine-tuning Behavior

Asynq punya banyak options buat control behavior setiap task.

### Semua Options yang Tersedia

```go
info, err := client.Enqueue(task,
    // ===== TIMING =====
    asynq.ProcessIn(30 * time.Second),         // Delay: proses 30 detik dari sekarang
    asynq.ProcessAt(time.Date(2026, 5, 1, 9, 0, 0, 0, time.Local)), // Jadwal: 1 Mei 2026 jam 9
    asynq.Deadline(time.Now().Add(24 * time.Hour)),  // Deadline: harus selesai dalam 24 jam

    // ===== RETRY =====
    asynq.MaxRetry(5),                         // Max 5x retry
    asynq.Timeout(2 * time.Minute),            // Timeout per attempt: 2 menit

    // ===== QUEUE =====
    asynq.Queue("critical"),                   // Masuk queue "critical"

    // ===== UNIQUENESS =====
    asynq.Unique(1 * time.Hour),               // Prevent duplicate: 1 task unik per 1 jam

    // ===== IDENTIFICATION =====
    asynq.TaskID("welcome-email-user-123"),    // Custom task ID

    // ===== RETENTION =====
    asynq.Retention(24 * time.Hour),           // Simpan completed task selama 24 jam (buat audit)
)
```

### Detail Setiap Option

#### 1. `ProcessIn()` — Delayed Task

```go
// Kirim email 10 detik setelah register (biar DB commit dulu)
client.Enqueue(task, asynq.ProcessIn(10 * time.Second))

// Kirim reminder 1 jam sebelum due date
client.Enqueue(reminderTask, asynq.ProcessIn(1 * time.Hour))
```

**NestJS Bull equivalent:**
```typescript
queue.add('send-email', data, { delay: 10000 }); // 10 detik
```

#### 2. `MaxRetry()` — Retry dengan Exponential Backoff

```go
// Retry 5 kali dengan exponential backoff
// Retry 1: ~10 detik setelah fail
// Retry 2: ~20 detik
// Retry 3: ~40 detik
// Retry 4: ~80 detik
// Retry 5: ~160 detik
// Total: ~5 menit sebelum masuk Dead Letter Queue
client.Enqueue(task, asynq.MaxRetry(5))
```

**NestJS Bull equivalent:**
```typescript
queue.add('send-email', data, {
  attempts: 5,
  backoff: { type: 'exponential', delay: 10000 },
});
```

#### 3. `Timeout()` — Per-Attempt Timeout

```go
// Kalau 1 attempt lebih dari 2 menit, cancel dan retry
client.Enqueue(task, asynq.Timeout(2 * time.Minute))
```

Di handler, kamu bisa cek context cancellation:

```go
func HandleGenerateReport(ctx context.Context, t *asynq.Task) error {
    // ctx.Done() akan di-close kalau timeout
    select {
    case <-ctx.Done():
        return ctx.Err() // deadline exceeded
    default:
        // continue processing
    }

    // Atau lebih praktis: pass ctx ke function yang bisa cancel
    report, err := generateReport(ctx, payload.BoardID)
    if err != nil {
        return err
    }
    return nil
}
```

**NestJS Bull equivalent:**
```typescript
queue.add('generate-report', data, { timeout: 120000 }); // 2 menit
```

#### 4. `Unique()` — Prevent Duplicate Tasks

```go
// Cegah kirim email welcome duplikat ke user yang sama
task, _ := worker.NewSendWelcomeEmailTask(payload)
client.Enqueue(task,
    asynq.Unique(1 * time.Hour), // 1 task unik per 1 jam
)

// Kalau user register lagi dalam 1 jam (misal double-click),
// Enqueue return ErrDuplicateTask — task ga diduplikat
```

Asynq pakai combination dari **task type + payload** sebagai uniqueness key. Jadi `email:send-welcome` dengan payload `{user_id: 1}` cuma bisa ada 1 dalam window 1 jam.

**NestJS Bull equivalent:**
```typescript
queue.add('send-welcome', data, {
  jobId: `welcome-${userId}`, // unique job ID
});
```

#### 5. `Queue()` — Queue Priority

```go
// Email penting — masuk queue critical
client.Enqueue(passwordResetTask, asynq.Queue("critical"))

// Welcome email — biasa aja
client.Enqueue(welcomeEmailTask, asynq.Queue("default"))

// Report generation — bisa nunggu
client.Enqueue(reportTask, asynq.Queue("low"))
```

Ini ngikutin priority yang kita set di server config:

```go
Queues: map[string]int{
    "critical": 6, // 60% processing power
    "default":  3, // 30%
    "low":      1, // 10%
},
```

Artinya kalau ada 10 concurrent worker, kira-kira:
- 6 worker fokus ke `critical`
- 3 worker ke `default`
- 1 worker ke `low`

#### 6. `Retention()` — Keep Completed Tasks

```go
// Simpan completed task 7 hari (buat audit/debugging)
client.Enqueue(task, asynq.Retention(7 * 24 * time.Hour))
```

Default-nya, completed task langsung dihapus. Dengan `Retention`, task yang sukses tetap disimpan di Redis selama durasi tertentu — berguna buat audit trail.

---

## 9.10 Monitoring — Asynq CLI dan Web UI

### Asynq CLI

Install CLI tool:

```bash
go install github.com/hibiken/asynq/tools/asynq@latest
```

Command yang berguna:

```bash
# Dashboard — overview semua queue
asynq dash

# Stats — lihat stats per queue
asynq stats

# List task di queue tertentu
asynq task list --queue=default --state=active
asynq task list --queue=default --state=pending
asynq task list --queue=default --state=retry
asynq task list --queue=default --state=archived  # dead letter queue

# Cancel task
asynq task cancel <task_id>

# Delete task
asynq task delete --queue=default --state=archived --id=<task_id>

# Drain queue (hapus semua task pending)
asynq queue remove --queue=low
```

### Asynqmon — Web UI

Asynqmon itu web dashboard buat monitor Asynq. Mirip Bull Board di NestJS.

**Install dan jalankan:**

```bash
# Install
go install github.com/hibiken/asynqmon/cmd/asynqmon@latest

# Jalankan (connect ke Redis)
asynqmon --redis-addr=localhost:6379

# Buka browser: http://localhost:8080
```

**Atau pakai Docker:**

```bash
docker run --rm \
    --name asynqmon \
    -p 8080:8080 \
    hibiken/asynqmon \
    --redis-addr=host.docker.internal:6379
```

### Apa yang bisa dilihat di Asynqmon:

| Feature | Deskripsi |
|---------|-----------|
| **Queue overview** | Jumlah task per queue, per state |
| **Active tasks** | Task yang sedang diproses |
| **Pending tasks** | Task yang nunggu giliran |
| **Scheduled tasks** | Task yang delayed/scheduled |
| **Retry tasks** | Task yang gagal dan nunggu retry |
| **Archived tasks** | Dead letter queue — task yang udah max retry |
| **Task detail** | Lihat payload, error message, retry count |
| **Actions** | Delete, run now, archive, batch operations |

### Perbandingan Monitoring Tool

| | Bull Board (NestJS) | Asynqmon (Go) |
|---|-------------------|---------------|
| **Install** | `npm install @bull-board/express` | `go install asynqmon` |
| **UI** | Web-based | Web-based |
| **Queue stats** | Ya | Ya |
| **Task detail** | Ya | Ya |
| **Retry/delete** | Ya | Ya |
| **Real-time** | Ya | Ya |
| **Docker** | Ya | Ya |

Dari segi fitur, keduanya **setara**. Asynqmon bahkan lebih clean dan informatif menurut banyak developer.

---

## 9.11 Practical Example — Full Implementation

Sekarang kita satuin semuanya ke project Task Manager kita.

### Step 1: File Task Definitions

```go
// internal/worker/tasks.go

package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task type constants
const (
	TaskSendWelcomeEmail      = "email:send-welcome"
	TaskSendTaskAssignedNotif = "notification:task-assigned"
)

// ===== Payloads =====

type EmailWelcomePayload struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

type TaskAssignedPayload struct {
	TaskID     uint   `json:"task_id"`
	TaskTitle  string `json:"task_title"`
	AssigneeID uint   `json:"assignee_id"`
	AssignerID uint   `json:"assigner_id"`
	BoardID    uint   `json:"board_id"`
}

// ===== Task Constructors =====

func NewSendWelcomeEmailTask(p EmailWelcomePayload) (*asynq.Task, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskSendWelcomeEmail, data), nil
}

func NewTaskAssignedNotifTask(p TaskAssignedPayload) (*asynq.Task, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskSendTaskAssignedNotif, data), nil
}
```

### Step 2: Email Handler

```go
// internal/worker/handler_email.go

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

func HandleSendWelcomeEmail(ctx context.Context, t *asynq.Task) error {
	var p EmailWelcomePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal failed: %w: %w", err, asynq.SkipRetry)
	}

	log.Printf("[Email Worker] Sending welcome email to %s (%s)...", p.Name, p.Email)

	// TODO: Ganti dengan real email service
	// Contoh pakai net/smtp:
	//
	// auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	// msg := fmt.Sprintf("Subject: Welcome %s!\r\n\r\nTerima kasih sudah register!", p.Name)
	// err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{p.Email}, []byte(msg))

	log.Printf("[Email Worker] Welcome email sent to %s", p.Email)
	return nil
}
```

### Step 3: Notification Handler

```go
// internal/worker/handler_notif.go

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

func HandleTaskAssignedNotif(ctx context.Context, t *asynq.Task) error {
	var p TaskAssignedPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal failed: %w: %w", err, asynq.SkipRetry)
	}

	log.Printf("[Notif Worker] User %d was assigned task '%s' (ID:%d) by user %d",
		p.AssigneeID, p.TaskTitle, p.TaskID, p.AssignerID)

	// TODO: Ganti dengan real notification service
	// Opsi:
	// 1. Simpan ke notification table di DB
	// 2. Kirim push notification via FCM
	// 3. Kirim email notification
	// 4. Post ke Slack/Discord webhook

	log.Printf("[Notif Worker] Notification sent for task %d", p.TaskID)
	return nil
}
```

### Step 4: Asynq Client (Producer)

```go
// internal/worker/client.go

package worker

import (
	"github.com/hibiken/asynq"
)

func NewAsynqClient(redisAddr, redisPassword string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})
}
```

### Step 5: Asynq Server (Consumer)

```go
// internal/worker/server.go

package worker

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
)

func NewAsynqServer(redisAddr, redisPassword string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("[Asynq Error] type=%s error=%v", task.Type(), err)
			}),
		},
	)
}

func NewServeMux() *asynq.ServeMux {
	mux := asynq.NewServeMux()

	// Register all task handlers
	mux.HandleFunc(TaskSendWelcomeEmail, HandleSendWelcomeEmail)
	mux.HandleFunc(TaskSendTaskAssignedNotif, HandleTaskAssignedNotif)

	return mux
}
```

### Step 6: Worker Entrypoint

```go
// cmd/worker/main.go

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/irzan-aldi/task-manager/internal/config"
	"github.com/irzan-aldi/task-manager/internal/worker"
)

func main() {
	log.Println("Starting Asynq worker...")

	// Load config (reuse config yang sama)
	cfg := config.Load()

	// Redis address
	redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port

	// Create Asynq server
	srv := worker.NewAsynqServer(redisAddr, cfg.Redis.Password)

	// Create mux dengan semua handlers
	mux := worker.NewServeMux()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		log.Println("Shutting down worker...")
		srv.Shutdown()
	}()

	// Start processing
	log.Println("Asynq worker started. Waiting for tasks...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Failed to start Asynq server: %v", err)
		os.Exit(1)
	}
}
```

### Step 7: Wire Asynq Client ke main.go (API Server)

```go
// cmd/api/main.go — UPDATED

package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/irzan-aldi/task-manager/internal/config"
	"github.com/irzan-aldi/task-manager/internal/handler"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/repository"
	"github.com/irzan-aldi/task-manager/internal/router"
	"github.com/irzan-aldi/task-manager/internal/service"
	"github.com/irzan-aldi/task-manager/internal/worker" // BARU
)

func main() {
	cfg := config.Load()
	db := config.ConnectDB(cfg)

	if err := db.AutoMigrate(&model.User{}, &model.Board{}, &model.Task{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated successfully")

	redisClient := config.ConnectRedis(cfg)

	// ========== SETUP ASYNQ CLIENT ==========
	redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
	asynqClient := worker.NewAsynqClient(redisAddr, cfg.Redis.Password)
	defer asynqClient.Close() // Tutup koneksi saat shutdown
	log.Println("Asynq client initialized")
	// =========================================

	// Repositories
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Services — inject asynqClient
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, asynqClient)
	userService := service.NewUserService(userRepo)
	boardService := service.NewBoardService(boardRepo)
	taskService := service.NewTaskService(taskRepo, boardRepo, asynqClient)

	// Handlers
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

### Step 8: Jalankan

```bash
# Terminal 1 — API Server
go run cmd/api/main.go

# Terminal 2 — Worker
go run cmd/worker/main.go

# Terminal 3 — Test
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Irzan", "email": "irzan@mail.com", "password": "password123"}'
```

**Output di Terminal 1 (API):**
```
Server starting on port 8080 (env: development)
Enqueued task: id=abc123 queue=default
```

**Output di Terminal 2 (Worker):**
```
Asynq worker started. Waiting for tasks...
[Email Worker] Sending welcome email to Irzan (irzan@mail.com)...
[Email Worker] Welcome email sent to irzan@mail.com
```

### Step 9: Update Makefile

```makefile
# Makefile — tambahkan target baru

.PHONY: run worker monitor

# Jalankan API server
run:
	go run cmd/api/main.go

# Jalankan Asynq worker
worker:
	go run cmd/worker/main.go

# Jalankan Asynqmon web UI
monitor:
	asynqmon --redis-addr=localhost:6379

# Jalankan semuanya (development)
dev:
	@echo "Starting API server and worker..."
	@make run &
	@make worker &
	@wait
```

---

## 9.12 Bull (NestJS) vs Asynq (Go) — Side by Side

Ini perbandingan lengkap buat kamu yang migrasi dari NestJS Bull ke Go Asynq.

### 1. Setup Queue

**NestJS Bull:**
```typescript
// app.module.ts
@Module({
  imports: [
    BullModule.forRoot({
      redis: {
        host: 'localhost',
        port: 6379,
      },
    }),
    BullModule.registerQueue(
      { name: 'email' },
      { name: 'notification' },
    ),
  ],
})
export class AppModule {}
```

**Go Asynq:**
```go
// internal/worker/client.go
func NewAsynqClient(redisAddr, password string) *asynq.Client {
    return asynq.NewClient(asynq.RedisClientOpt{
        Addr:     redisAddr,
        Password: password,
    })
}

// internal/worker/server.go
func NewAsynqServer(redisAddr, password string) *asynq.Server {
    return asynq.NewServer(
        asynq.RedisClientOpt{Addr: redisAddr, Password: password},
        asynq.Config{Concurrency: 10},
    )
}
```

### 2. Define dan Enqueue Job/Task

**NestJS Bull:**
```typescript
// Inject queue
constructor(@InjectQueue('email') private emailQueue: Queue) {}

// Enqueue
await this.emailQueue.add('send-welcome', {
  userId: user.id,
  email: user.email,
}, {
  delay: 10000,
  attempts: 3,
  priority: 1,
  backoff: { type: 'exponential', delay: 5000 },
  removeOnComplete: true,
  timeout: 60000,
});
```

**Go Asynq:**
```go
// Bikin task
task, _ := worker.NewSendWelcomeEmailTask(worker.EmailWelcomePayload{
    UserID: user.ID,
    Email:  user.Email,
})

// Enqueue
client.Enqueue(task,
    asynq.ProcessIn(10 * time.Second),  // delay
    asynq.MaxRetry(3),                  // attempts
    asynq.Queue("default"),             // queue
    asynq.Timeout(60 * time.Second),    // timeout
)
```

### 3. Process Job/Task

**NestJS Bull:**
```typescript
@Processor('email')
export class EmailProcessor {
  @Process('send-welcome')
  async handleSendWelcome(job: Job<{ userId: number; email: string }>) {
    const { userId, email } = job.data;
    await this.mailService.send(email, 'Welcome!', 'template.html');
    return { sent: true }; // return value disimpan
  }

  @OnQueueActive()
  onActive(job: Job) {
    console.log(`Processing job ${job.id}...`);
  }

  @OnQueueCompleted()
  onCompleted(job: Job, result: any) {
    console.log(`Job ${job.id} completed`, result);
  }

  @OnQueueFailed()
  onFailed(job: Job, err: Error) {
    console.error(`Job ${job.id} failed: ${err.message}`);
  }
}
```

**Go Asynq:**
```go
// Handler function
func HandleSendWelcomeEmail(ctx context.Context, t *asynq.Task) error {
    var p EmailWelcomePayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return fmt.Errorf("unmarshal: %w: %w", err, asynq.SkipRetry)
    }

    if err := mailService.Send(p.Email, "Welcome!", "template.html"); err != nil {
        return err // auto retry
    }
    return nil // success
}

// Register (mirip @Processor + @Process)
mux := asynq.NewServeMux()
mux.HandleFunc("email:send-welcome", HandleSendWelcomeEmail)

// Error handler (mirip @OnQueueFailed)
asynq.Config{
    ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
        log.Printf("Task %s failed: %v", task.Type(), err)
    }),
}
```

### 4. Scheduled/Recurring Job

**NestJS Bull:**
```typescript
// Recurring job — every day at 9am
await this.emailQueue.add('daily-digest', {}, {
  repeat: {
    cron: '0 9 * * *',
    tz: 'Asia/Jakarta',
  },
});
```

**Go Asynq:**
```go
// Asynq Scheduler — separate component
scheduler := asynq.NewScheduler(
    asynq.RedisClientOpt{Addr: "localhost:6379"},
    &asynq.SchedulerOpts{
        Location: time.FixedZone("WIB", 7*60*60),
    },
)

task := asynq.NewTask("email:daily-digest", nil)
entryID, _ := scheduler.Register("0 9 * * *", task)
// Run scheduler
scheduler.Run()
```

### 5. Job Events / Lifecycle Hooks

**NestJS Bull:**
```typescript
// Event-based, pakai decorators
@OnQueueActive()    // job mulai diproses
@OnQueueCompleted() // job selesai
@OnQueueFailed()    // job gagal
@OnQueueStalled()   // job stuck
@OnQueueProgress()  // job progress update
```

**Go Asynq:**
```go
// Middleware-based
mux.Use(func(next asynq.Handler) asynq.Handler {
    return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
        start := time.Now()
        log.Printf("[Start] %s", t.Type())

        err := next.ProcessTask(ctx, t)

        if err != nil {
            log.Printf("[Failed] %s: %v (took %v)", t.Type(), err, time.Since(start))
        } else {
            log.Printf("[Done] %s (took %v)", t.Type(), time.Since(start))
        }
        return err
    })
})
```

### Ringkasan Mapping

| Konsep | NestJS Bull | Go Asynq |
|--------|------------|----------|
| Queue module | `BullModule.registerQueue()` | `asynq.NewClient()` |
| Inject queue | `@InjectQueue('name')` | Manual inject `*asynq.Client` |
| Add job | `queue.add('name', data, opts)` | `client.Enqueue(task, ...opts)` |
| Process job | `@Process('name')` decorator | `mux.HandleFunc(type, handler)` |
| Processor class | `@Processor('queue')` class | Plain handler function |
| Job data | `job.data` | `json.Unmarshal(t.Payload(), &p)` |
| Return value | `return result` | `return nil` (success) / `return err` (retry) |
| Lifecycle hooks | `@OnQueue*()` decorators | Middleware function |
| Scheduled | `repeat: { cron }` | `asynq.Scheduler` |
| Web UI | Bull Board / Arena | Asynqmon |

---

## 9.13 Best Practices

### 1. Idempotency — Task Bisa Dijalankan Ulang Tanpa Side Effect

Ini **PALING PENTING**. Karena Asynq bisa retry task, handler kamu harus **idempotent** — dijalankan 2x atau 10x hasilnya sama.

```go
// BURUK — tidak idempotent
func HandleSendWelcomeEmail(ctx context.Context, t *asynq.Task) error {
    var p EmailWelcomePayload
    json.Unmarshal(t.Payload(), &p)

    // Kirim email — kalau di-retry, user dapat email 2x!
    sendEmail(p.Email, "Welcome!")
    return nil
}

// BAGUS — idempotent
func HandleSendWelcomeEmail(ctx context.Context, t *asynq.Task) error {
    var p EmailWelcomePayload
    json.Unmarshal(t.Payload(), &p)

    // Cek dulu apakah email sudah pernah dikirim
    sent, _ := checkEmailSent(p.UserID, "welcome")
    if sent {
        log.Printf("Welcome email already sent to user %d, skipping", p.UserID)
        return nil // Skip — ga kirim ulang
    }

    // Kirim email
    if err := sendEmail(p.Email, "Welcome!"); err != nil {
        return err
    }

    // Tandai sudah dikirim
    markEmailSent(p.UserID, "welcome")
    return nil
}
```

**Tips idempotency:**
- Cek state sebelum action (apakah email sudah dikirim?)
- Pakai database flag / Redis key sebagai "already processed" marker
- Pakai `asynq.Unique()` untuk prevent duplicate enqueue
- Design handler yang bisa di-rerun tanpa efek samping

### 2. Dead Letter Queue (DLQ) — Handle Task yang Gagal Terus

Task yang udah max retry dan masih gagal masuk ke **Archived** state (Dead Letter Queue). Kamu harus monitor ini!

```go
// Di server config, set apa yang terjadi dengan exhausted task
asynq.Config{
    RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
        // Custom retry delay:
        // Retry 1: 10s, Retry 2: 30s, Retry 3: 60s, Retry 4: 120s, Retry 5: 300s
        delays := []time.Duration{
            10 * time.Second,
            30 * time.Second,
            1 * time.Minute,
            2 * time.Minute,
            5 * time.Minute,
        }
        if n-1 < len(delays) {
            return delays[n-1]
        }
        return 5 * time.Minute
    },
}
```

**Monitor DLQ:**
```bash
# Lihat task yang masuk DLQ
asynq task list --queue=default --state=archived

# Retry task yang ada di DLQ
asynq task run --queue=default --state=archived --id=<task_id>

# Retry semua task di DLQ
asynq task run-all --queue=default --state=archived
```

### 3. Monitoring dan Alerting

```go
// Middleware untuk logging dan metrics
func loggingMiddleware(next asynq.Handler) asynq.Handler {
    return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
        start := time.Now()
        taskID, _ := asynq.GetTaskID(ctx)
        retryCount, _ := asynq.GetRetryCount(ctx)
        maxRetry, _ := asynq.GetMaxRetry(ctx)

        log.Printf("[Task Start] id=%s type=%s retry=%d/%d",
            taskID, t.Type(), retryCount, maxRetry)

        err := next.ProcessTask(ctx, t)

        duration := time.Since(start)
        if err != nil {
            log.Printf("[Task Failed] id=%s type=%s error=%v duration=%v retry=%d/%d",
                taskID, t.Type(), err, duration, retryCount, maxRetry)

            // Alert kalau ini retry terakhir
            if retryCount >= maxRetry-1 {
                log.Printf("[ALERT] Task %s exhausted all retries! Moving to DLQ", taskID)
                // Kirim alert ke Slack, PagerDuty, dll
            }
        } else {
            log.Printf("[Task Done] id=%s type=%s duration=%v",
                taskID, t.Type(), duration)
        }

        return err
    })
}

// Register middleware
mux := asynq.NewServeMux()
mux.Use(loggingMiddleware)
mux.HandleFunc(worker.TaskSendWelcomeEmail, worker.HandleSendWelcomeEmail)
```

### 4. Graceful Shutdown

Saat worker di-shutdown (deploy, restart, scaling down), pastikan task yang sedang diproses selesai dulu — bukan langsung di-kill.

```go
// cmd/worker/main.go — graceful shutdown

func main() {
    cfg := config.Load()
    redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
    srv := worker.NewAsynqServer(redisAddr, cfg.Redis.Password)
    mux := worker.NewServeMux()

    // Graceful shutdown handler
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
        sig := <-sigCh
        log.Printf("Received signal %v, shutting down gracefully...", sig)
        srv.Shutdown() // Wait for active tasks to finish
    }()

    // Start
    if err := srv.Run(mux); err != nil {
        log.Fatalf("Worker failed: %v", err)
    }
}
```

`srv.Shutdown()` akan:
1. **Stop accepting** new tasks
2. **Wait** for currently processing tasks to finish
3. Then **exit** cleanly

Ini penting banget buat deployment — kalau kamu pakai Kubernetes, `SIGTERM` dikirim saat pod mau di-terminate.

### 5. Payload Size — Keep It Small

```go
// BURUK — payload terlalu besar
type ReportPayload struct {
    Data []map[string]interface{} `json:"data"` // Bisa jutaan row!
}

// BAGUS — payload cuma ID, worker query sendiri
type ReportPayload struct {
    BoardID uint   `json:"board_id"`
    UserID  uint   `json:"user_id"`
    Format  string `json:"format"`
}
```

Rule: **Payload cuma berisi ID dan parameter minimal.** Worker yang query data lengkapnya dari database. Kenapa?
- Redis memory itu mahal — jangan simpan data besar di queue
- Data bisa berubah antara enqueue dan process — kalau payload punya data lama, bisa stale
- Payload kecil = serialization/deserialization cepat

### 6. Separate Queue per Priority

```go
// Email reset password — harus cepat!
client.Enqueue(task, asynq.Queue("critical"))

// Welcome email — bisa nunggu
client.Enqueue(task, asynq.Queue("default"))

// Generate weekly report — ga urgent
client.Enqueue(task, asynq.Queue("low"))
```

Jangan campur task critical dan non-critical di satu queue. Kalau ada 1000 report task di queue, password reset email bisa ke-delay.

### 7. Health Check untuk Worker

```go
// internal/worker/health.go — health check endpoint di worker

package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

func CheckWorkerHealth(redisAddr, redisPassword string) error {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})
	defer inspector.Close()

	queues, err := inspector.Queues()
	if err != nil {
		return fmt.Errorf("failed to get queues: %w", err)
	}

	for _, q := range queues {
		info, err := inspector.GetQueueInfo(q)
		if err != nil {
			return fmt.Errorf("failed to get queue info for %s: %w", q, err)
		}

		// Alert kalau pending tasks terlalu banyak
		if info.Pending > 1000 {
			fmt.Printf("[WARNING] Queue %s has %d pending tasks\n", q, info.Pending)
		}

		// Alert kalau ada task di DLQ
		if info.Archived > 0 {
			fmt.Printf("[WARNING] Queue %s has %d archived (dead) tasks\n", q, info.Archived)
		}
	}

	return nil
}
```

### 8. Testing Task Handlers

Handler itu just functions — gampang di-test!

```go
// internal/worker/handler_email_test.go

package worker_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/irzan-aldi/task-manager/internal/worker"
)

func TestHandleSendWelcomeEmail(t *testing.T) {
	// Arrange — bikin task dengan payload
	payload := worker.EmailWelcomePayload{
		UserID: 1,
		Email:  "test@mail.com",
		Name:   "Test User",
	}
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(worker.TaskSendWelcomeEmail, data)

	// Act — process task
	err := worker.HandleSendWelcomeEmail(context.Background(), task)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestHandleSendWelcomeEmail_InvalidPayload(t *testing.T) {
	// Arrange — payload rusak
	task := asynq.NewTask(worker.TaskSendWelcomeEmail, []byte("invalid json"))

	// Act
	err := worker.HandleSendWelcomeEmail(context.Background(), task)

	// Assert — harus error
	if err == nil {
		t.Error("expected error for invalid payload")
	}
}
```

---

## 9.14 Advanced: Asynq Scheduler — Recurring Tasks

Selain one-off tasks, Asynq juga support **scheduled/recurring tasks** — mirip cron job.

```go
// cmd/scheduler/main.go

package main

import (
	"log"
	"time"

	"github.com/hibiken/asynq"
)

func main() {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		&asynq.SchedulerOpts{
			Location: loc,
		},
	)

	// Daily digest email — setiap hari jam 9 pagi WIB
	dailyDigest := asynq.NewTask("email:daily-digest", nil)
	_, err := scheduler.Register("0 9 * * *", dailyDigest,
		asynq.Queue("low"),
		asynq.MaxRetry(2),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Weekly report — setiap Senin jam 8 pagi
	weeklyReport := asynq.NewTask("report:weekly", nil)
	_, err = scheduler.Register("0 8 * * 1", weeklyReport,
		asynq.Queue("low"),
		asynq.MaxRetry(3),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Scheduler started")
	if err := scheduler.Run(); err != nil {
		log.Fatal(err)
	}
}
```

**NestJS Bull equivalent:**
```typescript
await this.queue.add('daily-digest', {}, {
  repeat: { cron: '0 9 * * *', tz: 'Asia/Jakarta' },
});
```

---

## 9.15 Architecture Overview — Semuanya Disatukan

Setelah bab ini, architecture project kita jadi:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client (Frontend)                        │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTP Request
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│                     API Server (cmd/api)                        │
│                                                                 │
│  Handler → Service → Repository → PostgreSQL                    │
│                ↓                                                │
│          Asynq Client ──enqueue──→ Redis                        │
│                                    │                            │
└────────────────────────────────────┼────────────────────────────┘
                                     │
                          ┌──────────┼──────────┐
                          │     Redis Server     │
                          │  ┌────────────────┐  │
                          │  │ Queue: critical│  │
                          │  │ Queue: default │  │
                          │  │ Queue: low     │  │
                          │  │ Cache data     │  │
                          │  │ Rate limit     │  │
                          │  └────────────────┘  │
                          └──────────┬───────────┘
                                     │
┌────────────────────────────────────┼────────────────────────────┐
│                    Worker (cmd/worker)                          │
│                                                                 │
│  Asynq Server ──poll──→ Redis                                  │
│       ↓                                                         │
│  ServeMux (router)                                              │
│    ├── "email:send-welcome"     → HandleSendWelcomeEmail()     │
│    ├── "notification:task-*"    → HandleTaskAssignedNotif()     │
│    └── "report:generate"        → HandleGenerateReport()        │
│                                                                 │
│  External Services:                                             │
│    ├── SMTP Server (email)                                      │
│    ├── FCM/APNs (push notification)                             │
│    └── Slack/Discord (webhook)                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Process terpisah:**
```bash
# 3 process yang jalan bersamaan
go run cmd/api/main.go       # API server — port 8080
go run cmd/worker/main.go    # Worker — process background tasks
asynqmon --redis-addr=:6379  # Monitoring UI — port 8080 (ganti port kalau clash)
```

---

## Checklist Bab Ini

- [ ] Paham konsep message queue: producer, consumer, queue
- [ ] Paham kapan pakai dan kapan TIDAK pakai message queue
- [ ] Paham perbedaan RabbitMQ vs Kafka vs Asynq
- [ ] Paham kenapa pilih Asynq (Redis-based, mirip Bull, simple)
- [ ] Paham architecture Asynq: Client → Redis → Server/Worker
- [ ] Install `github.com/hibiken/asynq` di project
- [ ] Bisa define task type, payload struct, dan constructor
- [ ] Bisa enqueue task dari service layer (producer)
- [ ] Bisa bikin handler function (consumer)
- [ ] Paham task options: delay, retry, timeout, unique, queue priority
- [ ] Setup `cmd/worker/main.go` sebagai worker entrypoint
- [ ] Wire Asynq client ke `cmd/api/main.go`
- [ ] Install dan jalankan Asynqmon web UI
- [ ] Paham best practices: idempotency, DLQ monitoring, graceful shutdown
- [ ] Paham mapping Bull (NestJS) ke Asynq (Go)
- [ ] Test: register user → cek worker log ada "welcome email sent"
- [ ] Baca https://github.com/hibiken/asynq/wiki — self-study
