// Package tfsdklog provides helper functions for logging from SDKs and
// frameworks for building Terraform plugins.
//
// Plugin authors shouldn't need to use this package; it is meant for authors
// of the frameworks and SDKs for plugins. Plugin authors should use the tflog
// package.
//
// This package provides very similar functionality to tflog, except it uses a
// separate namespace for its logs.
package tfsdklog
