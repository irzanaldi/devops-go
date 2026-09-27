# Bab 18: Microservices Pattern — Distributed System dari Monolith ke Microservices

## Kenapa Bab Ini Penting?

Ini bab yang paling **konseptual** di seluruh ebook ini. Bab-bab sebelumnya fokus ke tools — Docker, CI/CD, AWS, Kubernetes, Terraform. Bab ini fokus ke **arsitektur** — gimana kamu design system yang bisa scale, maintain, dan evolve seiring waktu.

Di Bab 16, kamu udah belajar Kubernetes — orchestrate containers. Di Bab 17, Terraform — infrastructure as code. Sekarang pertanyaannya: **containers apa yang mau di-orchestrate?** Jawaban populernya: microservices.

Tapi — dan ini penting — microservices **bukan silver bullet**. Bukan jawaban untuk semua masalah. Bab ini bakal jujur soal trade-offs, kapan pakai, kapan JANGAN pakai, dan gimana migrate kalau memang waktunya tiba.

```
Sebelum bab ini:
- Task Manager API = 1 monolith Go app
- Semua fitur (auth, user, board, task) dalam 1 codebase
- 1 database PostgreSQL, 1 Redis
- Deploy sebagai 1 container
- Scale = scale semuanya

Setelah bab ini:
- Paham kapan monolith cukup dan kapan perlu microservices
- Bisa design microservices architecture
- Paham communication patterns (REST, gRPC, message queue)
- Bisa implement gRPC service di Go
- Paham API Gateway, Service Discovery, Circuit Breaker
- Paham Saga Pattern, Event-Driven Architecture
- Bisa setup distributed tracing dan logging
- Punya migration strategy dari monolith ke microservices
- Tau anti-patterns yang harus dihindari
```

> **Analogi gampang:** Bayangin warung makan. Satu tukang masak yang bisa masak semua — nasi goreng, mie ayam, bakso, es teh. Itu **monolith**. Sekarang bayangin **food court** — ada stall nasi goreng, stall mie ayam, stall bakso, stall minuman. Setiap stall spesialisasi, punya koki sendiri, bisa scale sendiri (stall bakso rame? tambah koki!). Itu **microservices**. Tapi food court butuh manajemen lebih — siapa yang atur antrian? Gimana kalau stall bakso tutup, stall mie masih buka? Gimana customer bayar kalau setiap stall punya kasir sendiri?

---

## 18.1 Apa Itu Microservices?

### Definisi Simple

**Microservices** itu arsitektur dimana satu aplikasi dipecah jadi **beberapa service kecil** yang masing-masing:
- Punya **satu tanggung jawab** (single responsibility)
- Bisa di-**deploy sendiri** (independently deployable)
- Punya **database sendiri** (decentralized data)
- **Berkomunikasi** via network (HTTP, gRPC, message queue)
- Ditulis oleh **tim terpisah** (bisa bahasa berbeda)

```
Monolith Architecture:
━━━━━━━━━━━━━━━━━━━━
┌─────────────────────────────────────────────────┐
│                 TASK MANAGER API                  │
│                                                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │   Auth    │  │   User   │  │  Board   │       │
│  │  Handler  │  │  Handler │  │  Handler │       │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘       │
│        │              │              │             │
│  ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐       │
│  │   Auth   │  │   User   │  │  Board   │       │
│  │  Service │  │  Service │  │  Service │       │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘       │
│        │              │              │             │
│  ┌─────▼──────────────▼──────────────▼─────┐     │
│  │           SHARED DATABASE                │     │
│  │           (PostgreSQL)                   │     │
│  └──────────────────────────────────────────┘     │
│                                                   │
│  1 codebase, 1 deployment, 1 database             │
└─────────────────────────────────────────────────┘


Microservices Architecture:
━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│   Auth   │  │   User   │  │  Board   │  │   Task   │  │  Notif   │
│  Service │  │  Service │  │  Service │  │  Service │  │  Service │
│          │  │          │  │          │  │          │  │          │
│  :3001   │  │  :3002   │  │  :3003   │  │  :3004   │  │  :3005   │
│          │  │          │  │          │  │          │  │          │
│ ┌──────┐ │  │ ┌──────┐ │  │ ┌──────┐ │  │ ┌──────┐ │  │ ┌──────┐ │
│ │ DB 1 │ │  │ │ DB 2 │ │  │ │ DB 3 │ │  │ │ DB 4 │ │  │ │Redis │ │
│ └──────┘ │  │ └──────┘ │  │ └──────┘ │  │ └──────┘ │  │ └──────┘ │
└──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘
     ↑              ↑              ↑              ↑              ↑
     └──────────────┴──────────────┴──────────────┴──────────────┘
                              │
                    ┌─────────▼─────────┐
                    │   API GATEWAY      │
                    │   :8080            │
                    │   Routing + Auth   │
                    │   + Rate Limiting  │
                    └─────────┬─────────┘
                              │
                         ┌────▼────┐
                         │ Client  │
                         └─────────┘

5 codebases, 5 deployments, 5 databases
```

### Kenapa Microservices Populer?

1. **Independent scaling** — Task Service rame? Scale Task Service aja, ga perlu scale Auth Service
2. **Independent deployment** — Update Notification Service tanpa deploy ulang Board Service
3. **Technology freedom** — Auth Service pakai Go, Notification Service pakai Python? Bisa
4. **Team autonomy** — Tim A handle Auth, Tim B handle Task. Ga saling ganggu
5. **Fault isolation** — Notification Service down? Task Service tetap jalan
6. **Easier to understand** — Setiap service kecil dan fokus. Lebih gampang dipahami 1 developer

### Kenapa BUKAN Selalu Jawaban yang Benar?

> **"If you can't build a well-structured monolith, what makes you think microservices is the answer?"** — Simon Brown

Ini yang sering dilupain: microservices itu **bukan upgrade dari monolith**. Microservices itu **trade-off**. Kamu tukar **complexity di code** dengan **complexity di infrastructure**.

```
Monolith complexity:
━━━━━━━━━━━━━━━━━━
- Codebase gede
- Coupling antar module
- Deploy semuanya bareng
- Scale semuanya bareng

Microservices complexity:
━━━━━━━━━━━━━━━━━━━━━━━━
- Network calls bisa fail (latency, timeout, partial failure)
- Distributed transactions (saga, compensation)
- Service discovery (gimana service A nemuin service B?)
- Distributed tracing (request lewat 5 service, error di mana?)
- Data consistency (eventual consistency, bukan ACID)
- Deployment complexity (10 services = 10 CI/CD pipelines)
- Monitoring complexity (10 services = 10 dashboards)
- Testing complexity (integration test across services)
```

### Monolith First

**Martin Fowler** (salah satu founding father software architecture) bilang:

> "Almost all the successful microservice stories have started with a monolith that got too big and was broken up."

Artinya: **mulai dari monolith dulu**. Kalau monolith udah terlalu besar dan painful, baru migrate ke microservices. Jangan mulai dari microservices.

```
Project baru → Monolith ✅
MVP / prototype → Monolith ✅
Tim < 5 orang → Monolith ✅
Belum tau domain boundaries → Monolith ✅
Udah product-market fit, tim growing → Consider microservices 🤔
Traffic pattern jelas, bottleneck teridentifikasi → Microservices ✅
```

---

## 18.2 Monolith vs Microservices — Detailed Comparison

### Comparison Table

| Aspek | Monolith | Microservices |
|-------|----------|---------------|
| **Deployment** | 1 deployment unit | Setiap service deploy sendiri |
| **Scaling** | Scale semuanya | Scale per service |
| **Codebase** | 1 repo (atau mono-repo) | Multi-repo atau mono-repo |
| **Database** | 1 shared database | Database per service |
| **Communication** | Function call (in-process) | Network call (HTTP/gRPC/queue) |
| **Latency** | Microseconds (function call) | Milliseconds (network call) |
| **Data consistency** | ACID transactions | Eventual consistency |
| **Team structure** | 1 tim besar | Banyak tim kecil (2-pizza teams) |
| **Technology** | 1 tech stack | Polyglot (bisa beda bahasa) |
| **Testing** | Unit + integration (simple) | Contract + integration + e2e (complex) |
| **Debugging** | Stack trace langsung | Distributed tracing (Jaeger, Zipkin) |
| **Error handling** | Try/catch | Circuit breaker, retry, timeout |
| **Complexity awal** | Rendah | Tinggi |
| **Complexity seiring grow** | Naik drastis | Terkontrol per service |
| **DevOps maturity needed** | Low-medium | High |
| **Cost** | Lebih murah (awal) | Lebih mahal (infrastructure overhead) |

### Kapan Monolith Lebih Baik?

```
Monolith wins kalau:
━━━━━━━━━━━━━━━━━━━

1. Tim kecil (1-5 developer)
   → Microservices butuh DevOps maturity. Tim kecil ga punya bandwidth.

2. Startup awal / MVP
   → Kamu belum tau fitur mana yang bakal survive. Jangan invest di architecture.

3. Domain belum clear
   → Kalau kamu salah pecah service, merge lagi itu PAINFUL.

4. Simple CRUD app
   → Ga semua app butuh microservices. Blog? E-commerce kecil? Monolith cukup.

5. Tight deadline
   → Microservices = lebih banyak setup awal. Monolith = langsung ngoding fitur.

6. Budget terbatas
   → 1 server monolith vs 10+ containers microservices. Beda cost.
```

### Kapan Migrate ke Microservices?

```
Sinyal butuh microservices:
━━━━━━━━━━━━━━━━━━━━━━━━━

1. Deployment jadi bottleneck
   → "Kita ga bisa deploy fitur notifikasi karena fitur payment belum ready"
   → Service yang bisa deploy sendiri = unblocked

2. Scaling ga efisien
   → "Auth endpoint kena 10x traffic, tapi kita harus scale SEMUA app"
   → Scale per service = cost efficient

3. Tim mulai saling block
   → "Tim A merge conflict terus sama Tim B"
   → Repo terpisah = ga ada conflict

4. Codebase terlalu besar
   → "Onboarding developer baru butuh 2 minggu buat paham codebase"
   → Service kecil = 1-2 hari paham

5. Different scaling needs
   → "Report generation butuh banyak CPU tapi auth ga"
   → Separate service = optimize per service

6. Technology constraints
   → "ML model kita harus pakai Python tapi backend pakai Go"
   → Polyglot = setiap service bahasa yang cocok
```

---

## 18.3 Microservices Principles

### 1. Single Responsibility

Setiap service punya **satu tanggung jawab yang jelas**. Ini bukan berarti 1 endpoint, tapi 1 **domain/business capability**.

```
✅ Correct service boundaries:
- Auth Service → handle authentication & authorization
- User Service → handle user profiles, CRUD
- Task Service → handle tasks CRUD, assignment
- Notification Service → handle email, push notification

❌ Wrong service boundaries (terlalu kecil):
- Login Service → cuma handle login
- Register Service → cuma handle register
- Password Service → cuma handle password reset
→ Ini NANO-SERVICES. Anti-pattern. Terlalu banyak overhead.

❌ Wrong service boundaries (terlalu besar):
- Backend Service → handle everything
→ Ini cuma monolith yang kamu rename jadi "service"
```

### 2. Independently Deployable

Setiap service bisa di-deploy **tanpa deploy service lain**. Ini berarti:
- Service punya **CI/CD pipeline sendiri**
- Service punya **Docker image sendiri**
- Service punya **version sendiri**
- Perubahan di Service A **ga butuh redeploy** Service B

```
Deployment independence:
━━━━━━━━━━━━━━━━━━━━━━

Monolith:
  Change auth logic → deploy EVERYTHING → risk breaking task feature

Microservices:
  Change auth logic → deploy ONLY auth service → task service ga kena
```

### 3. Decentralized Data Management

Setiap service **own** datanya sendiri. Ga ada shared database.

```
✅ Database per service:
┌──────────┐     ┌──────────┐     ┌──────────┐
│   Auth   │     │   Task   │     │  Notif   │
│  Service │     │  Service │     │  Service │
│          │     │          │     │          │
│  ┌────┐  │     │  ┌────┐  │     │  ┌────┐  │
│  │Auth│  │     │  │Task│  │     │  │Noti│  │
│  │ DB │  │     │  │ DB │  │     │  │f DB│  │
│  └────┘  │     │  └────┘  │     │  └────┘  │
└──────────┘     └──────────┘     └──────────┘

❌ Shared database (anti-pattern):
┌──────────┐     ┌──────────┐     ┌──────────┐
│   Auth   │     │   Task   │     │  Notif   │
│  Service │     │  Service │     │  Service │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     └────────────────┼────────────────┘
                      │
                ┌─────▼─────┐
                │ SHARED DB │  ← Coupling! Semua service terikat ke 1 schema
                └───────────┘
```

### 4. Design for Failure

Dalam monolith, function call hampir pasti succeed. Dalam microservices, **network call bisa gagal kapan aja**. Jadi kamu harus **expect failure**.

```
Monolith:
  result = userService.GetByID(123)  // Function call → almost always works

Microservices:
  result = HTTP GET user-service:3002/users/123
  → Network timeout? ← BISA
  → Service down? ← BISA
  → Response corrupt? ← BISA
  → 500 error? ← BISA
  → DNS resolution fail? ← BISA

Solusi:
  → Circuit Breaker (jangan terus panggil service yang down)
  → Retry with backoff (coba lagi, tapi jangan buru-buru)
  → Timeout (jangan nunggu selamanya)
  → Fallback (kalau gagal, kasih default response)
  → Graceful degradation (fitur A down, fitur B masih jalan)
```

### 5. Smart Endpoints, Dumb Pipes

Logika bisnis ada di **service** (smart endpoints), bukan di communication layer (dumb pipes). Message queue atau HTTP cuma transport — ga ada business logic di sana.

```
✅ Smart endpoints, dumb pipes:
  - Service A process business logic
  - Service A publish event ke message queue (dumb pipe)
  - Service B consume event dan process business logic
  → Queue cuma transport. Ga tau isi message-nya apa.

❌ Smart pipes:
  - ESB (Enterprise Service Bus) yang transform, route, dan apply logic
  → Complex, hard to debug, single point of failure
```

---

## 18.4 Task Manager as Microservices

Sekarang kita breakdown Task Manager API kita yang monolith jadi microservices. Ini cara kita pecah berdasarkan **domain boundary**.

### Current Monolith Structure

```
task-manager/
├── cmd/api/main.go           ← Single entry point
├── internal/
│   ├── config/               ← Shared config
│   ├── handler/
│   │   ├── auth.go           ← Auth endpoints
│   │   ├── user.go           ← User endpoints
│   │   ├── board.go          ← Board endpoints
│   │   └── task.go           ← Task endpoints
│   ├── middleware/
│   │   ├── auth.go           ← JWT middleware
│   │   └── ratelimit.go      ← Rate limiter
│   ├── model/
│   │   ├── user.go           ← User model
│   │   ├── board.go          ← Board model
│   │   └── task.go           ← Task model
│   ├── repository/
│   │   ├── user.go
│   │   ├── board.go
│   │   └── task.go
│   ├── router/router.go      ← Single router
│   └── service/
│       ├── auth.go
│       ├── user.go
│       ├── board.go
│       └── task.go
└── pkg/response/json.go

1 app, 1 DB, semua di-deploy bareng.
```

### Microservices Breakdown

#### 1. Auth Service

```
Responsibility: Authentication & Authorization
Endpoints:
  POST /auth/register   → Register user baru
  POST /auth/login      → Login, return JWT
  POST /auth/refresh    → Refresh token
  POST /auth/validate   → Validate JWT (internal, dipanggil service lain)

Database: auth_db (users table — email, password_hash, role)
Port: 3001
Communication: REST + gRPC (validate token di-call service lain)
```

#### 2. User Service

```
Responsibility: User profile management
Endpoints:
  GET    /users/profile  → Get own profile
  PUT    /users/profile  → Update profile
  GET    /users          → List users (admin)
  GET    /users/:id      → Get user by ID (internal)
  DELETE /users/:id      → Delete user (admin)

Database: user_db (users table — name, email, avatar, preferences)
Port: 3002
Communication: REST + gRPC
Events:
  - Listens: user.registered (dari Auth Service)
  - Publishes: user.updated, user.deleted
```

#### 3. Board Service

```
Responsibility: Board/project management
Endpoints:
  POST   /boards         → Create board
  GET    /boards         → List my boards
  GET    /boards/:id     → Get board detail
  PUT    /boards/:id     → Update board
  DELETE /boards/:id     → Delete board

Database: board_db (boards table)
Port: 3003
Communication: REST + gRPC
Events:
  - Publishes: board.created, board.deleted
  - Listens: user.deleted (cascade delete boards)
```

#### 4. Task Service

```
Responsibility: Task management
Endpoints:
  POST   /boards/:boardId/tasks  → Create task
  GET    /boards/:boardId/tasks  → List tasks by board
  GET    /tasks/my               → My assigned tasks
  GET    /tasks/:id              → Get task detail
  PUT    /tasks/:id              → Update task
  DELETE /tasks/:id              → Delete task

Database: task_db (tasks table)
Port: 3004
Communication: REST + gRPC
Events:
  - Publishes: task.created, task.updated, task.assigned, task.completed
  - Listens: board.deleted (cascade delete tasks)
```

#### 5. Notification Service

```
Responsibility: Send notifications
Endpoints:
  (Mostly event-driven, minimal REST)
  GET /notifications    → List user notifications
  PUT /notifications/:id/read → Mark as read

Database: notification_db (notifications table) atau Redis
Port: 3005
Communication: Event-driven (consumer)
Events:
  - Listens: task.assigned, task.completed, user.registered
  - Actions: Send email, push notification, in-app notification
```

#### 6. API Gateway

```
Responsibility: Single entry point for all clients
Functions:
  - Route requests ke service yang tepat
  - Authentication (validate JWT)
  - Rate limiting
  - Load balancing
  - Request/response transformation
  - Circuit breaking

Port: 8080
Communication: REST (external), gRPC (internal ke services)
```

### Architecture Diagram

```
                            ┌─────────────────┐
                            │     Clients      │
                            │  (Web, Mobile)   │
                            └────────┬────────┘
                                     │
                                     ▼
                          ┌──────────────────────┐
                          │    API GATEWAY        │
                          │    :8080              │
                          │                       │
                          │  ┌─────────────────┐  │
                          │  │ Auth Middleware  │  │
                          │  │ Rate Limiter    │  │
                          │  │ Router          │  │
                          │  │ Circuit Breaker │  │
                          │  └─────────────────┘  │
                          └──────────┬────────────┘
                                     │
              ┌──────────┬──────────┬┴──────────┬──────────┐
              │          │          │            │          │
              ▼          ▼          ▼            ▼          ▼
        ┌──────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
        │   Auth   │ │  User  │ │ Board  │ │  Task  │ │ Notif  │
        │ Service  │ │Service │ │Service │ │Service │ │Service │
        │  :3001   │ │ :3002  │ │ :3003  │ │ :3004  │ │ :3005  │
        └────┬─────┘ └───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘
             │           │         │           │          │
             ▼           ▼         ▼           ▼          ▼
        ┌────────┐  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
        │auth_db │  │user_db │ │board_db│ │task_db │ │  Redis │
        │(Postgres)│ │(Postgres)││(Postgres)││(Postgres)│ │        │
        └────────┘  └────────┘ └────────┘ └────────┘ └────────┘

                          ┌──────────────────┐
                          │   Message Broker  │
                          │   (NATS / Kafka)  │
                          │                   │
                          │  Events:          │
                          │  - task.created   │
                          │  - task.assigned  │
                          │  - user.registered│
                          │  - board.deleted  │
                          └──────────────────┘
```

### Perbandingan dengan NestJS

Kalau kamu dari NestJS, breakdown-nya mirip banget:

```typescript
// NestJS monolith (1 app, banyak modules)
@Module({
  imports: [
    AuthModule,
    UserModule,
    BoardModule,
    TaskModule,
    NotificationModule,
  ],
})
export class AppModule {}

// NestJS microservices (banyak apps)
// apps/auth-service/src/main.ts
const app = await NestFactory.createMicroservice(AuthModule, {
  transport: Transport.TCP,
  options: { port: 3001 },
});

// apps/task-service/src/main.ts
const app = await NestFactory.createMicroservice(TaskModule, {
  transport: Transport.TCP,
  options: { port: 3004 },
});
```

Di Go, kita ga punya framework "microservice" built-in kayak NestJS. Tapi justru itu **kelebihan Go** — kamu build exactly what you need, ga ada magic, semuanya explicit.

---

## 18.5 Communication Patterns

### Synchronous vs Asynchronous

Dua pertanyaan penting saat service A butuh ngomong ke service B:

1. **Apakah service A perlu response sekarang?** → Synchronous
2. **Apakah service A cukup "fire and forget"?** → Asynchronous

```
Synchronous (request-response):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Service A ──request──→ Service B
Service A ←─response── Service B
Service A NUNGGU sampai Service B selesai.

Use case:
- API Gateway → Auth Service: "Validate this JWT" (perlu jawaban SEKARANG)
- Task Service → User Service: "Get user name for task assignee" (perlu data)
- Board Service → Auth Service: "Is this user authorized?" (perlu yes/no)


Asynchronous (event-driven):
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Service A ──publish event──→ Message Broker ──deliver──→ Service B
Service A LANJUT KERJA. Ga nunggu.

Use case:
- Task Service → "task.created" → Notification Service (kirim email)
- Auth Service → "user.registered" → User Service (create profile)
- Board Service → "board.deleted" → Task Service (delete tasks)
```

### Synchronous Options

#### 1. REST (HTTP/JSON)

Yang paling familiar. Kamu udah pakai ini di Task Manager API.

```go
// Task Service call User Service via REST
func (s *TaskService) getUserName(userID uint) (string, error) {
    url := fmt.Sprintf("http://user-service:3002/users/%d", userID)

    resp, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("failed to call user service: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("user service returned %d", resp.StatusCode)
    }

    var user struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return "", fmt.Errorf("failed to decode response: %w", err)
    }

    return user.Name, nil
}
```

**Pros:** Simple, familiar, human-readable (JSON), tooling bagus (Postman, curl).
**Cons:** Slower (JSON serialize/deserialize), no type safety across services, no streaming.

#### 2. gRPC (Protocol Buffers)

Binary protocol, lebih cepat dari REST. Pakai **Protocol Buffers** (protobuf) untuk define contract.

```protobuf
// user.proto
syntax = "proto3";
package user;

service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
}

message GetUserRequest {
    uint64 id = 1;
}

message GetUserResponse {
    uint64 id = 1;
    string name = 2;
    string email = 3;
}
```

**Pros:** Faster (binary), type-safe, streaming, auto-generate client code.
**Cons:** Not human-readable, harder to debug, learning curve.

### Asynchronous Options

#### 1. Message Queue (RabbitMQ)

Traditional message broker. Produce message, consumer ambil dari queue.

```
Producer → Queue → Consumer
```

#### 2. Event Streaming (Kafka)

Distributed event log. Events di-persist dan bisa di-replay.

```
Producer → Topic (partitioned log) → Consumer Group
```

#### 3. Lightweight Messaging (NATS)

Cloud-native messaging. Super ringan dan cepat.

```
Publisher → Subject → Subscriber
```

### Comparison Table

| Aspek | REST | gRPC | RabbitMQ | Kafka | NATS |
|-------|------|------|----------|-------|------|
| **Type** | Sync | Sync | Async | Async | Async |
| **Protocol** | HTTP/1.1 + JSON | HTTP/2 + Protobuf | AMQP | Custom binary | Custom |
| **Speed** | Moderate | Fast | Moderate | Very fast | Very fast |
| **Latency** | ~1-10ms | ~0.1-1ms | ~1-5ms | ~2-10ms | ~0.1-1ms |
| **Type safety** | No (runtime) | Yes (compile time) | No | No | No |
| **Streaming** | No (SSE workaround) | Yes (bi-directional) | No | Yes (consumer) | Yes |
| **Human readable** | Yes (JSON) | No (binary) | Depends | No | Depends |
| **Persistence** | No | No | Optional | Yes (append log) | Optional (JetStream) |
| **Use case** | External API | Internal service-to-service | Background jobs | Event streaming | Lightweight pub/sub |
| **Complexity** | Low | Medium | Medium | High | Low |
| **Go library** | net/http | google.golang.org/grpc | amqp091-go | confluent-kafka-go | nats-io/nats.go |

### Kapan Pakai Apa?

```
REST:
  → External API (client-facing)
  → Simple CRUD calls
  → Kamu mau bisa debug pakai curl/Postman
  → Tim belum familiar gRPC

gRPC:
  → Internal service-to-service communication
  → High throughput, low latency needed
  → Streaming needed
  → Strong type safety between services

RabbitMQ:
  → Complex routing (topic, fanout)
  → Task queue dengan retry
  → Traditional enterprise messaging

Kafka:
  → Event streaming (jutaan events/sec)
  → Event sourcing, CQRS
  → Data pipeline, log aggregation
  → Event replay needed

NATS:
  → Lightweight pub/sub
  → Cloud-native, Kubernetes-friendly
  → Simple messaging tanpa complexity Kafka
  → Real-time communication
```

### Practical Rule

Untuk project Task Manager kita:

```
External (client → API Gateway):  REST (JSON)
Internal (service → service):     gRPC (protobuf)
Events (service → service async): NATS (lightweight, perfect for our scale)
```

---

## 18.6 gRPC — Remote Procedure Call di Go

### Apa Itu gRPC?

gRPC itu **framework RPC** (Remote Procedure Call) yang dibuat Google. "g" itu bisa berarti Google, gRPC, good, great — Google suka ganti artinya setiap release.

Intinya: kamu bisa **panggil function di service lain seolah-olah panggil function lokal**. Di belakang layar, gRPC yang handle serialization, network call, deserialization.

```
Tanpa gRPC:
  1. Serialize request ke JSON
  2. HTTP POST ke service lain
  3. Service lain deserialize JSON
  4. Process
  5. Serialize response ke JSON
  6. Return HTTP response
  7. Deserialize JSON response

Dengan gRPC:
  1. Call function → grpc handle SEMUA di atas

  result, err := userClient.GetUser(ctx, &pb.GetUserRequest{Id: 123})
  // Seolah-olah panggil function lokal, padahal ini network call
```

### Protocol Buffers (Protobuf)

gRPC pakai **Protocol Buffers** sebagai interface definition language (IDL) dan serialization format.

```
JSON:  {"id": 123, "name": "Irzan", "email": "irzan@test.com"}
       → ~55 bytes, human readable, slow to parse

Protobuf: [binary encoded]
          → ~20 bytes, not human readable, FAST to parse
```

### Setup gRPC di Go

#### Step 1: Install Tools

```bash
# Install protoc compiler
# macOS
brew install protobuf

# Linux
apt install -y protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Pastikan $GOPATH/bin ada di PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

#### Step 2: Define Proto File

Buat file proto untuk User Service:

```protobuf
// proto/user/user.proto
syntax = "proto3";

package user;

option go_package = "github.com/irzan-aldi/task-manager/proto/user";

// Service definition
service UserService {
    // Get user by ID
    rpc GetUser(GetUserRequest) returns (GetUserResponse);

    // Get multiple users by IDs
    rpc GetUsers(GetUsersRequest) returns (GetUsersResponse);

    // Validate JWT token (called by other services)
    rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}

// Messages
message GetUserRequest {
    uint64 id = 1;
}

message GetUserResponse {
    uint64 id = 1;
    string name = 2;
    string email = 3;
    string role = 4;
}

message GetUsersRequest {
    repeated uint64 ids = 1;
}

message GetUsersResponse {
    repeated GetUserResponse users = 1;
}

message ValidateTokenRequest {
    string token = 1;
}

message ValidateTokenResponse {
    bool valid = 1;
    uint64 user_id = 2;
    string email = 3;
    string role = 4;
}
```

#### Step 3: Generate Go Code

```bash
# Generate Go code dari proto file
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/user/user.proto
```

Ini bakal generate 2 file:
- `proto/user/user.pb.go` — message types (struct Go)
- `proto/user/user_grpc.pb.go` — gRPC service interface dan client

#### Step 4: Implement gRPC Server (User Service)

```go
// user-service/internal/grpc/server.go
package grpc

import (
    "context"
    "fmt"
    "log"
    "net"

    pb "github.com/irzan-aldi/task-manager/proto/user"
    "github.com/irzan-aldi/task-manager/user-service/internal/service"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type UserGRPCServer struct {
    pb.UnimplementedUserServiceServer
    userService *service.UserService
}

func NewUserGRPCServer(userService *service.UserService) *UserGRPCServer {
    return &UserGRPCServer{
        userService: userService,
    }
}

// GetUser implements UserServiceServer
func (s *UserGRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    user, err := s.userService.GetByID(uint(req.Id))
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
    }

    return &pb.GetUserResponse{
        Id:    uint64(user.ID),
        Name:  user.Name,
        Email: user.Email,
        Role:  string(user.Role),
    }, nil
}

// GetUsers implements UserServiceServer
func (s *UserGRPCServer) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
    var users []*pb.GetUserResponse

    for _, id := range req.Ids {
        user, err := s.userService.GetByID(uint(id))
        if err != nil {
            continue // Skip not found users
        }
        users = append(users, &pb.GetUserResponse{
            Id:    uint64(user.ID),
            Name:  user.Name,
            Email: user.Email,
            Role:  string(user.Role),
        })
    }

    return &pb.GetUsersResponse{Users: users}, nil
}

// ValidateToken implements UserServiceServer
func (s *UserGRPCServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
    claims, err := s.userService.ValidateToken(req.Token)
    if err != nil {
        return &pb.ValidateTokenResponse{Valid: false}, nil
    }

    return &pb.ValidateTokenResponse{
        Valid:  true,
        UserId: uint64(claims.UserID),
        Email:  claims.Email,
        Role:   claims.Role,
    }, nil
}

// StartGRPCServer starts the gRPC server
func StartGRPCServer(port string, userService *service.UserService) error {
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        return fmt.Errorf("failed to listen: %w", err)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterUserServiceServer(grpcServer, NewUserGRPCServer(userService))

    log.Printf("gRPC server listening on port %s", port)
    return grpcServer.Serve(lis)
}
```

#### Step 5: Implement gRPC Client (Task Service calling User Service)

```go
// task-service/internal/client/user_client.go
package client

import (
    "context"
    "fmt"
    "time"

    pb "github.com/irzan-aldi/task-manager/proto/user"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
    conn   *grpc.ClientConn
    client pb.UserServiceClient
}

func NewUserClient(address string) (*UserClient, error) {
    // Connect ke User Service gRPC
    conn, err := grpc.NewClient(address,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to connect to user service: %w", err)
    }

    client := pb.NewUserServiceClient(conn)
    return &UserClient{conn: conn, client: client}, nil
}

func (c *UserClient) Close() error {
    return c.conn.Close()
}

// GetUser calls User Service via gRPC
func (c *UserClient) GetUser(ctx context.Context, userID uint) (*pb.GetUserResponse, error) {
    // Set timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{
        Id: uint64(userID),
    })
    if err != nil {
        return nil, fmt.Errorf("gRPC GetUser failed: %w", err)
    }

    return resp, nil
}

// GetUsers calls User Service for multiple users
func (c *UserClient) GetUsers(ctx context.Context, userIDs []uint) ([]*pb.GetUserResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    ids := make([]uint64, len(userIDs))
    for i, id := range userIDs {
        ids[i] = uint64(id)
    }

    resp, err := c.client.GetUsers(ctx, &pb.GetUsersRequest{
        Ids: ids,
    })
    if err != nil {
        return nil, fmt.Errorf("gRPC GetUsers failed: %w", err)
    }

    return resp.Users, nil
}

// ValidateToken validates JWT token via Auth/User Service
func (c *UserClient) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
        Token: token,
    })
    if err != nil {
        return nil, fmt.Errorf("gRPC ValidateToken failed: %w", err)
    }

    return resp, nil
}
```

#### Step 6: Use gRPC Client di Task Service

```go
// task-service/internal/service/task.go
package service

import (
    "context"
    "fmt"

    "github.com/irzan-aldi/task-manager/task-service/internal/client"
    "github.com/irzan-aldi/task-manager/task-service/internal/model"
    "github.com/irzan-aldi/task-manager/task-service/internal/repository"
)

type TaskService struct {
    repo       *repository.TaskRepository
    userClient *client.UserClient
}

func NewTaskService(repo *repository.TaskRepository, userClient *client.UserClient) *TaskService {
    return &TaskService{
        repo:       repo,
        userClient: userClient,
    }
}

func (s *TaskService) CreateWithAssignee(ctx context.Context, req model.CreateTaskRequest) (*model.TaskWithAssignee, error) {
    // Create task
    task, err := s.repo.Create(&model.Task{
        Title:       req.Title,
        Description: req.Description,
        BoardID:     req.BoardID,
        AssigneeID:  req.AssigneeID,
        Priority:    req.Priority,
        Status:      "todo",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create task: %w", err)
    }

    result := &model.TaskWithAssignee{Task: *task}

    // Kalau ada assignee, fetch user data via gRPC
    if task.AssigneeID != nil {
        userResp, err := s.userClient.GetUser(ctx, *task.AssigneeID)
        if err != nil {
            // Graceful degradation — task tetap dibuat, assignee info aja yang kosong
            fmt.Printf("Warning: failed to get assignee info: %v\n", err)
        } else {
            result.AssigneeName = userResp.Name
            result.AssigneeEmail = userResp.Email
        }
    }

    return result, nil
}
```

#### Step 7: Start Both Services

```go
// user-service/cmd/main.go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/irzan-aldi/task-manager/user-service/internal/config"
    grpcserver "github.com/irzan-aldi/task-manager/user-service/internal/grpc"
    "github.com/irzan-aldi/task-manager/user-service/internal/repository"
    "github.com/irzan-aldi/task-manager/user-service/internal/service"
)

func main() {
    cfg := config.Load()
    db := config.ConnectDB(cfg)

    userRepo := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo, cfg.JWTSecret)

    // Start gRPC server
    go func() {
        if err := grpcserver.StartGRPCServer(cfg.GRPCPort, userService); err != nil {
            log.Fatalf("gRPC server failed: %v", err)
        }
    }()

    log.Printf("User Service started (gRPC: %s)", cfg.GRPCPort)

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down User Service...")
}
```

```go
// task-service/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/task-service/internal/client"
    "github.com/irzan-aldi/task-manager/task-service/internal/config"
    "github.com/irzan-aldi/task-manager/task-service/internal/handler"
    "github.com/irzan-aldi/task-manager/task-service/internal/repository"
    "github.com/irzan-aldi/task-manager/task-service/internal/service"
)

func main() {
    cfg := config.Load()
    db := config.ConnectDB(cfg)

    // Connect ke User Service via gRPC
    userClient, err := client.NewUserClient(cfg.UserServiceGRPC) // e.g., "user-service:50051"
    if err != nil {
        log.Fatalf("Failed to connect to User Service: %v", err)
    }
    defer userClient.Close()

    taskRepo := repository.NewTaskRepository(db)
    taskService := service.NewTaskService(taskRepo, userClient)
    taskHandler := handler.NewTaskHandler(taskService)

    r := gin.Default()
    // Setup routes...
    r.POST("/tasks", taskHandler.Create)
    r.GET("/tasks/:id", taskHandler.GetByID)

    port := cfg.Port
    if port == "" {
        port = "3004"
    }

    log.Printf("Task Service started on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
        os.Exit(1)
    }
}
```

### gRPC vs REST — Performance

```
Benchmark: Get User by ID (1000 requests)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

REST (HTTP/1.1 + JSON):
  - Avg latency: 2.3ms
  - Throughput: 4,200 req/sec
  - Payload size: 85 bytes

gRPC (HTTP/2 + Protobuf):
  - Avg latency: 0.4ms
  - Throughput: 18,500 req/sec
  - Payload size: 28 bytes

gRPC itu ~5x lebih cepat dan ~3x lebih kecil payload-nya.
Untuk internal service-to-service yang high frequency, ini significant.
```

---

## 18.7 API Gateway

### Kenapa Butuh API Gateway?

Tanpa API Gateway, client harus tau **alamat setiap service**:

```
Tanpa API Gateway:
━━━━━━━━━━━━━━━━━
Client harus tau:
  - Auth:  http://auth-service:3001/auth/login
  - User:  http://user-service:3002/users/profile
  - Board: http://board-service:3003/boards
  - Task:  http://task-service:3004/tasks

Problem:
  - Client harus manage banyak URLs
  - CORS di setiap service
  - Auth di setiap service
  - Rate limiting di setiap service
  - Kalau service pindah port/host → client harus update


Dengan API Gateway:
━━━━━━━━━━━━━━━━━━
Client cuma tau 1 URL: http://api.taskmanager.com

API Gateway yang route:
  /api/v1/auth/*   → auth-service:3001
  /api/v1/users/*  → user-service:3002
  /api/v1/boards/* → board-service:3003
  /api/v1/tasks/*  → task-service:3004
```

### API Gateway Functions

```
┌─────────────────────────────────────────────────┐
│                 API GATEWAY                      │
│                                                   │
│  1. ROUTING          → Route request ke service   │
│  2. AUTHENTICATION   → Validate JWT sekali        │
│  3. RATE LIMITING    → Protect dari abuse          │
│  4. LOAD BALANCING   → Distribute ke instances     │
│  5. CIRCUIT BREAKING → Protect dari cascade fail   │
│  6. LOGGING          → Central request logging     │
│  7. CORS             → Handle CORS sekali          │
│  8. TRANSFORMATION   → Request/response transform  │
│  9. CACHING          → Cache response yang sering  │
│ 10. SSL TERMINATION  → HTTPS → HTTP internal       │
└─────────────────────────────────────────────────┘
```

### API Gateway Options

| Option | Type | Pros | Cons |
|--------|------|------|------|
| **Kong** | Open-source | Plugin-rich, battle-tested, Lua-based | Complex setup, resource heavy |
| **Traefik** | Open-source | K8s native, auto-discovery, Go-based | Less plugins than Kong |
| **AWS API Gateway** | Managed | No ops, auto-scale, pay-per-use | Vendor lock-in, latency |
| **Nginx** | Open-source | Super fast, reliable | Config-based, less dynamic |
| **Custom (Go)** | Custom | Full control, lightweight | Build & maintain yourself |
| **Envoy** | Open-source | L7 proxy, service mesh ready | Complex config (YAML heavy) |

### Simple API Gateway di Go

Untuk belajar, kita bikin API Gateway sederhana pakai Go:

```go
// api-gateway/cmd/main.go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/api-gateway/internal/gateway"
    "github.com/irzan-aldi/task-manager/api-gateway/internal/middleware"
)

func main() {
    r := gin.Default()

    // CORS
    r.Use(middleware.CORS())

    // Rate limiting (global)
    r.Use(middleware.RateLimiter(100, 60)) // 100 req/min

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok", "service": "api-gateway"})
    })

    // Setup reverse proxy routes
    gw := gateway.New()

    api := r.Group("/api/v1")
    {
        // Public routes (no auth)
        api.Any("/auth/*path", gw.ProxyTo("auth-service:3001"))

        // Protected routes (with auth)
        protected := api.Group("")
        protected.Use(middleware.AuthMiddleware(gw.AuthClient))
        {
            protected.Any("/users/*path", gw.ProxyTo("user-service:3002"))
            protected.Any("/boards/*path", gw.ProxyTo("board-service:3003"))
            protected.Any("/tasks/*path", gw.ProxyTo("task-service:3004"))
            protected.Any("/notifications/*path", gw.ProxyTo("notification-service:3005"))
        }
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("API Gateway starting on port %s", port)
    r.Run(":" + port)
}
```

```go
// api-gateway/internal/gateway/proxy.go
package gateway

import (
    "fmt"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/api-gateway/internal/client"
)

type Gateway struct {
    AuthClient *client.AuthClient
    proxies    map[string]*httputil.ReverseProxy
}

func New() *Gateway {
    authClient, _ := client.NewAuthClient("auth-service:50051")
    return &Gateway{
        AuthClient: authClient,
        proxies:    make(map[string]*httputil.ReverseProxy),
    }
}

// ProxyTo returns a Gin handler that proxies requests to the target service
func (gw *Gateway) ProxyTo(target string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get or create reverse proxy
        proxy, ok := gw.proxies[target]
        if !ok {
            targetURL, err := url.Parse(fmt.Sprintf("http://%s", target))
            if err != nil {
                c.JSON(http.StatusBadGateway, gin.H{"error": "invalid target service"})
                return
            }

            proxy = httputil.NewSingleHostReverseProxy(targetURL)

            // Custom error handler
            proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
                w.WriteHeader(http.StatusBadGateway)
                w.Write([]byte(fmt.Sprintf(`{"error":"service unavailable: %s"}`, target)))
            }

            gw.proxies[target] = proxy
        }

        // Rewrite path — remove /api/v1/users prefix, keep the rest
        // /api/v1/users/profile → /users/profile
        originalPath := c.Request.URL.Path
        parts := strings.SplitN(originalPath, "/", 5) // ["", "api", "v1", "users", "profile"]
        if len(parts) >= 5 {
            c.Request.URL.Path = "/" + strings.Join(parts[3:], "/")
        } else if len(parts) == 4 {
            c.Request.URL.Path = "/" + parts[3]
        }

        // Forward user info from JWT (set by auth middleware)
        if userID, exists := c.Get("userID"); exists {
            c.Request.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
        }
        if userEmail, exists := c.Get("userEmail"); exists {
            c.Request.Header.Set("X-User-Email", fmt.Sprintf("%v", userEmail))
        }
        if userRole, exists := c.Get("userRole"); exists {
            c.Request.Header.Set("X-User-Role", fmt.Sprintf("%v", userRole))
        }

        // Forward request ID for tracing
        c.Request.Header.Set("X-Request-ID", c.GetString("requestID"))

        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

```go
// api-gateway/internal/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/api-gateway/internal/client"
)

func AuthMiddleware(authClient *client.AuthClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
            return
        }

        token := parts[1]

        // Validate token via Auth Service (gRPC)
        resp, err := authClient.ValidateToken(c.Request.Context(), token)
        if err != nil || !resp.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

        // Set user info in context
        c.Set("userID", resp.UserId)
        c.Set("userEmail", resp.Email)
        c.Set("userRole", resp.Role)

        c.Next()
    }
}
```

```go
// api-gateway/internal/middleware/ratelimit.go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
)

type visitor struct {
    count    int
    lastSeen time.Time
}

var (
    visitors = make(map[string]*visitor)
    mu       sync.Mutex
)

// RateLimiter implements a simple in-memory rate limiter
func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
    // Cleanup goroutine
    go func() {
        for {
            time.Sleep(window)
            mu.Lock()
            for ip, v := range visitors {
                if time.Since(v.lastSeen) > window {
                    delete(visitors, ip)
                }
            }
            mu.Unlock()
        }
    }()

    return func(c *gin.Context) {
        ip := c.ClientIP()

        mu.Lock()
        v, exists := visitors[ip]
        if !exists {
            visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
            mu.Unlock()
            c.Next()
            return
        }

        if time.Since(v.lastSeen) > window {
            v.count = 1
            v.lastSeen = time.Now()
            mu.Unlock()
            c.Next()
            return
        }

        v.count++
        v.lastSeen = time.Now()

        if v.count > maxRequests {
            mu.Unlock()
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
            })
            return
        }

        mu.Unlock()
        c.Next()
    }
}
```

---

## 18.8 Service Discovery

### Problem: Gimana Service A Nemuin Service B?

Di monolith, semua module ada di 1 process. Function call langsung. Di microservices, service A harus tau **alamat** service B.

```
Problem:
  Task Service mau call User Service.
  User Service ada di mana?
  - IP 10.0.0.5? Kalau server ganti?
  - Port 3002? Kalau port berubah?
  - Instance mana? Kalau ada 3 instances?

  Hardcode IP = fragile. Kalau service pindah, semua client harus update.
```

### Solution Options

#### 1. DNS-Based (Kubernetes Default)

Di Kubernetes, setiap **Service** otomatis dapat DNS name. Ini yang paling umum dan paling gampang.

```yaml
# k8s/user-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service          # ← DNS name: user-service.default.svc.cluster.local
  namespace: default
spec:
  selector:
    app: user-service
  ports:
    - name: http
      port: 3002
      targetPort: 3002
    - name: grpc
      port: 50051
      targetPort: 50051
```

```go
// Di Task Service, cukup pakai DNS name
userClient, err := client.NewUserClient("user-service:50051")
// Kubernetes DNS resolve "user-service" ke IP yang benar
// Kalau ada 3 instances, K8s load balance otomatis
```

```
Kubernetes DNS format:
━━━━━━━━━━━━━━━━━━━━

<service-name>                              → Same namespace
<service-name>.<namespace>                  → Cross namespace
<service-name>.<namespace>.svc.cluster.local → Full qualified

Contoh:
  user-service                              → http://user-service:3002
  user-service.production                   → http://user-service.production:3002
  user-service.production.svc.cluster.local → Full path
```

#### 2. Consul (HashiCorp)

Service registry yang populer. Service register dirinya ke Consul, dan service lain query Consul untuk nemuin alamat.

```
┌──────────┐  register   ┌──────────┐
│  User    │────────────→│  Consul  │
│  Service │             │  Server  │
└──────────┘             └────┬─────┘
                              │
┌──────────┐  discover   ┌────▼─────┐
│  Task    │←────────────│  Consul  │
│  Service │             │  Server  │
└──────────┘             └──────────┘
```

```go
// Register service ke Consul
import consul "github.com/hashicorp/consul/api"

func registerToConsul(serviceName, serviceID, address string, port int) error {
    config := consul.DefaultConfig()
    config.Address = "consul:8500"

    client, err := consul.NewClient(config)
    if err != nil {
        return err
    }

    registration := &consul.AgentServiceRegistration{
        ID:      serviceID,
        Name:    serviceName,
        Port:    port,
        Address: address,
        Check: &consul.AgentServiceCheck{
            HTTP:     fmt.Sprintf("http://%s:%d/health", address, port),
            Interval: "10s",
            Timeout:  "5s",
        },
    }

    return client.Agent().ServiceRegister(registration)
}

// Discover service dari Consul
func discoverService(serviceName string) (string, error) {
    config := consul.DefaultConfig()
    config.Address = "consul:8500"

    client, err := consul.NewClient(config)
    if err != nil {
        return "", err
    }

    services, _, err := client.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return "", err
    }

    if len(services) == 0 {
        return "", fmt.Errorf("no healthy instances of %s", serviceName)
    }

    // Simple round-robin (production pakai proper load balancing)
    service := services[rand.Intn(len(services))]
    return fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port), nil
}
```

#### 3. etcd

Key-value store yang dipakai Kubernetes sendiri. Bisa juga dipakai sebagai service registry.

### Rekomendasi

```
Pakai Kubernetes? → DNS-based (gratis, built-in, zero config)
Ga pakai K8s?    → Consul (feature-rich, health checking)
Already have etcd? → etcd (kalau ga mau tambah dependency)

Untuk project kita: Kubernetes DNS. Simple dan cukup.
```

---

## 18.9 Event-Driven Architecture

### Events vs Commands

```
Command (imperative):
  "Kirim email ke user 123"
  → Kamu suruh service tertentu melakukan sesuatu
  → Coupling: sender HARUS tau siapa yang handle

Event (declarative):
  "Task 456 telah dibuat"
  → Kamu cuma announce apa yang terjadi
  → Decoupling: sender GA PERLU tau siapa yang listen
  → Bisa ada 0, 1, atau 10 listener
```

```
Command pattern:
  Task Service ──"Send email to user 123"──→ Notification Service
  Task Service HARUS tau Notification Service ada.
  Kalau nanti mau tambah Analytics Service? Task Service harus diubah.

Event pattern:
  Task Service ──"task.created {taskID: 456}"──→ Message Broker
                                                   │
                                                   ├──→ Notification Service (kirim email)
                                                   ├──→ Analytics Service (record metric)
                                                   └──→ Search Service (index task)
  Task Service GA PERLU tau siapa yang listen.
  Mau tambah listener baru? Cukup subscribe. Task Service ga diubah.
```

### Event Sourcing (Konsep)

Alih-alih simpan **current state**, kamu simpan **semua events yang terjadi**. Current state = replay semua events.

```
Traditional (state-based):
  tasks table:
  | id | title      | status | assignee_id |
  | 1  | Fix bug    | done   | 5           |

  Kamu tau status sekarang, tapi ga tau HISTORY.

Event Sourcing:
  events table:
  | id | event_type      | data                                | timestamp       |
  | 1  | task.created    | {title: "Fix bug", status: "todo"}  | 2026-05-01 10:00|
  | 2  | task.assigned   | {assignee_id: 5}                    | 2026-05-01 10:05|
  | 3  | task.updated    | {status: "in_progress"}             | 2026-05-01 14:00|
  | 4  | task.completed  | {status: "done"}                    | 2026-05-01 16:00|

  Replay events → current state = {title: "Fix bug", status: "done", assignee: 5}
  Bonus: kamu punya FULL HISTORY. Bisa audit, debug, undo.
```

> **Kapan pakai event sourcing?** Kalau kamu butuh **audit trail** lengkap (finance, compliance), atau butuh **temporal query** ("tampilkan state board ini 1 minggu lalu"). Untuk kebanyakan CRUD app, ga perlu. Overkill.

### CQRS (Command Query Responsibility Segregation)

Pisahkan **write model** dan **read model**.

```
Traditional:
  API → Same Service → Same Database (read & write)

CQRS:
  Write: API → Command Handler → Write DB (normalized, ACID)
  Read:  API → Query Handler → Read DB (denormalized, fast)

Contoh di Task Manager:
  Write: Create task → PostgreSQL (normalized, foreign keys)
  Read:  List tasks with assignee names → Redis/Elasticsearch (denormalized, pre-joined)

  Event bridge: task.created → update read model dengan joined data
```

> **Kapan pakai CQRS?** Kalau read dan write pattern **sangat berbeda**. Misalnya write jarang tapi read sangat sering dan butuh complex query. Untuk kebanyakan app, overkill. Monolith CRUD cukup.

### Practical: Event-Driven dengan NATS

NATS itu lightweight messaging system. Perfect buat cloud-native microservices. Lebih simple dari Kafka, lebih modern dari RabbitMQ.

```go
// shared/events/events.go
package events

// Event types
const (
    TaskCreated   = "task.created"
    TaskAssigned  = "task.assigned"
    TaskCompleted = "task.completed"
    UserRegistered = "user.registered"
    BoardDeleted  = "board.deleted"
)

// Event payloads
type TaskCreatedEvent struct {
    TaskID      uint   `json:"task_id"`
    Title       string `json:"title"`
    BoardID     uint   `json:"board_id"`
    CreatorID   uint   `json:"creator_id"`
    AssigneeID  *uint  `json:"assignee_id,omitempty"`
}

type TaskAssignedEvent struct {
    TaskID     uint   `json:"task_id"`
    Title      string `json:"title"`
    AssigneeID uint   `json:"assignee_id"`
    AssignerID uint   `json:"assigner_id"`
}

type TaskCompletedEvent struct {
    TaskID      uint   `json:"task_id"`
    Title       string `json:"title"`
    CompletedBy uint   `json:"completed_by"`
    BoardID     uint   `json:"board_id"`
}

type UserRegisteredEvent struct {
    UserID uint   `json:"user_id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
}
```

```go
// shared/messaging/nats.go
package messaging

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
)

type NATSClient struct {
    conn *nats.Conn
}

func NewNATSClient(url string) (*NATSClient, error) {
    // Connect dengan retry
    var conn *nats.Conn
    var err error

    for i := 0; i < 5; i++ {
        conn, err = nats.Connect(url,
            nats.RetryOnFailedConnect(true),
            nats.MaxReconnects(10),
            nats.ReconnectWait(2*time.Second),
            nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
                log.Printf("NATS disconnected: %v", err)
            }),
            nats.ReconnectHandler(func(nc *nats.Conn) {
                log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
            }),
        )
        if err == nil {
            break
        }
        log.Printf("Failed to connect to NATS (attempt %d/5): %v", i+1, err)
        time.Sleep(2 * time.Second)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS after 5 attempts: %w", err)
    }

    log.Printf("Connected to NATS: %s", conn.ConnectedUrl())
    return &NATSClient{conn: conn}, nil
}

// Publish event ke NATS
func (c *NATSClient) Publish(subject string, data interface{}) error {
    payload, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    if err := c.conn.Publish(subject, payload); err != nil {
        return fmt.Errorf("failed to publish to %s: %w", subject, err)
    }

    log.Printf("Published event: %s", subject)
    return nil
}

// Subscribe ke event dari NATS
func (c *NATSClient) Subscribe(subject string, handler func(data []byte)) error {
    _, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
        log.Printf("Received event: %s", subject)
        handler(msg.Data)
    })
    if err != nil {
        return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
    }

    log.Printf("Subscribed to: %s", subject)
    return nil
}

// QueueSubscribe — multiple instances share the load (consumer group)
func (c *NATSClient) QueueSubscribe(subject, queue string, handler func(data []byte)) error {
    _, err := c.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
        log.Printf("Received event: %s (queue: %s)", subject, queue)
        handler(msg.Data)
    })
    if err != nil {
        return fmt.Errorf("failed to queue subscribe to %s: %w", subject, err)
    }

    log.Printf("Queue subscribed to: %s (queue: %s)", subject, queue)
    return nil
}

func (c *NATSClient) Close() {
    c.conn.Close()
}
```

```go
// task-service/internal/service/task.go
// Publishing event saat task dibuat
func (s *TaskService) Create(ctx context.Context, req CreateTaskRequest) (*Task, error) {
    task, err := s.repo.Create(&Task{
        Title:      req.Title,
        BoardID:    req.BoardID,
        AssigneeID: req.AssigneeID,
        Status:     "todo",
    })
    if err != nil {
        return nil, err
    }

    // Publish event — async, fire and forget
    go func() {
        event := events.TaskCreatedEvent{
            TaskID:     task.ID,
            Title:      task.Title,
            BoardID:    task.BoardID,
            CreatorID:  req.CreatorID,
            AssigneeID: task.AssigneeID,
        }
        if err := s.nats.Publish(events.TaskCreated, event); err != nil {
            log.Printf("Failed to publish task.created event: %v", err)
            // Event gagal publish bukan alasan untuk fail task creation
        }
    }()

    return task, nil
}
```

```go
// notification-service/cmd/main.go
// Consuming events
func main() {
    // Connect NATS
    natsClient, err := messaging.NewNATSClient("nats://nats:4222")
    if err != nil {
        log.Fatalf("Failed to connect NATS: %v", err)
    }
    defer natsClient.Close()

    notifService := service.NewNotificationService(db, emailClient)

    // Subscribe ke events
    // QueueSubscribe = kalau ada 3 instances notification service,
    // cuma 1 yang process setiap event (load balanced)
    natsClient.QueueSubscribe(events.TaskAssigned, "notification-workers", func(data []byte) {
        var event events.TaskAssignedEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal task.assigned event: %v", err)
            return
        }

        // Send notification
        if err := notifService.SendTaskAssigned(event); err != nil {
            log.Printf("Failed to send task assigned notification: %v", err)
        }
    })

    natsClient.QueueSubscribe(events.TaskCompleted, "notification-workers", func(data []byte) {
        var event events.TaskCompletedEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal task.completed event: %v", err)
            return
        }

        if err := notifService.SendTaskCompleted(event); err != nil {
            log.Printf("Failed to send task completed notification: %v", err)
        }
    })

    natsClient.QueueSubscribe(events.UserRegistered, "notification-workers", func(data []byte) {
        var event events.UserRegisteredEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal user.registered event: %v", err)
            return
        }

        if err := notifService.SendWelcomeEmail(event); err != nil {
            log.Printf("Failed to send welcome email: %v", err)
        }
    })

    log.Println("Notification Service started — listening for events...")

    // Block main goroutine
    select {}
}
```

---

## 18.10 Saga Pattern — Distributed Transactions

### Problem: Distributed Transactions

Di monolith, kamu bisa pakai **database transaction** untuk ensure consistency:

```go
// Monolith — simple transaction
func (s *BoardService) CreateBoardWithDefaults(ctx context.Context, req CreateBoardRequest) error {
    tx := s.db.Begin()

    // Step 1: Create board
    board := &Board{Title: req.Title, OwnerID: req.OwnerID}
    if err := tx.Create(board).Error; err != nil {
        tx.Rollback()
        return err
    }

    // Step 2: Create default tasks
    defaultTasks := []Task{
        {Title: "Welcome task", BoardID: board.ID, Status: "todo"},
        {Title: "Read documentation", BoardID: board.ID, Status: "todo"},
    }
    for _, task := range defaultTasks {
        if err := tx.Create(&task).Error; err != nil {
            tx.Rollback() // ← Rollback SEMUA kalau ada yang gagal
            return err
        }
    }

    // Step 3: Send notification
    // (di dalam transaction, guaranteed consistent)

    tx.Commit() // ← Commit SEMUA kalau semua sukses
    return nil
}
```

Di microservices, Board Service dan Task Service punya **database terpisah**. Kamu **ga bisa** pakai 1 database transaction yang span 2 service.

```
Problem:
  1. Board Service creates board (board_db) ← Success
  2. Task Service creates default tasks (task_db) ← FAIL!
  3. Notification Service sends email ← Never reached

  Board udah ada, tapi default tasks ga ada. INCONSISTENT!

  Ga bisa rollback board_db dari Task Service. Beda database.
```

### Solution: Saga Pattern

Saga itu **sequence of local transactions** dimana setiap step punya **compensating transaction** (undo action).

```
Saga Flow:
━━━━━━━━━━

Happy path (semua sukses):
  Step 1: Create Board → ✅
  Step 2: Create Default Tasks → ✅
  Step 3: Send Notification → ✅
  Done!

Failure path (step 2 gagal):
  Step 1: Create Board → ✅
  Step 2: Create Default Tasks → ❌ FAIL
  Compensate Step 1: DELETE Board → ✅ (undo step 1)
  Done! (consistent — seolah-olah nothing happened)

Failure path (step 3 gagal):
  Step 1: Create Board → ✅
  Step 2: Create Default Tasks → ✅
  Step 3: Send Notification → ❌ FAIL
  Compensate Step 2: DELETE Default Tasks → ✅ (undo step 2)
  Compensate Step 1: DELETE Board → ✅ (undo step 1)
  Done!
```

### Choreography vs Orchestration

Ada 2 cara implement Saga:

#### 1. Choreography (Event-based)

Setiap service **react** ke event dari service sebelumnya. Ga ada coordinator.

```
Board Service ──board.created──→ Message Broker
                                      │
                                      ▼
Task Service listen "board.created" → Create default tasks
Task Service ──tasks.created──→ Message Broker
                                      │
                                      ▼
Notification Service listen "tasks.created" → Send email

Kalau Task Service gagal:
Task Service ──tasks.creation.failed──→ Message Broker
                                            │
                                            ▼
Board Service listen "tasks.creation.failed" → Delete board (compensate)
```

**Pros:** Decoupled, simple for 2-3 steps.
**Cons:** Hard to track overall flow, debugging nightmare kalau banyak steps, no central view.

#### 2. Orchestration (Coordinator-based)

Ada satu **Saga Orchestrator** yang coordinate semua steps.

```
┌───────────────────────────────────────────────────┐
│               SAGA ORCHESTRATOR                    │
│                                                    │
│  Step 1: Call Board Service → Create Board         │
│          ↳ Compensate: Delete Board                │
│                                                    │
│  Step 2: Call Task Service → Create Default Tasks  │
│          ↳ Compensate: Delete Tasks                │
│                                                    │
│  Step 3: Call Notif Service → Send Email           │
│          ↳ Compensate: (no-op, email ga bisa undo) │
│                                                    │
│  If any step fails → run compensations in reverse  │
└───────────────────────────────────────────────────┘
```

**Pros:** Easy to understand, easy to track, centralized logic.
**Cons:** Orchestrator bisa jadi single point of failure, more coupling.

### Implementation: Saga Orchestrator di Go

```go
// saga/saga.go
package saga

import (
    "context"
    "fmt"
    "log"
)

// Step represents one step in a saga
type Step struct {
    Name       string
    Action     func(ctx context.Context) error
    Compensate func(ctx context.Context) error
}

// Saga orchestrates a sequence of steps with compensation
type Saga struct {
    name  string
    steps []Step
}

func New(name string) *Saga {
    return &Saga{name: name}
}

func (s *Saga) AddStep(step Step) *Saga {
    s.steps = append(s.steps, step)
    return s
}

// Execute runs all steps. If any fails, compensate completed steps in reverse.
func (s *Saga) Execute(ctx context.Context) error {
    log.Printf("Saga [%s] started", s.name)

    completedSteps := make([]Step, 0)

    for _, step := range s.steps {
        log.Printf("Saga [%s] executing step: %s", s.name, step.Name)

        if err := step.Action(ctx); err != nil {
            log.Printf("Saga [%s] step '%s' failed: %v", s.name, step.Name, err)

            // Compensate completed steps in reverse order
            if compErr := s.compensate(ctx, completedSteps); compErr != nil {
                log.Printf("Saga [%s] compensation failed: %v", s.name, compErr)
                return fmt.Errorf("saga failed at step '%s': %w (compensation also failed: %v)",
                    step.Name, err, compErr)
            }

            return fmt.Errorf("saga failed at step '%s' (compensated): %w", step.Name, err)
        }

        completedSteps = append(completedSteps, step)
        log.Printf("Saga [%s] step '%s' completed", s.name, step.Name)
    }

    log.Printf("Saga [%s] completed successfully", s.name)
    return nil
}

func (s *Saga) compensate(ctx context.Context, completedSteps []Step) error {
    log.Printf("Saga [%s] compensating %d steps...", s.name, len(completedSteps))

    var errors []error

    // Compensate in reverse order
    for i := len(completedSteps) - 1; i >= 0; i-- {
        step := completedSteps[i]
        if step.Compensate != nil {
            log.Printf("Saga [%s] compensating step: %s", s.name, step.Name)
            if err := step.Compensate(ctx); err != nil {
                log.Printf("Saga [%s] compensation for '%s' failed: %v", s.name, step.Name, err)
                errors = append(errors, err)
            }
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("%d compensation(s) failed", len(errors))
    }

    return nil
}
```

```go
// Usage: Create Board with defaults using Saga
func (h *BoardHandler) CreateBoardWithDefaults(c *gin.Context) {
    var req CreateBoardWithDefaultsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    ctx := c.Request.Context()
    var boardID uint
    var taskIDs []uint

    // Build saga
    createBoardSaga := saga.New("create-board-with-defaults")

    // Step 1: Create Board
    createBoardSaga.AddStep(saga.Step{
        Name: "create-board",
        Action: func(ctx context.Context) error {
            board, err := h.boardClient.CreateBoard(ctx, req.Title, req.OwnerID)
            if err != nil {
                return err
            }
            boardID = board.ID
            return nil
        },
        Compensate: func(ctx context.Context) error {
            return h.boardClient.DeleteBoard(ctx, boardID)
        },
    })

    // Step 2: Create Default Tasks
    createBoardSaga.AddStep(saga.Step{
        Name: "create-default-tasks",
        Action: func(ctx context.Context) error {
            defaults := []string{"Welcome!", "Read documentation", "Create your first task"}
            for _, title := range defaults {
                task, err := h.taskClient.CreateTask(ctx, boardID, title)
                if err != nil {
                    return err
                }
                taskIDs = append(taskIDs, task.ID)
            }
            return nil
        },
        Compensate: func(ctx context.Context) error {
            for _, id := range taskIDs {
                _ = h.taskClient.DeleteTask(ctx, id) // Best effort
            }
            return nil
        },
    })

    // Step 3: Send Notification (no compensation — email can't be unsent)
    createBoardSaga.AddStep(saga.Step{
        Name: "send-notification",
        Action: func(ctx context.Context) error {
            return h.nats.Publish("board.created", map[string]interface{}{
                "board_id": boardID,
                "owner_id": req.OwnerID,
                "title":    req.Title,
            })
        },
        Compensate: nil, // Can't unsend notification — acceptable
    })

    // Execute saga
    if err := createBoardSaga.Execute(ctx); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, gin.H{
        "message":  "board created with defaults",
        "board_id": boardID,
    })
}
```

### Kapan Pakai Saga?

```
PAKAI Saga kalau:
  - Multi-service transaction (create di service A, update di service B)
  - Business process yang span multiple services
  - Kamu butuh consistency guarantee

JANGAN PAKAI Saga kalau:
  - Single service operation (pakai DB transaction biasa)
  - Fire-and-forget (notification yang gagal ga masalah)
  - Semua data di 1 database (monolith / shared DB)
```

---

## 18.11 Circuit Breaker

### Problem: Cascade Failure

Bayangin Task Service call User Service, User Service call Auth Service. Kalau Auth Service down:

```
Tanpa Circuit Breaker:
━━━━━━━━━━━━━━━━━━━━━

Request → Task Service → User Service → Auth Service (DOWN!)
                                         ↑ Timeout 30s...
                         ← Waiting...    │
           ← Waiting...                  │
User nunggu 30 detik, terus error.

Request 2 → Task Service → User Service → Auth Service (STILL DOWN!)
                                           ↑ Timeout 30s lagi...

100 requests → 100 threads stuck → Thread pool exhausted
→ Task Service JUGA DOWN sekarang
→ CASCADE FAILURE!

Auth down → User down → Task down → EVERYTHING DOWN
```

### Solution: Circuit Breaker

Konsepnya mirip **circuit breaker** (MCB) di rumah kamu. Kalau ada korsleting (service down), MCB trip — putus aliran listrik (request). Setelah beberapa waktu, coba nyalain lagi.

```
Circuit Breaker States:
━━━━━━━━━━━━━━━━━━━━━

┌──────────┐    failure threshold    ┌──────────┐
│  CLOSED  │ ──────reached──────→    │   OPEN   │
│ (normal) │                         │  (block) │
│          │                         │          │
│ Allow all│                         │ Block all│
│ requests │    success              │ requests │
│          │ ←──────────────────     │          │
└──────────┘                    │    └─────┬────┘
      ↑                        │          │
      │                        │    timeout expired
      │                        │          │
      │              ┌─────────┴────┐     │
      └──success─────│  HALF-OPEN   │←────┘
                     │  (testing)   │
                     │              │
                     │ Allow 1 req  │
                     │ to test      │
                     └──────────────┘

CLOSED:    Semua request diterusin. Normal operation.
           Kalau gagal > threshold → switch ke OPEN.

OPEN:      Semua request langsung di-reject. Ga call service.
           Return error/fallback immediately.
           Setelah timeout → switch ke HALF-OPEN.

HALF-OPEN: Coba 1 request. Kalau sukses → CLOSED (recovered!).
           Kalau gagal → OPEN lagi.
```

### Implementation dengan sony/gobreaker

```go
// internal/circuitbreaker/breaker.go
package circuitbreaker

import (
    "fmt"
    "time"

    "github.com/sony/gobreaker/v2"
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string) *gobreaker.CircuitBreaker[[]byte] {
    settings := gobreaker.Settings{
        Name:        name,
        MaxRequests: 3,                // Max requests in half-open state
        Interval:    30 * time.Second, // Reset failure count after this interval (in closed state)
        Timeout:     10 * time.Second, // Time to wait before switching from open to half-open

        // ReadyToTrip decides when to switch from Closed to Open
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // Open circuit if failure ratio > 50% AND at least 5 requests
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 5 && failureRatio >= 0.5
        },

        // OnStateChange logs state changes
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            fmt.Printf("Circuit Breaker [%s]: %s → %s\n", name, from, to)
        },
    }

    return gobreaker.NewCircuitBreaker[[]byte](settings)
}
```

```go
// internal/client/user_client_with_cb.go
package client

import (
    "context"
    "fmt"
    "io"
    "net/http"

    "github.com/sony/gobreaker/v2"
    "github.com/irzan-aldi/task-manager/internal/circuitbreaker"
)

type UserClientWithCB struct {
    baseURL string
    cb      *gobreaker.CircuitBreaker[[]byte]
    client  *http.Client
}

func NewUserClientWithCB(baseURL string) *UserClientWithCB {
    return &UserClientWithCB{
        baseURL: baseURL,
        cb:      circuitbreaker.NewCircuitBreaker("user-service"),
        client:  &http.Client{Timeout: 5 * time.Second},
    }
}

func (c *UserClientWithCB) GetUser(ctx context.Context, userID uint) ([]byte, error) {
    // Wrap HTTP call dengan circuit breaker
    body, err := c.cb.Execute(func() ([]byte, error) {
        url := fmt.Sprintf("%s/users/%d", c.baseURL, userID)

        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return nil, err
        }

        resp, err := c.client.Do(req)
        if err != nil {
            return nil, err // Network error → counted as failure
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 500 {
            return nil, fmt.Errorf("user service returned %d", resp.StatusCode)
            // 5xx → counted as failure (server error)
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, err
        }

        return body, nil
        // Success! → failure count reset
    })

    if err != nil {
        // Circuit breaker open? Return fallback
        if err == gobreaker.ErrOpenState {
            return nil, fmt.Errorf("user service circuit breaker is OPEN — service likely down")
        }
        return nil, err
    }

    return body, nil
}
```

### Custom Circuit Breaker (DIY)

Kalau ga mau pakai library, circuit breaker itu cukup simple buat di-implement sendiri:

```go
// internal/circuitbreaker/simple.go
package circuitbreaker

import (
    "errors"
    "sync"
    "time"
)

type State int

const (
    StateClosed   State = iota // Normal — allow requests
    StateOpen                  // Tripped — block requests
    StateHalfOpen              // Testing — allow 1 request
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type SimpleBreaker struct {
    mu            sync.Mutex
    state         State
    failureCount  int
    successCount  int
    maxFailures   int
    timeout       time.Duration
    lastFailure   time.Time
}

func NewSimpleBreaker(maxFailures int, timeout time.Duration) *SimpleBreaker {
    return &SimpleBreaker{
        state:       StateClosed,
        maxFailures: maxFailures,
        timeout:     timeout,
    }
}

func (b *SimpleBreaker) Execute(fn func() error) error {
    b.mu.Lock()

    switch b.state {
    case StateOpen:
        // Check if timeout has passed
        if time.Since(b.lastFailure) > b.timeout {
            b.state = StateHalfOpen
            b.mu.Unlock()
            return b.tryHalfOpen(fn)
        }
        b.mu.Unlock()
        return ErrCircuitOpen

    case StateHalfOpen:
        b.mu.Unlock()
        return b.tryHalfOpen(fn)

    default: // Closed
        b.mu.Unlock()
        return b.tryClosed(fn)
    }
}

func (b *SimpleBreaker) tryClosed(fn func() error) error {
    err := fn()

    b.mu.Lock()
    defer b.mu.Unlock()

    if err != nil {
        b.failureCount++
        b.lastFailure = time.Now()
        if b.failureCount >= b.maxFailures {
            b.state = StateOpen
        }
        return err
    }

    b.failureCount = 0 // Reset on success
    return nil
}

func (b *SimpleBreaker) tryHalfOpen(fn func() error) error {
    err := fn()

    b.mu.Lock()
    defer b.mu.Unlock()

    if err != nil {
        b.state = StateOpen // Back to open
        b.lastFailure = time.Now()
        return err
    }

    // Success! Close the circuit
    b.state = StateClosed
    b.failureCount = 0
    return nil
}

func (b *SimpleBreaker) State() State {
    b.mu.Lock()
    defer b.mu.Unlock()
    return b.state
}
```

---

## 18.12 Retry & Timeout

### Kenapa Retry Penting?

Network call bisa gagal karena banyak alasan — temporary network glitch, service lagi restart, DNS hiccup. Banyak dari failures ini **transient** (sementara). Retry bisa solve.

### Naive Retry = Berbahaya

```
❌ Naive retry (JANGAN LAKUIN INI):
for i := 0; i < 10; i++ {
    resp, err := http.Get("http://user-service:3002/users/123")
    if err == nil {
        return resp
    }
    // Immediately retry — NO WAIT
}

Problem:
  Service lagi overloaded → kamu kirim 10 request tambahan → service MAKIN overloaded
  1000 clients retry 10x = 10,000 requests → DDoS your own service
  Ini namanya "retry storm" atau "thundering herd"
```

### Exponential Backoff with Jitter

**Exponential backoff** = tiap retry, nunggu lebih lama (1s, 2s, 4s, 8s, ...).
**Jitter** = tambah random delay supaya ga semua client retry di waktu yang sama.

```go
// internal/retry/retry.go
package retry

import (
    "context"
    "fmt"
    "log"
    "math"
    "math/rand"
    "time"
)

type Config struct {
    MaxRetries  int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func DefaultConfig() Config {
    return Config{
        MaxRetries: 3,
        BaseDelay:  500 * time.Millisecond,
        MaxDelay:   10 * time.Second,
        Multiplier: 2.0,
    }
}

// Do executes fn with exponential backoff retry
func Do(ctx context.Context, cfg Config, fn func() error) error {
    var lastErr error

    for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
        // Execute function
        if err := fn(); err == nil {
            return nil // Success!
        } else {
            lastErr = err
            log.Printf("Attempt %d/%d failed: %v", attempt+1, cfg.MaxRetries+1, err)
        }

        // Don't wait after last attempt
        if attempt == cfg.MaxRetries {
            break
        }

        // Calculate delay with exponential backoff + jitter
        delay := float64(cfg.BaseDelay) * math.Pow(cfg.Multiplier, float64(attempt))
        if delay > float64(cfg.MaxDelay) {
            delay = float64(cfg.MaxDelay)
        }

        // Add jitter (random ±25%)
        jitter := delay * 0.25 * (rand.Float64()*2 - 1) // -25% to +25%
        actualDelay := time.Duration(delay + jitter)

        log.Printf("Retrying in %v...", actualDelay)

        // Wait with context cancellation support
        select {
        case <-ctx.Done():
            return fmt.Errorf("retry cancelled: %w", ctx.Err())
        case <-time.After(actualDelay):
            // Continue to next attempt
        }
    }

    return fmt.Errorf("all %d attempts failed, last error: %w", cfg.MaxRetries+1, lastErr)
}
```

```go
// Usage
func (c *UserClient) GetUserWithRetry(ctx context.Context, userID uint) (*User, error) {
    var user *User

    err := retry.Do(ctx, retry.DefaultConfig(), func() error {
        resp, err := http.Get(fmt.Sprintf("http://user-service:3002/users/%d", userID))
        if err != nil {
            return err // Transient error → retry
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", resp.StatusCode) // 5xx → retry
        }

        if resp.StatusCode == 404 {
            return nil // 404 is not retryable — user doesn't exist
        }

        return json.NewDecoder(resp.Body).Decode(&user)
    })

    return user, err
}
```

### Retry Timeline

```
Attempt 1: Call service → FAIL
  Wait: 500ms + jitter (~125ms)
Attempt 2: Call service → FAIL
  Wait: 1000ms + jitter (~250ms)
Attempt 3: Call service → FAIL
  Wait: 2000ms + jitter (~500ms)
Attempt 4: Call service → SUCCESS ✅

Total time: ~4 seconds (vs immediate failure)


Tanpa jitter (semua client retry bersamaan):
  t=0:    1000 clients call → service overwhelmed → FAIL
  t=500ms: 1000 clients retry → service STILL overwhelmed → FAIL
  t=1.5s:  1000 clients retry → STILL overwhelmed → FAIL
  → Service never recovers

Dengan jitter (clients retry di waktu random):
  t=375ms:  333 clients retry → some succeed
  t=625ms:  333 clients retry → more succeed
  t=500ms:  334 clients retry → rest succeed
  → Service recovers gradually
```

### Context Timeout

Di Go, **context** itu cara standard untuk propagate timeout dan cancellation.

```go
// Set timeout di handler level
func (h *TaskHandler) GetTask(c *gin.Context) {
    // 5 second timeout for the entire request
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    taskID := c.Param("id")

    // This call will be cancelled if it takes > 5 seconds
    task, err := h.taskService.GetByID(ctx, taskID)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, task)
}

// Timeout propagated ke downstream calls
func (s *TaskService) GetByID(ctx context.Context, taskID string) (*Task, error) {
    task, err := s.repo.FindByID(ctx, taskID) // DB query respects ctx timeout
    if err != nil {
        return nil, err
    }

    // This gRPC call also respects the same timeout
    user, err := s.userClient.GetUser(ctx, task.AssigneeID) // ctx carries timeout
    if err != nil {
        // Graceful degradation
        task.AssigneeName = "Unknown"
    } else {
        task.AssigneeName = user.Name
    }

    return task, nil
}
```

### Retry + Circuit Breaker + Timeout = The Trifecta

```
Best practice: combine semua 3 pattern:

1. TIMEOUT: Jangan nunggu selamanya (5 seconds max)
2. RETRY: Coba lagi kalau gagal (max 3 retries, exponential backoff)
3. CIRCUIT BREAKER: Stop retry kalau service beneran down

Flow:
  Request → Circuit Breaker check
    ↓ (CLOSED? proceed)
    Retry loop:
      ↓
      Call service with TIMEOUT (5s)
        ↓ Success? → Return result
        ↓ Fail? → Circuit Breaker record failure
        ↓ Wait (exponential backoff + jitter)
        ↓ Retry...
    ↓ All retries failed?
    Circuit Breaker update (maybe switch to OPEN)
    ↓
    Return error / fallback
```

---

## 18.13 Distributed Tracing

### Problem

Di monolith, kalau ada error, kamu lihat **stack trace** — langsung tau error di file mana, line berapa.

Di microservices, 1 request bisa lewat **5+ services**. Error di mana?

```
Client → API Gateway → Auth Service → Board Service → Task Service → Notification Service
                                                         ↑
                                                     Error di sini!

Tapi yang client lihat cuma "500 Internal Server Error" dari API Gateway.
Log Auth Service: "request received, forwarded to Board Service ✅"
Log Board Service: "request received, forwarded to Task Service ✅"
Log Task Service: "ERROR: connection refused to notification-service:3005"
Log Notification Service: (no logs — it's down)

Tanpa tracing, kamu harus MANUAL check log setiap service.
Dengan 20 services, ini nightmare.
```

### Solution: Distributed Tracing

Setiap request dapat **Trace ID** yang unique. Trace ID ini di-propagate ke SEMUA service yang terlibat. Kamu bisa query by Trace ID untuk lihat **entire journey** dari 1 request.

```
Trace ID: abc-123-def
━━━━━━━━━━━━━━━━━━━

┌─ Trace: abc-123-def ─────────────────────────────────────┐
│                                                           │
│  Span 1: API Gateway (2ms)                               │
│  ├── Span 2: Auth Service - ValidateToken (1ms)          │
│  ├── Span 3: Task Service - GetTask (15ms)               │
│  │   ├── Span 4: Task DB Query (3ms)                     │
│  │   └── Span 5: User Service - GetUser (8ms) ← SLOW!   │
│  │       └── Span 6: User DB Query (5ms)                 │
│  └── Span 7: Notification Service - Error! (timeout)     │
│                                                           │
│  Total: 26ms (but Notification timed out)                 │
└───────────────────────────────────────────────────────────┘
```

### Konsep: Trace dan Span

```
Trace = 1 request dari awal sampai akhir (across all services)
Span  = 1 operasi di 1 service (function call, DB query, HTTP call)

1 Trace punya banyak Spans.
Span bisa punya child Spans (parent-child relationship).
```

### Setup OpenTelemetry di Go

OpenTelemetry (OTel) itu **standard** untuk distributed tracing. Vendor-neutral — bisa kirim data ke Jaeger, Zipkin, Datadog, dll.

```go
// internal/telemetry/tracer.go
package telemetry

import (
    "context"
    "fmt"
    "log"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(ctx context.Context, serviceName, collectorURL string) (func(), error) {
    // Create OTLP exporter (send traces to Jaeger/Collector)
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(collectorURL),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create exporter: %w", err)
    }

    // Create resource (service metadata)
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create resource: %w", err)
    }

    // Create tracer provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    // Set global tracer provider
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.TraceContext{})

    log.Printf("Tracer initialized for service: %s", serviceName)

    // Return cleanup function
    cleanup := func() {
        if err := tp.Shutdown(ctx); err != nil {
            log.Printf("Error shutting down tracer: %v", err)
        }
    }

    return cleanup, nil
}
```

```go
// Gin middleware untuk tracing
// internal/middleware/tracing.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/trace"
)

func Tracing(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)

    return func(c *gin.Context) {
        // Extract trace context from incoming request headers
        ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(),
            propagation.HeaderCarrier(c.Request.Header))

        // Start a new span
        ctx, span := tracer.Start(ctx, c.Request.Method+" "+c.FullPath(),
            trace.WithAttributes(
                attribute.String("http.method", c.Request.Method),
                attribute.String("http.url", c.Request.URL.String()),
                attribute.String("http.route", c.FullPath()),
            ),
        )
        defer span.End()

        // Set trace context in gin context
        c.Request = c.Request.WithContext(ctx)

        // Process request
        c.Next()

        // Add response attributes
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
        )
    }
}
```

```go
// Propagate trace context ke downstream service call
// internal/client/http_client.go
package client

import (
    "context"
    "net/http"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

func DoRequestWithTracing(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Inject trace context into outgoing request headers
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    return http.DefaultClient.Do(req)
}
```

```go
// main.go — Setup
func main() {
    ctx := context.Background()

    // Init tracer
    cleanup, err := telemetry.InitTracer(ctx, "task-service", "jaeger:4317")
    if err != nil {
        log.Fatalf("Failed to init tracer: %v", err)
    }
    defer cleanup()

    r := gin.Default()
    r.Use(middleware.Tracing("task-service"))

    // ... routes
    r.Run(":3004")
}
```

### Jaeger UI

Setelah setup, kamu bisa buka Jaeger UI di `http://localhost:16686` dan:
- Search traces by service name
- Filter by duration, status code, error
- See full trace waterfall (request flow across services)
- Identify bottlenecks (span mana yang paling lambat?)

---

## 18.14 Distributed Logging

### Problem: Log Tersebar di Banyak Service

```
1 request masuk → lewat 5 services → setiap service punya log sendiri

Auth Service log: "Token validated for user 123"
Board Service log: "Board 456 accessed by user 123"
Task Service log: "ERROR: failed to create task"
User Service log: (nothing relevant)
Notification Service log: (nothing relevant)

Pertanyaan: Log mana yang terkait request yang sama?
Jawaban: GA TAU kalau ga ada Correlation ID.
```

### Solution: Correlation ID

Setiap request dapat **Correlation ID** (atau Request ID) yang unique. ID ini di-propagate ke semua service dan di-include di SEMUA log entries.

```go
// middleware/correlation.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

func CorrelationID() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if correlation ID already exists (from upstream service)
        correlationID := c.GetHeader("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }

        // Set in context
        c.Set("correlationID", correlationID)

        // Set in response header
        c.Header("X-Correlation-ID", correlationID)

        // Set in request header (for downstream propagation)
        c.Request.Header.Set("X-Correlation-ID", correlationID)

        c.Next()
    }
}
```

### Structured Logging dengan zerolog

```go
// internal/logger/logger.go
package logger

import (
    "os"
    "time"

    "github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(serviceName string) {
    zerolog.TimeFieldFormat = time.RFC3339

    Log = zerolog.New(os.Stdout).With().
        Timestamp().
        Str("service", serviceName).
        Logger()
}

// WithCorrelationID returns logger with correlation ID
func WithCorrelationID(correlationID string) zerolog.Logger {
    return Log.With().
        Str("correlation_id", correlationID).
        Logger()
}
```

```go
// Usage di handler
func (h *TaskHandler) Create(c *gin.Context) {
    correlationID := c.GetString("correlationID")
    log := logger.WithCorrelationID(correlationID)

    log.Info().Msg("Creating task")

    var req model.CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        log.Error().Err(err).Msg("Invalid request body")
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    task, err := h.service.Create(c.Request.Context(), req)
    if err != nil {
        log.Error().Err(err).Msg("Failed to create task")
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    log.Info().
        Uint("task_id", task.ID).
        Str("title", task.Title).
        Msg("Task created successfully")

    c.JSON(201, task)
}
```

### Log Output (Structured JSON)

```json
{"level":"info","service":"task-service","correlation_id":"abc-123-def","time":"2026-05-01T10:00:00Z","message":"Creating task"}
{"level":"info","service":"task-service","correlation_id":"abc-123-def","task_id":42,"title":"Fix bug","time":"2026-05-01T10:00:01Z","message":"Task created successfully"}
{"level":"info","service":"notification-service","correlation_id":"abc-123-def","time":"2026-05-01T10:00:01Z","message":"Sending notification for task 42"}
```

Semua log dengan `correlation_id: abc-123-def` = 1 request flow. Tinggal query!

### Central Logging (ELK / Loki)

```
Service logs → Fluentd/Promtail (collector) → Elasticsearch/Loki (storage) → Kibana/Grafana (query)

┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Auth    │     │  Task    │     │  Board   │     │  Notif   │
│  Service │     │  Service │     │  Service │     │  Service │
│  (logs)  │     │  (logs)  │     │  (logs)  │     │  (logs)  │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     └────────────────┴────────────────┴────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Log Collector     │
                    │  (Promtail/        │
                    │   Fluentd)         │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Log Storage       │
                    │  (Loki /           │
                    │   Elasticsearch)   │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Query & Dashboard │
                    │  (Grafana /        │
                    │   Kibana)          │
                    └───────────────────┘

Query: correlation_id = "abc-123-def"
→ Dapet SEMUA log dari SEMUA service untuk 1 request
```

### ELK vs Loki

| Aspek | ELK (Elasticsearch) | Loki (Grafana) |
|-------|-------------------|----------------|
| **Full-text search** | Yes (powerful) | No (label-based) |
| **Resource usage** | Heavy (RAM hungry) | Lightweight |
| **Cost** | Expensive at scale | Cheap |
| **Setup** | Complex | Simple |
| **Best for** | Complex log analysis | Simple log queries |
| **Integration** | Kibana | Grafana (udah ada dari Bab 15!) |

Untuk project kita, **Loki** lebih cocok karena udah pakai Grafana dari Bab 15, dan ga butuh full-text search yang complex.

---

## 18.15 Database per Service

### Aturan Emas

> Setiap service **memiliki** datanya sendiri. Tidak ada service lain yang boleh langsung akses database service lain.

```
✅ Service owns its data:
  Task Service → task_db (hanya Task Service yang read/write)
  User Service → user_db (hanya User Service yang read/write)

  Kalau Task Service butuh data user? → Call User Service API/gRPC
  JANGAN direct query ke user_db!

❌ Shared database:
  Task Service → shared_db ← User Service juga query

  Problem:
  - Schema change di user table → break Task Service
  - Task Service bisa corrupt user data
  - Ga bisa deploy/scale independently
  - Coupling at database level = worst kind of coupling
```

### Data Duplication is OK

Di microservices, **data duplication** itu bukan anti-pattern. Itu **tradeoff yang sadar**.

```
Monolith approach:
  SELECT t.title, u.name AS assignee_name
  FROM tasks t
  JOIN users u ON t.assignee_id = u.id
  → Simple JOIN, 1 query, consistent data

Microservices approach:
  Option 1: API Call (synchronous)
    1. Task Service: SELECT * FROM tasks WHERE id = 1
    2. Task Service → User Service: GET /users/5
    3. Combine data in Task Service
    → 2 calls, slower, but always fresh data

  Option 2: Data Duplication (cached copy)
    1. Task Service stores: assignee_id AND assignee_name
    2. When user updates name → event → Task Service updates cached name
    → 1 query, fast, but data might be stale for a few seconds

  Option 3: CQRS Read Model
    1. Write: Task Service creates task (task_db)
    2. Event: task.created → Read Model Service
    3. Read Model Service: pre-join task + user data → read_db
    4. Read: query read_db (denormalized, fast)
    → Complex setup, but best read performance
```

### Event-Driven Data Sync

```
User changes name: "Irzan" → "Irzan Aldi"

1. User Service updates user_db → name = "Irzan Aldi"
2. User Service publishes: user.updated {id: 5, name: "Irzan Aldi"}
3. Task Service receives event
4. Task Service updates task_db: UPDATE tasks SET assignee_name = "Irzan Aldi" WHERE assignee_id = 5

Timeline:
  t=0ms:   User Service updates → user_db has "Irzan Aldi"
  t=0ms:   Task Service still has "Irzan" → INCONSISTENT (briefly)
  t=50ms:  Event arrives at Task Service
  t=55ms:  Task Service updates → task_db has "Irzan Aldi" → CONSISTENT

Ini namanya EVENTUAL CONSISTENCY.
Data AKAN konsisten, tapi ga INSTANT. Ada delay beberapa ms/detik.
```

### Kapan Eventual Consistency OK?

```
OK:
  - User profile name di task assignee → stale beberapa detik ga masalah
  - Notification count → eventual consistency fine
  - Search index → delay beberapa detik acceptable

NOT OK:
  - Bank transfer → uang harus konsisten SEKARANG
  - Inventory count saat checkout → harus real-time
  - Auth token validation → harus accurate NOW

  Untuk kasus NOT OK → pakai synchronous call atau saga pattern
```

---

## 18.16 Docker Compose Microservices

Sekarang kita satuin semua services di Docker Compose untuk development.

```yaml
# docker-compose.yml
version: "3.8"

services:
  # ============================================
  # API GATEWAY
  # ============================================
  api-gateway:
    build:
      context: ./api-gateway
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - AUTH_SERVICE_URL=http://auth-service:3001
      - AUTH_SERVICE_GRPC=auth-service:50051
      - USER_SERVICE_URL=http://user-service:3002
      - BOARD_SERVICE_URL=http://board-service:3003
      - TASK_SERVICE_URL=http://task-service:3004
      - NOTIFICATION_SERVICE_URL=http://notification-service:3005
      - RATE_LIMIT=100
    depends_on:
      auth-service:
        condition: service_healthy
      user-service:
        condition: service_healthy
      board-service:
        condition: service_healthy
      task-service:
        condition: service_healthy
    networks:
      - microservices

  # ============================================
  # AUTH SERVICE
  # ============================================
  auth-service:
    build:
      context: ./auth-service
      dockerfile: Dockerfile
    ports:
      - "3001:3001"
      - "50051:50051"  # gRPC
    environment:
      - PORT=3001
      - GRPC_PORT=50051
      - DB_HOST=auth-db
      - DB_PORT=5432
      - DB_USER=auth
      - DB_PASSWORD=auth_secret
      - DB_NAME=auth_db
      - JWT_SECRET=super-secret-jwt-key
      - NATS_URL=nats://nats:4222
    depends_on:
      auth-db:
        condition: service_healthy
      nats:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3001/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  auth-db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=auth
      - POSTGRES_PASSWORD=auth_secret
      - POSTGRES_DB=auth_db
    volumes:
      - auth-db-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U auth -d auth_db"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  # ============================================
  # USER SERVICE
  # ============================================
  user-service:
    build:
      context: ./user-service
      dockerfile: Dockerfile
    ports:
      - "3002:3002"
      - "50052:50052"  # gRPC
    environment:
      - PORT=3002
      - GRPC_PORT=50052
      - DB_HOST=user-db
      - DB_PORT=5432
      - DB_USER=user_svc
      - DB_PASSWORD=user_secret
      - DB_NAME=user_db
      - NATS_URL=nats://nats:4222
    depends_on:
      user-db:
        condition: service_healthy
      nats:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3002/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  user-db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=user_svc
      - POSTGRES_PASSWORD=user_secret
      - POSTGRES_DB=user_db
    volumes:
      - user-db-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user_svc -d user_db"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  # ============================================
  # BOARD SERVICE
  # ============================================
  board-service:
    build:
      context: ./board-service
      dockerfile: Dockerfile
    ports:
      - "3003:3003"
    environment:
      - PORT=3003
      - DB_HOST=board-db
      - DB_PORT=5432
      - DB_USER=board_svc
      - DB_PASSWORD=board_secret
      - DB_NAME=board_db
      - AUTH_SERVICE_GRPC=auth-service:50051
      - NATS_URL=nats://nats:4222
    depends_on:
      board-db:
        condition: service_healthy
      nats:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3003/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  board-db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=board_svc
      - POSTGRES_PASSWORD=board_secret
      - POSTGRES_DB=board_db
    volumes:
      - board-db-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U board_svc -d board_db"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  # ============================================
  # TASK SERVICE
  # ============================================
  task-service:
    build:
      context: ./task-service
      dockerfile: Dockerfile
    ports:
      - "3004:3004"
    environment:
      - PORT=3004
      - DB_HOST=task-db
      - DB_PORT=5432
      - DB_USER=task_svc
      - DB_PASSWORD=task_secret
      - DB_NAME=task_db
      - USER_SERVICE_GRPC=user-service:50052
      - AUTH_SERVICE_GRPC=auth-service:50051
      - NATS_URL=nats://nats:4222
    depends_on:
      task-db:
        condition: service_healthy
      nats:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3004/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  task-db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=task_svc
      - POSTGRES_PASSWORD=task_secret
      - POSTGRES_DB=task_db
    volumes:
      - task-db-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U task_svc -d task_db"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  # ============================================
  # NOTIFICATION SERVICE
  # ============================================
  notification-service:
    build:
      context: ./notification-service
      dockerfile: Dockerfile
    ports:
      - "3005:3005"
    environment:
      - PORT=3005
      - REDIS_URL=redis://redis:6379
      - NATS_URL=nats://nats:4222
      - SMTP_HOST=mailhog
      - SMTP_PORT=1025
    depends_on:
      redis:
        condition: service_healthy
      nats:
        condition: service_started
    networks:
      - microservices

  # ============================================
  # INFRASTRUCTURE
  # ============================================
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - microservices

  nats:
    image: nats:2-alpine
    ports:
      - "4222:4222"   # Client
      - "8222:8222"   # Monitoring
    command: ["--js"]  # Enable JetStream for persistence
    networks:
      - microservices

  # ============================================
  # OBSERVABILITY
  # ============================================
  jaeger:
    image: jaegertracing/all-in-one:1.55
    ports:
      - "16686:16686"  # UI
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - microservices

  mailhog:
    image: mailhog/mailhog:v1.0.1
    ports:
      - "8025:8025"  # Web UI
      - "1025:1025"  # SMTP
    networks:
      - microservices

# ============================================
# VOLUMES
# ============================================
volumes:
  auth-db-data:
  user-db-data:
  board-db-data:
  task-db-data:
  redis-data:

# ============================================
# NETWORKS
# ============================================
networks:
  microservices:
    driver: bridge
```

### Jalankan

```bash
# Build dan start semua services
docker compose up -d --build

# Lihat status
docker compose ps

# Expected output:
# NAME                    STATUS    PORTS
# api-gateway             Up        0.0.0.0:8080->8080/tcp
# auth-service            Up        0.0.0.0:3001->3001/tcp, 0.0.0.0:50051->50051/tcp
# user-service            Up        0.0.0.0:3002->3002/tcp, 0.0.0.0:50052->50052/tcp
# board-service           Up        0.0.0.0:3003->3003/tcp
# task-service            Up        0.0.0.0:3004->3004/tcp
# notification-service    Up        0.0.0.0:3005->3005/tcp
# auth-db                 Up        5432/tcp
# user-db                 Up        5432/tcp
# board-db                Up        5432/tcp
# task-db                 Up        5432/tcp
# redis                   Up        0.0.0.0:6379->6379/tcp
# nats                    Up        0.0.0.0:4222->4222/tcp
# jaeger                  Up        0.0.0.0:16686->16686/tcp
# mailhog                 Up        0.0.0.0:8025->8025/tcp

# Lihat logs specific service
docker compose logs -f task-service

# Test via API Gateway
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Irzan","email":"irzan@test.com","password":"secret123"}'

# Stop semua
docker compose down
```

Perhatiin: **14 containers** untuk 1 project. Ini trade-off microservices — complexity di infrastructure.

---

## 18.17 Kubernetes Microservices

### Namespace Strategy

```yaml
# Option 1: Namespace per environment (recommended for small teams)
apiVersion: v1
kind: Namespace
metadata:
  name: task-manager-dev
---
apiVersion: v1
kind: Namespace
metadata:
  name: task-manager-staging
---
apiVersion: v1
kind: Namespace
metadata:
  name: task-manager-prod
```

```yaml
# Option 2: Namespace per service (for large teams)
# Biasanya kalau setiap service dikelola tim berbeda
apiVersion: v1
kind: Namespace
metadata:
  name: auth-service
---
apiVersion: v1
kind: Namespace
metadata:
  name: task-service
```

### K8s Manifests untuk Task Service

```yaml
# k8s/task-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-service
  namespace: task-manager
  labels:
    app: task-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: task-service
  template:
    metadata:
      labels:
        app: task-service
    spec:
      containers:
        - name: task-service
          image: your-registry/task-service:v1.0.0
          ports:
            - name: http
              containerPort: 3004
          env:
            - name: PORT
              value: "3004"
            - name: DB_HOST
              value: "task-db-service"
            - name: DB_PORT
              value: "5432"
            - name: DB_NAME
              value: "task_db"
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: task-db-secret
                  key: username
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: task-db-secret
                  key: password
            - name: USER_SERVICE_GRPC
              value: "user-service:50052"  # K8s DNS!
            - name: AUTH_SERVICE_GRPC
              value: "auth-service:50051"  # K8s DNS!
            - name: NATS_URL
              value: "nats://nats:4222"    # K8s DNS!
            - name: JAEGER_ENDPOINT
              value: "jaeger-collector:4317"
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 256Mi
          livenessProbe:
            httpGet:
              path: /health
              port: 3004
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /health
              port: 3004
            initialDelaySeconds: 5
            periodSeconds: 10
---
# k8s/task-service/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: task-service
  namespace: task-manager
spec:
  selector:
    app: task-service
  ports:
    - name: http
      port: 3004
      targetPort: 3004
---
# k8s/task-service/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: task-service-hpa
  namespace: task-manager
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: task-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### Ingress untuk API Gateway

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: task-manager-ingress
  namespace: task-manager
  annotations:
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.taskmanager.com
      secretName: taskmanager-tls
  rules:
    - host: api.taskmanager.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  number: 8080
```

### Service Mesh (Konsep)

Service mesh itu **infrastructure layer** yang handle service-to-service communication. Kamu ga perlu implement circuit breaker, retry, mTLS, tracing di **setiap service**. Service mesh yang handle.

```
Tanpa Service Mesh:
━━━━━━━━━━━━━━━━━━
Setiap service implement sendiri:
  - Circuit breaker ← code di setiap service
  - Retry logic ← code di setiap service
  - mTLS ← config di setiap service
  - Tracing ← code di setiap service
  - Rate limiting ← code di setiap service

Dengan Service Mesh (e.g., Istio):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────────────────────────┐
│           Pod                │
│  ┌────────────┐  ┌────────┐ │
│  │   Your     │  │ Sidecar│ │
│  │   Service  │←→│ Proxy  │ │ ← Envoy proxy (auto-injected)
│  │   (Go app) │  │(Envoy) │ │
│  └────────────┘  └───┬────┘ │
└──────────────────────┼──────┘
                       │
              All traffic goes through
              sidecar proxy — handles:
              - mTLS (encryption)
              - Circuit breaking
              - Retry
              - Load balancing
              - Tracing
              - Metrics

              Your app code DOESN'T change!
```

**Popular service meshes:**
- **Istio** — Most feature-rich, most complex, backed by Google
- **Linkerd** — Lightweight, simple, written in Rust
- **Consul Connect** — HashiCorp's service mesh, integrates with Consul

> **Kapan pakai service mesh?** Kalau kamu punya **20+ microservices** dan tim DevOps yang dedicated. Untuk 5-10 services, implement circuit breaker dan retry di code aja cukup. Service mesh itu powerful tapi complex.

---

## 18.18 Testing Microservices

### Testing Pyramid untuk Microservices

```
                    ╱╲
                   ╱  ╲
                  ╱ E2E╲           ← Paling sedikit, paling lambat, paling mahal
                 ╱ Tests ╲
                ╱──────────╲
               ╱ Integration╲      ← Medium — test antar services
              ╱    Tests      ╲
             ╱────────────────╲
            ╱  Contract Tests  ╲   ← Test API contract antar services
           ╱────────────────────╲
          ╱     Component Tests  ╲  ← Test 1 service in isolation
         ╱────────────────────────╲
        ╱       Unit Tests         ╲ ← Paling banyak, paling cepat, paling murah
       ╱────────────────────────────╲
```

### 1. Unit Tests

Test logic di dalam 1 service. Mock semua external dependencies.

```go
// task-service/internal/service/task_test.go
func TestCreateTask_Success(t *testing.T) {
    // Mock repository
    mockRepo := &MockTaskRepository{}
    mockRepo.On("Create", mock.Anything).Return(&model.Task{
        ID:    1,
        Title: "Test Task",
    }, nil)

    // Mock user client (gRPC)
    mockUserClient := &MockUserClient{}

    svc := NewTaskService(mockRepo, mockUserClient)

    task, err := svc.Create(context.Background(), CreateTaskRequest{
        Title:   "Test Task",
        BoardID: 1,
    })

    assert.NoError(t, err)
    assert.Equal(t, "Test Task", task.Title)
    mockRepo.AssertExpectations(t)
}
```

### 2. Contract Tests (Pact)

**Contract test** memastikan bahwa **provider** (service yang dipanggil) masih memenuhi **contract** yang diharapkan **consumer** (service yang memanggil).

```
Problem tanpa contract test:
  Task Service expect: GET /users/123 → {"id": 123, "name": "Irzan"}
  User Service returns: GET /users/123 → {"userId": 123, "fullName": "Irzan"}

  Field name berubah! Task Service BREAK.
  Ini ga ketahuan sampai integration test (atau worse, production).

Dengan contract test:
  Consumer (Task Service) define: "I expect GET /users/123 to return {id, name}"
  Provider (User Service) verify: "Do I still return {id, name}?"

  Kalau provider change response format → contract test FAIL → ketahuan di CI/CD
```

```go
// task-service/contract_test.go (consumer side)
func TestUserServiceContract(t *testing.T) {
    // Define expected interaction
    pact := dsl.Pact{
        Consumer: "task-service",
        Provider: "user-service",
    }

    pact.AddInteraction().
        Given("User 123 exists").
        UponReceiving("A request for user 123").
        WithRequest(dsl.Request{
            Method: "GET",
            Path:   dsl.String("/users/123"),
        }).
        WillRespondWith(dsl.Response{
            Status: 200,
            Body: dsl.MapMatcher{
                "id":    dsl.Like(123),
                "name":  dsl.Like("Irzan"),
                "email": dsl.Like("irzan@test.com"),
            },
        })

    // Test
    err := pact.Verify(func() error {
        // Call the user client with mock server URL
        client := NewUserClient(fmt.Sprintf("http://localhost:%d", pact.Server.Port))
        user, err := client.GetUser(context.Background(), 123)

        assert.NoError(t, err)
        assert.Equal(t, "Irzan", user.Name)
        return nil
    })

    assert.NoError(t, err)
}
```

### 3. Integration Tests

Test interaksi nyata antar services (biasanya pakai Docker Compose).

```go
// integration_test.go
func TestCreateTaskIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup: docker compose up -d (or testcontainers)

    // 1. Register user via Auth Service
    registerResp := httpPost("http://localhost:8080/api/v1/auth/register", map[string]string{
        "name":     "Test User",
        "email":    "test@example.com",
        "password": "secret123",
    })
    assert.Equal(t, 201, registerResp.StatusCode)

    // 2. Login
    loginResp := httpPost("http://localhost:8080/api/v1/auth/login", map[string]string{
        "email":    "test@example.com",
        "password": "secret123",
    })
    token := loginResp.Body["access_token"]

    // 3. Create board
    boardResp := httpPostAuth("http://localhost:8080/api/v1/boards", token, map[string]string{
        "title": "Test Board",
    })
    boardID := boardResp.Body["id"]

    // 4. Create task
    taskResp := httpPostAuth(
        fmt.Sprintf("http://localhost:8080/api/v1/boards/%v/tasks", boardID),
        token,
        map[string]string{"title": "Test Task"},
    )
    assert.Equal(t, 201, taskResp.StatusCode)
    assert.Equal(t, "Test Task", taskResp.Body["title"])

    // 5. Verify notification was sent (check MailHog)
    time.Sleep(2 * time.Second) // Wait for async event
    mailResp := httpGet("http://localhost:8025/api/v2/messages")
    assert.Greater(t, len(mailResp.Body["items"]), 0)
}
```

### 4. End-to-End Tests

Test full flow dari perspective user. Biasanya pakai tools kayak Playwright atau Cypress kalau ada frontend.

```
E2E Test Flow:
  1. User opens app
  2. User registers
  3. User logs in
  4. User creates board
  5. User creates task
  6. User assigns task to another user
  7. Verify: notification email received
  8. User marks task as done
  9. Verify: board shows completed task
```

### Testing Strategy Summary

```
Unit Tests:
  - Run: Every commit
  - Speed: Milliseconds
  - Coverage: Business logic, utils, models
  - Mock: Everything external

Contract Tests:
  - Run: Every PR
  - Speed: Seconds
  - Coverage: API contracts between services
  - Mock: Provider behavior

Integration Tests:
  - Run: Every merge to main
  - Speed: Minutes
  - Coverage: Service interactions, DB queries
  - Mock: Nothing (real services via Docker)

E2E Tests:
  - Run: Before release
  - Speed: Minutes
  - Coverage: Full user journeys
  - Mock: Nothing
```

---

## 18.19 Migration Strategy — Monolith to Microservices

### Strangler Fig Pattern

Dinamakan dari **strangler fig** — tanaman yang tumbuh di sekitar pohon lain, pelan-pelan menggantikan pohon aslinya.

```
Phase 1: Monolith + Facade
━━━━━━━━━━━━━━━━━━━━━━━━━
                          ┌──────────────┐
Client ──→ API Gateway ──→│   MONOLITH    │
           (facade)       │              │
                          │ Auth         │
                          │ User         │
                          │ Board        │
                          │ Task         │
                          │ Notification │
                          └──────────────┘

Phase 2: Extract 1 Service (Notification)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
                          ┌──────────────┐     ┌──────────────┐
Client ──→ API Gateway ──→│   MONOLITH    │────→│ Notification │
           │              │              │     │   Service    │
           │              │ Auth         │     └──────────────┘
           │              │ User         │
           │              │ Board        │
           └──────────────│ Task         │
                          └──────────────┘

Phase 3: Extract More Services
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
                          ┌──────────────┐     ┌──────────────┐
Client ──→ API Gateway ──→│   MONOLITH    │────→│ Notification │
           │              │              │     └──────────────┘
           │              │ Auth         │
           │              │ User         │     ┌──────────────┐
           │              │ Board        │────→│    Task       │
           └──────────────└──────────────┘     │   Service    │
                                               └──────────────┘

Phase 4: Fully Decomposed
━━━━━━━━━━━━━━━━━━━━━━━━━
Client ──→ API Gateway ──→ Auth Service
                       ──→ User Service
                       ──→ Board Service
                       ──→ Task Service
                       ──→ Notification Service

Monolith GONE. (tapi ini bisa butuh bulan/tahun)
```

### Step-by-Step Migration Plan untuk Task Manager

```
Step 1: Add API Gateway di depan monolith (Week 1)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  - Deploy API Gateway (Nginx/Traefik/custom Go)
  - Route SEMUA traffic ke monolith
  - No functionality change — just add a facade
  - Ini "zero risk" — kalau gateway error, remove aja

Step 2: Extract Notification Service (Week 2-3)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Kenapa mulai dari Notification?
  - Paling sedikit coupling ke service lain
  - Event-driven (fire and forget)
  - Kalau error, ga affect core functionality
  - Perfect buat belajar dan validate architecture

  Steps:
  a. Create notification-service repo
  b. Implement event consumer (NATS/Kafka)
  c. Deploy notification service
  d. Monolith publish events instead of direct notification
  e. Test: notifications masih jalan
  f. Remove notification code dari monolith

Step 3: Extract Auth Service (Week 4-5)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Steps:
  a. Create auth-service repo
  b. Implement register, login, validate token (gRPC)
  c. Deploy auth service
  d. API Gateway route /auth/* ke auth service
  e. Other routes still go to monolith
  f. Monolith validate token via auth service gRPC
  g. Test: auth flow masih jalan
  h. Remove auth code dari monolith

Step 4: Extract User Service (Week 6-7)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  (similar pattern)

Step 5: Extract Board Service (Week 8-9)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  (similar pattern)

Step 6: Extract Task Service (Week 10-11)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  (similar pattern — this is the last one)

Step 7: Decommission Monolith (Week 12)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  - Verify semua traffic lewat microservices
  - Shutdown monolith
  - Celebrate 🎉 (maaf, ini bukan emoji, ini ASCII)

Total: ~3 bulan untuk full migration
Reality: biasanya 6-12 bulan di company beneran
```

### Tips Migration

```
DO:
  ✅ Mulai dari service yang paling independent (Notification)
  ✅ Keep monolith running selama migration
  ✅ Use feature flags — bisa switch antara monolith dan microservice
  ✅ Monitor closely setiap kali extract service
  ✅ Database migration terpisah — copy data, sync via events
  ✅ Run both systems in parallel sebelum cutover

DON'T:
  ❌ Big bang rewrite (rewrite semua sekaligus)
  ❌ Extract semua service bersamaan
  ❌ Skip testing di setiap phase
  ❌ Ignore data migration complexity
  ❌ Underestimate timeline — always takes longer
```

---

## 18.20 Common Anti-Patterns

### 1. Distributed Monolith

```
Anti-pattern: Kamu pecah monolith jadi services, tapi semua service
             HARUS di-deploy bareng. Change di Service A → deploy semua.
             Itu bukan microservices, itu DISTRIBUTED MONOLITH.

Cause:
  - Shared database
  - Synchronous chain (A → B → C → D)
  - Shared code libraries yang sering berubah
  - Tight API coupling

Solution:
  - Database per service
  - Async communication where possible
  - API versioning
  - Independent deployment pipeline per service
```

### 2. Shared Database

```
Anti-pattern: Semua services connect ke 1 database.

Problem:
  - Schema change affect ALL services
  - No independent deployment
  - Performance coupling (Service A query berat → Service B slow)
  - Data ownership unclear

Solution:
  - Database per service
  - Data sync via events
  - API calls for cross-service data
```

### 3. Synchronous Chain

```
Anti-pattern: A → B → C → D → E (semua synchronous)

Client → API GW → Auth → User → Board → Task → Notification
                                                  ↑ Timeout!
Total latency = sum of ALL services
If ANY service slow/down = entire request fails

Solution:
  - Minimize sync chain length (max 2-3 hops)
  - Use async (events) where possible
  - Circuit breaker at each hop
  - Parallel calls where possible
```

### 4. Nano-services (Terlalu Kecil)

```
Anti-pattern: Pecah service terlalu kecil.

  - LoginService
  - RegisterService
  - PasswordResetService
  - EmailValidationService
  → 4 services untuk auth? Overkill!

Problem:
  - Network overhead > benefit of separation
  - Deployment complexity increases
  - More services = more monitoring, more CI/CD, more infra cost

Solution:
  - Group by business domain, bukan by function
  - AuthService (login + register + password + validation) = 1 service
  - Rule of thumb: kalau service cuma punya 1-2 endpoints → too small
```

### 5. No Monitoring

```
Anti-pattern: Deploy microservices tanpa proper monitoring.

Di monolith: 1 app crash → 1 alert → check 1 log file
Di microservices: any of 10 services crash → need to know WHICH one

Tanpa monitoring:
  - "Something is slow" → which service?
  - "User can't login" → Auth? Gateway? Network?
  - "Task creation fails intermittently" → Task Service? DB? Event bus?

  Debugging microservices tanpa monitoring itu kayak nyari jarum
  di 10 tumpukan jerami. Blindfolded.

Must have:
  - Distributed tracing (Jaeger/Zipkin)
  - Centralized logging (ELK/Loki)
  - Metrics & alerting (Prometheus/Grafana)
  - Health check endpoints
  - Correlation IDs
```

### 6. Ignoring Network Failures

```
Anti-pattern: Code microservices seolah network call selalu success.

// JANGAN kayak gini
user, _ := userClient.GetUser(ctx, userID)  // Ignore error!

// HARUS kayak gini
user, err := userClient.GetUser(ctx, userID)
if err != nil {
    // Handle gracefully:
    // - Return cached data
    // - Return partial response
    // - Return default value
    // - Retry with backoff
    // - Circuit breaker
}
```

---

## 18.21 Kapan TIDAK Pakai Microservices

Ini bab yang PALING penting. Serius.

### Jangan Pakai Microservices Kalau:

```
1. Tim kecil (1-5 developer)
━━━━━━━━━━━━━━━━━━━━━━━━━━
   Microservices butuh DevOps maturity:
   - CI/CD per service
   - Monitoring per service
   - On-call rotation
   - Infrastructure management

   Tim 3 orang ga punya bandwidth buat maintain 10 services + infra.
   Monolith = focus on FEATURES, bukan infrastructure.

2. MVP / Prototype
━━━━━━━━━━━━━━━━━
   Kamu belum tau fitur mana yang bakal survive.
   Invest di architecture premature = waste.

   "Premature optimization is the root of all evil" — Knuth
   Premature decomposition is the root of all microservices evil.

3. Simple CRUD app
━━━━━━━━━━━━━━━━━
   Blog, simple e-commerce, internal tool, portfolio website.
   Monolith + PostgreSQL cukup. Bahkan monolith + SQLite cukup.

   Microservices buat blog itu kayak beli truk trailer buat ke warung.

4. No DevOps maturity
━━━━━━━━━━━━━━━━━━━━
   Kalau tim belum bisa:
   - Containerize app (Docker)
   - Setup CI/CD pipeline
   - Monitor services (Prometheus, Grafana)
   - Manage Kubernetes cluster

   JANGAN microservices. Fix DevOps foundation dulu.

5. Tight deadline
━━━━━━━━━━━━━━━
   Microservices = more upfront setup time.
   - 5 repos instead of 1
   - 5 CI/CD pipelines instead of 1
   - Docker Compose / K8s setup
   - Service communication setup
   - Distributed tracing setup

   Kalau deadline 2 minggu, monolith aja. Ship first.

6. Domain belum jelas
━━━━━━━━━━━━━━━━━━━━
   Kalau kamu masih experimenting, domain boundaries belum clear.
   Salah pecah service → merge lagi = sangat painful.

   Better: monolith dulu → domain boundaries jadi clear → THEN decompose.
```

### The Quote

> "If you can't build a well-structured monolith, what makes you think microservices is the answer?"

Microservices ga fix bad architecture. Microservices **amplify** whatever you have:
- Good architecture → microservices amplify the benefits
- Bad architecture → microservices amplify the problems

### Decision Framework

```
Pertanyaan sebelum pilih microservices:

1. Tim kamu ada berapa orang?
   < 5 → Monolith
   5-15 → Modular monolith, consider microservices
   > 15 → Microservices likely beneficial

2. Sudah punya DevOps maturity?
   No Docker, no CI/CD → Monolith
   Docker + CI/CD + monitoring → Consider microservices
   K8s + full observability → Ready for microservices

3. Domain sudah clear?
   Still exploring → Monolith
   Domain boundaries clear → Microservices OK

4. Ada scaling bottleneck yang specific?
   No → Monolith
   Yes, specific service → Extract that service only

5. Deployment blocking antar tim?
   No → Monolith
   Yes → Microservices
```

---

## 18.22 Tools & Ecosystem

### Go Microservices Frameworks

| Framework | Description | Stars | Use Case |
|-----------|-------------|-------|----------|
| **Go kit** | Toolkit for microservices | 26k+ | Enterprise, production-ready |
| **Go Micro** | Framework for distributed systems | 21k+ | Full-featured framework |
| **Kratos** | By Bilibili (Chinese YouTube) | 23k+ | Cloud-native, gRPC-first |
| **Gin** | HTTP framework (our choice) | 79k+ | REST API, simple services |
| **Fiber** | Express-inspired Go framework | 34k+ | REST API, performance |

```
Rekomendasi:
━━━━━━━━━━━

Pemula / Small project:
  Gin + manual wiring → Kamu paham setiap bagian

Medium project:
  Go kit → Production patterns built-in (transport, endpoint, service layers)

Large project (Chinese tech style):
  Kratos → Opinionated, gRPC-first, used by Bilibili at massive scale

Catatan: Kamu ga HARUS pakai framework khusus microservices.
Gin + gRPC + NATS + standard library = perfectly fine.
Go philosophy: keep it simple.
```

### Communication

| Tool | Type | Use Case |
|------|------|----------|
| **gRPC** | Sync RPC | Internal service-to-service |
| **NATS** | Async messaging | Lightweight pub/sub, events |
| **Kafka** | Event streaming | High-throughput event streaming |
| **RabbitMQ** | Message broker | Task queues, complex routing |
| **Redis Streams** | Streaming | Lightweight streaming (already have Redis) |

### Service Discovery & Configuration

| Tool | Type | Use Case |
|------|------|----------|
| **Kubernetes DNS** | DNS-based | K8s native (recommended) |
| **Consul** | Service registry | Non-K8s, health checking |
| **etcd** | Key-value store | Config, service discovery |
| **Vault** | Secret management | Secrets, certificates |

### Observability

| Tool | Type | Use Case |
|------|------|----------|
| **Jaeger** | Distributed tracing | Trace requests across services |
| **Zipkin** | Distributed tracing | Alternative to Jaeger |
| **Prometheus** | Metrics | Metrics collection & alerting |
| **Grafana** | Dashboards | Visualization |
| **Loki** | Log aggregation | Centralized logging |
| **OpenTelemetry** | Standard | Vendor-neutral telemetry SDK |

### Service Mesh

| Tool | Type | Complexity |
|------|------|------------|
| **Istio** | Service mesh | High (most feature-rich) |
| **Linkerd** | Service mesh | Low (lightweight, Rust-based) |
| **Consul Connect** | Service mesh | Medium |

### Quick Setup Guide

```
Minimal microservices stack:
━━━━━━━━━━━━━━━━━━━━━━━━━━
  Language:      Go (Gin)
  Communication: gRPC (sync) + NATS (async)
  Database:      PostgreSQL per service
  Cache:         Redis
  Container:     Docker + Docker Compose (dev)
  Orchestration: Kubernetes (prod)
  Tracing:       OpenTelemetry + Jaeger
  Logging:       zerolog + Loki
  Monitoring:    Prometheus + Grafana
  CI/CD:         GitHub Actions
  IaC:           Terraform

  Total tools: ~12. Welcome to microservices. 😅
  (Bandingkan monolith: Go + PostgreSQL + Redis + Docker. 4 tools.)
```

---

## 18.23 Cheat Sheet

### Pattern Decision Matrix

```
┌─────────────────────────────────────────────────────────┐
│              PATTERN DECISION MATRIX                     │
├─────────────────────┬───────────────────────────────────┤
│ Situation           │ Pattern                           │
├─────────────────────┼───────────────────────────────────┤
│ Service A needs     │ Synchronous:                      │
│ data from Service B │ - REST (simple, external)         │
│ RIGHT NOW           │ - gRPC (fast, internal)           │
├─────────────────────┼───────────────────────────────────┤
│ Service A needs to  │ Asynchronous:                     │
│ notify Service B    │ - NATS/Kafka (event-driven)       │
│ (fire & forget)     │ - Message Queue (task queue)      │
├─────────────────────┼───────────────────────────────────┤
│ Multi-service       │ Saga Pattern:                     │
│ transaction needed  │ - Choreography (event-based)      │
│                     │ - Orchestration (coordinator)     │
├─────────────────────┼───────────────────────────────────┤
│ Downstream service  │ Circuit Breaker:                  │
│ might be down       │ - sony/gobreaker                  │
│                     │ - Custom implementation           │
├─────────────────────┼───────────────────────────────────┤
│ Network call might  │ Retry + Timeout:                  │
│ fail temporarily    │ - Exponential backoff + jitter    │
│                     │ - Context timeout                 │
├─────────────────────┼───────────────────────────────────┤
│ Need to find        │ Service Discovery:                │
│ service address     │ - K8s DNS (if on K8s)             │
│                     │ - Consul (if not on K8s)          │
├─────────────────────┼───────────────────────────────────┤
│ Debug request       │ Distributed Tracing:              │
│ across services     │ - OpenTelemetry + Jaeger          │
│                     │ - Correlation ID in logs          │
├─────────────────────┼───────────────────────────────────┤
│ Single entry point  │ API Gateway:                      │
│ for clients         │ - Kong/Traefik (production)       │
│                     │ - Custom Go (learning)            │
├─────────────────────┼───────────────────────────────────┤
│ Migrate from        │ Strangler Fig Pattern:            │
│ monolith            │ - Extract 1 service at a time     │
│                     │ - Start with least coupled        │
└─────────────────────┴───────────────────────────────────┘
```

### Communication Patterns Quick Reference

```
┌───────────────┬──────────┬──────────┬──────────────┐
│               │   REST   │  gRPC    │  Event/Queue │
├───────────────┼──────────┼──────────┼──────────────┤
│ Type          │ Sync     │ Sync     │ Async        │
│ Format        │ JSON     │ Protobuf │ JSON/Protobuf│
│ Speed         │ Medium   │ Fast     │ N/A          │
│ Coupling      │ Medium   │ Medium   │ Low          │
│ Use for       │ External │ Internal │ Events       │
│ Debugging     │ Easy     │ Medium   │ Hard         │
│ Type safety   │ No       │ Yes      │ No           │
│ Streaming     │ No       │ Yes      │ Yes          │
│ Failure mode  │ Timeout  │ Timeout  │ At-least-once│
└───────────────┴──────────┴──────────┴──────────────┘
```

### Tools Comparison Quick Reference

```
Messaging:
  Simple pub/sub → NATS
  Task queue → Asynq (Redis) / RabbitMQ
  Event streaming → Kafka
  Already have Redis → Redis Streams

Service Discovery:
  Kubernetes → K8s DNS (free!)
  Non-K8s → Consul
  Simple → Environment variables

API Gateway:
  Production → Kong / Traefik
  AWS → API Gateway
  Learning → Custom Go
  K8s → Ingress Controller

Tracing:
  Standard → OpenTelemetry
  Backend → Jaeger / Zipkin / Tempo

Logging:
  Library → zerolog / slog (Go 1.21+)
  Central → Loki (+ Grafana) / ELK

Service Mesh:
  Feature-rich → Istio
  Lightweight → Linkerd
  HashiCorp stack → Consul Connect
  < 20 services → DON'T NEED IT
```

---

## 18.24 Checklist Bab Ini

### Konsep Dasar

- [ ] Paham apa itu microservices dan bedanya sama monolith
- [ ] Paham analogi warung vs food court
- [ ] Paham kenapa microservices populer DAN kenapa bukan selalu jawaban
- [ ] Paham konsep "monolith first"
- [ ] Bisa jelaskan kapan monolith lebih baik
- [ ] Bisa jelaskan kapan waktunya migrate ke microservices

### Microservices Principles

- [ ] Paham Single Responsibility untuk services
- [ ] Paham independently deployable
- [ ] Paham decentralized data management
- [ ] Paham design for failure
- [ ] Paham smart endpoints, dumb pipes

### Task Manager Decomposition

- [ ] Bisa identify service boundaries (Auth, User, Board, Task, Notification)
- [ ] Paham tanggung jawab setiap service
- [ ] Paham event flow antar services
- [ ] Paham role API Gateway

### Communication Patterns

- [ ] Paham bedanya synchronous vs asynchronous communication
- [ ] Paham kapan pakai REST vs gRPC vs message queue
- [ ] Paham trade-offs setiap pattern

### gRPC

- [ ] Paham apa itu gRPC dan Protocol Buffers
- [ ] Bisa nulis proto file
- [ ] Bisa generate Go code dari proto
- [ ] Bisa implement gRPC server di Go
- [ ] Bisa implement gRPC client di Go
- [ ] Paham kenapa gRPC lebih cepat dari REST

### API Gateway

- [ ] Paham fungsi API Gateway (routing, auth, rate limiting)
- [ ] Paham options (Kong, Traefik, custom Go, AWS API Gateway)
- [ ] Paham bagaimana API Gateway route requests ke services
- [ ] Paham reverse proxy concept

### Service Discovery

- [ ] Paham problem service discovery
- [ ] Paham Kubernetes DNS-based discovery
- [ ] Paham Consul concept
- [ ] Tau DNS name format di Kubernetes

### Event-Driven Architecture

- [ ] Paham bedanya events vs commands
- [ ] Paham event sourcing concept (dan kapan ga perlu)
- [ ] Paham CQRS concept (dan kapan ga perlu)
- [ ] Bisa implement pub/sub dengan NATS di Go

### Saga Pattern

- [ ] Paham problem distributed transactions
- [ ] Paham saga = local transactions + compensations
- [ ] Paham bedanya choreography vs orchestration
- [ ] Bisa implement simple saga orchestrator di Go

### Resilience Patterns

- [ ] Paham circuit breaker (Closed, Open, Half-Open states)
- [ ] Bisa implement circuit breaker (sony/gobreaker atau custom)
- [ ] Paham retry dengan exponential backoff + jitter
- [ ] Paham kenapa naive retry berbahaya
- [ ] Paham context timeout di Go
- [ ] Paham combo: timeout + retry + circuit breaker

### Observability

- [ ] Paham distributed tracing (Trace, Span)
- [ ] Paham OpenTelemetry + Jaeger
- [ ] Bisa setup tracing di Go
- [ ] Paham correlation ID untuk distributed logging
- [ ] Paham structured logging (zerolog/slog)
- [ ] Paham central logging (ELK/Loki)

### Database per Service

- [ ] Paham aturan "each service owns its data"
- [ ] Paham kenapa shared database itu anti-pattern
- [ ] Paham data duplication dan eventual consistency
- [ ] Paham event-driven data sync

### Infrastructure

- [ ] Bisa nulis docker-compose.yml untuk microservices stack
- [ ] Paham Kubernetes manifests untuk microservices
- [ ] Paham namespace strategy di K8s
- [ ] Paham service mesh concept (Istio/Linkerd)

### Testing

- [ ] Paham testing pyramid untuk microservices
- [ ] Paham contract testing (Pact)
- [ ] Paham integration testing across services
- [ ] Paham testing strategy per level

### Migration

- [ ] Paham Strangler Fig pattern
- [ ] Bisa plan migration dari monolith (step by step)
- [ ] Tau mulai dari service mana (paling independent)
- [ ] Paham tips dan pitfalls migration

### Anti-Patterns & When NOT to Use

- [ ] Paham 6 common anti-patterns
- [ ] Paham kapan TIDAK pakai microservices
- [ ] Bisa evaluate apakah project butuh microservices atau tidak

### Tools & Ecosystem

- [ ] Tau Go microservices frameworks (Go kit, Kratos, Go Micro)
- [ ] Tau messaging options (gRPC, NATS, Kafka, RabbitMQ)
- [ ] Tau observability stack (OTel, Jaeger, Loki, Prometheus)
- [ ] Tau service mesh options (dan kapan ga perlu)

---

## Refleksi: Monolith vs Microservices

Selamat! Kamu baru aja selesaiin bab terpanjang kedua setelah Kubernetes. Dan ini mungkin bab yang paling **penting untuk decision making** di karier kamu.

```
Key takeaways:
━━━━━━━━━━━━━

1. Microservices is a TRADE-OFF, bukan upgrade
   → Tukar code complexity dengan infrastructure complexity

2. Start with monolith
   → "Almost all successful microservice stories started with a monolith"

3. Extract services ONE AT A TIME
   → Strangler Fig, bukan big bang rewrite

4. Design for failure
   → Network calls fail. Plan for it.

5. Observability is NOT optional
   → Tanpa tracing + logging + monitoring = debugging nightmare

6. Keep it simple
   → 5 well-designed services > 50 nano-services

7. The best architecture is the one your team can maintain
   → Ga peduli seberapa elegant designnya kalau tim ga bisa maintain
```

```
Journey so far:
━━━━━━━━━━━━━━
Bab 1-4:   Go Fundamentals
Bab 5-10:  Backend Development (API, DB, Auth, Cache, Queue, Testing)
Bab 11-15: DevOps Core (Docker, CI/CD, AWS, Nginx, Monitoring)
Bab 16:    Container Orchestration (Kubernetes)
Bab 17:    Infrastructure as Code (Terraform)
Bab 18:    Distributed System Patterns (Microservices) ← KAMU DI SINI

What's next:
━━━━━━━━━━━
Bab 19: Performance — Load Testing & Optimization
Bab 20: Security — Best Practices
```

Dari developer yang cuma tau bikin monolith CRUD, sekarang kamu paham **distributed system patterns** — gRPC, event-driven, saga, circuit breaker, distributed tracing. Ini skill yang dicari di posisi **Senior Backend Engineer** dan **Platform Engineer** di company manapun.

Tapi inget: **paham** microservices bukan berarti kamu **harus** pakai microservices. Justru senior engineer yang beneran itu yang tau **kapan TIDAK pakai** microservices. Yang tau bilang "monolith cukup buat sekarang, kita decompose nanti kalau butuh."

> **Next up: Bab 19 — Performance & Load Testing.** Setelah paham architecture, sekarang kita optimize. Gimana measure performance? Gimana identify bottleneck? Gimana handle 10,000 concurrent users? Load testing, profiling, optimization — semua ada di bab selanjutnya.

---
