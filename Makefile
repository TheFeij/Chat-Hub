postgres:
	docker run --name postgres-container --network chat-server-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=1234 -d postgres:16rc1-alpine3.18

create-db:
	docker exec -it postgres-container createdb --username=root --owner=root chat_server

drop-db:
	docker exec -it postgres-container dropdb chat_server

create-test-db:
	docker exec -it postgres-container createdb --username=root --owner=root chat_server_test

drop-test-db:
	docker exec -it postgres-container dropdb chat_server_test

mockdb:
	mockgen -package mockdb -destination repository/mock/repository.go Chat-Server/repository Repository

mockmaker:
	mockgen -package mockmaker -destination token/mock/maker.go Chat-Server/token Maker


start-server:
	go run ./main.go

test:
	go test -v --cover ./...

