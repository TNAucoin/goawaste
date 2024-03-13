.PHONY: build
build:
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap main.go

.PHONY: zip
zip:
	zip myfunction.zip bootstrap

.PHONY: clean
clean:
	rm -f bootstrap myfunction.zip

.PHONY: upload
upload:
	aws lambda update-function-code --function-name myFunction \
    --zip-file fileb://myfunction.zip

.PHONY: deploy
deploy: build zip upload clean