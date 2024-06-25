.PHONY: otel_up
otel_up:
	x-www-browser http://localhost:16686/search
	docker-compose -f ./otel_jaeger/docker-compose.yml up

otel_down:
	docker-compose -f ./otel_jaeger/docker-compose.yml down -v

################################################################################################################

#.PHONY: otel_tempo_up
otel_tempo_up:
	x-www-browser http://localhost:3000/explore
	docker-compose -f ./otel_tempo/docker-compose.yml up

otel_tempo_down:
	docker-compose -f ./otel_tempo/docker-compose.yml down -v

################################################################################################################
