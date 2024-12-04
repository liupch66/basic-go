.PHONY: mock
mock:
	@mockgen -package=loggermocks -source=./webook/pkg/logger/type.go -destination=./webook/pkg/logger/mock/logger.mock.go
	@mockgen -package=jwtHdlmocks -source=./webook/internal/web/jwt/types.go -destination=./webook/internal/web/jwt/mock/handler.mcok.go
	@mockgen -package=svcmocks -source=./webook/internal/service/user.go -destination=./webook/internal/service/mocks/user.mock.go
	@mockgen -package=svcmocks -source=./webook/internal/service/code.go -destination=./webook/internal/service/mocks/code.mock.go
	@mockgen -package=svcmocks -source=./webook/internal/service/article.go -destination=./webook/internal/service/mocks/article.mock.go
	@mockgen -package=repomocks -source=./webook/internal/repository/user.go -destination=./webook/internal/repository/mocks/user.mock.go
	@mockgen -package=repomocks -source=./webook/internal/repository/code.go -destination=./webook/internal/repository/mocks/code.mock.go
	@mockgen -package=daomocks -source=./webook/internal/repository/dao/user.go -destination=./webook/internal/repository/dao/mocks/user.mock.go
	@mockgen -package=cachemocks -source=./webook/internal/repository/cache/user.go -destination=./webook/internal/repository/cache/mocks/user.mock.go
	@mockgen -package=cachemocks -source=./webook/internal/repository/cache/code.go -destination=./webook/internal/repository/cache/mocks/code.mock.go
	# 为 redis Cmdable 生成 mock 实现
	@mockgen -package=redismocks -destination=./webook/internal/repository/cache/redismocks/cmdable.mock.go github.com/redis/go-redis/v9 Cmdable
	@mockgen -package=artRepomocks -source=webook/internal/repository/article/article.go -destination=webook/internal/repository/article/mocks/article.mock.go
	@mockgen -package=artRepomocks -source=webook/internal/repository/article/article_author.go -destination=webook/internal/repository/article/mocks/article_author.mock.go
	@mockgen -package=artRepomocks -source=webook/internal/repository/article/article_reader.go -destination=webook/internal/repository/article/mocks/article_reader.mock.go
	@mockgen -package=smsmocks -source=./webook/internal/service/sms/types.go -destination=./webook/internal/service/sms/mocks/svc.mock.go
	@mockgen -package=limitmocks -source=./webook/pkg/ratelimit/types.go -destination=./webook/pkg/ratelimit/mocks/limit.mock.go
	@go mod tidy