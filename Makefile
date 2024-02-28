IMG=ghcr.io/sonalys/fake

.PHONYE: all

build:
	@CGO_ENABLED=0 go build -o ./bin/fake ./entrypoint/cli/main.go

image: build
	@docker build -t ${IMG}:latest -f dockerfile .

push:
	@docker push ${IMG}

build_all:
	@GOOS=windows GOARCH=amd64 go build -o ./bin/windows/amd64/fake.exe ./entrypoint/cli/main.go
	@GOOS=linux GOARCH=amd64 go build -o ./bin/linux/amd64/fake ./entrypoint/cli/main.go
	@GOOS=linux GOARCH=arm64 go build -o ./bin/linux/arm64/fake ./entrypoint/cli/main.go
	@mkdir releases
	@zip releases/fake_Windows_amd64.zip bin/windows/amd64/fake.exe
	@zip releases/fake_Linux_amd64.zip bin/linux/amd64/fake
	@zip releases/fake_Linux_arm64.zip bin/linux/arm64/fake