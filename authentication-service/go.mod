module github.com/andreiTn/go-micro/authentication-service

go 1.27.3

require (
	github.com/andreiTn/go-micro/pkg v0.0.0
	github.com/go-chi/chi/v5 v5.5.0
	github.com/jackc/pgx/v4 v4.18.3
	golang.org/x/crypto v0.58.0
)

replace github.com/andreiTn/go-micro/pkg => ../pkg

require (
	github.com/go-chi/cors v1.2.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/stretchr/testify v1.12.1 // indirect
	golang.org/x/text v0.43.0 // indirect
)
