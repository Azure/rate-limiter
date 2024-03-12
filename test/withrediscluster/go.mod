module github.com/Azure/distributed-rate-limiter/test/withrediscluster

go 1.21

require (
	github.com/gorilla/mux v1.8.0
	pkg/cache v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
)

replace (
	pkg/cache => ../../pkg/cache
	pkg/tokenbucket => ../../pkg/tokenbucket
	ratelimiter => ../../ratelimiter
)
