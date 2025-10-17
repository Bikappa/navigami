
build-frontend:
	npm --prefix frontend run build

build-server:
	go build -o server cmd/server/main.go
