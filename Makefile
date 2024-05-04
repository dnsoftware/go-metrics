PROJECT="go-metrics"
DATABASE_DSN="postgres://praktikum:praktikum@127.0.0.1:5532/praktikum?sslmode=disable"


default:
	echo ${PROJECT}

test:
	DATABASE_DSN=${DATABASE_DSN} go test -v -count=1 ./...

.PHONY: default test cover
cover:
	DATABASE_DSN=${DATABASE_DSN} go test -v -coverpkg=./... -coverprofile=coverage.out -covermode=count ./...
	go tool cover -func coverage.out | grep total | awk '{print $3}'

.PHONY: gen
gen:
	mockgen -source=internal/server/collector/collector.go -destination=internal/server/collector/mocks/mock_backup_storage.go

cover2:
	DATABASE_DSN=${DATABASE_DSN} go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm coverage.out

.PHONY: staticlint
staticlint:
	go build -o ./staticlint ./cmd/staticlint
	go vet -vettool=staticlint ./...
