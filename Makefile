.PHONY: mock
mock:
	@mockgen -package=svcmocks -source=./webook/internal/service/user.go -destination=./webook/internal/service/mocks/user.mock.go
	@mockgen -package=svcmocks -source=./webook/internal/service/code.go -destination=./webook/internal/service/mocks/code.mock.go
	@mockgen -package=repomocks -source=./webook/internal/repository/user.go -destination=./webook/internal/repository/mocks/user.mock.go
	@mockgen -package=repomocks -source=./webook/internal/repository/code.go -destination=./webook/internal/repository/mocks/code.mock.go
	@mockgen -package=daomocks -source=./webook/internal/repository/dao/user.go -destination=./webook/internal/repository/dao/mocks/user.mock.go
	@mockgen -package=cachemocks -source=./webook/internal/repository/cache/user.go -destination=./webook/internal/repository/cache/mocks/user.mock.go
	@mockgen -package=cachemocks -source=./webook/internal/repository/cache/code.go -destination=./webook/internal/repository/cache/mocks/code.mock.go
	# 为 redis Cmdable 生成 mock 实现
	@mockgen -package=redismocks -destination=./webook/internal/repository/cache/redismocks/cmdable.mock.go github.com/redis/go-redis/v9 Cmdable
	@go mod tidy