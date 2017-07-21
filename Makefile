all: .uptodate

prom-run: main.go
	GOOS=linux go build -ldflags "-extldflags \"-static\" -linkmode=external" -o $@ .

.uptodate: prom-run Dockerfile
	docker build -t tomwilkie/prom-run .

clean:
	rm -f prom-rum .uptodate
