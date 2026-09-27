# Bab 15: Monitoring — Prometheus + Grafana dari Nol Sampai Production

## Kenapa Bab Ini Penting?

Selamat. Kamu udah sampai di bab terakhir **Stage 2: DevOps Core**. Selama 4 bab terakhir — Docker, CI/CD, AWS, Nginx — kamu udah berhasil bikin pipeline yang bisa deploy Task Manager API ke production secara otomatis. API kamu jalan di cloud, ada HTTPS, ada reverse proxy, ada rate limiting. Keren.

Tapi ada satu masalah besar: **kamu ga tau apa yang terjadi di production.**

API kamu bisa aja error rate-nya naik 50% selama 3 jam, dan kamu ga tau sampai user komplain di Twitter. Database connection pool bisa aja penuh, response time naik dari 50ms jadi 5 detik, tapi kamu ga tau sampai bos nanya "Kok API-nya lemot ya?". Server bisa aja kehabisan memory, crash tengah malam, dan kamu baru tau pas bangun pagi buka Slack.

**Tanpa monitoring, kamu buta di production.**

```
Sebelum bab ini:
- API udah jalan di production (Docker + AWS + Nginx)
- CI/CD udah auto deploy
- Tapi... ga ada cara tau apakah API healthy
- Ga ada cara tau berapa request per detik
- Ga ada cara tau response time P95
- Ga ada alert kalau server down
- Kalau ada masalah, tau dari user yang komplain

Setelah bab ini:
- Prometheus scrape metrics dari Go app setiap 15 detik
- Grafana dashboard live — request rate, error rate, latency, memory
- Alert otomatis ke Telegram/Discord kalau ada masalah
- Structured logging (JSON) untuk debugging
- Loki untuk log aggregation
- Health check endpoints yang proper
- Full monitoring stack di docker-compose
- Kalau ada masalah, kamu tau SEBELUM user tau
```

Kalau kamu pernah pakai Laravel Telescope atau NestJS dengan PM2 monitoring, itu monitoring **level basic**. Di bab ini kita naik level — monitoring yang dipake perusahaan production beneran: Prometheus + Grafana.

> **Analogi gampang:** Monitoring itu kayak **dashboard mobil**. Speedometer ngasih tau kecepatan (request rate). Fuel gauge ngasih tau bensin sisa berapa (memory usage). RPM gauge ngasih tau mesin kerja seberapa keras (CPU usage). Engine warning light nyala kalau ada masalah (alerts). Bayangin nyetir mobil **tanpa dashboard** — ga tau kecepatan, ga tau bensin sisa berapa, ga tau engine overheat. Serem kan? Itulah production tanpa monitoring.

---

## 15.1 Kenapa Monitoring Penting

### "You Can't Fix What You Can't See"

Ini prinsip fundamental monitoring. Kalau kamu ga bisa **lihat** apa yang terjadi di system kamu, kamu ga bisa **fix** masalahnya. Simple as that.

```
Tanpa Monitoring:
━━━━━━━━━━━━━━━
1. User komplain: "API lemot"
2. Kamu: "Hmm, coba gue SSH ke server"
3. SSH ke server, cek manually
4. Coba curl endpoint, kayaknya normal
5. "Mungkin internet user-nya aja kali"
6. 2 jam kemudian, makin banyak yang komplain
7. Ternyata database connection pool penuh sejak tadi pagi
8. Fix: restart app
9. Total downtime: 5 jam
10. Reputasi: rusak

Dengan Monitoring:
━━━━━━━━━━━━━━━━━
1. Alert Telegram: "P95 latency > 500ms sejak 08:15"
2. Buka Grafana dashboard
3. Lihat: DB connection pool 100% utilized
4. Lihat: query time spike mulai 08:10
5. Cek logs: ada slow query yang baru di-deploy semalam
6. Fix: rollback atau optimize query
7. Total downtime: 15 menit
8. Reputasi: "Wah DevOps-nya responsive banget"
```

### Horror Stories Tanpa Monitoring

Ini bukan cerita fiktif — ini kejadian nyata di industri:

```
Horror Story #1: "Server Down 3 Jam, Nobody Knew"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Startup e-commerce. Server API mati jam 2 pagi (OOM kill).
Ga ada monitoring, ga ada alert.
Team baru tau jam 8 pagi pas masuk kantor.
6 jam downtime. Kehilangan transaksi weekend.
Estimated loss: Rp 50 juta+

Horror Story #2: "Memory Leak yang Slow"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Go service punya goroutine leak.
Memory naik 10MB per jam. Ga terasa.
Setelah 3 hari, server pakai 8GB RAM (dari 4GB allocated).
OOM killer bunuh process. 502 Bad Gateway.
Kalau ada monitoring — memory graph naik terus,
alert trigger di hari pertama.

Horror Story #3: "Disk Full, Database Crash"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PostgreSQL log file ga di-rotate.
Disk usage naik pelan-pelan: 60%... 80%... 95%... 100%.
Database ga bisa write. Semua INSERT/UPDATE gagal.
API return 500 Internal Server Error.
Fix: hapus old logs, setup logrotate.
Kalau ada monitoring — disk usage alert di 80%.
```

### Kenapa Developer Perlu Tau Monitoring?

"Tapi kan monitoring itu kerjaannya SRE/DevOps?"

Betul, di perusahaan besar ada team SRE khusus. Tapi:

1. **Di startup, kamu yang handle semuanya** — ga ada SRE dedicated
2. **Kamu yang nulis kode** — kamu yang paling tau metric apa yang penting
3. **On-call rotation** — banyak perusahaan yang developer ikut on-call
4. **Career growth** — developer yang paham monitoring = senior engineer

```
Junior Developer:
"API-nya error? Hmm, coba restart aja"

Senior Developer:
"Error rate naik 3% sejak deploy terakhir.
P95 latency naik dari 50ms ke 200ms.
Root cause: query N+1 di endpoint GET /tasks.
Fix: eager loading, deploy hotfix,
verify error rate balik ke baseline."
```

Bedanya? **Data.** Senior engineer pakai data, bukan feeling.

---

## 15.2 Observability Stack — Three Pillars

### Apa Itu Observability?

Observability = kemampuan untuk **memahami** apa yang terjadi di dalam system dari **output** yang dihasilkan.

Beda sama monitoring:
- **Monitoring** = tau *bahwa* ada masalah ("server down!")
- **Observability** = tau *kenapa* ada masalah ("server down karena memory leak di goroutine X yang dipanggil endpoint Y")

### Three Pillars of Observability

```
┌──────────────────────────────────────────────────────────┐
│                THREE PILLARS OF OBSERVABILITY              │
│                                                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   METRICS    │  │    LOGS      │  │   TRACES     │    │
│  │              │  │              │  │              │    │
│  │  Angka       │  │  Event       │  │  Journey     │    │
│  │  (numbers)   │  │  (text)      │  │  (path)      │    │
│  │              │  │              │  │              │    │
│  │  "Berapa?"   │  │  "Apa yang   │  │  "Lewat      │    │
│  │              │  │   terjadi?"  │  │   mana?"     │    │
│  │              │  │              │  │              │    │
│  │  CPU: 75%    │  │  "User login │  │  Request →   │    │
│  │  RPS: 1200   │  │   failed:    │  │  Auth →      │    │
│  │  Latency:    │  │   invalid    │  │  DB Query →  │    │
│  │  50ms        │  │   password"  │  │  Cache →     │    │
│  │              │  │              │  │  Response    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                            │
│  Tools:            Tools:            Tools:               │
│  Prometheus        ELK Stack         Jaeger               │
│  Datadog           Loki              Zipkin               │
│  CloudWatch        Fluentd           Tempo                │
│  InfluxDB          Papertrail        OpenTelemetry        │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

### Pillar 1: Metrics — "Berapa?"

Metrics itu **angka** yang menggambarkan kondisi system pada waktu tertentu. Metrics itu **aggregated** — bukan per-event, tapi summary.

```
Contoh metrics:
- Total HTTP requests: 1,234,567
- Request per second: 450 RPS
- Error rate: 2.3%
- P95 latency: 45ms
- Active goroutines: 127
- Memory usage: 256MB
- CPU usage: 35%
- Database connections: 15/20
```

**Kenapa penting:** Metrics ngasih kamu **big picture**. Kamu bisa lihat trend, compare timeframe, set alert threshold.

### Pillar 2: Logs — "Apa yang Terjadi?"

Logs itu **event** yang terjadi di aplikasi. Setiap log entry = satu kejadian spesifik.

```
Contoh logs:
[2026-05-01 10:15:33] INFO  User login successful: user_id=42, email=john@example.com
[2026-05-01 10:15:34] ERROR Failed to create task: board_id=5, error="board not found"
[2026-05-01 10:15:35] WARN  Rate limit exceeded: ip=203.0.113.50, limit=100/min
[2026-05-01 10:15:36] INFO  Task created: task_id=789, board_id=3, user_id=42
```

**Kenapa penting:** Logs ngasih kamu **detail**. Kalau metrics bilang "error rate naik", logs bilang "error-nya apa, di mana, kapan, user siapa".

### Pillar 3: Traces — "Lewat Mana?"

Traces itu **perjalanan satu request** melintasi seluruh system. Penting banget di microservices.

```
Trace ID: abc-123-def
├── API Gateway (2ms)
├── Auth Service - validate token (5ms)
├── Task Service - create task (45ms)
│   ├── PostgreSQL - INSERT task (12ms)
│   ├── Redis - invalidate cache (3ms)
│   └── RabbitMQ - publish event (8ms)
└── Total: 75ms
```

**Kenapa penting:** Traces ngasih kamu **path**. Kamu bisa lihat bottleneck ada di mana — DB query lambat? External API timeout?

### Focus Bab Ini

Di bab ini, kita fokus ke **Metrics** (Prometheus + Grafana) dan **Logs** (structured logging + Loki). Traces (Jaeger/Tempo) kita bahas nanti di bab microservices — karena traces paling berguna kalau udah multiple services.

```
Bab ini:
├── Metrics  → Prometheus + Grafana     ← FOKUS UTAMA
├── Logs     → Structured Logging + Loki ← FOKUS KEDUA
└── Traces   → (Bab 18: Microservices)   ← NANTI
```

---

## 15.3 Prometheus Overview

### Apa Itu Prometheus?

Prometheus itu **open-source monitoring system** dan **time-series database** yang dibuat oleh SoundCloud tahun 2012. Sekarang jadi project **CNCF graduated** (sama levelnya dengan Kubernetes).

```
Prometheus Facts:
━━━━━━━━━━━━━━━━
- Dibuat: 2012 di SoundCloud
- Open source: 2015
- CNCF graduated: 2018 (setelah Kubernetes)
- Bahasa: Go (tentu saja!)
- License: Apache 2.0
- Dipakai: SoundCloud, DigitalOcean, Docker, Uber, GitLab, dll
```

### Pull-Based Model (Scrape)

Yang bikin Prometheus beda dari tools monitoring lain: Prometheus **pull** (scrape) metrics dari target, bukan target yang **push** metrics ke Prometheus.

```
Push-Based (Datadog, StatsD):
━━━━━━━━━━━━━━━━━━━━━━━━━━━
App actively sends metrics to monitoring server.
App harus tau alamat monitoring server.

App A ─── push ───→ Monitoring Server
App B ─── push ───→ Monitoring Server
App C ─── push ───→ Monitoring Server

Pull-Based (Prometheus):
━━━━━━━━━━━━━━━━━━━━━━━
Prometheus yang actively fetch metrics dari apps.
Apps cuma perlu expose /metrics endpoint.

Prometheus ─── scrape ───→ App A /metrics
Prometheus ─── scrape ───→ App B /metrics
Prometheus ─── scrape ───→ App C /metrics
```

**Keuntungan pull-based:**
1. App ga perlu tau Prometheus ada di mana
2. Prometheus bisa detect kalau target down (scrape gagal)
3. Gampang debug — kamu bisa buka `/metrics` di browser
4. Satu Prometheus bisa scrape banyak target

### Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Prometheus Architecture                     │
│                                                               │
│  ┌──────────┐     scrape      ┌──────────────────┐           │
│  │  Go App   │ ←─────────── │    Prometheus     │           │
│  │ /metrics  │               │                   │           │
│  └──────────┘               │  ┌──────────────┐ │           │
│                              │  │  Time-Series │ │           │
│  ┌──────────┐     scrape    │  │   Database   │ │           │
│  │  Node     │ ←─────────── │  └──────────────┘ │           │
│  │ Exporter  │               │                   │           │
│  └──────────┘               │  ┌──────────────┐ │           │
│                              │  │   PromQL     │ │           │
│  ┌──────────┐     scrape    │  │   Engine     │ │           │
│  │ PostgreSQL│ ←─────────── │  └──────────────┘ │           │
│  │ Exporter  │               │                   │           │
│  └──────────┘               │  ┌──────────────┐ │           │
│                              │  │Alert Manager │ │           │
│                              │  └──────┬───────┘ │           │
│                              └─────────┼─────────┘           │
│                                        │                      │
│              ┌─────────────────────────┼──────────┐          │
│              │                         │          │          │
│              ▼                         ▼          ▼          │
│  ┌──────────────────┐    ┌──────────┐  ┌────────────┐       │
│  │     Grafana       │    │ Telegram │  │  Discord   │       │
│  │   (Visualization) │    │   Bot    │  │  Webhook   │       │
│  └──────────────────┘    └──────────┘  └────────────┘       │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

### Prometheus vs SaaS Alternatives

| Aspek | Prometheus + Grafana | Datadog | New Relic |
|-------|---------------------|---------|-----------|
| **Type** | Self-hosted (gratis) | SaaS (bayar) | SaaS (bayar) |
| **Cost** | $0 (infra doang) | $15-23/host/bulan | $25+/host/bulan |
| **Setup** | Manual (Docker) | Install agent, done | Install agent, done |
| **Customization** | Sangat fleksibel | Terbatas UI mereka | Terbatas UI mereka |
| **Data ownership** | Data di server kamu | Data di cloud mereka | Data di cloud mereka |
| **Learning curve** | Medium (PromQL) | Low | Low |
| **Community** | Huge (CNCF) | Good | Good |
| **Best for** | Tim yang mau kontrol penuh | Tim yang mau cepat setup | Enterprise |

```
Kapan Pakai Prometheus + Grafana:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Budget terbatas (startup, project pribadi)
- Mau kontrol penuh atas data
- Mau belajar monitoring beneran
- Udah pakai Kubernetes (Prometheus = standard)
- Data sensitive (ga mau di cloud orang lain)

Kapan Pakai Datadog/New Relic:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Budget banyak
- Ga mau maintain infrastructure monitoring
- Tim kecil, butuh cepat setup
- Butuh APM (Application Performance Monitoring) out-of-the-box
```

> **Fun fact:** Prometheus, Grafana, Docker, Kubernetes — semua ditulis pakai **Go**. Pattern-nya jelas: tools DevOps modern = Go. Makanya kamu belajar Go + DevOps itu combo yang pas banget.

---

## 15.4 Metric Types

Prometheus punya **4 tipe metric**. Kamu harus paham kapan pakai yang mana.

### 1. Counter — Naik Terus, Ga Pernah Turun

Counter itu angka yang **cuma naik** (atau reset ke 0 kalau restart). Cocok untuk hal-hal yang kamu **hitung**.

```
Counter = Odometer mobil
━━━━━━━━━━━━━━━━━━━━━━
Odometer cuma naik. Ga bisa turun.
Dari 10,000 km → 10,001 km → 10,002 km → ...
Reset cuma kalau mobil baru (restart).

Contoh Counter:
- http_requests_total: total request yang masuk (1, 2, 3, ..., 1000000)
- http_errors_total: total error (1, 2, 3, ..., 500)
- tasks_created_total: total task yang dibuat
- emails_sent_total: total email yang dikirim
- db_queries_total: total query ke database

JANGAN pakai counter untuk:
- CPU usage (bisa naik/turun)
- Active connections (bisa naik/turun)
- Temperature (bisa naik/turun)
```

### 2. Gauge — Bisa Naik, Bisa Turun

Gauge itu angka yang **bisa naik dan turun**. Cocok untuk hal-hal yang menunjukkan **kondisi saat ini**.

```
Gauge = Speedometer mobil
━━━━━━━━━━━━━━━━━━━━━━━
Speedometer bisa naik dan turun.
0 km/h → 60 km/h → 100 km/h → 80 km/h → 0 km/h

Contoh Gauge:
- active_connections: koneksi aktif saat ini (naik/turun)
- cpu_usage_percent: CPU usage saat ini (0-100%)
- memory_usage_bytes: memory yang dipakai saat ini
- goroutines_count: jumlah goroutine aktif
- queue_size: jumlah item di queue saat ini
- temperature_celsius: suhu server

JANGAN pakai gauge untuk:
- Total requests (harusnya counter — cuma naik)
- Total errors (harusnya counter)
```

### 3. Histogram — Distribusi Nilai

Histogram **mengukur distribusi** dari suatu nilai — biasanya **durasi** atau **ukuran**. Histogram bisa kasih kamu percentile (P50, P90, P95, P99).

```
Histogram = Catatan lap time di balapan
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Kamu catat setiap lap time:
Lap 1: 1:45, Lap 2: 1:50, Lap 3: 1:42, Lap 4: 2:01, ...

Dari data itu, kamu bisa hitung:
- Rata-rata lap time
- Lap time tercepat
- 95% lap selesai dalam waktu berapa (P95)

Contoh Histogram:
- http_request_duration_seconds: berapa lama request diproses
  → Bucket: 0.005s, 0.01s, 0.025s, 0.05s, 0.1s, 0.25s, 0.5s, 1s, 5s
  → P95: 95% request selesai dalam 45ms

- http_response_size_bytes: seberapa besar response
  → P50: 50% response ukurannya < 2KB

- db_query_duration_seconds: berapa lama query ke database
```

Histogram otomatis bikin **3 metrics**:
- `_bucket`: jumlah observasi per bucket (le="0.1", le="0.5", dll)
- `_sum`: total sum dari semua observasi
- `_count`: total jumlah observasi

### 4. Summary — Mirip Histogram, Tapi Hitung Quantile di Client

Summary mirip histogram, tapi **quantile dihitung di sisi aplikasi** (bukan di Prometheus).

```
Summary vs Histogram:
━━━━━━━━━━━━━━━━━━━━

Histogram:
- Bucket di-define di awal (0.01s, 0.05s, 0.1s, ...)
- Quantile dihitung di Prometheus pakai PromQL
- Bisa di-aggregate across instances
- Lebih fleksibel — RECOMMENDED

Summary:
- Quantile dihitung di aplikasi (phi=0.5, phi=0.95, phi=0.99)
- Lebih akurat untuk specific quantile
- TIDAK bisa di-aggregate across instances
- Cocok kalau kamu tau pasti quantile apa yang kamu mau
```

### Kapan Pakai Apa?

```
┌──────────────────────────────────────────────────────┐
│              Metric Type Cheat Sheet                  │
│                                                       │
│  "Berapa TOTAL nya?"          → COUNTER               │
│    total requests, total errors, total tasks          │
│                                                       │
│  "Berapa SEKARANG?"           → GAUGE                 │
│    active users, CPU%, memory, goroutines             │
│                                                       │
│  "Berapa LAMA / BESAR?"      → HISTOGRAM              │
│    request duration, response size, query time        │
│                                                       │
│  "Berapa PERCENTILE (exact)?" → SUMMARY               │
│    (jarang dipakai, biasanya histogram cukup)         │
│                                                       │
└──────────────────────────────────────────────────────┘
```

---

## 15.5 Instrumenting Go App

### Install prometheus/client_golang

Library official Prometheus untuk Go. Ini standard-nya — semua Go app yang pakai Prometheus pake library ini.

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/prometheus/client_golang/prometheus/promauto
```

### Setup /metrics Endpoint — Basic

Pertama, kita expose `/metrics` endpoint di Go app. Ini endpoint yang bakal di-scrape sama Prometheus.

```go
// internal/handler/metrics.go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler wraps promhttp handler for Gin
func MetricsHandler() gin.HandlerFunc {
    h := promhttp.Handler()
    return func(c *gin.Context) {
        h.ServeHTTP(c.Writer, c.Request)
    }
}
```

Register di router:

```go
// internal/router/router.go
func Setup(r *gin.Engine, h *Handlers, authService *service.AuthService, redisClient *redis.Client) {
    // Health check (public)
    r.GET("/health", handler.HealthCheck)

    // Prometheus metrics endpoint (public, tapi biasanya di-protect di Nginx)
    r.GET("/metrics", handler.MetricsHandler())

    // ... rest of routes
}
```

### Apa yang Udah Ke-expose?

Tanpa nulis custom metric apapun, Prometheus client_golang **otomatis** expose default Go metrics:

```bash
# Buka browser: http://localhost:8080/metrics

# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 12

# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 3.456e+06

# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 2.34e-05
go_gc_duration_seconds{quantile="0.25"} 3.45e-05
go_gc_duration_seconds{quantile="0.5"} 4.56e-05
go_gc_duration_seconds{quantile="0.75"} 5.67e-05
go_gc_duration_seconds{quantile="1"} 0.000123

# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge
go_threads 8

# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 0.45

# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 2.5e+07
```

Ini udah lumayan — kamu bisa monitor goroutines, memory, GC, CPU. Tapi kita perlu **custom metrics** yang spesifik untuk Task Manager API.

---

## 15.6 Custom Metrics untuk Task Manager

### Define Custom Metrics

Bikin file khusus untuk define semua custom metrics:

```go
// internal/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP Metrics
    // Counter: total HTTP requests, labeled by method, path, status code
    HttpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    // Histogram: HTTP request duration in seconds, labeled by method, path
    HttpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
        },
        []string{"method", "path"},
    )

    // Histogram: HTTP response size in bytes
    HttpResponseSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_response_size_bytes",
            Help:    "HTTP response size in bytes",
            Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000},
        },
        []string{"method", "path"},
    )

    // Business Metrics
    // Counter: total tasks created
    TasksCreatedTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "tasks_created_total",
            Help: "Total number of tasks created",
        },
    )

    // Counter: total tasks completed
    TasksCompletedTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "tasks_completed_total",
            Help: "Total number of tasks completed",
        },
    )

    // Counter: total user registrations
    UserRegistrationsTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "user_registrations_total",
            Help: "Total number of user registrations",
        },
    )

    // Counter: total login attempts (success/fail)
    LoginAttemptsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "login_attempts_total",
            Help: "Total number of login attempts",
        },
        []string{"status"}, // "success" or "failed"
    )

    // Gauge: active users (currently authenticated sessions)
    ActiveUsersGauge = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_users",
            Help: "Number of currently active users",
        },
    )

    // Gauge: database connection pool stats
    DBConnectionsActive = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_active",
            Help: "Number of active database connections",
        },
    )

    DBConnectionsIdle = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_idle",
            Help: "Number of idle database connections",
        },
    )

    // Histogram: database query duration
    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 5},
        },
        []string{"operation", "table"},
    )
)
```

### Prometheus Middleware — Auto Collect per Request

Ini bagian paling penting. Bikin middleware yang **otomatis** record HTTP metrics untuk setiap request:

```go
// internal/middleware/prometheus.go
package middleware

import (
    "fmt"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/internal/metrics"
)

// PrometheusMiddleware records HTTP metrics for every request
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip /metrics dan /health endpoint dari recording
        // (biar ga polusi metrics)
        path := c.FullPath()
        if path == "/metrics" || path == "/health" || path == "/ready" {
            c.Next()
            return
        }

        // Kalau FullPath kosong (404), pakai "not_found"
        if path == "" {
            path = "not_found"
        }

        start := time.Now()

        // Process request
        c.Next()

        // After request
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        method := c.Request.Method
        responseSize := float64(c.Writer.Size())

        // Record metrics
        metrics.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
        metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(duration)

        if responseSize > 0 {
            metrics.HttpResponseSize.WithLabelValues(method, path).Observe(responseSize)
        }
    }
}

// NormalizePath normalizes path parameters to reduce label cardinality
// /api/v1/tasks/123 → /api/v1/tasks/:id
// Gin's FullPath() already handles this, tapi ini backup kalau perlu
func NormalizePath(path string) string {
    // Gin's c.FullPath() returns the route pattern, not the actual path
    // So /api/v1/tasks/123 becomes /api/v1/tasks/:id automatically
    // This function is here as documentation
    return path
}
```

**Penting banget:** Kita pakai `c.FullPath()` bukan `c.Request.URL.Path`. Kenapa?

```
c.Request.URL.Path:
- /api/v1/tasks/1
- /api/v1/tasks/2
- /api/v1/tasks/3
- /api/v1/tasks/42
- ... (infinite cardinality!)

c.FullPath():
- /api/v1/tasks/:id
- /api/v1/tasks/:id
- /api/v1/tasks/:id
- /api/v1/tasks/:id
- ... (1 label value)

Kenapa ini penting?
Label cardinality tinggi = Prometheus pakai banyak memory.
1 juta unique label values = Prometheus bisa OOM.
SELALU normalize path parameters!
```

### Register Middleware di Router

```go
// internal/router/router.go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/internal/handler"
    "github.com/irzan-aldi/task-manager/internal/middleware"
    "github.com/irzan-aldi/task-manager/internal/service"
    "github.com/redis/go-redis/v9"
    "time"
)

func Setup(r *gin.Engine, h *Handlers, authService *service.AuthService, redisClient *redis.Client) {
    // Prometheus middleware — HARUS PERTAMA (sebelum middleware lain)
    r.Use(middleware.PrometheusMiddleware())

    // Health check endpoints (public)
    r.GET("/health", handler.HealthCheck)
    r.GET("/ready", handler.ReadinessCheck)  // tambah readiness check
    r.GET("/metrics", handler.MetricsHandler())

    // Rate limiter: 100 requests per minute per IP
    r.Use(middleware.RateLimiter(redisClient, 100, time.Minute))

    // ... rest of routes (sama kayak sebelumnya)
}
```

### Instrument Business Logic

Selain auto-collect di middleware, kamu juga perlu **manually instrument** business events:

```go
// internal/handler/task.go — tambah metrics increment
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/irzan-aldi/task-manager/internal/metrics"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/internal/service"
    "github.com/irzan-aldi/task-manager/pkg/response"
)

func (h *TaskHandler) Create(c *gin.Context) {
    userID := c.GetUint("userID")
    boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid board ID")
        return
    }

    var req model.CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    task, err := h.taskService.Create(uint(boardID), userID, req)
    if err != nil {
        response.Error(c, http.StatusForbidden, err.Error())
        return
    }

    // Increment task created counter
    metrics.TasksCreatedTotal.Inc()

    response.Success(c, http.StatusCreated, "Task created", task)
}
```

```go
// internal/handler/auth.go — tambah login metrics
func (h *AuthHandler) Login(c *gin.Context) {
    var req model.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    token, err := h.authService.Login(req.Email, req.Password)
    if err != nil {
        // Record failed login
        metrics.LoginAttemptsTotal.WithLabelValues("failed").Inc()
        response.Error(c, http.StatusUnauthorized, "Invalid credentials")
        return
    }

    // Record successful login
    metrics.LoginAttemptsTotal.WithLabelValues("success").Inc()
    metrics.ActiveUsersGauge.Inc()

    response.Success(c, http.StatusOK, "Login successful", token)
}
```

### Database Connection Pool Metrics

Bonus: expose database connection pool stats sebagai Prometheus metrics. Ini penting banget buat detect connection leaks.

```go
// internal/metrics/db_collector.go
package metrics

import (
    "database/sql"
    "time"
)

// StartDBMetricsCollector periodically collects DB pool stats
func StartDBMetricsCollector(sqlDB *sql.DB, interval time.Duration) {
    go func() {
        for {
            stats := sqlDB.Stats()
            DBConnectionsActive.Set(float64(stats.InUse))
            DBConnectionsIdle.Set(float64(stats.Idle))
            time.Sleep(interval)
        }
    }()
}
```

```go
// cmd/api/main.go — start DB metrics collector
func main() {
    // ... existing setup ...

    // Start DB metrics collector
    sqlDB, _ := db.DB()
    metrics.StartDBMetricsCollector(sqlDB, 15*time.Second)

    // ... rest of main
}
```

### Verify Metrics

Jalanin app dan buka `/metrics`:

```bash
go run cmd/api/main.go

# Di terminal lain:
curl http://localhost:8080/metrics | grep http_requests_total

# Output:
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/v1/tasks/my",status="200"} 15
http_requests_total{method="POST",path="/api/v1/auth/login",status="200"} 3
http_requests_total{method="POST",path="/api/v1/auth/login",status="401"} 1
http_requests_total{method="POST",path="/api/v1/boards/:boardId/tasks",status="201"} 7

# Cek request duration
curl http://localhost:8080/metrics | grep http_request_duration

# Output:
# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/v1/tasks/my",le="0.001"} 2
http_request_duration_seconds_bucket{method="GET",path="/api/v1/tasks/my",le="0.005"} 8
http_request_duration_seconds_bucket{method="GET",path="/api/v1/tasks/my",le="0.01"} 13
http_request_duration_seconds_bucket{method="GET",path="/api/v1/tasks/my",le="0.025"} 15
http_request_duration_seconds_bucket{method="GET",path="/api/v1/tasks/my",le="+Inf"} 15
http_request_duration_seconds_sum{method="GET",path="/api/v1/tasks/my"} 0.0543
http_request_duration_seconds_count{method="GET",path="/api/v1/tasks/my"} 15
```

Kalau kamu lihat output kayak gitu, **metrics udah jalan**. Sekarang tinggal setup Prometheus untuk scrape-nya.

---

## 15.7 Setup Prometheus

### Prometheus Config — prometheus.yml

Prometheus butuh config file yang kasih tau dia **target apa aja** yang harus di-scrape, dan **seberapa sering**.

```yaml
# monitoring/prometheus/prometheus.yml

# Global config
global:
  scrape_interval: 15s       # Scrape setiap 15 detik (default)
  evaluation_interval: 15s   # Evaluate alert rules setiap 15 detik
  scrape_timeout: 10s        # Timeout per scrape

# Alertmanager config (nanti kita setup)
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          # - alertmanager:9093

# Alert rules files
rule_files:
  - "alerts/*.yml"

# Scrape configurations — target apa aja yang di-scrape
scrape_configs:
  # Prometheus monitor dirinya sendiri
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]

  # Go Task Manager API
  - job_name: "task-manager-api"
    static_configs:
      - targets: ["api:8080"]    # "api" = Docker service name
    metrics_path: /metrics       # Default, tapi explicit lebih jelas
    scrape_interval: 10s         # Override global, scrape lebih sering

  # Node Exporter (system metrics — CPU, memory, disk, network)
  - job_name: "node-exporter"
    static_configs:
      - targets: ["node-exporter:9100"]

  # PostgreSQL Exporter
  - job_name: "postgres-exporter"
    static_configs:
      - targets: ["postgres-exporter:9187"]

  # Redis Exporter
  - job_name: "redis-exporter"
    static_configs:
      - targets: ["redis-exporter:9121"]

  # Nginx Exporter (kalau pakai nginx-prometheus-exporter)
  - job_name: "nginx-exporter"
    static_configs:
      - targets: ["nginx-exporter:9113"]
```

### Jalanin Prometheus via Docker

```bash
# Quick test — jalanin Prometheus standalone
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus:v3.4.0

# Buka browser: http://localhost:9090
# Masuk ke Status > Targets — lihat target yang di-scrape
```

### Verify Prometheus

Buka `http://localhost:9090` di browser:

```
1. Cek Targets: Status > Targets
   - Semua target harus "UP" (hijau)
   - Kalau "DOWN" (merah), cek network/firewall

2. Coba Query: masukkan di query box
   - up                    → lihat target mana yang UP/DOWN
   - go_goroutines         → jumlah goroutine di Go app
   - http_requests_total   → total HTTP requests

3. Cek Graph: klik tab "Graph"
   - Pilih range: 15m, 1h, 6h
   - Lihat grafik metrics over time
```

---

## 15.8 PromQL Basics

PromQL (Prometheus Query Language) itu query language untuk **query metrics** dari Prometheus. Ini yang bikin Prometheus powerful — kamu bisa slice and dice data sesuka hati.

### Instant Vector — Nilai Saat Ini

```promql
# Semua metrics dengan nama "http_requests_total"
http_requests_total

# Filter by label
http_requests_total{method="GET"}
http_requests_total{status="500"}
http_requests_total{method="POST", path="/api/v1/auth/login"}

# Regex match
http_requests_total{path=~"/api/v1/tasks.*"}
http_requests_total{status=~"5.."}   # 500, 501, 502, 503, ...
http_requests_total{method!="GET"}   # semua kecuali GET
```

### Range Vector — Nilai dalam Rentang Waktu

```promql
# HTTP requests dalam 5 menit terakhir
http_requests_total[5m]

# HTTP requests dalam 1 jam terakhir
http_requests_total[1h]

# Range vector ga bisa langsung di-graph
# Harus di-aggregate dulu (pakai rate, avg, dll)
```

### rate() — Request per Second

**Ini function paling penting di PromQL.** `rate()` menghitung per-second rate dari counter.

```promql
# Request per second (dalam 5 menit terakhir)
rate(http_requests_total[5m])

# Request per second, filtered by status 500
rate(http_requests_total{status="500"}[5m])

# Total request rate across all endpoints
sum(rate(http_requests_total[5m]))

# Request rate per endpoint
sum by (path) (rate(http_requests_total[5m]))

# Request rate per method
sum by (method) (rate(http_requests_total[5m]))
```

```
Kenapa rate() bukan langsung http_requests_total?
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

http_requests_total = 1,234,567
Itu ga berguna. "1 juta request" — sejak kapan? Sejak app start.

rate(http_requests_total[5m]) = 45.3
Ini berguna! "45.3 request per detik dalam 5 menit terakhir"
```

### Error Rate Percentage

```promql
# Error rate = (5xx requests / total requests) * 100
sum(rate(http_requests_total{status=~"5.."}[5m]))
/
sum(rate(http_requests_total[5m]))
* 100

# Hasil: 2.3 (artinya 2.3% error rate)
```

### Latency Percentiles (P50, P95, P99)

```promql
# P50 (median) — 50% request selesai dalam waktu segini
histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))

# P95 — 95% request selesai dalam waktu segini (SLO standard)
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))

# P99 — 99% request selesai dalam waktu segini
histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))

# P95 per endpoint
histogram_quantile(0.95,
  sum by (le, path) (rate(http_request_duration_seconds_bucket[5m]))
)
```

```
Kenapa P95, bukan average?
━━━━━━━━━━━━━━━━━━━━━━━━

Average itu BOHONG.

Contoh: 99 request selesai dalam 10ms, 1 request selesai dalam 10 detik.
Average = (99 * 10ms + 1 * 10000ms) / 100 = 109ms

Kelihatan okay — "rata-rata 109ms".
Tapi 1% user nunggu 10 DETIK. User itu pasti marah.

P95 = 10ms (95% request cepet)
P99 = 10000ms (1% request super lambat — ADA MASALAH!)

P95/P99 ngasih tau kamu WORST CASE experience.
Itu yang penting.
```

### Practical PromQL Queries

```promql
# ===== REQUEST METRICS =====

# Total request per second (semua endpoint)
sum(rate(http_requests_total[5m]))

# Request per second per endpoint (top 5)
topk(5, sum by (path) (rate(http_requests_total[5m])))

# Request per second per status code
sum by (status) (rate(http_requests_total[5m]))


# ===== ERROR METRICS =====

# Error rate percentage (5xx)
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# 4xx rate (client errors)
sum(rate(http_requests_total{status=~"4.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# Failed login rate
rate(login_attempts_total{status="failed"}[5m])


# ===== LATENCY METRICS =====

# Average request duration
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

# P95 latency (overall)
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))

# P95 latency per endpoint
histogram_quantile(0.95, sum by (le, path) (rate(http_request_duration_seconds_bucket[5m])))

# Slowest endpoints (by P95)
topk(5, histogram_quantile(0.95, sum by (le, path) (rate(http_request_duration_seconds_bucket[5m]))))


# ===== SYSTEM METRICS =====

# Goroutines
go_goroutines

# Memory usage (allocated)
go_memstats_alloc_bytes / 1024 / 1024    # in MB

# Memory usage trend (is it growing? → possible leak)
deriv(go_memstats_alloc_bytes[1h])

# GC pause duration P99
histogram_quantile(0.99, rate(go_gc_duration_seconds_bucket[5m]))

# CPU usage
rate(process_cpu_seconds_total[5m]) * 100


# ===== BUSINESS METRICS =====

# Tasks created per minute
rate(tasks_created_total[5m]) * 60

# User registrations per hour
increase(user_registrations_total[1h])

# Active users
active_users


# ===== DATABASE METRICS =====

# Active DB connections
db_connections_active

# DB connection utilization
db_connections_active / (db_connections_active + db_connections_idle) * 100

# DB query duration P95
histogram_quantile(0.95, sum by (le) (rate(db_query_duration_seconds_bucket[5m])))
```

---

## 15.9 Grafana Overview

### Apa Itu Grafana?

Grafana itu **open-source visualization platform**. Grafana ga nyimpen data sendiri — dia **query** data dari data source (Prometheus, MySQL, PostgreSQL, Elasticsearch, Loki, dll) dan **visualisasikan** dalam dashboard yang cantik.

```
Grafana Facts:
━━━━━━━━━━━━━
- Dibuat: 2014 oleh Torkel Odegaard
- Bahasa: Go (backend) + TypeScript (frontend)
- License: AGPLv3 (open source)
- Dipakai: hampir semua perusahaan tech
- Tipe: visualization, dashboarding, alerting
- Data sources: 100+ (Prometheus, MySQL, PostgreSQL, CloudWatch, dll)
```

### Grafana Concepts

```
┌──────────────────────────────────────────────────────┐
│                 Grafana Hierarchy                     │
│                                                       │
│  Organization                                         │
│  └── Dashboard                                        │
│      ├── Row (pengelompokan)                         │
│      │   ├── Panel (graph, stat, gauge, table, ...)  │
│      │   ├── Panel                                    │
│      │   └── Panel                                    │
│      ├── Row                                          │
│      │   ├── Panel                                    │
│      │   └── Panel                                    │
│      └── Variables (dynamic filters)                  │
│                                                       │
│  Data Source (Prometheus, Loki, ...)                  │
│  └── Queries (PromQL, LogQL, ...)                    │
│      └── Transformations                              │
│          └── Visualizations                           │
└──────────────────────────────────────────────────────┘
```

---

## 15.10 Setup Grafana

### Jalanin Grafana via Docker

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  -e GF_SECURITY_ADMIN_USER=admin \
  -e GF_SECURITY_ADMIN_PASSWORD=admin123 \
  grafana/grafana:11.6.0

# Buka browser: http://localhost:3000
# Login: admin / admin123
```

### Connect Prometheus Data Source

Langkah-langkah di Grafana UI:

```
1. Login ke Grafana (http://localhost:3000)
2. Sidebar kiri → Connections → Data Sources
3. Klik "Add data source"
4. Pilih "Prometheus"
5. Settings:
   - Name: Prometheus
   - URL: http://prometheus:9090  (kalau Docker network sama)
   - Access: Server (default)
6. Scroll bawah → klik "Save & Test"
7. Harus muncul: "Successfully queried the Prometheus API"
```

### Import Pre-built Dashboards

Grafana punya ribuan dashboard community yang siap pakai:

```
Popular Dashboard IDs (import di Grafana):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Go Metrics Dashboard: 10826
   - Goroutines, memory, GC, CPU
   - Langsung works dengan default Go metrics

2. Node Exporter Full: 1860
   - CPU, memory, disk, network (system level)
   - Butuh node-exporter running

3. PostgreSQL Database: 9628
   - Connections, transactions, query stats
   - Butuh postgres-exporter running

4. Redis Dashboard: 763
   - Memory, connections, commands, keys
   - Butuh redis-exporter running

Cara import:
1. Dashboards → New → Import
2. Masukkan dashboard ID (misal: 10826)
3. Klik "Load"
4. Pilih data source: Prometheus
5. Klik "Import"
```

---

## 15.11 Building Custom Dashboard — Task Manager API

Sekarang bagian yang seru. Kita bikin **custom dashboard** khusus untuk Task Manager API.

### Panel 1: Request Rate (Time Series)

```json
{
  "title": "Request Rate (req/s)",
  "type": "timeseries",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "sum(rate(http_requests_total[5m]))",
      "legendFormat": "Total RPS"
    },
    {
      "expr": "sum by (method) (rate(http_requests_total[5m]))",
      "legendFormat": "{{method}}"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "reqps",
      "custom": {
        "lineWidth": 2,
        "fillOpacity": 10
      }
    }
  }
}
```

### Panel 2: Error Rate (Stat)

```json
{
  "title": "Error Rate",
  "type": "stat",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m])) * 100",
      "legendFormat": "Error Rate %"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "percent",
      "thresholds": {
        "mode": "absolute",
        "steps": [
          { "color": "green", "value": null },
          { "color": "yellow", "value": 1 },
          { "color": "orange", "value": 3 },
          { "color": "red", "value": 5 }
        ]
      }
    }
  }
}
```

### Panel 3: Response Time P95 (Gauge)

```json
{
  "title": "Response Time P95",
  "type": "gauge",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))",
      "legendFormat": "P95"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "s",
      "min": 0,
      "max": 2,
      "thresholds": {
        "mode": "absolute",
        "steps": [
          { "color": "green", "value": null },
          { "color": "yellow", "value": 0.2 },
          { "color": "orange", "value": 0.5 },
          { "color": "red", "value": 1 }
        ]
      }
    }
  }
}
```

### Panel 4: Active Goroutines (Time Series)

```json
{
  "title": "Active Goroutines",
  "type": "timeseries",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "go_goroutines{job=\"task-manager-api\"}",
      "legendFormat": "Goroutines"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "short",
      "custom": {
        "lineWidth": 2,
        "fillOpacity": 20,
        "gradientMode": "scheme"
      }
    }
  }
}
```

### Panel 5: Memory Usage (Time Series)

```json
{
  "title": "Memory Usage",
  "type": "timeseries",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "go_memstats_alloc_bytes{job=\"task-manager-api\"} / 1024 / 1024",
      "legendFormat": "Allocated (MB)"
    },
    {
      "expr": "go_memstats_sys_bytes{job=\"task-manager-api\"} / 1024 / 1024",
      "legendFormat": "System (MB)"
    },
    {
      "expr": "process_resident_memory_bytes{job=\"task-manager-api\"} / 1024 / 1024",
      "legendFormat": "RSS (MB)"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "decmbytes",
      "custom": {
        "lineWidth": 2,
        "fillOpacity": 10
      }
    }
  }
}
```

### Panel 6: Top Endpoints by Request Count (Table)

```json
{
  "title": "Top Endpoints by Request Rate",
  "type": "table",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "topk(10, sum by (method, path) (rate(http_requests_total[5m])))",
      "legendFormat": "{{method}} {{path}}",
      "format": "table",
      "instant": true
    }
  ],
  "transformations": [
    {
      "id": "organize",
      "options": {
        "renameByName": {
          "method": "Method",
          "path": "Endpoint",
          "Value": "Req/s"
        }
      }
    }
  ]
}
```

### Panel 7: HTTP Status Codes Distribution (Pie Chart)

```json
{
  "title": "HTTP Status Distribution",
  "type": "piechart",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "sum by (status) (increase(http_requests_total[1h]))",
      "legendFormat": "{{status}}"
    }
  ]
}
```

### Panel 8: Database Connections (Time Series)

```json
{
  "title": "Database Connection Pool",
  "type": "timeseries",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "db_connections_active{job=\"task-manager-api\"}",
      "legendFormat": "Active"
    },
    {
      "expr": "db_connections_idle{job=\"task-manager-api\"}",
      "legendFormat": "Idle"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "short",
      "custom": {
        "lineWidth": 2,
        "fillOpacity": 30,
        "stacking": { "mode": "normal" }
      }
    }
  }
}
```

### Panel 9: Business Metrics — Tasks Created (Stat)

```json
{
  "title": "Tasks Created (last 24h)",
  "type": "stat",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "increase(tasks_created_total[24h])",
      "legendFormat": "Tasks"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "short",
      "thresholds": {
        "mode": "absolute",
        "steps": [
          { "color": "blue", "value": null }
        ]
      }
    }
  }
}
```

### Full Dashboard JSON Export

Ini contoh full dashboard JSON yang bisa kamu import langsung ke Grafana:

```json
{
  "dashboard": {
    "id": null,
    "uid": "task-manager-api",
    "title": "Task Manager API - Overview",
    "tags": ["task-manager", "api", "go"],
    "timezone": "browser",
    "refresh": "10s",
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "panels": [
      {
        "id": 1,
        "title": "Request Rate",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{job=\"task-manager-api\"}[5m]))",
            "legendFormat": "Total RPS"
          }
        ],
        "fieldConfig": {
          "defaults": { "unit": "reqps" }
        }
      },
      {
        "id": 2,
        "title": "Error Rate %",
        "type": "stat",
        "gridPos": { "h": 8, "w": 6, "x": 12, "y": 0 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{job=\"task-manager-api\",status=~\"5..\"}[5m])) / sum(rate(http_requests_total{job=\"task-manager-api\"}[5m])) * 100",
            "legendFormat": "Error %"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "thresholds": {
              "steps": [
                { "color": "green", "value": null },
                { "color": "yellow", "value": 1 },
                { "color": "red", "value": 5 }
              ]
            }
          }
        }
      },
      {
        "id": 3,
        "title": "P95 Latency",
        "type": "gauge",
        "gridPos": { "h": 8, "w": 6, "x": 18, "y": 0 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{job=\"task-manager-api\"}[5m])))",
            "legendFormat": "P95"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "s",
            "min": 0,
            "max": 2,
            "thresholds": {
              "steps": [
                { "color": "green", "value": null },
                { "color": "yellow", "value": 0.2 },
                { "color": "red", "value": 0.5 }
              ]
            }
          }
        }
      },
      {
        "id": 4,
        "title": "Goroutines",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 8 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "go_goroutines{job=\"task-manager-api\"}",
            "legendFormat": "Goroutines"
          }
        ]
      },
      {
        "id": 5,
        "title": "Memory Usage (MB)",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 12, "y": 8 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "go_memstats_alloc_bytes{job=\"task-manager-api\"} / 1024 / 1024",
            "legendFormat": "Alloc MB"
          },
          {
            "expr": "process_resident_memory_bytes{job=\"task-manager-api\"} / 1024 / 1024",
            "legendFormat": "RSS MB"
          }
        ],
        "fieldConfig": {
          "defaults": { "unit": "decmbytes" }
        }
      },
      {
        "id": 6,
        "title": "Top Endpoints",
        "type": "table",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 16 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "topk(10, sum by (method, path) (rate(http_requests_total{job=\"task-manager-api\"}[5m])))",
            "format": "table",
            "instant": true
          }
        ]
      },
      {
        "id": 7,
        "title": "DB Connections",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 12, "y": 16 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "db_connections_active{job=\"task-manager-api\"}",
            "legendFormat": "Active"
          },
          {
            "expr": "db_connections_idle{job=\"task-manager-api\"}",
            "legendFormat": "Idle"
          }
        ]
      },
      {
        "id": 8,
        "title": "Tasks Created (24h)",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 0, "y": 24 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "increase(tasks_created_total[24h])",
            "legendFormat": "Tasks"
          }
        ]
      },
      {
        "id": 9,
        "title": "Login Success Rate",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 6, "y": 24 },
        "datasource": "Prometheus",
        "targets": [
          {
            "expr": "rate(login_attempts_total{status=\"success\"}[5m]) / (rate(login_attempts_total{status=\"success\"}[5m]) + rate(login_attempts_total{status=\"failed\"}[5m])) * 100",
            "legendFormat": "Success %"
          }
        ],
        "fieldConfig": {
          "defaults": { "unit": "percent" }
        }
      }
    ]
  }
}
```

Import dashboard ini di Grafana:
```
1. Dashboards → New → Import
2. Paste JSON di atas
3. Klik "Load"
4. Pilih data source: Prometheus
5. Klik "Import"
```

---

## 15.12 Alerting

### Prometheus Alertmanager vs Grafana Alerts

```
Prometheus Alertmanager:
━━━━━━━━━━━━━━━━━━━━━━
- Alert rules di-define di Prometheus (YAML)
- Alertmanager handle routing, grouping, silencing
- Lebih powerful untuk complex alert routing
- Standard di production

Grafana Alerts:
━━━━━━━━━━━━━━
- Alert rules di-define di Grafana UI
- Lebih gampang setup (visual)
- Integrated langsung sama dashboard
- Cocok untuk setup sederhana

Rekomendasi:
- Mulai dengan Grafana Alerts (lebih gampang)
- Migrate ke Alertmanager kalau butuh complex routing
```

### Alert Rules — Prometheus

```yaml
# monitoring/prometheus/alerts/api-alerts.yml

groups:
  - name: task-manager-api
    rules:
      # Alert 1: High Error Rate (> 5%)
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{job="task-manager-api",status=~"5.."}[5m]))
          / sum(rate(http_requests_total{job="task-manager-api"}[5m]))
          * 100 > 5
        for: 2m                    # Alert setelah 2 menit terus-menerus
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | printf \"%.1f\" }}% (threshold: 5%)"

      # Alert 2: High Latency (P95 > 500ms)
      - alert: HighLatencyP95
        expr: |
          histogram_quantile(0.95,
            sum by (le) (rate(http_request_duration_seconds_bucket{job="task-manager-api"}[5m]))
          ) > 0.5
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "High P95 latency"
          description: "P95 latency is {{ $value | printf \"%.3f\" }}s (threshold: 500ms)"

      # Alert 3: Service Down (target not reachable)
      - alert: ServiceDown
        expr: up{job="task-manager-api"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Task Manager API is DOWN"
          description: "Target {{ $labels.instance }} has been down for more than 1 minute"

      # Alert 4: High Memory Usage (> 80% of system memory)
      - alert: HighMemoryUsage
        expr: |
          process_resident_memory_bytes{job="task-manager-api"}
          / 1024 / 1024 > 512
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value | printf \"%.0f\" }}MB (threshold: 512MB)"

      # Alert 5: Too Many Goroutines (possible goroutine leak)
      - alert: GoroutineLeak
        expr: go_goroutines{job="task-manager-api"} > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Possible goroutine leak"
          description: "Goroutine count is {{ $value }} (threshold: 1000)"

      # Alert 6: High DB Connection Usage
      - alert: HighDBConnections
        expr: |
          db_connections_active{job="task-manager-api"}
          / (db_connections_active{job="task-manager-api"} + db_connections_idle{job="task-manager-api"})
          * 100 > 80
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "High database connection usage"
          description: "DB connections are {{ $value | printf \"%.0f\" }}% utilized"

      # Alert 7: No Requests (app mungkin hang/frozen)
      - alert: NoRequests
        expr: sum(rate(http_requests_total{job="task-manager-api"}[5m])) == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "No requests received"
          description: "API has not received any requests in the last 5 minutes"
```

### Alertmanager Config — Send to Telegram

```yaml
# monitoring/alertmanager/alertmanager.yml

global:
  resolve_timeout: 5m

route:
  # Default receiver
  receiver: "telegram"

  # Group alerts by alert name
  group_by: ["alertname", "severity"]
  group_wait: 30s        # Wait 30s sebelum send grouped alerts
  group_interval: 5m     # Interval between grouped alerts
  repeat_interval: 4h    # Repeat alert setiap 4 jam kalau masih firing

  # Route specific alerts to specific receivers
  routes:
    - match:
        severity: critical
      receiver: "telegram"
      repeat_interval: 1h    # Critical alerts repeat setiap 1 jam

    - match:
        severity: warning
      receiver: "discord"
      repeat_interval: 4h

receivers:
  # Telegram Bot
  - name: "telegram"
    telegram_configs:
      - bot_token: "YOUR_BOT_TOKEN_HERE"    # Get from @BotFather
        chat_id: -1001234567890              # Your chat/group ID
        parse_mode: "HTML"
        message: |
          <b>{{ .Status | toUpper }}</b>
          {{ range .Alerts }}
          <b>Alert:</b> {{ .Labels.alertname }}
          <b>Severity:</b> {{ .Labels.severity }}
          <b>Summary:</b> {{ .Annotations.summary }}
          <b>Description:</b> {{ .Annotations.description }}
          {{ end }}

  # Discord Webhook
  - name: "discord"
    discord_configs:
      - webhook_url: "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"
        title: '{{ .Status | toUpper }}: {{ .CommonLabels.alertname }}'
        message: |
          {{ range .Alerts }}
          **Alert:** {{ .Labels.alertname }}
          **Severity:** {{ .Labels.severity }}
          **Description:** {{ .Annotations.description }}
          {{ end }}

  # Slack (kalau pakai)
  - name: "slack"
    slack_configs:
      - api_url: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
        channel: "#alerts"
        title: '{{ .Status | toUpper }}: {{ .CommonLabels.alertname }}'
        text: |
          {{ range .Alerts }}
          *Alert:* {{ .Labels.alertname }}
          *Severity:* {{ .Labels.severity }}
          *Description:* {{ .Annotations.description }}
          {{ end }}

  # Email (optional)
  - name: "email"
    email_configs:
      - to: "devops@yourcompany.com"
        from: "alertmanager@yourcompany.com"
        smarthost: "smtp.gmail.com:587"
        auth_username: "alertmanager@yourcompany.com"
        auth_password: "your-app-password"
        auth_identity: "alertmanager@yourcompany.com"
```

### Setup Telegram Bot (Step-by-step)

```
1. Buka Telegram, cari @BotFather
2. Send: /newbot
3. Kasih nama bot: "Task Manager Alert Bot"
4. Kasih username: task_manager_alert_bot
5. Dapat BOT TOKEN — simpan!

6. Bikin group baru di Telegram
7. Add bot ke group
8. Send pesan apapun di group

9. Get chat ID:
   curl https://api.telegram.org/bot<BOT_TOKEN>/getUpdates
   Cari "chat":{"id": -1001234567890, ...}

10. Masukkan BOT_TOKEN dan CHAT_ID ke alertmanager.yml
```

### Grafana Alert Rules (Alternative — Lebih Gampang)

Kalau mau setup alert langsung di Grafana (tanpa Alertmanager):

```
1. Buka Grafana → Alerting → Alert Rules
2. Klik "New Alert Rule"
3. Configure:
   - Name: "High Error Rate"
   - Query: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100
   - Condition: Is above 5
   - Evaluate every: 1m
   - For: 2m
4. Add notification:
   - Contact point: Telegram (setup di Alerting → Contact Points)
5. Save

Contact Points setup:
1. Alerting → Contact Points → New
2. Type: Telegram
3. BOT API Token: <your token>
4. Chat ID: <your chat id>
5. Test → Save
```

---

## 15.13 Structured Logging

### Kenapa Structured Logging?

Logging tradisional itu susah di-parse:

```
# Unstructured log — susah di-parse mesin
2026-05-01 10:15:33 INFO User login successful for john@example.com from IP 203.0.113.50
2026-05-01 10:15:34 ERROR Failed to create task: board not found (board_id=5, user_id=42)

# Structured log (JSON) — gampang di-parse mesin
{"time":"2026-05-01T10:15:33Z","level":"info","msg":"User login successful","email":"john@example.com","ip":"203.0.113.50","request_id":"abc-123"}
{"time":"2026-05-01T10:15:34Z","level":"error","msg":"Failed to create task","error":"board not found","board_id":5,"user_id":42,"request_id":"abc-124"}
```

```
Kenapa JSON logs lebih bagus:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Machine-readable — bisa di-index, search, filter
2. Consistent format — setiap log entry punya struktur yang sama
3. Queryable — "tampilkan semua log dengan level=error AND user_id=42"
4. Log aggregation — Loki/ELK bisa parse otomatis
5. Correlation — request_id bisa link logs dari satu request
```

### zerolog — Fast Structured Logging for Go

`zerolog` itu library logging Go yang **super cepat** (zero allocation). Alternative lainnya: `slog` (Go 1.21+ standard library), `zap` (Uber).

```bash
go get github.com/rs/zerolog
```

### Setup Logger

```go
// internal/logger/logger.go
package logger

import (
    "io"
    "os"
    "time"

    "github.com/rs/zerolog"
)

var Log zerolog.Logger

// Init initializes the global logger
func Init(env string) {
    // Set time format
    zerolog.TimeFieldFormat = time.RFC3339

    var output io.Writer

    if env == "development" {
        // Pretty print untuk development (human-readable)
        output = zerolog.ConsoleWriter{
            Out:        os.Stdout,
            TimeFormat: "15:04:05",
        }
    } else {
        // JSON output untuk production (machine-readable)
        output = os.Stdout
    }

    // Set log level
    level := zerolog.InfoLevel
    if env == "development" {
        level = zerolog.DebugLevel
    }

    Log = zerolog.New(output).
        Level(level).
        With().
        Timestamp().
        Str("service", "task-manager-api").
        Caller(). // include file:line
        Logger()
}
```

### Log Levels

```go
// Log levels dari paling rendah ke paling tinggi:
logger.Log.Trace().Msg("Very detailed info")    // -1: trace
logger.Log.Debug().Msg("Debug information")      //  0: debug
logger.Log.Info().Msg("Normal operation")        //  1: info
logger.Log.Warn().Msg("Something suspicious")    //  2: warn
logger.Log.Error().Msg("Something failed")       //  3: error
logger.Log.Fatal().Msg("Cannot continue")        //  4: fatal (exit 1)
logger.Log.Panic().Msg("Unexpected state")       //  5: panic

// Contoh penggunaan
logger.Log.Info().
    Str("email", "john@example.com").
    Str("ip", c.ClientIP()).
    Msg("User login successful")

logger.Log.Error().
    Err(err).
    Uint("board_id", boardID).
    Uint("user_id", userID).
    Msg("Failed to create task")

logger.Log.Warn().
    Str("ip", c.ClientIP()).
    Int("count", int(count)).
    Int("limit", limit).
    Msg("Rate limit exceeded")
```

### Request Logging Middleware with Correlation ID

Setiap request harus punya **correlation ID** (atau request ID) — biar kamu bisa track semua log dari satu request.

```go
// internal/middleware/request_logger.go
package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/irzan-aldi/task-manager/internal/logger"
)

// RequestLoggerMiddleware logs every HTTP request with correlation ID
func RequestLoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generate unique request ID
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery
        method := c.Request.Method
        clientIP := c.ClientIP()
        userAgent := c.Request.UserAgent()

        // Process request
        c.Next()

        // After request
        duration := time.Since(start)
        statusCode := c.Writer.Status()
        responseSize := c.Writer.Size()

        // Build log entry
        event := logger.Log.Info()
        if statusCode >= 500 {
            event = logger.Log.Error()
        } else if statusCode >= 400 {
            event = logger.Log.Warn()
        }

        event.
            Str("request_id", requestID).
            Str("method", method).
            Str("path", path).
            Str("query", query).
            Int("status", statusCode).
            Dur("duration", duration).
            Int("response_size", responseSize).
            Str("client_ip", clientIP).
            Str("user_agent", userAgent).
            Msg("HTTP request")
    }
}
```

### Init Logger di main.go

```go
// cmd/api/main.go
package main

import (
    "github.com/irzan-aldi/task-manager/internal/config"
    "github.com/irzan-aldi/task-manager/internal/logger"
    // ... other imports
)

func main() {
    // Load config
    cfg := config.Load()

    // Initialize structured logger
    logger.Init(cfg.Env)
    logger.Log.Info().
        Str("env", cfg.Env).
        Str("port", cfg.Port).
        Msg("Starting Task Manager API")

    // ... rest of setup

    // Register middleware in order
    r := gin.New() // pakai gin.New() bukan gin.Default() (kita bikin logger sendiri)
    r.Use(gin.Recovery()) // recovery dari panic
    r.Use(middleware.RequestLoggerMiddleware()) // structured logging
    r.Use(middleware.PrometheusMiddleware())    // prometheus metrics

    // ... rest of routes
}
```

### Output Contoh

Development mode (pretty print):
```
10:15:33 INF HTTP request method=GET path=/api/v1/tasks/my status=200 duration=12.345ms response_size=1234 client_ip=127.0.0.1 request_id=abc-123-def
10:15:34 ERR HTTP request method=POST path=/api/v1/boards/5/tasks status=403 duration=5.678ms response_size=89 client_ip=127.0.0.1 request_id=abc-124-def
```

Production mode (JSON):
```json
{"time":"2026-05-01T10:15:33Z","level":"info","service":"task-manager-api","caller":"middleware/request_logger.go:52","request_id":"abc-123-def","method":"GET","path":"/api/v1/tasks/my","status":200,"duration":12.345,"response_size":1234,"client_ip":"127.0.0.1","user_agent":"curl/8.1.2","msg":"HTTP request"}
{"time":"2026-05-01T10:15:34Z","level":"error","service":"task-manager-api","caller":"middleware/request_logger.go:52","request_id":"abc-124-def","method":"POST","path":"/api/v1/boards/5/tasks","status":403,"duration":5.678,"response_size":89,"client_ip":"127.0.0.1","user_agent":"curl/8.1.2","msg":"HTTP request"}
```

---

## 15.14 Log Aggregation — Loki + Grafana

### ELK vs Loki — Pilih yang Mana?

```
ELK Stack (Elasticsearch + Logstash + Kibana):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Full-text search (powerful)
- Complex queries
- Berat (Elasticsearch butuh banyak RAM — min 4GB)
- Cocok untuk perusahaan besar
- Setup complex

Loki + Grafana:
━━━━━━━━━━━━━━
- "Like Prometheus, but for logs"
- Ga full-text index (index by labels only — lebih hemat)
- Ringan (bisa jalan dengan 256MB RAM)
- Integrated langsung sama Grafana
- Setup simple
- Cocok untuk startup, small-medium team
- Pakai LogQL (mirip PromQL)

Rekomendasi: Loki + Grafana
Kenapa? Kita udah pakai Grafana untuk metrics (Prometheus).
Tambahin Loki = logs juga di Grafana.
Satu dashboard untuk semua.
```

### Setup Loki

```yaml
# monitoring/loki/loki-config.yml

auth_enabled: false

server:
  http_listen_port: 3100

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2020-10-24
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

limits_config:
  reject_old_samples: true
  reject_old_samples_max_age: 168h    # 7 days
  max_query_length: 721h              # 30 days

analytics:
  reporting_enabled: false
```

### Setup Promtail (Log Shipper)

Promtail itu agent yang **baca log files** dan **kirim ke Loki**. Bisa juga pakai Docker log driver langsung.

```yaml
# monitoring/promtail/promtail-config.yml

server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  # Scrape Docker container logs
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
    relabel_configs:
      # Use container name as label
      - source_labels: ["__meta_docker_container_name"]
        target_label: "container"
        regex: "/(.*)"
      # Use compose service name
      - source_labels: ["__meta_docker_container_label_com_docker_compose_service"]
        target_label: "service"
    pipeline_stages:
      # Parse JSON logs
      - json:
          expressions:
            level: level
            msg: msg
            request_id: request_id
            method: method
            path: path
            status: status
      - labels:
          level:
          service:
```

### Query Logs di Grafana (LogQL)

Setelah Loki terhubung ke Grafana sebagai data source, kamu bisa query logs pakai LogQL:

```logql
# Semua logs dari task-manager-api
{service="api"}

# Filter by log level
{service="api"} |= "error"
{service="api"} | json | level="error"

# Filter by specific request ID (untuk debugging)
{service="api"} |= "abc-123-def"

# Filter by path
{service="api"} | json | path="/api/v1/tasks"

# Filter by status code 500
{service="api"} | json | status >= 500

# Count errors per minute
count_over_time({service="api"} | json | level="error" [1m])

# Top error messages
{service="api"} | json | level="error" | line_format "{{.msg}}"
```

```
Setup Loki di Grafana:
1. Connections → Data Sources → Add data source
2. Pilih "Loki"
3. URL: http://loki:3100
4. Save & Test
5. Explore → pilih Loki → query logs
```

---

## 15.15 Docker Compose — Full Monitoring Stack

Ini yang kamu tunggu-tunggu. **Satu docker-compose** yang jalanin semuanya:

```yaml
# docker-compose.monitoring.yml
# Full production-like monitoring stack
# Usage: docker compose -f docker-compose.monitoring.yml up -d

services:
  # ==========================================
  # APPLICATION SERVICES
  # ==========================================

  # Go Task Manager API
  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: task-manager-api
    environment:
      - APP_ENV=production
      - PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=taskmanager
      - DB_PASSWORD=secretpassword
      - DB_NAME=task_manager
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=your-super-secret-jwt-key-change-this
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - app-network
    restart: unless-stopped
    labels:
      - "logging=promtail"
    # JANGAN expose port 8080 ke host — lewat Nginx aja
    expose:
      - "8080"

  # PostgreSQL Database
  postgres:
    image: postgres:17-alpine
    container_name: task-manager-db
    environment:
      - POSTGRES_USER=taskmanager
      - POSTGRES_PASSWORD=secretpassword
      - POSTGRES_DB=task_manager
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U taskmanager -d task_manager"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network
    restart: unless-stopped

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: task-manager-redis
    command: redis-server --maxmemory 128mb --maxmemory-policy allkeys-lru
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network
    restart: unless-stopped

  # Nginx Reverse Proxy
  nginx:
    image: nginx:alpine
    container_name: task-manager-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
    depends_on:
      - api
    networks:
      - app-network
    restart: unless-stopped

  # ==========================================
  # MONITORING SERVICES
  # ==========================================

  # Prometheus — Metrics Collection
  prometheus:
    image: prom/prometheus:v3.4.0
    container_name: prometheus
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--storage.tsdb.path=/prometheus"
      - "--storage.tsdb.retention.time=15d"        # Keep data 15 days
      - "--storage.tsdb.retention.size=5GB"         # Max 5GB storage
      - "--web.enable-lifecycle"                     # Enable config reload via API
      - "--web.enable-admin-api"
    volumes:
      - ./monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./monitoring/prometheus/alerts:/etc/prometheus/alerts:ro
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
    networks:
      - app-network
    restart: unless-stopped

  # Grafana — Visualization & Dashboards
  grafana:
    image: grafana/grafana:11.6.0
    container_name: grafana
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_SERVER_ROOT_URL=http://localhost:3000
      # Auto-provision data sources
      - GF_INSTALL_PLUGINS=grafana-clock-panel,grafana-simple-json-datasource
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
    ports:
      - "3000:3000"
    depends_on:
      - prometheus
      - loki
    networks:
      - app-network
    restart: unless-stopped

  # Loki — Log Aggregation
  loki:
    image: grafana/loki:3.5.0
    container_name: loki
    command: -config.file=/etc/loki/loki-config.yml
    volumes:
      - ./monitoring/loki/loki-config.yml:/etc/loki/loki-config.yml:ro
      - loki_data:/loki
    ports:
      - "3100:3100"
    networks:
      - app-network
    restart: unless-stopped

  # Promtail — Log Shipper (sends logs to Loki)
  promtail:
    image: grafana/promtail:3.5.0
    container_name: promtail
    command: -config.file=/etc/promtail/promtail-config.yml
    volumes:
      - ./monitoring/promtail/promtail-config.yml:/etc/promtail/promtail-config.yml:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro  # Access Docker logs
    depends_on:
      - loki
    networks:
      - app-network
    restart: unless-stopped

  # Alertmanager — Alert Routing
  alertmanager:
    image: prom/alertmanager:v0.28.1
    container_name: alertmanager
    command:
      - "--config.file=/etc/alertmanager/alertmanager.yml"
      - "--storage.path=/alertmanager"
    volumes:
      - ./monitoring/alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
      - alertmanager_data:/alertmanager
    ports:
      - "9093:9093"
    networks:
      - app-network
    restart: unless-stopped

  # Node Exporter — System Metrics (CPU, Memory, Disk, Network)
  node-exporter:
    image: prom/node-exporter:v1.9.1
    container_name: node-exporter
    command:
      - "--path.procfs=/host/proc"
      - "--path.rootfs=/rootfs"
      - "--path.sysfs=/host/sys"
      - "--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    networks:
      - app-network
    restart: unless-stopped

  # PostgreSQL Exporter — Database Metrics
  postgres-exporter:
    image: prometheuscommunity/postgres-exporter:v0.16.0
    container_name: postgres-exporter
    environment:
      - DATA_SOURCE_NAME=postgresql://taskmanager:secretpassword@postgres:5432/task_manager?sslmode=disable
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - app-network
    restart: unless-stopped

  # Redis Exporter — Cache Metrics
  redis-exporter:
    image: oliver006/redis_exporter:v1.67.0
    container_name: redis-exporter
    environment:
      - REDIS_ADDR=redis:6379
    depends_on:
      redis:
        condition: service_healthy
    networks:
      - app-network
    restart: unless-stopped

# ==========================================
# VOLUMES
# ==========================================
volumes:
  postgres_data:
    driver: local
  prometheus_data:
    driver: local
  grafana_data:
    driver: local
  loki_data:
    driver: local
  alertmanager_data:
    driver: local

# ==========================================
# NETWORKS
# ==========================================
networks:
  app-network:
    driver: bridge
```

### Grafana Auto-Provisioning

Biar data sources dan dashboards otomatis ke-setup waktu Grafana start:

```yaml
# monitoring/grafana/provisioning/datasources/datasources.yml

apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: true

  - name: Alertmanager
    type: alertmanager
    access: proxy
    url: http://alertmanager:9093
    jsonData:
      implementation: prometheus
    editable: true
```

```yaml
# monitoring/grafana/provisioning/dashboards/dashboards.yml

apiVersion: 1

providers:
  - name: "default"
    orgId: 1
    folder: "Task Manager"
    type: file
    disableDeletion: false
    editable: true
    options:
      path: /var/lib/grafana/dashboards
      foldersFromFilesStructure: true
```

### Folder Structure

```
monitoring/
├── prometheus/
│   ├── prometheus.yml              # Main config
│   └── alerts/
│       └── api-alerts.yml          # Alert rules
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── datasources.yml     # Auto-config data sources
│   │   └── dashboards/
│   │       └── dashboards.yml      # Auto-config dashboards
│   └── dashboards/
│       └── task-manager.json       # Dashboard JSON
├── loki/
│   └── loki-config.yml             # Loki config
├── promtail/
│   └── promtail-config.yml         # Promtail config
└── alertmanager/
    └── alertmanager.yml            # Alertmanager config
```

### Start Everything

```bash
# Start full stack
docker compose -f docker-compose.monitoring.yml up -d

# Check all containers running
docker compose -f docker-compose.monitoring.yml ps

# Expected output:
# NAME                   STATUS          PORTS
# task-manager-api       Up (healthy)    8080/tcp
# task-manager-db        Up (healthy)    5432/tcp
# task-manager-redis     Up (healthy)    6379/tcp
# task-manager-nginx     Up              0.0.0.0:80->80/tcp
# prometheus             Up              0.0.0.0:9090->9090/tcp
# grafana                Up              0.0.0.0:3000->3000/tcp
# loki                   Up              0.0.0.0:3100->3100/tcp
# promtail               Up              9080/tcp
# alertmanager           Up              0.0.0.0:9093->9093/tcp
# node-exporter          Up              9100/tcp
# postgres-exporter      Up              9187/tcp
# redis-exporter         Up              9121/tcp

# Access:
# App:          http://localhost (via Nginx)
# Prometheus:   http://localhost:9090
# Grafana:      http://localhost:3000 (admin/admin123)
# Alertmanager: http://localhost:9093
# Loki:         http://localhost:3100

# View logs
docker compose -f docker-compose.monitoring.yml logs -f api
docker compose -f docker-compose.monitoring.yml logs -f prometheus

# Stop everything
docker compose -f docker-compose.monitoring.yml down

# Stop and remove all data (clean start)
docker compose -f docker-compose.monitoring.yml down -v
```

---

## 15.16 Health Check Patterns

### Basic Health Check — /health

Ini yang udah kita punya:

```go
// internal/handler/health.go — EXISTING
func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
    })
}
```

Ini cuma cek "apakah app-nya hidup?". Ga cek apakah database connected, apakah Redis available. Untuk production, kita butuh lebih dari ini.

### Readiness Check — /ready

Readiness check itu cek **dependency** — database, Redis, external services. Kalau dependency ga available, app dianggap "not ready" untuk terima traffic.

```go
// internal/handler/health.go — UPDATED
package handler

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

type HealthHandler struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
    return &HealthHandler{db: db, redis: redis}
}

// HealthCheck — basic liveness check
// "Is the app alive?"
func (h *HealthHandler) HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "ok",
        "service": "task-manager-api",
        "time":    time.Now().UTC().Format(time.RFC3339),
    })
}

// ReadinessCheck — dependency check
// "Is the app ready to serve traffic?"
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    checks := make(map[string]string)
    healthy := true

    // Check database
    sqlDB, err := h.db.DB()
    if err != nil {
        checks["database"] = "error: " + err.Error()
        healthy = false
    } else if err := sqlDB.PingContext(ctx); err != nil {
        checks["database"] = "error: " + err.Error()
        healthy = false
    } else {
        checks["database"] = "ok"
    }

    // Check Redis
    if err := h.redis.Ping(ctx).Err(); err != nil {
        checks["redis"] = "error: " + err.Error()
        healthy = false
    } else {
        checks["redis"] = "ok"
    }

    status := http.StatusOK
    statusText := "ready"
    if !healthy {
        status = http.StatusServiceUnavailable
        statusText = "not ready"
    }

    c.JSON(status, gin.H{
        "status":  statusText,
        "service": "task-manager-api",
        "checks":  checks,
        "time":    time.Now().UTC().Format(time.RFC3339),
    })
}
```

### Response Contoh

```bash
# App healthy:
curl http://localhost:8080/ready
{
  "status": "ready",
  "service": "task-manager-api",
  "checks": {
    "database": "ok",
    "redis": "ok"
  },
  "time": "2026-05-01T10:15:33Z"
}

# Database down:
curl http://localhost:8080/ready
# HTTP 503
{
  "status": "not ready",
  "service": "task-manager-api",
  "checks": {
    "database": "error: connection refused",
    "redis": "ok"
  },
  "time": "2026-05-01T10:15:33Z"
}
```

### Liveness vs Readiness — Konsep Kubernetes

Meskipun kita belum pakai Kubernetes (itu Bab 16), konsep ini penting untuk dipahami sekarang:

```
Liveness Probe (/health):
━━━━━━━━━━━━━━━━━━━━━━━━
"Is the app alive?"
Kalau FAIL → Kubernetes RESTART container
Contoh gagal: app crash, deadlock, infinite loop

Readiness Probe (/ready):
━━━━━━━━━━━━━━━━━━━━━━━━
"Is the app ready to serve traffic?"
Kalau FAIL → Kubernetes STOP sending traffic (tapi ga restart)
Contoh gagal: database down, dependency unavailable, cold start

Startup Probe:
━━━━━━━━━━━━━━
"Has the app finished starting?"
Untuk slow-starting apps. Kalau FAIL → Kubernetes RESTART
Contoh: app butuh 30 detik warm up
```

```
Di Docker Compose, kita pakai healthcheck:

services:
  api:
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

### Prometheus Scrape Health Check

Prometheus juga otomatis jadi "health checker". Kalau scrape gagal, target statusnya `DOWN`:

```promql
# Cek target health
up{job="task-manager-api"}

# 1 = UP (healthy)
# 0 = DOWN (unreachable)
```

---

## 15.17 Uptime Monitoring — External Monitoring

### Kenapa Butuh External Monitoring?

```
Internal monitoring (Prometheus):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Monitor dari dalam infrastructure
- Kalau SERVER mati → Prometheus juga mati → ga ada alert
- Blind spot: network issue between server and internet

External monitoring (UptimeRobot):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Monitor dari LUAR infrastructure
- Cek apakah app accessible dari internet
- Kalau server mati → external monitor DETECT → alert
- Cover blind spot: network, DNS, CDN issues
```

### Tools External Monitoring

```
Free:
━━━━━
1. UptimeRobot (https://uptimerobot.com)
   - 50 monitors gratis
   - 5-minute interval (free), 1-minute (paid)
   - HTTP, keyword, ping monitoring
   - Email + Telegram + Slack alerts
   - Status page gratis
   - RECOMMENDED untuk mulai

2. Freshping (https://freshping.io)
   - 50 monitors gratis
   - 1-minute interval
   - Status page gratis

Paid:
━━━━━
3. Better Stack (https://betterstack.com)
   - 30-second interval
   - Incident management built-in
   - Beautiful status pages
   - Mulai $24/bulan

4. Pingdom (https://pingdom.com)
   - Industry standard
   - Synthetic monitoring
   - RUM (Real User Monitoring)
   - Mulai $15/bulan

5. Datadog Synthetics
   - API tests, browser tests
   - Multi-step monitoring
   - Integrated dengan Datadog platform
```

### Setup UptimeRobot (Free)

```
1. Daftar di https://uptimerobot.com (gratis)
2. Add New Monitor:
   - Monitor Type: HTTP(s)
   - Friendly Name: "Task Manager API"
   - URL: https://api.taskmanager.com/health
   - Monitoring Interval: 5 minutes
3. Alert Contacts:
   - Email: your@email.com
   - Telegram: (setup Telegram integration)
4. Advanced:
   - HTTP Method: GET
   - Expected Status: 200
   - Keyword: "ok" (cek response body contains "ok")

5. Add more monitors:
   - https://api.taskmanager.com/ready (readiness)
   - https://api.taskmanager.com/api/v1/auth/login (actual endpoint — POST)
```

### Status Page

UptimeRobot juga kasih **status page gratis** — halaman publik yang kasih tau user apakah service kamu UP:

```
Status Page URL: https://status.taskmanager.com

┌──────────────────────────────────────────┐
│          Task Manager Status             │
│                                          │
│  API Server          ● Operational       │
│  Database            ● Operational       │
│  Redis Cache         ● Operational       │
│  Website             ● Operational       │
│                                          │
│  Overall Uptime (30 days): 99.95%        │
│                                          │
│  Recent Incidents:                       │
│  May 1 - Brief API outage (5 min)       │
└──────────────────────────────────────────┘
```

---

## 15.18 Performance Profiling — Go pprof

### Kapan Perlu Profiling?

```
Monitoring bilang: "CPU usage tinggi!"
Logging bilang: "Banyak request ke /api/v1/tasks"
Tapi... KENAPA CPU usage tinggi? Function mana yang berat?

Jawaban: PROFILING.

Monitoring = WHAT (apa yang terjadi?)
Logging   = WHEN (kapan terjadi?)
Profiling = WHY (kenapa terjadi?)
```

### Go pprof — Built-in Profiler

Go punya profiler bawaan yang luar biasa powerful — `net/http/pprof`. Tinggal import, langsung jalan.

```go
// cmd/api/main.go — tambah pprof (HANYA di development/staging!)
package main

import (
    "net/http"
    _ "net/http/pprof"  // Register pprof handlers
)

func main() {
    // ... existing setup ...

    // Start pprof server di port terpisah (JANGAN expose ke public!)
    if cfg.Env != "production" {
        go func() {
            log.Println("pprof server starting on :6060")
            log.Println(http.ListenAndServe(":6060", nil))
        }()
    }

    // ... rest of main
}
```

### Profiling Tools

```bash
# CPU Profile — function mana yang pakai CPU paling banyak
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory Profile — function mana yang allocate memory paling banyak
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine Profile — goroutine apa aja yang jalan
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Block Profile — goroutine mana yang ke-block paling lama
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex Profile — mutex contention
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### Interactive pprof

```bash
# Download CPU profile 30 detik
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Di interactive mode:
(pprof) top 10          # Top 10 function by CPU
(pprof) top -cum 10     # Top 10 by cumulative time
(pprof) list functionName  # Show source code of function
(pprof) web             # Open flame graph di browser
(pprof) svg             # Generate SVG flame graph
```

### Flame Graphs

Flame graph itu visualisasi yang bikin kamu langsung tau function mana yang "paling panas" (paling banyak pakai resource).

```bash
# Install go-torch atau pakai pprof web langsung
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30

# Buka browser: http://localhost:8081
# Klik "Flame Graph" → lihat function yang berat
```

```
Flame Graph Reading:
━━━━━━━━━━━━━━━━━━

Lebar bar = waktu yang dihabiskan
Semakin lebar = semakin banyak waktu di function itu

┌─────────────────────────────────────────────────┐
│                    main.main                      │
├──────────────────────┬──────────────────────────┤
│  router.handleRequest │  db.Query                │
├────────┬─────────────┤──────────┬───────────────┤
│ auth() │  json.Marshal│ pgx.conn │ pgx.scan      │
└────────┴─────────────┴──────────┴───────────────┘

Dari flame graph ini: db.Query itu 50% waktu.
Optimasi query = performance boost terbesar.
```

### Kapan Profiling?

```
DO profile ketika:
- Alert: CPU usage > 80% terus-menerus
- Alert: Memory usage terus naik (possible leak)
- P95 latency tiba-tiba naik
- Goroutine count terus naik (goroutine leak)

DO NOT profile di production kalau bisa dihindari.
Profile di staging environment yang mirip production.
Kalau terpaksa di production, jalanin sebentar aja (30 detik).

JANGAN expose pprof endpoint ke public!
Port 6060 harus di-block di firewall / Nginx.
```

---

## 15.19 Incident Response

### Incident Response Flow

```
┌──────────────────────────────────────────────────────────┐
│                 Incident Response Flow                     │
│                                                            │
│  1. DETECT                                                 │
│     ├── Alert masuk (Telegram/Discord/Email)              │
│     ├── User report                                        │
│     └── External monitoring (UptimeRobot)                 │
│                                                            │
│  2. TRIAGE                                                 │
│     ├── Seberapa parah? (P1/P2/P3/P4)                    │
│     ├── Berapa user yang affected?                        │
│     └── Apakah ada workaround?                            │
│                                                            │
│  3. INVESTIGATE                                            │
│     ├── Buka Grafana dashboard                            │
│     ├── Cek metrics: error rate, latency, CPU, memory     │
│     ├── Cek logs: apa error message-nya?                  │
│     ├── Cek recent deploys: ada deploy baru?              │
│     └── Cek external deps: DB, Redis, 3rd party          │
│                                                            │
│  4. FIX                                                    │
│     ├── Quick fix (restart, rollback, scale up)           │
│     ├── Permanent fix (code fix, config change)           │
│     └── Verify: metrics kembali normal                    │
│                                                            │
│  5. COMMUNICATE                                            │
│     ├── Update status page                                 │
│     └── Notify stakeholders                               │
│                                                            │
│  6. POST-MORTEM                                            │
│     ├── Timeline of events                                 │
│     ├── Root cause analysis                                │
│     ├── What went well                                     │
│     ├── What went wrong                                    │
│     └── Action items (prevent recurrence)                 │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

### Severity Levels

```
P1 — CRITICAL (respond: 15 menit)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Service completely down
- Data loss / data corruption
- Security breach
Contoh: API return 502 untuk semua users

P2 — HIGH (respond: 30 menit)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Major feature broken
- Performance severely degraded
- Affecting large percentage of users
Contoh: Login ga bisa, tapi app lain masih jalan

P3 — MEDIUM (respond: 4 jam)
━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Minor feature broken
- Performance slightly degraded
- Workaround available
Contoh: Endpoint GET /tasks lambat (5s), tapi masih jalan

P4 — LOW (respond: next business day)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Cosmetic issue
- Non-critical feature
Contoh: Log format salah, metric label typo
```

### Runbook Template

Runbook itu **dokumentasi step-by-step** untuk handle incident spesifik. Bikin runbook sebelum incident terjadi.

```markdown
# Runbook: High Error Rate (> 5%)

## Detection
- Alert: "HighErrorRate" dari Prometheus/Grafana
- Dashboard: http://grafana:3000/d/task-manager-api

## Diagnosis Steps
1. Buka Grafana dashboard → cek Error Rate panel
2. Cek error rate per endpoint:
   topk(5, sum by (path, status) (rate(http_requests_total{status=~"5.."}[5m])))
3. Cek logs:
   {service="api"} | json | level="error" | line_format "{{.msg}}"
4. Cek apakah ada recent deploy:
   git log --oneline -5
5. Cek database health:
   curl http://localhost:8080/ready

## Common Root Causes
1. Bad deploy → rollback: git revert + redeploy
2. Database overloaded → check connections, slow queries
3. External service down → check dependency health
4. OOM → check memory metrics, restart if needed

## Resolution
1. If bad deploy: rollback to previous version
   docker compose up -d --build
2. If database:
   - Check connections: db_connections_active
   - Kill long queries
   - Scale DB if needed
3. If OOM: restart container
   docker compose restart api

## Escalation
- If not resolved in 30 min → ping #engineering Slack
- If data loss → ping CTO immediately
```

### On-Call Rotation

```
On-Call Rotation (konsep):
━━━━━━━━━━━━━━━━━━━━━━━━

Week 1: Developer A (primary), Developer B (secondary)
Week 2: Developer B (primary), Developer C (secondary)
Week 3: Developer C (primary), Developer A (secondary)

Primary: handle alert pertama kali
Secondary: backup kalau primary ga respond dalam 15 menit

Tools: PagerDuty, Opsgenie, Grafana OnCall (gratis!)

Yang penting:
- Alert harus actionable (jangan alert yang ga perlu action)
- Runbook harus up-to-date
- Handoff meeting setiap ganti rotation
- Jangan terlalu banyak alert (alert fatigue = berbahaya)
```

---

## 15.20 Best Practices

### USE Method — Infrastructure Monitoring

```
USE Method (by Brendan Gregg):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━

U — Utilization: seberapa sibuk resource?
S — Saturation: apakah ada yang antri/overflow?
E — Errors: ada error ga?

Aplikasi ke setiap resource:

CPU:
- Utilization: CPU usage percentage
- Saturation: CPU run queue length (waiting to be scheduled)
- Errors: CPU-related errors (rare)

Memory:
- Utilization: memory used / total memory
- Saturation: swap usage, OOM events
- Errors: allocation failures

Disk:
- Utilization: disk space used / total
- Saturation: I/O queue length
- Errors: disk read/write errors

Network:
- Utilization: bandwidth used / total
- Saturation: dropped packets, retransmits
- Errors: network errors, timeouts
```

### RED Method — Service Monitoring

```
RED Method (by Tom Wilkie — Grafana Labs):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

R — Rate: request per second
E — Errors: error rate
D — Duration: latency (P50, P95, P99)

Ini yang kita pakai di custom dashboard!

PromQL:
- Rate:     sum(rate(http_requests_total[5m]))
- Errors:   sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
- Duration: histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))

Kapan pakai RED vs USE:
- USE → infrastructure (CPU, memory, disk, network)
- RED → application services (API, microservices)
- Biasanya pakai DUA-DUANYA
```

### Label Cardinality — Jangan Bikin Prometheus OOM

```
GOOD labels (low cardinality):
- method: GET, POST, PUT, DELETE (4 values)
- status: 200, 201, 400, 401, 403, 404, 500 (~10 values)
- path: /api/v1/tasks/:id (~20 routes)
- env: production, staging (2 values)

BAD labels (high cardinality — JANGAN!):
- user_id: 1, 2, 3, ..., 1000000 (infinite!)
- request_id: abc-123, def-456, ... (infinite!)
- email: john@..., jane@..., ... (infinite!)
- timestamp: ... (infinite!)

Kenapa high cardinality berbahaya?
Setiap unique label combination = 1 time series di Prometheus.
method(4) x status(10) x path(20) = 800 time series → OK
method(4) x status(10) x user_id(1M) = 40M time series → OOM!

Rule of thumb:
- Max ~10 label values per label
- Total time series per metric: < 10,000
```

### Retention Policy

```
Berapa lama simpan data?

Short-term (1-2 minggu):
- Full resolution (15s interval)
- Untuk debugging recent issues
- Prometheus default: 15 hari

Medium-term (1-3 bulan):
- Downsampled (1m or 5m intervals)
- Untuk trend analysis
- Pakai Thanos/Cortex untuk long-term storage

Long-term (1+ tahun):
- Aggregated daily/weekly
- Untuk capacity planning, business reports
- Biasanya di separate storage

Rekomendasi awal:
- Prometheus retention: 15 hari (default)
- Loki retention: 7 hari
- Kalau butuh lebih lama: setup Thanos atau cloud storage
```

### Dashboard Organization

```
Dashboard Hierarchy:
━━━━━━━━━━━━━━━━━━

Level 1: Overview (executive summary)
├── All services at a glance
├── Total RPS, error rate, latency (overall)
├── System health (all targets UP/DOWN)
└── Link to detail dashboards

Level 2: Service Detail (per service)
├── Task Manager API Dashboard
│   ├── Request rate, error rate, latency
│   ├── Top endpoints
│   ├── Business metrics
│   └── Go runtime metrics
├── PostgreSQL Dashboard
│   ├── Connections, transactions
│   ├── Query stats, slow queries
│   └── Replication lag
└── Redis Dashboard
    ├── Memory, connections
    ├── Commands per second
    └── Hit rate, evictions

Level 3: Debug (for incident investigation)
├── Request tracing
├── Log exploration (Loki)
├── Detailed system metrics
└── Custom investigation panels

Tips:
- Jangan terlalu banyak panel di satu dashboard (max 15-20)
- Pakai variables untuk dynamic filtering
- Consistent color scheme (green=ok, yellow=warn, red=critical)
- Dashboard harus load cepat (< 3 detik)
```

---

## 15.21 Cheat Sheet

### PromQL Queries

```promql
# ===== BASIC =====
metric_name                           # Instant vector
metric_name[5m]                       # Range vector
metric_name{label="value"}            # Filter by label
metric_name{label=~"val.*"}           # Regex match
metric_name{label!="value"}           # Negative match

# ===== RATE & INCREASE =====
rate(counter[5m])                     # Per-second rate (for counter)
irate(counter[5m])                    # Instant rate (last 2 points)
increase(counter[1h])                 # Total increase in time range

# ===== AGGREGATION =====
sum(metric)                           # Sum all values
avg(metric)                           # Average
min(metric)                           # Minimum
max(metric)                           # Maximum
count(metric)                         # Count of series
topk(5, metric)                       # Top 5 values
bottomk(5, metric)                    # Bottom 5 values

# ===== BY LABEL =====
sum by (label) (metric)               # Sum grouped by label
avg by (label) (metric)               # Average grouped by label
sum without (label) (metric)          # Sum excluding label

# ===== HISTOGRAM =====
histogram_quantile(0.95, sum by (le) (rate(histogram_bucket[5m])))  # P95
histogram_quantile(0.99, sum by (le) (rate(histogram_bucket[5m])))  # P99
histogram_quantile(0.50, sum by (le) (rate(histogram_bucket[5m])))  # P50 (median)

# ===== MATH =====
metric_a / metric_b * 100            # Percentage
deriv(gauge[1h])                      # Rate of change (for gauge)
predict_linear(gauge[1h], 3600)       # Predict value 1h from now
abs(metric)                           # Absolute value
ceil(metric)                          # Round up
floor(metric)                         # Round down

# ===== TIME =====
time()                                # Current Unix timestamp
timestamp(metric)                     # Last scrape timestamp
changes(gauge[1h])                    # Number of value changes

# ===== TASK MANAGER SPECIFIC =====

# Request per second
sum(rate(http_requests_total[5m]))

# Error rate %
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# P95 latency
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))

# P95 per endpoint
histogram_quantile(0.95, sum by (le, path) (rate(http_request_duration_seconds_bucket[5m])))

# Top 5 endpoints by RPS
topk(5, sum by (path) (rate(http_requests_total[5m])))

# Tasks created per hour
increase(tasks_created_total[1h])

# Memory usage MB
process_resident_memory_bytes / 1024 / 1024

# Is memory leaking? (positive = growing)
deriv(go_memstats_alloc_bytes[1h])

# DB connection utilization %
db_connections_active / (db_connections_active + db_connections_idle) * 100
```

### Grafana Shortcuts

```
Keyboard Shortcuts (di Grafana):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

d + a     → Toggle auto-refresh
d + s     → Save dashboard
d + r     → Refresh dashboard
d + k     → Toggle kiosk mode (fullscreen)
d + e     → Expand all rows
d + c     → Collapse all rows
?         → Show all shortcuts

t + a     → Set time range to absolute
t + z     → Zoom out time range
t + left  → Move time range back
t + right → Move time range forward

e         → Edit panel (when hovering)
v         → View panel fullscreen
i         → Inspect panel data
p + s     → Toggle panel sharing
```

### Common Alert Rules

```yaml
# Alert: Service Down
- alert: ServiceDown
  expr: up == 0
  for: 1m
  labels: { severity: critical }

# Alert: High Error Rate (> 5%)
- alert: HighErrorRate
  expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100 > 5
  for: 2m
  labels: { severity: critical }

# Alert: High Latency (P95 > 500ms)
- alert: HighLatency
  expr: histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m]))) > 0.5
  for: 3m
  labels: { severity: warning }

# Alert: High Memory (> 80%)
- alert: HighMemory
  expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100 > 80
  for: 5m
  labels: { severity: warning }

# Alert: Disk Almost Full (> 85%)
- alert: DiskAlmostFull
  expr: (node_filesystem_size_bytes - node_filesystem_free_bytes) / node_filesystem_size_bytes * 100 > 85
  for: 5m
  labels: { severity: warning }

# Alert: Goroutine Leak
- alert: GoroutineLeak
  expr: go_goroutines > 1000
  for: 5m
  labels: { severity: warning }

# Alert: High DB Connections
- alert: HighDBConnections
  expr: db_connections_active / (db_connections_active + db_connections_idle) * 100 > 80
  for: 3m
  labels: { severity: warning }

# Alert: No Traffic (possible app freeze)
- alert: NoTraffic
  expr: sum(rate(http_requests_total[5m])) == 0
  for: 5m
  labels: { severity: warning }

# Alert: SSL Certificate Expiring (< 7 days)
- alert: SSLCertExpiring
  expr: probe_ssl_earliest_cert_expiry - time() < 86400 * 7
  for: 1h
  labels: { severity: warning }
```

### Docker Compose Commands

```bash
# Start monitoring stack
docker compose -f docker-compose.monitoring.yml up -d

# Stop monitoring stack
docker compose -f docker-compose.monitoring.yml down

# Restart specific service
docker compose -f docker-compose.monitoring.yml restart prometheus

# View logs
docker compose -f docker-compose.monitoring.yml logs -f grafana

# Reload Prometheus config (tanpa restart)
curl -X POST http://localhost:9090/-/reload

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Check Alertmanager alerts
curl http://localhost:9093/api/v2/alerts | jq '.[].labels.alertname'

# Test Alertmanager (send test alert)
curl -X POST http://localhost:9093/api/v2/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {"alertname": "TestAlert", "severity": "warning"},
    "annotations": {"summary": "This is a test alert"},
    "startsAt": "2026-05-01T00:00:00Z"
  }]'
```

### File Locations (dalam monitoring stack)

```
Project Structure:
━━━━━━━━━━━━━━━━

task-manager/
├── cmd/api/main.go                           # Entry point
├── internal/
│   ├── metrics/
│   │   ├── metrics.go                        # Custom metric definitions
│   │   └── db_collector.go                   # DB pool metrics
│   ├── middleware/
│   │   ├── prometheus.go                     # Prometheus middleware
│   │   └── request_logger.go                 # Structured logging middleware
│   ├── handler/
│   │   ├── health.go                         # Health + readiness checks
│   │   └── metrics.go                        # /metrics endpoint
│   └── logger/
│       └── logger.go                         # zerolog setup
├── monitoring/
│   ├── prometheus/
│   │   ├── prometheus.yml                    # Prometheus config
│   │   └── alerts/
│   │       └── api-alerts.yml                # Alert rules
│   ├── grafana/
│   │   ├── provisioning/
│   │   │   ├── datasources/datasources.yml   # Auto data sources
│   │   │   └── dashboards/dashboards.yml     # Auto dashboards
│   │   └── dashboards/
│   │       └── task-manager.json             # Dashboard JSON
│   ├── loki/
│   │   └── loki-config.yml                   # Loki config
│   ├── promtail/
│   │   └── promtail-config.yml               # Promtail config
│   └── alertmanager/
│       └── alertmanager.yml                  # Alertmanager config
├── docker-compose.monitoring.yml             # Full stack compose
└── Dockerfile                                # Multi-stage build
```

---

## 15.22 Checklist Bab Ini

### Konsep Monitoring

- [ ] Paham kenapa monitoring penting ("you can't fix what you can't see")
- [ ] Paham three pillars of observability: Metrics, Logs, Traces
- [ ] Paham perbedaan monitoring vs observability
- [ ] Paham analogi dashboard mobil

### Prometheus

- [ ] Paham apa itu Prometheus (time-series database, pull-based)
- [ ] Paham perbedaan pull vs push model
- [ ] Paham architecture: targets, scrape, TSDB, PromQL
- [ ] Paham perbandingan Prometheus vs Datadog/New Relic

### Metric Types

- [ ] Paham 4 tipe metric: Counter, Gauge, Histogram, Summary
- [ ] Tau kapan pakai Counter vs Gauge vs Histogram
- [ ] Paham label cardinality dan kenapa penting

### Go Instrumentation

- [ ] prometheus/client_golang ter-install
- [ ] `/metrics` endpoint expose default Go metrics
- [ ] Custom metrics defined (http_requests_total, http_request_duration, dll)
- [ ] Prometheus middleware auto-collect per request
- [ ] Business metrics instrumented (tasks_created, login_attempts)
- [ ] DB connection pool metrics exposed

### Prometheus Setup

- [ ] prometheus.yml configured dengan scrape targets
- [ ] Prometheus jalan via Docker
- [ ] Semua targets status UP di Prometheus UI
- [ ] Bisa query metrics di Prometheus UI

### PromQL

- [ ] Paham instant vector vs range vector
- [ ] Paham rate(), sum(), avg(), topk()
- [ ] Bisa query request rate per second
- [ ] Bisa query error rate percentage
- [ ] Bisa query P95 latency (histogram_quantile)
- [ ] Bisa query top endpoints

### Grafana

- [ ] Grafana jalan via Docker
- [ ] Prometheus connected sebagai data source
- [ ] Pre-built dashboards imported (Go metrics, Node exporter)
- [ ] Custom dashboard dibuat untuk Task Manager API
- [ ] Dashboard panels: RPS, error rate, latency, goroutines, memory

### Alerting

- [ ] Alert rules defined (high error rate, high latency, service down, dll)
- [ ] Alertmanager configured
- [ ] Notification channel setup (Telegram/Discord/Slack)
- [ ] Test alert berhasil dikirim
- [ ] Paham severity levels (P1-P4)

### Structured Logging

- [ ] zerolog/slog ter-install
- [ ] Logger initialized (JSON output production, pretty dev)
- [ ] Request logging middleware dengan correlation ID
- [ ] Log levels dipakai dengan benar (info, warn, error)

### Log Aggregation

- [ ] Paham ELK vs Loki (dan kenapa pilih Loki)
- [ ] Loki jalan via Docker
- [ ] Promtail configured untuk ship logs ke Loki
- [ ] Bisa query logs di Grafana (LogQL)

### Docker Compose

- [ ] docker-compose.monitoring.yml include full stack
- [ ] Semua services: API + DB + Redis + Nginx + Prometheus + Grafana + Loki + Promtail + Alertmanager + Exporters
- [ ] Grafana auto-provisioning (data sources + dashboards)
- [ ] `docker compose up -d` → everything works

### Health Checks

- [ ] `/health` endpoint (liveness — basic)
- [ ] `/ready` endpoint (readiness — cek DB + Redis)
- [ ] Paham liveness vs readiness probe (Kubernetes concept)
- [ ] Docker healthcheck configured

### Uptime Monitoring

- [ ] External monitoring setup (UptimeRobot / similar)
- [ ] Health endpoint dimonitor dari luar
- [ ] Alert configured kalau endpoint down
- [ ] Paham kenapa external monitoring perlu (selain Prometheus)

### Performance Profiling

- [ ] Paham Go pprof
- [ ] pprof endpoint tersedia (development only!)
- [ ] Bisa jalanin CPU profile
- [ ] Bisa jalanin memory profile
- [ ] Paham flame graph

### Incident Response

- [ ] Paham incident response flow (detect → triage → investigate → fix → post-mortem)
- [ ] Paham severity levels (P1-P4)
- [ ] Runbook template tersedia
- [ ] Paham on-call rotation concept

### Best Practices

- [ ] Paham USE method (Utilization, Saturation, Errors)
- [ ] Paham RED method (Rate, Errors, Duration)
- [ ] Paham label cardinality best practices
- [ ] Paham retention policy
- [ ] Dashboard terorganisir dengan baik

---

## Stage 2 Complete: DevOps Core

**Selamat!** Kamu baru aja menyelesaikan **Stage 2: DevOps Core**. Coba kita recap perjalanan dari Bab 11 sampai Bab 15:

```
Stage 2: DevOps Core — COMPLETED!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Bab 11: Docker
→ Containerize Task Manager API
→ Dockerfile multi-stage (20MB image)
→ docker-compose untuk development

Bab 12: CI/CD
→ GitHub Actions pipeline
→ Auto test → build → deploy
→ Push ke GitHub = auto deploy

Bab 13: Cloud — AWS
→ Deploy ke EC2 + RDS + ElastiCache
→ Production environment di cloud
→ Accessible dari internet

Bab 14: Nginx
→ Reverse proxy + HTTPS + SSL
→ Rate limiting + security headers
→ Load balancing + compression

Bab 15: Monitoring (BAB INI)
→ Prometheus + Grafana dashboards
→ Custom metrics + alerts
→ Structured logging + Loki
→ Health checks + incident response
→ Full monitoring stack di Docker Compose
```

Sekarang kamu punya **production-grade setup** yang lengkap:

```
User Request Journey (full stack):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

User
 │
 ▼
UptimeRobot (external monitoring)
 │
 ▼
DNS → api.taskmanager.com
 │
 ▼
Nginx (reverse proxy, SSL, rate limit, gzip)
 │
 ▼
Go Gin API (Prometheus metrics, structured logs)
 ├── PostgreSQL (RDS)
 ├── Redis (ElastiCache)
 └── /metrics → Prometheus → Grafana → Dashboard + Alerts
                                     → Telegram/Discord notification
```

**Dari developer yang cuma bisa `go run main.go`, sekarang kamu bisa:**
1. Containerize app (Docker)
2. Auto deploy (CI/CD)
3. Run di cloud (AWS)
4. Serve securely (Nginx + SSL)
5. Monitor everything (Prometheus + Grafana)
6. Get alerted when things go wrong (Alertmanager)
7. Debug with structured logs (zerolog + Loki)
8. Know app health (health checks)
9. Respond to incidents (runbooks)

Itu **DevOps Core yang solid**. Di Stage 3 (Bab 16+), kita masuk ke level advanced: Kubernetes, Terraform, Microservices, Performance Testing, dan Security. Tapi fondasi yang kamu bangun di Stage 2 ini bakal jadi dasar untuk semua itu.

> **Next up: Bab 16 — Kubernetes.** Kalau Docker itu containerize satu app, Kubernetes itu **orchestrate ratusan container** secara otomatis. Auto-scaling, self-healing, rolling deployments. Ini yang dipake Google, Netflix, Spotify untuk jalanin ribuan microservices. Dan sekarang kamu udah punya fondasi yang kuat untuk belajar itu.

---
