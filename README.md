# Prometheus Shell Command Runner

Periodically run a command and exports its return code as a prometheus metric.

```
Usage: ./prom-run (options) command...
  -listen-addr string
    	Address to listen on (default ":8080")
  -metric-name string
    	Metric name of exporter status code (default "status_code")
  -period duration
    	Period with which to run the command. (default 10s)
```

The output of the command is also made availible via HTTP at /.

To run this, try:

```
$ docker run -p 8080:8080 -ti tomwilkie/prom-run echo foo
2016/06/02 11:49:08 Listening on address :8080
2016/06/02 11:49:18 Running 'echo' with argments [foo]
2016/06/02 11:49:18 Command exited successfully
```
