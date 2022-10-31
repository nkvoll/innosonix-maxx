.PHONY = generate-openapi
generate-openapi:
	oapi-codegen -old-config-style -generate client,types -package api pkg/api/maxx_rest_openapi.yml > pkg/api/maxx.gen.go

.PHONY = build
build:
	go build -o out/imctl cmd/main/main.go

docker-build:
	docker build -t nkvoll/innosonix-maxx:latest .

docker-push:
	docker push docker.io/nkvoll/innosonix-maxx:latest