
#export GOOS=linux
#GOOS=linux GOARCH=386 go build -o myselfgo
# CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o myselfgo main.go
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o myselfgo main.go