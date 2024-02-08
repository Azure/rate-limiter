module test/withazureredis

go 1.21

require (
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.5.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis/v2 v2.3.0
	github.com/gorilla/mux v1.8.0
	pkg/cache v0.0.0-00010101000000-000000000000
	test/handlers v0.0.0-00010101000000-000000000000
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.9.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	pkg/tokenbucket v0.0.0-00010101000000-000000000000 // indirect
	ratelimiter v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	pkg/cache => ../../pkg/cache
	pkg/tokenbucket => ../../pkg/tokenbucket
	ratelimiter => ../../ratelimiter
	test/handlers => ../handlers
)
