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
