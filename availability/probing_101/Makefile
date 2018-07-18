WD:=$(shell pwd)

start-stack:
	docker volume prune -f
	cd stack \
	&& docker-compose down \
	&& docker-compose up

start-postgres-stack:
	# start postgres
	docker stop postgres
	docker container rm postgres
	docker run \
		-p "5432:5432" \
		-e POSTGRES_USER=root \
		-e POSTGRES_PASSWORD=root \
		-v "$(WD)/postgres-stack/config:/docker-entrypoint-initdb.d" \
		--name=postgres \
		-d \
		postgres:9.6

	# wait until initialized
	./postgres-stack/bin/wait-for-postgres.sh

	# start cloudprober
	docker stop cloudprober
	docker container rm cloudprober
	docker run \
		-p "9313:9313" \
		-v "$(WD)/postgres-stack/config/cloudprober-singletarget.cfg:/etc/cloudprober.cfg" \
		--network=host \
		--name cloudprober \
		-d \
		cloudprober/cloudprober

.PHONY: start-stack start-postgres-stack