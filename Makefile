.PHONY: otel_up
otel_up:
	docker-compose -f ./otel_jaeger/docker-compose.yml up

otel_down:
	docker-compose -f ./otel_jaeger/docker-compose.yml down -v

################################################################################################################
