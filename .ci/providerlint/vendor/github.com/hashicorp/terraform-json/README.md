# terraform-json

[![GoDoc](https://godoc.org/github.com/hashicorp/terraform-json?status.svg)](https://godoc.org/github.com/hashicorp/terraform-json)

This repository houses data types designed to help parse the data produced by
two [Terraform](https://www.terraform.io/) commands:

* [`terraform show -json`](https://www.terraform.io/docs/commands/show.html#json-output)
* [`terraform providers schema -json`](https://www.terraform.io/docs/commands/providers/schema.html#json)

While containing mostly data types, there are also a few helpers to assist with
working with the data.

This repository also serves as de facto documentation for the formats produced
by these commands. For more details, see the
[GoDoc](https://godoc.org/github.com/hashicorp/terraform-json).

## Should I use this library?

This library was built for a few specific applications, and is not intended for
general purpose use.

The Terraform core team **recommends against** using `terraform-json` if your
application has any of the following requirements:

* **Forward-compatibility**: each version of this library represents a specific
  snapshot of the [Terraform JSON output format](https://developer.hashicorp.com/terraform/internals/json-format),
  and it often slightly lags behind Terraform itself. The library supports
  [the 1.x compatibility promises](https://developer.hashicorp.com/terraform/language/v1-compatibility-promises)
  but you will need to upgrade the version promptly to use new additions. If you
  require full compatibility with future Terraform versions, we recommend
  implementing your own custom decoders for the parts of the JSON format you need.
* **Writing JSON output**: the structures in this library are not guaranteed to emit
  JSON data which is semantically equivalent to Terraform itself. If your application
  must robustly write JSON data to be consumed by systems which expect Terraform's
  format to be supported, you should implement your own custom encoders.
* **Filtering or round-tripping**: the Terraform JSON formats are designed to be
  forwards compatible, and permit new attributes to be added which may safely be
  ignored by earlier versions of consumers. This library **drops unknown attributes**,
  which means it is unsuitable for any application which intends to filter data
  or read-modify-write data which will be consumed downstream. Any application doing
  this will silently drop new data from new versions. For this application, you should
  implement a custom decoder and encoder which preserves any unknown attributes
  through a round-trip.

When is `terraform-json` suitable? We recommend using it for applications which
decode the core stable data types and use it directly, and don't attempt to emit
JSON to be consumed by applications which expect the Terraform format.

## Why a separate repository?

To reduce dependencies on any of Terraform core's internals, we've made a design
decision to make any helpers or libraries that work with the external JSON data
external and not a part of the Terraform GitHub repository itself.

While Terraform core will change often and be relatively unstable, this library
will see a smaller amount of change. Most of the major changes have already
happened leading up to 0.12, so you can expect this library to only see minor
incremental changes going forward.

For this reason, `terraform show -json` and `terraform providers schema -json`
is the recommended format for working with Terraform data externally, and as
such, if you require any help working with the data in these formats, or even a
reference of how the JSON is formatted, use this repository.
