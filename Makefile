list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

run-redis: rm-redis
	docker run -d --name redis -p 6379:6379 redis:alpine

rm-redis: 
	docker rm -f redis || true
	
run-jaeger: rm-jaeger
	docker run -d --name local-jaeger \
	-p5775:5775/udp \
	-p16686:16686 \
	jaegertracing/all-in-one:latest

rm-jaeger:
	docker rm -f local-jaeger || true