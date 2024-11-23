.PHONY: mock
mock:
	@mockgen -package=svcmocks -source=./webook/internal/service/user.go -destination=./webook/internal/service/mocks/user.mock.go
	@mockgen -package=svcmocks -source=./webook/internal/service/code.go -destination=./webook/internal/service/mocks/code.mock.go
	@go mod tidy