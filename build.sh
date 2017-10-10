go build uploader.go
docker build -t kuberlab/file-uploader:latest -f Dockerfile .
