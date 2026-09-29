# todo-app

Project belajar Go dengan pendekatan microservice sederhana. Terdiri dari dua service yang saling berkomunikasi:

- **`todo-svc`** — CRUD Todo (create, read, update, delete)
- **`notif-svc`** — menerima event dari `todo-svc` saat ada todo baru dibuat, lalu mencatat/menampilkan notifikasi

Dikembangkan sebagai monorepo menggunakan [Go Workspace](https://go.dev/ref/mod#workspaces) (`go.work`), supaya tiap service tetap punya `go.mod` independen tapi bisa dikembangkan bareng tanpa perlu publish package.

## Struktur folder

```
todo-app/
├── services/
│   ├── todo-svc/
│   └── notif-svc/
├── go.work
└── README.md
```

## Progress belajar

### Setup awal
- [x] Buat struktur folder `services/todo-svc` dan `services/notif-svc`
- [x] `go mod init` di masing-masing service
- [x] Buat `go.work` di root, hubungkan kedua service
- [x] Init git repo & push ke GitHub

### todo-svc
- [x] Server HTTP dasar (`net/http`), endpoint `/health`
- [ ] Model `Todo` (ID, Title, Done, CreatedAt)
- [ ] Repository in-memory (Create, GetAll, GetByID, Update, Delete)
- [ ] Service layer (validasi, logic bisnis)
- [ ] Handler & routing (CRUD via HTTP)
- [ ] Test end-to-end pakai curl/Postman

### notif-svc
- [ ] Server HTTP dasar (`net/http`), endpoint `/health`
- [ ] Endpoint `/notify` untuk menerima event dari `todo-svc`
- [ ] Log notifikasi ke console saat event diterima

### Integrasi antar service
- [ ] `todo-svc` memanggil `notif-svc` (HTTP call) saat todo baru dibuat
- [ ] Jalankan kedua service bersamaan, test alur end-to-end
- [ ] (Opsional) Buat `pkg/shared` untuk struct yang dipakai bersama, tambahkan ke `go.work`

### Eksplorasi lanjutan (setelah alur dasar jalan)
- [ ] Ganti in-memory repository ke database asli (PostgreSQL)
- [ ] Containerize tiap service dengan Docker + `docker-compose.yml`
- [ ] Ganti komunikasi HTTP sinkron ke message broker (RabbitMQ/NATS/Kafka)
- [ ] Tambah observability (structured logging, tracing)
- [ ] Tambah unit test & mock untuk tiap layer

## Cara menjalankan (development)

```bash
# Terminal 1
cd services/todo-svc
go run ./cmd

# Terminal 2
cd services/notif-svc
go run ./cmd
```

## Catatan
- `go.work` di-commit ke repo karena project ini dikerjakan solo — supaya siap jalan langsung setelah clone ulang tanpa perlu setup ulang workspace.
- Struktur internal tiap service mengikuti pola layered architecture: `handler` → `service` → `repository`.
