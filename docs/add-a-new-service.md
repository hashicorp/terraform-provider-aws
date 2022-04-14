## Adding a New AWS Service to the Provider

AWS frequently launches new services, and Terraform support is frequently desired by the community shortly after launch. Depending on the API surface area of the new service, this could be a major undertaking. The following steps should be followed to prepare for adding the resources that allow for Terraform management of that service.

### Add a Service Client


TODO: This will be generated

Before new resources are submitted, please raise a separate pull request containing just the new AWS Go SDK service client. Doing so will pull the AWS Go SDK service code into the project at the current version. Since the AWS Go SDK is updated frequently, these pull reqests can easily have merge conflicts or be out of date. The maintainers prioritize reviewing and merging these quickly to prevent those situations.

To add the AWS Go SDK service client:
- In `internal/conns/conns.go`: Add a string constant for the service. Follow these rules to name the constant.
    - The constant name should be the same as the service name used in the AWS Go SDK except:
        1. Drop "service" or "api" if the service name ends with either or both, and
        2. Shorten the service name if it is excessively long. Avoid names longer than 17 characters if possible. When shortening a service name, look to the endpoints ID, common usage in documentation and marketing, and discuss the change with the community and maintainers to get buy in. The goals for this alternate name are to be instantly recognizable, not verbose, and more easily managed.
    - The constant name should be capitalized following Go mixed-case rules. In other words:
        1. Do not use underscores,
        2. The first letter of each word is capitalized, and
        3. Abbreviations and initialisms are all caps.
    - Proper examples include `CognitoIdentity`, `DevOpsGuru`, `DynamoDB`, `ECS`, `Prometheus` ("Service" is dropped from end), and `ServerlessRepo` (shortened from "Serverless Application Repository").
    - The constant value is the same as the name but all lowercase (_e.g._, `DynamoDB = "dynamodb"`).
- In `internal/conns/conns.go`: Add a new entry to the `serviceData` map:
    1. The entry key is the string constant created above
    2. The `AWSClientName` is the exact name of the return type of the `New()` method of the service. For example, see the `New()` method in the [Application Auto Scaling documentation](https://docs.aws.amazon.com/sdk-for-go/api/service/applicationautoscaling/#New).
    3. For `AWSServiceName`, `AWSEndpointsID`, and `AWSServiceID`, directly reference the AWS Go SDK service package for the values. For example, `accessanalyzer.ServiceName`, `accessanalyzer.EndpointsID`, and `accessanalyzer.ServiceID` respectively.
    4. `ProviderNameUpper` is the exact same as the constant _name_ (_not_ value) as described above.
    5. In most cases, the `HCLKeys` slice will have one element, an all-lowercase string that matches the AWS SDK Go service name and provider constant value, described above. However, when these diverge, it may be helpful to add additional elements. Practitioners can use any of these names in the provider configuration when customizing service endpoints.
- In `internal/conns/conns.go`: Add a new import for the AWS Go SDK code. E.g.
`github.com/aws/aws-sdk-go/service/quicksight`
- In `internal/conns/conns.go`: Add a new `{ServiceName}Conn` field to the `AWSClient`
struct for the service client. The service name should match the constant name, capitalized the same, as described above.
_E.g._, `DynamoDBConn *dynamodb.DynamoDB`.
- In `internal/conns/conns.go`: Create the new service client in the `{ServiceName}Conn`
field in the `AWSClient` instantiation within `Client()`, using the constant created above as a key to a value in the `Endpoints` map. _E.g._,
`DynamoDBConn: dynamodb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints[DynamoDB])})),`.
- In `website/allowed-subcategories.txt`: Add a name acceptable for the documentation navigation.
- In `website/docs/guides/custom-service-endpoints.html.md`: Add the service
name in the list of customizable endpoints.
- In `infrastructure/repository/labels-service.tf`: Add the new service to create a repository label.
- In `.github/labeler-issue-triage.yml`: Add the new service to automated issue labeling. E.g., with the `quicksight` service
  ``yaml
  #... other services ...
  srvice/quicksight:
- '((\*|-) ?`?|(data|resource) "?)aws_quicksight_'
  #... other services ...
  ``

- In `.github/labeler-pr-triage.yml`: Add the new service to automated pull request labeling. E.g., with the `quicksight` service

    ```yaml
    # ... other services ...
    service/quicksight:
        - 'internal/service/quicksight/**/*'
        - '**/*_quicksight_*'
        - '**/quicksight_*'
    # ... other services ...
    ```

- Run the following then submit the pull request:

  ```sh
  make test
  go mod tidy
  ```

### Perform Service Design

Before adding a new service to the provider its a good idea to familarize yourself with the primary workflows practioners are likely to want to accomplish with the provider to ensure the provider design can solve for for this. Its not always necessary to cover 100% of the AWS service offering to unblock most workflows.

You should have an idea of what resources and datasources should be added, their dependencies and relative imprortance in relation to the workflow. This should give you an idea of the order in which resources to be added. It's important to note that generally, we like to review and merge resources in isolation, and avoid combining multiple new resources in one Pull Request.

Using the AWS API documentation as a reference, identify the various API's which correspond to the CRUD operations which consist of the management surface for that resource. These will be the set of API's called from the new resource. The API's model attributes will correspond to your resource schema.

From there begin to map out the list of resources you would like to implement, and note your plan on the GitHub issue relating to the service (or create one if one does not exist) for the community and maintainers to feedback.

From there you are ready to [create your first resource](add-a-new-resource.md)!