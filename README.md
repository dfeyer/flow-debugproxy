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

Currently this project is under development and not ready for a daily usage.

Build your own
--------------

    # Get the dependecies
    go get
    # Build
    go Build

Run the proxy
-------------

    # Don't forget to change the configuration of your IDE to use port 9010
    ./flow-debugproxy --xdebug 127.0.0.1:9000 --ide 127.0.0.1:9010 --vv

Show help
---------

    ./flow-debugproxy help

Acknowledgments
---------------

Development sponsored by [ttree ltd - neos solution provider](http://ttree.ch).

We try our best to craft this package with a lots of love, we are open to
sponsoring, support request, ... just contact us.

License
-------

Licensed under MIT, see [LICENSE](LICENSE)
