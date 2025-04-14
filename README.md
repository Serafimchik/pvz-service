Инструкция по запуску

PS C:\pvz-service> docker-compose up --build
PS C:\pvz-service> goose -dir migrations postgres "postgres://user:password@localhost:5432/pvz?sslmode=disable" up
PS C:\pvz-service\cmd> go run main.go

запускал тесты следующим образом
PS C:\pvz-service> go test ./... -v
PS C:\pvz-service> go test ./... -coverprofile=C:\pvz-service\cover.out
PS C:\pvz-service> go tool cover -func=C:\pvz-service\cover.out
PS C:\pvz-service> go tool cover -html=C:\pvz-service\cover.out -o C:\pvz-service\cover.html
PS C:\pvz-service> start C:\pvz-service\cover.html
