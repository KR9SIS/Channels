#DEV
build-dev:
	docker build -t channels -f containers/images/Dockerfile . && docker build -t turn  -f containers/images/Dockerfile .

clean-dev:
	docker-compose -f containers/composes/dc.dev.yml.down

run-dev:
	docker-compose -f containers/composes/dc.dev.yml up
