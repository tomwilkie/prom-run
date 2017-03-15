all: .uptodate

prom-run: main.go
	GOOS=linux go build -o $@ .

.uptodate: prom-run Dockerfile
	docker build -t tomwilkie/prom-run .

clean:
	rm -f prom-rum .uptodate