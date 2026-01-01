module github.com/oomph-ac/example-proxy

go 1.24.6

require (
	github.com/cooldogedev/spectrum v0.0.43
	github.com/df-mc/dragonfly v0.10.9
	github.com/getsentry/sentry-go v0.35.3
	github.com/oomph-ac/oconfig v0.0.0-20250912013507-a80d378a6595
	github.com/oomph-ac/oomph v0.0.0-20251002033530-3dd27115da92
	github.com/sandertv/gophertunnel v1.52.0
	golang.org/x/exp v0.0.0-20251002181428-27f1f14c8bb9
)

require (
	github.com/brentp/intintmap v0.0.0-20190211203843-30dc0ade9af9 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chewxy/math32 v1.10.1 // indirect
	github.com/cooldogedev/spectral v0.0.5 // indirect
	github.com/df-mc/goleveldb v1.1.9 // indirect
	github.com/df-mc/jsonc v1.0.5 // indirect
	github.com/df-mc/worldupgrader v1.0.20 // indirect
	github.com/ethaniccc/float32-cube v0.0.0-20250511224129-7af1f8c4ee12 // indirect
	github.com/francoispqt/gojay v1.2.13 // indirect
	github.com/go-gl/mathgl v1.2.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hjson/hjson-go/v4 v4.4.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/quic-go/quic-go v0.55.0 // indirect
	github.com/sandertv/go-raknet v1.14.3-0.20250305181847-6af3e95113d6 // indirect
	github.com/scylladb/go-set v1.0.2 // indirect
	github.com/segmentio/fasthash v1.0.3 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/oauth2 v0.32.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
)

replace github.com/oomph-ac/oomph => ./deps/oomph

replace github.com/df-mc/dragonfly => ./deps/dragonfly

replace github.com/oomph-ac/oconfig => ./deps/oconfig

replace github.com/cooldogedev/spectrum => ./deps/spectrum

replace github.com/sandertv/gophertunnel => ./deps/gophertunnel
