# Setup Go Environment — ITSD3/SD3

Panduan standarisasi environment Go untuk developer di departemen ITSD3/SD3.

---

## Daftar Isi

- [Persyaratan](#persyaratan)
- [Instalasi](#instalasi)
- [Struktur Folder](#struktur-folder)
- [Verifikasi Instalasi](#verifikasi-instalasi)
- [Penggunaan goswitch](#penggunaan-goswitch)
  - [Melihat Versi Tersedia](#melihat-versi-tersedia)
  - [Switch Versi Go](#switch-versi-go)
  - [Versi Tidak Ditemukan](#versi-tidak-ditemukan)
- [Konfigurasi VS Code](#konfigurasi-vs-code)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

---

## Persyaratan

| Kebutuhan         | Keterangan                              |
| ----------------- | --------------------------------------- |
| OS                | Windows 10 / 11 (64-bit)               |
| Disk              | Minimal 2 GB ruang kosong di drive `D:` |
| Koneksi Internet  | Diperlukan saat instalasi awal          |
| Hak Akses         | Administrator (script akan meminta otomatis) |

---

## Instalasi

### Langkah 1 — Download Script

Salin file **`setup_go_env.bat`** ke komputer Anda. Lokasi bebas, contoh:

```
D:\SD3\setup_go_env.bat
```

### Langkah 2 — Jalankan Script

**Double-click** file `setup_go_env.bat`. Script akan:

1. Meminta akses Administrator secara otomatis
2. Membuat folder `D:\go-apps\go-sdks\`
3. Mengunduh Go **1.24.6** dan Go **1.25.8** dari [go.dev/dl](https://go.dev/dl/)
4. Mengekstrak kedua SDK ke folder masing-masing
5. Mengatur environment variables (`GOROOT`, `GOPATH`, `PATH`)
6. Memverifikasi instalasi

> ⏳ **Catatan:** Proses download membutuhkan waktu tergantung kecepatan internet.
> Jangan tutup terminal selama proses berjalan.

### Langkah 3 — Restart Terminal

Setelah script selesai, **tutup terminal lalu buka terminal baru**.

Jalankan:

```bat
go version
```

Output yang diharapkan:

```
go version go1.25.8 windows/amd64
```

### Langkah 4 — Setup goswitch (Opsional, Direkomendasikan)

Salin file **`goswitch.bat`** ke `D:\go-apps\goswitch.bat`, lalu jalankan perintah ini **satu kali** agar bisa dipanggil dari mana saja:

```bat
setx PATH "D:\go-apps;%PATH%"
```

Restart terminal setelahnya.

---

## Struktur Folder

Setelah instalasi berhasil, struktur folder akan terlihat seperti ini:

```
D:\go-apps\
├── go-sdks\
│   ├── go1.24.6\          ← Go SDK versi 1.24.6
���   │   └── bin\
│   │       └── go.exe
│   └── go1.25.8\          ← Go SDK versi 1.25.8 (default)
│       └── bin\
│           └── go.exe
├── gopath\                ← GOPATH (modules, binaries)
│   ├── bin\
│   ├── pkg\
│   └── src\
└── goswitch.bat           ← Tool untuk switch versi Go
```

### Environment Variables

| Variable | Value                              | Keterangan                    |
| -------- | ---------------------------------- | ----------------------------- |
| `GOROOT` | `D:\go-apps\go-sdks\go1.25.8`     | Berubah saat switch versi     |
| `GOPATH` | `D:\go-apps\gopath`                | Tetap, tidak berubah          |
| `PATH`   | Includes `%GOROOT%\bin` & `%GOPATH%\bin` | Diatur otomatis oleh script |

---

## Verifikasi Instalasi

Buka **terminal baru** dan jalankan perintah berikut:

```bat
:: Cek versi Go aktif
go version

:: Cek environment
go env GOROOT
go env GOPATH

:: Cek instalasi berfungsi
go env
```

Contoh output:

```
> go version
go version go1.25.8 windows/amd64

> go env GOROOT
D:\go-apps\go-sdks\go1.25.8

> go env GOPATH
D:\go-apps\gopath
```

---

## Penggunaan goswitch

`goswitch` adalah tool sederhana untuk berpindah antar versi Go yang sudah terinstall.

### Melihat Versi Tersedia

```bat
goswitch
```

Output:

```
  GOSWITCH - Go Version Switcher
  ==============================

  Penggunaan: goswitch <versi>
  Contoh:     goswitch 1.24.6

  Versi tersedia:
       1.24.6
     *  1.25.8

  * = versi aktif saat ini
```

### Switch Versi Go

```bat
goswitch 1.24.6
```

Output:

```
  [PASS] Switched ke Go 1.24.6
  ========================
  go version go1.24.6 windows/amd64

  GOROOT = D:\go-apps\go-sdks\go1.24.6
  GOPATH = D:\go-apps\gopath
```

Switch kembali ke versi default:

```bat
goswitch 1.25.8
```

Output:

```
  [PASS] Switched ke Go 1.25.8
  ========================
  go version go1.25.8 windows/amd64

  GOROOT = D:\go-apps\go-sdks\go1.25.8
  GOPATH = D:\go-apps\gopath
```

> **Catatan:** `goswitch` langsung aktif di terminal saat ini DAN tersimpan permanen
> untuk terminal baru. Tidak perlu restart terminal setelah switch.

### Versi Tidak Ditemukan

```bat
goswitch 1.99.0
```

Output:

```
  [FAIL] Go 1.99.0 tidak ditemukan di D:\go-apps\go-sdks\go1.99.0

  Versi tersedia:
    1.24.6
    1.25.8
```

---

## Konfigurasi VS Code

### 1. Install Extension Go

Buka VS Code → Extensions (`Ctrl+Shift+X`) → Cari **"Go"** by Go Team at Google → Install.

### 2. Konfigurasi Settings

Buka Settings (`Ctrl+,`) → Cari `go.goroot`, atau tambahkan di `settings.json`:

```json
{
    "go.goroot": "D:\\go-apps\\go-sdks\\go1.25.8",
    "go.gopath": "D:\\go-apps\\gopath"
}
```

> **Tip:** Setelah `goswitch`, restart VS Code agar membaca `GOROOT` yang baru,
> atau ubah `go.goroot` di settings secara manual.

### 3. Install Go Tools

Buka Command Palette (`Ctrl+Shift+P`) → ketik **"Go: Install/Update Tools"** → pilih semua → OK.

---

## Troubleshooting

### `'go' is not recognized as an internal or external command`

**Penyebab:** Terminal belum membaca environment variables yang baru.

**Solusi:**
1. Tutup **semua** terminal/command prompt yang terbuka
2. Buka terminal baru
3. Jalankan `go version`

Jika masih error, verifikasi PATH secara manual:

```bat
echo %GOROOT%
echo %PATH%
```

Pastikan output `%GOROOT%` tidak kosong dan `%PATH%` mengandung `D:\go-apps\go-sdks\go1.25.8\bin`.

---

### `'goswitch' is not recognized`

**Penyebab:** Folder `D:\go-apps` belum ditambahkan ke PATH.

**Solusi:**

```bat
setx PATH "D:\go-apps;%PATH%"
```

Restart terminal.

---

### `GOROOT is empty` atau `The system cannot find the path specified`

**Penyebab:** Versi Go yang diminta belum terinstall.

**Solusi:**

```bat
:: Cek versi apa saja yang terinstall
dir D:\go-apps\go-sdks\

:: Switch ke versi yang ada
goswitch 1.25.8
```

---

### Script instalasi gagal download

**Penyebab:** Koneksi internet bermasalah atau versi Go tidak tersedia.

**Solusi:**
1. Pastikan bisa akses [https://go.dev/dl/](https://go.dev/dl/) dari browser
2. Pastikan tidak ada proxy/firewall yang memblokir
3. Jalankan ulang `setup_go_env.bat`

> Script akan skip versi yang sudah terinstall, jadi aman dijalankan berkali-kali.

---

### Module download error / `GOPROXY` timeout

**Penyebab:** Default proxy `https://proxy.golang.org` tidak bisa diakses.

**Solusi:**

```bat
go env -w GOPROXY=https://proxy.golang.org,direct
```

Atau jika di belakang corporate proxy:

```bat
set HTTP_PROXY=http://proxy-server:port
set HTTPS_PROXY=http://proxy-server:port
go mod tidy
```

---

## FAQ

### Apakah saya harus pakai drive D:?

Ya, script ini dikonfigurasi untuk `D:\go-apps`. Jika ingin mengubah lokasi,
edit variabel `$targetRoot` di `setup_go_env.bat` dan `SDK_ROOT` di `goswitch.bat`.

### Bisakah saya install versi Go tambahan?

Bisa. Download zip dari [go.dev/dl](https://go.dev/dl/), ekstrak ke
`D:\go-apps\go-sdks\go<VERSI>\`, lalu gunakan `goswitch <VERSI>`.

Contoh menambah Go 1.23.8:

1. Download `go1.23.8.windows-amd64.zip` dari [go.dev/dl](https://go.dev/dl/)
2. Ekstrak isi folder `go` di dalam zip ke `D:\go-apps\go-sdks\go1.23.8\`
3. Verifikasi: `D:\go-apps\go-sdks\go1.23.8\bin\go.exe` harus ada
4. Switch: `goswitch 1.23.8`

### Apakah goswitch mengubah versi Go secara permanent?

Ya. `goswitch` mengatur `GOROOT` secara permanen (via `setx`) DAN untuk terminal
saat ini. Terminal baru akan otomatis menggunakan versi terakhir yang di-switch.

### Apakah aman menjalankan setup_go_env.bat berkali-kali?

Ya. Script akan skip download untuk versi yang sudah ada di `D:\go-apps\go-sdks\`.

### Bagaimana cara uninstall?

1. Hapus folder `D:\go-apps\`
2. Hapus environment variables:
   ```bat
   setx GOROOT ""
   setx GOPATH ""
   ```
3. Hapus `D:\go-apps` dan `D:\go-apps\go-sdks\...` dari PATH secara manual
   melalui **System Properties → Environment Variables**

---

## File Reference

| File               | Lokasi                  | Fungsi                                     |
| ------------------- | ----------------------- | ------------------------------------------ |
| `setup_go_env.bat`  | Bebas (misal `D:\SD3\`) | Instalasi awal Go SDK + environment setup  |
| `goswitch.bat`      | `D:\go-apps\`           | Switch antar versi Go yang terinstall      |

---

> **Maintainer:** Tim ITSD3/SD3
> **Last Updated:** 2026-03-30