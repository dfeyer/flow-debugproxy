Flow Framework Debug Proxy for xDebug
-------------------------------------

Flow Framework is a web application platform enabling developers creating
excellent web solutions and bring back the joy of coding. It gives you fast
results. It is a reliable foundation for complex applications.

The biggest pain with Flow Framework come from the the proxy class, the
framework do not execute your own code, but a precompiled version. This is
required for advanced feature, like AOP and the security framework. So working
with Flow is a real pleasure, but adding xDebug in the setup can be a pain.

This project is an xDebug proxy, written in Go, to take care of the mapping
between your PHP file and the proxy class.

Build your own
--------------

    # Get the dependecies
    go get
    # Build
    go Build

Run the proxy
-------------

    # Don't forget to change the configuration of your IDE to use port 9010
    flow-debugproxy -vv --framework flow

How to debug the proxy class directly
-------------------------------------

You can disable to path mapping, in this case the proxy do not process xDebug
protocol:

    ./flow-debugproxy --framework dummy

Show help
---------

    ./flow-debugproxy help

Use with Docker
---------------

##### 1. Preparation:

You will need:
1. Your (W)LAN IP address.
2. Your docker-machine's IP address. CMD: `docker-machine ip default` (use 127.0.0.1 on linux)
3. A compiled flow-debugproxy binary
4. Access to the PHP container's php.ini
5. xdebug must be working

##### 2. Installation & Debugging:

1. Copy the flow-debugproxy binary to your container or a mounted folder.
2. Identify and set these environment variables or replace them in the upcoming commands:
* `HOST_IP` (Primary (W)LAN address of your device)
* `XDEBUG_PORT` (PhpStorm settings: Language & Framework -> PHP -> Debug: Xdebug: Debug Port)
* `FLOW_DEBUG_BIN_PATH` (The path of the binary **inside** the container)
3. Set following php.ini values in your php/web container: (xdebug will now try to (only) connect to the php container itself.)
```
xdebug.remote_host = 127.0.0.1
xdebug.remote_port = 9002
```
4. Start the debug proxy (Replace `$(docker-compose ps -q app)` with your container if you don't use docker-compose)
```
docker exec -e PHP_IDE_CONFIG='serverName=app' $(docker-compose ps -q app) ${FLOW_DEBUG_BIN_PATH} --xdebug 0.0.0.0:9002 --ide ${HOST_IP}:${XDEBUG_PORT} > /dev/null &
```
5. Enable your IDE's xdebug listener, ensure xdebug is enabled (e.g. if you use bookm)
6. Use `xdebug_break()` in your code to force your first break.
7. The first time you have to configure your IDE with the popup that should open on the first `xdebug_break();` hit. (Or "Click to set up path mappings" in your debug console UI)
If not, configure your PHP Server settings yourself.

Debugging the debugger:

Start the debug proxy with verbose flags if it does not connect to your IDE.
The debug proxy does not quit after stopping the process that started it. You have to kill it in the container manually.

Hint:

If you use the env variable `FLOW_PATH_TEMPORARY_BASE`, please be sure to keep
`Data/Temporary` inside the path, without this the mapper will not detect the
proxy classes.

```
FLOW_PATH_TEMPORARY_BASE=/tmp/flow/Data/Temporary
```

##### Using with --framework dummy

If your debugging target is the code generated by Flow's AOP Framework then you can start the debugging proxy with --framework dummy
In that case it won't remap from the generated code to your source but "pass through" the debugger steps.
To see what's going on you have to have the generated code in a folder visible to your IDE (in your project).
You can either abstain from `FLOW_PATH_TEMPORARY_BASE` or set it to a path that is in your IDE's project.

Acknowledgments
---------------

Development sponsored by [ttree ltd - neos solution provider](http://ttree.ch).

This project is highly inspired by the PHP based Debug proxy:
https://github.com/sandstorm/debugproxy thanks to the Sandstorm team. The goal
of the Go version of the proxy is to solve the performance issue that the PHP
version has.

We try our best to craft this package with a lots of love, we are open to
sponsoring, support request, ... just contact us.

License
-------

Licensed under MIT, see [LICENSE](LICENSE)
