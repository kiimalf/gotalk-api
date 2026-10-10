# GoTalk API

**GoTalk API** adalah layanan backend RESTful API untuk komunikasi pesan langsung (*1-on-1 Direct Messaging*) privat antar-pengguna, yang dibangun menggunakan bahasa pemrograman **Go (Golang)**, framework **Go Fiber v2**, dan basis data relasional **PostgreSQL**.

API ini dirancang sebagai layanan *headless* yang ringan, aman, dan berlatensi rendah. Setiap percakapan diwadahi dalam ruang obrolan (*chat room*) unik yang dikelola secara otomatis oleh sistem tanpa memerlukan relasi pertemanan terlebih dahulu (*friendship-less model*).

Sistem ini berfokus pada privasi dan keandalan pertukaran pesan teks, dilengkapi dengan autentikasi stateless JWT dengan rotasi *refresh token*, pembatasan laju login (*rate limiting*), penelusuran riwayat pesan berbasis kursor (*cursor pagination*), kontrol kepemilikan pesan (*edit* dan *soft-delete*), serta kontrol akses berbasis peran (*Role-Based Access Control* / RBAC).

---

## Panduan Instalasi & Menjalankan Sistem

Ikuti langkah-langkah berikut untuk menyiapkan dan menjalankan GoTalk API di lingkungan lokal Anda.

### 1. Persyaratan Sistem
Sebelum memulai, pastikan perangkat Anda telah terpasang:
* **Go Compiler:** Versi 1.22 atau yang lebih baru (disarankan v1.26+)
* **PostgreSQL:** Versi 15 atau yang lebih baru

### 2. Konfigurasi Lingkungan (.env)
Salin berkas template konfigurasi `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```

Buka berkas `.env` lalu sesuaikan pengaturan koneksi basis data dan kunci rahasia JWT:
```env
APP_NAME=gotalk-api
APP_PORT=3000
ALLOWED_ORIGINS=*

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=rahasia
DB_NAME=gotalk_db
DB_SSLMODE=disable
DB_MAX_CONNS=10

JWT_SECRET=super_secret_jwt_key_minimal_32_karakter_acak!
JWT_ISSUER=gotalk-api
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7
```
> **Catatan:** Nilai `JWT_SECRET` wajib memiliki panjang minimal 32 karakter demi keamanan sistem.

### 3. Migrasi Basis Data
Pastikan basis data PostgreSQL telah dibuat terlebih dahulu:
```bash
createdb -h localhost -U postgres gotalk_db
```

Jalankan seluruh skrip SQL migrasi yang tersedia pada direktori `migrations/` secara berurutan:
```bash
for file in migrations/*.sql; do
    echo "Menjalankan migrasi: $file"
    psql -h localhost -U postgres -d gotalk_db -f "$file"
done
```

### 4. Menjalankan Server
1. Unduh seluruh dependensi Go:
   ```bash
   go mod download
   ```
2. Jalankan server:
   ```bash
   go run main.go
   ```
Server akan aktif dan siap menerima permintaan pada: `http://localhost:3000`

### 5. Pengujian Otomatis (Opsional)
Untuk memverifikasi kebenaran fungsi logika bisnis, kriptografi, dan aturan hak akses, jalankan pengujian unit bawaan:
```bash
go test -v ./tests/...
```

---

## Tautan Repositori
* **GitHub:** [https://github.com/kiimalf/gotalk-api](https://github.com/kiimalf/gotalk-api)
