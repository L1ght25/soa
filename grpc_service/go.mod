module grpc_service

go 1.22.0

require (
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	common v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
)

require (
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.9.0
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)

replace common => ../common
