docker:
	@rm demojob || true
	@docker rmi -f liupch6/demojob:v0.0.1
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags=k8s -o demojob .
	@docker build -t liupch6/demojob:v0.0.1 .