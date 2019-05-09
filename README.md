Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 0.10+
- [Go](https://golang.org/doc/install) 1.12 (to build the provider plugin)

Building the Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-aws`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-aws
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-aws
$ make build
```

Using the Provider
----------------------
If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it. Documentation about the provider specific configuration options can be found on the [provider's website](https://www.terraform.io/docs/providers/aws/index.html).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.12+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-aws
...
```

Testing the Provider
---------------------------

In order to test the provider, you can simply run `make test`.

*Note:* Make sure no `AWS_ACCESS_KEY_ID` or `AWS_SECRET_ACCESS_KEY` variables are set, and there's no `[default]` section in the AWS credentials file `~/.aws/credentials`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

Contributing
---------------------------

Terraform is the work of thousands of contributors. We appreciate your help!

To contribute, please read the contribution guidelines: [Contributing to Terraform - AWS Provider](.github/CONTRIBUTING.md)

Issues on GitHub are intended to be related to bugs or feature requests with provider codebase. See https://www.terraform.io/docs/extend/community/index.html for a list of community resources to ask questions about Terraform.

