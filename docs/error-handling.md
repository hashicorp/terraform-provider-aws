<!-- markdownlint-configure-file { "code-block-style": false } -->
# Error Handling

The Terraform AWS Provider codebase bridges the implementation of a [Terraform Plugin](https://www.terraform.io/plugin/how-terraform-works) and an AWS API client to support AWS operations as Terraform Resources.
An important aspect of performing remote actions is properly handling operations which are not guaranteed to succeed.
Some common examples include unstable network connections, missing permissions, incorrect Terraform configurations, or unexpected responses from the remote system.
All of these situations lead to an unexpected workflow action that must be surfaced to Terraform for operators to troubleshoot.
This guide is intended to document best practices for surfacing these issues properly.

For further details about how the AWS SDK for Go and the Terraform AWS Provider handle retryable errors, see the [Retries and Waiters documentation](retries-and-waiters.md).

## General Guidelines and Helpers

### Naming and Check Style

Following typical Go conventions, error variables in the Terraform AWS Provider codebase should be named `err`, e.g.

```go
result, err := strconv.Itoa("oh no!")
```

The code that then checks these errors should prefer `if` conditionals that usually `return` (or in the case of looping constructs, `break`/`continue`) early, especially in the case of multiple error checks, e.g.

```go
if /* ... something checking err first ... */ {
    // ... return, break, continue, etc. ...
}

if err != nil {
    // ... return, break, continue, etc. ...
}

// all good!
```

This is in preference to some other styles of error checking, such as `switch` conditionals without a condition.

### Wrap Errors

Go implements error wrapping, which means that a deeply nested function call can return a particular error type, while each function up the stack can provide additional error message context without losing the ability to determine the original error.
Additional information about this concept can be found in the [Go blog entry titled Working with Errors in Go 1.13](https://blog.golang.org/go1.13-errors).

For most use cases in this codebase, this means if code is receiving an error and needs to return it, it should implement [`fmt.Errorf()`](https://pkg.go.dev/fmt#Errorf) and the `%w` verb, e.g.

```go
return fmt.Errorf("adding some additional message: %w", err)
```

This type of error wrapping should be applied to all Terraform resource logic where errors (not diagnostics) are returned.

### AWS SDK for Go Errors

V1 and V2 of the AWS SDK for Go approach errors differently.
Most notably, V2 makes use of more modern techniques like error wrapping that v1 does not.
The [Errors Types](https://aws.github.io/aws-sdk-go-v2/docs/migrating/#errors-types) section in the migration guide provides additional context on these differences.

The sections below contain documentation for both SDKs. See [AWS Go SDK Versions](aws-go-sdk-versions.md) for direction on which to use.

=== "AWS Go SDK V2 (Preferred)"
    The [AWS SDK for Go v2 documentation](https://aws.github.io/aws-sdk-go-v2/docs/) includes a [section on handling errors](https://aws.github.io/aws-sdk-go-v2/docs/handling-errors/), which is recommended reading.

    For the purposes of this documentation, the most important concepts in handling these errors are:

    - The SDK wraps all errors returned by service clients with the [`smithy.OperationError`](https://pkg.go.dev/github.com/aws/smithy-go/#OperationError) type.
    - [`errors.As`](https://golang.org/pkg/errors#As) should be used to unwrap errors when inspecting for a specific error type (e.g, a `BucketAlreadyExists` error from the S3 API).

=== "AWS Go SDK V1"
    The [AWS SDK for Go v1 documentation](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/welcome.html) includes a [section on handling errors](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/handling-errors.html), which is recommended reading.

    For the purposes of this documentation, the most important concepts in handling these errors are:

    - Each response error (which eventually implements `awserr.Error`) has a `string` error code (`Code`) and `string` error message (`Message`). When printed as a string, they format as: `Code: Message`, e.g., `InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX cannot be assumed by AWS Backup`.
    - Error handling is almost exclusively done via those `string` fields and not other response information, such as HTTP Status Codes.
    - When the error code is non-specific, the error message should also be checked. Unfortunately, AWS APIs generally do not provide documentation or API modeling with the contents of these messages and often the Terraform AWS Provider code must rely on substring matching.
    - Not all errors are returned in the response error from an AWS API operation. This is service- and sometimes API-call-specific. For example, the [EC2 `DeleteVpcEndpoints` API call](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DeleteVpcEndpoints.html) can return a "successful" response (in terms of no response error) but include information in an `Unsuccessful` field in the response body.

    When working with AWS SDK for Go v1 errors, it is preferred to use the helpers outlined below and use the `%w` format verb. Code should generally avoid type assertions with the underlying `awserr.Error` type or calling its `Code()`, `Error()`, `Message()`, or `String()` receiver methods. Using the `%v`, `%#v`, or `%+v` format verbs generally provides extraneous information that is not helpful to operators or code maintainers.

#### AWS SDK for Go Error Helpers

=== "AWS Go SDK V2 (Preferred)"
    To simplify operations with AWS SDK for Go error types, the following helpers are available via the `github.com/hashicorp/aws-sdk-go-base/v2/tfawserr` Go package:

    - `tfawserr.ErrCodeEquals` - Preferred when the error code is specific enough for the check condition. For example, a `ResourceNotFoundError` code provides enough information that the requested resource does not exist.
    - `tfawserr.ErrMessageContains`: Preferred when an error code can cover multiple failure modes, and additional parsing of the error message is required to determine if it matches a specific condition.

=== "AWS Go SDK V1"
    To simplify operations with AWS SDK for Go error types, the following helpers are available via the `github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr` Go package:

    - `tfawserr.ErrCodeEquals(err, "Code")`: Preferred when the error code is specific enough for the check condition. For example, a `ResourceNotFoundError` code provides enough information that the requested resource does not exist.
    - `tfawserr.ErrMessageContains(err, "Code", "MessageContains")`: Does simple substring matching for the error message.

The recommendation for error message checking is to be just specific enough to capture the anticipated issue, but not include _too_ much matching as the AWS API can change over time without notice.
The maintainers have observed changes in wording and capitalization cause unexpected issues in the past.

For example, given this error code and message:

```
InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX cannot be assumed by AWS Backup
```

An error check for this might be:

```go
if tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "cannot be assumed") {
    // Special handling here
}
```

The Amazon Resource Name in the error message will be different for every environment and does not add value to the check.
The AWS Backup suffix is also extraneous and could change should the service ever rename.

#### AWS SDK for Go Error Constants

=== "AWS Go SDK V2 (Preferred)"
    Each AWS SDK for Go v2 service API typically implements common error codes, which get exported as public structs in the SDK. In the [AWS SDK for Go v2 API Reference](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2), these can be found in each of the service packages `types` subpackage (typically named `{ErrorType}Exception`).

=== "AWS Go SDK V1"
    Each AWS SDK for Go v1 service API typically implements common error codes, which get exported as public constants in the SDK. In the [AWS SDK for Go v1 API Reference](https://docs.aws.amazon.com/sdk-for-go/api/), these can be found in each of the service packages under the `Constants` section (typically named `ErrCode{ExceptionName}`).

If an AWS SDK for Go service API is missing an error code constant, an AWS Support case should be submitted and a new constant can be added to `internal/service/{SERVICE}/errors.go` file (created if not present), e.g.

```go
const(
    ErrCodeInvalidParameterException = "InvalidParameterException"
)
```

Then referencing code can use it via:

```go
// imports
tf{SERVICE} "github.com/hashicorp/terraform-provider-aws/internal/service/{SERVICE}"

// logic
tfawserr.ErrCodeEquals(err, tf{SERVICE}.ErrCodeInvalidParameterException)
```

### Terraform Plugin Types and Helpers

The Terraform Plugin SDK includes some error types which are used in certain operations and typically preferred over implementing new types:

* [`retry.NotFoundError`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry#NotFoundError)
* [`retry.TimeoutError`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry#TimeoutError)
    * Returned from [`retry.RetryContext()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry#RetryContext) and
    [`(retry.StateChangeConf).WaitForStateContext()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry#StateChangeConf.WaitForStateContext)

!!! note
    While these helpers currently reside in the Terraform Plugin SDK V2 package, they can be used with Plugin Framework based resources. In the future these functions will likely be migrated into the provider itself, or a standalone library as there is no direct dependency on Plugin SDK functionality.

The Terraform AWS Provider codebase implements some additional helpers for working with these in the `internal/tfresource` package:

- `tfresource.NotFound(err)`: Returns true if the error is a `retry.NotFoundError`.
- `tfresource.TimedOut(err)`: Returns true if the error is a `retry.TimeoutError` and contains no `LastError`. This typically signifies that the retry logic was never signaled for a retry, which can happen when AWS API operations are automatically retrying before returning.

## Resource Lifecycle Guidelines

Terraform CLI and the Terraform Plugin libraries have certain expectations and automatic behaviors depending on the lifecycle operation of a resource.
This section highlights some common issues that can occur and their expected resolution.

### Resource Creation

For Terraform Plugin Framework based resources, creation is implemented on the `Create` method of the resource struct.
For Terraform Plugin SDK V2 based resources, creation is implemented on the `CreateWithoutTimeout` function of the resource schema definition.

#### Creation Error Message Context

=== "Terraform Plugin Framework (Preferred)"
    Errors encountered during creation should include additional messaging about the location or cause of the error for operators and code maintainers.
    Plugin Framework based resources will append error [diagnostics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics) to the `resource.CreateResponse` `Diagnostics` field.
    The `create.ProblemStandardMessage` helper function can be used for constructing the appropriate context to accompany the error.

    ```go
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionCreating, ResName{{ .Resource }}, plan.Name.String(), err),
            err.Error(),
        )
        return
    }
    ```

    e.g.

    ```go
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameVPCConnection, plan.Name.String(), err),
            err.Error(),
        )
        return
    }
    ```

    Resources that use other operations that return errors (e.g. waiters) should follow a similar pattern, including the resource identifier since it has typically been set before this execution:

    ```go
    createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
    waitOut, err := waitVPCConnectionCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForCreation, ResNameVPCConnection, plan.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

=== "Terraform Plugin SDK V2"
    Errors encountered during creation should include additional messaging about the location or cause of the error for operators and code maintainers.
    The `create.AppendDiagError` helper function can be used to convert a native error into a [diagnostic](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics), which is the method for surfacing warnings and errors in Terraform providers.

    ```go
    if err != nil {
        return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionCreating, ResName{{ .Resource }}, d.Get("name").(string), err)...)
    }
    ```

    e.g.

    ```go
    if err != nil {
        return create.AppendDiagError(diags, names.IVS, create.ErrActionCreating, ResNameRecordingConfiguration, d.Id(), err)
    }
    ```

    Resources that use other operations that return errors (e.g. waiters) should follow a similar pattern, including the resource identifier since it has typically been set before this execution:

    ```go
    if _, err := waitRecordingConfigurationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
        return create.AppendDiagError(diags, names.IVS, create.ErrActionWaitingForCreation, ResNameRecordingConfiguration, d.Id(), err)
    }
    ```

#### d.IsNewResource() Checks

!!! note
    This type of check only applies to Plugin SDK V2 based resources. Plugin Framework based resources rely solely on [retries and waiters](retries-and-waiters.md) to handle eventually consistent resources.

During resource creation, Terraform CLI expects either a properly applied state for the new resource or an error. To signal proper resource existence, the Terraform Plugin SDK uses an underlying resource identifier (set via `d.SetId(/* some value */)`). If for some reason the resource creation is returned without an error, but also without the resource identifier being set, Terraform CLI will return an error such as:

```sh
Error: Provider produced inconsistent result after apply

When applying changes to aws_sns_topic_subscription.sqs,
provider "registry.terraform.io/hashicorp/aws" produced an unexpected new
value: Root resource was present, but now absent.

This is a bug in the provider, which should be reported in the provider's own
issue tracker.
```

A typical pattern in resource implementations in the `CreateWithoutTimeout` function is to `return` the `ReadWithoutTimeout` function at the end to fill in the Terraform State for all attributes. Another typical pattern in resource implementations in the `ReadWithoutTimeout` function is to remove the resource from the Terraform State if the remote system returns an error or status that indicates the remote resource no longer exists by explicitly calling `d.SetId("")` and returning no error. If the remote system is not strongly read-after-write consistent (eventually consistent), this means the resource creation can return no error and also return no resource state.

To prevent this type of Terraform CLI error, the resource implementation should also check against `d.IsNewResource()` before removing from the Terraform State and returning no error. If that check is `true`, then remote operation error (or one synthesized from the non-existent status) should be returned instead. While adding this check will not fix the resource implementation to handle the eventually consistent nature of the remote system, the error being returned will be less opaque for operators and code maintainers to troubleshoot.

In the Terraform AWS Provider, an initial fix for the Terraform CLI error will typically look like:

```go
func resourceServiceThingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
 	var diags diag.Diagnostics

   /* ... */

    return append(diags, resourceServiceThingRead(ctx, d, meta)...)
}

func resourceServiceThingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
  	var diags diag.Diagnostics

   /* ... */

    output, err := conn.DescribeServiceThing(input)

    if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
        log.Printf("[WARN] {Service} {Thing} (%s) not found, removing from state", d.Id())
        d.SetId("")
        return diags
    }

    if err != nil {
        return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionReading, ResName{{ .Resource }}, d.Id(), err)
    }

    /* ... */
}
```

If the remote system is eventually consistent, see the [Retries and Waiters documentation on Resource Lifecycle Retries](retries-and-waiters.md#resource-lifecycle-retries) for how to prevent consistency-type errors.

### Resource Read

For Terraform Plugin Framework based resources, read is implemented on the `Read` method of the resource struct.
For Terraform Plugin SDK V2 based resources, read is implemented on the `ReadWithoutTimeout` function of the resource schema definition.

#### Read Error Message Context

=== "Terraform Plugin Framework (Preferred)"
    Errors encountered during read should include the resource identifier (for managed resources) and additional messaging about the location or cause of the error for operators and code maintainers.
    Plugin Framework based resources will append error [diagnostics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics) to the `resource.ReadResponse` `Diagnostics` field.
    The `create.ProblemStandardMessage` helper function can be used for constructing the appropriate context to accompany the error.

    ```go
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionReading, ResName{{ .Resource }}, state.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

    e.g.

    ```go
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, ResNameVPCConnection, state.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

=== "Terraform Plugin SDK V2"
    Errors encountered during read should include the resource identifier (for managed resources) and additional messaging about the location or cause of the error for operators and code maintainers.
    The `create.AppendDiagError` helper function can be used to convert a native error into a [diagnostic](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics), which is the method for surfacing warnings and errors in Terraform providers.

    ```go
    if err != nil {
        return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionReading, ResName{{ .Resource }}, d.Id(), err)
    }
    ```

    e.g.

    ```go
    if err != nil {
        return create.AppendDiagError(diags, names.IVS, create.ErrActionReading, ResNameRecordingConfiguration, d.Id(), err)
    }
    ```

#### Singular Data Source Errors

A data source which is expected to return Terraform State about a single remote resource is commonly referred to as a "singular" data source.
Implementation-wise, it may use any available describe or list functionality from the remote system to retrieve the information.
In addition to remote operation and data handling errors, errors should also be returned if:

- Zero results are found.
- Multiple results are found.

For remote operations that are designed to return an error when the remote resource is not found, this error is typically just passed through similar to other remote operation errors.
For remote operations that are designed to return a successful result whether there are zero, one, or multiple results, the error must be generated.

For example in pseudo-code:

```go
output, err := conn.ListServiceThings(input)

if err != nil {
    // Return error diagnostic wrapping remote error
}

if output == nil || len(output.Results) == 0 {
    // Return custom error diagnostic indicating empty results
}

if len(output.Results) > 1 {
    // Return custom error diagnostic indicating multiple results
}
```

#### Plural Data Source Errors

An emergent concept is a data source that returns multiple results, acting similarly to listing functionality from the remote system.
These types of data sources should return _not_ return errors if:

- Zero results are found.
- Multiple results are found.

Remote operation and other data handling errors should still be returned.

### Resource Update

For Terraform Plugin Framework based resources, update is implemented on the `Update` method of the resource struct.
For Terraform Plugin SDK V2 based resources, update is implemented on the `UpdateWithoutTimeout` function of the resource schema definition.

#### Update Error Message Context

=== "Terraform Plugin Framework (Preferred)"
    Errors ecnountered during update should include the resource identifier and additional messaging about the location or cause of the error for operators and code maintainers.
    Plugin Framework based resources will append error [diagnostics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics) to the `resource.UpdateResponse` `Diagnostics` field.
    The `create.ProblemStandardMessage` helper function can be used for constructing the appropriate context to accompany the error.

    ```go
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionUpdating, ResName{{ .Resource }}, plan.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

    e.g.

    ```go
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameVPCConnection, plan.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

    Resources that use other operations that return errors (e.g. waiters) should follow a similar pattern:

    ```go
    updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
    waitOut, err := waitVPCConnectionUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForUpdate, ResNameVPCConnection, plan.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

=== "Terraform Plugin SDK V2"
    Errors encountered during update should include the resource identifier and additional messaging about the location or cause of the error for operators and code maintainers.
    The `create.AppendDiagError` helper function can be used to convert a native error into a [diagnostic](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics), which is the method for surfacing warnings and errors in Terraform providers.

    ```go
    if err != nil {
        return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionUpdating, ResName{{ .Resource }}, d.Id(), err)
    }
    ```

    e.g.

    ```go
    if err != nil {
         return create.AppendDiagError(diags, names.IVS, create.ErrActionUpdating, ResNameRecordingConfiguration, d.Id(), err)
    }
    ```

    Resources that use other operations that return errors (e.g. waiters) should follow a similar pattern:

    ```go
    if _, err := waitRecordingConfigurationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
         return create.AppendDiagError(diags, names.IVS, create.ErrActionWaitingForUpdate, ResNameRecordingConfiguration, d.Id(), err)
    }
    ```

### Resource Deletion

For Terraform Plugin Framework based resources, deletion is implemented on the `Delete` method of the resource struct.
For Terraform Plugin SDK V2 based resources, deletion is implemented on the `DeleteWithoutTimeout` function of the resource schema definition.

#### Deletion Error Message Context

=== "Terraform Plugin Framework (Preferred)"
    Errors encountered during deletion should include the resource identifier and additional messaging about the location or cause of the error for operators and code maintainers.
    Plugin Framework based resources will append error [diagnostics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics) to the `resource.DeleteResponse` `Diagnostics` field.
    The `create.ProblemStandardMessage` helper function can be used for constructing the appropriate context to accompany the error.

    ```go
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionDeleting, ResName{{ .Resource }}, state.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

    e.g.

    ```go
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameVPCConnection, state.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

    Resources that use other operations that return errors (e.g. waiters) should follow a similar pattern:

    ```go
    deleteTimeout := r.DeleteTimeout(ctx, plan.Timeouts)
    waitOut, err := waitVPCConnectionDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForDeletion, ResNameVPCConnection, state.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

=== "Terraform Plugin SDK V2"
    Errors encountered during deletion should include the resource identifier and additional messaging about the location or cause of the error for operators and code maintainers.
    The `create.AppendDiagError` helper function can be used to convert a native error into a [diagnostic](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics), which is the method for surfacing warnings and errors in Terraform providers.

    ```go
    if err != nil {
        return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionDeleting, ResName{{ .Resource }}, d.Id(), err)
    }
    ```

    e.g.

    ```go
    if err != nil {
        return create.AppendDiagError(diags, names.IVS, create.ErrActionDeleting, ResNameRecordingConfiguration, d.Id(), err)
    }
    ```

    Resources that use other operations that return errors (e.g. waiters) should follow a similar pattern:

    ```go
    if _, err := waitRecordingConfigurationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
        return create.AppendDiagError(diags, names.IVS, create.ErrActionWaitingForDeletion, ResNameRecordingConfiguration, d.Id(), err)
    }
    ```

#### Resource Already Deleted

A typical pattern for resource deletion is to immediately perform the remote system deletion operation without checking existence.
This is generally acceptable as operators are encouraged to always refresh their Terraform State prior to performing changes.
However, in certain scenarios, such as external systems modifying the remote system prior to the Terraform execution, it is still possible that the remote system will return an error signifying that the resource does not exist.
In these cases, resources should implement logic that skips returning the error.

=== "Terraform Plugin Framework (Preferred)"
    !!! note
        The Terraform Plugin Framework automatically handles the equivalent of `resp.State.RemoveResource()` on deletion, so it is not necessary to include it.

    ```go
    output, err := conn.DeleteServiceThing(input)

    if tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
        return
    }

    if err != nil {
        resp.Diagnostics.AddError(
            create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionDeleting, ResName{{ .Resource }}, state.ID.String(), err),
            err.Error(),
        )
        return
    }
    ```

=== "Terraform Plugin SDK V2"
    !!! note
        The Terraform Plugin SDK V2 automatically handles the equivalent of `d.SetId("")` on deletion, so it is not necessary to include it.

    ```go
    output, err := conn.DeleteServiceThing(input)

    if tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
        return diags
    }

    if err != nil {
        return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionDeleting, ResName{{ .Resource }}, d.Id(), err)
    }
    ```
