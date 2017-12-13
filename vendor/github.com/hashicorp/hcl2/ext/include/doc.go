// Package include implements a zcl extension that allows inclusion of
// one HCL body into another using blocks of type "include", with the following
// structure:
//
//     include {
//       path = "./foo.hcl"
//     }
//
// The processing of the given path is delegated to the calling application,
// allowing it to decide how to interpret the path and which syntaxes to
// support for referenced files.
package include
