go build -o jpkg cmd/main.go
go build -o jpx runner/main.go

cp jpkg ~/.amber/bin/jpkg
cp jpx ~/.amber/bin/jpx