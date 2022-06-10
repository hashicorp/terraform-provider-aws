# Terraform AWS Provider - Contributor Guide

<p>The AWS Provider is the work of thousands of contributors, and is maintained by a small team within HashiCorp. There are many way to contribute, and we have extensive documentation on them. Below are the main ways most users engage with the provider repository and the high level steps required to ensure the contribution goes smoothly.</p>

## What would you like do do?

### Add a new AWS Service to the Provider

Allow Terraform (via the AWS Provider) to manage an entirely new AWS service by introducing the resources and datasources required to manage configuration of the service.

- [Add Service Client](add-a-new-service#add-a-service-client)
- [Perform Service Design](add-a-new-service#perform-service-design)
- [Add a new resource](add-a-new-service#add-a-new-resource)

### Add a new resource to the Provider

Allow the management of a logical resource within AWS by adding a new resource to the Terraform AWS Provider.

- [Name the resource](add-a-new-resource#add-a-new-resource)
- [Fork the Provider](add-a-new-resource#fork-the-provider)
- [Fill out the Resource Schema](add-a-new-resource#fill-out-the-resource-schema)
- [Implement CRUD Handlers](add-a-new-resource#implement-crud-handlers)
- [Write passing Acceptance Tests](add-a-new-resource#write-passing-acceptance-tests)
- [Create Documentation for the Resource](add-a-new-resource#create-documentation-for-the-resource)
- [Ensure format and lint checks are passing](add-a-new-resource#ensure-format-and-link-checks-are-passing)
- [Raise a Pull Request](raising-a-pull-request.md)
- [Wait for Prioritization](prioritization.md)

### Add a new datasource to the Provider

Allow reading an existing AWS resource by id or filter, and allow its use within Terraform configurations.


- [Name the resource](add-a-new-datasource#add-a-new-datasource)
- [Fork the Provider](add-a-new-datasource#fork-the-provider)
- [Fill out the Resource Schema](add-a-new-datasource#fill-out-the-datasource-schema)
- [Implement CRUD Handlers](add-a-new-datasource#implement-crud-handlers)
- [Write passing Acceptance Tests](add-a-new-datasource#write-passing-acceptance-tests)
- [Create Documentation for the Resource](add-a-new-resource#create-documentation-for-the-resource)
- [Ensure format and lint checks are passing](add-a-new-datasource#ensure-format-and-link-checks-are-passing)
- [Raise a Pull Request](raising-a-pull-request.md)
- [Wait for Prioritization](prioritization.md)

### Add a newly released AWS Region

New regions are immediately usable with the provider with the caveat that a configuration workaround is required to skip validation of the region during cli operations. A small set of changes are required to makes this workaround unecessary.

- [Add a new Region](add-a-new-region)


### Fix a bug, or make an enhancement to a resource or datasource

These constitute the majority of pull requests submitted, many of which we address and merge regardless of priority in our regular internal gardening days.

- [Making small changes to existing resources.](bugs-and-enhancements)

### Add Resource Name Generation Support

Allow a resource, to either fully, or partially generate its own resource names. TODO:This can be useful.

- [Resource Name Generation](resource-name-generation)

### Add Tagging support for an existing Resource

Many AWS resources allow [assigning metadata via tags](https://docs.aws.amazon.com/general/latest/gr/aws_tagging.html) which can be enormously useful to manage and track cloud resources from a governance and cost perspective. Often though, new aws services or features are launched without tagging support, which can follow later. In this case tagging support will need to be explicitly added to those resources. Typically this support follows an established pattern so support is normally straightforward to add.

- [Add tag support for a resource](resource-tagging)

### Add Import support for a Resource

Adding import support allows `terraform import` to be run targeting an existing unmanaged resource and pulling its configuration into Terraform state. Typically import support is added during initial resource implementation but in some cases this may not have been done.

- [Add import support](add-import-support.md)

### End User Documentation Changes

The provider documentation is displayed on the <a href="https://registry.terraform.io/providers/hashicorp/aws/latest">Terraform Registry and is sourced and refreshed from the provider repository during the release process.

- [How to modify provider documentation](documentation-changes)

### Report a bug, or make a feature request

In both cases follow the template laid out in the submission forms. Details are important, it can take a considerable amount of time for the maintainers to reproduce issues. So bug reports without full reproductions are likely to be uninvestigated, or in some case closed.

- [Report a Bug](https://github.com/hashicorp/terraform-provider-aws/issues/new?assignees=&labels=&template=Bug_Report.md)
- [Submit a Feature Request](https://github.com/hashicorp/terraform-provider-aws/issues/new?assignees=&labels=enhancement&template=Feature_Request.md)    
