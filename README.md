Commands to run application:

1-) go mod init edatasance/subscription-manager 
2-) go get .
3-) DB_HOST=192.168.1.21 DB_PORT=3306 DB_USERNAME=test DB_PASSWORD=test DB_NAME=testdb APPLICATION_PORT=3535 go run main.go

Trigger application example commands:

1-) curl http://localhost:3535/subscription 
2-) curl-X POST -H "Content-Type: application/json" -d '{"id": 1, "subscription_info": "examplebrk"}' http://localhost:3535/subscription

PS: MYSQL database should be installed before those operations . In addition to this "subscription_table" should be created
