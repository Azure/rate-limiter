module github.com/Azure/rate-limiter/test/handlers

go 1.21.3

require (
	github.com/gorilla/mux v1.8.0
	pkg/tokenbucket v0.0.0-00010101000000-000000000000
	ratelimiter v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	pkg/cache v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	pkg/cache => ../../pkg/cache
	pkg/tokenbucket => ../../pkg/tokenbucket
	ratelimiter => ../../ratelimiter
)
