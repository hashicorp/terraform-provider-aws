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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_dashboard", name="Dashboard")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.Dashboard")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func resourceDashboard() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDashboardCreate,
		ReadWithoutTimeout:   resourceDashboardRead,
		UpdateWithoutTimeout: resourceDashboardUpdate,
		DeleteWithoutTimeout: resourceDashboardDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
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
				"dashboard_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"dashboard_publish_options": quicksightschema.DashboardPublishOptionsSchema(),
				"definition":                quicksightschema.DashboardDefinitionSchema(),
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
				"source_entity":       quicksightschema.DashboardSourceEntitySchema(),
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
				"theme_arn": {
					Type:     schema.TypeString,
					Optional: true,
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

const (
	dashboardLatestVersion int64 = -1
)

func resourceDashboardCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	dashboardID := d.Get("dashboard_id").(string)
	id := dashboardCreateResourceID(awsAccountID, dashboardID)
	input := &quicksight.CreateDashboardInput{
		AwsAccountId: aws.String(awsAccountID),
		DashboardId:  aws.String(dashboardID),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("dashboard_publish_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DashboardPublishOptions = quicksightschema.ExpandDashboardPublishOptions(d.Get("dashboard_publish_options").([]any))
	}

	if v, ok := d.GetOk("definition"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Definition = quicksightschema.ExpandDashboardDefinition(d.Get("definition").([]any))
	}

	if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Parameters = quicksightschema.ExpandParameters(d.Get(names.AttrParameters).([]any))
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("source_entity"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.SourceEntity = quicksightschema.ExpandDashboardSourceEntity(v.([]any))
	}

	if v, ok := d.GetOk("version_description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	_, err := conn.CreateDashboard(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Dashboard (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDashboardCreated(ctx, conn, awsAccountID, dashboardID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Dashboard (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDashboardRead(ctx, d, meta)...)
}

func resourceDashboardRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dashboardID, err := dashboardParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dashboard, err := findDashboardByThreePartKey(ctx, conn, awsAccountID, dashboardID, dashboardLatestVersion)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Dashboard (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Dashboard (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dashboard.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrCreatedTime, dashboard.CreatedTime.Format(time.RFC3339))
	d.Set("dashboard_id", dashboard.DashboardId)
	d.Set(names.AttrLastUpdatedTime, dashboard.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, dashboard.Name)
	d.Set("source_entity_arn", dashboard.Version.SourceEntityArn)
	d.Set(names.AttrStatus, dashboard.Version.Status)
	d.Set("version_description", dashboard.Version.Description)
	version := aws.ToInt64(dashboard.Version.VersionNumber)
	d.Set("version_number", version)

	outputDDD, err := findDashboardDefinitionByThreePartKey(ctx, conn, awsAccountID, dashboardID, version)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Dashboard (%s) definition: %s", d.Id(), err)
	}

	if err := d.Set("definition", quicksightschema.FlattenDashboardDefinition(outputDDD.Definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}

	if err := d.Set("dashboard_publish_options", quicksightschema.FlattenDashboardPublishOptions(outputDDD.DashboardPublishOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dashboard_publish_options: %s", err)
	}

	permissions, err := findDashboardPermissionsByTwoPartKey(ctx, conn, awsAccountID, dashboardID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Dashboard (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dashboardID, err := dashboardParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		inputUD := &quicksight.UpdateDashboardInput{
			AwsAccountId:       aws.String(awsAccountID),
			DashboardId:        aws.String(dashboardID),
			Name:               aws.String(d.Get(names.AttrName).(string)),
			VersionDescription: aws.String(d.Get("version_description").(string)),
		}

		if v, ok := d.GetOk("dashboard_publish_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			inputUD.DashboardPublishOptions = quicksightschema.ExpandDashboardPublishOptions(d.Get("dashboard_publish_options").([]any))
		}

		if v, ok := d.GetOk("source_entity"); ok {
			inputUD.SourceEntity = quicksightschema.ExpandDashboardSourceEntity(v.([]any))
		} else {
			inputUD.Definition = quicksightschema.ExpandDashboardDefinition(d.Get("definition").([]any))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			inputUD.Parameters = quicksightschema.ExpandParameters(d.Get(names.AttrParameters).([]any))
		}

		output, err := conn.UpdateDashboard(ctx, inputUD)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Dashboard (%s): %s", d.Id(), err)
		}

		updatedVersionNumber := versionFromDashboardARN(aws.ToString(output.VersionArn))

		if _, err := waitDashboardUpdated(ctx, conn, awsAccountID, dashboardID, updatedVersionNumber, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Dashboard (%s) update: %s", d.Id(), err)
		}

		inputUDPV := &quicksight.UpdateDashboardPublishedVersionInput{
			AwsAccountId:  aws.String(awsAccountID),
			DashboardId:   aws.String(dashboardID),
			VersionNumber: aws.Int64(updatedVersionNumber),
		}

		_, err = conn.UpdateDashboardPublishedVersion(ctx, inputUDPV)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Dashboard (%s) published version: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		o, n := d.GetChange(names.AttrPermissions)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		toGrant, toRevoke := quicksightschema.DiffPermissions(os.List(), ns.List())

		input := &quicksight.UpdateDashboardPermissionsInput{
			AwsAccountId: aws.String(awsAccountID),
			DashboardId:  aws.String(dashboardID),
		}

		if len(toGrant) > 0 {
			input.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			input.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateDashboardPermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Dashboard (%s) permissions: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDashboardRead(ctx, d, meta)...)
}

func resourceDashboardDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dashboardID, err := dashboardParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Dashboard: %s", d.Id())
	_, err = conn.DeleteDashboard(ctx, &quicksight.DeleteDashboardInput{
		AwsAccountId: aws.String(awsAccountID),
		DashboardId:  aws.String(dashboardID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Dashboard (%s): %s", d.Id(), err)
	}

	return diags
}

const dashboardResourceIDSeparator = ","

func dashboardCreateResourceID(awsAccountID, dashboardID string) string {
	parts := []string{awsAccountID, dashboardID}
	id := strings.Join(parts, dashboardResourceIDSeparator)

	return id
}

func dashboardParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, dashboardResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sDASHBOARD_ID", id, dashboardResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func versionFromDashboardARN(arn string) int64 {
	return flex.StringValueToInt64Value(arn[strings.LastIndex(arn, "/")+1:])
}

func findDashboardByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dashboardID string, version int64) (*awstypes.Dashboard, error) {
	input := &quicksight.DescribeDashboardInput{
		AwsAccountId: aws.String(awsAccountID),
		DashboardId:  aws.String(dashboardID),
	}
	if version != dashboardLatestVersion {
		input.VersionNumber = aws.Int64(version)
	}

	return findDashboard(ctx, conn, input)
}

func findDashboard(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDashboardInput) (*awstypes.Dashboard, error) {
	output, err := conn.DescribeDashboard(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Dashboard == nil || output.Dashboard.Version == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Dashboard, nil
}

func findDashboardDefinitionByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dashboardID string, version int64) (*quicksight.DescribeDashboardDefinitionOutput, error) {
	input := &quicksight.DescribeDashboardDefinitionInput{
		AwsAccountId:  aws.String(awsAccountID),
		DashboardId:   aws.String(dashboardID),
		VersionNumber: aws.Int64(version),
	}

	return findDashboardDefinition(ctx, conn, input)
}

func findDashboardDefinition(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDashboardDefinitionInput) (*quicksight.DescribeDashboardDefinitionOutput, error) {
	output, err := conn.DescribeDashboardDefinition(ctx, input)

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

	return output, nil
}

func findDashboardPermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dashboardID string) ([]awstypes.ResourcePermission, error) {
	input := &quicksight.DescribeDashboardPermissionsInput{
		AwsAccountId: aws.String(awsAccountID),
		DashboardId:  aws.String(dashboardID),
	}

	return findDashboardPermissions(ctx, conn, input)
}

func findDashboardPermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDashboardPermissionsInput) ([]awstypes.ResourcePermission, error) {
	output, err := conn.DescribeDashboardPermissions(ctx, input)

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

func statusDashboard(ctx context.Context, conn *quicksight.Client, awsAccountID, dashboardID string, version int64) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDashboardByThreePartKey(ctx, conn, awsAccountID, dashboardID, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Version.Status), nil
	}
}

func waitDashboardCreated(ctx context.Context, conn *quicksight.Client, awsAccountID, dashboardID string, timeout time.Duration) (*awstypes.Dashboard, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusDashboard(ctx, conn, awsAccountID, dashboardID, dashboardLatestVersion),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Dashboard); ok {
		if status, apiErrors := output.Version.Status, output.Version.Errors; status == awstypes.ResourceStatusCreationFailed {
			tfresource.SetLastError(err, dashboardError(apiErrors))
		}

		return output, err
	}

	return nil, err
}

func waitDashboardUpdated(ctx context.Context, conn *quicksight.Client, awsAccountID, dashboardID string, version int64, timeout time.Duration) (*awstypes.Dashboard, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusUpdateInProgress, awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusUpdateSuccessful, awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusDashboard(ctx, conn, awsAccountID, dashboardID, version),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Dashboard); ok {
		if status, apiErrors := output.Version.Status, output.Version.Errors; status == awstypes.ResourceStatusUpdateFailed {
			tfresource.SetLastError(err, dashboardError(apiErrors))
		}

		return output, err
	}

	return nil, err
}

func dashboardError(apiObjects []awstypes.DashboardError) error {
	errs := tfslices.ApplyToAll(apiObjects, func(v awstypes.DashboardError) error {
		return fmt.Errorf("%s: %s", v.Type, aws.ToString(v.Message))
	})

	return errors.Join(errs...)
}
