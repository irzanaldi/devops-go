# Bab 10: Testing di Go — Dari Unit Test Sampai Integration Test

## Overview

Testing itu bukan "nice to have" — itu **kebutuhan**. Kalau di NestJS kamu udah terbiasa sama Jest, describe/it/expect, mocking pakai `jest.mock()`, dan test runner yang fancy — di Go, semuanya **built-in** dan jauh lebih **minimalis**. Ga perlu install test framework terpisah. Ga ada magic. Tapi powerful.

Bab ini bakal cover **semua** yang kamu perlu tau tentang testing di Go, dari yang paling basic sampai integration test. Semua contoh diambil langsung dari project Task Manager API kita. Bukan contoh mainan.

> **Mental model shift:** Di Jest, kamu describe/it/expect. Di Go, kamu bikin function `TestXxx`, panggil function yang mau di-test, terus check hasilnya pakai `if`. Itu aja. Simpel banget.

---

## 1. Kenapa Testing Penting

### Bukan Cuma Teori

"Testing itu penting" — udah sering denger kan? Tapi kenapa **beneran** penting? Ini real-world scenario:

**Scenario 1: Bug Prevention**

```
Kamu bikin fitur Register. Jalan lancar.
Seminggu kemudian, teammate refactor UserRepository — ganti return type.
AuthService.Register tiba-tiba broken.
Tanpa test → bug masuk production, user ga bisa register.
Dengan test → CI gagal, PR ditolak, bug ketangkep sebelum merge.
```

**Scenario 2: Refactor Confidence**

```
Kamu mau refactor AuthService — pecah generateTokenPair jadi 2 function.
Takut break sesuatu? Kalau ada test suite yang cover semua case,
tinggal jalanin `go test ./...` — kalau hijau semua, aman.
```

**Scenario 3: Documentation yang Hidup**

```go
func TestAuthService_Login_InvalidPassword(t *testing.T) {
    // Test ini sekaligus jadi dokumentasi:
    // "Kalau password salah, Login harus return error 'invalid email or password'"
}
```

Test file itu **dokumentasi yang selalu up-to-date**. Beda sama README yang bisa basi.

**Scenario 4: CI/CD Requirement**

```yaml
# Di pipeline CI/CD, kalau test gagal → deploy dibatalin
- name: Run Tests
  run: go test ./... -v
  # Kalau ada 1 aja test fail → pipeline merah → ga jadi deploy
```

### Rangkuman

| Alasan | Penjelasan |
|--------|-----------|
| **Bug prevention** | Tangkep bug sebelum masuk production |
| **Refactor confidence** | Ubah kode internal tanpa takut break |
| **Living documentation** | Test = spesifikasi yang executable |
| **CI/CD gate** | Ga lolos test = ga boleh deploy |
| **Better design** | Kode yang gampang di-test = kode yang well-designed |

---

## 2. Testing di Go vs Jest

### Perbedaan Fundamental

Di JavaScript/NestJS, testing butuh **ekosistem terpisah**: Jest, Mocha, Chai, Sinon, dll. Di Go? **Semua udah built-in**.

### Comparison Table

| Aspek | Jest (NestJS) | Go Testing |
|-------|--------------|------------|
| **Test runner** | Jest (install terpisah) | `go test` (built-in) |
| **Assertion** | `expect().toBe()` | `if` + `t.Errorf()` (manual) |
| **Mocking** | `jest.mock()`, `jest.fn()` | Interface + manual mock |
| **Test file** | `*.spec.ts` atau `*.test.ts` | `*_test.go` |
| **Test function** | `describe/it` | `func TestXxx(t *testing.T)` |
| **Config** | `jest.config.ts` | Ga perlu config |
| **Coverage** | `jest --coverage` | `go test -cover` |
| **Benchmark** | External (benchmark.js) | Built-in `testing.B` |
| **Setup/Teardown** | `beforeEach/afterEach` | `TestMain` atau manual |
| **Subtests** | `describe/it` nesting | `t.Run()` |
| **Parallel** | `--runInBand` / default parallel | `t.Parallel()` explicit |
| **Watch mode** | `jest --watch` | External (air, watchexec) |
| **Snapshot testing** | Built-in | External / golden files |

### File Naming Convention

```
NestJS:
  auth.service.ts        → source
  auth.service.spec.ts   → test (atau auth.service.test.ts)

Go:
  auth.go                → source
  auth_test.go           → test (HARUS suffix _test.go)
```

Di Go, file test **harus** punya suffix `_test.go`. Ini bukan convention — ini **enforced** oleh compiler. File `_test.go` ga akan masuk ke build binary.

### Menjalankan Test

```bash
# Jest
npx jest                          # run semua
npx jest auth.service.spec.ts     # run satu file
npx jest --watch                  # watch mode
npx jest --coverage               # dengan coverage

# Go
go test ./...                     # run semua test di semua package
go test ./internal/service/       # run test di satu package
go test -run TestAuthService      # run test yang match nama
go test -v ./...                  # verbose output
go test -cover ./...              # dengan coverage
```

### Anatomi Test

**Jest:**
```typescript
describe('AuthService', () => {
  describe('login', () => {
    it('should return tokens for valid credentials', () => {
      const result = authService.login(validCredentials);
      expect(result.accessToken).toBeDefined();
    });

    it('should throw for invalid password', () => {
      expect(() => authService.login(invalidCredentials))
        .toThrow('invalid email or password');
    });
  });
});
```

**Go:**
```go
func TestAuthService_Login_ValidCredentials(t *testing.T) {
    result, err := authService.Login(validCredentials)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if result.AccessToken == "" {
        t.Error("expected access token, got empty string")
    }
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
    _, err := authService.Login(invalidCredentials)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if err.Error() != "invalid email or password" {
        t.Errorf("expected 'invalid email or password', got '%s'", err.Error())
    }
}
```

Lihat bedanya? Go ga punya `describe/it`. Satu test = satu function. Naming convention pake underscore: `TestNamaStruct_Method_Scenario`.

---

## 3. Unit Test Basics

### testing.T — Si Bos Test

Semua test function di Go terima parameter `*testing.T`. Ini "test context" yang punya method-method untuk report hasil test.

### Method-method testing.T

```go
package service

import "testing"

func TestBasicMethods(t *testing.T) {
    // === t.Log / t.Logf ===
    // Print info (cuma muncul kalau -v atau test fail)
    t.Log("ini info biasa")
    t.Logf("user ID: %d", 42)

    // === t.Error / t.Errorf ===
    // Mark test FAIL, tapi LANJUT eksekusi
    got := 2 + 2
    want := 5
    if got != want {
        t.Errorf("got %d, want %d", got, want) // test fail, tapi lanjut ke baris bawah
    }
    t.Log("ini masih jalan meskipun error di atas")

    // === t.Fatal / t.Fatalf ===
    // Mark test FAIL dan STOP eksekusi function ini
    result := someFunction()
    if result == nil {
        t.Fatal("result is nil, can't continue") // STOP di sini
    }
    // Baris ini ga akan jalan kalau Fatal di atas kena
    t.Log(result.Name)

    // === t.Skip / t.Skipf ===
    // Skip test (misal: butuh DB yang ga available)
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
}
```

**Kapan pake Error vs Fatal?**

```
t.Error/t.Errorf → "Test fail, tapi masih bisa check hal lain"
                    Cocok untuk assertion yang independent

t.Fatal/t.Fatalf → "Test fail, ga ada gunanya lanjut"
                    Cocok kalau result nil dan mau akses field-nya
```

**Jest equivalent:**

| Go | Jest |
|----|------|
| `t.Error("...")` | `expect(x).toBe(y)` (test lanjut) |
| `t.Fatal("...")` | `throw` di test (test stop) |
| `t.Skip("...")` | `it.skip('...')` atau `xit('...')` |
| `t.Log("...")` | `console.log()` |

### t.Run — Subtests

`t.Run()` itu equivalent-nya `describe/it` nesting di Jest. Bikin grup test di dalam satu function.

```go
func TestUser_HashPassword(t *testing.T) {
    t.Run("should hash password successfully", func(t *testing.T) {
        user := &model.User{Password: "secret123"}
        err := user.HashPassword()
        if err != nil {
            t.Fatalf("expected no error, got %v", err)
        }
        if user.Password == "secret123" {
            t.Error("password should be hashed, not plain text")
        }
    })

    t.Run("should produce different hash for same password", func(t *testing.T) {
        user1 := &model.User{Password: "secret123"}
        user2 := &model.User{Password: "secret123"}

        user1.HashPassword()
        user2.HashPassword()

        if user1.Password == user2.Password {
            t.Error("same password should produce different hashes (bcrypt salt)")
        }
    })

    t.Run("hashed password should be verifiable", func(t *testing.T) {
        user := &model.User{Password: "secret123"}
        user.HashPassword()

        if !user.CheckPassword("secret123") {
            t.Error("CheckPassword should return true for correct password")
        }
    })
}
```

**Jest equivalent:**
```typescript
describe('User.hashPassword', () => {
  it('should hash password successfully', () => {
    // ...
  });

  it('should produce different hash for same password', () => {
    // ...
  });

  it('hashed password should be verifiable', () => {
    // ...
  });
});
```

### t.Parallel — Run Test Secara Paralel

```go
func TestParallelExample(t *testing.T) {
    t.Run("test A", func(t *testing.T) {
        t.Parallel() // <-- ini jalan paralel dengan test B
        // ... test logic
    })

    t.Run("test B", func(t *testing.T) {
        t.Parallel() // <-- ini jalan paralel dengan test A
        // ... test logic
    })
}
```

Di Jest, test **default paralel** (per file). Di Go, test **default sequential**. Kamu harus explicitly call `t.Parallel()` kalau mau paralel.

**Kapan pakai `t.Parallel()`?**
- Test yang **independent** (ga share state)
- Test yang lambat (bisa jalan barengan untuk speed up)
- **Jangan** pakai kalau test share database/state yang sama

### Menjalankan Test

```bash
# Run test function yang namanya match "HashPassword"
go test -run TestUser_HashPassword ./internal/model/

# Run specific subtest
go test -run TestUser_HashPassword/should_hash_password_successfully ./internal/model/

# Verbose — lihat semua test name + PASS/FAIL
go test -v ./internal/model/

# Short mode — skip test yang di-mark t.Skip kalau testing.Short()
go test -short ./...
```

---

## 4. Table-Driven Tests — THE Go Pattern

Ini pattern **paling idiomatic** di Go. Kalau kamu cuma belajar satu hal dari bab ini, belajar ini. Seriously.

### Konsep

Daripada nulis banyak test function yang mirip-mirip, kamu bikin **slice of test cases**, terus loop dan jalanin satu-satu.

### Contoh: Test CheckPassword dari Project Kita

```go
// File: internal/model/user_test.go
package model

import "testing"

func TestUser_CheckPassword(t *testing.T) {
    // Setup: bikin user dengan hashed password
    user := &User{Password: "secret123"}
    if err := user.HashPassword(); err != nil {
        t.Fatalf("failed to hash password: %v", err)
    }

    // Table-driven tests
    tests := []struct {
        name     string
        password string
        want     bool
    }{
        {
            name:     "correct password",
            password: "secret123",
            want:     true,
        },
        {
            name:     "wrong password",
            password: "wrongpassword",
            want:     false,
        },
        {
            name:     "empty password",
            password: "",
            want:     false,
        },
        {
            name:     "similar password",
            password: "secret124",
            want:     false,
        },
        {
            name:     "password with extra space",
            password: "secret123 ",
            want:     false,
        },
        {
            name:     "case sensitive",
            password: "Secret123",
            want:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := user.CheckPassword(tt.password)
            if got != tt.want {
                t.Errorf("CheckPassword(%q) = %v, want %v", tt.password, got, tt.want)
            }
        })
    }
}
```

### Kenapa Table-Driven Itu Idiomatic Go?

1. **DRY** — logic test ditulis sekali, case tinggal tambah di table
2. **Readable** — semua test case visible di satu tempat
3. **Easy to add** — mau nambah case? Tambahin struct baru di slice
4. **Consistent naming** — setiap case punya `name` yang muncul di output
5. **Go standard library sendiri pake pattern ini**

### Output-nya

```bash
$ go test -v -run TestUser_CheckPassword ./internal/model/
=== RUN   TestUser_CheckPassword
=== RUN   TestUser_CheckPassword/correct_password
=== RUN   TestUser_CheckPassword/wrong_password
=== RUN   TestUser_CheckPassword/empty_password
=== RUN   TestUser_CheckPassword/similar_password
=== RUN   TestUser_CheckPassword/password_with_extra_space
=== RUN   TestUser_CheckPassword/case_sensitive
--- PASS: TestUser_CheckPassword (0.25s)
    --- PASS: TestUser_CheckPassword/correct_password (0.04s)
    --- PASS: TestUser_CheckPassword/wrong_password (0.04s)
    --- PASS: TestUser_CheckPassword/empty_password (0.04s)
    --- PASS: TestUser_CheckPassword/similar_password (0.04s)
    --- PASS: TestUser_CheckPassword/password_with_extra_space (0.04s)
    --- PASS: TestUser_CheckPassword/case_sensitive (0.04s)
PASS
```

### Contoh: Test ToResponse dari User Model

```go
func TestUser_ToResponse(t *testing.T) {
    tests := []struct {
        name string
        user User
        want UserResponse
    }{
        {
            name: "member user",
            user: User{
                ID:    1,
                Name:  "Irzan Aldi",
                Email: "irzan@test.com",
                Role:  RoleMember,
            },
            want: UserResponse{
                ID:    1,
                Name:  "Irzan Aldi",
                Email: "irzan@test.com",
                Role:  RoleMember,
            },
        },
        {
            name: "admin user",
            user: User{
                ID:    2,
                Name:  "Admin",
                Email: "admin@test.com",
                Role:  RoleAdmin,
            },
            want: UserResponse{
                ID:    2,
                Name:  "Admin",
                Email: "admin@test.com",
                Role:  RoleAdmin,
            },
        },
        {
            name: "password should not leak",
            user: User{
                ID:       3,
                Name:     "Test",
                Email:    "test@test.com",
                Password: "should-not-appear",
                Role:     RoleMember,
            },
            want: UserResponse{
                ID:    3,
                Name:  "Test",
                Email: "test@test.com",
                Role:  RoleMember,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.user.ToResponse()
            if got.ID != tt.want.ID {
                t.Errorf("ID = %d, want %d", got.ID, tt.want.ID)
            }
            if got.Name != tt.want.Name {
                t.Errorf("Name = %s, want %s", got.Name, tt.want.Name)
            }
            if got.Email != tt.want.Email {
                t.Errorf("Email = %s, want %s", got.Email, tt.want.Email)
            }
            if got.Role != tt.want.Role {
                t.Errorf("Role = %s, want %s", got.Role, tt.want.Role)
            }
        })
    }
}
```

### Jest Equivalent: describe.each

```typescript
// Jest punya describe.each / test.each — mirip tapi beda syntax
describe.each([
  { password: 'secret123', expected: true },
  { password: 'wrong', expected: false },
  { password: '', expected: false },
])('checkPassword($password)', ({ password, expected }) => {
  it(`should return ${expected}`, () => {
    expect(user.checkPassword(password)).toBe(expected);
  });
});
```

Konsepnya sama — define data, loop, test. Tapi Go punya kontrol lebih karena kamu tulis loop sendiri.

---

## 5. Testify Library — Assertion yang Lebih Enak

### Masalah Testing Tanpa Assertion Library

Go standard library ga punya assertion. Kamu harus nulis `if-then-t.Error` manual:

```go
// Vanilla Go — verbose
if got != want {
    t.Errorf("got %v, want %v", got, want)
}
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
if result == nil {
    t.Fatal("result should not be nil")
}
```

Ini jadi repetitif. Masuk **testify**.

### Install Testify

```bash
go get github.com/stretchr/testify
```

### assert vs require

Testify punya dua package utama:

| Package | Behavior | Equivalent |
|---------|----------|------------|
| `assert` | Test fail tapi **lanjut** | `t.Error()` |
| `require` | Test fail dan **stop** | `t.Fatal()` |

```go
package model

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUser_HashPassword_WithTestify(t *testing.T) {
    user := &User{Password: "secret123"}

    // require — kalau gagal, STOP di sini
    err := user.HashPassword()
    require.NoError(t, err)                         // kalau error → stop
    require.NotEmpty(t, user.Password)               // kalau empty → stop

    // assert — kalau gagal, LANJUT check yang lain
    assert.NotEqual(t, "secret123", user.Password)   // password harus udah di-hash
    assert.True(t, user.CheckPassword("secret123"))  // harus bisa verify
    assert.False(t, user.CheckPassword("wrong"))     // password salah harus false
    assert.Len(t, user.Password, 60)                 // bcrypt hash selalu 60 char
}
```

### assert Method Cheat Sheet

| Testify | Vanilla Go | Jest |
|---------|-----------|------|
| `assert.Equal(t, want, got)` | `if got != want { t.Error }` | `expect(got).toBe(want)` |
| `assert.NotEqual(t, a, b)` | `if a == b { t.Error }` | `expect(a).not.toBe(b)` |
| `assert.Nil(t, obj)` | `if obj != nil { t.Error }` | `expect(obj).toBeNull()` |
| `assert.NotNil(t, obj)` | `if obj == nil { t.Error }` | `expect(obj).not.toBeNull()` |
| `assert.True(t, val)` | `if !val { t.Error }` | `expect(val).toBe(true)` |
| `assert.False(t, val)` | `if val { t.Error }` | `expect(val).toBe(false)` |
| `assert.NoError(t, err)` | `if err != nil { t.Error }` | Ga ada langsung |
| `assert.Error(t, err)` | `if err == nil { t.Error }` | `expect(() => ...).toThrow()` |
| `assert.Contains(t, str, sub)` | `if !strings.Contains { t.Error }` | `expect(str).toContain(sub)` |
| `assert.Len(t, slice, 3)` | `if len(slice) != 3 { t.Error }` | `expect(arr).toHaveLength(3)` |
| `assert.Empty(t, slice)` | `if len(slice) != 0 { t.Error }` | `expect(arr).toHaveLength(0)` |
| `assert.EqualError(t, err, "msg")` | `if err.Error() != "msg" { t.Error }` | `expect(fn).toThrow("msg")` |
| `assert.Greater(t, a, b)` | `if a <= b { t.Error }` | `expect(a).toBeGreaterThan(b)` |
| `assert.WithinDuration(t, a, b, d)` | Time comparison manual | Ga ada langsung |
| `require.NoError(t, err)` | `if err != nil { t.Fatal }` | Ga ada (auto-throw) |

### Table-Driven Test dengan Testify

Kombinasi terbaik — table-driven + testify:

```go
func TestUser_CheckPassword_Testify(t *testing.T) {
    user := &User{Password: "secret123"}
    require.NoError(t, user.HashPassword())

    tests := []struct {
        name     string
        password string
        want     bool
    }{
        {"correct password", "secret123", true},
        {"wrong password", "wrong", false},
        {"empty string", "", false},
        {"case sensitive", "Secret123", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := user.CheckPassword(tt.password)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

Jauh lebih clean kan?

### Testify Suite (Optional)

Testify juga punya `suite` package — mirip test class di Jest dengan `beforeEach/afterEach`:

```go
package service

import (
    "testing"

    "github.com/stretchr/testify/suite"
)

type AuthServiceSuite struct {
    suite.Suite
    service *AuthService
    // mock dependencies
}

// SetupTest jalan sebelum SETIAP test method — mirip beforeEach
func (s *AuthServiceSuite) SetupTest() {
    // setup fresh instance
    s.service = NewAuthService(mockRepo, "test-secret")
}

// TearDownTest jalan setelah SETIAP test method — mirip afterEach
func (s *AuthServiceSuite) TearDownTest() {
    // cleanup
}

func (s *AuthServiceSuite) TestRegister_Success() {
    // s.Assert() otomatis ada dari suite.Suite
    result, err := s.service.Register(validReq)
    s.NoError(err)
    s.NotNil(result)
    s.Equal("irzan@test.com", result.Email)
}

func (s *AuthServiceSuite) TestLogin_InvalidPassword() {
    _, err := s.service.Login(invalidReq)
    s.Error(err)
    s.EqualError(err, "invalid email or password")
}

// Entry point — wajib ada
func TestAuthServiceSuite(t *testing.T) {
    suite.Run(t, new(AuthServiceSuite))
}
```

**Jest equivalent:**
```typescript
describe('AuthService', () => {
  let service: AuthService;

  beforeEach(() => {
    service = new AuthService(mockRepo, 'test-secret');
  });

  afterEach(() => {
    // cleanup
  });

  it('register success', () => { ... });
  it('login invalid password', () => { ... });
});
```

> **Rekomendasi:** Untuk project kecil-menengah, table-driven + testify assert/require udah cukup. Suite overkill kalau test-nya simpel.

---

## 6. Mocking — Interface-Based, No Magic

### Perbedaan Fundamental dengan Jest

Di Jest, mocking itu **magic**:

```typescript
// Jest — mock seluruh module dengan satu baris
jest.mock('./user.repository');
const mockRepo = jest.mocked(UserRepository);
mockRepo.findByEmail.mockResolvedValue(mockUser);
```

Di Go? **Ga ada magic**. Mocking dilakukan lewat **interface**. Kalau code kamu depend on concrete struct, kamu ga bisa mock. Kalau depend on interface, kamu bisa inject mock implementation.

### Step 1: Define Interface

Saat ini, `AuthService` di project kita depend on concrete `*repository.UserRepository`:

```go
type AuthService struct {
    userRepo  *repository.UserRepository  // ← concrete type, susah di-mock
    jwtSecret string
}
```

Untuk bikin testable, kita perlu **interface**:

```go
// File: internal/repository/interfaces.go
package repository

import "github.com/irzan-aldi/task-manager/internal/model"

// UserRepositoryInterface — kontrak yang harus dipenuhi
type UserRepositoryInterface interface {
    Create(user *model.User) error
    FindByID(id uint) (*model.User, error)
    FindByEmail(email string) (*model.User, error)
    FindAll(page, limit int) ([]model.User, int64, error)
    Update(user *model.User) error
    Delete(id uint) error
}
```

Terus update `AuthService` untuk terima interface:

```go
// File: internal/service/auth.go (versi testable)
type AuthService struct {
    userRepo  repository.UserRepositoryInterface  // ← interface, bisa di-mock!
    jwtSecret string
}

func NewAuthService(userRepo repository.UserRepositoryInterface, jwtSecret string) *AuthService {
    return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}
```

`UserRepository` struct yang udah ada **otomatis implement** interface ini karena semua method-nya cocok. Ini duck typing — ga perlu `implements` keyword.

### Step 2: Manual Mock

Cara paling basic — bikin struct yang implement interface:

```go
// File: internal/service/auth_test.go
package service

import (
    "errors"

    "github.com/irzan-aldi/task-manager/internal/model"
)

// MockUserRepository — manual mock implementation
type MockUserRepository struct {
    Users    map[string]*model.User  // simulated storage
    CreateFn func(user *model.User) error
}

func NewMockUserRepository() *MockUserRepository {
    return &MockUserRepository{
        Users: make(map[string]*model.User),
    }
}

func (m *MockUserRepository) Create(user *model.User) error {
    if m.CreateFn != nil {
        return m.CreateFn(user)
    }
    user.ID = uint(len(m.Users) + 1)
    m.Users[user.Email] = user
    return nil
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
    user, ok := m.Users[email]
    if !ok {
        return nil, errors.New("user not found")
    }
    return user, nil
}

func (m *MockUserRepository) FindByID(id uint) (*model.User, error) {
    for _, u := range m.Users {
        if u.ID == id {
            return u, nil
        }
    }
    return nil, errors.New("user not found")
}

func (m *MockUserRepository) FindAll(page, limit int) ([]model.User, int64, error) {
    var users []model.User
    for _, u := range m.Users {
        users = append(users, *u)
    }
    return users, int64(len(users)), nil
}

func (m *MockUserRepository) Update(user *model.User) error {
    m.Users[user.Email] = user
    return nil
}

func (m *MockUserRepository) Delete(id uint) error {
    for email, u := range m.Users {
        if u.ID == id {
            delete(m.Users, email)
            return nil
        }
    }
    return errors.New("user not found")
}
```

### Step 3: Testify Mock (Lebih Powerful)

Manual mock oke untuk case simpel. Tapi kalau kamu mau **verify** apakah method dipanggil, berapa kali, dengan argument apa — pakai `testify/mock`:

```go
// File: internal/mocks/user_repository.go
package mocks

import (
    "github.com/stretchr/testify/mock"
    "github.com/irzan-aldi/task-manager/internal/model"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (*model.User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindAll(page, limit int) ([]model.User, int64, error) {
    args := m.Called(page, limit)
    return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Update(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
    args := m.Called(id)
    return args.Error(0)
}
```

**Cara pakainya:**

```go
func TestWithTestifyMock(t *testing.T) {
    mockRepo := new(mocks.MockUserRepository)

    // Setup expectation — "kalau FindByEmail dipanggil dengan 'irzan@test.com', return user ini"
    mockRepo.On("FindByEmail", "irzan@test.com").Return(&model.User{
        ID:    1,
        Email: "irzan@test.com",
        Name:  "Irzan",
    }, nil)

    // Panggil function yang di-test
    authService := NewAuthService(mockRepo, "test-secret")
    // ... test logic ...

    // Verify — pastikan method beneran dipanggil
    mockRepo.AssertExpectations(t)
    mockRepo.AssertCalled(t, "FindByEmail", "irzan@test.com")
    mockRepo.AssertNumberOfCalls(t, "FindByEmail", 1)
}
```

### Comparison: Go Mock vs Jest Mock

| Aspek | Jest | Go (testify/mock) |
|-------|------|-------------------|
| **Setup** | `jest.mock('./repo')` | Define mock struct + `mock.Mock` |
| **Return value** | `mockFn.mockReturnValue(x)` | `m.On("Method", args).Return(x)` |
| **Verify called** | `expect(fn).toHaveBeenCalled()` | `m.AssertCalled(t, "Method", args)` |
| **Call count** | `expect(fn).toHaveBeenCalledTimes(2)` | `m.AssertNumberOfCalls(t, "Method", 2)` |
| **Arguments** | `expect(fn).toHaveBeenCalledWith(arg)` | `m.AssertCalled(t, "Method", arg)` |
| **Magic level** | Tinggi (proxy whole module) | Rendah (explicit interface) |
| **Effort** | Rendah | Tinggi (tapi lebih safe) |

**Key insight:** Di Go, kemampuan mock itu **consequence of good design**. Kalau code kamu susah di-mock, itu tanda design-nya perlu diperbaiki (too coupled). Di Jest, kamu bisa mock apa aja — termasuk code yang design-nya jelek.

---

## 7. Testing Repository Layer

### Opsi Testing Repository

| Opsi | Pros | Cons |
|------|------|------|
| **Mock DB (sqlmock)** | Cepat, isolated, ga perlu DB | Ga test query beneran |
| **Testcontainers** | Test dengan DB beneran (Docker) | Butuh Docker, lebih lambat |
| **SQLite in-memory** | Cepat, real SQL | Beda behavior dari PostgreSQL |
| **Separate test DB** | Real PostgreSQL | Setup manual, cleanup ribet |

### Contoh: Test UserRepository dengan Mock Manual

Karena `UserRepository` depend on `*gorm.DB`, kita bisa test pakai SQLite in-memory sebagai pengganti PostgreSQL:

```go
// File: internal/repository/user_test.go
package repository

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/irzan-aldi/task-manager/internal/model"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// setupTestDB bikin in-memory SQLite database untuk testing
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    // Migrate schema
    err = db.AutoMigrate(&model.User{})
    require.NoError(t, err)

    return db
}

func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)

    user := &model.User{
        Name:     "Irzan Aldi",
        Email:    "irzan@test.com",
        Password: "hashedpassword",
        Role:     model.RoleMember,
    }

    err := repo.Create(user)
    require.NoError(t, err)
    assert.NotZero(t, user.ID)            // ID harus auto-generated
    assert.Equal(t, "Irzan Aldi", user.Name)
}

func TestUserRepository_FindByEmail(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)

    // Seed data
    user := &model.User{
        Name:     "Irzan Aldi",
        Email:    "irzan@test.com",
        Password: "hashedpassword",
        Role:     model.RoleMember,
    }
    require.NoError(t, repo.Create(user))

    t.Run("should find existing user", func(t *testing.T) {
        found, err := repo.FindByEmail("irzan@test.com")
        require.NoError(t, err)
        assert.Equal(t, "Irzan Aldi", found.Name)
        assert.Equal(t, "irzan@test.com", found.Email)
    })

    t.Run("should return error for non-existing email", func(t *testing.T) {
        _, err := repo.FindByEmail("notfound@test.com")
        assert.Error(t, err)
    })
}

func TestUserRepository_FindByID(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)

    user := &model.User{
        Name:     "Test User",
        Email:    "test@test.com",
        Password: "hashedpassword",
        Role:     model.RoleMember,
    }
    require.NoError(t, repo.Create(user))

    t.Run("should find by ID", func(t *testing.T) {
        found, err := repo.FindByID(user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Email, found.Email)
    })

    t.Run("should return error for non-existing ID", func(t *testing.T) {
        _, err := repo.FindByID(9999)
        assert.Error(t, err)
    })
}

func TestUserRepository_Update(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)

    user := &model.User{
        Name:     "Old Name",
        Email:    "update@test.com",
        Password: "hashedpassword",
        Role:     model.RoleMember,
    }
    require.NoError(t, repo.Create(user))

    // Update name
    user.Name = "New Name"
    err := repo.Update(user)
    require.NoError(t, err)

    // Verify
    found, err := repo.FindByID(user.ID)
    require.NoError(t, err)
    assert.Equal(t, "New Name", found.Name)
}

func TestUserRepository_Delete(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)

    user := &model.User{
        Name:     "Delete Me",
        Email:    "delete@test.com",
        Password: "hashedpassword",
        Role:     model.RoleMember,
    }
    require.NoError(t, repo.Create(user))

    // Delete
    err := repo.Delete(user.ID)
    require.NoError(t, err)

    // Verify — soft delete, GORM ga beneran hapus
    _, err = repo.FindByID(user.ID)
    assert.Error(t, err) // harusnya ga ketemu lagi
}

func TestUserRepository_FindAll(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)

    // Seed 5 users
    for i := 1; i <= 5; i++ {
        user := &model.User{
            Name:     "User " + string(rune('A'+i-1)),
            Email:    "user" + string(rune('0'+i)) + "@test.com",
            Password: "hashedpassword",
            Role:     model.RoleMember,
        }
        require.NoError(t, repo.Create(user))
    }

    t.Run("should return paginated results", func(t *testing.T) {
        users, total, err := repo.FindAll(1, 3)
        require.NoError(t, err)
        assert.Equal(t, int64(5), total)
        assert.Len(t, users, 3)
    })

    t.Run("should return second page", func(t *testing.T) {
        users, total, err := repo.FindAll(2, 3)
        require.NoError(t, err)
        assert.Equal(t, int64(5), total)
        assert.Len(t, users, 2) // sisa 2 dari 5
    })
}
```

### Jest Equivalent (NestJS + TypeORM)

```typescript
describe('UserRepository', () => {
  let repo: UserRepository;

  beforeEach(async () => {
    const module = await Test.createTestingModule({
      imports: [TypeOrmModule.forRoot({ type: 'sqlite', database: ':memory:' })],
      providers: [UserRepository],
    }).compile();
    repo = module.get(UserRepository);
  });

  it('should create user', async () => {
    const user = await repo.create({ name: 'Test', email: 'test@test.com' });
    expect(user.id).toBeDefined();
  });
});
```

**Insight:** Di Go, setup test DB lebih straightforward — cuma `gorm.Open(sqlite.Open(":memory:"))`. Di NestJS butuh TestingModule + imports. Logic-nya sama — pakai in-memory DB untuk speed.

---

## 8. Testing Service Layer — Mock Repository, Test Business Logic

Ini bagian paling penting. Service layer punya business logic — inilah yang **paling perlu di-test**.

### Full Example: Testing AuthService

```go
// File: internal/service/auth_test.go
package service

import (
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "github.com/irzan-aldi/task-manager/internal/model"
)

// --- Mock UserRepository ---

type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) Create(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepo) FindByEmail(email string) (*model.User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(id uint) (*model.User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindAll(page, limit int) ([]model.User, int64, error) {
    args := m.Called(page, limit)
    return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepo) Update(user *model.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepo) Delete(id uint) error {
    args := m.Called(id)
    return args.Error(0)
}

// --- Test Register ---

func TestAuthService_Register(t *testing.T) {
    t.Run("success - new user", func(t *testing.T) {
        mockRepo := new(MockUserRepo)

        // FindByEmail return nil → email belum terdaftar
        mockRepo.On("FindByEmail", "irzan@test.com").Return(nil, errors.New("not found"))
        // Create sukses
        mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)

        authService := NewAuthService(mockRepo, "test-jwt-secret")

        result, err := authService.Register(model.RegisterRequest{
            Name:     "Irzan Aldi",
            Email:    "irzan@test.com",
            Password: "secret123",
        })

        require.NoError(t, err)
        require.NotNil(t, result)
        assert.Equal(t, "Irzan Aldi", result.Name)
        assert.Equal(t, "irzan@test.com", result.Email)
        assert.Equal(t, model.RoleMember, result.Role)

        // Verify mock dipanggil
        mockRepo.AssertCalled(t, "FindByEmail", "irzan@test.com")
        mockRepo.AssertCalled(t, "Create", mock.AnythingOfType("*model.User"))
        mockRepo.AssertExpectations(t)
    })

    t.Run("fail - email already registered", func(t *testing.T) {
        mockRepo := new(MockUserRepo)

        // FindByEmail return user → email udah ada
        existingUser := &model.User{
            ID:    1,
            Email: "irzan@test.com",
            Name:  "Existing User",
        }
        mockRepo.On("FindByEmail", "irzan@test.com").Return(existingUser, nil)

        authService := NewAuthService(mockRepo, "test-jwt-secret")

        result, err := authService.Register(model.RegisterRequest{
            Name:     "Irzan Aldi",
            Email:    "irzan@test.com",
            Password: "secret123",
        })

        assert.Nil(t, result)
        assert.Error(t, err)
        assert.EqualError(t, err, "email already registered")

        // Create TIDAK boleh dipanggil
        mockRepo.AssertNotCalled(t, "Create", mock.Anything)
    })

    t.Run("fail - repository create error", func(t *testing.T) {
        mockRepo := new(MockUserRepo)

        mockRepo.On("FindByEmail", "new@test.com").Return(nil, errors.New("not found"))
        mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(errors.New("db error"))

        authService := NewAuthService(mockRepo, "test-jwt-secret")

        result, err := authService.Register(model.RegisterRequest{
            Name:     "New User",
            Email:    "new@test.com",
            Password: "secret123",
        })

        assert.Nil(t, result)
        assert.Error(t, err)
        assert.EqualError(t, err, "failed to create user")
    })
}

// --- Test Login ---

func TestAuthService_Login(t *testing.T) {
    // Bikin user dengan password yang udah di-hash
    hashedUser := &model.User{
        ID:       1,
        Name:     "Irzan Aldi",
        Email:    "irzan@test.com",
        Password: "secret123",
        Role:     model.RoleMember,
    }
    // Hash password beneran supaya CheckPassword jalan
    require.NoError(t, hashedUser.HashPassword())

    t.Run("success - valid credentials", func(t *testing.T) {
        mockRepo := new(MockUserRepo)
        mockRepo.On("FindByEmail", "irzan@test.com").Return(hashedUser, nil)

        authService := NewAuthService(mockRepo, "test-jwt-secret-key-minimum-32-chars!")

        tokens, err := authService.Login(model.LoginRequest{
            Email:    "irzan@test.com",
            Password: "secret123",
        })

        require.NoError(t, err)
        require.NotNil(t, tokens)
        assert.NotEmpty(t, tokens.AccessToken)
        assert.NotEmpty(t, tokens.RefreshToken)
        assert.Equal(t, int64(900), tokens.ExpiresIn)

        mockRepo.AssertExpectations(t)
    })

    t.Run("fail - user not found", func(t *testing.T) {
        mockRepo := new(MockUserRepo)
        mockRepo.On("FindByEmail", "notfound@test.com").Return(nil, errors.New("not found"))

        authService := NewAuthService(mockRepo, "test-jwt-secret-key-minimum-32-chars!")

        tokens, err := authService.Login(model.LoginRequest{
            Email:    "notfound@test.com",
            Password: "secret123",
        })

        assert.Nil(t, tokens)
        assert.Error(t, err)
        assert.EqualError(t, err, "invalid email or password")
    })

    t.Run("fail - wrong password", func(t *testing.T) {
        mockRepo := new(MockUserRepo)
        mockRepo.On("FindByEmail", "irzan@test.com").Return(hashedUser, nil)

        authService := NewAuthService(mockRepo, "test-jwt-secret-key-minimum-32-chars!")

        tokens, err := authService.Login(model.LoginRequest{
            Email:    "irzan@test.com",
            Password: "wrongpassword",
        })

        assert.Nil(t, tokens)
        assert.Error(t, err)
        assert.EqualError(t, err, "invalid email or password")
        // Perhatiin: error message sama untuk email not found dan wrong password
        // Ini security best practice — prevent email enumeration
    })
}

// --- Test ValidateToken ---

func TestAuthService_ValidateToken(t *testing.T) {
    mockRepo := new(MockUserRepo)
    authService := NewAuthService(mockRepo, "test-jwt-secret-key-minimum-32-chars!")

    // Generate token dulu lewat login
    hashedUser := &model.User{
        ID:       1,
        Name:     "Test User",
        Email:    "test@test.com",
        Password: "password123",
        Role:     model.RoleMember,
    }
    require.NoError(t, hashedUser.HashPassword())
    mockRepo.On("FindByEmail", "test@test.com").Return(hashedUser, nil)

    tokens, err := authService.Login(model.LoginRequest{
        Email:    "test@test.com",
        Password: "password123",
    })
    require.NoError(t, err)

    t.Run("valid token", func(t *testing.T) {
        claims, err := authService.ValidateToken(tokens.AccessToken)
        require.NoError(t, err)
        assert.Equal(t, uint(1), claims.UserID)
        assert.Equal(t, "test@test.com", claims.Email)
        assert.Equal(t, model.RoleMember, claims.Role)
    })

    t.Run("invalid token", func(t *testing.T) {
        _, err := authService.ValidateToken("invalid-token-string")
        assert.Error(t, err)
    })

    t.Run("token signed with different secret", func(t *testing.T) {
        otherService := NewAuthService(mockRepo, "completely-different-secret-key-here!")
        _, err := otherService.ValidateToken(tokens.AccessToken)
        assert.Error(t, err)
    })
}

// --- Test RefreshToken ---

func TestAuthService_RefreshToken(t *testing.T) {
    hashedUser := &model.User{
        ID:       1,
        Name:     "Test User",
        Email:    "test@test.com",
        Password: "password123",
        Role:     model.RoleMember,
    }
    require.NoError(t, hashedUser.HashPassword())

    t.Run("success - valid refresh token", func(t *testing.T) {
        mockRepo := new(MockUserRepo)
        mockRepo.On("FindByEmail", "test@test.com").Return(hashedUser, nil)
        mockRepo.On("FindByID", uint(1)).Return(hashedUser, nil)

        authService := NewAuthService(mockRepo, "test-jwt-secret-key-minimum-32-chars!")

        // Login dulu untuk dapat refresh token
        tokens, err := authService.Login(model.LoginRequest{
            Email:    "test@test.com",
            Password: "password123",
        })
        require.NoError(t, err)

        // Refresh
        newTokens, err := authService.RefreshToken(tokens.RefreshToken)
        require.NoError(t, err)
        require.NotNil(t, newTokens)
        assert.NotEmpty(t, newTokens.AccessToken)
        assert.NotEmpty(t, newTokens.RefreshToken)
        // Token baru harus beda dari yang lama
        assert.NotEqual(t, tokens.AccessToken, newTokens.AccessToken)
    })

    t.Run("fail - invalid refresh token", func(t *testing.T) {
        mockRepo := new(MockUserRepo)
        authService := NewAuthService(mockRepo, "test-jwt-secret-key-minimum-32-chars!")

        _, err := authService.RefreshToken("invalid-token")
        assert.Error(t, err)
        assert.EqualError(t, err, "invalid refresh token")
    })

    t.Run("fail - user deleted after token issued", func(t *testing.T) {
        mockRepo := new(MockUserRepo)
        mockRepo.On("FindByEmail", "test@test.com").Return(hashedUser, nil)
        mockRepo.On("FindByID", uint(1)).Return(nil, errors.New("not found"))

        authService := NewAuthService(mockRepo, "test-jwt-secret-key-minimum-32-chars!")

        tokens, err := authService.Login(model.LoginRequest{
            Email:    "test@test.com",
            Password: "password123",
        })
        require.NoError(t, err)

        _, err = authService.RefreshToken(tokens.RefreshToken)
        assert.Error(t, err)
        assert.EqualError(t, err, "user not found")
    })
}
```

### Jest Equivalent

```typescript
describe('AuthService', () => {
  let authService: AuthService;
  let mockUserRepo: jest.Mocked<UserRepository>;

  beforeEach(() => {
    mockUserRepo = {
      findByEmail: jest.fn(),
      create: jest.fn(),
      findByID: jest.fn(),
    } as any;
    authService = new AuthService(mockUserRepo, 'test-secret');
  });

  describe('register', () => {
    it('should register new user', async () => {
      mockUserRepo.findByEmail.mockResolvedValue(null);
      mockUserRepo.create.mockResolvedValue(mockUser);

      const result = await authService.register(registerDto);

      expect(result.email).toBe('irzan@test.com');
      expect(mockUserRepo.findByEmail).toHaveBeenCalledWith('irzan@test.com');
      expect(mockUserRepo.create).toHaveBeenCalled();
    });

    it('should throw if email exists', async () => {
      mockUserRepo.findByEmail.mockResolvedValue(existingUser);

      await expect(authService.register(registerDto))
        .rejects.toThrow('email already registered');
    });
  });
});
```

Pattern-nya mirip: mock dependency → inject → test → verify. Bedanya, di Go semua explicit via interface.

---

## 9. Testing Handler Layer — httptest + Gin Test Mode

### Cara Test HTTP Handler di Go

Go punya package `net/http/httptest` yang built-in. Bisa bikin fake HTTP request dan record response — tanpa butuh server beneran.

### Setup Gin untuk Testing

```go
// Helper function — bikin Gin engine dalam test mode
func setupTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New() // gin.New() tanpa Logger/Recovery — bersih untuk test
    return r
}
```

### Full Example: Testing Auth Handler

```go
// File: internal/handler/auth_test.go
package handler

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/pkg/response"
)

// --- Mock AuthService ---
// Di real project, AuthService juga perlu interface supaya bisa di-mock.
// Untuk contoh ini, kita bikin mock sederhana.

type MockAuthService struct {
    RegisterFn     func(req model.RegisterRequest) (*model.UserResponse, error)
    LoginFn        func(req model.LoginRequest) (*TokenPair, error)
    RefreshTokenFn func(token string) (*TokenPair, error)
}

// TokenPair untuk mock (samakan dengan service.TokenPair)
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

// --- Helper Functions ---

func setupTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    return gin.New()
}

// makeRequest bikin HTTP request dan return response recorder
func makeRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
    var reqBody *bytes.Reader
    if body != nil {
        jsonBytes, _ := json.Marshal(body)
        reqBody = bytes.NewReader(jsonBytes)
    } else {
        reqBody = bytes.NewReader([]byte{})
    }

    req := httptest.NewRequest(method, path, reqBody)
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    return w
}

// parseResponse parse JSON response ke struct
func parseResponse(t *testing.T, w *httptest.ResponseRecorder) response.Response {
    var resp response.Response
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)
    return resp
}

// --- Test Register Handler ---

func TestAuthHandler_Register(t *testing.T) {
    t.Run("success - valid request", func(t *testing.T) {
        router := setupTestRouter()

        // Langsung test handler dengan mock service
        // Karena AuthService di project kita itu concrete struct,
        // dalam prakteknya kamu perlu bikin interface-nya dulu.
        // Di sini kita demo pattern httptest-nya.

        router.POST("/api/v1/auth/register", func(c *gin.Context) {
            var req model.RegisterRequest
            if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, response.Response{
                    Status:  http.StatusBadRequest,
                    Message: err.Error(),
                })
                return
            }

            // Simulate successful register
            c.JSON(http.StatusCreated, response.Response{
                Status:  http.StatusCreated,
                Message: "User registered successfully",
                Data: model.UserResponse{
                    ID:    1,
                    Name:  req.Name,
                    Email: req.Email,
                    Role:  model.RoleMember,
                },
            })
        })

        w := makeRequest(router, "POST", "/api/v1/auth/register", map[string]string{
            "name":     "Irzan Aldi",
            "email":    "irzan@test.com",
            "password": "secret123",
        })

        assert.Equal(t, http.StatusCreated, w.Code)

        resp := parseResponse(t, w)
        assert.Equal(t, http.StatusCreated, resp.Status)
        assert.Equal(t, "User registered successfully", resp.Message)
    })

    t.Run("fail - missing required fields", func(t *testing.T) {
        router := setupTestRouter()

        router.POST("/api/v1/auth/register", func(c *gin.Context) {
            var req model.RegisterRequest
            if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, response.Response{
                    Status:  http.StatusBadRequest,
                    Message: err.Error(),
                })
                return
            }
        })

        // Kirim tanpa name
        w := makeRequest(router, "POST", "/api/v1/auth/register", map[string]string{
            "email":    "irzan@test.com",
            "password": "secret123",
        })

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })

    t.Run("fail - invalid email format", func(t *testing.T) {
        router := setupTestRouter()

        router.POST("/api/v1/auth/register", func(c *gin.Context) {
            var req model.RegisterRequest
            if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, response.Response{
                    Status:  http.StatusBadRequest,
                    Message: err.Error(),
                })
                return
            }
        })

        w := makeRequest(router, "POST", "/api/v1/auth/register", map[string]string{
            "name":     "Irzan",
            "email":    "bukan-email",
            "password": "secret123",
        })

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })

    t.Run("fail - password too short", func(t *testing.T) {
        router := setupTestRouter()

        router.POST("/api/v1/auth/register", func(c *gin.Context) {
            var req model.RegisterRequest
            if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, response.Response{
                    Status:  http.StatusBadRequest,
                    Message: err.Error(),
                })
                return
            }
        })

        w := makeRequest(router, "POST", "/api/v1/auth/register", map[string]string{
            "name":     "Irzan",
            "email":    "irzan@test.com",
            "password": "12345", // min 6
        })

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })
}

// --- Test Login Handler ---

func TestAuthHandler_Login(t *testing.T) {
    t.Run("fail - empty body", func(t *testing.T) {
        router := setupTestRouter()

        router.POST("/api/v1/auth/login", func(c *gin.Context) {
            var req model.LoginRequest
            if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, response.Response{
                    Status:  http.StatusBadRequest,
                    Message: err.Error(),
                })
                return
            }
        })

        w := makeRequest(router, "POST", "/api/v1/auth/login", nil)

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })

    t.Run("fail - invalid email", func(t *testing.T) {
        router := setupTestRouter()

        router.POST("/api/v1/auth/login", func(c *gin.Context) {
            var req model.LoginRequest
            if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, response.Response{
                    Status:  http.StatusBadRequest,
                    Message: err.Error(),
                })
                return
            }
        })

        w := makeRequest(router, "POST", "/api/v1/auth/login", map[string]string{
            "email":    "not-an-email",
            "password": "secret123",
        })

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })
}

// --- Test Me Handler ---

func TestAuthHandler_Me(t *testing.T) {
    t.Run("success - returns current user", func(t *testing.T) {
        router := setupTestRouter()

        router.GET("/api/v1/auth/me", func(c *gin.Context) {
            // Simulate middleware udah set context
            c.Set("userID", uint(1))
            c.Set("email", "irzan@test.com")
            c.Set("role", model.RoleMember)

            userID, _ := c.Get("userID")
            email, _ := c.Get("email")
            role, _ := c.Get("role")

            c.JSON(http.StatusOK, response.Response{
                Status:  http.StatusOK,
                Message: "Current user",
                Data: gin.H{
                    "user_id": userID,
                    "email":   email,
                    "role":    role,
                },
            })
        })

        req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)

        resp := parseResponse(t, w)
        assert.Equal(t, "Current user", resp.Message)
    })
}
```

### Pattern httptest yang Perlu Dihafal

```go
// 1. Bikin request
req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(jsonBody))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token) // kalau perlu auth

// 2. Bikin recorder (fake ResponseWriter)
w := httptest.NewRecorder()

// 3. Serve request ke router
router.ServeHTTP(w, req)

// 4. Check response
assert.Equal(t, http.StatusOK, w.Code)          // status code
assert.Contains(t, w.Body.String(), "success")   // body
```

### Jest Equivalent (Supertest)

```typescript
import request from 'supertest';

describe('AuthController', () => {
  it('POST /auth/register should return 201', async () => {
    const response = await request(app.getHttpServer())
      .post('/auth/register')
      .send({ name: 'Irzan', email: 'irzan@test.com', password: 'secret123' })
      .expect(201);

    expect(response.body.data.email).toBe('irzan@test.com');
  });

  it('POST /auth/register without name should return 400', async () => {
    await request(app.getHttpServer())
      .post('/auth/register')
      .send({ email: 'irzan@test.com', password: 'secret123' })
      .expect(400);
  });
});
```

Mirip kan? `httptest.NewRequest` + `httptest.ResponseRecorder` itu equivalent-nya `supertest`. Bedanya, di Go built-in — ga perlu install apa-apa.

---

## 10. Integration Test — Full Flow

### Apa itu Integration Test?

Unit test = test satu unit (function/method) secara isolated.
Integration test = test **beberapa komponen bareng** — handler + service + repository, end-to-end.

### Full Example: Test Auth Flow End-to-End

```go
// File: internal/integration/auth_integration_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/irzan-aldi/task-manager/internal/handler"
    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/irzan-aldi/task-manager/internal/repository"
    "github.com/irzan-aldi/task-manager/internal/service"
    "github.com/irzan-aldi/task-manager/pkg/response"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// Setup everything — real DB, real services, real handlers
func setupIntegrationTest(t *testing.T) (*gin.Engine, *gorm.DB) {
    // 1. In-memory database
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    // 2. Migrate
    err = db.AutoMigrate(&model.User{}, &model.Board{}, &model.Task{})
    require.NoError(t, err)

    // 3. Real dependencies (bukan mock)
    userRepo := repository.NewUserRepository(db)
    authService := service.NewAuthService(userRepo, "integration-test-secret-key-32chars!")
    authHandler := handler.NewAuthHandler(authService)

    // 4. Setup Gin
    gin.SetMode(gin.TestMode)
    r := gin.New()

    auth := r.Group("/api/v1/auth")
    {
        auth.POST("/register", authHandler.Register)
        auth.POST("/login", authHandler.Login)
        auth.POST("/refresh", authHandler.RefreshToken)
    }

    return r, db
}

func TestAuthFlow_Integration(t *testing.T) {
    router, _ := setupIntegrationTest(t)

    var accessToken string
    var refreshToken string

    // ============================================
    // Step 1: Register
    // ============================================
    t.Run("1. Register new user", func(t *testing.T) {
        body, _ := json.Marshal(map[string]string{
            "name":     "Irzan Aldi",
            "email":    "irzan@test.com",
            "password": "secret123",
        })

        req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusCreated, w.Code)

        var resp response.Response
        json.Unmarshal(w.Body.Bytes(), &resp)
        assert.Equal(t, "User registered successfully", resp.Message)
    })

    // ============================================
    // Step 2: Register duplicate email — should fail
    // ============================================
    t.Run("2. Register duplicate email", func(t *testing.T) {
        body, _ := json.Marshal(map[string]string{
            "name":     "Another User",
            "email":    "irzan@test.com", // same email
            "password": "secret456",
        })

        req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusConflict, w.Code)
    })

    // ============================================
    // Step 3: Login
    // ============================================
    t.Run("3. Login with valid credentials", func(t *testing.T) {
        body, _ := json.Marshal(map[string]string{
            "email":    "irzan@test.com",
            "password": "secret123",
        })

        req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)

        // Parse response untuk ambil token
        var resp struct {
            Status  int    `json:"status"`
            Message string `json:"message"`
            Data    struct {
                AccessToken  string `json:"access_token"`
                RefreshToken string `json:"refresh_token"`
                ExpiresIn    int64  `json:"expires_in"`
            } `json:"data"`
        }
        json.Unmarshal(w.Body.Bytes(), &resp)

        assert.NotEmpty(t, resp.Data.AccessToken)
        assert.NotEmpty(t, resp.Data.RefreshToken)
        assert.Equal(t, int64(900), resp.Data.ExpiresIn)

        // Simpan untuk test berikutnya
        accessToken = resp.Data.AccessToken
        refreshToken = resp.Data.RefreshToken
    })

    // ============================================
    // Step 4: Login with wrong password
    // ============================================
    t.Run("4. Login with wrong password", func(t *testing.T) {
        body, _ := json.Marshal(map[string]string{
            "email":    "irzan@test.com",
            "password": "wrongpassword",
        })

        req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })

    // ============================================
    // Step 5: Refresh token
    // ============================================
    t.Run("5. Refresh token", func(t *testing.T) {
        body, _ := json.Marshal(map[string]string{
            "refresh_token": refreshToken,
        })

        req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)

        var resp struct {
            Data struct {
                AccessToken  string `json:"access_token"`
                RefreshToken string `json:"refresh_token"`
            } `json:"data"`
        }
        json.Unmarshal(w.Body.Bytes(), &resp)

        // Token baru harus beda dari token lama
        assert.NotEmpty(t, resp.Data.AccessToken)
        assert.NotEqual(t, accessToken, resp.Data.AccessToken)
    })

    // ============================================
    // Step 6: Refresh with invalid token
    // ============================================
    t.Run("6. Refresh with invalid token", func(t *testing.T) {
        body, _ := json.Marshal(map[string]string{
            "refresh_token": "invalid-token",
        })

        req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}
```

### Setup/Teardown Pattern

```go
// TestMain jalan SEKALI sebelum semua test di package ini
func TestMain(m *testing.M) {
    // === Setup ===
    fmt.Println("Setting up integration tests...")
    // Setup DB, seed data, dll

    // === Run tests ===
    code := m.Run()

    // === Teardown ===
    fmt.Println("Cleaning up integration tests...")
    // Close DB connections, cleanup files, dll

    os.Exit(code)
}
```

### Pisahkan Unit Test dan Integration Test

```bash
# Cara 1: Pakai build tags
# Di atas file integration test, tambahin:
//go:build integration

# Run unit test aja (default, skip integration)
go test ./...

# Run integration test
go test -tags=integration ./...

# Cara 2: Pakai naming convention + -run flag
# Integration test selalu prefix "TestIntegration_"
go test -run TestIntegration ./...

# Cara 3: Pakai -short flag
# Di integration test, tambahin di awal:
func TestAuthFlow_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    // ...
}

# Run tanpa integration test
go test -short ./...
```

---

## 11. Test Coverage — Ukur Seberapa Ter-cover Code Kamu

### Menjalankan Coverage

```bash
# Basic coverage percentage
go test -cover ./...

# Output:
# ok   github.com/irzan-aldi/task-manager/internal/model     0.5s  coverage: 85.7% of statements
# ok   github.com/irzan-aldi/task-manager/internal/service    0.8s  coverage: 72.3% of statements
# ok   github.com/irzan-aldi/task-manager/internal/handler    1.2s  coverage: 68.1% of statements

# Generate coverage profile file
go test -coverprofile=coverage.out ./...

# Lihat coverage per function
go tool cover -func=coverage.out

# Output:
# github.com/irzan-aldi/task-manager/internal/model/user.go:29:    HashPassword      100.0%
# github.com/irzan-aldi/task-manager/internal/model/user.go:38:    CheckPassword     100.0%
# github.com/irzan-aldi/task-manager/internal/model/user.go:63:    ToResponse        100.0%
# github.com/irzan-aldi/task-manager/internal/service/auth.go:34:  Register          80.0%
# github.com/irzan-aldi/task-manager/internal/service/auth.go:60:  Login             75.0%
# total:                                                           (statements)      76.4%

# Generate HTML report (SANGAT berguna)
go tool cover -html=coverage.out -o coverage.html
# Buka coverage.html di browser — baris yang ke-cover hijau, yang ga ke-cover merah
```

### Coverage HTML Report

HTML report itu **game changer**. Kamu bisa lihat persis baris mana yang belum ke-test:

```
Hijau  = baris yang dijalanin saat test
Merah  = baris yang ga pernah dijalanin (perlu test)
Abu    = baris yang ga bisa dijalanin (import, type declaration)
```

### Target Coverage Strategy

**Jangan target 100%.** Serius.

| Layer | Target | Kenapa |
|-------|--------|--------|
| **Model** | 90-100% | Simple, pure logic, gampang di-test |
| **Service** | 80-90% | Business logic, paling penting |
| **Handler** | 70-80% | Mostly delegation, test validation + response |
| **Repository** | 60-70% | Bergantung DB, integration test lebih berguna |
| **Middleware** | 70-80% | Test happy path + error cases |
| **Overall** | 70-80% | Sweet spot |

**Kenapa bukan 100%?**

1. **Diminishing returns** — dari 0% ke 70% gampang. Dari 70% ke 90% mulai effort. Dari 90% ke 100% ngabisin waktu untuk test trivial code
2. **Some code isn't worth testing** — `main.go`, config loading, database connection setup
3. **False sense of security** — 100% coverage bukan berarti bug-free. Bisa aja test-nya cuma check "ga error" tanpa verify behavior

### Jest Coverage Comparison

```bash
# Jest
npx jest --coverage
# Generates:
# - Statements: 82.14%
# - Branches: 71.43%
# - Functions: 90.91%
# - Lines: 82.14%

# Go cuma report statement coverage
go test -cover ./...
# - coverage: 82.1% of statements
```

Go cuma punya statement coverage (bukan branch/function/line). Lebih simpel, tapi cukup informatif.

---

## 12. Benchmark — Test Performance

### testing.B

Selain test, Go punya built-in **benchmark**. Ga perlu install library benchmark terpisah.

```go
// File: internal/model/user_benchmark_test.go
package model

import "testing"

// Benchmark function HARUS prefix "Benchmark" dan terima *testing.B
func BenchmarkUser_HashPassword(b *testing.B) {
    user := &User{Password: "secret123"}

    // b.N otomatis ditentuin Go — run sampai statistically significant
    for i := 0; i < b.N; i++ {
        user.Password = "secret123" // reset password tiap iterasi
        user.HashPassword()
    }
}

func BenchmarkUser_CheckPassword(b *testing.B) {
    user := &User{Password: "secret123"}
    user.HashPassword()

    b.ResetTimer() // Reset timer — jangan hitung setup time

    for i := 0; i < b.N; i++ {
        user.CheckPassword("secret123")
    }
}

// Benchmark buat compare dua implementasi
func BenchmarkToResponse_Direct(b *testing.B) {
    user := User{
        ID:    1,
        Name:  "Irzan Aldi",
        Email: "irzan@test.com",
        Role:  RoleMember,
    }

    for i := 0; i < b.N; i++ {
        _ = user.ToResponse()
    }
}

// Benchmark JSON marshaling — ini sering jadi bottleneck
func BenchmarkUser_JSONMarshal(b *testing.B) {
    user := User{
        ID:    1,
        Name:  "Irzan Aldi",
        Email: "irzan@test.com",
        Role:  RoleMember,
    }

    for i := 0; i < b.N; i++ {
        json.Marshal(user)
    }
}
```

### Menjalankan Benchmark

```bash
# Run semua benchmark
go test -bench=. ./internal/model/

# Output:
# BenchmarkUser_HashPassword-8          6        197485633 ns/op
# BenchmarkUser_CheckPassword-8         6        199712417 ns/op
# BenchmarkToResponse_Direct-8    38621754          31.24 ns/op
# BenchmarkUser_JSONMarshal-8      3271456         367.8 ns/op

# Baca output:
# -8              = jumlah CPU core
# 6               = b.N (berapa kali dijalanin)
# 197485633 ns/op = ~200ms per operasi (bcrypt emang lambat by design!)
# 31.24 ns/op     = 31 nanosecond per operasi (super cepat)

# Run benchmark dengan memory allocation stats
go test -bench=. -benchmem ./internal/model/

# Output tambahan:
# BenchmarkToResponse_Direct-8    38621754    31.24 ns/op    64 B/op    1 allocs/op
#                                                             ↑           ↑
#                                                       bytes allocated   allocation count

# Run specific benchmark
go test -bench=BenchmarkUser_HashPassword ./internal/model/

# Run benchmark N kali (untuk consistency)
go test -bench=. -count=5 ./internal/model/
```

### b.ResetTimer — Kapan Pakainya

```go
func BenchmarkExpensiveSetup(b *testing.B) {
    // Setup yang berat — ga mau dihitung ke benchmark
    db := setupHeavyDatabase()
    data := loadLargeFixture()

    b.ResetTimer() // <-- timer direset di sini

    for i := 0; i < b.N; i++ {
        processData(data, db) // yang di-benchmark cuma ini
    }
}
```

### Jest Equivalent

```typescript
// Jest ga punya built-in benchmark. Harus pakai library terpisah:
// npm install benchmark
import Benchmark from 'benchmark';

const suite = new Benchmark.Suite();
suite
  .add('hashPassword', () => { user.hashPassword(); })
  .add('checkPassword', () => { user.checkPassword('secret123'); })
  .on('complete', function() { console.log(this.filter('fastest').map('name')); })
  .run();
```

Di Go, benchmark itu **first-class citizen**. Built-in, integrated sama `go test`, standardized output format.

---

## 13. Test Fixtures & Helpers

### testdata Folder

Go punya convention: folder `testdata/` **diabaikan oleh Go toolchain**. Perfect buat simpan test fixtures (JSON, SQL, golden files).

```
internal/
  service/
    auth.go
    auth_test.go
    testdata/                 ← Go toolchain ga compile ini
      register_request.json
      login_response.json
      users_seed.json
```

```go
// Baca fixture di test
func TestWithFixture(t *testing.T) {
    // os.ReadFile baca file relative ke package directory
    data, err := os.ReadFile("testdata/register_request.json")
    require.NoError(t, err)

    var req model.RegisterRequest
    err = json.Unmarshal(data, &req)
    require.NoError(t, err)

    assert.Equal(t, "Irzan Aldi", req.Name)
}
```

### Test Helper Functions

Buat helper yang sering dipake di banyak test:

```go
// File: internal/testutil/helpers.go
package testutil

import (
    "testing"

    "github.com/irzan-aldi/task-manager/internal/model"
    "github.com/stretchr/testify/require"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// SetupTestDB bikin in-memory SQLite — bisa dipake di semua test
func SetupTestDB(t *testing.T) *gorm.DB {
    t.Helper() // <-- ini bikin error message nunjuk ke caller, bukan ke helper ini

    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    err = db.AutoMigrate(&model.User{}, &model.Board{}, &model.Task{})
    require.NoError(t, err)

    return db
}

// CreateTestUser bikin user dengan password yang udah di-hash
func CreateTestUser(t *testing.T, name, email, password string, role model.Role) *model.User {
    t.Helper()

    user := &model.User{
        Name:     name,
        Email:    email,
        Password: password,
        Role:     role,
    }
    require.NoError(t, user.HashPassword())

    return user
}

// StringPtr helper buat bikin pointer ke string (berguna untuk UpdateRequest)
func StringPtr(s string) *string {
    return &s
}

// UintPtr helper buat pointer ke uint
func UintPtr(u uint) *uint {
    return &u
}
```

**`t.Helper()`** itu penting banget. Tanpa ini, kalau test fail di dalam helper, error message nunjuk ke baris di helper function — bikin bingung. Dengan `t.Helper()`, error message nunjuk ke baris yang **memanggil** helper.

### TestMain — Setup/Teardown Per Package

```go
// File: internal/service/auth_test.go
package service

import (
    "fmt"
    "os"
    "testing"
)

// TestMain jalan SEKALI sebelum semua test di package ini
func TestMain(m *testing.M) {
    // === SETUP ===
    fmt.Println("=== Setting up service tests ===")
    // Bisa: connect DB, seed data, setup env vars, dll

    // === RUN ALL TESTS ===
    exitCode := m.Run()

    // === TEARDOWN ===
    fmt.Println("=== Cleaning up service tests ===")
    // Bisa: close connections, cleanup temp files, dll

    os.Exit(exitCode)
}
```

**Jest equivalent:**

```typescript
// Jest punya global setup/teardown
// jest.config.ts
module.exports = {
  globalSetup: './test/setup.ts',
  globalTeardown: './test/teardown.ts',
};

// Atau per-file
beforeAll(() => { /* setup */ });
afterAll(() => { /* teardown */ });
```

| Go | Jest |
|----|------|
| `TestMain` | `globalSetup` / `beforeAll` |
| `t.Cleanup()` | `afterEach` / `afterAll` |
| `t.Helper()` | Ga ada equivalent langsung |

### t.Cleanup — Teardown Per Test

```go
func TestWithCleanup(t *testing.T) {
    db := setupTestDB(t)

    // Register cleanup — jalan SETELAH test selesai (termasuk subtest)
    t.Cleanup(func() {
        sqlDB, _ := db.DB()
        sqlDB.Close()
    })

    // ... test logic ...
    // Setelah function ini selesai, cleanup otomatis dipanggil
}
```

Ini mirip `afterEach` tapi **per-test**. Kalau test punya resource yang harus di-cleanup, pakai `t.Cleanup()`.

---

## 14. Testing Best Practices

### 14.1 AAA Pattern (Arrange/Act/Assert)

Setiap test harus punya 3 bagian yang jelas:

```go
func TestAuthService_Register_Success(t *testing.T) {
    // === ARRANGE — setup dependencies, data, expectations ===
    mockRepo := new(MockUserRepo)
    mockRepo.On("FindByEmail", "new@test.com").Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything).Return(nil)
    authService := NewAuthService(mockRepo, "test-secret")

    req := model.RegisterRequest{
        Name:     "New User",
        Email:    "new@test.com",
        Password: "secret123",
    }

    // === ACT — jalanin function yang di-test ===
    result, err := authService.Register(req)

    // === ASSERT — check hasilnya ===
    require.NoError(t, err)
    assert.Equal(t, "new@test.com", result.Email)
    assert.Equal(t, model.RoleMember, result.Role)
    mockRepo.AssertExpectations(t)
}
```

**Jest equivalent (same pattern):**
```typescript
it('should register new user', async () => {
  // Arrange
  mockRepo.findByEmail.mockResolvedValue(null);
  mockRepo.create.mockResolvedValue(mockUser);

  // Act
  const result = await authService.register(registerDto);

  // Assert
  expect(result.email).toBe('new@test.com');
  expect(mockRepo.create).toHaveBeenCalled();
});
```

### 14.2 Test Naming Convention

```go
// Pattern: Test{Struct}_{Method}_{Scenario}
func TestAuthService_Register_Success(t *testing.T) { ... }
func TestAuthService_Register_DuplicateEmail(t *testing.T) { ... }
func TestAuthService_Login_InvalidPassword(t *testing.T) { ... }
func TestAuthService_Login_UserNotFound(t *testing.T) { ... }

// Atau pakai t.Run untuk group:
func TestAuthService_Register(t *testing.T) {
    t.Run("success", func(t *testing.T) { ... })
    t.Run("duplicate email", func(t *testing.T) { ... })
    t.Run("invalid input", func(t *testing.T) { ... })
}
```

**Naming rules:**
- Prefix `Test` wajib (Go convention)
- Nama harus **descriptive** — baca nama test = tau apa yang di-test
- Jangan pake `Test1`, `Test2` — ga bermakna

### 14.3 Don't Test Private Functions

```go
// auth.go
func (s *AuthService) Register(req model.RegisterRequest) (*model.UserResponse, error) { ... }  // exported ← TEST INI
func (s *AuthService) generateTokenPair(user *model.User) (*TokenPair, error) { ... }            // unexported ← JANGAN test langsung

// Kenapa?
// generateTokenPair itu implementation detail.
// Kalau kamu test langsung, terus refactor (misalnya rename atau pecah),
// test-nya ikut break meskipun behavior Register/Login ga berubah.

// Test-nya via public function:
func TestAuthService_Login_ReturnsTokenPair(t *testing.T) {
    // Login internally panggil generateTokenPair
    // Kamu test behavior Login → otomatis generateTokenPair ke-test juga
}
```

### 14.4 Test Behavior, Not Implementation

```go
// SALAH — test implementation detail
func TestAuthService_Register_CallsHashPassword(t *testing.T) {
    // Test apakah HashPassword dipanggil → fragile
    // Kalau ganti cara hash → test break meskipun behavior benar
}

// BENAR — test behavior/outcome
func TestAuthService_Register_PasswordIsHashed(t *testing.T) {
    result, err := authService.Register(req)
    require.NoError(t, err)
    // Check behavior: password yang disimpan bukan plaintext
    // Ga peduli gimana cara hash-nya
}
```

### 14.5 One Assert Per Concept

```go
// KURANG IDEAL — terlalu banyak assert yang ga related
func TestRegister(t *testing.T) {
    result, err := authService.Register(req)
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "irzan@test.com", result.Email)
    assert.Equal(t, model.RoleMember, result.Role)
    assert.NotEmpty(t, result.ID)
    // kalau assert pertama fail, yang lain juga biasanya fail
    // tapi kamu ga tau mana yang root cause
}

// LEBIH BAIK — pecah per concern, pakai subtests
func TestAuthService_Register(t *testing.T) {
    mockRepo := new(MockUserRepo)
    // ... setup ...
    result, err := authService.Register(req)

    t.Run("should not return error", func(t *testing.T) {
        require.NoError(t, err)
    })

    t.Run("should return user with correct email", func(t *testing.T) {
        require.NotNil(t, result)
        assert.Equal(t, "irzan@test.com", result.Email)
    })

    t.Run("should set role to member", func(t *testing.T) {
        require.NotNil(t, result)
        assert.Equal(t, model.RoleMember, result.Role)
    })
}
```

### 14.6 Golden Files

Golden files = expected output yang disimpan di file. Berguna untuk test yang output-nya complex (JSON response, HTML template, dll).

```go
func TestComplexJSONResponse(t *testing.T) {
    result := generateComplexResponse()
    got, _ := json.MarshalIndent(result, "", "  ")

    goldenFile := "testdata/complex_response.golden.json"

    // Update golden file kalau flag -update
    if *update {
        os.WriteFile(goldenFile, got, 0644)
        return
    }

    // Compare dengan golden file
    want, err := os.ReadFile(goldenFile)
    require.NoError(t, err)
    assert.JSONEq(t, string(want), string(got))
}

// Jalanin dengan -update untuk regenerate golden files:
// go test -run TestComplexJSONResponse -update
```

**Jest equivalent:**
```typescript
// Jest punya built-in snapshot testing
it('should match snapshot', () => {
  const result = generateComplexResponse();
  expect(result).toMatchSnapshot();
});

// Update snapshots:
// npx jest --updateSnapshot
```

---

## 15. Jest -> Go Test Cheat Sheet

Ini **reference utama** kamu. Bookmark halaman ini.

### Test Structure

| Jest | Go |
|------|-----|
| `describe('Auth', () => {})` | `func TestAuth(t *testing.T) {}` atau pisah per function |
| `it('should login', () => {})` | `t.Run("should login", func(t *testing.T) {})` |
| `test('login works', () => {})` | `func TestAuth_Login(t *testing.T) {}` |
| `describe.each(cases)` | Table-driven tests (slice + loop + t.Run) |
| `it.skip('...')` | `t.Skip("reason")` |
| `it.only('...')` | `go test -run TestSpecificName` |
| `xit('...')` | Comment out atau `t.Skip()` |

### Assertions

| Jest | Go (testify) |
|------|-------------|
| `expect(a).toBe(b)` | `assert.Equal(t, b, a)` |
| `expect(a).not.toBe(b)` | `assert.NotEqual(t, b, a)` |
| `expect(a).toBeNull()` | `assert.Nil(t, a)` |
| `expect(a).not.toBeNull()` | `assert.NotNil(t, a)` |
| `expect(a).toBeDefined()` | `assert.NotNil(t, a)` |
| `expect(a).toBeUndefined()` | `assert.Nil(t, a)` (Go ga ada undefined) |
| `expect(a).toBeTruthy()` | `assert.True(t, a)` |
| `expect(a).toBeFalsy()` | `assert.False(t, a)` |
| `expect(a).toBeGreaterThan(b)` | `assert.Greater(t, a, b)` |
| `expect(a).toBeLessThan(b)` | `assert.Less(t, a, b)` |
| `expect(str).toContain(sub)` | `assert.Contains(t, str, sub)` |
| `expect(arr).toHaveLength(n)` | `assert.Len(t, arr, n)` |
| `expect(obj).toEqual(other)` | `assert.Equal(t, other, obj)` (deep equal) |
| `expect(obj).toMatchObject(partial)` | Ga ada langsung — check field per field |
| `expect(fn).toThrow()` | `assert.Error(t, err)` |
| `expect(fn).toThrow('msg')` | `assert.EqualError(t, err, "msg")` |
| `expect(fn).not.toThrow()` | `assert.NoError(t, err)` |
| `expect(json).toMatchSnapshot()` | Golden files (manual) |

### Setup/Teardown

| Jest | Go |
|------|-----|
| `beforeAll()` | `TestMain(m *testing.M)` |
| `afterAll()` | `TestMain` cleanup section |
| `beforeEach()` | Setup di awal test / `suite.SetupTest()` |
| `afterEach()` | `t.Cleanup()` / `suite.TearDownTest()` |

### Mocking

| Jest | Go |
|------|-----|
| `jest.mock('./module')` | Define interface + mock struct |
| `jest.fn()` | Method on mock struct |
| `mockFn.mockReturnValue(x)` | `mock.On("Method").Return(x)` |
| `mockFn.mockResolvedValue(x)` | `mock.On("Method").Return(x, nil)` |
| `mockFn.mockRejectedValue(err)` | `mock.On("Method").Return(nil, err)` |
| `mockFn.mockImplementation(fn)` | `mock.On("Method").Run(fn).Return(...)` |
| `expect(fn).toHaveBeenCalled()` | `mock.AssertCalled(t, "Method", ...)` |
| `expect(fn).toHaveBeenCalledTimes(n)` | `mock.AssertNumberOfCalls(t, "Method", n)` |
| `expect(fn).toHaveBeenCalledWith(args)` | `mock.AssertCalled(t, "Method", args)` |
| `jest.spyOn(obj, 'method')` | Ga ada — Go pake interface mock |
| `jest.clearAllMocks()` | Bikin mock instance baru |

### Running Tests

| Jest | Go |
|------|-----|
| `npx jest` | `go test ./...` |
| `npx jest auth.spec.ts` | `go test ./internal/service/` |
| `npx jest --watch` | `watchexec -e go "go test ./..."` (external) |
| `npx jest --verbose` | `go test -v ./...` |
| `npx jest --coverage` | `go test -cover ./...` |
| `npx jest --testNamePattern "login"` | `go test -run "Login" ./...` |
| `npx jest --bail` | `go test -failfast ./...` |
| `npx jest --runInBand` | Default behavior (sequential) |
| `npx jest --forceExit` | Ga perlu — Go test always exits |

### Async Testing

| Jest | Go |
|------|-----|
| `async/await` in test | Ga ada — Go synchronous by default |
| `expect(promise).resolves.toBe(x)` | `result, err := fn(); assert.Equal(t, x, result)` |
| `expect(promise).rejects.toThrow()` | `_, err := fn(); assert.Error(t, err)` |
| `jest.setTimeout(10000)` | `go test -timeout 30s ./...` |
| Fake timers `jest.useFakeTimers()` | Inject time as dependency |

### HTTP Testing

| Jest (Supertest) | Go (httptest) |
|-----------------|---------------|
| `request(app).get('/path')` | `httptest.NewRequest("GET", "/path", nil)` |
| `.send(body)` | Body as `io.Reader` parameter |
| `.set('Header', 'value')` | `req.Header.Set("Header", "value")` |
| `.expect(200)` | `assert.Equal(t, 200, w.Code)` |
| `.expect('Content-Type', /json/)` | `assert.Contains(t, w.Header().Get("Content-Type"), "json")` |
| Response body | `w.Body.String()` atau `json.Unmarshal(w.Body.Bytes(), &resp)` |

---

## 16. Checklist Bab Ini

### Fundamental

- [ ] Paham kenapa testing penting (bukan cuma teori, tapi real-world impact)
- [ ] Paham perbedaan testing Go vs Jest (built-in, no assertion, file naming)
- [ ] Bisa jalanin `go test ./...` dan baca output-nya

### Unit Test

- [ ] Bisa nulis test function `TestXxx(t *testing.T)`
- [ ] Paham `t.Error` vs `t.Fatal` vs `t.Skip`
- [ ] Bisa pakai `t.Run()` untuk subtests
- [ ] Bisa nulis table-driven tests (THE Go pattern)

### Testify

- [ ] Udah install `github.com/stretchr/testify`
- [ ] Paham perbedaan `assert` (lanjut) vs `require` (stop)
- [ ] Bisa pakai `assert.Equal`, `assert.NoError`, `assert.Nil`, dll

### Mocking

- [ ] Paham kenapa Go pakai interface-based mocking (bukan magic)
- [ ] Bisa bikin manual mock struct
- [ ] Bisa pakai `testify/mock` (`mock.On`, `mock.Return`, `AssertCalled`)

### Layer Testing

- [ ] Bisa test model layer (HashPassword, CheckPassword, ToResponse)
- [ ] Bisa test repository layer (pakai SQLite in-memory)
- [ ] Bisa test service layer (mock repository, test business logic)
- [ ] Bisa test handler layer (httptest, gin test mode)

### Advanced

- [ ] Bisa bikin integration test (full flow tanpa mock)
- [ ] Bisa run dan baca coverage report (`go test -cover`, HTML report)
- [ ] Bisa nulis benchmark (`BenchmarkXxx(b *testing.B)`)
- [ ] Paham `testdata/` folder, `t.Helper()`, `TestMain`
- [ ] Paham AAA pattern, test naming, test behavior not implementation

### Migration Skill

- [ ] Bisa translate Jest `describe/it/expect` ke Go test
- [ ] Bisa translate Jest mock ke Go interface mock
- [ ] Hafal Jest -> Go cheat sheet (atau bookmark halaman ini)

---

> **Next:** [Bab 11: Docker & Containerization](11-docker.md) — kita bakal containerize project Task Manager API kita, bikin Dockerfile yang production-ready, dan setup Docker Compose untuk development.
