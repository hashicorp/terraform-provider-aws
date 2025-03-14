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

// @SDKResource("aws_quicksight_analysis", name="Analysis")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.Analysis")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func resourceAnalysis() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAnalysisCreate,
		ReadWithoutTimeout:   resourceAnalysisRead,
		UpdateWithoutTimeout: resourceAnalysisUpdate,
		DeleteWithoutTimeout: resourceAnalysisDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set("recovery_window_in_days", 30) //nolint:mnd // 30days is the default value (see below)
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"analysis_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
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
				"definition": quicksightschema.AnalysisDefinitionSchema(),
				"last_published_time": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrLastUpdatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 2048),
				},
				names.AttrParameters:  quicksightschema.ParametersSchema(),
				names.AttrPermissions: quicksightschema.PermissionsSchema(),
				"recovery_window_in_days": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  30,
					ValidateFunc: validation.Any(
						validation.IntBetween(7, 30),
						validation.IntInSlice([]int{0}),
					),
				},
				"source_entity": quicksightschema.AnalysisSourceEntitySchema(),
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"theme_arn": {
					Type:     schema.TypeString,
					Optional: true,
				},
			}
		},
	}
}

func resourceAnalysisCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	analysisID := d.Get("analysis_id").(string)
	id := analysisCreateResourceID(awsAccountID, analysisID)
	input := &quicksight.CreateAnalysisInput{
		AwsAccountId: aws.String(awsAccountID),
		AnalysisId:   aws.String(analysisID),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("definition"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Definition = quicksightschema.ExpandAnalysisDefinition(d.Get("definition").([]any))
	}

	if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Parameters = quicksightschema.ExpandParameters(d.Get(names.AttrParameters).([]any))
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("source_entity"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.SourceEntity = quicksightschema.ExpandAnalysisSourceEntity(v.([]any))
	}

	if v, ok := d.Get("theme_arn").(string); ok && v != "" {
		input.ThemeArn = aws.String(v)
	}

	_, err := conn.CreateAnalysis(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Analysis (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitAnalysisCreated(ctx, conn, awsAccountID, analysisID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Analysis (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAnalysisRead(ctx, d, meta)...)
}

func resourceAnalysisRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, analysisID, err := analysisParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	analysis, err := findAnalysisByTwoPartKey(ctx, conn, awsAccountID, analysisID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Analysis (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Analysis (%s): %s", d.Id(), err)
	}

	d.Set("analysis_id", analysis.AnalysisId)
	d.Set(names.AttrARN, analysis.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrCreatedTime, analysis.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, analysis.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, analysis.Name)
	d.Set(names.AttrStatus, analysis.Status)
	d.Set("theme_arn", analysis.ThemeArn)

	definition, err := findAnalysisDefinitionByTwoPartKey(ctx, conn, awsAccountID, analysisID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Analysis (%s) definition: %s", d.Id(), err)
	}

	if err := d.Set("definition", quicksightschema.FlattenAnalysisDefinition(definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}

	permissions, err := findAnalysisPermissionsByTwoPartKey(ctx, conn, awsAccountID, analysisID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Analysis (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceAnalysisUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, analysisID, err := analysisParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		input := &quicksight.UpdateAnalysisInput{
			AnalysisId:   aws.String(analysisID),
			AwsAccountId: aws.String(awsAccountID),
			Name:         aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk("source_entity"); ok {
			input.SourceEntity = quicksightschema.ExpandAnalysisSourceEntity(v.([]any))
		} else {
			input.Definition = quicksightschema.ExpandAnalysisDefinition(d.Get("definition").([]any))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.Parameters = quicksightschema.ExpandParameters(d.Get(names.AttrParameters).([]any))
		}

		if v, ok := d.Get("theme_arn").(string); ok && v != "" {
			input.ThemeArn = aws.String(v)
		}

		_, err := conn.UpdateAnalysis(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Analysis (%s): %s", d.Id(), err)
		}

		if _, err := waitAnalysisUpdated(ctx, conn, awsAccountID, analysisID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Analysis (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		o, n := d.GetChange(names.AttrPermissions)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		toGrant, toRevoke := quicksightschema.DiffPermissions(os.List(), ns.List())

		input := &quicksight.UpdateAnalysisPermissionsInput{
			AnalysisId:   aws.String(analysisID),
			AwsAccountId: aws.String(awsAccountID),
		}

		if len(toGrant) > 0 {
			input.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			input.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateAnalysisPermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Analysis (%s) permissions: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAnalysisRead(ctx, d, meta)...)
}

func resourceAnalysisDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, analysisID, err := analysisParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &quicksight.DeleteAnalysisInput{
		AnalysisId:   aws.String(analysisID),
		AwsAccountId: aws.String(awsAccountID),
	}

	if v := d.Get("recovery_window_in_days").(int); v == 0 {
		input.ForceDeleteWithoutRecovery = true
	} else {
		input.RecoveryWindowInDays = aws.Int64(int64(v))
	}

	log.Printf("[INFO] Deleting QuickSight Analysis: %s", d.Id())
	_, err = conn.DeleteAnalysis(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Analysis (%s): %s", d.Id(), err)
	}

	return diags
}

const analysisResourceIDSeparator = ","

func analysisCreateResourceID(awsAccountID, analysisID string) string {
	parts := []string{awsAccountID, analysisID}
	id := strings.Join(parts, analysisResourceIDSeparator)

	return id
}

func analysisParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, analysisResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sANALYSIS_ID", id, analysisResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findAnalysisByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, analysisID string) (*awstypes.Analysis, error) {
	input := &quicksight.DescribeAnalysisInput{
		AnalysisId:   aws.String(analysisID),
		AwsAccountId: aws.String(awsAccountID),
	}

	output, err := findAnalysis(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.ResourceStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findAnalysis(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeAnalysisInput) (*awstypes.Analysis, error) {
	output, err := conn.DescribeAnalysis(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Analysis == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Analysis, nil
}

func findAnalysisDefinitionByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, analysisID string) (*awstypes.AnalysisDefinition, error) {
	input := &quicksight.DescribeAnalysisDefinitionInput{
		AnalysisId:   aws.String(analysisID),
		AwsAccountId: aws.String(awsAccountID),
	}

	return findAnalysisDefinition(ctx, conn, input)
}

func findAnalysisDefinition(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeAnalysisDefinitionInput) (*awstypes.AnalysisDefinition, error) {
	output, err := conn.DescribeAnalysisDefinition(ctx, input)

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

func findAnalysisPermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, analysisID string) ([]awstypes.ResourcePermission, error) {
	input := &quicksight.DescribeAnalysisPermissionsInput{
		AnalysisId:   aws.String(analysisID),
		AwsAccountId: aws.String(awsAccountID),
	}

	return findAnalysisPermissions(ctx, conn, input)
}

func findAnalysisPermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeAnalysisPermissionsInput) ([]awstypes.ResourcePermission, error) {
	output, err := conn.DescribeAnalysisPermissions(ctx, input)

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

func statusAnalysis(ctx context.Context, conn *quicksight.Client, awsAccountID, analysisID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findAnalysisByTwoPartKey(ctx, conn, awsAccountID, analysisID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitAnalysisCreated(ctx context.Context, conn *quicksight.Client, awsAccountID, analysisID string, timeout time.Duration) (*awstypes.Analysis, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusAnalysis(ctx, conn, awsAccountID, analysisID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Analysis); ok {
		if status, apiErrors := output.Status, output.Errors; status == awstypes.ResourceStatusCreationFailed {
			tfresource.SetLastError(err, analysisError(apiErrors))
		}

		return output, err
	}

	return nil, err
}

func waitAnalysisUpdated(ctx context.Context, conn *quicksight.Client, awsAccountID, analysisID string, timeout time.Duration) (*awstypes.Analysis, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusUpdateInProgress, awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusUpdateSuccessful, awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusAnalysis(ctx, conn, awsAccountID, analysisID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Analysis); ok {
		if status, apiErrors := output.Status, output.Errors; status == awstypes.ResourceStatusUpdateFailed {
			tfresource.SetLastError(err, analysisError(apiErrors))
		}

		return output, err
	}

	return nil, err
}

func analysisError(apiObjects []awstypes.AnalysisError) error {
	errs := tfslices.ApplyToAll(apiObjects, func(v awstypes.AnalysisError) error {
		return fmt.Errorf("%s: %s", v.Type, aws.ToString(v.Message))
	})

	return errors.Join(errs...)
}
