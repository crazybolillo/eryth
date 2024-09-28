.PHONY: swagger docs db

docs:
	swag init --generalInfo cmd/main.go --outputTypes=yaml

swagger:
	docker compose up --wait swagger

db:
	docker compose up --wait db
