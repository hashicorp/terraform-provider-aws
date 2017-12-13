# HCL

HCL is a structured configuration language that is both human- and
machine-friendly, for use with command-line tools. Although intended to be
generally useful, it is primarily targeted towards devops tools, servers, etc.

HCL has both a _native syntax_, intended to be pleasant to read and write for
humans, and a JSON-based variant that is easier for machines to generate
and parse.

The HCL native syntax is inspired by [libucl](https://github.com/vstakhov/libucl),
[nginx configuration](http://nginx.org/en/docs/beginners_guide.html#conf_structure),
and others.

It includes an expression syntax that allows basic inline computation and,
with support from the calling application, use of variables and functions
for more dynamic configuration languages.

## Experimental HCL2

This repository contains the experimental version 2 of HCL. This new version
combines the initial iteration of HCL with the interpolation language HIL
to produce a single configuration language that supports arbitrary expressions.

At this time the HCL2 syntax and the Go API are still evolving.
Backward-compatibility is not guaranteed and so any application using this
library should use vendoring.

The new implementation has a completely new parser and Go API, with no direct
migration path. Although the syntax is similar, the implementation takes some
very different approaches to improve on some "rough edges" that existed with
the original implementation and to allow for more robust error handling.

Once this new implementation reaches stability, its package paths will be
changed to reflect that it is the _current_ HCL implementation. At that time,
the original implementation will be archived.

## Why?

Newcomers to HCL often ask: why not JSON, YAML, etc?

Whereas JSON and YAML are formats for serializing data structures, HCL is
a syntax and API specifically designed for building structured configuration
formats.

HCL attempts to strike a compromise between generic serialization formats
such as JSON and configuration formats built around full programming languages
such as Ruby. HCL syntax is designed to be easily read and written by humans,
and allows _declarative_ logic to permit its use in more complex applications.

HCL is intended as a base syntax for configuration formats built
around key-value pairs and heirarchical blocks whose structure is well-defined
by the calling application, and this definition of the configuration structure
allows for better error messages and more convenient definition within the
calling application.

It can't be denied that JSON is very convenient as a _lingua franca_
for interoperability between different pieces of software. Because of this,
HCL defines a common configuration model that can be parsed from either its
native syntax or from a well-defined equivalent JSON structure. This allows
configuration to be provided as a mixture of human-authored configuration
files in the native syntax and machine-generated files in JSON.

## Information Model and Syntax

HCL is built around two primary concepts: _attributes_ and _blocks_. In
native syntax, a configuration file for a hypothetical application might look
something like this:

```hcl
io_mode = "async"

service "http" "web_proxy" {
  listen_addr = "127.0.0.1:8080"
  
  process "main" {
    command = ["/usr/local/bin/awesome-app", "server"]
  }

  process "mgmt" {
    command = ["/usr/local/bin/awesome-app", "mgmt"]
  }
}
```

The JSON equivalent of this configuration is the following:

```json
{
  "io_mode": "async",
  "service": {
    "http": {
      "web_proxy": {
        "listen_addr": "127.0.0.1:8080",
        "process": {
          "main": {
            "command": ["/usr/local/bin/awesome-app", "server"]
          },
          "mgmt": {
            "command": ["/usr/local/bin/awesome-app", "mgmt"]
          },
        }
      }
    }
  }
}
```

Regardless of which syntax is used, the API within the calling application
is the same. It can either work directly with the low-level attributes and
blocks, for more advanced use-cases, or it can use one of the _decoder_
packages to declaratively extract into either Go structs or dynamic value
structures.

Attribute values can be expressions as well as just literal values:

```hcl
# Arithmetic with literals and application-provided variables
sum = 1 + addend

# String interpolation and templates
message = "Hello, ${name}!"

# Application-provided functions
shouty_message = upper(message)
```

Although JSON syntax doesn't permit direct use of expressions, the interpolation
syntax allows use of arbitrary expressions within JSON strings:

```json
{
  "sum": "${1 + addend}",
  "message": "Hello, ${name}!",
  "shouty_message": "${upper(message)}"
}
```

For more information, see the detailed specifications:

* [Syntax-agnostic Information Model](hcl/spec.md)
* [HCL Native Syntax](hcl/hclsyntax/spec.md)
* [JSON Representation](hcl/json/spec.md)

## Acknowledgements

HCL was heavily inspired by [libucl](https://github.com/vstakhov/libucl),
by [Vsevolod Stakhov](https://github.com/vstakhov).

HCL and HIL originate in [HashiCorp Terraform](https://terraform.io/),
with the original parsers for each written by
[Mitchell Hashimoto](https://github.com/mitchellh).

The original HCL parser was ported to pure Go (from yacc) by
[Fatih Arslan](https://github.com/fatih). The structure-related portions of
the new native syntax parser build on that work.

The original HIL parser was ported to pure Go (from yacc) by
[Martin Atkins](https://github.com/apparentlymart). The expression-related
portions of the new native syntax parser build on that work.

HCL2, which merged the original HCL and HIL languages into this single new
language, builds on design and prototyping work by
[Martin Atkins](https://github.com/apparentlymart) in
[zcl](https://github.com/zclconf/go-zcl).
