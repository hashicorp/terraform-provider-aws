# HCL Extensions

This directory contains some packages implementing some extensions to HCL
that add features by building on the core API in the main `hcl` package.

These serve as optional language extensions for use-cases that are limited only
to specific callers. Generally these make the language more expressive at
the expense of increased dynamic behavior that may be undesirable for
applications that need to impose more rigid structure on configuration.
