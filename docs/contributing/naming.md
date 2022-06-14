# Naming Conventions for the AWS Provider

<!-- TOC depthFrom:2 -->

- [Service Identifier](#service-identifier)
- [Packages](#packages)
- [Resources and Data Sources](#resources-and-data-sources)
- [Files](#files)
- [MixedCaps](#mixedcaps)
- [Functions](#functions)
- [Variables and Constants](#variables-and-constants)
- [Acceptance and Unit Tests](#acceptance-and-unit-tests)
- [Test Support Functions](#test-support-functions)
- [Acceptance Test Configurations](#acceptance-test-configurations)

<!-- /TOC -->

## Service Identifier

In the AWS Provider, a service identifier should consistently identify an AWS service from code to documentation to provider use by a practitioner. Prominent places you will see service identifiers:

* The package name (e.g., `internal/service/<serviceidentifier>`)
* In resource and data source names (e.g., `aws_<serviceidentifier>_thing`)
* Documentation file names (e.g., `website/docs/r/<serviceidentifier>_thing`)

Typically, choosing the AWS Provider identifier for a service is simple. AWS consistently uses one name and we use that name as the identifier. However, some services are not simple. To provide consistency, and to help contributors and practitioners know what to expect, we provide this rule for defining a service identifier:

### Rule

1. Determine the service package name for [AWS Go SDK v2](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2).
2. Determine the [AWS CLI v2](https://awscli.amazonaws.com/v2/documentation/api/latest/index.html) _command_ corresponding to the service (i.e., the word following `aws` in CLI commands; e.g., for `aws sts get-caller-identity`, `sts` is the _command_, `get-caller-identity` is the _subcommand_).
3. If the SDK and CLI agree, use that. If the service only exists in one, use that.
4. If they differ, use the shorter of the two.
5. Use lowercase letters and do not include any underscores (`_`).

### How Well Is It Followed?

With 156+ services having some level of implementation, the following is a summary of how well this rule is currently followed.

For AWS provider service package names, only five packages violate this rule: `appautoscaling` should be `applicationautoscaling`, `codedeploy` should be `deploy`, `elasticsearch` should be `es`, `cloudwatchlogs` should be `logs`, and `simpledb` should be `sdb`.

For the service identifiers used in resource and data source configuration names (e.g., `aws_acmpca_certificate_authority`), 32 wholly or partially violate the rule.

* EC2, ELB, ELBv2, and RDS have legacy but heavily used resources and data sources that do not or inconsistently use service identifiers.
* The remaining 28 services violate the rule in a consistent way: `appautoscaling` should be `applicationautoscaling`, `codedeploy` should be `deploy`, `elasticsearch` should be `es`, `cloudwatch_log` should be `logs`, `simpledb` should be `sdb`, `prometheus` should be `amp`, `api_gateway` should be `apigateway`, `cloudcontrolapi` should be `cloudcontrol`, `cognito_identity` should be `cognitoidentity`, `cognito` should be `cognitoidp`, `config` should be `configservice`, `dx` should be `directconnect`, `directory_service` should be `ds`, `elastic_beanstalk` should be `elasticbeanstalk`, `cloudwatch_event` should be `events`, `kinesis_firehose` should be `firehose`, `msk` should be `kafka`, `mskconnect` should be `kafkaconnect`, `kinesis_analytics` should be `kinesisanalytics`, `kinesis_video` should be `kinesisvideo`, `lex` should be `lexmodels`, `media_convert` should be `mediaconvert`, `media_package` should be `mediapackage`, `media_store` should be `mediastore`, `route53_resolver` should be `route53resolver`, relevant `s3` should be `s3control`, `serverlessapplicationrepository` should be `serverlessrepo`, and `service_discovery` should be `servicediscovery`.

## Packages

Package names are not seen or used by practitioners. However, they should still be carefully considered.

### Rule

1. For service packages (i.e., packages under `internal/service`), use the AWS Provider [service identifier](#service-identifier) as the package name.
2. For other packages, use a short name for the package. Common Go lengths are 3-9 characters.
3. Use a descriptive name. The name should capture the key purpose of the package.
4. Use lowercase letters and do not include any underscores (`_`).
5. Avoid useless names like `helper`. These names convey zero information. Everything in the AWS Provider is helping something or someone do something so the name `helper` doesn't narrow down the purpose of the package within the codebase.
6. Use a name that is not too narrow or too broad as Go packages should not be too big or too small. Tiny packages can be combined using a broader name encompassing both. For example, `verify` is a good name because it tells you _what_ the package does and allows a broad set of validation, comparison, and checking functionality.

## Resources and Data Sources

When creating a new resource or data source, it is important to get names right. Once practitioners rely on names, we can only change them through breaking changes. If you are unsure about what to call a resource or data source, discuss it with the community and maintainers.

### Rule

1. Follow the [AWS SDK for Go v2](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2). Almost always, the API operations make determining the name simple. For example, the [Amazon CloudWatch Evidently](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/evidently) service includes `CreateExperiment`, `GetExperiment`, `UpdateExperiment`, and `DeleteExperiment`. Thus, the resource (or data source) name is "Experiment."
2. Give a resource its Terraform configuration (i.e., HCL) name (e.g., `aws_imagebuilder_image_pipeline`) by joining these three parts with underscores:
    * `aws` prefix
    * [Service identifier](#service-identifier) (service identifiers do not include underscores), all lower case (e.g., `imagebuilder`)
    * Resource (or data source) name in snake case (spaces replaced with underscores, if any), all lower case (e.g., `image_pipeline`)
3. Name the main resource function `Resource<ResourceName>()`, with the resource name in [MixedCaps](#mixedcaps). Do not include the service name or identifier. For example, define `ResourceImagePipeline()` in a file called `internal/service/imagebuilder/image_pipeline.go`.
4. Similarly, name the main data source function `DataSource<ResourceName>()`, with the data source name in [MixedCaps](#mixedcaps). Do not include the service name or identifier. For example, define `DataSourceImagePipeline()` in a file called `internal/service/imagebuilder/image_pipeline_data_source.go`.

## Files

File names should follow Go and Markdown conventions with these additional points.

### Resource and Data Source Documentation Rule

1. Resource markdown goes in the `website/docs/r` directory. Data source markdown goes in the `website/docs/d` directory.
2. Use the [service identifier](#service-identifier) and resource or data source name, separated by an underscore (`_`).
3. All letters are lowercase.
4. Use `.html.markdown` as the extension.
5. Do not include "aws" in the name.

A correct example is `accessanalyzer_analyzer.html.markdown`. An incorrect example is `service_discovery_instance.html.markdown` because the [service identifier](#service-identifier) should not include an underscore.

### Go File Rule

1. Resource and data source files are in the `internal/service/<service>` directory.
2. Do not include the service as part of the file name.
3. Data sources should include `_data_source` after the data source name (e.g., `application_data_source.go`).
4. Put unit and acceptance tests in a file ending with `_test.go` (e.g., `custom_domain_association_test.go`).
5. Use snake case for multiword names (i.e., all letters are lowercase, words separated by underscores).
6. Use the `.go` extension.
7. Idiomatic names for common non-resource, non-data-source files include `consts.go` (service-wide constants), `find.go` (finders), `flex.go` (FLatteners and EXpanders), `generate.go` (directives for code generation), `id.go` (ID creators and parsers), `status.go` (status functions), `sweep.go` (sweepers), `tags_gen.go` (generated tag code), `validate.go` (validators), and `wait.go` (waiters).

## MixedCaps

Write multiword names in Go using _MixedCaps_ (or _mixedCaps_) rather than underscores.

For more details on capitalizations we enforce with CI Semgrep tests, see the [Caps List](../../names/caps.md).

Initialisms and other abbreviations are a key difference between many camel/Pascal case interpretations and mixedCaps. **Abbreviations in mixedCaps should be the correct, human-readable case.** After all, names in code _are for humans_. (The mixedCaps convention aligns with HashiCorp's emphasis on pragmatism and beauty.)

For example, an initialism such as "VPC" should either be all capitalized ("VPC") or all lower case ("vpc"), never "Vpc" or "vPC." Similarly, in mixedCaps, "DynamoDB" should either be "DynamoDB" or "dynamoDB", depending on whether an initial cap is needed or not, and never "dynamoDb" or "DynamoDb."

### Rule

1. Use _mixedCaps_ for function, type, method, variable, and constant names in the Terraform AWS Provider Go code.

## Functions

In general, follow Go best practices for good function naming. This rule is for functions defined outside of the _test_ context (i.e., not in a file ending with `_test.go`). For test functions, see [Test Support Functions](#test-support-functions) or [Acceptance Test Configurations](#acceptance-test-configurations) below.

### Rule

1. Only export functions (capitalize) when necessary, i.e., when the function is used outside the current package, including in the `_test` (`.test`) package.
2. Use [MixedCaps](#mixedcaps) (exported) or [mixedCaps](#mixedcaps) (not exported). Do not use underscores for multiwords.
3. Do not include the service name in the function name. (If functions are used outside the current package, the import package clarifies a function's origin. For example, the EC2 function `FindVPCEndpointByID()` is used outside the `internal/service/ec2` package but where it is used, the call is `tfec2.FindVPCEndpointByID()`.)
4. For CRUD functions for resources, use this format: `resource<ResourceName><CRUDFunction>`. For example, `resourceImageRecipeUpdate()`, `resourceBaiduChannelRead()`.
5. For data sources, for Read functions, use this format: `dataSource<DataSourceName>Read`. For example, `dataSourceBrokerRead()`, `dataSourceEngineVersionRead()`.
6. To improve readability, consider including the resource name in helper function names that pertain only to that resource. For example, for an expander function for an "App" resource and a "Campaign Hook" expander, use `expandAppCampaignHook()`.
7. Do not include "AWS" or "Aws" in the name.

## Variables and Constants

In general, follow Go best practices for good variable and constant naming.

### Rule

1. Only export variables and constants (capitalize) when necessary, i.e., the variable or constant is used outside the current package, including in the `_test` (`.test`) package.
2. Use [MixedCaps](#mixedcaps) (exported) or [mixedCaps](#mixedcaps) (not exported). Do not use underscores for multiwords.
3. Do not include the service name in variable or constant names. (If variables or constants are used outside the current package, the import package clarifies its origin. For example, IAM's `PropagationTimeout` is widely used outside of IAM but each instance is through the package import alias, `tfiam.PropagationTimeout`. "IAM" is unnecessary in the constant name.)
4. To improve readability, consider including the resource name in variable and constant names that pertain only to that resource. For example, for a string constant for a "Role" resource and a "not found" status, use `roleStatusNotFound` or `RoleStatusNotFound`, if used outside the service's package.
5. Do not include "AWS" or "Aws" in the name.

**NOTE:** Give priority to using constants from the AWS SDK for Go rather than defining new constants for the same values.

## Acceptance and Unit Tests

With about 6000 acceptance and unit tests, following these naming conventions is essential to organization and (human) context switching between services.

There are three types of tests in the AWS Provider: (regular) acceptance tests, serialized acceptance tests, and unit tests. All are functions that take a variable of type `*testing.T`. Acceptance tests and unit tests have exported (i.e., capitalized) names while serialized tests do not. Serialized tests are called by another exported acceptance test, often ending with `_serial`. The majority of tests in the AWS provider are acceptance tests.

### Acceptance Test Rule

Acceptance test names have a minimum of two (e.g., `TestAccBackupPlan_tags`) or a maximum of three (e.g., `TestAccDynamoDBTable_Replica_multiple`) parts, joined with underscores:

1. First part: All have a _prefix_ (i.e., `TestAcc`), _service name_ (e.g., `Backup`, `DynamoDB`), and _resource name_ (e.g., `Plan`, `Table`), [MixedCaps](#mixedcaps) without underscores between. Do not include "AWS" or "Aws" in the name.
2. Middle part (Optional): _Test group_ (e.g., `Replica`), uppercase, [MixedCaps](#mixedcaps). Consider a metaphor where tests are chapters in a book. If it is helpful, tests can be grouped together like chapters in a book that are sometimes grouped into parts or sections of the book.
3. Last part: _Test identifier_ (e.g., `basic`, `tags`, or `multiple`), lowercase, [mixedCaps](#mixedcaps)). The identifier should make the test's purpose clear but be concise. For example, the identifier `conflictsWithCloudFrontDefaultCertificate` (41 characters) conveys no more information than `conflictDefaultCertificate` (26 characters), since "CloudFront" is implied and "with" is _always_ implicit. Avoid words that convey no meaning or whose meaning is implied. For example, "with" (e.g., `_withTags`) is not needed because we imply the name is telling us what the test is _with_. `withTags` can be simplified to `tags`.

### Serialized Acceptance Test Rule

The names of serialized acceptance tests follow the regular [acceptance test name rule](#acceptance-test-rule) **_except_** serialized acceptance test names:

1. Start with `testAcc` instead of `TestAcc`
2. Do not include the name of the service (e.g., a serialized acceptance test would be called `testAccApp_basic` not `testAccAmplifyApp_basic`).

### Unit Test Rule

Unit test names follow the same [rule](#acceptance-test-rule) as acceptance test names except unit test names:

1. Start with `Test`, not `TestAcc`
2. Do not include the name of the service
3. Usually do not have any underscores
4. If they test a function, should include the function name (e.g., a unit test of `ExpandListener()` should be called `TestExpandListener()`)

## Test Support Functions

This rule is for functions defined in the _test_ context (i.e., in a file ending with `_test.go`) that do not return a string with Terraform configuration. For non-test functions, see [Functions](#functions) above. Or, see [Acceptance Test Configurations](#acceptance-test-configurations) below.

### Rule

1. Only export functions (capitalize) when necessary, i.e., when the function is used outside the current package. _This is very rare._
2. Use [MixedCaps](#mixedcaps) (exported) or [mixedCaps](#mixedcaps) (not exported). Do not use underscores for multiwords.
3. Do not include the service name in the function name. For example, `testAccCheckAMPWorkspaceExists()` should be named `testAccCheckWorkspaceExists()` instead, dropping the service name.
4. Several types of support functions occur commonly and should follow these patterns:
    * Destroy: `testAccCheck<Resource>Destroy`
    * Disappears: `testAccCheck<Resource>Disappears`
    * Exists: `testAccCheck<Resource>Exists`
    * Not Recreated: `testAccCheck<Resource>NotRecreated`
    * PreCheck: `testAccPreCheck` (often, only one PreCheck is needed per service so no resource name is needed)
    * Recreated: `testAccCheck<Resource>Recreated`
5. Do not include "AWS" or "Aws" in the name.

## Acceptance Test Configurations

This rule is for functions defined in the _test_ context (i.e., in a file ending with `_test.go`) that return a string with Terraform configuration. For test support functions, see [Test Support Functions](#test-support-functions) above. Or, for non-test functions, see [Functions](#functions) above.

**NOTE:** This rule is not widely used currently. However, new functions and functions you change should follow it.

### Rule

1. Only export functions (capitalize) when necessary, i.e., when the function is used outside the current package. _This is very rare._
2. Use [MixedCaps](#mixedcaps) (exported) or [mixedCaps](#mixedcaps) (not exported). Do not use underscores for multiwords.
3. Do not include the service name in the function name.
4. Follow this pattern: `testAccConfig<Resource>_<TestGroup>_<configDescription>`
    * `_<TestGroup>` is optional. Refer to the [Acceptance Test Rule](#acceptance-test-rule) test group discussion.
    * Especially when an acceptance test only uses one configuration, the `<configDescription>` should be the same as the test identifier discussed in the [Acceptance Test Rule](#acceptance-test-rule).
5. Do not include "AWS" or "Aws" in the name.
