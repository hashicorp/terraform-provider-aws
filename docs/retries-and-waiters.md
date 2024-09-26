<!-- markdownlint-configure-file { "code-block-style": false } -->
# Retries and Waiters

Terraform plugins may run into situations where calling the remote system after an operation may be necessary. These typically fall under three classes where:

- The request never reaches the remote system.
- The request reaches the remote system and responds that it cannot handle the request temporarily.
- The implementation of the remote system requires additional requests to ensure success.

This guide describes the behavior of the Terraform AWS Provider and provides code implementations that help ensure success in each of these situations.

!!! note
    The helper functions detailed below **are** compatible with Terraform Plugin Framework based resources (the required library for net-new resources).
    While these functions currently reside in the legacy Terraform Plugin SDK repository, they are not directly tied to functionality exclusive to this library, and likely will be moved to a standalone library or into the AWS provider itself in the future.

## Terraform Plugin SDK Functionality

The [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk/), which the AWS Provider uses, provides vital tools for handling consistency: the `retry.StateChangeConf{}` struct, and the retry function `retry.RetryContext()`.
We will discuss these throughout the rest of this guide.
Since they help keep the AWS Provider code consistent, we heavily prefer them over custom implementations.

This guide goes beyond the [Terraform Plugin SDK v2 documentation](https://www.terraform.io/plugin/sdkv2/resources/retries-and-customizable-timeouts) by providing additional context and emergent implementations specific to the Terraform AWS Provider.

### State Change Configuration and Functions

The [`retry.StateChangeConf` type](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry#StateChangeConf), along with its receiver methods [`WaitForState()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry#StateChangeConf.WaitForState) and [`WaitForStateContext()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry#StateChangeConf.WaitForStateContext) is a generic primitive for repeating operations in Terraform resource logic until desired value(s) are received. The "state change" in this case is generic to any value and not specific to the Terraform State. Among other functionality, it supports some of these desirable optional properties:

- Expecting specific value(s) while waiting for the target value(s) to be reached. Unexpected values are returned as an error which can be augmented with additional details.
- Expecting the target value(s) to be returned multiple times in succession.
- Allowing various polling configurations such as delaying the initial request and setting the time between polls.

### Retry Functions

The [`retry.RetryContext()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry#RetryContext) function provides a simplified retry implementation around `retry.StateChangeConf`.
The most common use is for simple error-based retries.

## AWS Request Handling

The Terraform AWS Provider's requests to AWS service APIs happen on top of Hypertext Transfer Protocol (HTTP). The following is a simplified description of the layers and handling that requests pass through:

- A Terraform resource calls an AWS Go SDK function.
- The AWS Go SDK generates an AWS-compatible HTTP request using the [Go standard library `net/http` package](https://pkg.go.dev/net/http/). This includes the following:
    - Adding HTTP headers for authentication and signing of requests to ensure authenticity.
    - Converting operation inputs into required HTTP URI parameters and/or request body type (XML or JSON).
    - If debug logging is enabled, logging of the HTTP request.
- The AWS Go SDK transmits the `net/http` request using Go's standard handling of the Operating System (OS) and Domain Name System (DNS) configuration.
- The AWS service potentially receives the request and responds, typically adding a request identifier HTTP header which can be used for AWS Support cases.
- The OS and Go `net/http` receive the response and pass it to the AWS Go SDK.
- The AWS Go SDK attempts to handle the response. This may include:
    - Parsing output
    - Converting errors into operation errors (Go `error` type of wrapped [`awserr.Error` type](https://docs.aws.amazon.com/sdk-for-go/api/aws/awserr/#Error)).
    - Converting response elements into operation outputs (AWS Go SDK operation-specific types).
    - Triggering automatic request retries based on default and custom logic.
- The Terraform resource receives the response, including any output and errors, from the AWS Go SDK.

In cases where custom operation handling is configured for a specific service client, the changes can typically be found in `internal/service/{service-name}/service_package.go`.

### Default AWS Go SDK Retries

In some situations, while handling a response, the AWS Go SDK automatically retries a request before returning the output and error. The retry mechanism implements an exponential backoff algorithm. The default conditions triggering automatic retries (implemented through [`client.DefaultRetryer`](https://docs.aws.amazon.com/sdk-for-go/api/aws/client/#DefaultRetryer)) include:

- Certain network errors. A common exception to this is connection reset errors.
- HTTP status codes 429 and 5xx.
- Certain API error codes, which are common across various AWS services (e.g., `ThrottledException`). However, not all AWS services implement these error codes consistently. A common exception to this is certain expired credentials errors.

By default, the Terraform AWS Provider sets the maximum number of AWS Go SDK retries based on the [`max_retries` provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#max_retries). The provider configuration defaults to 25 and the exponential backoff roughly equates to one hour of retries. This very high default value was present before the Terraform AWS Provider codebase was split from Terraform CLI in version 0.10.

### Lower Network Error Retries

Given the very high default number of AWS Go SDK retries configured in the Terraform AWS Provider and the excessive wait that practitioners would face, the [`hashicorp/aws-sdk-go-base` codebase](https://github.com/hashicorp/aws-sdk-go-base/blob/57529b4c2d2f8f3b5299d66a829b01259fa800d7/session.go#L108-L130) lowers retries to 10 for certain network errors that typically cannot be remediated via retries. This roughly equates to 30 seconds of retries.

### Terraform AWS Provider Service Retries

The AWS Go SDK provides hooks for injecting custom logic into service client handlers.
We prefer this handling in situations where contributors would need to apply the retry behavior to many resources.
For example, in cases where the AWS service API does not mark an error code as automatically retriable.
The AWS Provider includes other retry-changing behaviors using this method.
When custom service client configurations are applied, these will be defined in `internal/service/{service-name}/service_package.go`.

=== "AWS Go SDK V2 (Preferred)"
    With V2 of the AWS Go SDK, the retrier is extended directly in client construction.

    ```go
    // NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
    func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*s3_sdkv2.Client, error) {
    	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))
    
    	return s3_sdkv2.NewFromConfig(cfg,
			s3.WithEndpointResolverV2(newEndpointResolverSDKv2()),
			withBaseEndpoint(config[names.AttrEndpoint].(string)),
			func(o *s3_sdkv2.Options) {
				// ..other configuration..
		
				o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws_sdkv2.RetryerV2), retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
					if tfawserr_sdkv2.ErrMessageContains(err, errCodeOperationAborted, "A conflicting conditional operation is currently in progress against this resource. Please try again.") {
						return aws_sdkv2.TrueTernary
					}
					return aws_sdkv2.UnknownTernary // Delegate to configured Retryer.
				}))
			},
		), nil
    }
    ```

=== "AWS Go SDK V1"
    With V1 of the AWS Go SDK, a `CustomizeConn` helper is used to augment retry logic on an existing client after construction.

    ```go
    // CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
    func (p *servicePackage) CustomizeConn(ctx context.Context, conn *kafka_sdkv1.Kafka) (*kafka_sdkv1.Kafka, error) {
    	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
    		if tfawserr.ErrMessageContains(r.Error, kafka_sdkv1.ErrCodeTooManyRequestsException, "Too Many Requests") {
    			r.Retryable = aws_sdkv1.Bool(true)
    		}
    	})
    
    	return conn, nil
    }
    ```

## Eventual Consistency

Eventual consistency is a temporary condition where the remote system can return outdated information or errors due to not being strongly read-after-write consistent.
This is a pattern found in remote systems that must be highly scaled for broad usage.

Terraform expects any planned resource lifecycle change (create, update, destroy of the resource itself) and planned resource attribute value change to match after being applied.
Conversely, operators typically expect that Terraform resources also implement the concept of drift detection for resources and their attributes, which requires reading information back from the remote system after an operation.
A common implementation is calling the underlying `Get`/`Describe` AWS API after `Create` and `Update`.

These two concepts conflict with each other and require additional handling in Terraform resource logic as shown in the following sections.
These issues are _not_ reliably reproducible, especially in the case of writing acceptance testing, so they can be elusive with false positives to verify fixes.

### Operation Specific Error Retries

Even given a properly ordered Terraform configuration, eventual consistency can unexpectedly prevent downstream operations from succeeding.
A simple retry after a few seconds resolves many of these issues.
To reduce frustrating behavior for operators, wrap AWS Go SDK operations with the `retry.RetryContext()` function.
These retries should have a reasonably low timeout (typically two minutes but up to five minutes).
Save them in a constant for reusability.
These functions are preferably in line with the associated resource logic to remove any indirection with the code.

Do not use this type of logic to overcome improperly ordered Terraform configurations.
The approach may not work in larger environments.

```go
const (
	// Maximum amount of time to wait for Thing operation eventual consistency
	ThingOperationTimeout = 2 * time.Minute
)
```

=== "AWS Go SDK V2 (Preferred)"
    ```go
    // internal/service/{service}/{thing}.go

    // ... Create, Read, Update, or Delete function ...
    	err := retry.RetryContext(ctx, ThingOperationTimeout, func() *retry.RetryError {
    		_, err := conn./* ... AWS Go SDK operation with eventual consistency errors ... */

    		// Retryable conditions which can be checked.
    		// These must be updated to match the AWS service API error code and message.
            if errs.IsAErrorMessageContains[/* error type */](err, /* error message */) {
    			return retry.RetryableError(err)
    		}
    
    		if err != nil {
    			return retry.NonRetryableError(err)
    		}

    		return nil
    	})
    
    	// This check is important - it handles when the AWS Go SDK operation retries without returning.
    	// e.g., any automatic retries due to network or throttling errors.
    	if tfresource.TimedOut(err) {
    		// The use of equals assignment (over colon equals) is also important here.
    		// This overwrites the error variable to simplify logic.
    		_, err = conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */
    	}

    	if err != nil {
    		return fmt.Errorf("... error message context ... : %w", err)
    	}
    ```

=== "AWS Go SDK V1"
    ```go
    // internal/service/{service}/{thing}.go

    // ... Create, Read, Update, or Delete function ...
    	err := retry.RetryContext(ctx, ThingOperationTimeout, func() *retry.RetryError {
    		_, err := conn./* ... AWS Go SDK operation with eventual consistency errors ... */

    		// Retryable conditions which can be checked.
    		// These must be updated to match the AWS service API error code and message.
    		if tfawserr.ErrMessageContains(err, /* error code */, /* error message */) {
    			return retry.RetryableError(err)
    		}

    		if err != nil {
    			return retry.NonRetryableError(err)
    		}

    		return nil
    	})

    	// This check is important - it handles when the AWS Go SDK operation retries without returning.
    	// e.g., any automatic retries due to network or throttling errors.
    	if tfresource.TimedOut(err) {
    		// The use of equals assignment (over colon equals) is also important here.
    		// This overwrites the error variable to simplify logic.
    		_, err = conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */
    	}

    	if err != nil {
    		return fmt.Errorf("... error message context ... : %w", err)
    	}
    ```

#### IAM Error Retries

A common eventual consistency issue is an error returned due to IAM permissions. The IAM service itself is eventually consistent along with the propagation of its components and permissions to other AWS services. For example, if the following operations occur in quick succession:

- Create an IAM Role
- Attach an IAM Policy to the IAM Role
- Reference the new IAM Role in another AWS service, such as creating a Lambda Function

The last operation can receive varied API errors ranging from:

- IAM Role being reported as not existing
- IAM Role being reported as not having permissions for the other service to use it (assume role permissions)
- IAM Role being reported as not having sufficient permissions (inline or attached role permissions)

Each AWS service API (and sometimes even operations within the same API) varies in the implementation of these errors. To handle them, it is recommended to use the [Operation Specific Error Retries](#operation-specific-error-retries) pattern. The Terraform AWS Provider implements a standard timeout constant of two minutes in the `internal/service/iam` package which should be used for all retry timeouts associated with IAM errors. This timeout was derived from years of Terraform operational experience with all AWS APIs.

=== "AWS Go SDK V2 (Preferred)"
    ```go
    // internal/service/{service}/{thing}.go

    import (
    	// ... other imports ...
    	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
    )

    // ... Create and typically Update function ...
    	err := retry.RetryContext(ctx, iamwaiter.PropagationTimeout, func() *retry.RetryError {
    		_, err := conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */

    		// Example retryable condition
    		// This must be updated to match the AWS service API error code and message.
    		if errs.IsAErrorMessageContains[/* error type */](err, /* error message */) {
    			return retry.RetryableError(err)
    		}

    		if err != nil {
    			return retry.NonRetryableError(err)
    		}

    		return nil
    	})

    	if tfresource.TimedOut(err) {
    		_, err = conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */
    	}

    	if err != nil {
    		return fmt.Errorf("... error message context ... : %w", err)
    	}
    ```

=== "AWS Go SDK V1"
    ```go
    // internal/service/{service}/{thing}.go

    import (
    	// ... other imports ...
    	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
    )

    // ... Create and typically Update function ...
    	err := retry.RetryContext(ctx, iamwaiter.PropagationTimeout, func() *retry.RetryError {
    		_, err := conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */

    		// Example retryable condition
    		// This must be updated to match the AWS service API error code and message.
    		if tfawserr.ErrMessageContains(err, /* error code */, /* error message */) {
    			return retry.RetryableError(err)
    		}

    		if err != nil {
    			return retry.NonRetryableError(err)
    		}

    		return nil
    	})

    	if tfresource.TimedOut(err) {
    		_, err = conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */
    	}

    	if err != nil {
    		return fmt.Errorf("... error message context ... : %w", err)
    	}
    ```

#### Asynchronous Operation Error Retries

Some remote system operations run asynchronously as detailed in the [Asynchronous Operations section](#asynchronous-operations). In these cases, it is possible that the initial operation will immediately return as successful, but potentially return a retryable failure while checking the operation status that requires starting everything over. The handling for these is complicated by the fact that there are two timeouts, one for the retryable failure and one for the asynchronous operation status checking.

The below code example highlights this situation for a resource creation that also exhibited IAM eventual consistency.

```go
// internal/service/{service}/{thing}.go

import (
	// ... other imports ...
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

// ... Create function ...

	// Underlying IAM eventual consistency errors can occur after the creation
	// operation. The goal is only retry these types of errors up to the IAM
	// timeout. Since the creation process is asynchronous and can take up to
	// its own timeout, we store a stop time upfront for checking.
	iamwaiterStopTime := time.Now().Add(tfiam.PropagationTimeout)

	// Ensure to add IAM eventual consistency timeout in case of retries
	err = retry.RetryContext(ctx, tfiam.PropagationTimeout+ThingOperationTimeout, func() *retry.RetryError {
		// Only retry IAM eventual consistency errors up to that timeout
		iamwaiterRetry := time.Now().Before(iamwaiterStopTime)

		_, err := conn./* ... AWS Go SDK operation without eventual consistency errors ... */

		if err != nil {
			return retry.NonRetryableError(err)
		}

		_, err = ThingOperation(conn, d.Id())

		if err != nil {
			if iamwaiterRetry && /* eventual consistency error checking */ {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn./* ... AWS Go SDK operation without eventual consistency errors ... */

		if err != nil {
			return err
		}

		_, err = ThingOperation(conn, d.Id())

		if err != nil {
			return err
		}
	}
```

### Resource Lifecycle Retries

Resource lifecycle eventual consistency is a type of consistency issue that relates to the existence or state of an AWS infrastructure component.
For example, if you create a resource and immediately try to get information about it, some AWS services and operations will return a "not found" error.
Depending on the service and general AWS load, these errors can be frequent or rare.

In order to avoid this issue, identify operations that make changes.
Then, when calling any other operations that rely on the changes, account for the possibility that the AWS service has not yet fully realized them.

Handling eventual consistency looks slightly different for Terraform Plugin Framework based resources versus Plugin SDK V2. For Terraform Framework based resources, a `Get`/`Describe` API call should be made after the create request completes, all within the `Create` method. For Terraform Plugin SDK V2 resources, the `Create` function should return by calling the `Read` function. Both approaches fill in computed attributes and ensure that the AWS service applied the configuration correctly. Add retry logic to the `Get`/`Describe` API call (Plugin Framework) or `Read` function (Plugin SDK V2) to overcome the temporary condition on resource creation.

!!! note
    For eventually consistent resources, "not found" errors can still occur in the `Read` function even after implementing [Resource Lifecycle Waiters](#resource-lifecycle-waiters).

```go
const (
	// Maximum amount of time to wait for Thing eventual consistency on creation
	ThingCreationTimeout = 2 * time.Minute
)
```

=== "Terraform Plugin Framework (Preferred)"
    ```go
    // internal/service/{service}/{thing}.go

    func (r *resourceThing) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    	conn := meta.(*AWSClient).ExampleClient()

    	// ...Creation steps...

    	input := &example.OperationInput{/* ... */}

    	var output *example.OperationOutput
        createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
    	err := retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
    		var err error
    		output, err = conn.Operation(input)

    		if errs.IsA[*types.ResourceNotFoundException(err) {
    			return retry.RetryableError(err)
    		}

    		if err != nil {
    			return retry.NonRetryableError(err)
    		}

    		return nil
    	})

    	// Retry AWS Go SDK operation if no response from automatic retries.
    	if tfresource.TimedOut(err) {
    		output, err = exampleconn.Operation(input)
    	}

    	if err != nil {
            resp.Diagnostics.AddError(
                create.ProblemStandardMessage(names.Example, create.ErrActionWaitingForCreation, ResNameThing, plan.ID.String(), err),
                err.Error(),
            )
            return
    	}

    	// Prevent panics.
    	if output == nil {
            resp.Diagnostics.AddError(
                create.ProblemStandardMessage(names.Example, create.ErrActionWaitingForCreation, ResNameThing, plan.ID.String(), errors.New("empty output")),
                err.Error(),
            )
            return
    	}

    	// ... refresh Terraform state as normal ...
    }
    ```

=== "Terraform Plugin SDK V2"
    ```go
    // internal/service/{service}/{thing}.go

    function ExampleThingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		var diags diag.Diagnostics
    	// ...
    	return append(diags, ExampleThingRead(ctx, d, meta)...)
    }

    function ExampleThingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		var diags diag.Diagnostics

    	conn := meta.(*AWSClient).ExampleConn()

    	input := &example.OperationInput{/* ... */}

    	var output *example.OperationOutput
    	err := retry.RetryContext(ctx, ThingCreationTimeout, func() *retry.RetryError {
    		var err error
    		output, err = conn.Operation(input)

    		// Retry on any API "not found" errors, but only on new resources.
    		if d.IsNewResource() && tfawserr.ErrorCodeEquals(err, example.ErrCodeResourceNotFoundException) {
    			return retry.RetryableError(err)
    		}

    		if err != nil {
    			return retry.NonRetryableError(err)
    		}

    		return nil
    	})

    	// Retry AWS Go SDK operation if no response from automatic retries.
    	if tfresource.TimedOut(err) {
    		output, err = exampleconn.Operation(input)
    	}

    	// Prevent confusing Terraform error messaging to operators by
    	// Only ignoring API "not found" errors if not a new resource.
    	if !d.IsNewResource() && tfawserr.ErrorCodeEquals(err, example.ErrCodeNoSuchEntityException) {
    		log.Printf("[WARN] Example Thing (%s) not found, removing from state", d.Id())
    		d.SetId("")
    		return diags
    	}

    	if err != nil {
    		return sdkdiag.AppendErrorf(diags, "reading Example Thing (%s): %w", d.Id(), err)
    	}

    	// Prevent panics.
    	if output == nil {
    		return sdkdiag.AppendErrorf(diags, "reading Example Thing (%s): empty response", d.Id())
    	}

    	// ... refresh Terraform state as normal ...
    	d.Set("arn", output.Arn)
    }
    ```

    Some other general guidelines are:
    
    - If the `Create` function uses `retry.StateChangeConf`, the underlying `resource.RefreshStateFunc` should `return nil, "", nil` instead of the API "not found" error. This way the `StateChangeConf` logic will automatically retry.
    - If the `Create` function uses `retry.RetryContext()`, the API "not found" error should be caught and `return retry.RetryableError(err)` to automatically retry.
    
    In rare cases, it may be easier to duplicate all `Read` function logic in the `Create` function to handle all retries in one place.

### Resource Attribute Value Waiters

An emergent solution for handling eventual consistency with attribute values on updates is to introduce a custom `retry.StateChangeConf` and `resource.RefreshStateFunc` handlers. For example, the waiting logic can be implemented as:

=== "AWS SDK Go V2 (Preferred)"
    ```go
    // ThingAttribute fetches the Thing and its Attribute
    func ThingAttribute(ctx context.Context, conn *example.Client, id string) retry.StateRefreshFunc {
        return func() (interface{}, string, error) {
            output, err := /* ... AWS Go SDK operation to fetch resource/value ... */

    		if errs.IsA[*types.ResourceNotFoundException](err) {
    			return nil, "", nil
    		}
    
    		if err != nil {
    			return nil, "", err
    		}
    
    		if output == nil {
    			return nil, "", nil
    		}
    
    		return output, aws.ToString(output.Attribute), nil
    	}
    }
    
    // waitThingAttributeUpdated is an attribute waiter for Thing.Attribute
    func waitThingAttributeUpdated(ctx context.Context, conn *example.Example, id string, expectedValue string, timeout time.Duration) (*example.Thing, error) {
    	stateConf := &retry.StateChangeConf{
    		Target:  []string{expectedValue},
    		Refresh: ThingAttribute(ctx, conn, id),
    		Timeout: timeout,
    	}
    
    	outputRaw, err := stateConf.WaitForState()
    
    	if output, ok := outputRaw.(*example.Thing); ok {
    		return output, err
    	}
    
    	return nil, err
    }
    ```

=== "AWS SDK Go V1"
    ```go
    // ThingAttribute fetches the Thing and its Attribute
    func ThingAttribute(ctx context.Context, conn *example.Example, id string) retry.StateRefreshFunc {
        return func() (interface{}, string, error) {
            output, err := /* ... AWS Go SDK operation to fetch resource/value ... */

    		if tfawserr.ErrCodeEquals(err, example.ErrCodeResourceNotFoundException) {
    			return nil, "", nil
    		}

    		if err != nil {
    			return nil, "", err
    		}

    		if output == nil {
    			return nil, "", nil
    		}

    		return output, aws.StringValue(output.Attribute), nil
    	}
    }
    
    // waitThingAttributeUpdated is an attribute waiter for Thing.Attribute
    func waitThingAttributeUpdated(ctx context.Context, conn *example.Example, id string, expectedValue string, timeout time.Duration) (*example.Thing, error) {
    	stateConf := &retry.StateChangeConf{
    		Target:  []string{expectedValue},
    		Refresh: ThingAttribute(ctx, conn, id),
    		Timeout: timeout,
    	}

    	outputRaw, err := stateConf.WaitForState()

    	if output, ok := outputRaw.(*example.Thing); ok {
    		return output, err
    	}

    	return nil, err
    }
    ```

And consumed within the resource update workflow as follows:

=== "Terraform Plugin Framework (Preferred)"
    ```go
    func (r *resourceThing) Update(ctx context.Context, req resource.UpdateRequest, resp resource.UpdateResponse) {
        // ...

    	!plan.Attribute.Equal(state.Attribute) {
    		// ... AWS Go SDK logic to update attribute ...

            updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
    		if _, err := waitThingAttributeUpdated(ctx, conn, plan.ID.StringValue(), plan.Attribute.StringValue(), updateTimeout); err != nil {
                resp.Diagnostics.AddError(
                    create.ProblemStandardMessage(names.Example, create.ErrActionWaitingForUpdate, ResNameThing, plan.ID.String(), err),
                    err.Error(),
                )
            return
    		}
    	}

    	// ...
    }
    ```

=== "Terraform Plugin SDK V2"
    ```go
    function resourceThingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diags.Diagnostics {
        // ...

    	d.HasChange("attribute") {
    		// ... AWS Go SDK logic to update attribute ...

    		if _, err := waitThingAttributeUpdated(ctx, conn, d.Id(), d.Get("attribute").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
                return create.AppendDiagError(diags, names.Example, create.ErrActionWaitingForUpdate, ResNameThing, d.Id(), err)
    		}
    	}

    	// ...
    }
    ```

## Asynchronous Operations

When you initiate a long-running operation, an AWS service may return a successful response immediately and continue working on the request asynchronously. A resource can track the status with a component-level field (e.g., `CREATING`, `UPDATING`, etc.) or an explicit tracking identifier.

Terraform resources should wait for these background operations to complete. Failing to do so can introduce incomplete state information and downstream errors in other resources. In rare scenarios involving very long-running operations, operators may request a flag to skip the waiting. However, these should only be implemented case-by-case to prevent those previously mentioned confusing issues.

### AWS Go SDK Waiters

In limited cases, the AWS service API model includes the information to automatically generate a waiter function in the AWS Go SDK for an operation. These are typically named with the prefix `WaitUntil...`. If available, these functions can be used for an initial resource implementation. For example:

```go
if err := conn.WaitUntilEndpointInService(input); err != nil {
	return fmt.Errorf("waiting for Example Thing (%s) ...: %w", d.Id(), err)
}
```

If it is necessary to customize the timeouts and polling, we generally prefer using [Resource Lifecycle Waiters](#resource-lifecycle-waiters) instead since they are more commonly used throughout the codebase.

### Resource Lifecycle Waiters

Most of the codebase uses `retry.StateChangeConf` and `retry.StateRefreshFunc` handlers for tracking either component-level status fields or explicit tracking identifiers.
These should be placed in the `internal/service/{SERVICE}` package and split into separate functions. For example:

=== "AWS SDK Go V2 (Preferred)"
    ```go
    // ThingStatus fetches the Thing and its Status
    func ThingStatus(ctx context.Context, conn *example.Client, id string) retry.StateRefreshFunc {
        return func() (interface{}, string, error) {
            output, err := /* ... AWS Go SDK operation to fetch resource/status ... */

            if errs.IsA[*types.ResourceNotFoundException](err) {
    			return nil, "", nil
    		}

    		if err != nil {
    			return nil, "", err
    		}

    		if output == nil {
    			return nil, "", nil
    		}

    		return output, aws.ToString(output.Status), nil
    	}
    }
    ```

=== "AWS SDK Go V1"
    ```go
    // ThingStatus fetches the Thing and its Status
    func ThingStatus(ctx context.Context, conn *example.Example, id string) retry.StateRefreshFunc {
        return func() (interface{}, string, error) {
            output, err := /* ... AWS Go SDK operation to fetch resource/status ... */

    		if tfawserr.ErrCodeEquals(err, example.ErrCodeResourceNotFoundException) {
    			return nil, "", nil
    		}

    		if err != nil {
    			return nil, "", err
    		}

    		if output == nil {
    			return nil, "", nil
    		}

    		return output, aws.StringValue(output.Status), nil
    	}
    }
    ```

```go
// waitThingCreated is a resource waiter for Thing creation
func waitThingCreated(ctx context.Context, conn *example.Example, id string, timeout time.Duration) (*example.Thing, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{example.StatusCreating},
		Target:  []string{example.StatusCreated},
		Refresh: ThingStatus(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*example.Thing); ok {
		return output, err
	}

	return nil, err
}

// waitThingDeleted is a resource waiter for Thing deletion
func waitThingDeleted(ctx context.Context, conn *example.Example, id string, timeout time.Duration) (*example.Thing, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{example.StatusDeleting},
		Target:  []string{}, // Use empty list if the resource disappears and does not have "deleted" status
		Refresh: ThingStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*example.Thing); ok {
		return output, err
	}

	return nil, err
}
```

=== "Terraform Plugin Framework (Preferred)"
    ```go
    func (r *resourceThing) Create(ctx context.Context, req resource.CreateRequest, resp resource.CreateResponse) {
        // ... AWS Go SDK logic to create resource ...

        createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
        if _, err = waitThingCreated(ctx, conn, plan.ID.ValueString(), createTimeout); err != nil {
            resp.Diagnostics.AddError(
                create.ProblemStandardMessage(names.Example, create.ErrActionWaitingForCreation, ResNameThing, plan.ID.String(), err),
                err.Error(),
            )
            return
        }

    	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
    }

    func (r *resourceThing) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
        // ... AWS Go SDK logic to delete resource ...

        deleteTimeout := r.DeleteTimeout(ctx, plan.Timeouts)
        if _, err = waitThingDeleted(ctx, conn, plan.ID.ValueString(), deleteTimeout); err != nil {
            resp.Diagnostics.AddError(
                create.ProblemStandardMessage(names.Example, create.ErrActionWaitingForDeletion, ResNameThing, plan.ID.String(), err),
                err.Error(),
            )
            return
        }
    }
    ```

=== "Terraform Plugin SDK V2"
    ```go
    func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
        var diags diag.Diagnostics

        // ... AWS Go SDK logic to create resource ...

    	if _, err := waitThingCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)) {
            return create.AppendDiagError(diags, names.Example, create.ErrActionWaitingForCreation, ResNameThing, d.Id(), err)
    	}

    	return append(diags, ExampleThingRead(ctx, d, meta)...)
    }

    func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
        // ... AWS Go SDK logic to delete resource ...

    	if _, err := waitThingDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
            return create.AppendDiagError(diags, names.Example, create.ErrActionWaitingForDeletion, ResNameThing, d.Id(), err)
    	}

    	return diags
    }
    ```

Typically, the AWS Go SDK should include constants for various status field values (e.g., `StatusCreating` for `CREATING`). If not, create them in a file named `internal/service/{SERVICE}/consts.go`.
