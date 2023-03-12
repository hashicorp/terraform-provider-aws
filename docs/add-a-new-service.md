# Adding a New AWS Service

AWS frequently launches new services, and Terraform support is frequently desired by the community shortly after launch. Depending on the API surface area of the new service, this could be a major undertaking. The following steps should be followed to prepare for adding the resources that allow for Terraform management of that service.

## Perform Service Design

Before adding a new service to the provider its a good idea to familiarize yourself with the primary workflows practitioners are likely to want to accomplish with the provider to ensure the provider design can solve for this. Its not always necessary to cover 100% of the AWS service offering to unblock most workflows.

You should have an idea of what resources and data sources should be added, their dependencies and relative importance in relation to the workflow. This should give you an idea of the order in which resources to be added. It's important to note that generally, we like to review and merge resources in isolation, and avoid combining multiple new resources in one Pull Request.

Using the AWS API documentation as a reference, identify the various API's which correspond to the CRUD operations which consist of the management surface for that resource. These will be the set of API's called from the new resource. The API's model attributes will correspond to your resource schema.

From there begin to map out the list of resources you would like to implement, and note your plan on the GitHub issue relating to the service (or create one if one does not exist) for the community and maintainers to feedback.

## Add a Service Client

Before new resources are submitted, please raise a separate pull request containing just the new AWS SDK for Go service client.

To add an AWS SDK for Go service client:

1. Check the file `names/names_data.csv` for the service.
  If it is already there, you are ready to implement the first [resource](./add-a-new-resource.md) or [data source](./add-a-new-datasource.md).

1. Otherwise, determine the service identifier using the rule described in [the Naming Guide](naming.md#service-identifier).

1. In `names/names_data.csv`, add a new line with all the requested information for the service following the guidance in the [`names` README](https://github.com/hashicorp/terraform-provider-aws/blob/main/names/README.md).
  **_Be very careful when adding or changing data in `names_data.csv`!
  The Provider and generators depend on the file being correct.
  We strongly recommend using an editor with CSV support._**

1. Run the following then submit the pull request:

  ```sh
  make gen
  make test
  go mod tidy
  ```

Once the service client has been added, implement the first [resource](./add-a-new-resource.md) or [data source](./add-a-new-datasource.md) in a separate PR.
