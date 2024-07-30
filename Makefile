.PHONY: swagger docs

docs:
	swag init --generalInfo cmd/main.go --outputTypes=yaml

swagger:
	docker run --detach --name eryth-swagger -p 4000:8080 -e API_URL=/doc/swagger.yaml --mount 'type=bind,src=$(shell pwd)/docs,dst=/usr/share/nginx/html/doc' swaggerapi/swagger-ui
