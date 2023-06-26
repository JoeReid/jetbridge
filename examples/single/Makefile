.PHONY: run
run: run.sh .lambda-arn .stream-created
	./run.sh

.PHONY: clean
clean:
	-rm handler handler.zip .lambda-arn

.stream-created: .docker-up stream.json
	nats stream add --config stream.json
	touch .stream-created

.lambda-arn: handler.zip .docker-up
	LAMBDA_ARN=$(shell awslocal lambda create-function \
	    --function-name my-function \
	    --runtime go1.x \
	    --zip-file fileb://handler.zip \
	    --role arn:aws:iam::123456:role/irrelevant \
	    --environment Variables={FOO=bar} \
	    --query 'FunctionArn' \
	    --output text); \
	echo $$LAMBDA_ARN > .lambda-arn

handler.zip: main.go
	go build -o handler ./main.go
	zip handler.zip handler

.docker-up:
	docker compose build
	docker compose up -d
	touch .docker-up

.PHONY: docker-down
docker-down:
	docker-compose down
	-rm .docker-up
