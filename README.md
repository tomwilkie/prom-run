# Prometheus Shell Command Runner

Periodically run a command and exports its return code as a prometheus metric.

```
$ docker run -ti tomwilkie/prom-run -h
Usage: ./prom-run (options) command...
  -listen-addr string
    	Address to listen on (default ":9152")
  -period duration
    	Period with which to run the command. (default 10s)
  -timeout duration
    	Amount of time to give the command to run. (default 10m0s)
```

The output of the command is also made availible via HTTP at /.

To run this, try:

```
$ docker run -p 9152:9152 -ti tomwilkie/prom-run echo foo
INFO[0000] Listening on address :9152                   
INFO[0010] Running 'echo' with argments [foo]           
INFO[0010] Command exited successfully 
```
