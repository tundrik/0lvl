pub:
	go run -race cmd/publisher/main.go

run:
	go run -race cmd/orderstorage/main.go

vet:
	go vet cmd/orderstorage/main.go

lint:
	golint cmd/orderstorage/main.go
