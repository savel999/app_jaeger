.PHONY: otel_up
otel_up:
	URL="http://localhost:16686/search"
	OS=$(uname -s)

	echo $OS

	case "${OS}" in
		Linux*)     x-www-browser "$URL" || xdg-open "$URL" || echo "Не удалось открыть браузер";;
		Darwin*)    open "$URL";;
		*)          echo "Неизвестная ОС";;
	esac

	docker-compose -f ./jaeger_v2/docker-compose.yml up

otel_down:
	docker-compose -f ./jaeger_v2/docker-compose.yml down -v

################################################################################################################

#.PHONY: otel_tempo_up
otel_tempo_up:
	x-www-browser http://localhost:3000/explore
	docker-compose -f ./otel_tempo/docker-compose.yml up

otel_tempo_down:
	docker-compose -f ./otel_tempo/docker-compose.yml down -v

################################################################################################################
