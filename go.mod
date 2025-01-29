module github.com/green-ecolution/tbz-baumkataster-plugin

go 1.23.3

replace github.com/green-ecolution/green-ecolution-backend => ../green-ecolution-management/green-ecolution-backend

replace github.com/green-ecolution/green-ecolution-backend/client => ../green-ecolution-management/green-ecolution-backend/pkg/client

replace github.com/green-ecolution/green-ecolution-backend/plugin => ../green-ecolution-management/green-ecolution-backend/pkg/plugin

require (
	github.com/green-ecolution/green-ecolution-backend/client v0.0.0-00010101000000-000000000000
	github.com/green-ecolution/green-ecolution-backend/plugin v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/omniscale/go-proj/v2 v2.0.0-20221006090944-6c8a5f5a510d
	golang.org/x/oauth2 v0.0.0-20210323180902-22b0adad7558
)

require (
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/green-ecolution/green-ecolution-backend v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)
