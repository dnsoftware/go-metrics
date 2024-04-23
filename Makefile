PROJECT="go-metrics"

default:
	echo ${PROJECT}

test:
	go test -v -count=1 ./...

.PHONY: default test cover
cover:
	go test -v -coverpkg=./... -coverprofile=coverage.out -covermode=count ./...
	go tool cover -func coverage.out | grep total | awk '{print $3}'

.PHONY: gen
gen:
	mockgen -source=internal/server/collector/collector.go -destination=internal/server/collector/mocks/mock_backup_storage.go


