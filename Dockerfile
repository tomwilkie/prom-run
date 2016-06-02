FROM alpine
ADD prom-run ./
ENTRYPOINT ["./prom-run"]

