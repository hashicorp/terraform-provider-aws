# Changelog Process

HashiCorp’s open-source projects have always maintained user-friendly, readable `CHANGELOG.md` that allows users to tell at a glance whether a release should have any effect on them, and to gauge the risk of an upgrade.

We use [go-changelog](https://github.com/hashicorp/go-changelog) to generate the changelog from files created in the `.changelog/` directory.
It is important that when you raise your pull request, there is a changelog entry which describes the changes your contribution makes.
Not all changes require an entry in the changelog, guidance follows on what changes do.

## Changelog format

The changelog format requires an entry in the following format, where HEADER corresponds to the changelog category, and the entry is the changelog entry itself. The entry should be included in a file in the `.changelog` directory with the naming convention `{PR-NUMBER}.txt`. For example, to create a changelog entry for pull request 1234, there should be a file named `.changelog/1234.txt`.

``````
```release-note:{HEADER}
{ENTRY}
```
``````

If a pull request should contain multiple changelog entries, then multiple blocks can be added to the same changelog file. For example:

``````
```release-note:note
resource/aws_example_thing: The `broken` attribute has been deprecated. All configurations using `broken` should be updated to use the new `not_broken` attribute instead.
```

```release-note:enhancement
resource/aws_example_thing: Add `not_broken` attribute
```
``````

## Pull request types to CHANGELOG

The CHANGELOG is intended to show operator-impacting changes to the codebase for a particular version. If every change or commit to the code resulted in an entry, the CHANGELOG would become less useful for operators. The lists below are general guidelines and examples for when a decision needs to be made to decide whether a change should have an entry.

### Changes that should have a CHANGELOG entry

#### New resource

A new resource entry should only contain the name of the resource, and use the `release-note:new-resource` header.

``````
```release-note:new-resource
aws_secretsmanager_secret_policy
```
``````

#### New data source

A new data source entry should only contain the name of the data source, and use the `release-note:new-data-source` header.

``````
```release-note:new-data-source
aws_workspaces_workspace
```
``````

#### New full-length documentation guides (e.g., EKS Getting Started Guide, IAM Policy Documents with Terraform)

A new full-length documentation entry gives the title of the documentation added, using the `release-note:new-guide` header.

``````
```release-note:new-guide
Custom Service Endpoint Configuration
```
``````

#### Resource and provider bug fixes

A new bug entry should use the `release-note:bug` header and have a prefix indicating the resource or data source it corresponds to, a colon, then followed by a brief summary. Use a `provider` prefix for provider-level fixes.

``````
```release-note:bug
resource/aws_glue_classifier: Fix quote_symbol being optional
```
``````

#### Resource and provider enhancements

A new enhancement entry should use the `release-note:enhancement` header and have a prefix indicating the resource or data source it corresponds to, a colon, then followed by a brief summary. Use a `provider` prefix for provider-level enhancements.

``````
```release-note:enhancement
resource/aws_eip: Add network_border_group argument
```
``````

#### Deprecations

A deprecation entry should use the `release-note:note` header and have a prefix indicating the resource or data source it corresponds to, a colon, then followed by a brief summary. Use a `provider` prefix for provider-level changes.

``````
```release-note:note
resource/aws_dx_gateway_association: The vpn_gateway_id attribute is being deprecated in favor of the new associated_gateway_id attribute to support transit gateway associations
```
``````

#### Breaking changes and removals

A breaking-change entry should use the `release-note:breaking-change` header and have a prefix indicating the resource or data source it corresponds to, a colon, then followed by a brief summary. Use a `provider` prefix for provider-level changes.

``````
```release-note:breaking-change
resource/aws_lambda_alias: Resource import no longer converts Lambda Function name to ARN
```
``````

#### Region validation support

``````
```release-note:note
provider: Region validation now automatically supports the new `XX-XXXXX-#` (Location) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform AWS Provider will return errors during credential validation (e.g., `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`) or AWS operations will throw their own errors (e.g., `data.aws_availability_zones.available: Error fetching Availability Zones: AuthFailure: AWS was not able to validate the provided access credentials`). [GH-####]
```

```release-note:enhancement
provider: Support automatic region validation for `XX-XXXXX-#` [GH-####]
```
``````

### Changes that may have a CHANGELOG entry

Dependency updates: If the update contains relevant bug fixes or enhancements that affect operators, those should be called out.
Any changes which do not fit into the above categories but warrant highlighting.
Use resource/data source/provider prefixes where appropriate.

``````
```release-note:note
resource/aws_lambda_alias: Resource import no longer converts Lambda Function name to ARN
```
``````

### Changes that should _not_ have a CHANGELOG entry

- Resource and provider documentation updates
- Testing updates
- Code refactoring
