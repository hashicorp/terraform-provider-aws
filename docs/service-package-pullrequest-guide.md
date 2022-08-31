# Service Package Refactor Pull Request Guide

Pull request
[#21306](https://github.com/hashicorp/terraform-provider-aws/pull/21306) has
significantly refactored the AWS provider codebase. Specifically, the code for
all AWS resources and data sources has been relocated from a single `aws`
directory to a large number of separate directories in `internal/service`, each
corresponding to a particular AWS service. In addition to vastly simplifying
the codebase's overall structure, this change has also allowed us to simplify
the names of a number of underlying functions -- without encountering namespace
collisions. Issue
[#20000](https://github.com/hashicorp/terraform-provider-aws/issues/20000)
contains a more complete description of these changes.

As a result, nearly every pull request opened prior to the refactoring has merge
conflicts; they are attempting to apply changes to files that have since been
relocated. Furthermore, any new files or functions introduced must be brought
into line with the codebase's new conventions. The following steps are intended
to resolve such a conflict -- though it should be noted that this guide is an
active work in progress as additional pull requests are amended.

These fixes, however, *in no way affect the prioritization* of a particular
pull request. Once a pull request has been selected for review, the necessary
changes will be made by a maintainer -- either directly or in collaboration
with the pull request author.

## Fixing a Pre-Refactor Pull Request

1. `git checkout` the branch pertaining to the pull request you wish to amend

1. Begin a merge of the latest version of `main` branch into your local branch:
   `git pull origin main`. Merge conflicts are expected.

1. For any **new file**, rename and move the file to its appropriate service
   package directory:

   **Resource Files**

   ```
     git mv aws/resource_aws_{service_name}_{resource_name}.go \
     internal/service/{service_name}/{resource_name}.go
   ```

   **Resource Test Files**

   ```
     git mv aws/resource_aws_{service_name}_{resource_name}_test.go \
     internal/service/{service_name}/{resource_name}_test.go
   ```

   **Data Source Files**

   ```
     git mv aws/data_source_aws_{service_name}_{resource_name}.go \
     internal/service/{service_name}/{resource_name}_data_source.go
   ```

   **Data Source Test Files**

   ```
     git mv aws/data_source_aws_{service_name}_{resource_name}_test.go \
     internal/service/{service_name}/{resource_name}_data_source_test.go
   ```

1. For any new **function**, rename the function appropriately:

   **Resource Schema Functions**

   ```
     func resourceAws{ResourceName}() =>
     func Resource{ResourceName}()
   ```

   **Resource Generic Functions**

   ```
     func resourceAws{ServiceName}{ResourceName}{FunctionName}() =>
     func resource{ResourceName}{FunctionName}()
   ```

   **Resource Acceptance Test Functions**

   ```
     func TestAccAWS{ServiceName}{ResourceName}_{testType}() =>
     func TestAcc{ResourceName}_{testType}()
   ```

   **Data Source Schema Functions**

   ```
     func dataSourceAws{ResourceName}() =>
     func DataSource{ResourceName}()
   ```

   **Data Source Generic Functions**

   ```
     func dataSourceAws{ServiceName}{ResourceName}{FunctionName}() =>
     func dataSource{ResourceName}{FunctionName}()
   ```

   **Data Source Acceptance Test Functions**

   ```
     func TestAccDataSourceAWS{ServiceName}{ResourceName}_{testType}() =>
     func TestAcc{ResourceName}DataSource_{testType}()
   ```

   **Finder Functions**

   ```
     func finder.{FunctionName}() =>
     func Find{FunctionName}()
   ```

   **Status Functions**

   ```
     func waiter.{FunctionName}Status() =>
     func status{FunctionName}()
   ```

   **Waiter Functions**

   ```
     func waiter.{FunctionName}() =>
     func wait{FunctionName}()
   ```

1. If a file has a package declaration of `package aws`, you will need to change
   it to the new package location. For example, if you moved a file to `internal/service/ecs`,
   the declaration will now be `package ecs`.

   Any file that imports `"github.com/hashicorp/terraform-provider-aws/internal/acctest"` _must_
   be in the `<package>_test` package. For example, `internal/service/ecs/account_setting_default_test.go`
   does import the `acctest` package and must have a package declaration of `package ecs_test`.

1. If you have made any changes to `aws/provider.go`, you will have to manually
   re-enact those changes on the new `internal/provider/provider.go` file.

   Most commonly, these changes involve the addition of an entry to either the
   `DataSourcesMap` or `ResourcesMap`. If this is the case for your PR, you will have
   to adapt your entry to follow our new code conventions.

   **Resources Map Entries**

   ```
     "{aws_terraform_resource_type}":   resourceAws{ServiceName}{ResourceName}(), =>
     "{aws_terraform_resource_type}":   {serviceName}.Resource{ResourceName}(),
   ```

   **Data Source Map Entries**

   ```
     "{aws_terraform_data_source_type}":   dataSourceAws{ServiceName}{ResourceName}(), =>
     "{aws_terraform_data_source_type}":   {serviceName}.DataSource{ResourceName}(),
   ```

1. Some functions, constants, and variables have been moved, removed, or renamed.
   This table shows some of the common changes you may need to make to fix compile errors.

   | Before | Now |
   | --- | --- |
   | `isAWSErr(α, β, "<message>")` | `tfawserr.ErrMessageContains(α, β, "<message>")` |
   | `isAWSErr(α, β, "")` | `tfawserr.ErrCodeEquals(α, β)` |
   | `isResourceNotFoundError(α)` | `tfresource.NotFound(α)` |
   | `isResourceTimeoutError(α)` | `tfresource.TimedOut(α)` |
   | `testSweepSkipResourceError(α)` | `tfawserr.ErrCodeContains(α, "AccessDenied")` |
   | `testAccPreCheck(t)` | `acctest.PreCheck(t)` |
   | `testAccProviders` | `acctest.Providers` |
   | `acctest.RandomWithPrefix("tf-acc-test")` | `sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)` |
   | `composeConfig(α)` | `acctest.ConfigCompose(α)` |

1. Use `git status` to report the state of the merge. Review any merge
   conflicts -- being sure to adopt the new naming conventions described in the
   previous step where relevant. Use `git add` to add any new files to the commit.
