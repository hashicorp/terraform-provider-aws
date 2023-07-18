## v1.4.10

BUG FIXES:

* additional notes: ensure to close files [GH-241](https://github.com/hashicorp/go-plugin/pull/241)]

ENHANCEMENTS:

* deps: Remove direct dependency on golang.org/x/net [GH-240](https://github.com/hashicorp/go-plugin/pull/240)]

## v1.4.9

ENHANCEMENTS:

* client: Remove log warning introduced in 1.4.5 when SecureConfig is nil. [[GH-238](https://github.com/hashicorp/go-plugin/pull/238)]

## v1.4.8

BUG FIXES:

* Fix windows build: [[GH-227](https://github.com/hashicorp/go-plugin/pull/227)]

## v1.4.7

ENHANCEMENTS:

* More detailed error message on plugin start failure: [[GH-223](https://github.com/hashicorp/go-plugin/pull/223)]

## v1.4.6

BUG FIXES:

* server: Prevent gRPC broker goroutine leak when using `GRPCServer` type `GracefulStop()` or `Stop()` methods [[GH-220](https://github.com/hashicorp/go-plugin/pull/220)]

## v1.4.5

ENHANCEMENTS:

* client: log warning when SecureConfig is nil [[GH-207](https://github.com/hashicorp/go-plugin/pull/207)]


## v1.4.4

ENHANCEMENTS:

* client: increase level of plugin exit logs [[GH-195](https://github.com/hashicorp/go-plugin/pull/195)]

BUG FIXES:

* Bidirectional communication: fix bidirectional communication when AutoMTLS is enabled [[GH-193](https://github.com/hashicorp/go-plugin/pull/193)]
* RPC: Trim a spurious log message for plugins using RPC [[GH-186](https://github.com/hashicorp/go-plugin/pull/186)]
