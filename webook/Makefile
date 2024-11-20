.PHONY: docker
docker:
	# 把上次编译的删掉
	@rm webook || true
	@docker rmi -f liupch6/webook:v0.0.1
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker build -t liupch6/webook:v0.0.1 .