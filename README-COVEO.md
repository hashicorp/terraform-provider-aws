# Naming a version

Previously, we were using a tagging scheme where we added `-coveo.n` to official tagged version.

However, this causes problems when a user is restricting the version of its provider. If a user add:

```hcl
provider "aws" {
  version = ">= 2.28.1"
  region  = var.region
}
```

or:

```hcl
terraform {
  required_version = ">= 0.12.9"
  required_providers {
    aws        = ">= 2.52.0"
    local      = ">= 1.4"
    null       = ">= 2.1"
    template   = ">= 2.1"
    random     = ">= 2.1"
    kubernetes = ">= 1.11.1"
  }
}```

Then, our forked version will be ignored and the official one will be used.

To avoid this, starting with version 2.70, we will now release our forked version by adding 10 to patch version.

```text
terraform-provider-aws_v2.70.0 => terraform-provider-aws_v2.70.10
terraform-provider-aws_v2.70.1 => terraform-provider-aws_v2.70.11
```
