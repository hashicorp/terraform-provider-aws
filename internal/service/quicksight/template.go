// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_template", name="Template")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.Template")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTemplateCreate,
		ReadWithoutTimeout:   resourceTemplateRead,
		UpdateWithoutTimeout: resourceTemplateUpdate,
		DeleteWithoutTimeout: resourceTemplateDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				names.AttrCreatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"definition": quicksightschema.TemplateDefinitionSchema(),
				names.AttrLastUpdatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 2048),
				},
				names.AttrPermissions: quicksightschema.PermissionsSchema(),
				"source_entity":       quicksightschema.TemplateSourceEntitySchema(),
				"source_entity_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"template_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"version_description": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 512),
				},
				"version_number": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			}
		},
	}
}

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	templateID := d.Get("template_id").(string)
	id := templateCreateResourceID(awsAccountID, templateID)
	input := &quicksight.CreateTemplateInput{
		AwsAccountId: aws.String(awsAccountID),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Tags:         getTagsIn(ctx),
		TemplateId:   aws.String(templateID),
	}

	if v, ok := d.GetOk("definition"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Definition = quicksightschema.ExpandTemplateDefinition(d.Get("definition").([]any))
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("source_entity"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.SourceEntity = quicksightschema.ExpandTemplateSourceEntity(v.([]any))
	}

	if v, ok := d.GetOk("version_description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	_, err := conn.CreateTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Template (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTemplateCreated(ctx, conn, awsAccountID, templateID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Template (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, templateID, err := templateParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	template, err := findTemplateByTwoPartKey(ctx, conn, awsAccountID, templateID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Template (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, template.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrCreatedTime, template.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, template.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, template.Name)
	d.Set(names.AttrStatus, template.Version.Status)
	d.Set("source_entity_arn", template.Version.SourceEntityArn)
	d.Set("template_id", template.TemplateId)
	d.Set("version_description", template.Version.Description)
	version := aws.ToInt64(template.Version.VersionNumber)
	d.Set("version_number", version)

	definition, err := findTemplateDefinitionByThreePartKey(ctx, conn, awsAccountID, templateID, version)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Template (%s) definition: %s", d.Id(), err)
	}

	if err := d.Set("definition", quicksightschema.FlattenTemplateDefinition(definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}

	permissions, err := findTemplatePermissionsByTwoPartKey(ctx, conn, awsAccountID, templateID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Template (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, templateID, err := templateParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		input := &quicksight.UpdateTemplateInput{
			AwsAccountId:       aws.String(awsAccountID),
			Name:               aws.String(d.Get(names.AttrName).(string)),
			TemplateId:         aws.String(templateID),
			VersionDescription: aws.String(d.Get("version_description").(string)),
		}

		// One of source_entity or definition is required for update
		if v, ok := d.GetOk("source_entity"); ok {
			input.SourceEntity = quicksightschema.ExpandTemplateSourceEntity(v.([]any))
		} else {
			input.Definition = quicksightschema.ExpandTemplateDefinition(d.Get("definition").([]any))
		}

		_, err := conn.UpdateTemplate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Template (%s): %s", d.Id(), err)
		}

		if _, err := waitTemplateUpdated(ctx, conn, awsAccountID, templateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Template (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		o, n := d.GetChange(names.AttrPermissions)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		toGrant, toRevoke := quicksightschema.DiffPermissions(os.List(), ns.List())

		input := &quicksight.UpdateTemplatePermissionsInput{
			AwsAccountId: aws.String(awsAccountID),
			TemplateId:   aws.String(templateID),
		}

		if len(toGrant) > 0 {
			input.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			input.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateTemplatePermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Template (%s) permissions: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, templateID, err := templateParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Template: %s", d.Id())
	_, err = conn.DeleteTemplate(ctx, &quicksight.DeleteTemplateInput{
		AwsAccountId: aws.String(awsAccountID),
		TemplateId:   aws.String(templateID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Template (%s): %s", d.Id(), err)
	}

	return diags
}

const templateResourceIDSeparator = ","

func templateCreateResourceID(awsAccountID, templateID string) string {
	parts := []string{awsAccountID, templateID}
	id := strings.Join(parts, templateResourceIDSeparator)

	return id
}

func templateParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, templateResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sTEMPLATE_ID", id, templateResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findTemplateByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, templateID string) (*awstypes.Template, error) {
	input := &quicksight.DescribeTemplateInput{
		AwsAccountId: aws.String(awsAccountID),
		TemplateId:   aws.String(templateID),
	}

	return findTemplate(ctx, conn, input)
}

func findTemplate(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeTemplateInput) (*awstypes.Template, error) {
	output, err := conn.DescribeTemplate(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Template == nil || output.Template.Version == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Template, nil
}

func findTemplateDefinitionByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, templateID string, version int64) (*awstypes.TemplateVersionDefinition, error) {
	input := &quicksight.DescribeTemplateDefinitionInput{
		AwsAccountId:  aws.String(awsAccountID),
		TemplateId:    aws.String(templateID),
		VersionNumber: aws.Int64(version),
	}

	return findTemplateDefinition(ctx, conn, input)
}

func findTemplateDefinition(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeTemplateDefinitionInput) (*awstypes.TemplateVersionDefinition, error) {
	output, err := conn.DescribeTemplateDefinition(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Definition == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Definition, nil
}

func findTemplatePermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, templateID string) ([]awstypes.ResourcePermission, error) {
	input := &quicksight.DescribeTemplatePermissionsInput{
		AwsAccountId: aws.String(awsAccountID),
		TemplateId:   aws.String(templateID),
	}

	return findTemplatePermissions(ctx, conn, input)
}

func findTemplatePermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeTemplatePermissionsInput) ([]awstypes.ResourcePermission, error) {
	output, err := conn.DescribeTemplatePermissions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Permissions, nil
}

func statusTemplate(ctx context.Context, conn *quicksight.Client, awsAccountID, templateID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTemplateByTwoPartKey(ctx, conn, awsAccountID, templateID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Version.Status), nil
	}
}

func waitTemplateCreated(ctx context.Context, conn *quicksight.Client, awsAccountID, templateID string, timeout time.Duration) (*awstypes.Template, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusTemplate(ctx, conn, awsAccountID, templateID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Template); ok {
		if status, apiErrors := output.Version.Status, output.Version.Errors; status == awstypes.ResourceStatusCreationFailed {
			tfresource.SetLastError(err, templateError(apiErrors))
		}

		return output, err
	}

	return nil, err
}

func waitTemplateUpdated(ctx context.Context, conn *quicksight.Client, awsAccountID, templateID string, timeout time.Duration) (*awstypes.Template, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusUpdateInProgress, awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusUpdateSuccessful, awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusTemplate(ctx, conn, awsAccountID, templateID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Template); ok {
		if status, apiErrors := output.Version.Status, output.Version.Errors; status == awstypes.ResourceStatusUpdateFailed {
			tfresource.SetLastError(err, templateError(apiErrors))
		}

		return output, err
	}

	return nil, err
}

func templateError(apiObjects []awstypes.TemplateError) error {
	errs := tfslices.ApplyToAll(apiObjects, func(v awstypes.TemplateError) error {
		return fmt.Errorf("%s: %s", v.Type, aws.ToString(v.Message))
	})

	return errors.Join(errs...)
}
