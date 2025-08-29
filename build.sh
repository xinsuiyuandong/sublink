GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o sublink_linux_amd64 main.go
GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o sublink_linux_arm64 main.go
GOOS=windows  GOARCH=amd64  go build -ldflags="-w -s" -o sublink_windows_amd64.exe main.go
