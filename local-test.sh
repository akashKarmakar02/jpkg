go build -o jpkg cmd/main.go
go build -o jpx runner/main.go

cp jpkg ~/.amber/bin/jpkg.tmp
mv ~/.amber/bin/jpkg.tmp ~/.amber/bin/jpkg

cp jpx ~/.amber/bin/jpx.tmp
mv ~/.amber/bin/jpx.tmp ~/.amber/bin/jpx
