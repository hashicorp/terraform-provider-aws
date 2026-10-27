// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudcontrol

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	cfschema "github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudcontrolapi_resource", name="Resource")
func resourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceCreate,
		DeleteWithoutTimeout: resourceResourceDelete,
		ReadWithoutTimeout:   resourceResourceRead,
		UpdateWithoutTimeout: resourceResourceUpdate,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Hour),
			Delete: schema.DefaultTimeout(2 * time.Hour),
			Update: schema.DefaultTimeout(2 * time.Hour),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"desired_state": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrProperties: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrRoleARN: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrSchema: {
					Type:      schema.TypeString,
					Optional:  true,
					Computed:  true,
					Sensitive: true,
				},
				"type_name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z]{2,64}::[0-9A-Za-z]{2,64}::[0-9A-Za-z]{2,64}`), "must be three alphanumeric sections separated by double colons (::)"),
				},
				"type_version_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
			}
		},

		CustomizeDiff: customdiff.Sequence(
			resourceResourceCustomizeDiffGetSchema,
			resourceResourceCustomizeDiffSchemaDiff,
			customdiff.ComputedIf(names.AttrProperties, func(ctx context.Context, diff *schema.ResourceDiff, meta any) bool {
				return diff.HasChange("desired_state")
			}),
		),
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	typeName := d.Get("type_name").(string)
	input := cloudcontrol.CreateResourceInput{
		ClientToken:  aws.String(create.UniqueId(ctx)),
		DesiredState: aws.String(d.Get("desired_state").(string)),
		TypeName:     aws.String(typeName),
	}
	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("type_version_id"); ok {
		input.TypeVersionId = aws.String(v.(string))
	}

	output, err := conn.CreateResource(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cloud Control API (%s) Resource: %s", typeName, err)
	}

	// Always try to capture the identifier before returning errors.
	d.SetId(aws.ToString(output.ProgressEvent.Identifier))

	requestToken := aws.ToString(output.ProgressEvent.RequestToken)
	var progressEvent *awstypes.ProgressEvent
	waiter := cloudcontrol.NewResourceRequestSuccessWaiter(conn, func(o *cloudcontrol.ResourceRequestSuccessWaiterOptions) {
		o.Retryable = retryGetResourceRequestStatus(&progressEvent)
	})
	if err := waiter.Wait(ctx, &cloudcontrol.GetResourceRequestStatusInput{RequestToken: aws.String(requestToken)}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cloud Control API (%s) Resource (%s) create: %s", typeName, d.Id(), err)
	}

	// Some resources do not set the identifier until after creation.
	if d.Id() == "" && progressEvent != nil {
		d.SetId(aws.ToString(progressEvent.Identifier))
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	typeName := d.Get("type_name").(string)
	resourceDescription, err := findResourceByFourPartKey(ctx, conn,
		d.Id(),
		typeName,
		d.Get("type_version_id").(string),
		d.Get(names.AttrRoleARN).(string),
	)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Cloud Control API Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cloud Control API (%s) Resource (%s): %s", typeName, d.Id(), err)
	}

	d.Set(names.AttrProperties, resourceDescription.Properties)

	return diags
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	if d.HasChange("desired_state") {
		oldRaw, newRaw := d.GetChange("desired_state")
		patchDocument, err := patchDocument(oldRaw.(string), newRaw.(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating JSON Patch: %s", err)
		}

		typeName := d.Get("type_name").(string)
		input := cloudcontrol.UpdateResourceInput{
			ClientToken:   aws.String(create.UniqueId(ctx)),
			Identifier:    aws.String(d.Id()),
			PatchDocument: aws.String(patchDocument),
			TypeName:      aws.String(typeName),
		}
		if v, ok := d.GetOk(names.AttrRoleARN); ok {
			input.RoleArn = aws.String(v.(string))
		}
		if v, ok := d.GetOk("type_version_id"); ok {
			input.TypeVersionId = aws.String(v.(string))
		}

		output, err := conn.UpdateResource(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Cloud Control API (%s) Resource (%s): %s", typeName, d.Id(), err)
		}

		requestToken := aws.ToString(output.ProgressEvent.RequestToken)
		waiter := cloudcontrol.NewResourceRequestSuccessWaiter(conn, func(o *cloudcontrol.ResourceRequestSuccessWaiterOptions) {
			o.Retryable = retryGetResourceRequestStatus(nil)
		})
		if err := waiter.Wait(ctx, &cloudcontrol.GetResourceRequestStatusInput{RequestToken: aws.String(requestToken)}, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Cloud Control API (%s) Resource (%s) update: %s", typeName, d.Id(), err)
		}
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	typeName := d.Get("type_name").(string)
	input := cloudcontrol.DeleteResourceInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
		Identifier:  aws.String(d.Id()),
		TypeName:    aws.String(typeName),
	}
	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("type_version_id"); ok {
		input.TypeVersionId = aws.String(v.(string))
	}

	log.Printf("[INFO] Deleting Cloud Control API (%s) Resource: %s", typeName, d.Id())
	output, err := conn.DeleteResource(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cloud Control API (%s) Resource (%s): %s", typeName, d.Id(), err)
	}

	requestToken := aws.ToString(output.ProgressEvent.RequestToken)
	waiter := cloudcontrol.NewResourceRequestSuccessWaiter(conn, func(o *cloudcontrol.ResourceRequestSuccessWaiterOptions) {
		o.Retryable = retryGetResourceRequestStatus(nil)
	})
	if err := waiter.Wait(ctx, &cloudcontrol.GetResourceRequestStatusInput{RequestToken: aws.String(requestToken)}, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cloud Control API (%s) Resource (%s) delete: %s", typeName, d.Id(), err)
	}

	return diags
}

func resourceResourceCustomizeDiffGetSchema(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	resourceSchema := diff.Get(names.AttrSchema).(string)
	if resourceSchema != "" {
		return nil
	}

	typeName := diff.Get("type_name").(string)
	output, err := tfcloudformation.FindTypeByName(ctx, conn, typeName)

	if err != nil {
		return fmt.Errorf("reading CloudFormation Type (%s): %w", typeName, err)
	}

	if err := diff.SetNew(names.AttrSchema, output.Schema); err != nil {
		return fmt.Errorf("setting schema New: %w", err)
	}

	return nil
}

func resourceResourceCustomizeDiffSchemaDiff(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
	oldDesiredStateRaw, newDesiredStateRaw := diff.GetChange("desired_state")
	newSchema := diff.Get(names.AttrSchema).(string)

	newDesiredState, ok := newDesiredStateRaw.(string)
	if !ok {
		return fmt.Errorf("unexpected new desired_state value type: %T", newDesiredStateRaw)
	}

	// desired_state can be empty if unknown
	if newDesiredState == "" {
		return nil
	}

	newSchema, err := cfschema.Sanitize(newSchema)

	if err != nil {
		return fmt.Errorf("sanitizing CloudFormation Resource Schema JSON: %w", err)
	}

	cfResourceSchema, err := cfschema.NewResourceJsonSchemaDocument(newSchema)

	if err != nil {
		return fmt.Errorf("parsing CloudFormation Resource Schema JSON: %w", err)
	}

	if err := cfResourceSchema.ValidateConfigurationDocument(newDesiredState); err != nil {
		return fmt.Errorf("validating desired_state against CloudFormation Resource Schema: %w", err)
	}

	// Do nothing further for new resources or if desired state is not changed
	if diff.Id() == "" || !diff.HasChange("desired_state") {
		return nil
	}

	cfResource, err := cfResourceSchema.Resource()

	if err != nil {
		return fmt.Errorf("converting CloudFormation Resource Schema JSON: %w", err)
	}

	patches, err := tfjson.CreatePatchFromStrings(oldDesiredStateRaw.(string), newDesiredState)

	if err != nil {
		return fmt.Errorf("creating desired_state JSON Patch: %w", err)
	}

	for _, patch := range patches {
		if cfResource.IsCreateOnlyPropertyPath(patch.Path) {
			if err := diff.ForceNew("desired_state"); err != nil {
				return fmt.Errorf("setting desired_state ForceNew: %w", err)
			}

			break
		}
	}

	return nil
}

func findResourceByFourPartKey(ctx context.Context, conn *cloudcontrol.Client, resourceID, typeName, typeVersionID, roleARN string) (*awstypes.ResourceDescription, error) {
	input := cloudcontrol.GetResourceInput{
		Identifier: aws.String(resourceID),
		TypeName:   aws.String(typeName),
	}
	if roleARN != "" {
		input.RoleArn = aws.String(roleARN)
	}
	if typeVersionID != "" {
		input.TypeVersionId = aws.String(typeVersionID)
	}

	return findResource(ctx, conn, &input)
}

func findResource(ctx context.Context, conn *cloudcontrol.Client, input *cloudcontrol.GetResourceInput) (*awstypes.ResourceDescription, error) {
	output, err := conn.GetResource(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	// Some CloudFormation Resources do not correctly re-map "not found" errors, instead returning a HandlerFailureException.
	// These should be reported and fixed upstream over time, but for now work around the issue.
	if errs.Contains(err, "not found") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResourceDescription == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ResourceDescription, nil
}

// retryGetResourceRequestStatus returns a custom retryable function for the GetResourceRequestStatus operation.
func retryGetResourceRequestStatus(pProgressEvent **awstypes.ProgressEvent) func(context.Context, *cloudcontrol.GetResourceRequestStatusInput, *cloudcontrol.GetResourceRequestStatusOutput, error) (bool, error) {
	return func(ctx context.Context, input *cloudcontrol.GetResourceRequestStatusInput, output *cloudcontrol.GetResourceRequestStatusOutput, err error) (bool, error) {
		if err == nil {
			progressEvent := output.ProgressEvent
			if pProgressEvent != nil {
				*pProgressEvent = progressEvent
			}

			switch value := progressEvent.OperationStatus; value {
			case awstypes.OperationStatusSuccess:
				return false, nil

			case awstypes.OperationStatusCancelComplete:
				return false, waiterError(string(value), aws.ToString(progressEvent.StatusMessage))

			case awstypes.OperationStatusFailed:
				if progressEvent.ErrorCode == awstypes.HandlerErrorCodeNotFound && progressEvent.Operation == awstypes.OperationDelete {
					// Resource not found error on delete is OK.
					return false, nil
				}

				// Retry on throttling errors
				if progressEvent.ErrorCode == awstypes.HandlerErrorCodeThrottling {
					return true, err
				}

				// Build enhanced error message with hook information
				waiterErr := waiterError(string(value), aws.ToString(progressEvent.StatusMessage))
				waiterErr.withErrorCode(string(progressEvent.ErrorCode))

				// Add hook information if available
				if len(output.HooksProgressEvent) > 0 {
					for _, hookEvent := range output.HooksProgressEvent {
						hookStatus := aws.ToString(hookEvent.HookStatus)
						// HOOK_COMPLETE_FAILED: The Hook invocation is complete with a failed result.
						// HOOK_FAILED: The Hook invocation didn't complete successfully.
						if hookStatus == "HOOK_COMPLETE_FAILED" || hookStatus == "HOOK_FAILED" {
							waiterErr.withHookEvent(hookEvent)
						}
					}
				}

				return false, waiterErr
			}
		}

		return true, err
	}
}

type waiterResultError struct {
	errorCode  string
	errorMsg   string
	hookErrors []string
}

func (e *waiterResultError) Error() string {
	if e.errorCode != "" {
		e.errorMsg += fmt.Sprintf(". ErrorCode: %s", e.errorCode)
	}

	if len(e.hookErrors) > 0 {
		e.errorMsg += ". Hook failures: " + strings.Join(e.hookErrors, "; ")
	}

	return e.errorMsg
}

func waiterError(status, msg string) *waiterResultError {
	e := waiterResultError{}
	e.errorMsg = fmt.Sprintf("waiter state transitioned to %s. StatusMessage: %s", status, msg)

	return &e
}

func (e *waiterResultError) withErrorCode(code string) {
	e.errorCode = code
}

func (e *waiterResultError) withHookEvent(hookEvent awstypes.HookProgressEvent) {
	e.hookErrors = append(
		e.hookErrors,
		fmt.Sprintf("HookName: %s, HookArn: %s, HookVersion: %s, Time: %s, HookMessage: %s",
			aws.ToString(hookEvent.HookTypeName),
			aws.ToString(hookEvent.HookTypeArn),
			aws.ToString(hookEvent.HookTypeVersionId),
			hookEvent.HookEventTime.Format(time.RFC3339),
			aws.ToString(hookEvent.HookStatusMessage),
		),
	)
}

// patchDocument returns a JSON Patch document describing the difference between `old` and `new`.
func patchDocument(old, new string) (string, error) {
	patch, err := tfjson.CreatePatchFromStrings(old, new)

	if err != nil {
		return "", err
	}

	return tfjson.EncodeToString(patch)
}
