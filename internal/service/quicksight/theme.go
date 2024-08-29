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

	"github.com/YakDriver/regexache"
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

// @SDKResource("aws_quicksight_theme", name="Theme")
// @Tags(identifierAttribute="arn")
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
				names.AttrConfiguration: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ThemeConfiguration.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_color_palette": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataColorPalette.html
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"colors": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 8, // Colors size needs to be in the range between 8 and 20
											MaxItems: 20,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
											},
										},
										"empty_fill_color": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"min_max_gradient": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 2, // MinMaxGradient size needs to be 2
											MaxItems: 2,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
											},
										},
									},
								},
							},
							"sheet": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetStyle.html
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"tile": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TileStyle.html
											Type:     schema.TypeList,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"border": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BorderStyle.html
														Type:     schema.TypeList,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"show": {
																	Type:     schema.TypeBool,
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
										"tile_layout": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TileLayoutStyle.html
											Type:     schema.TypeList,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"gutter": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GutterStyle.html
														Type:     schema.TypeList,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"show": {
																	Type:     schema.TypeBool,
																	Optional: true,
																},
															},
														},
													},
													"margin": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MarginStyle.html
														Type:     schema.TypeList,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"show": {
																	Type:     schema.TypeBool,
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"typography": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Typography.html
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"font_families": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Font.html
											Type:     schema.TypeList,
											MaxItems: 5,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"font_family": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"ui_color_palette": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_UIColorPalette.html
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"accent": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"accent_foreground": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"danger": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"danger_foreground": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"dimension": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"dimension_foreground": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"measure": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"measure_foreground": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"primary_background": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"primary_foreground": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"secondary_background": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"secondary_foreground": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"success": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"success_foreground": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"warning": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
										"warning_foreground": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""),
										},
									},
								},
							},
						},
					},
				},
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThemeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID
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

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Configuration = expandThemeConfiguration(v.([]interface{}))
	}

	if v, ok := d.Get(names.AttrPermissions).(*schema.Set); ok && v.Len() > 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.List())
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

func resourceThemeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	if err := d.Set(names.AttrConfiguration, flattenThemeConfiguration(theme.Version.Configuration)); err != nil {
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

func resourceThemeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

		if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Configuration = expandThemeConfiguration(v.([]interface{}))
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

func resourceThemeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	return func() (interface{}, string, error) {
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

func expandThemeConfiguration(tfList []interface{}) *awstypes.ThemeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ThemeConfiguration{}

	if v, ok := tfMap["data_color_palette"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataColorPalette = expandDataColorPalette(v)
	}
	if v, ok := tfMap["sheet"].([]interface{}); ok && len(v) > 0 {
		apiObject.Sheet = expandSheetStyle(v)
	}
	if v, ok := tfMap["typography"].([]interface{}); ok && len(v) > 0 {
		apiObject.Typography = expandTypography(v)
	}
	if v, ok := tfMap["ui_color_palette"].([]interface{}); ok && len(v) > 0 {
		apiObject.UIColorPalette = expandUIColorPalette(v)
	}

	return apiObject
}

func expandDataColorPalette(tfList []interface{}) *awstypes.DataColorPalette {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataColorPalette{}

	if v, ok := tfMap["colors"].([]interface{}); ok {
		apiObject.Colors = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["empty_fill_color"].(string); ok && v != "" {
		apiObject.EmptyFillColor = aws.String(v)
	}
	if v, ok := tfMap["min_max_gradient"].([]interface{}); ok {
		apiObject.MinMaxGradient = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandSheetStyle(tfList []interface{}) *awstypes.SheetStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SheetStyle{}

	if v, ok := tfMap["tile"].([]interface{}); ok && len(v) > 0 {
		apiObject.Tile = expandTileStyle(v)
	}
	if v, ok := tfMap["tile_layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.TileLayout = expandTileLayoutStyle(v)
	}

	return apiObject
}

func expandTileStyle(tfList []interface{}) *awstypes.TileStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TileStyle{}

	if v, ok := tfMap["border"].([]interface{}); ok && len(v) > 0 {
		apiObject.Border = expandBorderStyle(v)
	}

	return apiObject
}

func expandBorderStyle(tfList []interface{}) *awstypes.BorderStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.BorderStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		apiObject.Show = aws.Bool(v)
	}

	return apiObject
}

func expandTileLayoutStyle(tfList []interface{}) *awstypes.TileLayoutStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TileLayoutStyle{}

	if v, ok := tfMap["gutter"].([]interface{}); ok && len(v) > 0 {
		apiObject.Gutter = expandGutterStyle(v)
	}
	if v, ok := tfMap["margin"].([]interface{}); ok && len(v) > 0 {
		apiObject.Margin = expandMarginStyle(v)
	}

	return apiObject
}

func expandGutterStyle(tfList []interface{}) *awstypes.GutterStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GutterStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		apiObject.Show = aws.Bool(v)
	}

	return apiObject
}

func expandMarginStyle(tfList []interface{}) *awstypes.MarginStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.MarginStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		apiObject.Show = aws.Bool(v)
	}

	return apiObject
}

func expandTypography(tfList []interface{}) *awstypes.Typography {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.Typography{}

	if v, ok := tfMap["font_families"].([]interface{}); ok && len(v) > 0 {
		apiObject.FontFamilies = expandFonts(v)
	}

	return apiObject
}

func expandFonts(tfList []interface{}) []awstypes.Font {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Font

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandFont(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFont(tfMap map[string]interface{}) *awstypes.Font {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Font{}

	if v, ok := tfMap["font_family"].(string); ok && v != "" {
		apiObject.FontFamily = aws.String(v)
	}

	return apiObject
}

func expandUIColorPalette(tfList []interface{}) *awstypes.UIColorPalette {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.UIColorPalette{}

	if v, ok := tfMap["accent"].(string); ok && v != "" {
		apiObject.Accent = aws.String(v)
	}
	if v, ok := tfMap["accent_foreground"].(string); ok && v != "" {
		apiObject.AccentForeground = aws.String(v)
	}
	if v, ok := tfMap["danger"].(string); ok && v != "" {
		apiObject.Danger = aws.String(v)
	}
	if v, ok := tfMap["danger_foreground"].(string); ok && v != "" {
		apiObject.DangerForeground = aws.String(v)
	}
	if v, ok := tfMap["dimension"].(string); ok && v != "" {
		apiObject.Dimension = aws.String(v)
	}
	if v, ok := tfMap["dimension_foreground"].(string); ok && v != "" {
		apiObject.DimensionForeground = aws.String(v)
	}
	if v, ok := tfMap["measure"].(string); ok && v != "" {
		apiObject.Measure = aws.String(v)
	}
	if v, ok := tfMap["measure_foreground"].(string); ok && v != "" {
		apiObject.MeasureForeground = aws.String(v)
	}
	if v, ok := tfMap["primary_background"].(string); ok && v != "" {
		apiObject.PrimaryBackground = aws.String(v)
	}
	if v, ok := tfMap["primary_foreground"].(string); ok && v != "" {
		apiObject.PrimaryForeground = aws.String(v)
	}
	if v, ok := tfMap["secondary_background"].(string); ok && v != "" {
		apiObject.SecondaryBackground = aws.String(v)
	}
	if v, ok := tfMap["secondary_foreground"].(string); ok && v != "" {
		apiObject.SecondaryForeground = aws.String(v)
	}
	if v, ok := tfMap["success"].(string); ok && v != "" {
		apiObject.Success = aws.String(v)
	}
	if v, ok := tfMap["success_foreground"].(string); ok && v != "" {
		apiObject.SuccessForeground = aws.String(v)
	}
	if v, ok := tfMap["warning"].(string); ok && v != "" {
		apiObject.Warning = aws.String(v)
	}
	if v, ok := tfMap["warning_foreground"].(string); ok && v != "" {
		apiObject.WarningForeground = aws.String(v)
	}

	return apiObject
}

func flattenThemeConfiguration(apiObject *awstypes.ThemeConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DataColorPalette != nil {
		tfMap["data_color_palette"] = flattenDataColorPalette(apiObject.DataColorPalette)
	}
	if apiObject.Sheet != nil {
		tfMap["sheet"] = flattenSheetStyle(apiObject.Sheet)
	}
	if apiObject.Typography != nil {
		tfMap["typography"] = flattenTypography(apiObject.Typography)
	}
	if apiObject.UIColorPalette != nil {
		tfMap["ui_color_palette"] = flattenUIColorPalette(apiObject.UIColorPalette)
	}

	return []interface{}{tfMap}
}

func flattenDataColorPalette(apiObject *awstypes.DataColorPalette) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Colors != nil {
		tfMap["colors"] = apiObject.Colors
	}
	if apiObject.EmptyFillColor != nil {
		tfMap["empty_fill_color"] = aws.ToString(apiObject.EmptyFillColor)
	}
	if apiObject.MinMaxGradient != nil {
		tfMap["min_max_gradient"] = apiObject.MinMaxGradient
	}

	return []interface{}{tfMap}
}

func flattenSheetStyle(apiObject *awstypes.SheetStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Tile != nil {
		tfMap["tile"] = flattenTileStyle(apiObject.Tile)
	}
	if apiObject.TileLayout != nil {
		tfMap["tile_layout"] = flattenTileLayoutStyle(apiObject.TileLayout)
	}

	return []interface{}{tfMap}
}

func flattenTileStyle(apiObject *awstypes.TileStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Border != nil {
		tfMap["border"] = flattenBorderStyle(apiObject.Border)
	}

	return []interface{}{tfMap}
}

func flattenBorderStyle(apiObject *awstypes.BorderStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Show != nil {
		tfMap["show"] = aws.ToBool(apiObject.Show)
	}

	return []interface{}{tfMap}
}

func flattenTileLayoutStyle(apiObject *awstypes.TileLayoutStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Gutter != nil {
		tfMap["gutter"] = flattenGutterStyle(apiObject.Gutter)
	}
	if apiObject.Margin != nil {
		tfMap["margin"] = flattenMarginStyle(apiObject.Margin)
	}

	return []interface{}{tfMap}
}

func flattenGutterStyle(apiObject *awstypes.GutterStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Show != nil {
		tfMap["show"] = aws.ToBool(apiObject.Show)
	}

	return []interface{}{tfMap}
}

func flattenMarginStyle(apiObject *awstypes.MarginStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Show != nil {
		tfMap["show"] = aws.ToBool(apiObject.Show)
	}

	return []interface{}{tfMap}
}

func flattenTypography(apiObject *awstypes.Typography) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.FontFamilies != nil {
		tfMap["font_families"] = flattenFonts(apiObject.FontFamilies)
	}

	return []interface{}{tfMap}
}

func flattenFonts(apiObject []awstypes.Font) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObject {
		tfMap := map[string]interface{}{}

		if apiObject.FontFamily != nil {
			tfMap["font_family"] = aws.ToString(apiObject.FontFamily)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenUIColorPalette(apiObject *awstypes.UIColorPalette) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Accent != nil {
		tfMap["accent"] = aws.ToString(apiObject.Accent)
	}
	if apiObject.AccentForeground != nil {
		tfMap["accent_foreground"] = aws.ToString(apiObject.AccentForeground)
	}
	if apiObject.Danger != nil {
		tfMap["danger"] = aws.ToString(apiObject.Danger)
	}
	if apiObject.DangerForeground != nil {
		tfMap["danger_foreground"] = aws.ToString(apiObject.DangerForeground)
	}
	if apiObject.Dimension != nil {
		tfMap["dimension"] = aws.ToString(apiObject.Dimension)
	}
	if apiObject.DimensionForeground != nil {
		tfMap["dimension_foreground"] = aws.ToString(apiObject.DimensionForeground)
	}
	if apiObject.Measure != nil {
		tfMap["measure"] = aws.ToString(apiObject.Measure)
	}
	if apiObject.MeasureForeground != nil {
		tfMap["measure_foreground"] = aws.ToString(apiObject.MeasureForeground)
	}
	if apiObject.PrimaryBackground != nil {
		tfMap["primary_background"] = aws.ToString(apiObject.PrimaryBackground)
	}
	if apiObject.PrimaryForeground != nil {
		tfMap["primary_foreground"] = aws.ToString(apiObject.PrimaryForeground)
	}
	if apiObject.SecondaryBackground != nil {
		tfMap["secondary_background"] = aws.ToString(apiObject.SecondaryBackground)
	}
	if apiObject.SecondaryForeground != nil {
		tfMap["secondary_foreground"] = aws.ToString(apiObject.SecondaryForeground)
	}
	if apiObject.Success != nil {
		tfMap["success"] = aws.ToString(apiObject.Success)
	}
	if apiObject.SuccessForeground != nil {
		tfMap["success_foreground"] = aws.ToString(apiObject.SuccessForeground)
	}
	if apiObject.Warning != nil {
		tfMap["warning"] = aws.ToString(apiObject.Warning)
	}
	if apiObject.WarningForeground != nil {
		tfMap["warning_foreground"] = aws.ToString(apiObject.WarningForeground)
	}

	return []interface{}{tfMap}
}
