# CGO_ENABLED=0 GOOS=linux go build -a ./cmd/vaultd/
# CGO_ENABLED --> C 종속성에 바인딩 x 빌드 크기 줄일 수 있음
FROM scratch
ADD vaultd vaultd
EXPOSE 8080 8081
ENTRYPOINT [ "/vaultd" ]