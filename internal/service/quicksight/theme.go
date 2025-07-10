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

// @SDKResource("aws_quicksight_theme", name="Theme")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.Theme")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func resourceTheme() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThemeCreate,
		ReadWithoutTimeout:   resourceThemeRead,
		UpdateWithoutTimeout: resourceThemeUpdate,
		DeleteWithoutTimeout: resourceThemeDelete,

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
				"base_theme_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrConfiguration: quicksightschema.ThemeConfigurationSchema(),
				names.AttrCreatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"theme_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
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
				names.AttrPermissions: quicksightschema.PermissionsSchema(),
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"version_description": {
					Type:         schema.TypeString,
					Optional:     true,
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

func resourceThemeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	themeID := d.Get("theme_id").(string)
	id := themeCreateResourceID(awsAccountID, themeID)
	input := &quicksight.CreateThemeInput{
		AwsAccountId: aws.String(awsAccountID),
		BaseThemeId:  aws.String(d.Get("base_theme_id").(string)),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Tags:         getTagsIn(ctx),
		ThemeId:      aws.String(themeID),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Configuration = quicksightschema.ExpandThemeConfiguration(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("version_description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	_, err := conn.CreateTheme(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Theme (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitThemeCreated(ctx, conn, awsAccountID, themeID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Theme (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceThemeRead(ctx, d, meta)...)
}

func resourceThemeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, themeID, err := themeParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	theme, err := findThemeByTwoPartKey(ctx, conn, awsAccountID, themeID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Theme (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Theme (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, theme.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set("base_theme_id", theme.Version.BaseThemeId)
	if err := d.Set(names.AttrConfiguration, quicksightschema.FlattenThemeConfiguration(theme.Version.Configuration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}
	d.Set(names.AttrCreatedTime, theme.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, theme.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, theme.Name)
	d.Set(names.AttrStatus, theme.Version.Status)
	d.Set("theme_id", theme.ThemeId)
	d.Set("version_description", theme.Version.Description)
	d.Set("version_number", theme.Version.VersionNumber)

	permissions, err := findThemePermissionsByTwoPartKey(ctx, conn, awsAccountID, themeID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Theme (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceThemeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, themeID, err := themeParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		input := &quicksight.UpdateThemeInput{
			AwsAccountId: aws.String(awsAccountID),
			BaseThemeId:  aws.String(d.Get("base_theme_id").(string)),
			Name:         aws.String(d.Get(names.AttrName).(string)),
			ThemeId:      aws.String(themeID),
		}

		if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.Configuration = quicksightschema.ExpandThemeConfiguration(v.([]any))
		}

		_, err := conn.UpdateTheme(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Theme (%s): %s", d.Id(), err)
		}

		if _, err := waitThemeUpdated(ctx, conn, awsAccountID, themeID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Theme (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		o, n := d.GetChange(names.AttrPermissions)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		toGrant, toRevoke := quicksightschema.DiffPermissions(os.List(), ns.List())

		input := &quicksight.UpdateThemePermissionsInput{
			AwsAccountId: aws.String(awsAccountID),
			ThemeId:      aws.String(themeID),
		}

		if len(toGrant) > 0 {
			input.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			input.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateThemePermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Theme (%s) permissions: %s", d.Id(), err)
		}
	}

	return append(diags, resourceThemeRead(ctx, d, meta)...)
}

func resourceThemeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, themeID, err := themeParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Theme: %s", d.Id())
	_, err = conn.DeleteTheme(ctx, &quicksight.DeleteThemeInput{
		AwsAccountId: aws.String(awsAccountID),
		ThemeId:      aws.String(themeID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Theme (%s): %s", d.Id(), err)
	}

	return diags
}

const themeResourceIDSeparator = ","

func themeCreateResourceID(awsAccountID, themeID string) string {
	parts := []string{awsAccountID, themeID}
	id := strings.Join(parts, themeResourceIDSeparator)

	return id
}

func themeParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, themeResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sTHEME_ID", id, themeResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findThemeByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, themeID string) (*awstypes.Theme, error) {
	input := &quicksight.DescribeThemeInput{
		AwsAccountId: aws.String(awsAccountID),
		ThemeId:      aws.String(themeID),
	}

	return findTheme(ctx, conn, input)
}

func findTheme(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeThemeInput) (*awstypes.Theme, error) {
	output, err := conn.DescribeTheme(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Theme == nil || output.Theme.Version == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Theme, nil
}

func findThemePermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, themeID string) ([]awstypes.ResourcePermission, error) {
	input := &quicksight.DescribeThemePermissionsInput{
		AwsAccountId: aws.String(awsAccountID),
		ThemeId:      aws.String(themeID),
	}

	return findThemePermissions(ctx, conn, input)
}

func findThemePermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeThemePermissionsInput) ([]awstypes.ResourcePermission, error) {
	output, err := conn.DescribeThemePermissions(ctx, input)

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

func statusTheme(ctx context.Context, conn *quicksight.Client, awsAccountID, themeID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findThemeByTwoPartKey(ctx, conn, awsAccountID, themeID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Version.Status), nil
	}
}

func waitThemeCreated(ctx context.Context, conn *quicksight.Client, awsAccountID, themeID string, timeout time.Duration) (*awstypes.Theme, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusTheme(ctx, conn, awsAccountID, themeID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Theme); ok {
		if status, apiErrors := output.Version.Status, output.Version.Errors; status == awstypes.ResourceStatusCreationFailed {
			tfresource.SetLastError(err, themeError(apiErrors))
		}

		return output, err
	}

	return nil, err
}

func waitThemeUpdated(ctx context.Context, conn *quicksight.Client, awsAccountID, themeID string, timeout time.Duration) (*awstypes.Theme, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusUpdateInProgress, awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusUpdateSuccessful, awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusTheme(ctx, conn, awsAccountID, themeID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Theme); ok {
		if status, apiErrors := output.Version.Status, output.Version.Errors; status == awstypes.ResourceStatusUpdateFailed {
			tfresource.SetLastError(err, themeError(apiErrors))
		}

		return output, err
	}

	return nil, err
}

func themeError(apiObjects []awstypes.ThemeError) error {
	errs := tfslices.ApplyToAll(apiObjects, func(v awstypes.ThemeError) error {
		return fmt.Errorf("%s: %s", v.Type, aws.ToString(v.Message))
	})

	return errors.Join(errs...)
}
