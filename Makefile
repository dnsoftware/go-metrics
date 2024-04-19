PROJECT="go-metrics"

default:
	echo ${PROJECT}

test:
	go test -v -count=1 ./...

.PHONY: default test cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm coverage.out

.PHONY: gen
gen:
	mockgen -source=internal/server/collector/collector.go -destination=internal/server/collector/mocks/mock_backup_storage.go
