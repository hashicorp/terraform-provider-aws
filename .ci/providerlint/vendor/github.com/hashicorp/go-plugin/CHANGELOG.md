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


