# 🏗️ Fixio Backend — Forum Kritik & Solusi Kebijakan Publik

REST API backend untuk platform forum di mana masyarakat menyampaikan **kritik terhadap kebijakan** dan **mengusulkan solusi**.

## Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| Language | Go 1.25+ |
| Framework | Fiber v2 |
| ORM | GORM v2 |
| Database | PostgreSQL 16+ |
| Cache | Redis |
| Auth | OAuth 2.0 (Google) + JWT |
| Config | Viper |

## Architecture

Go Clean Architecture dengan 5-folder module pattern:

```
cmd/main.go → server → module.go (DI Wiring) → Controllers → Services → Repositories → DB
```

Setiap module mengikuti pattern:
```
modules/<module>/
├── entity/       # GORM models
├── dto/          # Request/Response + validation
├── repository/   # Data access layer
├── usecase/      # Business logic
└── delivery/http/ # HTTP controllers
```

## Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL 16+
- Redis (optional)

### Setup

```bash
# 1. Clone repository
git clone <repo-url> fixio-backend
cd fixio-backend

# 2. Copy environment config
cp .env.example .env
# Edit .env with your database & OAuth credentials

# 3. Install dependencies
go mod tidy

# 4. Run the server
make run
# or
go run cmd/main.go
```

Server will start at `http://localhost:8080`

### API Endpoints

**Auth (OAuth):**
- `GET /api/auth/google` — Login via Google
- `GET /api/auth/google/callback` — Google callback
- `GET /api/auth/me` — Get current user (auth required)
- `POST /api/auth/logout` — Logout

**Posts:**
- `GET /api/posts` — List approved posts (filter, sort, paginate)
- `GET /api/posts/:id` — Post detail
- `POST /api/posts` — Create post (auth required)
- `PUT /api/posts/:id` — Edit post (owner only)
- `DELETE /api/posts/:id` — Delete post (owner/admin)

**Votes:**
- `POST /api/posts/:id/vote` — Vote up/down (auth required)
- `DELETE /api/posts/:id/vote` — Remove vote

**Comments:**
- `GET /api/posts/:id/comments` — List comments
- `POST /api/posts/:id/comments` — Add comment (auth required)
- `DELETE /api/comments/:id` — Delete comment (owner/admin)

**Users:**
- `GET /api/users/:id` — User profile + posts

**Moderation (Moderator+ only):**
- `GET /api/moderation/queue` — Pending review queue
- `PUT /api/moderation/posts/:id/approve` — Approve post
- `PUT /api/moderation/posts/:id/reject` — Reject post
- `GET /api/moderation/history` — Review history

**Admin (Administrator only):**
- `GET /api/admin/users` — List all users
- `PUT /api/admin/users/:id/role` — Change user role
- `GET /api/admin/stats` — Platform statistics

### Roles

| Role | Akses |
|------|-------|
| Creator | Create posts, comments, votes |
| Moderator | + Approve/reject posts, view queue |
| Administrator | + Manage users, sectors, regions, stats |

## Development

```bash
make run          # Run server
make dev          # Run with hot-reload
make test         # Run tests
make swagger      # Generate API docs
make tidy         # Tidy dependencies
```

## License

Private — Hak cipta dilindungi.
