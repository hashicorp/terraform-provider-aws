# Terraform AWS Provider - Contributor Guide

The AWS Provider is the work of thousands of contributors, and is maintained by a small team within HashiCorp. There are many way to contribute, and we have extensive documentation on them. Below are the main ways most users engage with the provider repository and the high level steps required to ensure the contribution goes smoothly.

## What would you like do do?

### Add a New AWS Service to the Provider

Allow Terraform (via the AWS Provider) to manage an entirely new AWS service by introducing the resources and datasources required to manage configuration of the service.

- [Add Service Client](add-a-new-service.md#add-a-service-client)
- [Perform Service Design](add-a-new-service.md#perform-service-design)
- [Add a New Resource](add-a-new-service.md#add-a-new-resource)

### Add a New Resource to the Provider

Allow the management of a logical resource within AWS by adding a new resource to the Terraform AWS Provider.

- [Name the Resource](add-a-new-resource.md#add-a-new-resource)
- [Fork the Provider](add-a-new-resource.md#fork-the-provider)
- [Fill out the Resource Schema](add-a-new-resource.md#fill-out-the-resource-schema)
- [Implement CRUD Handlers](add-a-new-resource.md#implement-crud-handlers)
- [Write Passing Acceptance Tests](add-a-new-resource.md#write-passing-acceptance-tests)
- [Create Documentation for the Resource](add-a-new-resource.md#create-documentation-for-the-resource)
- [Ensure Format and Lint Checks are Passing](add-a-new-resource.md#ensure-format-and-link-checks-are-passing)
- [Raise a Pull Request](raising-a-pull-request.md)
- [Wait for Prioritization](prioritization.md)

### Add a New Datasource to the Provider

Allow reading an existing AWS resource by id or filter, and allow its use within Terraform configurations.


- [Name the Datasource](add-a-new-datasource.md#add-a-new-datasource)
- [Fork the Provider](add-a-new-datasource.md#fork-the-provider)
- [Fill out the Datasource Schema](add-a-new-datasource.md#fill-out-the-datasource-schema)
- [Implement CRUD Handlers](add-a-new-datasource.md#implement-crud-handlers)
- [Write Passing Acceptance Tests](add-a-new-datasource.md#write-passing-acceptance-tests)
- [Create Documentation for the Resource](add-a-new-resource.md#create-documentation-for-the-resource)
- [Ensure Format and Lint Checks are Passing](add-a-new-datasource.md#ensure-format-and-link-checks-are-passing)
- [Raise a Pull Request](raising-a-pull-request.md)
- [Wait for Prioritization](prioritization.md)

### Add a Newly Released AWS Region

New regions are immediately usable with the provider with the caveat that a configuration workaround is required to skip validation of the region during cli operations. A small set of changes are required to makes this workaround necessary.

- [Add a New Region](add-a-new-region.md)


### Fix a bug, or make an enhancement to a resource or datasource

These constitute the majority of pull requests submitted, many of which we address and merge regardless of priority in our regular internal gardening days.

- [Making Small Changes to Existing Resources.](bugs-and-enhancements.md)

### Add Resource Name Generation Support

Allow a resource, to either fully, or partially generate its own resource names. TODO:This can be useful.

- [Resource Name Generation](resource-name-generation.md)

### Add Tagging support for an existing Resource

Many AWS resources allow [assigning metadata via tags](https://docs.aws.amazon.com/general/latest/gr/aws_tagging.html) which can be enormously useful to manage and track cloud resources from a governance and cost perspective. Often though, new aws services or features are launched without tagging support, which can follow later. In this case tagging support will need to be explicitly added to those resources. Typically this support follows an established pattern so support is normally straightforward to add.

- [Add tag support for a resource](resource-tagging)

### Add Import Support for a Resource

Adding import support allows `terraform import` to be run targeting an existing unmanaged resource and pulling its configuration into Terraform state. Typically import support is added during initial resource implementation but in some cases this may not have been done.

- [Add Import Support](add-import-support.md)

### End User Documentation Changes

The provider documentation is displayed on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) and is sourced and refreshed from the provider repository during the release process.

- [How to Modify Provider Documentation](documentation-changes.md)

### Report a bug, or make a feature request

In both cases follow the template laid out in the submission forms. Details are important, it can take a considerable amount of time for the maintainers to reproduce issues. So bug reports without full reproductions are likely to be uninvestigated, or in some case closed.

- [Report a Bug](https://github.com/hashicorp/terraform-provider-aws/issues/new?assignees=&labels=&template=Bug_Report.md)
- [Submit a Feature Request](https://github.com/hashicorp/terraform-provider-aws/issues/new?assignees=&labels=enhancement&template=Feature_Request.md)
