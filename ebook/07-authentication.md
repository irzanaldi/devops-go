# Bab 7: Authentication — JWT + RBAC

## Overview

Chapter ini bakal jadi salah satu yang paling penting. Tanpa auth, API lu cuma open buffet — siapa aja bisa masuk. Kita bakal bahas dari teori sampai implementasi lengkap: JWT, password hashing, register, login, refresh token, middleware, dan RBAC (Role-Based Access Control).

Semua code di chapter ini **diambil langsung dari project Task Manager kita**. Bukan contoh mainan.

---

## 1. Authentication Overview — Session vs JWT

### Session-Based (Traditional)

```
Client login → Server bikin session → Simpan di server (memory/Redis)
                                    → Kasih session ID via cookie
Client request → Kirim cookie → Server cek session di storage
```

**Cara kerja:**
- Server **menyimpan state** user (siapa yang login) di memory atau database
- Client cuma pegang session ID (biasanya di cookie)
- Setiap request, server harus lookup session ID ke storage

**Masalah session untuk API modern:**
- **Stateful** — server harus inget siapa aja yang login. Kalau restart? Hilang (kecuali simpan di Redis)
- **Scaling susah** — kalau ada 5 server di belakang load balancer, session ada di server A, tapi request berikutnya masuk ke server B. Sticky session jadi solution, tapi ribet
- **Cross-domain pain** — cookie-based, susah dipakai mobile app atau beda domain
- **CSRF vulnerable** — cookie otomatis dikirim browser, harus pakai CSRF token

### JWT (JSON Web Token)

```
Client login → Server generate JWT → Kasih ke client (body response)
Client request → Kirim JWT di header → Server validasi JWT (stateless)
```

**Cara kerja:**
- Server **tidak menyimpan state**. Semua info user ada di dalam token itu sendiri
- Client simpan token (localStorage/memory), kirim di header setiap request
- Server cuma perlu **verify signature** — tidak perlu lookup ke database

**Kenapa JWT untuk REST API:**
- **Stateless** — server ga perlu simpan apa-apa. Perfect buat horizontal scaling
- **Cross-platform** — bisa dipakai web, mobile, microservices, third-party
- **Self-contained** — info user (ID, email, role) ada di dalam token
- **Standard** — RFC 7519, library ada di semua bahasa

### Perbandingan

| Aspek | Session | JWT |
|-------|---------|-----|
| State | Server-side | Client-side (stateless) |
| Storage | Redis/memory di server | Client (localStorage/memory) |
| Scalability | Butuh shared session store | Mudah scale horizontal |
| Mobile friendly | Ribet (cookie) | Gampang (header) |
| Revocation | Mudah (hapus session) | Susah (harus blacklist) |
| Payload | Minimal (cuma ID) | Bisa bawa data user |

> **Keputusan kita:** Pakai JWT karena API kita stateless, mau bisa di-consume mobile/frontend mana aja, dan mau gampang di-scale.

---

## 2. JWT Deep Dive — Header, Payload, Signature

### Struktur JWT

JWT itu cuma **string panjang** yang terdiri dari 3 bagian, dipisah titik:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6ImFkbWluQHRlc3QuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzE0MzAwMDAwfQ.abc123signature
|___________________________________|.|___________________________________________________________________________________|.|________________|
            Header (Base64)                                  Payload (Base64)                                              Signature
```

### 1. Header

```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

- `alg` — algoritma signing. Kita pakai **HS256** (HMAC-SHA256)
- `typ` — tipe token

### 2. Payload (Claims)

```json
{
  "user_id": 1,
  "email": "admin@test.com",
  "role": "admin",
  "exp": 1714300000,
  "iat": 1714299100
}
```

Claims itu "klaim" — informasi yang token bawa:
- **Custom claims**: `user_id`, `email`, `role` — data yang kita masukin sendiri
- **Registered claims** (standard): `exp` (expiry), `iat` (issued at), `sub` (subject), `iss` (issuer)

> **PENTING:** Payload itu cuma di-encode (Base64), **BUKAN di-encrypt**. Siapa aja bisa baca isinya! Jangan masukin data sensitif (password, credit card, dll).

### 3. Signature

```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret_key
)
```

Ini yang bikin JWT aman. Server pakai **secret key** untuk sign. Kalau ada yang modify payload, signature-nya ga cocok lagi — server bakal reject.

### Access Token vs Refresh Token

| | Access Token | Refresh Token |
|--|-------------|--------------|
| **Tujuan** | Akses resource/API | Dapat access token baru |
| **Expiry** | Pendek (15 menit) | Panjang (7 hari) |
| **Dikirim** | Setiap API request | Cuma saat refresh |
| **Disimpan** | Memory/variable | Secure storage |
| **Kalau bocor** | Bahaya 15 menit | Bahaya 7 hari |

**Flow-nya:**

```
1. Login → Dapat access_token (15min) + refresh_token (7hari)
2. Request API → Pakai access_token di header
3. Access token expired → Kirim refresh_token ke /auth/refresh
4. Dapat access_token baru + refresh_token baru
5. Ulangi dari step 2
```

Kenapa dua token?
- Access token sengaja **pendek umurnya** biar kalau bocor, damage-nya kecil
- Refresh token umurnya panjang biar user ga perlu login ulang tiap 15 menit
- Best practice: kalau refresh token dipakai, **generate pair baru** (rotation)

---

## 3. Password Hashing — Bcrypt

### Kenapa Bukan MD5 atau SHA?

```
MD5("password123")  → "482c811da5d5b4bc6d497ffa98491e38"
SHA256("password123") → "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f"
```

**Masalah MD5/SHA:**
1. **Terlalu cepat** — GPU modern bisa hash miliaran MD5 per detik. Brute force jadi gampang
2. **Tidak ada salt** — password yang sama = hash yang sama. Bisa pakai rainbow table
3. **Bukan untuk password** — MD5/SHA dirancang untuk checksum, bukan security

### Kenapa Bcrypt?

```
Bcrypt("password123") → "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
```

**Keunggulan bcrypt:**
1. **Intentionally slow** — dirancang lambat, brute force jadi ga feasible
2. **Built-in salt** — setiap hash punya salt unik, password sama = hash beda
3. **Cost factor** — bisa adjust seberapa lambat (naik seiring hardware makin kuat)
4. **Battle-tested** — standar industri sejak 1999, masih aman sampai sekarang

### Anatomy of Bcrypt Hash

```
$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
 │  │  │______________________││____________________________│
 │  │  │         Salt          │       Hashed password
 │  │  Cost factor (2^10 = 1024 rounds)
 │  Version
 Algorithm identifier
```

### Implementasi di Project Kita

File: **`internal/model/user.go`**

```go
package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Email     string         `json:"email" gorm:"size:100;uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"size:255;not null"`
	Role      Role           `json:"role" gorm:"size:20;default:member"`
	Boards    []Board        `json:"boards,omitempty" gorm:"foreignKey:OwnerID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (u *User) HashPassword() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
```

**Yang perlu diperhatiin:**

1. **`json:"-"`** pada field Password — ini penting banget! Artinya password **ga pernah** masuk ke JSON response. Kalau lu lupa ini, password hash bisa ke-expose ke client
2. **`bcrypt.DefaultCost`** — nilainya 10 (2^10 = 1024 rounds). Cukup untuk sekarang. Kalau hardware makin kuat, naikin cost-nya
3. **`GenerateFromPassword`** otomatis generate salt — lu ga perlu manage salt sendiri
4. **`CompareHashAndPassword`** — ini yang dipakai pas login. Compare password plain text dari user dengan hash di database

### NestJS Equivalent

```typescript
// NestJS biasanya pakai bcryptjs
import * as bcrypt from 'bcryptjs';

// Hash
const salt = await bcrypt.genSalt(10);
const hashedPassword = await bcrypt.hash(password, salt);

// Compare
const isMatch = await bcrypt.compare(password, hashedPassword);
```

Di Go lebih clean — `GenerateFromPassword` handle salt generation otomatis.

---

## 4. Register Flow

### Request/Response DTO

Masih di **`internal/model/user.go`**:

```go
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
	}
}
```

**Perhatikan:**
- **Binding tags** (`binding:"required,min=2,max=100"`) — ini Gin punya validator. Equivalent dengan `class-validator` di NestJS
- **Separate DTO** — `RegisterRequest` untuk input, `UserResponse` untuk output. Password ga pernah masuk response
- **`ToResponse()`** method — convert User entity ke safe response

### Service Layer — Register Logic

File: **`internal/service/auth.go`** (bagian Register)

```go
func (s *AuthService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
	// Step 1: Check if email already exists
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Step 2: Create user object
	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     model.RoleMember, // Default role: member
	}

	// Step 3: Hash password
	if err := user.HashPassword(); err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Step 4: Save to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Step 5: Return safe response (tanpa password)
	resp := user.ToResponse()
	return &resp, nil
}
```

**Flow Register:**

```
1. Validate input (Gin binding — otomatis di handler)
2. Cek email duplicate → query database
3. Create User struct, set role default "member"
4. Hash password pakai bcrypt
5. Simpan ke database
6. Return UserResponse (tanpa password)
```

### Handler Layer — Register Endpoint

File: **`internal/handler/auth.go`** (bagian Register)

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/service"
	"github.com/irzan-aldi/task-manager/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", user)
}
```

**Handler ini cuma 3 hal:**
1. **Parse & validate** request body (`ShouldBindJSON` — kalau binding gagal, otomatis error)
2. **Delegate** ke service layer
3. **Return** response dengan status code yang tepat

### Test dengan cURL

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Irzan Aldi",
    "email": "irzan@test.com",
    "password": "secret123"
  }'

# Response (201 Created)
{
  "status": 201,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "name": "Irzan Aldi",
    "email": "irzan@test.com",
    "role": "member"
  }
}
```

---

## 5. Login Flow — Find User, Check Password, Generate Token Pair

### Service Layer — Login Logic

File: **`internal/service/auth.go`** (bagian Login + token generation)

```go
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type JWTClaims struct {
	UserID uint       `json:"user_id"`
	Email  string     `json:"email"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(req model.LoginRequest) (*TokenPair, error) {
	// Step 1: Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Step 2: Check password
	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Step 3: Generate token pair
	return s.generateTokenPair(user)
}
```

**Perhatiin error message-nya:** `"invalid email or password"` — **sama** untuk email not found dan password salah. Ini best practice! Kalau beda, attacker bisa enumerate email mana yang terdaftar.

### Token Generation

```go
func (s *AuthService) generateTokenPair(user *model.User) (*TokenPair, error) {
	// Access token — 15 minutes
	accessClaims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	// Refresh token — 7 days
	refreshClaims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessString,
		RefreshToken: refreshString,
		ExpiresIn:    900, // 15 minutes in seconds
	}, nil
}
```

**Breakdown:**

1. **`JWTClaims`** struct — embed `jwt.RegisteredClaims` + custom fields (`UserID`, `Email`, `Role`)
2. **`jwt.NewWithClaims(jwt.SigningMethodHS256, claims)`** — buat token baru dengan claims
3. **`token.SignedString([]byte(s.jwtSecret))`** — sign token pakai secret key. Secret dari env variable
4. **Access token** = 15 menit, **Refresh token** = 7 hari
5. **`ExpiresIn: 900`** — kasih tau client kapan token expired (dalam detik). Client bisa set timer buat auto-refresh

### Login DTO

```go
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
```

### Handler Layer — Login Endpoint

```go
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.authService.Login(req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Login successful", tokens)
}
```

### Test dengan cURL

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "irzan@test.com",
    "password": "secret123"
  }'

# Response (200 OK)
{
  "status": 200,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6ImlyemFuQHRlc3QuY29tIiwicm9sZSI6Im1lbWJlciIsImV4cCI6MTcxNDMwMDkwMCwiaWF0IjoxNzE0MzAwMDAwfQ.xxxxx",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6ImlyemFuQHRlc3QuY29tIiwicm9sZSI6Im1lbWJlciIsImV4cCI6MTcxNDkwNDgwMCwiaWF0IjoxNzE0MzAwMDAwfQ.yyyyy",
    "expires_in": 900
  }
}
```

---

## 6. Refresh Token Flow

### Kenapa Perlu Refresh?

Access token sengaja umurnya pendek (15 menit). Tapi user ga mau login ulang tiap 15 menit. Solusinya: **refresh token**.

```
Timeline:
├── 0:00  — Login, dapat access_token (15min) + refresh_token (7hari)
├── 0:14  — Request masih OK, access_token valid
├── 0:15  — Access token expired! API return 401
├── 0:15  — Client kirim refresh_token ke /auth/refresh
├── 0:15  — Dapat access_token BARU + refresh_token BARU
├── 0:30  — Access token expired lagi → refresh lagi
└── 7:00  — Refresh token juga expired → harus login ulang
```

### Service Layer — Refresh Token

```go
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	// Step 1: Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Step 2: Find user (pastikan masih ada di database)
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Step 3: Generate token pair baru
	return s.generateTokenPair(user)
}
```

**Kenapa fetch user lagi dari database?**
- Karena bisa aja user udah di-delete atau di-ban sejak token terakhir dibuat
- Role user bisa berubah (dari member jadi admin, atau sebaliknya)
- Ini juga moment yang bagus untuk **rotate** refresh token (kasih yang baru, invalidate yang lama)

### Token Validation (Dipakai di Refresh dan Middleware)

```go
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check signing method — PENTING buat security
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
```

**Poin penting:**
1. **Check signing method** — ini security measure. Ada attack yang namanya "algorithm confusion" dimana attacker ganti alg dari HS256 ke "none". Dengan check explicit, kita prevent itu
2. **`jwt.ParseWithClaims`** otomatis check expiry — kalau expired, langsung return error
3. Return **claims** supaya caller bisa akses `UserID`, `Email`, `Role`

### Handler Layer — Refresh Endpoint

```go
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Token refreshed", tokens)
}
```

### Test dengan cURL

```bash
# Refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.yyyyy..."
  }'

# Response (200 OK) — token pair baru
{
  "status": 200,
  "message": "Token refreshed",
  "data": {
    "access_token": "NEW_ACCESS_TOKEN...",
    "refresh_token": "NEW_REFRESH_TOKEN...",
    "expires_in": 900
  }
}
```

---

## 7. JWT Middleware — Protecting Routes

Middleware ini jalan **sebelum** handler. Tugasnya: extract token dari header, validate, dan set user info ke context.

### Full Middleware Code

File: **`internal/middleware/auth.go`**

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/irzan-aldi/task-manager/internal/model"
	"github.com/irzan-aldi/task-manager/internal/service"
	"github.com/irzan-aldi/task-manager/pkg/response"
)

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Ambil Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Step 2: Parse "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization format. Use: Bearer <token>")
			c.Abort()
			return
		}

		// Step 3: Validate token
		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		// Step 4: Set user info di context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// Step 5: Lanjut ke handler berikutnya
		c.Next()
	}
}
```

**Alur middleware:**

```
Request masuk
    │
    ├── Ada Authorization header? ── Tidak ──→ 401 "Authorization header required"
    │
    ├── Format "Bearer <token>"? ── Tidak ──→ 401 "Invalid authorization format"
    │
    ├── Token valid & belum expired? ── Tidak ──→ 401 "Invalid or expired token"
    │
    ├── Set userID, email, role ke gin.Context
    │
    └── c.Next() → lanjut ke handler
```

**`c.Abort()`** — ini penting! Tanpa ini, request bakal tetap lanjut ke handler meskipun auth gagal.

**`c.Set()` / `c.Get()`** — ini cara Gin pass data antar middleware dan handler. Mirip `req.user` di Express/NestJS.

### Akses User Info di Handler

Setelah middleware set context, handler bisa akses user info:

```go
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	response.Success(c, http.StatusOK, "Current user", gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}
```

### NestJS Equivalent

```typescript
// NestJS Guard
@Injectable()
export class JwtAuthGuard extends AuthGuard('jwt') {}

// Passport Strategy
@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy) {
  constructor() {
    super({
      jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
      secretOrKey: process.env.JWT_SECRET,
    });
  }

  async validate(payload: any) {
    return { userId: payload.sub, email: payload.email, role: payload.role };
  }
}

// Pakai di Controller
@UseGuards(JwtAuthGuard)
@Get('me')
getMe(@Request() req) {
  return req.user;
}
```

Di NestJS ada magic dari Passport + decorator `@UseGuards`. Di Go, kita nulis sendiri — lebih verbose tapi lu tau **persis** apa yang terjadi. Ga ada hidden behavior.

---

## 8. RBAC (Role-Based Access Control)

### Konsep

RBAC = user punya **role**, dan setiap role punya **permission** yang beda.

```
Admin → bisa semua (manage users, delete users, manage boards, dll)
Member → cuma bisa manage data sendiri (boards sendiri, tasks sendiri)
```

### Role Definition

Di **`internal/model/user.go`**:

```go
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)
```

Simple — pakai `string` type alias. Kenapa bukan `int`? Karena lebih readable di database dan di JWT payload. Kalau nanti role-nya banyak, bisa pindah ke database table tersendiri.

### Role Middleware

File: **`internal/middleware/auth.go`** (bagian RoleMiddleware)

```go
func RoleMiddleware(roles ...model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Ambil role dari context (sudah di-set oleh AuthMiddleware)
		userRole, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "User role not found")
			c.Abort()
			return
		}

		// Step 2: Check apakah role user ada di list yang diizinkan
		role := userRole.(model.Role)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		// Step 3: Kalau ga ada — forbidden
		response.Error(c, http.StatusForbidden, "Insufficient permissions")
		c.Abort()
	}
}
```

**Cara kerjanya:**
- `RoleMiddleware(model.RoleAdmin)` — cuma admin yang boleh
- `RoleMiddleware(model.RoleAdmin, model.RoleMember)` — admin dan member boleh
- Variadic parameter (`...model.Role`) bikin flexible — bisa 1 atau banyak role

**Perbedaan 401 vs 403:**
- **401 Unauthorized** — "Siapa lu? Belum login atau token invalid"
- **403 Forbidden** — "Gue tau lu siapa, tapi lu ga punya akses"

### Flow Request dengan RBAC

```
Request: DELETE /api/v1/users/5
    │
    ├── AuthMiddleware: Token valid? ── Ya ──→ set userID=2, role="member"
    │
    ├── RoleMiddleware(RoleAdmin): role == "admin"? ── Tidak!
    │
    └── 403 "Insufficient permissions"
```

```
Request: DELETE /api/v1/users/5
    │
    ├── AuthMiddleware: Token valid? ── Ya ──→ set userID=1, role="admin"
    │
    ├── RoleMiddleware(RoleAdmin): role == "admin"? ── Ya!
    │
    └── Handler: Delete user 5 ──→ 200 OK
```

---

## 9. Protecting Routes — Public vs Protected vs Admin-Only

Ini bagian favorit gue. Di sinilah semua puzzle di atas **dirangkai jadi satu**.

### Router Setup

File: **`internal/router/router.go`**

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
		// ============================================
		// PUBLIC ROUTES — Siapa aja bisa akses
		// ============================================
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.Auth.Register)
			auth.POST("/login", h.Auth.Login)
			auth.POST("/refresh", h.Auth.RefreshToken)
		}

		// ============================================
		// PROTECTED ROUTES — Harus login (punya token)
		// ============================================
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Auth
			protected.GET("/auth/me", h.Auth.Me)

			// Users
			protected.GET("/users/profile", h.User.GetProfile)
			protected.PUT("/users/profile", h.User.Update)

			// ============================================
			// ADMIN-ONLY ROUTES — Harus login + role admin
			// ============================================
			admin := protected.Group("")
			admin.Use(middleware.RoleMiddleware(model.RoleAdmin))
			{
				admin.GET("/users", h.User.GetAll)
				admin.DELETE("/users/:id", h.User.Delete)
			}

			// Boards (semua user yang login)
			protected.POST("/boards", h.Board.Create)
			protected.GET("/boards", h.Board.GetMyBoards)
			protected.GET("/boards/:id", h.Board.GetByID)
			protected.PUT("/boards/:id", h.Board.Update)
			protected.DELETE("/boards/:id", h.Board.Delete)

			// Tasks (semua user yang login)
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

### Route Groups Visualized

```
/health                          → PUBLIC    (no middleware)
/api/v1/auth/register            → PUBLIC    (no auth needed)
/api/v1/auth/login               → PUBLIC    (no auth needed)
/api/v1/auth/refresh             → PUBLIC    (no auth needed)
                                   ─────────────────────────
/api/v1/auth/me                  → PROTECTED (AuthMiddleware)
/api/v1/users/profile            → PROTECTED (AuthMiddleware)
/api/v1/boards/*                 → PROTECTED (AuthMiddleware)
/api/v1/tasks/*                  → PROTECTED (AuthMiddleware)
                                   ─────────────────────────
/api/v1/users        GET         → ADMIN     (AuthMiddleware + RoleMiddleware)
/api/v1/users/:id    DELETE      → ADMIN     (AuthMiddleware + RoleMiddleware)
```

### Gin Route Group Pattern

```go
// Group = kumpulan routes yang share middleware
protected := v1.Group("")
protected.Use(middleware.AuthMiddleware(authService))  // <-- semua route di bawah ini butuh auth

// Nested group — admin routes inherit AuthMiddleware + tambah RoleMiddleware
admin := protected.Group("")
admin.Use(middleware.RoleMiddleware(model.RoleAdmin))  // <-- chain: auth + role check
```

**Key insight:** Gin middleware **di-chain**. Route di group `admin` kena 2 middleware:
1. `AuthMiddleware` (dari parent group `protected`)
2. `RoleMiddleware` (dari group `admin` sendiri)

### NestJS Equivalent

```typescript
// NestJS pakai decorator
@Controller('users')
export class UsersController {
  @Get('profile')
  @UseGuards(JwtAuthGuard)  // Protected
  getProfile(@Request() req) { ... }

  @Get()
  @UseGuards(JwtAuthGuard, RolesGuard)  // Admin only
  @Roles('admin')
  getAll() { ... }
}
```

Di NestJS, guard ditempel per-method pakai decorator. Di Go (Gin), pakai **route group** — semua route di dalam group otomatis kena middleware yang sama. Lebih clean kalau punya banyak routes.

---

## 10. Security Best Practices

### Token Expiry

```go
// Access token — PENDEK (15 menit)
ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),

// Refresh token — MEDIUM (7 hari)
ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
```

**Rules:**
- Access token: 5-30 menit. Semakin pendek, semakin aman (tapi semakin sering refresh)
- Refresh token: 1-30 hari. Tergantung app. Banking app? 1 hari. Social media? 30 hari
- **JANGAN** bikin access token yang ga expired. Pernah liat token tanpa `exp`? Red flag

### JWT Secret Management

Di **`internal/config/config.go`**:

```go
JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
```

**Rules:**
- **JANGAN** hardcode secret di code. Pakai environment variable
- Default value `"your-secret-key-change-in-production"` cuma untuk development. Di production, **wajib** set env variable
- Secret minimal **32 karakter**, random. Pakai generator:

```bash
# Generate random secret
openssl rand -hex 32
# Output: a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456
```

- Simpan di `.env` file (jangan commit ke git!) atau secrets manager (AWS Secrets Manager, Vault)

### Secret Rotation

Secret rotation = ganti JWT secret secara berkala.

**Masalahnya:** Kalau ganti secret, semua token yang udah di-generate jadi invalid. Semua user harus login ulang.

**Solusi: Dual secret**

```go
// Pseudocode — support old + new secret
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
    // Coba pakai secret baru dulu
    claims, err := validateWithSecret(tokenString, s.newSecret)
    if err == nil {
        return claims, nil
    }

    // Kalau gagal, coba secret lama (grace period)
    claims, err = validateWithSecret(tokenString, s.oldSecret)
    if err == nil {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}
```

Grace period: support 2 secret selama masa transisi (misal 24 jam), terus matiin secret lama.

### HTTPS

```
HTTP:  Client ──[plain text]──→ Server
       Token bisa disadap (man-in-the-middle)

HTTPS: Client ──[encrypted]──→ Server
       Token aman di perjalanan
```

**Rules:**
- **SELALU** pakai HTTPS di production. Ga ada alasan untuk HTTP
- JWT dikirim di header — tanpa HTTPS, siapa aja di network bisa baca token lu
- Pakai Nginx sebagai reverse proxy dengan SSL termination (bakal dibahas di Bab 14)

### Tambahan Security Tips

| Practice | Kenapa |
|----------|--------|
| **Jangan simpan token di localStorage** | XSS bisa baca localStorage. Pakai httpOnly cookie atau memory |
| **Validate signing method** | Prevent "alg: none" attack (udah kita implement di `ValidateToken`) |
| **Error message generic** | "Invalid email or password" — jangan kasih tau yang mana yang salah |
| **Rate limit login** | Prevent brute force (udah kita implement pakai Redis) |
| **Hash password, jangan encrypt** | Hash = one-way. Encrypt = bisa di-decrypt. Password harus one-way |
| **Blacklist token saat logout** | Simpan token yang di-revoke di Redis dengan TTL = sisa expiry |
| **CORS config** | Restrict origin yang boleh akses API |

### Token Blacklist (Bonus Pattern)

Kalau mau implement logout yang bener:

```go
// Pseudocode — simpan di Redis
func (s *AuthService) Logout(token string) error {
    claims, err := s.ValidateToken(token)
    if err != nil {
        return err
    }

    // Hitung sisa waktu expiry
    ttl := time.Until(claims.ExpiresAt.Time)

    // Simpan di Redis sebagai blacklist
    return s.redis.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}

// Di ValidateToken, tambah check:
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
    // ... existing validation ...

    // Check blacklist
    _, err = s.redis.Get(ctx, "blacklist:"+tokenString).Result()
    if err == nil {
        return nil, errors.New("token has been revoked")
    }

    return claims, nil
}
```

---

## 11. Full Wiring — Dari main.go Sampai Jalan

Biar lu ngerti gimana semuanya nyambung, ini flow lengkap dari `main.go`:

File: **`cmd/api/main.go`**

```go
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
)

func main() {
	// 1. Load config (termasuk JWT_SECRET dari env)
	cfg := config.Load()

	// 2. Connect database
	db := config.ConnectDB(cfg)

	// 3. Auto-migrate (buat table User, Board, Task)
	db.AutoMigrate(&model.User{}, &model.Board{}, &model.Task{})

	// 4. Connect Redis
	redisClient := config.ConnectRedis(cfg)

	// 5. Setup repositories
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// 6. Setup services (inject repo + secret)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	boardService := service.NewBoardService(boardRepo)
	taskService := service.NewTaskService(taskRepo, boardRepo)

	// 7. Setup handlers (inject service)
	handlers := &router.Handlers{
		Auth:  handler.NewAuthHandler(authService),
		User:  handler.NewUserHandler(userService),
		Board: handler.NewBoardHandler(boardService),
		Task:  handler.NewTaskHandler(taskService),
	}

	// 8. Setup Gin + routes + middleware
	r := gin.Default()
	router.Setup(r, handlers, authService, redisClient)

	// 9. Start server
	r.Run(":" + cfg.Port)
}
```

**Dependency graph:**

```
Config (env variables)
    │
    ├──→ Database (PostgreSQL) ──→ Repository ──→ Service ──→ Handler
    │                                              │
    ├──→ Redis ──→ RateLimiter middleware           │
    │                                              │
    └──→ JWTSecret ──→ AuthService ──→ AuthMiddleware
                            │
                            └──→ AuthHandler
```

Semuanya **explicit**. Ga ada magic DI container. Lu bisa trace dari `main.go` ke mana aja dependency-nya pergi.

---

## 12. NestJS/Passport vs Go — Full Comparison

### Setup Comparison

**NestJS (banyak file + decorator + config):**

```typescript
// 1. Install
npm install @nestjs/passport @nestjs/jwt passport passport-jwt bcryptjs

// 2. JwtStrategy (file baru)
@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy) {
  constructor(private configService: ConfigService) {
    super({
      jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
      ignoreExpiration: false,
      secretOrKey: configService.get('JWT_SECRET'),
    });
  }
  async validate(payload: any) {
    return { userId: payload.sub, email: payload.email, role: payload.role };
  }
}

// 3. AuthModule (register strategy, JWT module)
@Module({
  imports: [
    PassportModule,
    JwtModule.registerAsync({
      useFactory: (config: ConfigService) => ({
        secret: config.get('JWT_SECRET'),
        signOptions: { expiresIn: '15m' },
      }),
      inject: [ConfigService],
    }),
  ],
  providers: [AuthService, JwtStrategy],
  controllers: [AuthController],
})
export class AuthModule {}

// 4. AuthService
@Injectable()
export class AuthService {
  constructor(
    private jwtService: JwtService,
    private usersService: UsersService,
  ) {}

  async login(user: User) {
    const payload = { sub: user.id, email: user.email, role: user.role };
    return { access_token: this.jwtService.sign(payload) };
  }
}

// 5. Guard pakai di controller
@UseGuards(JwtAuthGuard)
@Get('profile')
getProfile(@Request() req) { return req.user; }

// 6. Roles Guard (file baru lagi)
@Injectable()
export class RolesGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const requiredRoles = this.reflector.get<string[]>('roles', context.getHandler());
    const { user } = context.switchToHttp().getRequest();
    return requiredRoles.some((role) => user.role === role);
  }
}
```

**Go (straightforward, no magic):**

```go
// 1. Install
// go get github.com/golang-jwt/jwt/v5
// go get golang.org/x/crypto

// 2. Auth Service — generate & validate token
func (s *AuthService) generateTokenPair(user *model.User) (*TokenPair, error) {
    claims := &JWTClaims{UserID: user.ID, Email: user.Email, Role: user.Role, ...}
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(s.jwtSecret))
    // ...
}

// 3. Auth Middleware — extract & validate
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractBearerToken(c)
        claims, err := authService.ValidateToken(token)
        c.Set("userID", claims.UserID)
        c.Next()
    }
}

// 4. Role Middleware
func RoleMiddleware(roles ...model.Role) gin.HandlerFunc { ... }

// 5. Router — apply middleware ke group
protected.Use(middleware.AuthMiddleware(authService))
admin.Use(middleware.RoleMiddleware(model.RoleAdmin))
```

### Side-by-Side Summary

| Aspek | NestJS/Passport | Go |
|-------|----------------|-----|
| Dependencies | 5+ packages | 2 packages (`golang-jwt`, `bcrypt`) |
| Files for auth | 5-7 files (Strategy, Guard, Module, Service, Controller, Decorator) | 3-4 files (Service, Handler, Middleware, Model) |
| Token generation | `JwtService.sign()` (abstracted) | `jwt.NewWithClaims()` + `SignedString()` (explicit) |
| Token validation | Passport does it (hidden) | Kita tulis sendiri di `ValidateToken()` |
| Guard/Middleware | `@UseGuards(JwtAuthGuard)` decorator | `group.Use(middleware.AuthMiddleware())` |
| Role check | `@Roles('admin')` + `RolesGuard` + `Reflector` | `middleware.RoleMiddleware(model.RoleAdmin)` |
| Magic level | Tinggi (DI, decorator, reflection) | Rendah (explicit everything) |
| Debugging | Kadang bingung di mana error-nya | Stack trace jelas |

**Verdict:**
- NestJS lebih sedikit code kalau udah setup. Tapi debugging bisa bikin pusing karena banyak abstraction layer
- Go lebih verbose tapi **crystal clear**. Lu tau persis apa yang terjadi di setiap step. Ga ada hidden behavior

---

## 13. Response Helper

Biar response API konsisten, kita punya helper:

File: **`pkg/response/json.go`**

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

Semua response punya format yang sama:

```json
// Success
{
  "status": 200,
  "message": "Login successful",
  "data": { ... }
}

// Error
{
  "status": 401,
  "message": "Invalid or expired token"
}
```

---

## 14. Testing Auth Endpoints

### Full cURL Test Flow

```bash
# ============================================
# 1. REGISTER
# ============================================
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin User","email":"admin@test.com","password":"admin123"}' | jq

# ============================================
# 2. LOGIN — simpan token
# ============================================
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"admin123"}')

ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.refresh_token')

echo "Access Token: $ACCESS_TOKEN"

# ============================================
# 3. PROTECTED ROUTE — pakai token
# ============================================
curl -s http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq

# ============================================
# 4. TANPA TOKEN — harus 401
# ============================================
curl -s http://localhost:8080/api/v1/auth/me | jq
# {"status":401,"message":"Authorization header required"}

# ============================================
# 5. TOKEN INVALID — harus 401
# ============================================
curl -s http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer invalid-token-here" | jq
# {"status":401,"message":"Invalid or expired token"}

# ============================================
# 6. ADMIN ROUTE (kalau role member) — harus 403
# ============================================
curl -s http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
# {"status":403,"message":"Insufficient permissions"}

# ============================================
# 7. REFRESH TOKEN
# ============================================
curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" | jq
```

---

## 15. Common Mistakes & Troubleshooting

### Mistake 1: Password ke-expose di response

```go
// SALAH — langsung return user entity
response.Success(c, http.StatusCreated, "User registered", user)
// Password hash ikut ke-return karena field Password ada di struct

// BENAR — pakai DTO
resp := user.ToResponse()  // UserResponse tanpa Password
response.Success(c, http.StatusCreated, "User registered", &resp)
```

Tapi kita udah handle ini pakai `json:"-"` di model:
```go
Password string `json:"-" gorm:"size:255;not null"` // json:"-" = never serialize
```

Double protection — DTOnya ga ada field Password, dan model-nya juga ada `json:"-"`.

### Mistake 2: Lupa c.Abort() di middleware

```go
// SALAH — request tetap lanjut ke handler!
if authHeader == "" {
    response.Error(c, http.StatusUnauthorized, "No token")
    return // return aja ga cukup
}

// BENAR — abort chain
if authHeader == "" {
    response.Error(c, http.StatusUnauthorized, "No token")
    c.Abort() // <-- ini yang stop request
    return
}
```

### Mistake 3: Hardcode JWT secret

```go
// SALAH
jwtSecret := "my-super-secret-key-123"

// BENAR — dari environment
jwtSecret := os.Getenv("JWT_SECRET")
// Atau pakai config loader seperti di project kita:
JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
```

### Mistake 4: Ga check signing method

```go
// SALAH — vulnerable ke algorithm confusion attack
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte(secret), nil
})

// BENAR — explicitly check algorithm
token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, errors.New("unexpected signing method")
    }
    return []byte(secret), nil
})
```

### Mistake 5: Specific error messages saat login

```go
// SALAH — attacker bisa enumerate email
if err != nil {
    return nil, errors.New("email not found")      // Email enumeration!
}
if !user.CheckPassword(req.Password) {
    return nil, errors.New("wrong password")        // Confirms email exists!
}

// BENAR — generic message
return nil, errors.New("invalid email or password") // Ga tau yang mana yang salah
```

---

## Checklist Bab Ini

- [ ] Paham perbedaan Session vs JWT dan kenapa JWT untuk API
- [ ] Paham struktur JWT (header, payload, signature)
- [ ] Paham kenapa bcrypt, bukan MD5/SHA untuk password
- [ ] Register flow jalan — validate, hash, save, return
- [ ] Login flow jalan — find user, check password, generate token pair
- [ ] Refresh token flow jalan — validate refresh, generate new pair
- [ ] JWT middleware bisa extract token, validate, set context
- [ ] RBAC middleware bisa check role (admin vs member)
- [ ] Routes terproteksi dengan benar (public vs protected vs admin-only)
- [ ] Paham security best practices (expiry, secret rotation, HTTPS)
- [ ] Bisa test semua endpoint pakai cURL
- [ ] Paham bedanya implementasi Go vs NestJS/Passport

---

> **Next:** [Bab 8: Caching dengan Redis](08-caching-redis.md) — kita bakal pakai Redis yang udah di-setup di project ini untuk caching, rate limiting, dan session management.
