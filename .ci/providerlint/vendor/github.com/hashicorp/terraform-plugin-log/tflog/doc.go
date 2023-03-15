// Package tflog provides helper functions for writing log output and creating
// loggers for Terraform plugins.
//
// For most plugin authors, building on an SDK or framework, the SDK or
// framework will take care of injecting a logger using New.
//
// tflog also allows plugin authors to create subsystem loggers, which are
// loggers for sufficiently distinct areas of the codebase or concerns. The
// benefit of using distinct loggers for these concerns is doing so allows
// plugin authors and practitioners to configure different log levels for each
// subsystem's log, allowing log output to be turned on or off without
// recompiling.
package tflog
