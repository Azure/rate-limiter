module go.goms.io/rate-limiter-backed-by-redis-cache/cmd/azureredis

go 1.20

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gorilla/mux v1.8.0
	pkg/redisclient v0.0.0-00010101000000-000000000000
	pkg/handlers v0.0.0-00010101000000-000000000000
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.4.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis/v2 v2.3.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-jwt/jwt/v5 v5.0.0 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	go.goms.io/rate-limiter-backed-by-redis-cache/tokenbucket v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)

replace (
	go.goms.io/rate-limiter-backed-by-redis-cache/tokenbucket => ../../tokenbucket
	pkg/redisclient => ../../pkg/redisclient
	pkg/handlers => ../../pkg/handlers
)
