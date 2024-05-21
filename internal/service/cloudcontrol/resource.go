// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	cfschema "github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/mattbaird/jsonpatch"
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

		Schema: map[string]*schema.Schema{
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
		},

		CustomizeDiff: customdiff.Sequence(
			resourceResourceCustomizeDiffGetSchema,
			resourceResourceCustomizeDiffSchemaDiff,
			customdiff.ComputedIf(names.AttrProperties, func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("desired_state")
			}),
		),
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	typeName := d.Get("type_name").(string)
	input := &cloudcontrol.CreateResourceInput{
		ClientToken:  aws.String(id.UniqueId()),
		DesiredState: aws.String(d.Get("desired_state").(string)),
		TypeName:     aws.String(typeName),
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_version_id"); ok {
		input.TypeVersionId = aws.String(v.(string))
	}

	output, err := conn.CreateResource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cloud Control API (%s) Resource: %s", typeName, err)
	}

	// Always try to capture the identifier before returning errors.
	d.SetId(aws.ToString(output.ProgressEvent.Identifier))

	output.ProgressEvent, err = waitProgressEventOperationStatusSuccess(ctx, conn, aws.ToString(output.ProgressEvent.RequestToken), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cloud Control API (%s) Resource (%s) create: %s", typeName, d.Id(), err)
	}

	// Some resources do not set the identifier until after creation.
	if d.Id() == "" {
		d.SetId(aws.ToString(output.ProgressEvent.Identifier))
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	typeName := d.Get("type_name").(string)
	resourceDescription, err := findResource(ctx, conn,
		d.Id(),
		typeName,
		d.Get("type_version_id").(string),
		d.Get(names.AttrRoleARN).(string),
	)

	if !d.IsNewResource() && tfresource.NotFound(err) {
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

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	if d.HasChange("desired_state") {
		oldRaw, newRaw := d.GetChange("desired_state")

		patchDocument, err := patchDocument(oldRaw.(string), newRaw.(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating JSON Patch: %s", err)
		}

		typeName := d.Get("type_name").(string)
		input := &cloudcontrol.UpdateResourceInput{
			ClientToken:   aws.String(id.UniqueId()),
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

		output, err := conn.UpdateResource(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Cloud Control API (%s) Resource (%s): %s", typeName, d.Id(), err)
		}

		if _, err := waitProgressEventOperationStatusSuccess(ctx, conn, aws.ToString(output.ProgressEvent.RequestToken), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Cloud Control API (%s) Resource (%s) update: %s", typeName, d.Id(), err)
		}
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CloudControlClient(ctx)

	typeName := d.Get("type_name").(string)
	input := &cloudcontrol.DeleteResourceInput{
		ClientToken: aws.String(id.UniqueId()),
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
	output, err := conn.DeleteResource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cloud Control API (%s) Resource (%s): %s", typeName, d.Id(), err)
	}

	progressEvent, err := waitProgressEventOperationStatusSuccess(ctx, conn, aws.ToString(output.ProgressEvent.RequestToken), d.Timeout(schema.TimeoutDelete))

	if progressEvent != nil && progressEvent.ErrorCode == types.HandlerErrorCodeNotFound {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cloud Control API (%s) Resource (%s) delete: %s", typeName, d.Id(), err)
	}

	return diags
}

func resourceResourceCustomizeDiffGetSchema(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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

func resourceResourceCustomizeDiffSchemaDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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

	patches, err := jsonpatch.CreatePatch([]byte(oldDesiredStateRaw.(string)), []byte(newDesiredStateRaw.(string)))

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

func findResource(ctx context.Context, conn *cloudcontrol.Client, resourceID, typeName, typeVersionID, roleARN string) (*types.ResourceDescription, error) {
	input := &cloudcontrol.GetResourceInput{
		Identifier: aws.String(resourceID),
		TypeName:   aws.String(typeName),
	}
	if roleARN != "" {
		input.RoleArn = aws.String(roleARN)
	}
	if typeVersionID != "" {
		input.TypeVersionId = aws.String(typeVersionID)
	}

	output, err := conn.GetResource(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	// Some CloudFormation Resources do not correctly re-map "not found" errors, instead returning a HandlerFailureException.
	// These should be reported and fixed upstream over time, but for now work around the issue.
	if errs.Contains(err, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResourceDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResourceDescription, nil
}

func findProgressEventByRequestToken(ctx context.Context, conn *cloudcontrol.Client, requestToken string) (*types.ProgressEvent, error) {
	input := &cloudcontrol.GetResourceRequestStatusInput{
		RequestToken: aws.String(requestToken),
	}

	output, err := conn.GetResourceRequestStatus(ctx, input)

	if errs.IsA[*types.RequestTokenNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ProgressEvent == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ProgressEvent, nil
}

func statusProgressEventOperation(ctx context.Context, conn *cloudcontrol.Client, requestToken string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findProgressEventByRequestToken(ctx, conn, requestToken)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.OperationStatus), nil
	}
}

func waitProgressEventOperationStatusSuccess(ctx context.Context, conn *cloudcontrol.Client, requestToken string, timeout time.Duration) (*types.ProgressEvent, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.OperationStatusInProgress, types.OperationStatusPending),
		Target:  enum.Slice(types.OperationStatusSuccess),
		Refresh: statusProgressEventOperation(ctx, conn, requestToken),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ProgressEvent); ok {
		if output.OperationStatus == types.OperationStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", output.ErrorCode, aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

// patchDocument returns a JSON Patch document describing the difference between `old` and `new`.
func patchDocument(old, new string) (string, error) {
	patch, err := jsonpatch.CreatePatch([]byte(old), []byte(new))

	if err != nil {
		return "", err
	}

	b, err := json.Marshal(patch)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
