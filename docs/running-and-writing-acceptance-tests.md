# Running and Writing Acceptance Tests

Terraform includes an acceptance test harness that does most of the repetitive
work involved in testing a resource. For additional information about testing
Terraform Providers, see the [SDKv2 documentation](https://www.terraform.io/plugin/sdkv2/testing).

## In context

To help place acceptance testing in context, here is an overview of the Terraform AWS Provider's three types of tests.

1. **Acceptance tests** (_You are here!_) are end-to-end evaluations of interactions with AWS. They validate functionalities like creating, reading, and destroying resources within AWS.
2. [**Unit tests**](unit-tests.md) focus on testing isolated units of code within the software, typically at the function level. They assess functionalities solely within the provider itself.
3. [**Continuous integration tests**](continuous-integration.md) encompass a suite of automated tests that are executed on every pull request and include linting, compiling code, running unit tests, and performing static analysis.

## Acceptance Tests Often Cost Money to Run

Our acceptance test suite creates real resources, and as a result, they cost real money to run.
Because the resources only exist for a short period of time, the total amount
of money required is usually relatively small. That said there are particular services
which are very expensive to run and it's important to be prepared for those costs.

Some services which can be cost-prohibitive include (among others):

- ACM (Amazon Certificate Manager)
- Bedrock
- EC2
- ElastiCache
- FSx
- Glue
- Kinesis Analytics
- OpenSearch
- RDS
- Storage Gateway
- WorkSpaces

We don't want financial limitations to be a barrier to contribution, so if you are unable to
pay to run acceptance tests for your contribution, mention this in your
pull request. We will happily accept "best effort" implementations of
acceptance tests and run them for you on our side. This might mean that your PR
takes a bit longer to merge, but it most definitely is not a blocker for
contributions.

## Running an Acceptance Test

Acceptance tests can be run using the `testacc` target in the Terraform
`Makefile`. The individual tests to run can be controlled using a regular
expression. Prior to running the tests provider configuration details such as
access keys must be made available as environment variables.

For example, to run an acceptance test against the Amazon Web Services
provider, the following environment variables must be set:

```sh
# Using a profile
export AWS_PROFILE=...
# Otherwise
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_DEFAULT_REGION=...
```

Please note that the default region for the testing is `us-west-2` and must be
overridden via the `AWS_DEFAULT_REGION` environment variable, if necessary. This
is especially important for testing AWS GovCloud (US), which requires:

```sh
export AWS_DEFAULT_REGION=us-gov-west-1
```

Tests can then be run by specifying a regular expression defining the tests to
run and the package in which the tests are defined:

```console
make testacc TESTS=TestAccCloudWatchDashboard_updateName PKG=cloudwatch
```

```
==> Checking that code complies with gofmt requirements...
TF_ACC=1 go test ./internal/service/cloudwatch/... -v -count 1 -parallel 20 -run=TestAccCloudWatchDashboard_updateName -timeout 180m
=== RUN   TestAccCloudWatchDashboard_updateName
=== PAUSE TestAccCloudWatchDashboard_updateName
=== CONT  TestAccCloudWatchDashboard_updateName
--- PASS: TestAccCloudWatchDashboard_updateName (25.33s)
PASS
ok  	github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch	25.387s
```

Entire resource test suites can be targeted by using the naming convention to
write the regular expression. For example, to run all tests of the
`aws_cloudwatch_dashboard` resource rather than just the `updateName` test, you
can start testing like this:

```console
make testacc TESTS=TestAccCloudWatchDashboard PKG=cloudwatch
```

```
==> Checking that code complies with gofmt requirements...
TF_ACC=1 go test ./internal/service/cloudwatch/... -v -count 1 -parallel 20 -run=TestAccCloudWatchDashboard -timeout 180m
=== RUN   TestAccCloudWatchDashboard_basic
=== PAUSE TestAccCloudWatchDashboard_basic
=== RUN   TestAccCloudWatchDashboard_update
=== PAUSE TestAccCloudWatchDashboard_update
=== RUN   TestAccCloudWatchDashboard_updateName
=== PAUSE TestAccCloudWatchDashboard_updateName
=== CONT  TestAccCloudWatchDashboard_basic
=== CONT  TestAccCloudWatchDashboard_updateName
=== CONT  TestAccCloudWatchDashboard_update
--- PASS: TestAccCloudWatchDashboard_basic (15.83s)
--- PASS: TestAccCloudWatchDashboard_updateName (26.69s)
--- PASS: TestAccCloudWatchDashboard_update (27.72s)
PASS
ok  	github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch	27.783s
```

Running acceptance tests requires version 0.12.26 or higher of the Terraform CLI to be installed.

For advanced developers, the acceptance testing framework accepts some additional environment variables that can be used to control Terraform CLI binary selection, logging, and other behaviors. See the [SDKv2 documentation](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests#environment-variables) for more information.

Please Note: On macOS 10.14 and later (and some Linux distributions), the default user open file limit is 256. This may cause unexpected issues when running the acceptance testing since this can prevent various operations from occurring such as opening network connections to AWS. To view this limit, the `ulimit -n` command can be run. To update this limit, run `ulimit -n 1024`  (or higher).

### Running Cross-Account Tests

Certain testing requires multiple AWS accounts. This additional setup is not typically required and the testing will return an error (shown below) if your current setup does not have the secondary AWS configuration:

```console
make testacc TESTS=TestAccRDSInstance_DBSubnetGroupName_ramShared PKG=rds
```

```
TF_ACC=1 go test ./internal/service/rds/... -v -count 1 -parallel 20 -run=TestAccRDSInstance_DBSubnetGroupName_ramShared -timeout 180m
=== RUN   TestAccRDSInstance_DBSubnetGroupName_ramShared
=== PAUSE TestAccRDSInstance_DBSubnetGroupName_ramShared
=== CONT  TestAccRDSInstance_DBSubnetGroupName_ramShared
    acctest.go:674: skipping test because at least one environment variable of [AWS_ALTERNATE_PROFILE AWS_ALTERNATE_ACCESS_KEY_ID] must be set. Usage: credentials for running acceptance testing in alternate AWS account.
--- SKIP: TestAccRDSInstance_DBSubnetGroupName_ramShared (0.85s)
PASS
ok      github.com/hashicorp/terraform-provider-aws/internal/service/rds        0.888s
```

Running these acceptance tests is the same as before, except the following additional AWS credential information is required:

```sh
# Using a profile
export AWS_ALTERNATE_PROFILE=...
# Otherwise
export AWS_ALTERNATE_ACCESS_KEY_ID=...
export AWS_ALTERNATE_SECRET_ACCESS_KEY=...
```

### Running Cross-Region Tests

Certain testing requires multiple AWS regions. Additional setup is not typically required because the testing defaults the second AWS region to `us-east-1` and the third AWS region to `us-east-2`.

Running these acceptance tests is the same as before, but if you wish to override the second and third regions:

```sh
export AWS_ALTERNATE_REGION=...
export AWS_THIRD_REGION=...
```

### Running Only Short Tests

Some tests have been manually marked as long-running (longer than 300 seconds) and can be skipped using the `-short` flag. However, we are adding long-running guards little by little and many services have no guarded tests.

Where guards have been implemented, do not always skip long-running tests. However, for intermediate test runs during development, or to verify functionality unrelated to the specific long-running tests, skipping long-running tests makes work more efficient. We recommend that for the final test run before submitting a PR you run affected tests without the `-short` flag.

If you want to run only short-running tests, you can use either one of these equivalent statements. Note the use of `-short`.

For example:

```console
make testacc TESTS='TestAccECSTaskDefinition_' PKG=ecs TESTARGS=-short
```

Or:

```console
TF_ACC=1 go test ./internal/service/ecs/... -v -count 1 -parallel 20 -run='TestAccECSTaskDefinition_' -short -timeout 180m
```

## Writing an Acceptance Test

Terraform has a framework for writing acceptance tests which minimizes the
amount of boilerplate code necessary to use common testing patterns. This guide is meant to augment the general [SDKv2 documentation](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests) with Terraform AWS Provider specific conventions and helpers.

### Anatomy of an Acceptance Test

This section describes in detail how the Terraform acceptance testing framework operates with respect to the Terraform AWS Provider. We recommend those unfamiliar with this provider, or Terraform resource testing in general, take a look here first to generally understand how we interact with AWS and the resource code to verify functionality.

The entry point to the framework is the `resource.ParallelTest()` function. This wraps our testing to work with the standard Go testing framework, while also preventing unexpected usage of AWS by requiring the `TF_ACC=1` environment variable. This function accepts a `TestCase` parameter, which has all the details about the test itself. For example, this includes the test steps (`TestSteps`) and how to verify resource deletion in the API after all steps have been run (`CheckDestroy`).

Each `TestStep` proceeds by applying some
Terraform configuration using the provider under test, and then verifying that
results are as expected by making assertions using the provider API. It is
common for a single test function to exercise both the creation of and updates
to a single resource. Most tests follow a similar structure.

1. Pre-flight checks are made to ensure that sufficient provider configuration
   is available to be able to proceed - for example in an acceptance test
   targeting AWS, `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` must be set prior
   to running acceptance tests. This is common to all tests exercising a single
   provider.

Most assertion
functions are defined out of band with the tests. This keeps the tests
readable, and allows reuse of assertion functions across different tests of the
same type of resource. The definition of a complete test looks like this:

```go
func TestAccCloudWatchDashboard_basic(t *testing.T) {
  ctx := acctest.Context(t)
	var dashboard cloudwatch.GetDashboardOutput
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDashboardDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(ctx, "aws_cloudwatch_dashboard.foobar", &dashboard),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foobar", "dashboard_name", testAccDashboardName(rInt)),
				),
			},
		},
	})
}
```

When executing the test, the following steps are taken for each `TestStep`:

1. The Terraform configuration required for the test is applied. This is
   responsible for configuring the resource under test, and any dependencies it
   may have. For example, to test the `aws_cloudwatch_dashboard` resource, a valid configuration with the requisite fields is required. This results in a configuration which looks like this:

    ```terraform
    resource "aws_cloudwatch_dashboard" "foobar" {
      dashboard_name = "terraform-test-dashboard-%[1]d"
      dashboard_body = <<EOF
      {
        "widgets": [{
          "type": "text",
          "x": 0,
          "y": 0,
          "width": 6,
          "height": 6,
          "properties": {
            "markdown": "Hi there from Terraform: CloudWatch"
          }
        }]
      }
      EOF
    }
    ```

1. Assertions are run using the provider API. These use the provider API
   directly rather than asserting against the resource state. For example, to
   verify that the `aws_cloudwatch_dashboard` described above was created
   successfully, a test function like this is used:

    ```go
    func testAccCheckDashboardExists(ctx context.Context, n string, dashboard *cloudwatch.GetDashboardOutput) resource.TestCheckFunc {
      return func(s *terraform.State) error {
        rs, ok := s.RootModule().Resources[n]
        if !ok {
          return fmt.Errorf("Not found: %s", n)
        }

        conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn(ctx)
        params := cloudwatch.GetDashboardInput{
          DashboardName: aws.String(rs.Primary.ID),
        }

        resp, err := conn.GetDashboardWithContext(ctx, &params)
        if err != nil {
          return err
        }

        *dashboard = *resp

        return nil
      }
    }
    ```

   Notice that the only information used from the Terraform state is the ID of
   the resource. For computed properties, we instead assert that the value saved in the Terraform state was the
   expected value if possible. The testing framework provides helper functions
   for several common types of checks - for example:

   ```go
   resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foobar", "dashboard_name", testAccDashboardName(rInt)),
   ```

1. The resources created by the test are destroyed. This step happens
   automatically, and is the equivalent of calling `terraform destroy`.

1. Assertions are made against the provider API to verify that the resources
   have indeed been removed. If these checks fail, the test fails and reports
   "dangling resources". The code to ensure that the `aws_cloudwatch_dashboard` shown
   above has been destroyed looks like this:

    ```go
    func testAccCheckDashboardDestroy(ctx context.Context) resource.TestCheckFunc {
	    return func(s *terraform.State) error {
        conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn(ctx)

        for _, rs := range s.RootModule().Resources {
          if rs.Type != "aws_cloudwatch_dashboard" {
            continue
          }

          params := cloudwatch.GetDashboardInput{
            DashboardName: aws.String(rs.Primary.ID),
          }

          _, err := conn.GetDashboardWithContext(ctx, &params)
          if err == nil {
            return fmt.Errorf("Dashboard still exists: %s", rs.Primary.ID)
          }
          if !isDashboardNotFoundErr(err) {
            return err
          }
        }

        return nil
      }
    }
    ```

   These functions usually test only for the resource directly under test.

### Resource Acceptance Testing

Most resources that implement standard Create, Read, Update, and Delete functionality should follow the pattern below. Each test type has a section that describes them in more detail:

- **basic**: This represents the bare minimum verification that the resource can be created, read, deleted, and optionally imported.
- **disappears**: A test that verifies Terraform will offer to recreate a resource if it is deleted outside of Terraform (e.g., via the Console) instead of returning an error that it cannot be found.
- **Per Attribute**: A test that verifies the resource with a single additional argument can be created, read, optionally updated (or force resource recreation), deleted, and optionally imported.

The leading sections below highlight additional recommended patterns.

#### Test Configurations

Most of the existing test configurations you will find in the Terraform AWS Provider are written in the following function-based style:

```go
func TestAccExampleThing_basic(t *testing.T) {
  // ... omitted for brevity ...

  resource.ParallelTest(t, resource.TestCase{
    // ... omitted for brevity ...
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingConfig(),
        // ... omitted for brevity ...
      },
    },
  })
}

func testAccExampleThingConfig() string {
  return `
resource "aws_example_thing" "test" {
  # ... omitted for brevity ...
}
`
}
```

Even when no values need to be passed in to the test configuration, we have found this setup to be the most flexible for allowing that to be easily implemented. Any configurable values are handled via `fmt.Sprintf()`. Using `text/template` or other templating styles is explicitly forbidden.

For consistency, resources in the test configuration should be named `resource "..." "test"` unless multiple of that resource are necessary.

We discourage re-using test configurations across test files (except for some common configuration helpers we provide) as it is much harder to discover potential testing regressions.

Please also note that the newline on the first line of the configuration (before `resource`) and the newline after the last line of configuration (after `}`) are important to allow test configurations to be easily combined without generating Terraform configuration language syntax errors.

#### Test Configuration Independence

_Across the entire provider_, all test configurations should be as independent of each other as possible. For example, a common place this concept comes up is with the default VPC. Since we have tests that reconfigure the default VPC, if your configuration requires a VPC, it should not rely on the default VPC. Instead, include a VPC that will be created and destroyed as part of the test.

Make sure that your test configuration:

1. Includes everything required for Terraform to run the test
1. Does not assume that any user-managed infrastructure will be in place, such as S3 buckets, IAM roles, KMS keys, VPCs, subnets, etc.

#### Combining Test Configurations

We include a helper function, `acctest.ConfigCompose()` for iteratively building and chaining test configurations together. It accepts any number of configurations to combine them. This simplifies a single resource's testing by allowing the creation of a "base" test configuration for all the other test configurations (if necessary) and also allows the maintainers to curate common configurations. Each of these is described in more detail in the below sections.

Please note that we do discourage _excessive_ chaining of configurations such as implementing multiple layers of "base" configurations. Usually these configurations are harder for maintainers and other future readers to understand due to the multiple levels of indirection.

##### Base Test Configurations

If a resource requires the same Terraform configuration as a prerequisite for all test configurations, then a common pattern is implementing a "base" test configuration that is combined with each test configuration.

For example:

```go
func testAccExampleThingConfigBase() string {
  return `
resource "aws_iam_role" "test" {
  # ... omitted for brevity ...
}

resource "aws_iam_role_policy" "test" {
  # ... omitted for brevity ...
}
`
}

func testAccExampleThingConfig() string {
  return acctest.ConfigCompose(
    testAccExampleThingConfigBase(),
    `
resource "aws_example_thing" "test" {
  # ... omitted for brevity ...
}
`)
}
```

##### Available Common Test Configurations

These test configurations are typical implementations we have found or allow testing to implement best practices easier, since the Terraform AWS Provider testing is expected to run against various AWS Regions and Partitions.

- `acctest.AvailableEC2InstanceTypeForRegion("type1", "type2", ...)`: Typically used to replace hardcoded EC2 Instance Types. Uses `aws_ec2_instance_type_offering` data source to return an available EC2 Instance Type in preferred ordering. Reference the instance type via: `data.aws_ec2_instance_type_offering.available.instance_type`. Use `acctest.AvailableEC2InstanceTypeForRegionNamed("name", "type1", "type2", ...)` to specify a name for the data source
- `acctest.ConfigLatestAmazonLinuxHVMEBSAMI()`: Typically used to replace hardcoded EC2 Image IDs (`ami-12345678`). Uses `aws_ami` data source to find the latest Amazon Linux image. Reference the AMI ID via: `data.aws_ami.amzn-ami-minimal-hvm-ebs.id`

#### Randomized Naming

For AWS resources that require unique naming, the tests should implement a randomized name, typically coded as a `rName` variable in the test and passed as a parameter to create the test configuration.

For example:

```go
func TestAccExampleThing_basic(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  // ... omitted for brevity ...

  resource.ParallelTest(t, resource.TestCase{
    // ... omitted for brevity ...
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingConfigName(rName),
        // ... omitted for brevity ...
      },
    },
  })
}

func testAccExampleThingConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q
}
`, rName)
}
```

Typically the `rName` is always the first argument to the test configuration function, if used, for consistency.

Note that if `rName` (or any other variable) is used multiple times in the `fmt.Sprintf()` statement, _do not_ repeat `rName` in the `fmt.Sprintf()` arguments. Using `fmt.Sprintf(..., rName, rName)`, for example, would not be correct. Instead, use the indexed `%[1]q` (or `%[x]q`, `%[x]s`, `%[x]t`, or `%[x]d`, where `x` represents the index number) verb multiple times. For example:

```go
func testAccExampleThingConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
```

#### Other Recommended Variables

We also typically recommend saving a `resourceName` variable in the test that contains the resource reference, e.g., `aws_example_thing.test`, which is repeatedly used in the checks.

For example:

```go
func TestAccExampleThing_basic(t *testing.T) {
  ctx := acctest.Context(t)
  // ... omitted for brevity ...
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    // ... omitted for brevity ...
    Steps: []resource.TestStep{
      {
        // ... omitted for brevity ...
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName),
          acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "example", fmt.Sprintf("thing/%s", rName)),
          resource.TestCheckResourceAttr(resourceName, "description", ""),
          resource.TestCheckResourceAttr(resourceName, "name", rName),
        ),
      },
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

// below all TestAcc functions

func testAccExampleThingConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q
}
`, rName)
}
```

#### Basic Acceptance Tests

Usually this test is implemented first. The test configuration should contain only required arguments (`Required: true` attributes) and it should check the values of all read-only attributes (`Computed: true` without `Optional: true`). If the resource supports it, it verifies import. It should _NOT_ perform other `TestStep` such as updates or verify recreation.

These are typically named `TestAcc{SERVICE}{THING}_basic`, e.g., `TestAccCloudWatchDashboard_basic`

For example:

```go
func TestAccExampleThing_basic(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckExampleThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName),
          acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "example", fmt.Sprintf("thing/%s", rName)),
          resource.TestCheckResourceAttr(resourceName, "description", ""),
          resource.TestCheckResourceAttr(resourceName, "name", rName),
        ),
      },
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

// below all TestAcc functions

func testAccExampleThingConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q
}
`, rName)
}
```

#### PreChecks

Acceptance test cases have a PreCheck. The PreCheck ensures that the testing environment meets certain preconditions. If the environment does not meet the preconditions, Go skips the test. Skipping a test avoids reporting a failure and wasting resources where the test cannot succeed.

Here is an example of the default PreCheck:

```go
func TestAccExampleThing_basic(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { acctest.PreCheck(ctx, t) },
    // ... additional checks follow ...
  })
}
```

Extend the default PreCheck by adding calls to functions in the anonymous PreCheck function. The functions can be existing functions in the provider or custom functions you add for new capabilities.

##### Standard Provider PreChecks

If you add a new test that has preconditions which are checked by an existing provider function, use that standard PreCheck instead of creating a new one. Some existing tests are missing standard PreChecks and you can help by adding them where appropriate.

These are some of the standard provider PreChecks:

* `acctest.PreCheckRegion(t *testing.T, regions ...string)` checks that the test region is one of the specified AWS Regions.
* `acctest.PreCheckRegionNot(t *testing.T, regions ...string)` checks that the test region is not one of the specified AWS Regions.
* `acctest.PreCheckAlternateRegionIs(t *testing.T, region string)` checks that the alternate test region is the specified AWS Region.
* `acctest.PreCheckPartition(t *testing.T, partition string)` checks that the test partition is the specified partition.
* `acctest.PreCheckPartitionNot(t *testing.T, partitions ...string)` checks that the test partition is not one of the specified partitions.
* `acctest.PreCheckPartitionHasService(t *testing.T, serviceID string)` checks whether the current partition lists the service as part of its offerings. Note: AWS may not add new or public preview services to the service list immediately. This function will return a false positive in that case.
* `acctest.PreCheckOrganizationsAccount(ctx context.Context, t *testing.T)` checks whether the current account can perform AWS Organizations tests.
* `acctest.PreCheckAlternateAccount(t *testing.T)` checks whether the environment is set up for tests across accounts.
* `acctest.PreCheckMultipleRegion(t *testing.T, regions int)` checks whether the environment is set up for tests across regions.

This is an example of using a standard PreCheck function. For an established service, such as WAF or FSx, use `acctest.PreCheckPartitionHasService()` and the service endpoint ID to check that a partition supports the service.

```go
func TestAccExampleThing_basic(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, waf.EndpointsID) },
    // ... additional checks follow ...
  })
}
```

##### Custom PreChecks

In situations where standard PreChecks do not test for the required preconditions, create a custom PreCheck.

Below is an example of adding a custom PreCheck function. For a new or preview service that AWS does not include in the partition service list yet, you can verify the existence of the service with a simple read-only request (e.g., list all X service things). (For acceptance tests of established services, use `acctest.PreCheckPartitionHasService()` instead.)

```go
func TestAccExampleThing_basic(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { acctest.PreCheck(ctx, t), testAccPreCheckExample(ctx, t) },
    // ... additional checks follow ...
  })
}

func testAccPreCheckExample(ctx context.Context, t *testing.T) {
  conn := acctest.Provider.Meta().(*conns.AWSClient).ExampleConn(ctx)
	input := &example.ListThingsInput{}
	_, err := conn.ListThingsWithContext(ctx, input)
	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
```

#### ErrorChecks

Acceptance test cases have an ErrorCheck. The ErrorCheck provides a chance to take a look at errors before the test fails. While most errors should result in test failure, some should not. For example, an error that indicates an API operation is not supported in a particular region should cause the test to skip instead of fail. Since errors should flow through the ErrorCheck, do not handle the vast majority of failing conditions. Instead, in ErrorCheck, focus on the rare errors that should cause a test to skip, or in other words, be ignored.

##### Common ErrorCheck

In many situations, the common ErrorCheck is sufficient. It will skip tests for several normal occurrences such as when AWS reports a feature is not supported in the current region.

Here is an example of the common ErrorCheck:

```go
func TestAccExampleThing_basic(t *testing.T) {
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    // PreCheck
    ErrorCheck:   acctest.ErrorCheck(t, names.ExampleServiceID),
    // ... additional checks follow ...
  })
}
```

##### Service-Specific ErrorChecks

However, some services have special conditions that aren't caught by the common ErrorCheck. In these cases, you can create a service-specific ErrorCheck.

To add a service-specific ErrorCheck, follow these steps:

1. Make sure there is not already an ErrorCheck for the service you have in mind. For example, search the codebase for `acctest.RegisterServiceErrorCheckFunc(names.ExampleServiceID` replacing "service" with the package name of the service you're working on (e.g., `ec2`). If there is already an ErrorCheck for the service, add it to the existing service-specific ErrorCheck.
2. Create the service-specific ErrorCheck in an `_test.go` file for the service. See the example below.
3. Register the new service-specific ErrorCheck in the `init()` at the top of the `_test.go` file. See the example below.

An example of adding a service-specific ErrorCheck:

```go
// just after the imports, create or add to the init() function
func init() {
  acctest.RegisterServiceErrorCheck(names.ExampleServiceID, testAccErrorCheckSkipService)
}

// ... additional code and tests ...

// this is the service-specific ErrorCheck
func testAccErrorCheckSkipService(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Error message specific to the service that indicates unsupported features",
		"You can include from one to many portions of error messages",
		"Be careful to not inadvertently capture errors that should not be skipped",
	)
}
```

#### Long-Running Test Guards

For any acceptance tests that typically run longer than 300 seconds (5 minutes), add a `-short` test guard at the top of the test function.

For example:

```go
func TestAccExampleThing_longRunningTest(t *testing.T) {
  if testing.Short() {
    t.Skip("skipping long-running test in short mode")
  }

  // ... omitted for brevity ...

  resource.ParallelTest(t, resource.TestCase{
    // ... omitted for brevity ...
  })
}
```

When running acceptance tests, tests with these guards can be skipped using the Go `-short` flag. See [Running Only Short Tests](#running-only-short-tests) for examples.

#### Disappears Acceptance Tests

This test is generally implemented second. It is straightforward to set up once the basic test is passing since it can reuse that test configuration. It prevents a common bug report with Terraform resources that error when they can not be found (e.g., deleted outside Terraform).

These are typically named `TestAcc{SERVICE}{THING}_disappears`, e.g., `TestAccCloudWatchDashboard_disappears`

For example:

```go
func TestAccExampleThing_disappears(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckExampleThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName, &job),
          acctest.CheckResourceDisappears(ctx, acctest.Provider, ResourceExampleThing(), resourceName),
        ),
        ExpectNonEmptyPlan: true,
      },
    },
  })
}
```

If this test does fail, the fix for this is generally adding error handling immediately after the `Read` API call that catches the error and tells Terraform to remove the resource before returning the error:

```go
output, err := conn.GetThing(input)

if !d.IsNewResource() && tfresource.NotFound(err) {
  log.Printf("[WARN] Example Thing (%s) not found, removing from state", d.Id())
  d.SetId("")
  return nil
}

if err != nil {
  return fmt.Errorf("reading Example Thing (%s): %w", d.Id(), err)
}
```

For children resources that are encapsulated by a parent resource, it is also preferable to verify that removing the parent resource will not generate an error either. These are typically named `TestAcc{SERVICE}{THING}_disappears_{PARENT}`, e.g., `TestAccRoute53ZoneAssociation_disappears_Vpc`

```go
func TestAccExampleChildThing_disappears_ParentThing(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  parentResourceName := "aws_example_parent_thing.test"
  resourceName := "aws_example_child_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckExampleChildThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName),
          acctest.CheckResourceDisappears(ctx, acctest.Provider, ResourceExampleParentThing(), parentResourceName),
        ),
        ExpectNonEmptyPlan: true,
      },
    },
  })
}
```

#### Per Attribute Acceptance Tests

These are typically named `TestAcc{SERVICE}{THING}_{ATTRIBUTE}`, e.g., `TestAccCloudWatchDashboard_Name`

For example:

```go
func TestAccExampleThing_Description(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckExampleThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingConfigDescription(rName, "description1"),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName),
          resource.TestCheckResourceAttr(resourceName, "description", "description1"),
        ),
      },
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
      {
        Config: testAccExampleThingConfigDescription(rName, "description2"),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName),
          resource.TestCheckResourceAttr(resourceName, "description", "description2"),
        ),
      },
    },
  })
}

// below all TestAcc functions

func testAccExampleThingConfigDescription(rName string, description string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}
```

#### Cross-Account Acceptance Tests

When testing requires AWS infrastructure in a second AWS account, the below changes to the normal setup will allow the management or reference of resources and data sources across accounts:

- In the `PreCheck` function, include `acctest.PreCheckOrganizationsAccount(ctx, t)` to ensure a standardized set of information is required for cross-account testing credentials
- Switch usage of `ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories` to `ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t)`
- Add `acctest.ConfigAlternateAccountProvider()` to the test configuration and use `provider = awsalternate` for cross-account resources. The resource that is the focus of the acceptance test should _not_ use the alternate provider identification to simplify the testing setup.
- For any `TestStep` that includes `ImportState: true`, add the `Config` that matches the previous `TestStep` `Config`

An example acceptance test implementation can be seen below:

```go
func TestAccExample_basic(t *testing.T) {
  ctx := acctest.Context(t)
  resourceName := "aws_example.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck: func() {
      acctest.PreCheck(ctx, t)
      acctest.PreCheckOrganizationsAccount(ctx, t)
    },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
    ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
    CheckDestroy:             testAccCheckExampleDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleConfig(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleExists(ctx, resourceName),
          // ... additional checks ...
        ),
      },
      {
        Config:            testAccExampleConfig(),
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

func testAccExampleConfig() string {
  return acctest.ConfigCompose(
    acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
# Cross account resources should be handled by the cross account provider.
# The standardized provider block to use is awsalternate as seen below.
resource "aws_cross_account_example" "test" {
  provider = awsalternate

  # ... configuration ...
}

# The resource that is the focus of the testing should be handled by the default provider,
# which is automatically done by not specifying the provider configuration in the resource.
resource "aws_example" "test" {
  # ... configuration ...
}
`, ...))
}
```

Searching for the usage of `acctest.PreCheckOrganizationsAccount` in the codebase will yield real-world examples of this setup in action.

#### Cross-Region Acceptance Tests

When testing requires AWS infrastructure in a second or third AWS region, the below changes to the normal setup will allow the management or reference of resources and data sources across regions:

- In the `PreCheck` function, include `acctest.PreCheckMultipleRegion(t, ###)` to ensure a standardized set of information is required for cross-region testing configuration. If the infrastructure in the second AWS region is also in a second AWS account also include `acctest.PreCheckOrganizationsAccount(ctx, t)`
- Switch usage of `ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories` to `ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2)` (where the last parameter is number of regions, 2 or 3)
- Add `acctest.ConfigMultipleRegionProvider(###)` to the test configuration and use `provider = awsalternate` (and potentially `provider = awsthird`) for cross-region resources. The resource that is the focus of the acceptance test should _not_ use the alternative providers to simplify the testing setup. If the infrastructure in the second AWS region is also in a second AWS account use `testAccAlternateAccountAlternateRegionProviderConfig()` (EC2) instead
- For any `TestStep` that includes `ImportState: true`, add the `Config` that matches the previous `TestStep` `Config`

An example acceptance test implementation can be seen below:

```go
func TestAccExample_basic(t *testing.T) {
  ctx := acctest.Context(t)
  var providers []*schema.Provider
  resourceName := "aws_example.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck: func() {
      acctest.PreCheck(ctx, t)
      acctest.PreCheckMultipleRegion(t, 2)
    },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
    ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
    CheckDestroy:             testAccCheckExampleDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleConfig(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleExists(ctx, resourceName),
          // ... additional checks ...
        ),
      },
      {
        Config:            testAccExampleConfig(),
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

func testAccExampleConfig() string {
  return acctest.ConfigCompose(
    acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
# Cross region resources should be handled by the cross region provider.
# The standardized provider is awsalternate as seen below.
resource "aws_cross_region_example" "test" {
  provider = awsalternate

  # ... configuration ...
}

# The resource that is the focus of the testing should be handled by the default provider,
# which is automatically done by not specifying the provider configuration in the resource.
resource "aws_example" "test" {
  # ... configuration ...
}
`, ...))
}
```

Searching for the usage of `acctest.PreCheckMultipleRegion` in the codebase will yield real-world examples of this setup in action.

#### Acceptance Test Concurrency

Certain AWS service APIs allow a limited number of a certain component, while the acceptance testing runs at a default concurrency of twenty tests at a time. For example as of this writing, the SageMaker service only allows one SageMaker Domain per AWS Region. Running the tests with the default concurrency will fail with API errors relating to the component quota being exceeded.

When encountering these types of components, acceptance testing can be set up to limit the available concurrency of that particular component. When limited to one component at a time, this may also be referred to as serializing the acceptance tests.

To convert to serialized (one test at a time) acceptance testing:

- Convert all existing capital `T` test functions with the limited component to begin with a lowercase `t`, e.g., `TestAccSageMakerDomain_basic` becomes `testDomain_basic`. This will prevent the test framework from executing these tests directly as the prefix `Test` is required.
    - In each of these test functions, convert `resource.ParallelTest` to `resource.Test`
- Create a capital `T` `TestAcc{Service}{Thing}_serial` test function that then references all the lowercase `t` test functions. If multiple test files are referenced, this new test be created in a new shared file such as `internal/service/{SERVICE}/{SERVICE}_test.go`. The contents of this test can be set like the following:

```go
func TestAccExampleThing_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Thing": {
			"basic":        testAccThing_basic,
			"disappears":   testAccThing_disappears,
			// ... potentially other resource tests ...
		},
		// ... potentially other top level resource test groups ...
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
```

!!! note
    Future iterations of these acceptance testing concurrency instructions will include the ability to handle more than one component at a time including service quota lookup, if supported by the service API.

### Data Source Acceptance Testing

Writing acceptance testing for data sources is similar to resources, with the biggest changes being:

- Adding `DataSource` to the test and configuration naming, such as `TestAccExampleThingDataSource_Filter`
- The basic test _may_ be named after the easiest lookup attribute instead, e.g., `TestAccExampleThingDataSource_Name`
- No disappears testing
- Almost all checks should be done with [`resource.TestCheckResourceAttrPair()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-testing/helper/resource?tab=doc#TestCheckResourceAttrPair) to compare the data source attributes to the resource attributes
- The usage of an additional `dataSourceName` variable to store a data source reference, e.g., `data.aws_example_thing.test`

Data sources testing should still use the `CheckDestroy` function of the resource, just to continue verifying that there are no dangling AWS resources after a test is run.

Please note that we do not recommend re-using test configurations between resources and their associated data source as it is harder to discover testing regressions. Authors are encouraged to potentially implement similar "base" configurations though.

For example:

```go
func TestAccExampleThingDataSource_Name(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  dataSourceName := "data.aws_example_thing.test"
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
    ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckExampleThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingDataSourceConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName),
          resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
          resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
          resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
        ),
      },
    },
  })
}

// below all TestAcc functions

func testAccExampleThingDataSourceConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q
}

data "aws_example_thing" "test" {
  name = aws_example_thing.test.name
}
`, rName)
}
```

## Acceptance Test Sweepers

When running the acceptance tests, especially when developing or troubleshooting Terraform resources, it's possible for code bugs or other issues to prevent the proper destruction of AWS infrastructure. To prevent lingering resources from consuming quota or causing unexpected billing, the Terraform Plugin SDK supports the test sweeper framework to clear out an AWS region of all resources. This section is meant to augment the [SDKv2 documentation on test sweepers](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests/sweepers) with Terraform AWS Provider specific details.

### Running Test Sweepers

!!! warning
    Test Sweepers will destroy AWS infrastructure and backups in the target AWS account and region! These are designed to override any API deletion protection. Never run these outside a development AWS account that should be completely empty of resources. <!-- markdownlint-disable-line code-block-style -->

To run the sweepers for all resources in `us-west-2` and `us-east-1` (default testing regions):

```console
make sweep
```

To run a specific resource sweeper:

```console
SWEEPARGS=-sweep-run=aws_example_thing make sweep
```

To run sweepers with an assumed role, use the following additional environment variables:

* `TF_AWS_ASSUME_ROLE_ARN` - Required.
* `TF_AWS_ASSUME_ROLE_DURATION` - Optional, defaults to 1 hour (3600).
* `TF_AWS_ASSUME_ROLE_EXTERNAL_ID` - Optional.
* `TF_AWS_ASSUME_ROLE_SESSION_NAME` - Optional.

### Sweeper Checklists

- __Add Resource Sweeper Implementation__: See [Writing Test Sweepers](#writing-test-sweepers).
- __Add Service To Sweeper List__: Once a `sweep.go` file is present in the service subdirectory, run `make gen` to regenerate the list of imports in `internal/sweep/sweep_test.go`.

### Writing Test Sweepers

Sweeper logic should be written to a file called `sweep.go` in the appropriate service subdirectory (`internal/service/{serviceName}`). This file should include the following build tags above the package declaration:

```go
package example
```

Next, register the resource into the test sweeper framework:

```go
func RegisterSweepers() {
  resource.AddTestSweepers("aws_example_thing", &resource.Sweeper{
    Name: "aws_example_thing",
    F:    sweepThings,
    // Optionally
    Dependencies: []string{
      "aws_other_thing",
    },
  })
}
```

Then add the actual implementation. Preferably, if a paginated SDK call is available:

```go
func sweepThings(region string) error {
  ctx := sweep.Context(region)
  client, err := sweep.SharedRegionalSweepClient(ctx, region)

  if err != nil {
    return fmt.Errorf("getting client: %w", err)
  }

  conn := client.ExampleConn(ctx)
  sweepResources := make([]sweep.Sweepable, 0)
  var errs *multierror.Error

  input := &example.ListThingsInput{}

  err = conn.ListThingsPages(input, func(page *example.ListThingsOutput, lastPage bool) bool {
    if page == nil {
      return !lastPage
    }

    for _, thing := range page.Things {
      r := ResourceThing()
      d := r.Data(nil)

      id := aws.StringValue(thing.Id)
      d.SetId(id)

      // Perform resource specific pre-sweep setup.
      // For example, you may need to perform one or more of these types of pre-sweep tasks, specific to the resource:
      //
      // err := sdk.ReadResource(ctx, r, d, client) // fill in data
      // d.Set("skip_final_snapshot", true)           // set an argument in order to delete

      // This "if" is only needed if the pre-sweep setup can produce errors.
      // Otherwise, do not include it.
      if err != nil {
        err := fmt.Errorf("reading Example Thing (%s): %w", id, err)
        errs = multierror.Append(errs, err)
        continue
      }

      sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
    }

    return !lastPage
  })

  if awsv1.SkipSweepError(err) {
    log.Printf("[WARN] Skipping Example Thing sweep for %s: %s", region, errs)
    return nil
  }
  if err != nil {
    errs = multierror.Append(errs, fmt.Errorf("listing Example Things for %s: %w", region, err))
  }

  if err := sweep.SweepOrchestrator(sweepResources); err != nil {
    errs = multierror.Append(errs, fmt.Errorf("sweeping Example Things for %s: %w", region, err))
  }

  return errs.ErrorOrNil()
}
```

If no paginated SDK call is available,
consider generating one using the [`listpages` generator](https://github.com/hashicorp/terraform-provider-aws/blob/main/internal/generate/listpages/README.md),
or implement the sweeper as follows:

```go
func sweepThings(region string) error {
  ctx := sweep.Context(region)
  client, err := sweep.SharedRegionalSweepClient(ctx, region)

  if err != nil {
    return fmt.Errorf("getting client: %w", err)
  }

  conn := client.ExampleConn(ctx)
  sweepResources := make([]sweep.Sweepable, 0)
  var errs *multierror.Error

  input := &example.ListThingsInput{}

  for {
    output, err := conn.ListThings(input)
    if awsv1.SkipSweepError(err) {
      log.Printf("[WARN] Skipping Example Thing sweep for %s: %s", region, errs)
      return nil
    }
    if err != nil {
      errs = multierror.Append(errs, fmt.Errorf("listing Example Things for %s: %w", region, err))
      return errs.ErrorOrNil()
    }

    for _, thing := range output.Things {
      r := ResourceThing()
      d := r.Data(nil)

      id := aws.StringValue(thing.Id)
      d.SetId(id)

      // Perform resource specific pre-sweep setup.
      // For example, you may need to perform one or more of these types of pre-sweep tasks, specific to the resource:
      //
      // err := sdk.ReadResource(ctx, r, d, client) // fill in data
      // d.Set("skip_final_snapshot", true)           // set an argument in order to delete

      // This "if" is only needed if the pre-sweep setup can produce errors.
      // Otherwise, do not include it.
      if err != nil {
        err := fmt.Errorf("reading Example Thing (%s): %w", id, err)
        errs = multierror.Append(errs, err)
        continue
      }

      sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
    }

    if aws.StringValue(output.NextToken) == "" {
      break
    }

    input.NextToken = output.NextToken
  }

  if err := sweep.SweepOrchestrator(sweepResources); err != nil {
    errs = multierror.Append(errs, fmt.Errorf("sweeping Example Thing for %s: %w", region, err))
  }

  return errs.ErrorOrNil()
}
```

## Acceptance Test Checklists

There are several aspects to writing good acceptance tests. These checklists will help ensure effective testing from the design stage through to implementation details.

### Basic Acceptance Test Design

These are basic principles to help guide the creation of acceptance tests.

- __Covers Changes__: Every line of resource or data source code added or changed should be covered by one or more tests. For example, if a resource has two ways of functioning, tests should cover both possible paths. Nearly every codebase change needs test coverage to ensure functionality and prevent future regressions. If a bug or other problem prompted a fix, a test should be added that previously would have failed, especially if the report included a configuration.
- __Follows the Single Responsibility Principle__: Every test should have a single responsibility and effectively test that responsibility. This may include individual tests for verifying basic functionality of the resource (Create, Read, Delete), separately verifying using and updating a single attribute in a resource, or separately changing between two attributes to verify two "modes"/"types" possible with a resource configuration. In following this principle, test configurations should be as simple as possible. For example, not including extra configuration unless it is necessary for the specific test.

### Test Implementation

Below are the required items that will be noted during the submission review and prevent immediate merging:

- __Implements CheckDestroy__: Resource testing should include a `CheckDestroy` function (typically named `testAccCheck{SERVICE}{RESOURCE}Destroy`) that calls the API to verify that the Terraform resource has been deleted or disassociated as appropriate. More information about `CheckDestroy` functions can be found in the [SDKv2 TestCase documentation](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests/testcase#checkdestroy).
- __Implements Exists Check Function__: Resource testing should include a `TestCheckFunc` function (typically named `testAccCheck{SERVICE}{RESOURCE}Exists`) that calls the API to verify that the Terraform resource has been created or associated as appropriate. Preferably, this function will also accept a pointer to an API object representing the Terraform resource from the API response that can be set for potential usage in later `TestCheckFunc`. More information about these functions can be found in the [SDKv2 Custom Check Functions documentation](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests/teststep#custom-check-functions).
- __Excludes Provider Declarations__: Test configurations should not include `provider "aws" {...}` declarations. If necessary, only the provider declarations in `acctest.go` should be used for multiple account/region or otherwise specialized testing.
- __Passes in us-west-2 Region__: Tests default to running in `us-west-2` and at a minimum should pass in that region or include necessary `PreCheck` functions to skip the test when run outside an expected environment.
- __Includes ErrorCheck__: All acceptance tests should include a call to the common ErrorCheck (`ErrorCheck:   acctest.ErrorCheck(t, names.ExampleServiceID),`).
- __Uses resource.ParallelTest__: Tests should use [`resource.ParallelTest()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#ParallelTest) instead of [`resource.Test()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#Test) except where serialized testing is absolutely required.
- [ ] __Uses fmt.Sprintf()__: Test configurations preferably should be separated into their own functions (typically named `testAcc{SERVICE}{RESOURCE}Config{PURPOSE}`) that call [`fmt.Sprintf()`](https://golang.org/pkg/fmt/#Sprintf) for variable injection or a string `const` for completely static configurations. Test configurations should avoid `var` or other variable injection functionality such as [`text/template`](https://golang.org/pkg/text/template/).
- __Uses Randomized Infrastructure Naming__: Test configurations that use resources where a unique name is required should generate a random name. Typically this is created via `rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)` in the acceptance test function before generating the configuration.
- __Prevents S3 Bucket Deletion Errors__: Test configurations that use `aws_s3_bucket` resources as a logging destination should include the `force_destroy = true` configuration. This is to prevent race conditions where logging objects may be written during the testing duration which will cause `BucketNotEmpty` errors during deletion.

For resources that support import, the additional item below is required that will be noted during the submission review and prevent immediate merging:

- __Implements ImportState Testing__: Tests should include an additional `TestStep` configuration that verifies resource import via `ImportState: true` and `ImportStateVerify: true`. This `TestStep` should be added to all possible tests for the resource to ensure that all infrastructure configurations are properly imported into Terraform.

The below are style-based items that _may_ be noted during review and are recommended for simplicity, consistency, and quality assurance:

- __Uses Builtin Check Functions__: Tests should use already available check functions, e.g. `resource.TestCheckResourceAttr()`, to verify values in the Terraform state over creating custom `TestCheckFunc`. More information about these functions can be found in the [SDKv2 Builtin Check Functions documentation](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests/teststep#builtin-check-functions).
- __Uses TestCheckResourceAttrPair() for Data Sources__: Tests should use [`resource.TestCheckResourceAttrPair()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#TestCheckResourceAttrPair) to verify values in the Terraform state for data sources attributes to compare them with their expected resource attributes.
- __Excludes Timeouts Configurations__: Test configurations should not include `timeouts {...}` configuration blocks except for explicit testing of customizable timeouts (typically very short timeouts with `ExpectError`).
- __Implements Default and Zero Value Validation__: The basic test for a resource (typically named `TestAcc{SERVICE}{RESOURCE}_basic`) should use available check functions, e.g. `resource.TestCheckResourceAttr()`, to verify default and zero values in the Terraform state for all attributes. Empty/missing configuration blocks can be verified with `resource.TestCheckResourceAttr(resourceName, "{ATTRIBUTE}.#", "0")` and empty maps with `resource.TestCheckResourceAttr(resourceName, "{ATTRIBUTE}.%", "0")`

### Avoid Hard Coding

Avoid hard coding values in acceptance test checks and configurations for consistency and testing flexibility. Resource testing is expected to pass across multiple AWS environments supported by the Terraform AWS Provider (e.g., AWS Standard and AWS GovCloud (US)). Contributors are not expected or required to perform testing outside of AWS Standard, e.g., running only in the `us-west-2` region is perfectly acceptable. However, contributors are expected to avoid hard coding with these guidelines.

#### Hardcoded Account IDs

- __Uses Account Data Sources__: Any hardcoded account numbers in configuration, e.g., `137112412989`, should be replaced with a data source. Depending on the situation, there are several data sources for account IDs including:
    - [`aws_caller_identity` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity),
    - [`aws_canonical_user_id` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/canonical_user_id),
    - [`aws_billing_service_account` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/billing_service_account), and
    - [`aws_sagemaker_prebuilt_ecr_image` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/sagemaker_prebuilt_ecr_image).
- __Uses Account Test Checks__: Any check required to verify an AWS Account ID of the current testing account or another account should use one of the following available helper functions over the usage of `resource.TestCheckResourceAttrSet()` and `resource.TestMatchResourceAttr()`:
    - `acctest.CheckResourceAttrAccountID()`: Validates the state value equals the AWS Account ID of the current account running the test. This is the most common implementation.
    - `acctest.MatchResourceAttrAccountID()`: Validates the state value matches any AWS Account ID (e.g. a 12 digit number). This is typically only used in data source testing of AWS managed components.

Here's an example of using `aws_caller_identity`:

```terraform
data "aws_caller_identity" "current" {}

resource "aws_backup_selection" "test" {
  plan_id      = aws_backup_plan.test.id
  name         = "tf_acc_test_backup_selection_%[1]d"
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"
}
```

#### Hardcoded AMI IDs

- __Uses aws_ami Data Source__: Any hardcoded AMI ID configuration, e.g. `ami-12345678`, should be replaced with the [`aws_ami` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ami) pointing to an Amazon Linux image. The package `internal/acctest` includes test configuration helper functions to simplify these lookups:
    - `acctest.ConfigLatestAmazonLinuxHVMEBSAMI()`: The recommended AMI for most situations, using Amazon Linux, HVM virtualization, and EBS storage. To reference the AMI ID in the test configuration: `data.aws_ami.amzn-ami-minimal-hvm-ebs.id`.
    - `testAccLatestAmazonLinuxHVMInstanceStoreAMIConfig()` (EC2): AMI lookup using Amazon Linux, HVM virtualization, and Instance Store storage. Should only be used in testing that requires Instance Store storage rather than EBS. To reference the AMI ID in the test configuration: `data.aws_ami.amzn-ami-minimal-hvm-instance-store.id`.
    - `testAccLatestAmazonLinuxPVEBSAMIConfig()` (EC2): AMI lookup using Amazon Linux, Paravirtual virtualization, and EBS storage. Should only be used in testing that requires Paravirtual over Hardware Virtual Machine (HVM) virtualization. To reference the AMI ID in the test configuration: `data.aws_ami.amzn-ami-minimal-pv-ebs.id`.
    - `configLatestAmazonLinuxPvInstanceStoreAmi` (EC2): AMI lookup using Amazon Linux, Paravirtual virtualization, and Instance Store storage. Should only be used in testing that requires Paravirtual virtualization over HVM and Instance Store storage over EBS. To reference the AMI ID in the test configuration: `data.aws_ami.amzn-ami-minimal-pv-instance-store.id`.
    - `testAccLatestWindowsServer2016CoreAMIConfig()` (EC2): AMI lookup using Windows Server 2016 Core, HVM virtualization, and EBS storage. Should only be used in testing that requires Windows. To reference the AMI ID in the test configuration: `data.aws_ami.win2016core-ami.id`.

Here's an example of using `acctest.ConfigLatestAmazonLinuxHVMEBSAMI()` and `data.aws_ami.amzn-ami-minimal-hvm-ebs.id`:

```go
func testAccLaunchConfigurationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "m1.small"
}
`, rName))
}
```

#### Hardcoded Availability Zones

- __Uses aws_availability_zones Data Source__: Any hardcoded AWS Availability Zone configuration, e.g. `us-west-2a`, should be replaced with the [`aws_availability_zones` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/availability_zones). Use the convenience function called `acctest.ConfigAvailableAZsNoOptIn()` (defined in `internal/acctest/acctest.go`) to declare `data "aws_availability_zones" "available" {...}`. You can then reference the data source via `data.aws_availability_zones.available.names[0]` or `data.aws_availability_zones.available.names[count.index]` in resources using `count`.

Here's an example of using `acctest.ConfigAvailableAZsNoOptIn()` and `data.aws_availability_zones.available.names[0]`:

```go
func testAccInstanceVpcConfigBasic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id
}
`, rName))
}
```

#### Hardcoded Database Versions

- __Uses Database Version Data Sources__: Hardcoded database versions, e.g., RDS MySQL Engine Version `5.7.42`, should be removed (which means the AWS-defined default version will be used) or replaced with a list of preferred versions using a data source. Because versions change over time and version offerings vary from region to region and partition to partition, using the default version or providing a list of preferences ensures a version will be available. Depending on the situation, there are several data sources for versions, including:
    - [`aws_rds_engine_version` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/rds_engine_version),
    - [`aws_docdb_engine_version` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/docdb_engine_version), and
    - [`aws_neptune_engine_version` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/neptune_engine_version).

Here's an example of using `aws_rds_engine_version` and `data.aws_rds_engine_version.default.version`:

```terraform
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]
}

resource "aws_db_instance" "bar" {
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot  = true
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
}
```

#### Hardcoded Direct Connect Locations

- __Uses aws_dx_locations Data Source__: Hardcoded AWS Direct Connect locations, e.g., `EqSe2`, should be replaced with the [`aws_dx_locations` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/dx_locations).

Here's an example using `data.aws_dx_locations.test.location_codes`:

```terraform
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = "Test LAG"
  connections_bandwidth = "1Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
  force_destroy         = true
}
```

#### Hardcoded Instance Types

- __Uses Instance Type Data Source__: Singular hardcoded instance types and classes, e.g., `t2.micro` and `db.t2.micro`, should be replaced with a list of preferences using a data source. Because offerings vary from region to region and partition to partition, providing a list of preferences dramatically improves the likelihood that one of the options will be available. Depending on the situation, there are several data sources for instance types and classes, including:
    - [`aws_ec2_instance_type_offering` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ec2_instance_type_offering) - Convenience functions declare configurations that are referenced with `data.aws_ec2_instance_type_offering.available` including:
        - The `acctest.AvailableEC2InstanceTypeForAvailabilityZone()` function for test configurations using an EC2 Subnet which is inherently within a single Availability Zone
        - The `acctest.AvailableEC2InstanceTypeForRegion()` function for test configurations that do not include specific Availability Zones
    - [`aws_rds_orderable_db_instance` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/rds_orderable_db_instance),
    - [`aws_neptune_orderable_db_instance` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/neptune_orderable_db_instance), and
    - [`aws_docdb_orderable_db_instance` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/docdb_orderable_db_instance).

Here's an example of using `acctest.AvailableEC2InstanceTypeForRegion()` and `data.aws_ec2_instance_type_offering.available.instance_type`:

```go
func testAccSpotInstanceRequestConfig(rInt int) string {
	return acctest.ConfigCompose(
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
}
`, rInt))
}
```

Here's an example of using `aws_rds_orderable_db_instance` and `data.aws_rds_orderable_db_instance.test.instance_class`:

```terraform
data "aws_rds_orderable_db_instance" "test" {
  engine                     = "mysql"
  engine_version             = "5.7.31"
  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t3.small"]
}

resource "aws_db_instance" "test" {
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  username            = "test"
}
```

#### Hardcoded Partition DNS Suffix

- __Uses aws_partition Data Source__: Any hardcoded DNS suffix configuration, e.g., the `amazonaws.com` in a `ec2.amazonaws.com` service principal, should be replaced with the [`aws_partition` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/partition). A common pattern is declaring `data "aws_partition" "current" {}` and referencing it via `data.aws_partition.current.dns_suffix`.

Here's an example of using `aws_partition` and `data.aws_partition.current.dns_suffix`:

```terraform
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}
```

#### Hardcoded Partition in ARN

- __Uses aws_partition Data Source__: Any hardcoded AWS Partition configuration, e.g. the `aws` in a `arn:aws:SERVICE:REGION:ACCOUNT:RESOURCE` ARN, should be replaced with the [`aws_partition` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/partition). A common pattern is declaring `data "aws_partition" "current" {}` and referencing it via `data.aws_partition.current.partition`.
- __Uses Builtin ARN Check Functions__: Tests should use available ARN check functions to validate ARN attribute values in the Terraform state over `resource.TestCheckResourceAttrSet()` and `resource.TestMatchResourceAttr()`:

    - `acctest.CheckResourceAttrRegionalARN()` verifies that an ARN matches the account ID and region of the test execution with an exact resource value
    - `acctest.MatchResourceAttrRegionalARN()` verifies that an ARN matches the account ID and region of the test execution with a regular expression of the resource value
    - `acctest.CheckResourceAttrGlobalARN()` verifies that an ARN matches the account ID of the test execution with an exact resource value
    - `acctest.MatchResourceAttrGlobalARN()` verifies that an ARN matches the account ID of the test execution with a regular expression of the resource value
    - `acctest.CheckResourceAttrRegionalARNNoAccount()` verifies that an ARN has no account ID and matches the current region of the test execution with an exact resource value
    - `acctest.CheckResourceAttrGlobalARNNoAccount()` verifies that an ARN has no account ID and matches an exact resource value
    - `acctest.CheckResourceAttrRegionalARNAccountID()` verifies that an ARN matches a specific account ID and the current region of the test execution with an exact resource value
    - `acctest.CheckResourceAttrGlobalARNAccountID()` verifies that an ARN matches a specific account ID with an exact resource value

Here's an example of using `aws_partition` and `data.aws_partition.current.partition`:

```terraform
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}
```

#### Hardcoded Region

- __Uses aws_region Data Source__: Any hardcoded AWS Region configuration, e.g., `us-west-2`, should be replaced with the [`aws_region` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region). A common pattern is declaring `data "aws_region" "current" {}` and referencing it via `data.aws_region.current.name`

Here's an example of using `aws_region` and `data.aws_region.current.name`:

```terraform
data "aws_region" "current" {}

resource "aws_route53_zone" "test" {
  vpc {
    vpc_id     = aws_vpc.test.id
    vpc_region = data.aws_region.current.name
  }
}
```

#### Hardcoded Spot Price

- __Uses aws_ec2_spot_price Data Source__: Any hardcoded spot prices, e.g., `0.05`, should be replaced with the [`aws_ec2_spot_price` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ec2_spot_price). A common pattern is declaring `data "aws_ec2_spot_price" "current" {}` and referencing it via `data.aws_ec2_spot_price.current.spot_price`.

Here's an example of using `aws_ec2_spot_price` and `data.aws_ec2_spot_price.current.spot_price`:

```terraform
data "aws_ec2_spot_price" "current" {
  instance_type = "t3.medium"

  filter {
    name   = "product-description"
    values = ["Linux/UNIX"]
  }
}

resource "aws_spot_fleet_request" "test" {
  spot_price      = data.aws_ec2_spot_price.current.spot_price
  target_capacity = 2
}
```

#### Hardcoded SSH Keys

- __Uses acctest.RandSSHKeyPair() or RandSSHKeyPairSize() Functions__: Any hardcoded SSH keys should be replaced with random SSH keys generated by either the acceptance testing framework's function [`RandSSHKeyPair()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/helper/acctest#RandSSHKeyPair) or the provider function `RandSSHKeyPairSize()`. `RandSSHKeyPair()` generates 1024-bit keys.

Here's an example using `aws_key_pair`

```go
func TestAccKeyPair_basic(t *testing.T) {
  ...

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey, _, err := acctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("generating random SSH key: %s", err)
	}

  resource.ParallelTest(t, resource.TestCase{
		...
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig(rName, publicKey),
        ...
      },
    },
  })
}

func testAccKeyPairConfig(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}
`, rName, publicKey)
}
```

#### Hardcoded Email Addresses

- __Uses either acctest.DefaultEmailAddress Constant or acctest.RandomEmailAddress() Function__: Any hardcoded email addresses should replaced with either the constant `acctest.DefaultEmailAddress` or the function `acctest.RandomEmailAddress()`.

Using `acctest.DefaultEmailAddress` is preferred when using a single email address in an acceptance test.

Here's an example using `acctest.DefaultEmailAddress`

```go
func TestAccSNSTopicSubscription_email(t *testing.T) {
	...

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		...
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionEmailConfig(rName, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					...
					resource.TestCheckResourceAttr(resourceName, "endpoint", acctest.DefaultEmailAddress),
				),
			},
		},
	})
}
```

Here's an example using `acctest.RandomEmailAddress()`

```go
func TestAccPinpointEmailChannel_basic(t *testing.T) {
	...

	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		...
		Steps: []resource.TestStep{
			{
				Config: testAccEmailChannelConfig_FromAddress(domain, address1),
				Check: resource.ComposeTestCheckFunc(
					...
					resource.TestCheckResourceAttr(resourceName, "from_address", address1),
				),
			},
			{
				Config: testAccEmailChannelConfig_FromAddress(domain, address2),
				Check: resource.ComposeTestCheckFunc(
					...
					resource.TestCheckResourceAttr(resourceName, "from_address", address2),
				),
			},
		},
	})
}
```
