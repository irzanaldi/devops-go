# Bab 6: Database — PostgreSQL + GORM

> Kalau REST API itu skeleton, database itu jantungnya. Di bab ini kita bahas kenapa PostgreSQL, cara pakai GORM (ORM-nya Go), dan gimana project Task Manager kita connect ke database — mulai dari model, migration, CRUD, sampai relationship.

---

## 6.1 Kenapa PostgreSQL?

Kamu mungkin di dunia NestJS biasa pakai MongoDB (Mongoose) atau MySQL. Kenapa pindah ke PostgreSQL?

### Comparison: PostgreSQL vs MySQL vs MongoDB

| Aspek | PostgreSQL | MySQL | MongoDB |
|-------|-----------|-------|---------|
| Type system | **Sangat kaya** (JSONB, Array, UUID, Enum) | Standard | Schema-less |
| ACID compliance | **Full** | Full (InnoDB) | Partial |
| JSON support | **JSONB native + indexing** | JSON (basic) | Native |
| Full-text search | **Built-in (tsvector)** | Basic | Atlas Search |
| Performance (complex query) | **Excellent** | Good | Depends |
| Scalability | Vertical + Citus (horizontal) | Read replicas | Horizontal |
| DevOps/Cloud adoption | **#1 di AWS RDS, GCP, Railway** | Populer | Atlas |
| Open source | **100% open** | Oracle-owned | SSPL license |

### Kenapa PostgreSQL jadi default di 2026?

1. **Industry standard** — Hampir semua startup modern pakai PostgreSQL
2. **JSONB** — Bisa simpan JSON dengan indexing, best of both worlds (relational + document)
3. **Extension ecosystem** — PostGIS (geospatial), pgvector (AI/embedding), TimescaleDB (time series)
4. **Cloud-friendly** — AWS RDS, Google Cloud SQL, Supabase, Neon, Railway — semua default PostgreSQL
5. **Tooling** — Migration tools, ORMs, semua support PostgreSQL first-class

### Kalau kamu dari MongoDB...

MongoDB bagus untuk prototyping cepat, tapi di production sering ketemu masalah:
- Data inconsistency (no strict schema)
- Transaction support terbatas
- Join antar collection itu pain

PostgreSQL kasih kamu **strict schema** (data pasti konsisten) + **JSONB column** kalau butuh flexibility. Best of both worlds.

---

## 6.2 GORM — ORM di Go

### Apa itu GORM?

GORM = **Go Object-Relational Mapping**. Library paling populer untuk interaksi database di Go. Fungsinya sama kayak TypeORM/Prisma di NestJS — mapping struct Go ke table database.

### Comparison: GORM vs TypeORM vs Prisma

| Aspek | GORM (Go) | TypeORM (NestJS) | Prisma (NestJS) |
|-------|-----------|-----------------|-----------------|
| Approach | Active Record + Data Mapper | Decorator-based | Schema-first |
| Model definition | **Struct + tags** | Class + decorators | `.prisma` schema file |
| Migration | AutoMigrate / manual SQL | `synchronize` / CLI | `prisma migrate` |
| Query builder | Chainable methods | QueryBuilder | Fluent API |
| Raw SQL | `db.Raw()` | `query()` | `$queryRaw` |
| Relationship | Struct field + tag | `@ManyToOne()` etc | `@relation` in schema |
| Transaction | `db.Transaction()` | `queryRunner` | `$transaction()` |
| Performance | **Cepat (compiled)** | Runtime overhead | Runtime overhead |
| Maturity | **Sangat mature (10+ tahun)** | Mature | Newer |

### GORM vs Raw SQL (sqlx)

Di Go, ada 2 camp:
- **GORM camp** — "Pakai ORM, productivity tinggi"
- **sqlx camp** — "Raw SQL, full control, no magic"

Untuk project ini kita pakai **GORM** karena:
1. Familiar buat kamu yang dari TypeORM (konsep sama)
2. AutoMigrate mempercepat development
3. Relationship handling built-in
4. Nanti kalau butuh raw SQL, GORM support `db.Raw()`

> **Tips:** Di production besar, banyak company hybrid — GORM untuk CRUD simple, raw SQL untuk query complex/reporting.

---

## 6.3 Install Dependencies

```bash
# PostgreSQL driver + GORM
go get gorm.io/gorm
go get gorm.io/driver/postgres

go mod tidy
```

**NestJS equivalent:**
```bash
npm install @nestjs/typeorm typeorm pg
```

---

## 6.4 Connection Setup

Ini file koneksi database di project kita.

### File: `internal/config/config.go`

```go
package config

import "os"

type Config struct {
	Env       string
	Port      string
	DB        DatabaseConfig
	Redis     RedisConfig
	JWTSecret string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load() *Config {
	return &Config{
		Env:       getEnv("APP_ENV", "development"),
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "task_manager"),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
```

**NestJS comparison** — di NestJS kamu biasa pakai `ConfigModule`:
```typescript
// NestJS — app.module.ts
TypeOrmModule.forRoot({
  type: 'postgres',
  host: process.env.DB_HOST,
  port: +process.env.DB_PORT,
  username: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME,
})
```

Di Go, kita buat config struct sendiri. Ga ada framework magic — semua explicit.

### File: `internal/config/database.go`

Ini jantung koneksi database:

```go
package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name,
	)

	logLevel := logger.Info
	if cfg.Env == "production" {
		logLevel = logger.Error
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connected successfully")
	return db
}
```

### Bedah Satu-Satu

#### 1. DSN (Data Source Name)
```go
dsn := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name,
)
```

DSN itu connection string ke PostgreSQL. Format key=value, bukan URL. `sslmode=disable` untuk development — di production kamu harus set `sslmode=require`.

**NestJS equivalent** — ini sama kayak `DATABASE_URL` atau object config di TypeORM.

#### 2. GORM Logger
```go
logLevel := logger.Info
if cfg.Env == "production" {
    logLevel = logger.Error
}
```

Di development, GORM nge-log semua SQL query (berguna untuk debugging). Di production, cuma log error biar ga noisy.

**NestJS equivalent:** `logging: true` di TypeORM config.

#### 3. Connection Pool
```go
sqlDB.SetMaxIdleConns(10)   // Koneksi idle yang tetap hidup
sqlDB.SetMaxOpenConns(100)  // Maksimal koneksi bersamaan
```

Kenapa butuh connection pool? Database punya limit koneksi. Kalau setiap request buka koneksi baru → database overload. Pool = reuse koneksi yang sudah ada.

- `MaxIdleConns(10)` — 10 koneksi standby, siap dipakai tanpa overhead buat buka baru
- `MaxOpenConns(100)` — maksimal 100 koneksi barengan. Kalau lebih, request harus antri

**NestJS/TypeORM equivalent:**
```typescript
TypeOrmModule.forRoot({
  extra: {
    max: 100, // pool size
  }
})
```

#### 4. `gorm.Open()` Return Value
```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{...})
```

`*gorm.DB` ini object utama yang bakal kamu pass ke semua repository. Satu instance, di-share ke seluruh app. Thread-safe.

### File `.env`

```env
APP_ENV=development
PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=task_manager

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-secret-key-change-in-production
```

---

## 6.5 Model Definition — Struct Tags

Di NestJS/TypeORM, kamu define entity pakai **class + decorators**. Di Go, kita pakai **struct + tags**.

### File: `internal/model/user.go`

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
```

### Bedah Struct Tags

Setiap field punya **2 tag** utama: `json` dan `gorm`.

#### `json` tag — Gimana field muncul di JSON response

| Tag | Artinya |
|-----|---------|
| `json:"id"` | Field `ID` jadi `"id"` di JSON |
| `json:"-"` | **Ga muncul di JSON** (Password, DeletedAt) |
| `json:"boards,omitempty"` | Kalau kosong/nil, ga muncul di JSON |

**Penting:** `json:"-"` di Password artinya password ga pernah terkirim ke client. Di NestJS kamu biasa pakai `@Exclude()` dari class-transformer. Di Go, cukup `json:"-"`.

#### `gorm` tag — Gimana field di-map ke database column

| Tag | Artinya | SQL Equivalent |
|-----|---------|---------------|
| `gorm:"primaryKey"` | Primary key | `SERIAL PRIMARY KEY` |
| `gorm:"size:100"` | Panjang varchar | `VARCHAR(100)` |
| `gorm:"not null"` | Ga boleh kosong | `NOT NULL` |
| `gorm:"uniqueIndex"` | Unique + index | `UNIQUE INDEX` |
| `gorm:"default:member"` | Default value | `DEFAULT 'member'` |
| `gorm:"type:text"` | Custom column type | `TEXT` |
| `gorm:"index"` | Bikin index | `CREATE INDEX` |
| `gorm:"foreignKey:OwnerID"` | Foreign key reference | `REFERENCES users(id)` |

#### NestJS/TypeORM Comparison — Side by Side

**NestJS/TypeORM:**
```typescript
@Entity()
export class User {
  @PrimaryGeneratedColumn()
  id: number;

  @Column({ length: 100 })
  name: string;

  @Column({ unique: true, length: 100 })
  email: string;

  @Column({ length: 255 })
  @Exclude()
  password: string;

  @Column({ default: 'member', length: 20 })
  role: string;

  @OneToMany(() => Board, board => board.owner)
  boards: Board[];

  @CreateDateColumn()
  created_at: Date;

  @UpdateDateColumn()
  updated_at: Date;

  @DeleteDateColumn()
  deleted_at: Date;
}
```

**Go/GORM:**
```go
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
```

Konsepnya **identik**. Cuma syntax-nya beda — decorator vs tag. Kalau kamu udah paham TypeORM entity, kamu udah paham GORM model.

#### Enum di Go

Perhatiin gimana kita define enum:

```go
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)
```

Di TypeScript biasa pakai `enum Role { Admin = 'admin', Member = 'member' }`. Di Go ga ada `enum` keyword — kita pakai **custom type + constants**. Hasilnya sama — type-safe string yang nilainya terbatas.

#### Method di Model

```go
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

Di Go, method bisa di-attach ke struct. `(u *User)` artinya method ini milik User. Sama kayak method di class TypeScript, tapi tanpa class keyword.

#### Request/Response DTO

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

`binding` tag itu dari Gin (framework). Sama kayak `class-validator` di NestJS (`@IsEmail()`, `@MinLength(6)`).

`ToResponse()` method ini pattern biar kita ga expose Password ke API response. Di NestJS biasa pakai `class-transformer` + `@Exclude()`. Di Go, kita explicit bikin response struct terpisah.

---

### File: `internal/model/board.go`

```go
package model

import (
	"time"

	"gorm.io/gorm"
)

type Board struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"size:500"`
	OwnerID     uint           `json:"owner_id" gorm:"not null"`
	Owner       User           `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Tasks       []Task         `json:"tasks,omitempty" gorm:"foreignKey:BoardID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateBoardRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=500"`
}

type UpdateBoardRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}
```

### File: `internal/model/task.go`

```go
package model

import (
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Task struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      TaskStatus     `json:"status" gorm:"size:20;default:todo"`
	Priority    Priority       `json:"priority" gorm:"size:20;default:medium"`
	BoardID     uint           `json:"board_id" gorm:"not null"`
	Board       Board          `json:"board,omitempty" gorm:"foreignKey:BoardID"`
	AssigneeID  *uint          `json:"assignee_id"`
	Assignee    *User          `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
	DueDate     *time.Time     `json:"due_date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
```

---

## 6.6 Relationships — HasMany, BelongsTo

Ini bagian yang paling mirip sama TypeORM. Cuma syntax-nya beda.

### Mapping Relationship di Project Kita

```
User (1) ──── hasMany ────> Board (N)      # User punya banyak Board
Board (1) ──── hasMany ────> Task (N)       # Board punya banyak Task
User (1) ──── hasMany ────> Task (N)        # User bisa di-assign ke banyak Task
```

### Gimana Define Relationship di GORM?

#### HasMany (User → Boards)
```go
// Di User struct
Boards []Board `json:"boards,omitempty" gorm:"foreignKey:OwnerID"`
```

**NestJS/TypeORM equivalent:**
```typescript
@OneToMany(() => Board, board => board.owner)
boards: Board[];
```

#### BelongsTo (Board → User)
```go
// Di Board struct
OwnerID uint `json:"owner_id" gorm:"not null"`       // FK column
Owner   User `json:"owner,omitempty" gorm:"foreignKey:OwnerID"` // Relation
```

**NestJS/TypeORM equivalent:**
```typescript
@Column()
ownerId: number;

@ManyToOne(() => User, user => user.boards)
@JoinColumn({ name: 'owner_id' })
owner: User;
```

#### Nullable Relationship (Task → Assignee)

Perhatiin Task model — `AssigneeID` itu **opsional** (task bisa ga di-assign ke siapa-siapa):

```go
AssigneeID  *uint  `json:"assignee_id"`                                   // Pointer = nullable
Assignee    *User  `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"` // Pointer = nullable
```

`*uint` (pointer) artinya field ini **bisa nil**. Kalau pakai `uint` (tanpa pointer), nilainya selalu 0 — ga bisa null. Ini penting!

**NestJS/TypeORM equivalent:**
```typescript
@Column({ nullable: true })
assigneeId?: number;

@ManyToOne(() => User, { nullable: true })
assignee?: User;
```

### Relationship Summary

| Relationship | GORM | TypeORM |
|-------------|------|---------|
| HasMany | `[]Board` field + `gorm:"foreignKey:X"` | `@OneToMany()` decorator |
| BelongsTo | `User` field + FK field + `gorm:"foreignKey:X"` | `@ManyToOne()` + `@JoinColumn()` |
| Nullable FK | `*uint` (pointer) | `{ nullable: true }` |

---

## 6.7 Auto-Migration vs Manual Migration

### Auto-Migration (Yang Kita Pakai)

Di `cmd/api/main.go`:

```go
// Auto-migrate models
if err := db.AutoMigrate(&model.User{}, &model.Board{}, &model.Task{}); err != nil {
    log.Fatalf("Failed to migrate database: %v", err)
}
log.Println("Database migrated successfully")
```

`AutoMigrate` baca struct kamu, terus:
- Bikin table kalau belum ada
- Tambah column baru kalau ada field baru di struct
- Bikin index

**NestJS/TypeORM equivalent:** `synchronize: true` di config.

### Apa yang AutoMigrate **GA** lakuin?

- **Ga hapus column** — kalau kamu remove field dari struct, column tetap ada di DB
- **Ga rename column** — rename field = bikin column baru + column lama tetap ada
- **Ga ubah type column** — ga bisa ganti `VARCHAR(100)` jadi `VARCHAR(200)` otomatis
- **Ga hapus table** — drop table harus manual

### Manual Migration (Production Recommended)

Untuk production, harusnya pakai migration tool kayak **golang-migrate**:

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Bikin migration file
migrate create -ext sql -dir migrations -seq create_users_table
```

Hasilnya 2 file:
```
migrations/
├── 000001_create_users_table.up.sql
└── 000001_create_users_table.down.sql
```

```sql
-- 000001_create_users_table.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_deleted_at ON users(deleted_at);
```

```sql
-- 000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

Jalanin:
```bash
migrate -database "postgres://postgres:@localhost:5432/task_manager?sslmode=disable" -path migrations up
```

### Kapan Pakai Yang Mana?

| | AutoMigrate | Manual Migration |
|---|-------------|-----------------|
| Development | Cocok | Overkill |
| Production | **Bahaya** | **Wajib** |
| Team project | Risky | **Version controlled** |
| Rename/drop column | Ga bisa | Bisa |
| Rollback | Ga bisa | `migrate down` |
| Speed | Instant | Perlu bikin file SQL |

**NestJS comparison:**
| | GORM AutoMigrate | TypeORM synchronize | Prisma Migrate |
|---|-----------------|-------------------|----------------|
| Auto-sync | Ya | Ya | Tidak |
| Migration files | Tidak | Tidak | **Ya (auto-generate)** |
| Safe for prod | Tidak | Tidak | **Ya** |

> **Rekomendasi:** Pakai AutoMigrate waktu development (kayak kita sekarang). Begitu mau deploy ke production, switch ke golang-migrate. Kita bahas migration tool lebih detail di bab Docker/deployment.

---

## 6.8 Repository Pattern — CRUD Operations

Di project kita, semua database query ada di layer **repository**. Handler ga pernah langsung akses `*gorm.DB`. Ini pattern yang sama kayak NestJS best practice.

```
Handler → Service → Repository → Database
```

### File: `internal/repository/user.go`

```go
package repository

import (
	"github.com/irzan-aldi/task-manager/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll(page, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)

	offset := (page - 1) * limit
	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}
```

### Bedah Setiap Operation

#### CREATE
```go
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}
```

`db.Create(user)` — insert row baru. GORM otomatis:
- Generate SQL `INSERT INTO users (name, email, ...) VALUES (?, ?, ...)`
- Set `ID` di struct setelah insert (auto-increment)
- Set `CreatedAt` dan `UpdatedAt`

**NestJS/TypeORM equivalent:**
```typescript
async create(user: User): Promise<User> {
  return this.userRepository.save(user);
}
```

#### FIND BY ID
```go
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```

`db.First(&user, id)` = `SELECT * FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1`

Perhatiin: GORM otomatis tambahin `AND deleted_at IS NULL` karena model kita punya `gorm.DeletedAt` (soft delete — dijelaskan di section berikutnya).

**NestJS/TypeORM equivalent:**
```typescript
async findById(id: number): Promise<User> {
  return this.userRepository.findOne({ where: { id } });
}
```

#### FIND BY FIELD (Where clause)
```go
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```

`db.Where("email = ?", email)` — parameterized query. Tanda `?` itu placeholder — GORM otomatis escape nilainya (anti SQL injection).

**JANGAN PERNAH** tulis kayak gini:
```go
// BAHAYA! SQL Injection!
db.Where(fmt.Sprintf("email = '%s'", email)).First(&user)
```

Selalu pakai `?` placeholder.

**NestJS/TypeORM equivalent:**
```typescript
async findByEmail(email: string): Promise<User> {
  return this.userRepository.findOne({ where: { email } });
}
```

#### FIND ALL + PAGINATION
```go
func (r *UserRepository) FindAll(page, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)

	offset := (page - 1) * limit
	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}
```

Ini pattern pagination standard:
1. `Count(&total)` — hitung total records (untuk response metadata)
2. `Offset(offset).Limit(limit)` — ambil subset records
3. Return `users` (data), `total` (jumlah semua), `err`

Rumus offset: `(page - 1) * limit`
- Page 1, Limit 10 → Offset 0 → ambil record 1-10
- Page 2, Limit 10 → Offset 10 → ambil record 11-20
- Page 3, Limit 10 → Offset 20 → ambil record 21-30

**NestJS/TypeORM equivalent:**
```typescript
async findAll(page: number, limit: number) {
  const [users, total] = await this.userRepository.findAndCount({
    skip: (page - 1) * limit,
    take: limit,
  });
  return { users, total };
}
```

#### UPDATE
```go
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}
```

`db.Save(user)` — kalau `user.ID` ada → `UPDATE`. Kalau ga ada → `INSERT`. GORM otomatis update field `UpdatedAt`.

**Alternatif — update field tertentu aja:**
```go
// Update satu field
db.Model(&user).Update("name", "New Name")

// Update beberapa field
db.Model(&user).Updates(map[string]interface{}{
    "name":  "New Name",
    "email": "new@email.com",
})

// Update pakai struct (hati-hati: zero value ga ke-update)
db.Model(&user).Updates(model.User{Name: "New Name", Email: "new@email.com"})
```

> **Gotcha:** Kalau pakai `Updates` dengan struct, field yang nilainya zero value (`""`, `0`, `false`) **ga akan di-update**. Pakai `map[string]interface{}` kalau mau set field ke zero value.

**NestJS/TypeORM equivalent:**
```typescript
async update(id: number, data: Partial<User>) {
  await this.userRepository.update(id, data);
}
```

#### DELETE
```go
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}
```

Karena model User punya `gorm.DeletedAt`, ini **bukan** hard delete. GORM generate:
```sql
UPDATE users SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL
```

Bukan `DELETE FROM users WHERE id = ?`. Ini namanya **Soft Delete**.

---

## 6.9 Soft Delete — `gorm.DeletedAt`

Soft delete artinya data ga beneran dihapus dari database. Cuma di-tandain "sudah dihapus" dengan set field `deleted_at`.

### Gimana Cara Kerjanya?

Cukup tambahin field ini di model:

```go
DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
```

Setelah itu, GORM otomatis:

| Operation | SQL yang di-generate |
|-----------|---------------------|
| `db.Delete(&user, id)` | `UPDATE users SET deleted_at = NOW() WHERE id = ?` |
| `db.Find(&users)` | `SELECT * FROM users WHERE deleted_at IS NULL` |
| `db.First(&user, id)` | `SELECT * FROM users WHERE id = ? AND deleted_at IS NULL` |

Semua query otomatis filter `deleted_at IS NULL`. Data yang "dihapus" jadi invisible tapi masih ada di database.

### Mau Lihat Data yang Sudah Di-Delete?

```go
// Include soft-deleted records
db.Unscoped().Find(&users)

// Hard delete (beneran hapus dari DB)
db.Unscoped().Delete(&user, id)
```

### Kenapa Pakai Soft Delete?

1. **Recovery** — user ga sengaja hapus? Tinggal set `deleted_at = NULL`
2. **Audit trail** — data tetap ada untuk keperluan audit
3. **Referential integrity** — data yang di-reference table lain tetap ada
4. **Compliance** — GDPR/regulasi kadang butuh data retention

**NestJS/TypeORM equivalent:**
```typescript
@DeleteDateColumn()
deleted_at: Date;

// TypeORM juga otomatis soft delete kalau ada @DeleteDateColumn
await this.userRepository.softDelete(id);
```

---

## 6.10 Preloading Relationships — `Preload()`

Ini salah satu fitur paling powerful di GORM. Preload = load data relationship sekaligus.

### File: `internal/repository/board.go`

```go
package repository

import (
	"github.com/irzan-aldi/task-manager/internal/model"
	"gorm.io/gorm"
)

type BoardRepository struct {
	db *gorm.DB
}

func NewBoardRepository(db *gorm.DB) *BoardRepository {
	return &BoardRepository{db: db}
}

func (r *BoardRepository) Create(board *model.Board) error {
	return r.db.Create(board).Error
}

func (r *BoardRepository) FindByID(id uint) (*model.Board, error) {
	var board model.Board
	err := r.db.Preload("Owner").Preload("Tasks").First(&board, id).Error
	if err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *BoardRepository) FindByOwnerID(ownerID uint, page, limit int) ([]model.Board, int64, error) {
	var boards []model.Board
	var total int64

	query := r.db.Where("owner_id = ?", ownerID)
	query.Model(&model.Board{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Tasks").Offset(offset).Limit(limit).Find(&boards).Error
	return boards, total, err
}

func (r *BoardRepository) Update(board *model.Board) error {
	return r.db.Save(board).Error
}

func (r *BoardRepository) Delete(id uint) error {
	return r.db.Delete(&model.Board{}, id).Error
}
```

### Bedah Preload

#### Single Preload
```go
db.Preload("Owner").First(&board, id)
```

Ini generate **2 query**:
```sql
SELECT * FROM boards WHERE id = ? AND deleted_at IS NULL LIMIT 1;
SELECT * FROM users WHERE id = ? AND deleted_at IS NULL; -- Owner
```

Hasilnya `board.Owner` terisi data User. Tanpa Preload, `board.Owner` bakal kosong (zero value).

#### Multiple Preload
```go
db.Preload("Owner").Preload("Tasks").First(&board, id)
```

Generate **3 query**:
```sql
SELECT * FROM boards WHERE id = ? AND deleted_at IS NULL LIMIT 1;
SELECT * FROM users WHERE id = ? AND deleted_at IS NULL;
SELECT * FROM tasks WHERE board_id = ? AND deleted_at IS NULL;
```

#### Preload di List Query
```go
query.Preload("Tasks").Offset(offset).Limit(limit).Find(&boards)
```

Ini load Tasks untuk **setiap** board. GORM cukup pintar — ga pakai N+1 query. Dia batch jadi 1 query:
```sql
SELECT * FROM boards WHERE owner_id = ? AND deleted_at IS NULL LIMIT ? OFFSET ?;
SELECT * FROM tasks WHERE board_id IN (?, ?, ?) AND deleted_at IS NULL;
```

### Nested Preload

Mau load Board → Tasks → Assignee? Bisa!

```go
db.Preload("Tasks.Assignee").First(&board, id)
```

### Conditional Preload

Mau preload Tasks tapi cuma yang statusnya "done"?

```go
db.Preload("Tasks", "status = ?", "done").First(&board, id)
```

Atau pakai function:

```go
db.Preload("Tasks", func(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "done").Order("created_at DESC")
}).First(&board, id)
```

### NestJS/TypeORM Comparison

**TypeORM:**
```typescript
// Eager loading
const board = await this.boardRepository.findOne({
  where: { id },
  relations: ['owner', 'tasks'],
});

// Conditional
const board = await this.boardRepository
  .createQueryBuilder('board')
  .leftJoinAndSelect('board.tasks', 'task', 'task.status = :status', { status: 'done' })
  .where('board.id = :id', { id })
  .getOne();
```

**GORM:**
```go
// Eager loading
db.Preload("Owner").Preload("Tasks").First(&board, id)

// Conditional
db.Preload("Tasks", "status = ?", "done").First(&board, id)
```

Menurut gue, syntax GORM lebih clean dan readable.

### Task Repository — Preload in Action

```go
// internal/repository/task.go

func (r *TaskRepository) FindByID(id uint) (*model.Task, error) {
	var task model.Task
	err := r.db.Preload("Assignee").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) FindByBoardID(boardID uint, page, limit int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.Where("board_id = ?", boardID)
	query.Model(&model.Task{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Assignee").Offset(offset).Limit(limit).
		Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}
```

Perhatiin: setiap kali load Task, kita selalu `Preload("Assignee")` supaya client tau siapa yang di-assign ke task itu.

---

## 6.11 Query Builder — Where, Order, Offset, Limit

GORM punya chainable query builder yang mirip Prisma/TypeORM QueryBuilder.

### Contoh-contoh Query

```go
// WHERE sederhana
db.Where("email = ?", "irzan@email.com").First(&user)
// SELECT * FROM users WHERE email = 'irzan@email.com' LIMIT 1

// WHERE multiple conditions
db.Where("role = ? AND created_at > ?", "admin", time.Now().AddDate(0, -1, 0)).Find(&users)
// SELECT * FROM users WHERE role = 'admin' AND created_at > '2026-03-28' AND deleted_at IS NULL

// OR condition
db.Where("role = ?", "admin").Or("role = ?", "member").Find(&users)

// IN clause
db.Where("id IN ?", []uint{1, 2, 3}).Find(&users)
// SELECT * FROM users WHERE id IN (1, 2, 3)

// LIKE
db.Where("name LIKE ?", "%irzan%").Find(&users)

// NOT
db.Not("role = ?", "admin").Find(&users)

// ORDER BY
db.Order("created_at DESC").Find(&users)
db.Order("priority ASC, created_at DESC").Find(&users)

// LIMIT + OFFSET (pagination)
db.Offset(10).Limit(5).Find(&users)
// SELECT * FROM users LIMIT 5 OFFSET 10

// SELECT specific columns
db.Select("id", "name", "email").Find(&users)
// SELECT id, name, email FROM users

// COUNT
var count int64
db.Model(&model.User{}).Where("role = ?", "admin").Count(&count)

// GROUP BY + HAVING
type Result struct {
    Status string
    Count  int
}
var results []Result
db.Model(&model.Task{}).
    Select("status, COUNT(*) as count").
    Group("status").
    Having("COUNT(*) > ?", 5).
    Find(&results)
```

### Chaining Pattern dari Project Kita

Lihat gimana `FindByBoardID` di task repository chain beberapa method:

```go
err := query.
    Preload("Assignee").    // Load relationship
    Offset(offset).         // Skip N records
    Limit(limit).           // Take N records
    Order("created_at DESC"). // Sort newest first
    Find(&tasks).Error      // Execute query
```

Ini mirip kayak TypeORM QueryBuilder:
```typescript
const tasks = await this.taskRepository
  .createQueryBuilder('task')
  .leftJoinAndSelect('task.assignee', 'assignee')
  .where('task.boardId = :boardId', { boardId })
  .orderBy('task.created_at', 'DESC')
  .skip(offset)
  .take(limit)
  .getMany();
```

### Raw SQL — Kalau Query Terlalu Complex

```go
var results []model.Task
db.Raw(`
    SELECT t.* FROM tasks t
    JOIN boards b ON b.id = t.board_id
    WHERE b.owner_id = ?
    AND t.status = ?
    AND t.due_date < NOW()
    ORDER BY t.priority DESC
`, userID, "todo").Scan(&results)
```

`db.Raw()` untuk query yang terlalu complex untuk query builder. Hasilnya di-scan ke struct yang sama.

---

## 6.12 Transaction

Transaction penting kalau kamu perlu execute beberapa query yang **harus sukses semua** atau **gagal semua** (atomic).

### Contoh: Transfer Task ke Board Lain

```go
func (r *TaskRepository) TransferTask(taskID, newBoardID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil task
		var task model.Task
		if err := tx.First(&task, taskID).Error; err != nil {
			return err // rollback
		}

		// 2. Verify new board exists dan user punya akses
		var board model.Board
		if err := tx.First(&board, newBoardID).Error; err != nil {
			return err // rollback
		}

		if board.OwnerID != userID {
			return errors.New("not authorized") // rollback
		}

		// 3. Update task board
		task.BoardID = newBoardID
		if err := tx.Save(&task).Error; err != nil {
			return err // rollback
		}

		// Kalau semua sukses, commit otomatis
		return nil
	})
}
```

### Gimana Transaction Bekerja?

```go
db.Transaction(func(tx *gorm.DB) error {
    // Semua query di dalam sini pakai `tx`, bukan `db`

    if err := tx.Create(&something).Error; err != nil {
        return err // return error = ROLLBACK
    }

    if err := tx.Create(&somethingElse).Error; err != nil {
        return err // return error = ROLLBACK
    }

    return nil // return nil = COMMIT
})
```

Rules:
- Pakai `tx` (transaction object), **bukan** `db` (original)
- Return `error` → otomatis **ROLLBACK**
- Return `nil` → otomatis **COMMIT**

### NestJS/TypeORM Transaction Comparison

**TypeORM:**
```typescript
await this.connection.transaction(async (manager) => {
  await manager.save(task);
  await manager.save(board);
  // throw error = rollback, success = commit
});
```

**GORM:**
```go
db.Transaction(func(tx *gorm.DB) error {
    tx.Save(&task)
    tx.Save(&board)
    return nil // commit
})
```

Konsepnya sama persis. Cuma di Go, kamu return error (bukan throw exception).

### Manual Transaction (Kalau Butuh Lebih Kontrol)

```go
tx := db.Begin()

if err := tx.Create(&task).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Save(&board).Error; err != nil {
    tx.Rollback()
    return err
}

return tx.Commit().Error
```

> **Rekomendasi:** Pakai `db.Transaction()` (callback style). Lebih aman karena rollback otomatis kalau panic.

---

## 6.13 Wiring di main.go

Semua component di-wiring manual di `main.go`:

```go
func main() {
	// Load config
	cfg := config.Load()

	// Connect database
	db := config.ConnectDB(cfg)

	// Auto-migrate models
	if err := db.AutoMigrate(&model.User{}, &model.Board{}, &model.Task{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated successfully")

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Setup services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	boardService := service.NewBoardService(boardRepo)
	taskService := service.NewTaskService(taskRepo, boardRepo)

	// ... handler setup, router, server start
}
```

Flow dependency:
```
Config → DB Connection → Repository → Service → Handler → Router
```

Semua explicit. Ga ada magic DI container. Kalau ada dependency yang salah, compiler langsung kasih error.

**NestJS equivalent:**
```typescript
@Module({
  imports: [TypeOrmModule.forFeature([User, Board, Task])],
  providers: [UserService, BoardService, TaskService],
  controllers: [UserController, BoardController, TaskController],
})
export class AppModule {}
```

Di NestJS, DI container handle semuanya. Di Go, kamu handle sendiri — tapi jadi 100% transparan dan predictable.

---

## 6.14 Tips & Best Practices

### 1. Jangan Expose `*gorm.DB` ke Handler

```go
// JELEK — handler langsung akses database
func (h *Handler) GetUser(c *gin.Context) {
    var user model.User
    h.db.First(&user, id) // handler tau soal database? Bad separation!
}

// BAGUS — lewat repository
func (h *Handler) GetUser(c *gin.Context) {
    user, err := h.service.GetByID(id) // handler cuma tau service
}
```

### 2. Pakai Pointer untuk Optional Fields

```go
// Update request — semua field opsional
type UpdateBoardRequest struct {
	Title       *string `json:"title"`       // pointer = bisa nil
	Description *string `json:"description"` // pointer = bisa nil
}

// Di service
if req.Title != nil {
    board.Title = *req.Title
}
```

Kenapa pointer? Karena di Go, string kosong `""` dan "field ga dikirim" itu beda. Dengan pointer, kamu bisa bedain keduanya — `nil` artinya ga dikirim, `""` artinya memang mau dikosongkan.

### 3. Index Field yang Sering Di-Query

```go
Email string `gorm:"uniqueIndex"` // sering dipakai untuk login
DeletedAt gorm.DeletedAt `gorm:"index"` // semua query filter ini
```

### 4. Selalu Handle Error

```go
// Setiap operasi GORM, cek .Error
if err := db.Create(&user).Error; err != nil {
    return fmt.Errorf("create user: %w", err)
}
```

### 5. Log Level Berdasarkan Environment

```go
logLevel := logger.Info       // development — log semua SQL
if cfg.Env == "production" {
    logLevel = logger.Error   // production — cuma log error
}
```

---

## 6.15 Cheat Sheet: GORM vs TypeORM

| Operation | GORM | TypeORM |
|-----------|------|---------|
| **Connect** | `gorm.Open(postgres.Open(dsn))` | `TypeOrmModule.forRoot({...})` |
| **Model** | `struct` + tags | `class` + decorators |
| **Create** | `db.Create(&user)` | `repo.save(user)` |
| **Find one** | `db.First(&user, id)` | `repo.findOne({ where: { id } })` |
| **Find all** | `db.Find(&users)` | `repo.find()` |
| **Where** | `db.Where("email = ?", email)` | `repo.find({ where: { email } })` |
| **Update** | `db.Save(&user)` | `repo.save(user)` |
| **Delete** | `db.Delete(&user, id)` | `repo.softDelete(id)` |
| **Preload** | `db.Preload("Tasks")` | `relations: ['tasks']` |
| **Pagination** | `db.Offset(x).Limit(y)` | `skip(x).take(y)` |
| **Order** | `db.Order("created_at DESC")` | `order: { created_at: 'DESC' }` |
| **Transaction** | `db.Transaction(func(tx) error)` | `connection.transaction(async manager)` |
| **Raw SQL** | `db.Raw("SELECT...")` | `query("SELECT...")` |
| **Count** | `db.Model(&User{}).Count(&n)` | `repo.count()` |
| **Migrate** | `db.AutoMigrate(&User{})` | `synchronize: true` |
| **Soft delete** | `gorm.DeletedAt` field | `@DeleteDateColumn()` |

---

## 6.16 Checklist Bab Ini

- [ ] Paham kenapa PostgreSQL jadi pilihan utama (vs MySQL/MongoDB)
- [ ] Paham GORM sebagai ORM dan perbedaan dengan TypeORM/Prisma
- [ ] Paham DSN, connection pooling, dan config setup
- [ ] Bisa define model dengan struct tags (`gorm`, `json`)
- [ ] Paham relationship: HasMany, BelongsTo, nullable FK
- [ ] Paham perbedaan AutoMigrate vs manual migration
- [ ] Bisa implement CRUD: Create, FindByID, FindAll + pagination, Update, Delete
- [ ] Paham Soft Delete dan gimana `gorm.DeletedAt` bekerja
- [ ] Bisa pakai Preload untuk load relationships
- [ ] Paham query builder: Where, Order, Offset, Limit
- [ ] Paham gimana Transaction bekerja
- [ ] Bisa mapping setiap konsep ke equivalent NestJS/TypeORM
- [ ] Database connected dan AutoMigrate jalan di `main.go`

---

> **Next:** [Bab 7: Authentication — JWT + RBAC](07-authentication.md) — Kita bahas gimana protect API routes pakai JWT token dan role-based access control.
