build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o main ./main.go
	zip ur-checker.zip main

deploy:
	aws lambda update-function-code \
	--function-name ur-checker \
	--zip-file fileb://ur-checker.zip \
	--profile ${profile} \
	--region ap-northeast-1