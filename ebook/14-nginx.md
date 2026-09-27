# Bab 14: Nginx & Reverse Proxy — Pintu Depan Aplikasi Kamu

## Kenapa Bab Ini Penting?

Di bab sebelumnya, kamu udah deploy Task Manager API ke AWS — EC2, RDS, ElastiCache, Docker, CI/CD, semuanya jalan. Tapi ada satu masalah: **user harus akses API lewat port 8080**. URL-nya kayak gini: `http://13.250.xx.xx:8080/api/v1/tasks`. Jelek. Ga profesional. Ga aman.

Di dunia nyata:
- User akses port **80** (HTTP) dan **443** (HTTPS)
- Domain yang cantik: `https://api.taskmanager.com`
- HTTPS dengan SSL certificate
- Protection dari serangan (DDoS, brute force)
- Compression biar response lebih kecil
- Caching biar server ga kerja keras

**Siapa yang handle semua itu?** Nginx.

```
Sebelum bab ini:
- User → http://13.250.xx.xx:8080 → Go app
- Ga ada HTTPS
- Ga ada compression
- Ga ada rate limiting di level server
- Ga ada security headers
- 1 server, 1 app, udah mentok

Setelah bab ini:
- User → https://api.taskmanager.com → Nginx → Go app
- HTTPS dengan Let's Encrypt (gratis!)
- Gzip compression (response 70% lebih kecil)
- Rate limiting di level Nginx
- Security headers lengkap
- Load balancing ke multiple backend (horizontal scaling)
- Static files served langsung tanpa hit Go app
```

Kalau kamu pernah deploy Laravel atau NestJS, kamu pasti udah kenal Nginx. Laravel pakai Nginx + PHP-FPM. NestJS pakai Nginx sebagai reverse proxy. **Konsepnya sama persis** — di bab ini kita pakai Nginx buat Go (Gin) API.

> **Analogi gampang:** Nginx itu **resepsionis hotel**. Tamu (user) datang ke lobby (port 80/443), resepsionis terima tamu, cek identitas (SSL, rate limit, security), terus arahin ke kamar yang tepat (backend server port 8080). Tamu ga perlu tau nomor kamar langsung — resepsionis yang handle semuanya.

---

## 14.1 Apa Itu Nginx?

### Sejarah Singkat

Nginx (dibaca "engine-x") diciptakan oleh **Igor Sysoev** tahun 2002, open-source tahun 2004. Awalnya dibuat untuk menyelesaikan **C10K problem** — masalah handling 10,000 concurrent connections.

```
C10K Problem (tahun 2000an):
━━━━━━━━━━━━━━━━━━━━━━━━━━
Apache pada waktu itu pakai model "1 thread per connection".
10,000 users = 10,000 threads = RAM habis, server mati.

Solusi Nginx:
━━━━━━━━━━━━
Event-driven, non-blocking I/O.
1 worker process bisa handle RIBUAN connections.
10,000 users? No problem. 100,000? Masih santai.
```

### 3 Peran Utama Nginx

Nginx bukan cuma satu hal — dia bisa jadi **tiga sekaligus**:

```
┌─────────────────────────────────────────────────────────┐
│                  Nginx — 3 Peran                         │
│                                                          │
│  1. WEB SERVER                                          │
│     Serve static files (HTML, CSS, JS, images)          │
│     Kayak Apache — tapi lebih cepet                     │
│                                                          │
│  2. REVERSE PROXY                                        │
│     Terima request dari client,                          │
│     forward ke backend server (Go, Node, Python)        │
│     Client ga tau backend di mana                        │
│                                                          │
│  3. LOAD BALANCER                                        │
│     Distribute traffic ke multiple backend servers       │
│     Server 1 sibuk? Kirim ke server 2                   │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Forward Proxy vs Reverse Proxy

Banyak yang bingung bedanya. Ini penting:

```
FORWARD PROXY (client-side):
━━━━━━━━━━━━━━━━━━━━━━━━━━
Client → [Forward Proxy] → Internet → Server

Contoh: VPN, corporate proxy
Fungsi: Client mau anonymous, bypass firewall
Siapa yang setup: CLIENT

REVERSE PROXY (server-side):
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Client → Internet → [Reverse Proxy] → Backend Server

Contoh: Nginx, Caddy, HAProxy
Fungsi: Server mau protect backend, load balance, SSL
Siapa yang setup: SERVER ADMIN / DEVOPS
```

```
Analogi:

Forward Proxy = AGEN RAHASIA
Kamu (client) mau beli rumah tanpa ketahuan siapa kamu.
Kamu kirim agen untuk beli atas nama dia.
Penjual ga tau siapa pembeli aslinya.

Reverse Proxy = RESEPSIONIS HOTEL
Tamu datang ke lobby, bilang mau ketemu "Room Service".
Resepsionis hubungi room service internal.
Tamu ga perlu tau room service ada di lantai berapa.
```

### Apache vs Nginx

Kalau kamu dari background Laravel, kamu mungkin pernah pakai Apache (LAMP stack). Ini perbandingannya:

| Aspek | Apache | Nginx |
|-------|--------|-------|
| **Architecture** | Process/Thread per request | Event-driven, non-blocking |
| **Concurrent connections** | Berat di 10K+ | Santai sampai 100K+ |
| **Memory usage** | Tinggi (per-thread overhead) | **Rendah** (event loop) |
| **Static files** | Bagus | **Lebih cepat** |
| **Dynamic content** | mod_php (embedded) | Proxy ke PHP-FPM (external) |
| **Config** | .htaccess (per-directory) | Centralized (nginx.conf) |
| **Market share (2025)** | ~25% | **~35%** (nomor 1) |
| **Best for** | Shared hosting, legacy | Modern apps, reverse proxy |

```
Kenapa Nginx Menang untuk DevOps:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Lebih hemat RAM — penting di cloud (bayar per resource)
2. Lebih cepat serve static files
3. Reverse proxy lebih natural (proxy_pass)
4. Config lebih clean (ga ada .htaccess mess)
5. Load balancing built-in
6. Industri standard untuk container/microservice
```

> **Fun fact:** Nginx, Docker, Kubernetes — semua bagian dari stack modern DevOps. Nginx Inc. sekarang dimiliki F5 Networks, tapi Nginx tetap open-source. Ada juga **Nginx Plus** (berbayar) dengan fitur tambahan seperti advanced load balancing dan monitoring.

---

## 14.2 Kenapa Butuh Reverse Proxy?

Go app kamu (Gin) listen di port `8080`. Kenapa ga langsung expose port 8080 ke publik?

```
Tanpa Reverse Proxy:
━━━━━━━━━━━━━━━━━━━
User → http://yourserver.com:8080 → Go App

Problems:
1. Port 8080 jelek — user expect port 80/443
2. Ga ada HTTPS — data dikirim plaintext (password, token, dll)
3. Ga ada compression — response besar, bandwidth boros
4. Ga ada security headers — vulnerable to XSS, clickjacking
5. Ga bisa load balance — 1 server, 1 app
6. Static files (images, uploads) di-serve oleh Go — waste
7. Rate limiting cuma di level Go — DDoS langsung ke app
8. Kalau Go crash, user langsung lihat error
```

```
Dengan Nginx Reverse Proxy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━
User → https://api.taskmanager.com → Nginx (port 443)
                                         │
                                         ├── SSL termination (HTTPS → HTTP)
                                         ├── Gzip compression
                                         ├── Security headers
                                         ├── Rate limiting
                                         ├── Caching
                                         ├── Static files (langsung serve)
                                         │
                                         ▼
                                    Go App (port 8080)
                                    Cuma terima HTTP biasa
                                    Fokus business logic aja
```

### 7 Alasan Pakai Reverse Proxy

| No | Alasan | Detail |
|----|--------|--------|
| 1 | **SSL Termination** | Nginx handle HTTPS. Go app tetap HTTP biasa. Lebih simple. |
| 2 | **Security** | Backend port ga exposed. User cuma lihat port 80/443. |
| 3 | **Load Balancing** | Distribute traffic ke multiple Go instances. |
| 4 | **Compression** | Gzip response sebelum kirim ke client. Response 70% lebih kecil. |
| 5 | **Caching** | Cache response yang ga sering berubah. Reduce load ke Go app. |
| 6 | **Rate Limiting** | Block abusive traffic SEBELUM sampai ke app. |
| 7 | **Static Files** | Nginx serve images/uploads langsung. Go ga perlu handle. |

### Comparison: NestJS vs Laravel vs Go + Nginx

```
Laravel Deployment:
User → Nginx → PHP-FPM → Laravel
                ↑
          Nginx WAJIB (Laravel ga bisa serve sendiri secara production)

NestJS Deployment:
User → Nginx → Node.js → NestJS (port 3000)
                ↑
          Nginx recommended (Node BISA serve sendiri, tapi Nginx lebih baik)

Go (Gin) Deployment:
User → Nginx → Go binary → Gin (port 8080)
                ↑
          Nginx recommended (Go BISA serve sendiri, tapi Nginx lebih baik)
```

Go sebenarnya punya HTTP server yang **sangat capable** — `net/http` di standard library. Bahkan bisa handle HTTPS sendiri. Tapi di production, tetap pakai Nginx karena:

1. **Separation of concerns** — Nginx handle infra stuff, Go handle business logic
2. **Battle-tested** — Nginx udah di-test jutaan production server
3. **Features** — Rate limiting, caching, compression tanpa nulis code
4. **Standard** — Tim DevOps expect Nginx di depan backend app

---

## 14.3 Install Nginx

### macOS (Homebrew)

```bash
# Install Nginx
brew install nginx

# Start Nginx
brew services start nginx

# Verify — buka browser http://localhost:8080
# Harusnya muncul "Welcome to nginx!"
curl http://localhost:8080

# Note: di macOS, Nginx default listen port 8080 (bukan 80)
# karena port < 1024 butuh sudo

# Config file location di macOS:
# /opt/homebrew/etc/nginx/nginx.conf (Apple Silicon)
# /usr/local/etc/nginx/nginx.conf (Intel Mac)

# Nginx commands via brew
brew services start nginx     # Start
brew services stop nginx      # Stop
brew services restart nginx   # Restart
```

### Ubuntu/Debian (apt)

```bash
# Update package list
sudo apt update

# Install Nginx
sudo apt install nginx -y

# Start Nginx
sudo systemctl start nginx

# Enable auto-start saat server boot
sudo systemctl enable nginx

# Check status
sudo systemctl status nginx
# Output: active (running)

# Verify — buka browser http://your-server-ip
# Harusnya muncul "Welcome to nginx!"
curl http://localhost
```

```bash
# Config file location di Ubuntu:
/etc/nginx/nginx.conf           # Main config
/etc/nginx/sites-available/     # Config files per-site
/etc/nginx/sites-enabled/       # Symlink ke sites-available (yang aktif)
/etc/nginx/conf.d/              # Additional config files
/var/log/nginx/                 # Log files
/var/www/html/                  # Default web root
```

### Docker

```bash
# Quick run
docker run -d --name nginx -p 80:80 nginx:alpine

# Verify
curl http://localhost

# Run dengan custom config
docker run -d --name nginx \
  -p 80:80 \
  -v ./nginx.conf:/etc/nginx/nginx.conf:ro \
  nginx:alpine

# Run dengan custom config + static files
docker run -d --name nginx \
  -p 80:80 \
  -v ./nginx.conf:/etc/nginx/nginx.conf:ro \
  -v ./static:/var/www/static:ro \
  nginx:alpine
```

### Verify Installation

```bash
# Check Nginx version
nginx -v
# nginx version: nginx/1.27.x

# Check full version + compile options
nginx -V
# Ini bakal nunjukin semua modules yang ter-compile

# Test config syntax
sudo nginx -t
# Output (kalau OK):
# nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
# nginx: configuration file /etc/nginx/nginx.conf test is successful

# Kalau error, dia bakal kasih tau line berapa yang salah
```

### Nginx Commands

```bash
# ============ SYSTEMCTL (Ubuntu/Debian) ============
sudo systemctl start nginx      # Start Nginx
sudo systemctl stop nginx       # Stop Nginx
sudo systemctl restart nginx    # Restart (downtime sebentar)
sudo systemctl reload nginx     # Reload config (zero downtime!)
sudo systemctl status nginx     # Check status

# ============ NGINX SIGNAL (Universal) ============
sudo nginx                      # Start
sudo nginx -s stop              # Stop (fast shutdown)
sudo nginx -s quit              # Stop (graceful shutdown)
sudo nginx -s reload            # Reload config
sudo nginx -s reopen            # Reopen log files

# ============ CONFIG ============
sudo nginx -t                   # Test config syntax
sudo nginx -T                   # Test + print full config
```

> **Penting:** Selalu jalanin `nginx -t` sebelum `reload`! Kalau config salah dan kamu langsung reload, Nginx bisa crash.

---

## 14.4 Nginx Architecture

### Master Process vs Worker Process

Nginx pakai model **master-worker**:

```
┌────────────────────────────────────────────────────┐
│                   Nginx Architecture                │
│                                                     │
│  ┌─────────────────┐                               │
│  │  Master Process  │  PID 1                       │
│  │                  │                               │
│  │  - Read config   │                               │
│  │  - Manage workers│                               │
│  │  - Bind ports    │                               │
│  │  - NO request    │                               │
│  │    handling!     │                               │
│  └────────┬────────┘                               │
│           │                                         │
│     ┌─────┼──────────────┐                         │
│     │     │              │                         │
│     ▼     ▼              ▼                         │
│  ┌──────┐ ┌──────┐  ┌──────┐                      │
│  │Worker│ │Worker│  │Worker│   (biasanya = CPU cores)│
│  │  1   │ │  2   │  │  3   │                        │
│  │      │ │      │  │      │                        │
│  │ 1000+│ │ 1000+│  │ 1000+│  connections each      │
│  │ conn │ │ conn │  │ conn │                        │
│  └──────┘ └──────┘  └──────┘                       │
│                                                     │
│  Total capacity: 3 workers x 1000+ = 3000+ conn    │
└────────────────────────────────────────────────────┘
```

**Master Process:**
- Jalan sebagai root (butuh akses port 80/443)
- Baca dan validasi config
- Manage worker processes (start, stop, restart)
- **Ga pernah handle request client**

**Worker Process:**
- Jalan sebagai user biasa (www-data / nginx)
- Handle semua client requests
- Event-driven — 1 worker handle ribuan connections
- Non-blocking I/O — ga ada thread yang nganggur nunggu

### Event-Driven vs Thread-Based

```
Apache (Thread-Based):
━━━━━━━━━━━━━━━━━━━━
Request 1 → Thread 1 ████████████ (nunggu DB response)
Request 2 → Thread 2 ████████████ (nunggu file I/O)
Request 3 → Thread 3 ████████████ (nunggu external API)
Request 4 → Thread 4 ████████████
...
Request 10000 → Thread 10000 → OUT OF MEMORY!

Setiap thread = ~2MB RAM.
10,000 threads = 20GB RAM cuma buat threads.

Nginx (Event-Driven):
━━━━━━━━━━━━━━━━━━━━
Worker 1:
  Event loop:
    Request 1 → kirim ke DB → NEXT! (ga nunggu)
    Request 2 → kirim ke file → NEXT!
    Request 3 → kirim ke API → NEXT!
    DB response for Request 1 → process → send response
    File ready for Request 2 → process → send response
    ...handles 1000+ requests dengan 1 process

Satu worker process = ~10MB RAM.
4 workers = 40MB RAM, handle 4000+ connections.
```

Ini konsep yang sama kayak **event loop di Node.js**. Kalau kamu udah paham Node.js/NestJS — kamu udah paham kenapa Nginx bisa handle banyak connections.

### Config File Structure

```
/etc/nginx/
├── nginx.conf              ← Main config file
├── mime.types              ← MIME type mappings
├── conf.d/                 ← Additional configs (*.conf auto-loaded)
│   └── default.conf
├── sites-available/        ← Config per-site (available tapi belum aktif)
│   └── taskmanager.conf
├── sites-enabled/          ← Symlink ke sites-available (yang aktif)
│   └── taskmanager.conf → ../sites-available/taskmanager.conf
├── snippets/               ← Reusable config snippets
│   └── ssl-params.conf
└── modules-enabled/        ← Dynamic modules
```

> **Note:** `sites-available` / `sites-enabled` itu pattern dari Debian/Ubuntu. Di CentOS/RHEL atau Docker, biasanya cuma pakai `conf.d/`. Dua-duanya valid.

---

## 14.5 Nginx Config Basics

### Anatomy of nginx.conf

```nginx
# ============================================
# nginx.conf — Minimal dengan penjelasan
# ============================================

# ---- MAIN CONTEXT (global settings) ----
# Ini setting yang berlaku untuk seluruh Nginx

worker_processes auto;          # Jumlah worker = jumlah CPU cores
error_log /var/log/nginx/error.log warn;  # Error log location
pid /run/nginx.pid;             # File untuk simpan PID master process

# ---- EVENTS CONTEXT ----
# Setting untuk connection handling

events {
    worker_connections 1024;    # Max connections per worker
    # Total max connections = worker_processes x worker_connections
    # auto (4 cores) x 1024 = 4096 concurrent connections
}

# ---- HTTP CONTEXT ----
# Setting untuk HTTP server (semua web-related config di sini)

http {
    # MIME types — supaya browser tau .css itu CSS, .js itu JS
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging format
    log_format main '$remote_addr - $remote_user [$time_local] '
                    '"$request" $status $body_bytes_sent '
                    '"$http_referer" "$http_user_agent"';
    access_log /var/log/nginx/access.log main;

    # Performance
    sendfile on;            # Efisien kirim file (bypass user space)
    keepalive_timeout 65;   # Keep connection alive 65 detik

    # ---- SERVER BLOCK ----
    # Ini = virtual host. Satu server block = satu website/app

    server {
        listen 80;                    # Listen di port 80
        server_name example.com;      # Domain name

        # ---- LOCATION BLOCK ----
        # Routing berdasarkan URL path

        location / {
            root /var/www/html;       # Serve files dari directory ini
            index index.html;         # Default file
        }

        location /api/ {
            proxy_pass http://localhost:8080;  # Forward ke backend
        }
    }
}
```

### Config Hierarchy

```
nginx.conf (main context)
│
├── events { }          ← Connection settings
│
└── http { }            ← All HTTP settings
    │
    ├── server { }      ← Virtual host 1 (api.domain.com)
    │   ├── location / { }
    │   ├── location /api/ { }
    │   └── location /static/ { }
    │
    └── server { }      ← Virtual host 2 (admin.domain.com)
        ├── location / { }
        └── location /assets/ { }
```

### Directives dan Blocks

```nginx
# DIRECTIVE = setting key-value
# Format: directive_name value;
# Selalu diakhiri semicolon (;)

worker_processes 4;                    # Simple directive
listen 80;                             # Simple directive
server_name api.taskmanager.com;       # Simple directive

# BLOCK = container yang punya { }
# Bisa nested (block di dalam block)

http {                                 # Block
    server {                           # Nested block
        location /api/ {               # Nested nested block
            proxy_pass http://localhost:8080;  # Directive di dalam block
        }
    }
}
```

### Syntax Rules

```
1. Setiap directive diakhiri SEMICOLON (;)
   ✓ listen 80;
   ✗ listen 80     ← Error!

2. Block pakai CURLY BRACES { }
   ✓ server { ... }
   ✗ server ( ... )  ← Error!

3. Comments pakai HASH (#)
   # This is a comment
   listen 80;  # This is also a comment

4. Include file lain
   include /etc/nginx/conf.d/*.conf;
   include /etc/nginx/sites-enabled/*;

5. Variables pakai DOLLAR SIGN ($)
   $remote_addr    ← IP client
   $request_uri    ← Full request URI
   $host           ← Hostname dari request
```

### Location Block Matching

Location block itu **kunci routing di Nginx**. Ini cara Nginx decide mana block yang match:

```nginx
# EXACT MATCH — paling prioritas
location = /health {
    # Cuma match /health (bukan /health/ atau /healthcheck)
    return 200 "OK";
}

# PREFIX MATCH (prioritized)
location ^~ /static/ {
    # Match URL yang START WITH /static/
    # ^~ = kalau prefix match, STOP cari regex
    root /var/www;
}

# REGEX MATCH (case-sensitive)
location ~ \.php$ {
    # Match URL yang END WITH .php
    # ~ = case-sensitive regex
}

# REGEX MATCH (case-insensitive)
location ~* \.(jpg|jpeg|png|gif)$ {
    # Match URL yang end dengan image extensions
    # ~* = case-insensitive
}

# PREFIX MATCH (normal)
location /api/ {
    # Match URL yang start with /api/
    # Kalau ada regex yang juga match, regex menang
    proxy_pass http://localhost:8080;
}

# CATCH-ALL
location / {
    # Match SEMUA URL (fallback)
    # Priority paling rendah
    root /var/www/html;
}
```

```
Priority Order (dari tinggi ke rendah):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. = /exact/path          (exact match)
2. ^~ /prefix/path        (prioritized prefix)
3. ~ regex                (case-sensitive regex)
4. ~* regex               (case-insensitive regex)
5. /prefix/path           (normal prefix)
6. /                      (catch-all)
```

---

## 14.6 Reverse Proxy Setup — Nginx di Depan Go App

Ini dia bagian utamanya. Kita setup Nginx sebagai reverse proxy ke Task Manager Go API.

### Skenario

```
User Request:
  https://api.taskmanager.com/api/v1/tasks

Flow:
  User → Nginx (port 80/443) → Go App (port 8080)

Nginx config:
  proxy_pass http://localhost:8080;
```

### Config File

Buat file config di `/etc/nginx/sites-available/taskmanager.conf`:

```nginx
# /etc/nginx/sites-available/taskmanager.conf
# Reverse proxy untuk Task Manager Go API

server {
    listen 80;
    server_name api.taskmanager.com;
    # Ganti dengan domain kamu, atau pakai IP address:
    # server_name 13.250.xx.xx;

    # ============================================
    # Reverse Proxy ke Go App
    # ============================================
    location / {
        # Forward semua request ke Go app di port 8080
        proxy_pass http://localhost:8080;

        # ---- PROXY HEADERS ----
        # Headers yang penting banget untuk reverse proxy

        # Kirim hostname asli dari client
        proxy_set_header Host $host;

        # IP asli client (bukan IP Nginx)
        proxy_set_header X-Real-IP $remote_addr;

        # Chain of proxies (kalau ada multiple proxy)
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # Protocol asli (http atau https)
        proxy_set_header X-Forwarded-Proto $scheme;

        # ---- TIMEOUTS ----
        proxy_connect_timeout 60s;   # Timeout connect ke backend
        proxy_send_timeout 60s;      # Timeout kirim request ke backend
        proxy_read_timeout 60s;      # Timeout nunggu response dari backend

        # ---- BUFFERING ----
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }
}
```

### Enable Config

```bash
# Buat symlink dari sites-available ke sites-enabled
sudo ln -s /etc/nginx/sites-available/taskmanager.conf /etc/nginx/sites-enabled/

# Hapus default config (opsional, tapi recommended)
sudo rm /etc/nginx/sites-enabled/default

# Test config
sudo nginx -t

# Reload Nginx (zero downtime)
sudo systemctl reload nginx
```

### Kenapa Proxy Headers Penting?

Tanpa proxy headers, Go app kamu ga tau info penting tentang client:

```
Tanpa proxy headers:
━━━━━━━━━━━━━━━━━━
Request masuk ke Go app:
  RemoteAddr = "127.0.0.1"    ← IP Nginx, bukan IP user!
  Host = "localhost:8080"      ← Internal hostname
  Scheme = "http"              ← Selalu http (Nginx → Go)

Go app pikir SEMUA request dari localhost.
Rate limiter ga bisa bedain user.
Log ga berguna (semua IP sama).

Dengan proxy headers:
━━━━━━━━━━━━━━━━━━━━
Request masuk ke Go app:
  X-Real-IP = "203.0.113.50"         ← IP user yang sebenarnya
  X-Forwarded-For = "203.0.113.50"   ← Chain of IPs
  X-Forwarded-Proto = "https"        ← User akses via HTTPS
  Host = "api.taskmanager.com"       ← Domain asli
```

### Update Go App untuk Baca Proxy Headers

Di Gin framework, kamu perlu bilang ke Gin bahwa dia ada di belakang proxy:

```go
// cmd/api/main.go

func main() {
    // ... existing code ...

    r := gin.Default()

    // PENTING: Set trusted proxies
    // Ini supaya Gin baca X-Real-IP dan X-Forwarded-For
    r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
    // Artinya: percaya proxy dari localhost (Nginx yang jalan di server yang sama)

    // Kalau Nginx jalan di server berbeda:
    // r.SetTrustedProxies([]string{"10.0.1.10"})  // IP Nginx server

    router.Setup(r, handlers, authService, redisClient)

    // ... existing code ...
}
```

Sekarang `c.ClientIP()` di middleware rate limiter akan return IP user yang sebenarnya, bukan IP Nginx.

### Test Reverse Proxy

```bash
# Go app harus running di port 8080
# (jalanin di terminal/screen/tmux terpisah)

# Test via Nginx (port 80)
curl http://api.taskmanager.com/health
# {"status":"ok"}

# Test API endpoint
curl http://api.taskmanager.com/api/v1/auth/login \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password123"}'

# Check headers yang diterima Go app (tambah logging kalau perlu)
curl -v http://api.taskmanager.com/api/v1/tasks
```

---

## 14.7 SSL/HTTPS dengan Let's Encrypt

### Kenapa HTTPS Wajib?

```
HTTP (tanpa SSL):
━━━━━━━━━━━━━━━
User ──[plaintext]──→ Server
Password: "mypassword123" ← BISA DIBACA siapa aja di network!
JWT Token: "eyJhbGciO..." ← BISA DICURI!

HTTPS (dengan SSL):
━━━━━━━━━━━━━━━━━━
User ──[encrypted]──→ Server
Password: "a8f3k2j5..." ← Encrypted, ga bisa dibaca
JWT Token: "9fk3m2p1..." ← Encrypted, aman
```

Sejak 2018, browser Chrome udah nandain HTTP sebagai **"Not Secure"**. Google juga prioritasin HTTPS di search ranking. **HTTPS bukan opsional lagi — wajib.**

### Let's Encrypt — SSL Gratis

Let's Encrypt itu **Certificate Authority (CA) gratis** dari Internet Security Research Group (ISRG). Certificate-nya valid 90 hari dan bisa di-renew otomatis.

```
SSL Certificate Options:
━━━━━━━━━━━━━━━━━━━━━━
Let's Encrypt    → GRATIS, auto-renewal, 90 hari
Cloudflare       → GRATIS (kalau pakai Cloudflare DNS)
AWS ACM          → GRATIS (kalau pakai AWS ALB/CloudFront)
DigiCert         → $200-1000/tahun (enterprise)
Comodo           → $50-500/tahun

Untuk project kita → Let's Encrypt (gratis + reliable)
```

### Install Certbot

Certbot itu tool resmi untuk obtain dan manage Let's Encrypt certificates.

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install certbot python3-certbot-nginx -y

# macOS (untuk testing — di production selalu Linux)
brew install certbot

# Verify
certbot --version
# certbot 2.x.x
```

### Obtain Certificate

```bash
# Pastikan:
# 1. Domain sudah pointing ke IP server kamu (DNS A record)
# 2. Port 80 accessible dari internet
# 3. Nginx running

# Obtain certificate (Certbot auto-edit Nginx config)
sudo certbot --nginx -d api.taskmanager.com

# Interactive prompts:
# 1. Enter email address → irzan@email.com
# 2. Agree to Terms of Service → Y
# 3. Share email with EFF → N (opsional)
# 4. Redirect HTTP to HTTPS → 2 (yes, redirect)

# Certbot akan:
# 1. Verify domain ownership (HTTP-01 challenge)
# 2. Download SSL certificate
# 3. Auto-edit Nginx config untuk HTTPS
# 4. Setup HTTP → HTTPS redirect
```

```bash
# Certificate files location:
/etc/letsencrypt/live/api.taskmanager.com/
├── fullchain.pem    ← Certificate + intermediate CA
├── privkey.pem      ← Private key
├── cert.pem         ← Certificate only
└── chain.pem        ← Intermediate CA only
```

### Full Nginx Config dengan SSL

Ini config yang biasanya Certbot generate (kita enhance sedikit):

```nginx
# /etc/nginx/sites-available/taskmanager.conf
# Task Manager API — HTTPS + Reverse Proxy

# ============================================
# HTTP → HTTPS Redirect
# ============================================
server {
    listen 80;
    server_name api.taskmanager.com;

    # Redirect SEMUA HTTP request ke HTTPS
    return 301 https://$host$request_uri;
    # 301 = permanent redirect
    # $host = domain dari request
    # $request_uri = path + query string
}

# ============================================
# HTTPS Server
# ============================================
server {
    listen 443 ssl http2;
    server_name api.taskmanager.com;

    # ---- SSL CERTIFICATE ----
    ssl_certificate /etc/letsencrypt/live/api.taskmanager.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.taskmanager.com/privkey.pem;

    # ---- SSL SETTINGS (Security Best Practices) ----

    # Hanya allow TLS 1.2 dan 1.3 (disable TLS 1.0, 1.1 — udah ga aman)
    ssl_protocols TLSv1.2 TLSv1.3;

    # Cipher suites yang aman
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # SSL session caching (performance)
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    # OCSP Stapling — verify certificate faster
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;

    # HSTS — paksa browser SELALU pakai HTTPS
    # max-age=63072000 = 2 tahun
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;

    # ---- REVERSE PROXY ----
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

### Auto-Renewal

Let's Encrypt certificates expire setiap **90 hari**. Certbot auto-install cron/timer untuk renewal:

```bash
# Check auto-renewal timer (systemd)
sudo systemctl list-timers | grep certbot
# certbot.timer — biasanya run 2x sehari

# Atau check crontab
sudo crontab -l | grep certbot

# Test renewal (dry run — ga beneran renew)
sudo certbot renew --dry-run

# Manual renew (kalau perlu)
sudo certbot renew

# Force renew (kalau belum expire tapi mau renew)
sudo certbot renew --force-renewal
```

Kalau auto-renewal belum ke-setup, tambahin sendiri:

```bash
# Tambah cron job untuk renewal
sudo crontab -e

# Tambah line ini (cek renewal 2x sehari, jam 2 dan jam 14):
0 2,14 * * * certbot renew --quiet --post-hook "systemctl reload nginx"

# --quiet = ga print output kalau sukses
# --post-hook = reload Nginx setelah renew
```

### SSL Test

```bash
# Test SSL grade (harus A atau A+)
# Buka: https://www.ssllabs.com/ssltest/analyze.html?d=api.taskmanager.com

# Atau pakai command line:
curl -vI https://api.taskmanager.com 2>&1 | grep -E "(SSL|TLS|certificate)"
```

---

## 14.8 Load Balancing

Satu Go app di satu server — kalau traffic naik, server bisa kewalahan. Solusinya: **jalanin multiple instances** dan distribusikan traffic pakai Nginx load balancer.

### Upstream Block

```nginx
# ============================================
# Load Balancing — Multiple Go Instances
# ============================================

# Definisikan group of backend servers
upstream task_manager_backend {
    server 127.0.0.1:8080;    # Instance 1
    server 127.0.0.1:8081;    # Instance 2
    server 127.0.0.1:8082;    # Instance 3
}

server {
    listen 80;
    server_name api.taskmanager.com;

    location / {
        # proxy_pass ke upstream group (bukan ke satu server)
        proxy_pass http://task_manager_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Load Balancing Algorithms

```nginx
# ============================================
# 1. ROUND ROBIN (default)
# ============================================
# Request didistribusikan secara bergiliran
# Request 1 → Server 1
# Request 2 → Server 2
# Request 3 → Server 3
# Request 4 → Server 1 (balik lagi)

upstream backend {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
    # Ga perlu directive tambahan — round robin itu default
}

# ============================================
# 2. LEAST CONNECTIONS
# ============================================
# Request dikirim ke server yang paling sedikit active connections
# Bagus kalau request processing time-nya beda-beda

upstream backend {
    least_conn;
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

# ============================================
# 3. IP HASH
# ============================================
# Request dari IP yang sama SELALU ke server yang sama
# Bagus kalau app punya session/state (sticky sessions)
# Task Manager kita pakai JWT (stateless), jadi ga perlu ini

upstream backend {
    ip_hash;
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

# ============================================
# 4. WEIGHTED
# ============================================
# Server dengan weight lebih tinggi dapat lebih banyak traffic
# Berguna kalau server punya spec berbeda

upstream backend {
    server 127.0.0.1:8080 weight=5;   # 5x traffic
    server 127.0.0.1:8081 weight=3;   # 3x traffic
    server 127.0.0.1:8082 weight=1;   # 1x traffic
    # Dari 9 request: 5 ke :8080, 3 ke :8081, 1 ke :8082
}
```

### Health Check dan Failover

```nginx
upstream task_manager_backend {
    server 127.0.0.1:8080 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8081 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8082 max_fails=3 fail_timeout=30s backup;

    # max_fails=3 → kalau gagal 3x, tandai sebagai DOWN
    # fail_timeout=30s → coba lagi setelah 30 detik
    # backup → server ini cuma dipake kalau semua server utama DOWN
}
```

```
Flow:
━━━━━
1. Nginx kirim request ke Server 1
2. Server 1 ga response (timeout/error)
3. Nginx coba lagi → gagal lagi → gagal 3x
4. Nginx tandai Server 1 sebagai DOWN
5. Semua traffic ke Server 2 dan 3
6. Setelah 30 detik, Nginx coba Server 1 lagi
7. Kalau Server 1 udah recover → masuk rotation lagi
8. Kalau Server 1 & 2 DOWN → traffic ke Server 3 (backup)
```

### Horizontal Scaling dengan Docker

Cara paling gampang untuk jalanin multiple instances:

```bash
# Jalanin 3 Go app instances di port berbeda
PORT=8080 ./task-manager-api &
PORT=8081 ./task-manager-api &
PORT=8082 ./task-manager-api &

# Atau pakai Docker
docker run -d -p 8080:8080 -e PORT=8080 task-manager-api
docker run -d -p 8081:8080 -e PORT=8080 task-manager-api
docker run -d -p 8082:8080 -e PORT=8080 task-manager-api
```

### Rekomendasi untuk Task Manager API

```
Task Manager API pakai JWT (stateless) dan Redis untuk session.
Semua instance share database (PostgreSQL) dan cache (Redis).

Rekomendasi:
- Algorithm: least_conn (karena beberapa endpoint lebih berat dari lain)
- Ga perlu ip_hash (stateless, ga ada server-side session)
- Set max_fails dan fail_timeout untuk auto-failover
- Punya minimal 1 backup server
```

---

## 14.9 Caching

Nginx bisa cache response dari backend. Artinya request yang sama ga perlu hit Go app lagi — Nginx langsung return response dari cache.

### Kapan Cache API Response?

```
CACHE:
- GET /api/v1/boards (list boards — jarang berubah)
- GET /api/v1/tasks (list tasks — bisa cache 30 detik)
- GET /health (health check — always same)
- Static assets (images, CSS, JS — cache lama)

JANGAN CACHE:
- POST/PUT/DELETE (mutating operations)
- GET /api/v1/auth/me (user-specific, berubah setelah login)
- Endpoints dengan authentication (response beda per user!)
- Real-time data yang harus selalu fresh
```

> **Penting:** Untuk Task Manager API kita, hampir semua endpoint perlu authentication dan response-nya **per-user**. Jadi caching di level Nginx **kurang cocok** untuk API endpoint. Caching lebih cocok di level **Redis** (yang udah kita implement di Bab 8). Tapi tetap penting tau caranya untuk use case lain (serve static content, public API).

### Setup Proxy Cache

```nginx
# Di http block (bukan di server block)
http {
    # ============================================
    # Cache Zone Definition
    # ============================================
    proxy_cache_path /var/cache/nginx
        levels=1:2                   # Directory structure (1 char/2 char subdirs)
        keys_zone=task_cache:10m     # Cache name dan shared memory size (10MB metadata)
        max_size=1g                  # Max disk space untuk cache (1GB)
        inactive=60m                 # Hapus cache yang ga di-akses selama 60 menit
        use_temp_path=off;           # Tulis langsung ke cache dir (ga pakai temp)

    server {
        listen 80;
        server_name api.taskmanager.com;

        # ============================================
        # Cached Location (public endpoints only)
        # ============================================
        location /health {
            proxy_pass http://localhost:8080;
            proxy_cache task_cache;
            proxy_cache_valid 200 5m;         # Cache 200 OK selama 5 menit
            proxy_cache_valid 404 1m;         # Cache 404 selama 1 menit
            proxy_cache_key "$request_uri";   # Cache key = URL

            # Header untuk debug — apakah response dari cache?
            add_header X-Cache-Status $upstream_cache_status;
            # MISS = ga ada di cache, request ke backend
            # HIT = response dari cache
            # EXPIRED = cache expired, request ke backend
            # BYPASS = cache di-bypass (lihat di bawah)
        }

        # ============================================
        # Non-Cached Location (API endpoints)
        # ============================================
        location /api/ {
            proxy_pass http://localhost:8080;
            proxy_no_cache 1;                 # Jangan cache
            proxy_cache_bypass 1;             # Bypass cache
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

### Cache Bypass

```nginx
# Bypass cache berdasarkan kondisi
location /api/v1/public/ {
    proxy_pass http://localhost:8080;
    proxy_cache task_cache;
    proxy_cache_valid 200 10m;

    # Bypass cache kalau ada cookie "nocache" atau query param "nocache"
    proxy_cache_bypass $cookie_nocache $arg_nocache;

    # Bypass cache untuk request dengan Authorization header
    proxy_no_cache $http_authorization;
    proxy_cache_bypass $http_authorization;
}
```

### Verify Cache

```bash
# Request pertama (MISS)
curl -I http://api.taskmanager.com/health
# X-Cache-Status: MISS

# Request kedua (HIT — dari cache!)
curl -I http://api.taskmanager.com/health
# X-Cache-Status: HIT

# Lihat cache files
ls -la /var/cache/nginx/

# Clear cache manual
sudo rm -rf /var/cache/nginx/*
sudo nginx -s reload
```

---

## 14.10 Gzip Compression

Gzip compression bisa mengurangi ukuran response **sampai 70-90%**. Artinya website load lebih cepat dan bandwidth lebih hemat.

### Setup Gzip

```nginx
http {
    # ============================================
    # Gzip Compression
    # ============================================

    gzip on;                  # Enable gzip
    gzip_vary on;             # Add "Vary: Accept-Encoding" header
    gzip_proxied any;         # Compress response dari proxy juga
    gzip_comp_level 6;        # Compression level (1-9, 6 = good balance)
    gzip_min_length 256;      # Jangan compress kalau < 256 bytes (ga worth it)

    # MIME types yang di-compress
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml
        application/xml+rss
        application/vnd.ms-fontobject
        font/ttf
        font/opentype
        image/svg+xml;

    # JANGAN compress file yang udah compressed:
    # - Images (JPEG, PNG, GIF) — udah compressed
    # - Videos (MP4, WebM) — udah compressed
    # - ZIP, GZIP files — udah compressed
}
```

### Before vs After

```
Tanpa Gzip:
━━━━━━━━━━━
GET /api/v1/tasks → Response: 15,000 bytes (15KB)
GET /api/v1/boards → Response: 8,500 bytes (8.5KB)
100 requests → 1.5MB + 850KB = 2.35MB bandwidth

Dengan Gzip:
━━━━━━━━━━━━
GET /api/v1/tasks → Response: 3,200 bytes (3.2KB) — 79% smaller!
GET /api/v1/boards → Response: 1,800 bytes (1.8KB) — 79% smaller!
100 requests → 320KB + 180KB = 500KB bandwidth

Hemat: 2.35MB - 0.5MB = 1.85MB per 100 requests
Di scale besar (1M requests/hari): hemat ~18.5GB bandwidth/hari
```

### Compression Level Trade-off

```
Level 1:  Compression rendah, CPU rendah   — cepat tapi file masih besar
Level 6:  Compression medium, CPU medium   — sweet spot (recommended)
Level 9:  Compression tinggi, CPU tinggi   — file kecil tapi CPU kerja keras

Untuk API JSON responses:
- Level 6 udah cukup
- Bedanya level 6 vs 9 cuma ~2-5% lebih kecil
- Tapi CPU usage level 9 bisa 2-3x lipat
```

### Verify Gzip

```bash
# Request tanpa gzip (default curl ga kirim Accept-Encoding)
curl -s -o /dev/null -w "Size: %{size_download} bytes\n" \
  http://api.taskmanager.com/api/v1/tasks

# Request dengan gzip
curl -s -o /dev/null -w "Size: %{size_download} bytes\n" \
  -H "Accept-Encoding: gzip" \
  --compressed \
  http://api.taskmanager.com/api/v1/tasks

# Check response headers
curl -I -H "Accept-Encoding: gzip" http://api.taskmanager.com/api/v1/tasks
# Harus ada: Content-Encoding: gzip
```

---

## 14.11 Rate Limiting di Nginx

### Nginx Rate Limiting vs Redis Rate Limiting

Di Bab 8, kita udah implement rate limiting pakai Redis di Go app:

```go
// internal/middleware/ratelimit.go
func RateLimiter(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    // ... 100 requests per minute per IP via Redis
}
```

Kenapa perlu rate limiting **lagi** di Nginx?

```
Rate Limiting di DUA Level:
━━━━━━━━━━━━━━━━━━━━━━━━━━

Level 1: NGINX (pintu depan)
├── Block excessive traffic SEBELUM hit Go app
├── Protect dari DDoS / brute force
├── Murah (ga perlu Redis, ga perlu Go processing)
├── Limit: 10 requests/second per IP
└── Kalau kena limit → 429 langsung dari Nginx

Level 2: GO APP + REDIS (pintu dalam)
├── Rate limiting per-user (bukan per-IP)
├── Granular — beda limit per endpoint
├── 100 requests/minute per user (setelah login)
├── Perlu authentication dulu baru bisa check user
└── Business logic rate limiting

Analogy:
Nginx rate limit = satpam di gerbang kompleks
Redis rate limit = access card di pintu apartemen

Satpam block orang mencurigakan yang terus-terusan masuk.
Access card limit berapa kali kamu bisa masuk per jam.
Dua-duanya perlu — layered security.
```

### Setup Rate Limiting

```nginx
http {
    # ============================================
    # Rate Limit Zone Definition
    # ============================================

    # Zone 1: General API rate limit
    # $binary_remote_addr = IP address client (binary format, hemat memory)
    # zone=api_limit:10m = nama zone "api_limit", 10MB shared memory
    #   10MB = bisa track ~160,000 unique IPs
    # rate=10r/s = 10 requests per second per IP
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

    # Zone 2: Login endpoint (lebih ketat — prevent brute force)
    limit_req_zone $binary_remote_addr zone=login_limit:5m rate=5r/m;
    # 5r/m = 5 requests per MENIT (login attempt)

    # Zone 3: Register endpoint (lebih ketat lagi)
    limit_req_zone $binary_remote_addr zone=register_limit:5m rate=2r/m;
    # 2r/m = 2 requests per menit (register attempt)

    # Custom response untuk rate limit
    limit_req_status 429;

    server {
        listen 80;
        server_name api.taskmanager.com;

        # ============================================
        # Rate Limited Locations
        # ============================================

        # General API — 10 req/s with burst of 20
        location /api/ {
            limit_req zone=api_limit burst=20 nodelay;
            # burst=20 → buffer 20 request lebih dari limit
            # nodelay → proses burst requests langsung (ga delay/queue)

            proxy_pass http://localhost:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Login — 5 req/menit (prevent brute force)
        location /api/v1/auth/login {
            limit_req zone=login_limit burst=5 nodelay;

            proxy_pass http://localhost:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Register — 2 req/menit (prevent spam)
        location /api/v1/auth/register {
            limit_req zone=register_limit burst=2 nodelay;

            proxy_pass http://localhost:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

### Burst dan Nodelay Explained

```
rate=10r/s tanpa burst:
━━━━━━━━━━━━━━━━━━━━━━
Detik 1: 10 requests → OK
Detik 1: Request ke-11 → 429 REJECTED!

rate=10r/s burst=20:
━━━━━━━━━━━━━━━━━━━━
Detik 1: 10 requests → OK (processed immediately)
Detik 1: Request 11-30 → di-QUEUE (burst buffer)
Detik 1: Request 31+ → 429 REJECTED!
Burst requests diproses pelan-pelan (1 per 100ms = rate 10/s)

rate=10r/s burst=20 nodelay:
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Detik 1: 30 requests → OK (semua diproses langsung!)
Detik 1: Request 31+ → 429 REJECTED!
"nodelay" = burst requests ga di-queue, langsung diproses
Tapi burst "slot" yang udah dipake butuh waktu refill
```

### Test Rate Limiting

```bash
# Kirim banyak request cepat (pakai ab — Apache Benchmark)
ab -n 50 -c 10 http://api.taskmanager.com/api/v1/tasks

# Atau pakai curl loop
for i in $(seq 1 50); do
    curl -s -o /dev/null -w "%{http_code}\n" \
      http://api.taskmanager.com/api/v1/auth/login \
      -X POST -d '{"email":"test@test.com","password":"wrong"}'
done

# Output:
# 200
# 200
# 200
# 200
# 200
# 429  ← Rate limited!
# 429
# 429
# ...

# Check Nginx error log untuk rate limit info
sudo tail -f /var/log/nginx/error.log
# [error] limiting requests, excess: 5.500 by zone "login_limit"
```

---

## 14.12 Security Headers

Security headers itu HTTP response headers yang instruksikan browser untuk **protect user** dari serangan umum (XSS, clickjacking, MIME sniffing, dll).

### Full Security Headers Config

```nginx
server {
    listen 443 ssl http2;
    server_name api.taskmanager.com;

    # ... SSL config ...

    # ============================================
    # Security Headers
    # ============================================

    # 1. X-Content-Type-Options
    # Cegah browser "sniffing" MIME type
    # Tanpa ini: browser bisa interpret .txt sebagai .js dan execute script
    add_header X-Content-Type-Options "nosniff" always;

    # 2. X-Frame-Options
    # Cegah website kamu di-embed dalam iframe (anti clickjacking)
    # DENY = ga boleh di-iframe sama sekali
    # SAMEORIGIN = boleh di-iframe dari domain yang sama
    add_header X-Frame-Options "DENY" always;

    # 3. X-XSS-Protection
    # Enable browser built-in XSS filter (legacy, tapi masih berguna)
    add_header X-XSS-Protection "1; mode=block" always;

    # 4. Referrer-Policy
    # Control info apa yang dikirim di Referer header
    # strict-origin-when-cross-origin = kirim origin saja kalau cross-origin
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # 5. Content-Security-Policy (CSP)
    # Untuk API backend, CSP sederhana aja
    # default-src 'none' = ga boleh load resource apapun (API ga perlu)
    add_header Content-Security-Policy "default-src 'none'; frame-ancestors 'none'" always;

    # 6. Permissions-Policy (dulunya Feature-Policy)
    # Disable fitur browser yang ga diperlukan API
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;

    # 7. HSTS (sudah di-cover di SSL section)
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;

    # 8. Hide Nginx version (security through obscurity)
    server_tokens off;
    # Tanpa ini: response header "Server: nginx/1.27.3"
    # Dengan ini: response header "Server: nginx"

    # ... location blocks ...
}
```

### Penjelasan "always"

```nginx
add_header X-Frame-Options "DENY" always;
#                                  ^^^^^^
# "always" = tambah header untuk SEMUA response codes
# Tanpa "always" = cuma untuk 200, 201, 204, 206, 301, 302, 303, 304, 307, 308
# Error pages (4xx, 5xx) ga dapat header tanpa "always"
```

### Verify Security Headers

```bash
# Check semua response headers
curl -I https://api.taskmanager.com/health

# Output:
# HTTP/2 200
# server: nginx
# x-content-type-options: nosniff
# x-frame-options: DENY
# x-xss-protection: 1; mode=block
# referrer-policy: strict-origin-when-cross-origin
# content-security-policy: default-src 'none'; frame-ancestors 'none'
# permissions-policy: camera=(), microphone=(), geolocation=()
# strict-transport-security: max-age=63072000; includeSubDomains; preload

# Online tool: https://securityheaders.com
# Masukkan domain kamu → harus dapat grade A atau A+
```

---

## 14.13 Logging

### Access Log dan Error Log

```nginx
http {
    # ============================================
    # Custom Log Format
    # ============================================
    log_format main '$remote_addr - $remote_user [$time_local] '
                    '"$request" $status $body_bytes_sent '
                    '"$http_referer" "$http_user_agent"';

    # Log format yang lebih detail (include response time)
    log_format detailed '$remote_addr - $remote_user [$time_local] '
                        '"$request" $status $body_bytes_sent '
                        '"$http_referer" "$http_user_agent" '
                        'rt=$request_time '           # Total request time
                        'uct=$upstream_connect_time '  # Time to connect ke backend
                        'uht=$upstream_header_time '   # Time sampai first header
                        'urt=$upstream_response_time '  # Time sampai full response
                        'cs=$upstream_cache_status';    # Cache status

    # Log format JSON (easier to parse, buat ELK/Grafana)
    log_format json_log escape=json
        '{'
            '"time": "$time_iso8601", '
            '"remote_addr": "$remote_addr", '
            '"request_method": "$request_method", '
            '"request_uri": "$request_uri", '
            '"status": $status, '
            '"body_bytes_sent": $body_bytes_sent, '
            '"request_time": $request_time, '
            '"upstream_response_time": "$upstream_response_time", '
            '"http_user_agent": "$http_user_agent", '
            '"http_referer": "$http_referer"'
        '}';

    server {
        # Access log — setiap request yang masuk
        access_log /var/log/nginx/taskmanager_access.log detailed;

        # Error log — errors dan warnings
        error_log /var/log/nginx/taskmanager_error.log warn;

        # Log levels (dari paling detail ke paling sedikit):
        # debug → info → notice → warn → error → crit → alert → emerg
        # Production biasanya pakai "warn" atau "error"

        # Disable access log untuk health check (terlalu noisy)
        location /health {
            access_log off;
            proxy_pass http://localhost:8080;
        }

        location / {
            proxy_pass http://localhost:8080;
            # ... proxy headers ...
        }
    }
}
```

### Log Rotation

Nginx log files bisa membesar dengan cepat. Log rotation otomatis pakai `logrotate`:

```bash
# /etc/logrotate.d/nginx (biasanya udah ada setelah install)
# Kalau belum ada, bikin manual:

sudo tee /etc/logrotate.d/nginx << 'EOF'
/var/log/nginx/*.log {
    daily           # Rotate setiap hari
    missingok       # Ga error kalau file ga ada
    rotate 14       # Simpan 14 file rotated (2 minggu)
    compress        # Compress file lama (.gz)
    delaycompress   # Delay compress 1 rotation
    notifempty      # Jangan rotate kalau file kosong
    create 0640 www-data adm    # Permission file baru
    sharedscripts
    postrotate
        # Reopen log files setelah rotation
        if [ -f /run/nginx.pid ]; then
            kill -USR1 $(cat /run/nginx.pid)
        fi
    endscript
}
EOF
```

### Useful Log Commands

```bash
# Lihat real-time access log
sudo tail -f /var/log/nginx/taskmanager_access.log

# Lihat error log (debugging)
sudo tail -f /var/log/nginx/taskmanager_error.log

# Cari request yang lambat (> 1 detik)
grep 'rt=[1-9]' /var/log/nginx/taskmanager_access.log

# Count status codes
awk '{print $9}' /var/log/nginx/taskmanager_access.log | sort | uniq -c | sort -rn
# Output:
# 5432 200
#  321 304
#   45 404
#   12 429
#    3 502

# Top 10 IP yang paling banyak request (detect potential DDoS)
awk '{print $1}' /var/log/nginx/taskmanager_access.log | sort | uniq -c | sort -rn | head -10

# Top 10 URL yang paling banyak di-akses
awk '{print $7}' /var/log/nginx/taskmanager_access.log | sort | uniq -c | sort -rn | head -10

# Filter by time range (request hari ini jam 14)
grep "01/May/2026:14" /var/log/nginx/taskmanager_access.log
```

---

## 14.14 Static Files

Nginx itu **sangat cepat** untuk serve static files. Jauh lebih cepat daripada Go app serve file via `http.FileServer`.

### Setup Static Files

```nginx
server {
    listen 80;
    server_name api.taskmanager.com;

    # ============================================
    # Static Files — Serve langsung dari Nginx
    # ============================================

    # Uploads (user-uploaded files)
    location /uploads/ {
        alias /var/www/taskmanager/uploads/;

        # Cache static files di browser (7 hari)
        expires 7d;
        add_header Cache-Control "public, immutable";

        # Disable access log untuk static files (too noisy)
        access_log off;
    }

    # Static assets (images, icons, dll)
    location /static/ {
        alias /var/www/taskmanager/static/;
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # Favicon
    location = /favicon.ico {
        alias /var/www/taskmanager/static/favicon.ico;
        access_log off;
        log_not_found off;    # Jangan log 404 kalau favicon ga ada
    }

    # ============================================
    # API — Forward ke Go App
    # ============================================
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### root vs alias

```nginx
# ROOT — append location path ke root
location /static/ {
    root /var/www/taskmanager;
    # Request: /static/logo.png
    # File: /var/www/taskmanager/static/logo.png
    #        ^^^^^^^^^^^^^^^^^^^^^ root + /static/logo.png
}

# ALIAS — replace location path dengan alias
location /static/ {
    alias /var/www/taskmanager/assets/;
    # Request: /static/logo.png
    # File: /var/www/taskmanager/assets/logo.png
    #        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ alias + logo.png (tanpa /static/)
}

# Kapan pakai mana?
# root = kalau directory structure match URL path
# alias = kalau directory structure BEDA dari URL path
```

### Kenapa Ini Penting?

```
Tanpa Nginx static files:
━━━━━━━━━━━━━━━━━━━━━━━━
User → GET /uploads/avatar.jpg → Nginx → Go App → Read file → Response
Setiap request = Go app process + file I/O
Go worker busy serving files, ga bisa handle API requests

Dengan Nginx static files:
━━━━━━━━━━━━━━━━━━━━━━━━━
User → GET /uploads/avatar.jpg → Nginx → Read file → Response (langsung!)
Go app ga tau ada request ini.
Go app 100% fokus ke business logic.
Nginx bisa serve ribuan static files per detik.
```

---

## 14.15 WebSocket Proxy

Task Manager API mungkin butuh real-time features nanti — notifications, live updates, chat. WebSocket itu protocol yang memungkinkan **two-way communication** antara client dan server.

### Problem: WebSocket lewat Nginx

WebSocket butuh **connection upgrade** dari HTTP ke WebSocket. By default, Nginx **ga forward** Upgrade headers — jadi WebSocket connection bakal gagal.

### Config WebSocket

```nginx
server {
    listen 80;
    server_name api.taskmanager.com;

    # ============================================
    # WebSocket Proxy
    # ============================================
    location /ws/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;

        # KUNCI WebSocket: Upgrade headers
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket timeout — lebih lama dari HTTP biasa
        # Kalau ga ada activity dalam 3600s (1 jam), connection di-close
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # Regular HTTP API
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Penjelasan WebSocket Headers

```
HTTP/1.1 Upgrade Request:
━━━━━━━━━━━━━━━━━━━━━━━━
GET /ws/ HTTP/1.1
Host: api.taskmanager.com
Upgrade: websocket              ← "Gue mau upgrade ke WebSocket"
Connection: upgrade             ← "Upgrade-in connection ini"
Sec-WebSocket-Key: dGhlIHN...  ← Random key untuk handshake
Sec-WebSocket-Version: 13

HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: upgrade
Sec-WebSocket-Accept: s3pPL...

Connection upgraded! Now it's WebSocket.
Two-way communication: client ↔ server
```

```
Kenapa proxy_http_version 1.1?
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
HTTP/1.0 ga support persistent connections.
WebSocket BUTUH persistent connection.
Jadi kita HARUS set HTTP/1.1 untuk WebSocket proxy.
```

### WebSocket Handler di Go (Preview)

```go
// Contoh WebSocket handler di Gin (pakai gorilla/websocket)
// Ini preview — detail implementation di bab yang lebih advanced

import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins (tighten di production)
    },
}

func WebSocketHandler(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        // Echo message back (contoh sederhana)
        conn.WriteMessage(messageType, message)
    }
}

// Di router:
// r.GET("/ws/", WebSocketHandler)
```

---

## 14.16 Docker + Nginx

Di production, semua service jalan di Docker containers. Ini setup lengkap: Nginx + Go App + PostgreSQL + Redis dalam docker-compose.

### Nginx Dockerfile

```dockerfile
# Dockerfile.nginx
# Custom Nginx image dengan config kita

FROM nginx:1.27-alpine

# Hapus default config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom config
COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/conf.d/ /etc/nginx/conf.d/

# (Opsional) Copy static files
# COPY static/ /var/www/taskmanager/static/

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

EXPOSE 80 443
```

### Nginx Config untuk Docker

```nginx
# nginx/conf.d/taskmanager.conf
# Nginx config for Docker environment

# ============================================
# Upstream — Go App
# ============================================
upstream task_api {
    server api:8080;
    # "api" = service name di docker-compose
    # Docker DNS resolve "api" ke container IP
}

# ============================================
# HTTP → HTTPS Redirect (production)
# Uncomment ini kalau pakai SSL
# ============================================
# server {
#     listen 80;
#     server_name api.taskmanager.com;
#     return 301 https://$host$request_uri;
# }

# ============================================
# Main Server
# ============================================
server {
    listen 80;
    # Untuk production dengan SSL, ganti ke:
    # listen 443 ssl http2;
    server_name api.taskmanager.com;

    # ---- Security ----
    server_tokens off;

    # ---- Security Headers ----
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'none'; frame-ancestors 'none'" always;
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;

    # ---- Gzip ----
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_min_length 256;
    gzip_types text/plain text/css text/xml text/javascript
               application/json application/javascript application/xml
               application/xml+rss image/svg+xml;

    # ---- Static Files ----
    location /static/ {
        alias /var/www/taskmanager/static/;
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # ---- Health Check (for Docker/LB) ----
    location /health {
        proxy_pass http://task_api;
        access_log off;
    }

    # ---- API Reverse Proxy ----
    location / {
        proxy_pass http://task_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # ---- WebSocket (kalau perlu) ----
    location /ws/ {
        proxy_pass http://task_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # ---- Error Pages ----
    error_page 502 /502.html;
    location = /502.html {
        root /usr/share/nginx/html;
        internal;
    }
}
```

### Nginx Main Config

```nginx
# nginx/nginx.conf
# Main Nginx configuration

worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # ---- Logging ----
    log_format main '$remote_addr - $remote_user [$time_local] '
                    '"$request" $status $body_bytes_sent '
                    '"$http_referer" "$http_user_agent" '
                    'rt=$request_time uct=$upstream_connect_time '
                    'uht=$upstream_header_time urt=$upstream_response_time';

    access_log /var/log/nginx/access.log main;

    # ---- Performance ----
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 10m;    # Max upload size 10MB

    # ---- Rate Limiting ----
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=login_limit:5m rate=5r/m;
    limit_req_status 429;

    # ---- Include site configs ----
    include /etc/nginx/conf.d/*.conf;
}
```

### Full docker-compose.yml dengan Nginx

```yaml
# docker-compose.yml
# Task Manager API — Full Stack with Nginx
# Usage: docker compose up -d

services:
  # ============================================
  # Nginx — Reverse Proxy / Load Balancer
  # ============================================
  nginx:
    build:
      context: .
      dockerfile: Dockerfile.nginx
    container_name: task-manager-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      # Nginx config (kalau mau edit tanpa rebuild)
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      # SSL certificates (kalau pakai Let's Encrypt)
      # - /etc/letsencrypt:/etc/letsencrypt:ro
      # Static files
      - ./static:/var/www/taskmanager/static:ro
      # Logs
      - nginx_logs:/var/log/nginx
    depends_on:
      api:
        condition: service_healthy
    networks:
      - task-network
    restart: unless-stopped

  # ============================================
  # Go API — Task Manager
  # ============================================
  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: task-manager-api
    # Port TIDAK di-expose ke host!
    # Cuma accessible dari dalam Docker network (via Nginx)
    expose:
      - "8080"
    environment:
      - APP_ENV=production
      - PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD:-supersecret}
      - DB_NAME=task_manager
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD:-}
      - JWT_SECRET=${JWT_SECRET:-change-this-in-production}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - task-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s

  # ============================================
  # PostgreSQL — Database
  # ============================================
  postgres:
    image: postgres:16-alpine
    container_name: task-manager-db
    # Port TIDAK di-expose ke host (security)
    expose:
      - "5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD:-supersecret}
      - POSTGRES_DB=task_manager
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - task-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  # ============================================
  # Redis — Cache & Rate Limiting
  # ============================================
  redis:
    image: redis:7-alpine
    container_name: task-manager-redis
    command: redis-server --appendonly yes
    expose:
      - "6379"
    volumes:
      - redis_data:/data
    networks:
      - task-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

# ============================================
# Volumes
# ============================================
volumes:
  postgres_data:
    name: task-manager-postgres-data
  redis_data:
    name: task-manager-redis-data
  nginx_logs:
    name: task-manager-nginx-logs

# ============================================
# Networks
# ============================================
networks:
  task-network:
    name: task-manager-network
    driver: bridge
```

### Perhatikan Perbedaan: `ports` vs `expose`

```yaml
# PORTS — Expose ke host machine (accessible dari luar)
nginx:
  ports:
    - "80:80"     # User bisa akses http://server-ip:80
    - "443:443"   # User bisa akses https://server-ip:443

# EXPOSE — Cuma accessible dari dalam Docker network
api:
  expose:
    - "8080"      # Cuma Nginx yang bisa akses (internal Docker network)

postgres:
  expose:
    - "5432"      # Cuma Go app yang bisa akses (internal)

# Security best practice:
# HANYA Nginx yang expose ports ke public
# Backend services (Go, Postgres, Redis) cuma expose ke internal network
```

### Build dan Run

```bash
# Build semua services
docker compose build

# Start semua services
docker compose up -d

# Check status
docker compose ps
# NAME                    STATUS     PORTS
# task-manager-nginx      Up         0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp
# task-manager-api        Up         8080/tcp (not published to host!)
# task-manager-db         Up         5432/tcp (not published to host!)
# task-manager-redis      Up         6379/tcp (not published to host!)

# Test
curl http://localhost/health
# {"status":"ok"}

# Check Nginx logs
docker compose logs nginx

# Check all logs
docker compose logs -f
```

### Directory Structure

```
project/
├── Dockerfile              ← Go app (multi-stage, production)
├── Dockerfile.nginx        ← Custom Nginx
├── docker-compose.yml      ← Full stack
├── docker-compose.dev.yml  ← Development (tanpa Nginx)
├── nginx/
│   ├── nginx.conf          ← Main Nginx config
│   └── conf.d/
│       └── taskmanager.conf  ← Site config
├── static/                 ← Static files (served by Nginx)
│   └── ...
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   └── ...
└── ...
```

---

## 14.17 Multiple Apps / Subdomains

Satu Nginx server bisa handle **banyak aplikasi** di domain/subdomain berbeda. Ini namanya **virtual hosting**.

### Setup Multiple Server Blocks

```nginx
# /etc/nginx/conf.d/api.taskmanager.conf
# Backend API (Go)
server {
    listen 80;
    server_name api.taskmanager.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# /etc/nginx/conf.d/admin.taskmanager.conf
# Admin Dashboard (React/Vue frontend)
server {
    listen 80;
    server_name admin.taskmanager.com;

    # Serve frontend build (static files)
    root /var/www/taskmanager-admin/dist;
    index index.html;

    # SPA routing — semua route ke index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy (kalau frontend perlu akses API)
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# /etc/nginx/conf.d/docs.taskmanager.conf
# API Documentation (Swagger/Redoc)
server {
    listen 80;
    server_name docs.taskmanager.com;

    root /var/www/taskmanager-docs;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}
```

### Docker Compose dengan Multiple Apps

```yaml
# docker-compose.yml — Multi-app setup

services:
  nginx:
    image: nginx:1.27-alpine
    container_name: nginx-proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
    depends_on:
      - api
      - admin
    networks:
      - app-network
    restart: unless-stopped

  api:
    build: ./backend
    container_name: task-api
    expose:
      - "8080"
    networks:
      - app-network

  admin:
    build: ./frontend
    container_name: task-admin
    expose:
      - "3000"
    networks:
      - app-network
```

```nginx
# nginx/conf.d/apps.conf — Multiple apps in Docker

# Backend API
upstream api_backend {
    server api:8080;
}

# Admin Frontend
upstream admin_frontend {
    server admin:3000;
}

server {
    listen 80;
    server_name api.taskmanager.com;

    location / {
        proxy_pass http://api_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name admin.taskmanager.com;

    location / {
        proxy_pass http://admin_frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### DNS Setup

```
Untuk multiple subdomains, kamu perlu DNS records:

A Record:
api.taskmanager.com    → 13.250.xx.xx (IP server)
admin.taskmanager.com  → 13.250.xx.xx (IP server yang sama!)
docs.taskmanager.com   → 13.250.xx.xx (IP server yang sama!)

Semua subdomain point ke IP yang SAMA.
Nginx yang memutuskan request mana ke app mana
berdasarkan "Host" header dari request.
```

---

## 14.18 Performance Tuning

### Worker Processes dan Connections

```nginx
# ============================================
# Performance Tuning — nginx.conf
# ============================================

# Worker processes = jumlah CPU cores
# "auto" = detect otomatis
worker_processes auto;

# CPU affinity — bind worker ke specific CPU core
# Ini opsional, tapi bisa improve performance di high-traffic
# worker_cpu_affinity auto;

# Max open files per worker (increase untuk high traffic)
worker_rlimit_nofile 65535;

events {
    # Max connections per worker
    # Rule of thumb: worker_connections = worker_rlimit_nofile / worker_processes
    worker_connections 4096;

    # Accept multiple connections at once
    multi_accept on;

    # Use epoll (Linux) — most efficient event method
    use epoll;
    # Di macOS otomatis pakai kqueue
}

http {
    # ---- Keepalive ----
    keepalive_timeout 65;     # Keep connection alive 65s
    keepalive_requests 1000;  # Max requests per keepalive connection

    # ---- Buffers ----
    client_body_buffer_size 16k;     # Buffer untuk request body
    client_header_buffer_size 1k;    # Buffer untuk request header
    large_client_header_buffers 4 8k; # Large header buffer
    client_max_body_size 10m;        # Max upload size (10MB)

    # ---- Timeouts ----
    client_body_timeout 12;      # Timeout baca client request body
    client_header_timeout 12;    # Timeout baca client request header
    send_timeout 10;             # Timeout kirim response ke client

    # ---- File handling ----
    sendfile on;           # Efficient file transfer (bypass user space)
    tcp_nopush on;         # Send headers in one packet
    tcp_nodelay on;        # Disable Nagle's algorithm (reduce latency)

    # ---- Open file cache ----
    open_file_cache max=1000 inactive=20s;
    open_file_cache_valid 30s;
    open_file_cache_min_uses 2;
    open_file_cache_errors on;
}
```

### Benchmark dengan ab (Apache Benchmark)

```bash
# Install ab (biasanya udah ada di server)
sudo apt install apache2-utils -y   # Ubuntu
brew install httpd                   # macOS

# Basic benchmark — 1000 requests, 10 concurrent
ab -n 1000 -c 10 http://api.taskmanager.com/health

# Output yang penting:
# Requests per second: 2500.00 [#/sec]     ← RPS
# Time per request: 4.000 [ms]             ← Avg response time
# Transfer rate: 500.00 [Kbytes/sec]
#
# Percentage of the requests served within a certain time (ms)
#   50%      3
#   90%      5
#   99%      12
```

### Benchmark dengan wrk (lebih advanced)

```bash
# Install wrk
sudo apt install wrk -y    # Ubuntu
brew install wrk            # macOS

# Basic benchmark — 10 threads, 100 connections, 30 seconds
wrk -t10 -c100 -d30s http://api.taskmanager.com/health

# Output:
# Running 30s test @ http://api.taskmanager.com/health
#   10 threads and 100 connections
#   Thread Stats   Avg      Stdev     Max   +/- Stdev
#     Latency     2.50ms    1.20ms   30ms    90%
#     Req/Sec     4.2k      500      6.0k    70%
#   1,260,000 requests in 30s, 200MB read
# Requests/sec:  42,000.00        ← Ini bagus!
# Transfer/sec:  6.67MB

# Benchmark dengan specific endpoint + headers
wrk -t10 -c100 -d30s \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://api.taskmanager.com/api/v1/tasks
```

### Performance Target

```
Target untuk API:
━━━━━━━━━━━━━━━
Health check endpoint: > 10,000 RPS
API endpoint (with DB): > 1,000 RPS
P99 latency: < 100ms

Kalau belum achieve:
1. Check Nginx config (worker_processes, connections)
2. Check Go app (goroutine leaks, slow queries)
3. Check database (indexes, connection pool)
4. Check Redis (cache hit rate)
5. Scale horizontally (more instances + load balancing)
```

---

## 14.19 Common Issues & Debugging

### 502 Bad Gateway

```
Artinya: Nginx bisa terima request, tapi BACKEND GA BISA DIREACH.

Penyebab umum:
1. Go app ga running
2. Go app listen di port yang salah
3. proxy_pass URL salah
4. Firewall block connection internal
5. Docker: service name salah di proxy_pass

Debugging:
```

```bash
# 1. Check Go app running ga?
curl http://localhost:8080/health
# Kalau error → Go app ga running

# 2. Check process
ps aux | grep task-manager
# Atau di Docker:
docker compose ps

# 3. Check Go app logs
docker compose logs api

# 4. Check Nginx error log
sudo tail -f /var/log/nginx/error.log
# Biasanya ada: "connect() failed (111: Connection refused)"
# Artinya: backend ga available

# 5. Verify proxy_pass URL
nginx -T | grep proxy_pass
```

### 504 Gateway Timeout

```
Artinya: Backend TERLALU LAMA respond. Nginx nyerah nunggu.

Penyebab umum:
1. Query database lambat (missing index, heavy query)
2. External API call timeout
3. Deadlock atau infinite loop di Go app
4. proxy_read_timeout terlalu kecil

Fix:
```

```nginx
# Increase timeout (tapi jangan terlalu besar!)
location / {
    proxy_pass http://localhost:8080;
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 120s;    # Increase dari 60s ke 120s
}
```

```bash
# Cari endpoint yang lambat di Nginx log
grep 'rt=[1-9]' /var/log/nginx/taskmanager_access.log | tail -20

# Check Go app performance
# Tambah logging di middleware untuk response time
```

### Permission Denied

```
Artinya: Nginx ga punya permission buat akses file/directory.

Penyebab umum:
1. Static files owned by root, tapi Nginx jalan sebagai www-data
2. Log directory ga writable
3. SELinux blocking (di CentOS/RHEL)

Fix:
```

```bash
# Check Nginx user
grep "user" /etc/nginx/nginx.conf
# user www-data;

# Fix file permissions
sudo chown -R www-data:www-data /var/www/taskmanager/
sudo chmod -R 755 /var/www/taskmanager/

# Fix log permissions
sudo chown -R www-data:adm /var/log/nginx/
```

### Config Test dan Reload vs Restart

```bash
# SELALU test config sebelum reload/restart!
sudo nginx -t
# nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
# nginx: configuration file /etc/nginx/nginx.conf test is successful

# RELOAD (recommended — zero downtime)
sudo systemctl reload nginx
# - Worker processes baru di-start dengan config baru
# - Worker processes lama selesaikan request yang ongoing
# - Setelah selesai, worker lama di-kill
# - ZERO DOWNTIME!

# RESTART (kalau reload ga cukup)
sudo systemctl restart nginx
# - Nginx di-stop
# - Nginx di-start dengan config baru
# - ADA DOWNTIME (sebentar, tapi ada)
# - Pakai ini kalau binary Nginx di-update
```

### Common Config Mistakes

```nginx
# MISTAKE 1: Lupa semicolon
server_name api.taskmanager.com    # ← Missing semicolon!

# MISTAKE 2: Lupa closing bracket
server {
    location / {
        proxy_pass http://localhost:8080;
    }
# ← Missing closing bracket for server!

# MISTAKE 3: proxy_pass trailing slash
location /api/ {
    proxy_pass http://localhost:8080;     # Request: /api/tasks → Backend: /api/tasks
    proxy_pass http://localhost:8080/;    # Request: /api/tasks → Backend: /tasks
    #                                ^^^ Trailing slash = strip location prefix!
}

# MISTAKE 4: Duplicate listen + server_name
# Dua server block dengan listen 80 dan server_name yang sama
# → Nginx ambil yang pertama, ignore yang kedua

# MISTAKE 5: Config file ga di-include
# File ada di conf.d/ tapi extension-nya bukan .conf
# nginx.conf: include /etc/nginx/conf.d/*.conf;
# ← File .txt atau tanpa extension ga ke-include!
```

### Debug Checklist

```bash
# 1. Config valid?
sudo nginx -t

# 2. Nginx running?
sudo systemctl status nginx

# 3. Port 80/443 accessible?
sudo ss -tlnp | grep -E ':80|:443'

# 4. Backend accessible dari Nginx server?
curl http://localhost:8080/health

# 5. DNS resolve ke IP yang benar?
nslookup api.taskmanager.com
dig api.taskmanager.com

# 6. Firewall blocking?
sudo ufw status        # Ubuntu
sudo iptables -L       # Generic

# 7. Error log bilang apa?
sudo tail -50 /var/log/nginx/error.log

# 8. Access log ada request masuk?
sudo tail -50 /var/log/nginx/access.log

# 9. Docker: container semua running?
docker compose ps

# 10. Docker: bisa ping service lain?
docker compose exec nginx ping api
```

---

## 14.20 Nginx vs Alternatives

Nginx bukan satu-satunya pilihan. Ada beberapa alternatif yang worth considering:

### Caddy — Auto HTTPS, Zero Config

```
Caddy:
━━━━━━
- Written in Go (kita kan belajar Go!)
- AUTO HTTPS — Let's Encrypt otomatis, zero config
- Config JAUH lebih simple dari Nginx
- Perfect untuk project kecil-menengah
- Reverse proxy, static files, load balancing — semua ada

Kelebihan:
+ Auto HTTPS (ga perlu Certbot, ga perlu manual renewal)
+ Config file super simple (Caddyfile)
+ HTTP/2 dan HTTP/3 out of the box
+ Built-in access log, error log
+ Cocok banget untuk developer yang ga mau ribet

Kekurangan:
- Lebih lambat dari Nginx untuk static files
- Community lebih kecil
- Kurang customizable dibanding Nginx
- Jarang dipakai di enterprise/large scale
```

```
# Caddyfile — seluruh reverse proxy + HTTPS dalam 4 baris!
api.taskmanager.com {
    reverse_proxy localhost:8080
}

# Bandingkan dengan Nginx yang butuh 30+ baris config.
# Caddy auto-generate SSL certificate dari Let's Encrypt!
```

### Traefik — Docker-Native

```
Traefik:
━━━━━━━
- Designed untuk Docker dan Kubernetes
- Auto-discover containers (ga perlu manual config)
- Dashboard built-in
- Let's Encrypt built-in

Kelebihan:
+ Auto-discover Docker containers (pakai labels)
+ Perfect untuk microservices
+ Dashboard untuk monitoring
+ Native support Docker, Kubernetes, Consul

Kekurangan:
- Config lebih complex (YAML/TOML)
- Performance dibawah Nginx untuk static files
- Overkill untuk single-app deployment
- Learning curve lebih curam
```

```yaml
# docker-compose.yml dengan Traefik
services:
  traefik:
    image: traefik:v3
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock

  api:
    build: .
    labels:
      - "traefik.http.routers.api.rule=Host(`api.taskmanager.com`)"
      - "traefik.http.services.api.loadbalancer.server.port=8080"
    # Traefik auto-detect container dan route traffic!
```

### HAProxy — Pure Load Balancer

```
HAProxy:
━━━━━━━
- Fokus ke load balancing dan proxying
- Paling mature, paling battle-tested
- Dipake oleh GitHub, Stack Overflow, Reddit

Kelebihan:
+ Performance terbaik untuk pure proxying
+ Advanced load balancing algorithms
+ Health checking yang sangat baik
+ Rate limiting yang lebih granular

Kekurangan:
- Ga bisa serve static files
- Ga ada SSL termination built-in (butuh tambahan)
- Config syntax kurang intuitif
- Overkill untuk project kecil
```

### Kapan Pakai Apa?

```
┌────────────────────────────────────────────────────────────┐
│                   Decision Matrix                           │
│                                                             │
│  Project kecil, 1 developer, mau simple?                   │
│  → CADDY (auto HTTPS, config 4 baris)                      │
│                                                             │
│  Docker-heavy, microservices, dynamic containers?           │
│  → TRAEFIK (auto-discover, Docker-native)                  │
│                                                             │
│  Need extreme performance, enterprise, huge traffic?        │
│  → NGINX (battle-tested, most configurable)                 │
│                                                             │
│  Pure load balancing, ga perlu static files?                │
│  → HAPROXY (best-in-class LB)                              │
│                                                             │
│  Learning DevOps, interview prep, industri standard?        │
│  → NGINX (paling banyak diminta di job postings)            │
│                                                             │
│  Kita di bab ini: NGINX ← karena industri standard         │
│  Nanti di Kubernetes: bisa pakai Traefik atau Nginx Ingress│
└────────────────────────────────────────────────────────────┘
```

| Tool | Best For | Auto HTTPS | Static Files | Docker Native | Market Share |
|------|----------|------------|-------------|---------------|-------------|
| **Nginx** | General purpose, enterprise | Manual (Certbot) | Excellent | Good | 35% |
| **Caddy** | Small-medium projects | Built-in | Good | Good | ~5% |
| **Traefik** | Microservices, Docker/K8s | Built-in | No | Excellent | ~8% |
| **HAProxy** | Pure load balancing | No | No | Good | ~10% |

---

## 14.21 Full Production Config

Ini dia **complete production config** untuk Task Manager API. Copy-paste ready. Semua yang udah kita bahas digabung jadi satu.

### nginx.conf (Main Config)

```nginx
# /etc/nginx/nginx.conf
# Task Manager API — Production Configuration
# ==============================================

user www-data;
worker_processes auto;
worker_rlimit_nofile 65535;
pid /run/nginx.pid;
error_log /var/log/nginx/error.log warn;

events {
    worker_connections 4096;
    multi_accept on;
    use epoll;
}

http {
    # ============================================
    # Basic Settings
    # ============================================
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    charset utf-8;
    server_tokens off;            # Hide Nginx version

    # ============================================
    # Logging
    # ============================================
    log_format main '$remote_addr - $remote_user [$time_local] '
                    '"$request" $status $body_bytes_sent '
                    '"$http_referer" "$http_user_agent" '
                    'rt=$request_time uct=$upstream_connect_time '
                    'uht=$upstream_header_time urt=$upstream_response_time';

    log_format json escape=json
        '{'
            '"time": "$time_iso8601", '
            '"remote_addr": "$remote_addr", '
            '"request": "$request", '
            '"status": $status, '
            '"body_bytes_sent": $body_bytes_sent, '
            '"request_time": $request_time, '
            '"upstream_response_time": "$upstream_response_time", '
            '"user_agent": "$http_user_agent"'
        '}';

    access_log /var/log/nginx/access.log main;

    # ============================================
    # Performance
    # ============================================
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    keepalive_requests 1000;
    types_hash_max_size 2048;

    # Buffers
    client_body_buffer_size 16k;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 8k;
    client_max_body_size 10m;

    # Timeouts
    client_body_timeout 12;
    client_header_timeout 12;
    send_timeout 10;

    # Open file cache
    open_file_cache max=1000 inactive=20s;
    open_file_cache_valid 30s;
    open_file_cache_min_uses 2;

    # ============================================
    # Gzip Compression
    # ============================================
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_min_length 256;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml
        application/xml+rss
        application/vnd.ms-fontobject
        font/ttf
        font/opentype
        image/svg+xml;

    # ============================================
    # Rate Limiting
    # ============================================
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=login_limit:5m rate=5r/m;
    limit_req_zone $binary_remote_addr zone=register_limit:5m rate=2r/m;
    limit_req_status 429;
    limit_req_log_level warn;

    # ============================================
    # Upstream (Backend Servers)
    # ============================================
    upstream task_api {
        least_conn;
        server 127.0.0.1:8080 max_fails=3 fail_timeout=30s;
        # Tambah server di bawah untuk load balancing:
        # server 127.0.0.1:8081 max_fails=3 fail_timeout=30s;
        # server 127.0.0.1:8082 max_fails=3 fail_timeout=30s;
    }

    # ============================================
    # Include site configs
    # ============================================
    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
```

### taskmanager.conf (Site Config)

```nginx
# /etc/nginx/sites-available/taskmanager.conf
# Task Manager API — Site Configuration (Production)
# ====================================================

# ============================================
# HTTP → HTTPS Redirect
# ============================================
server {
    listen 80;
    listen [::]:80;
    server_name api.taskmanager.com;

    # Let's Encrypt challenge (keep this for certificate renewal)
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    # Redirect everything else to HTTPS
    location / {
        return 301 https://$host$request_uri;
    }
}

# ============================================
# HTTPS Server — Main Config
# ============================================
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.taskmanager.com;

    # ---- SSL CERTIFICATE ----
    ssl_certificate /etc/letsencrypt/live/api.taskmanager.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.taskmanager.com/privkey.pem;

    # ---- SSL SETTINGS ----
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;

    # ---- SECURITY HEADERS ----
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'none'; frame-ancestors 'none'" always;
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;

    # ---- ACCESS LOG ----
    access_log /var/log/nginx/taskmanager_access.log main;
    error_log /var/log/nginx/taskmanager_error.log warn;

    # ============================================
    # Health Check (no logging, no rate limit)
    # ============================================
    location = /health {
        access_log off;
        proxy_pass http://task_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # ============================================
    # Static Files
    # ============================================
    location /uploads/ {
        alias /var/www/taskmanager/uploads/;
        expires 7d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    location /static/ {
        alias /var/www/taskmanager/static/;
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    location = /favicon.ico {
        alias /var/www/taskmanager/static/favicon.ico;
        access_log off;
        log_not_found off;
    }

    # ============================================
    # Login Endpoint (strict rate limit)
    # ============================================
    location = /api/v1/auth/login {
        limit_req zone=login_limit burst=5 nodelay;

        proxy_pass http://task_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # ============================================
    # Register Endpoint (strict rate limit)
    # ============================================
    location = /api/v1/auth/register {
        limit_req zone=register_limit burst=2 nodelay;

        proxy_pass http://task_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # ============================================
    # WebSocket
    # ============================================
    location /ws/ {
        proxy_pass http://task_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # ============================================
    # API — Catch All (general rate limit)
    # ============================================
    location / {
        limit_req zone=api_limit burst=20 nodelay;

        proxy_pass http://task_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # ============================================
    # Block common attack paths
    # ============================================
    location ~ /\. {
        deny all;          # Block .env, .git, .htaccess, dll
        access_log off;
        log_not_found off;
    }

    location ~* ^/(wp-admin|wp-login|xmlrpc|phpmyadmin) {
        deny all;          # Block WordPress/PHP attack vectors
        access_log off;
        log_not_found off;
    }
}
```

### Deploy Checklist untuk Config Ini

```bash
# 1. Copy config files
sudo cp nginx.conf /etc/nginx/nginx.conf
sudo cp taskmanager.conf /etc/nginx/sites-available/taskmanager.conf

# 2. Enable site
sudo ln -sf /etc/nginx/sites-available/taskmanager.conf /etc/nginx/sites-enabled/

# 3. Bikin directories
sudo mkdir -p /var/www/taskmanager/{static,uploads}
sudo chown -R www-data:www-data /var/www/taskmanager/
sudo mkdir -p /var/www/certbot

# 4. Test config
sudo nginx -t

# 5. Obtain SSL certificate
sudo certbot --nginx -d api.taskmanager.com

# 6. Reload Nginx
sudo systemctl reload nginx

# 7. Verify
curl -I https://api.taskmanager.com/health
# HTTP/2 200
# x-content-type-options: nosniff
# x-frame-options: DENY
# strict-transport-security: max-age=63072000...

# 8. Test SSL grade
# https://www.ssllabs.com/ssltest/analyze.html?d=api.taskmanager.com
```

---

## 14.22 Cheat Sheet

### Nginx Commands

```bash
# ========== INSTALL ==========
brew install nginx                       # macOS
sudo apt install nginx -y                # Ubuntu/Debian
docker run -d -p 80:80 nginx:alpine      # Docker

# ========== SERVICE ==========
sudo systemctl start nginx               # Start
sudo systemctl stop nginx                # Stop
sudo systemctl restart nginx             # Restart (ada downtime)
sudo systemctl reload nginx              # Reload config (zero downtime)
sudo systemctl status nginx              # Check status
sudo systemctl enable nginx              # Auto-start saat boot

# ========== CONFIG ==========
sudo nginx -t                            # Test config syntax
sudo nginx -T                            # Test + print full config
sudo nginx -V                            # Show version + compile options

# ========== SIGNALS ==========
sudo nginx -s stop                       # Fast shutdown
sudo nginx -s quit                       # Graceful shutdown
sudo nginx -s reload                     # Reload config
sudo nginx -s reopen                     # Reopen log files

# ========== LOGS ==========
sudo tail -f /var/log/nginx/access.log   # Follow access log
sudo tail -f /var/log/nginx/error.log    # Follow error log
sudo tail -100 /var/log/nginx/error.log  # Last 100 lines

# ========== SSL (CERTBOT) ==========
sudo certbot --nginx -d domain.com       # Obtain + auto-config
sudo certbot renew --dry-run             # Test renewal
sudo certbot renew                       # Manual renewal
sudo certbot certificates                # List certificates

# ========== DEBUG ==========
curl -I http://localhost                 # Check response headers
sudo ss -tlnp | grep nginx              # Check listening ports
ps aux | grep nginx                      # Check processes
```

### Common Config Patterns

```nginx
# ========== REVERSE PROXY ==========
location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

# ========== LOAD BALANCING ==========
upstream backend {
    least_conn;
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
}
location / {
    proxy_pass http://backend;
}

# ========== HTTPS REDIRECT ==========
server {
    listen 80;
    server_name domain.com;
    return 301 https://$host$request_uri;
}

# ========== STATIC FILES ==========
location /static/ {
    alias /var/www/app/static/;
    expires 30d;
}

# ========== GZIP ==========
gzip on;
gzip_types text/plain text/css application/json application/javascript;
gzip_min_length 256;
gzip_comp_level 6;

# ========== RATE LIMITING ==========
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
location /api/ {
    limit_req zone=api burst=20 nodelay;
}

# ========== SECURITY HEADERS ==========
add_header X-Content-Type-Options "nosniff" always;
add_header X-Frame-Options "DENY" always;
add_header X-XSS-Protection "1; mode=block" always;
server_tokens off;

# ========== WEBSOCKET ==========
location /ws/ {
    proxy_pass http://localhost:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 3600s;
}

# ========== BLOCK HIDDEN FILES ==========
location ~ /\. {
    deny all;
}

# ========== CUSTOM ERROR PAGE ==========
error_page 502 /502.html;
location = /502.html {
    root /usr/share/nginx/html;
    internal;
}
```

### Config File Locations

```
macOS (Homebrew — Apple Silicon):
  Config:  /opt/homebrew/etc/nginx/nginx.conf
  Logs:    /opt/homebrew/var/log/nginx/
  Root:    /opt/homebrew/var/www/

macOS (Homebrew — Intel):
  Config:  /usr/local/etc/nginx/nginx.conf
  Logs:    /usr/local/var/log/nginx/
  Root:    /usr/local/var/www/

Ubuntu/Debian:
  Config:  /etc/nginx/nginx.conf
  Sites:   /etc/nginx/sites-available/
  Enabled: /etc/nginx/sites-enabled/
  Conf.d:  /etc/nginx/conf.d/
  Logs:    /var/log/nginx/
  Root:    /var/www/html/
  PID:     /run/nginx.pid

Docker (nginx:alpine):
  Config:  /etc/nginx/nginx.conf
  Conf.d:  /etc/nginx/conf.d/
  Logs:    /var/log/nginx/
  Root:    /usr/share/nginx/html/
```

---

## 14.23 Checklist Bab Ini

### Konsep Dasar

- [ ] Paham apa itu Nginx (web server, reverse proxy, load balancer)
- [ ] Paham perbedaan forward proxy vs reverse proxy
- [ ] Paham kenapa Nginx (C10K problem, event-driven, non-blocking)
- [ ] Paham perbedaan Apache vs Nginx
- [ ] Paham analogi resepsionis hotel

### Architecture

- [ ] Paham master process vs worker process
- [ ] Paham event-driven vs thread-based
- [ ] Paham config file structure (nginx.conf, sites-available, conf.d)
- [ ] Paham config hierarchy (main → events → http → server → location)

### Installation

- [ ] Nginx terinstall (brew / apt / Docker)
- [ ] Bisa jalanin nginx -t untuk test config
- [ ] Paham reload vs restart
- [ ] Tau lokasi config files di OS kamu

### Config Basics

- [ ] Paham directives dan blocks
- [ ] Paham syntax rules (semicolon, curly braces, comments)
- [ ] Paham location block matching dan priority

### Reverse Proxy

- [ ] Udah setup reverse proxy ke Go app (proxy_pass)
- [ ] Paham kenapa proxy headers penting (X-Real-IP, X-Forwarded-For, dll)
- [ ] Go app udah di-update untuk SetTrustedProxies
- [ ] Test reverse proxy berhasil (curl lewat Nginx)

### SSL/HTTPS

- [ ] Paham kenapa HTTPS wajib
- [ ] Certbot terinstall
- [ ] Udah obtain SSL certificate dari Let's Encrypt
- [ ] HTTP → HTTPS redirect berjalan
- [ ] SSL best practices (TLS 1.2+, HSTS, OCSP stapling)
- [ ] Auto-renewal cron/timer terpasang
- [ ] SSL grade A atau A+ di ssllabs.com

### Load Balancing

- [ ] Paham upstream block
- [ ] Paham algorithms: round-robin, least_conn, ip_hash, weighted
- [ ] Paham health check (max_fails, fail_timeout, backup)
- [ ] Tau kapan pakai algorithm yang mana

### Caching

- [ ] Paham proxy_cache setup
- [ ] Paham kapan cache vs tidak cache API response
- [ ] Paham cache bypass
- [ ] Tau bedanya caching di Nginx vs Redis

### Compression

- [ ] Gzip enabled di config
- [ ] Paham gzip_types dan gzip_comp_level
- [ ] Verify compression working (Content-Encoding: gzip)

### Rate Limiting

- [ ] Paham limit_req_zone dan limit_req
- [ ] Paham burst dan nodelay
- [ ] Setup rate limit berbeda per endpoint (login lebih ketat)
- [ ] Paham bedanya Nginx rate limit vs Redis rate limit

### Security

- [ ] Security headers lengkap (X-Content-Type-Options, X-Frame-Options, dll)
- [ ] server_tokens off (hide version)
- [ ] Block hidden files (location ~ /\\.)
- [ ] Block common attack paths (wp-admin, phpmyadmin)
- [ ] Verify di securityheaders.com

### Logging

- [ ] Access log dan error log configured
- [ ] Custom log format dengan response time
- [ ] Health check access log disabled
- [ ] Paham log rotation (logrotate)
- [ ] Bisa analisis log (top IPs, status codes, slow requests)

### Static Files

- [ ] Paham serve static files langsung dari Nginx
- [ ] Paham root vs alias
- [ ] Cache headers untuk static files (expires, Cache-Control)

### WebSocket

- [ ] Paham WebSocket proxy config
- [ ] Paham kenapa perlu Upgrade dan Connection headers
- [ ] Paham kenapa perlu proxy_http_version 1.1

### Docker Integration

- [ ] Nginx jalan sebagai Docker container
- [ ] docker-compose.yml include Nginx + Go + PostgreSQL + Redis
- [ ] Paham ports vs expose (security)
- [ ] Backend services TIDAK exposed ke host (cuma Nginx)
- [ ] Full stack bisa start dengan `docker compose up -d`

### Multiple Apps

- [ ] Paham virtual hosting (multiple server blocks)
- [ ] Bisa setup subdomain routing (api.*, admin.*)
- [ ] Paham DNS setup untuk subdomains

### Performance

- [ ] Paham worker_processes, worker_connections, keepalive
- [ ] Paham buffer sizes dan timeouts
- [ ] Bisa benchmark dengan ab atau wrk
- [ ] Tau target performance (RPS, latency)

### Debugging

- [ ] Bisa debug 502 Bad Gateway
- [ ] Bisa debug 504 Gateway Timeout
- [ ] Paham debug checklist (10 steps)
- [ ] Paham common config mistakes

### Alternatives

- [ ] Paham Caddy (auto HTTPS, simple config)
- [ ] Paham Traefik (Docker-native, microservices)
- [ ] Paham HAProxy (pure load balancer)
- [ ] Tau kapan pakai Nginx vs alternatives

### Production Ready

- [ ] Full production config sudah dibuat
- [ ] Config include: reverse proxy + SSL + gzip + rate limit + security headers + logging
- [ ] Config sudah di-test (nginx -t)
- [ ] Config sudah di-deploy dan verified

---

> **Next:** [Bab 15: Monitoring — Prometheus + Grafana](15-monitoring.md) — setelah Nginx di depan, kita perlu **pantau** semua yang terjadi. Berapa request per detik? Berapa error rate? CPU usage berapa? Response time naik ga? Di bab berikutnya kita setup Prometheus untuk collect metrics dan Grafana untuk visualisasi. Ini yang bikin kamu bisa tidur tenang kalau production running — karena kamu tau persis apa yang terjadi di system kamu.