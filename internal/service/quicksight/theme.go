// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_theme", name="Theme")
// @Tags(identifierAttribute="arn")
func ResourceTheme() *schema.Resource {
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
				names.AttrPermissions: {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 64,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrActions: {
								Type:     schema.TypeSet,
								Required: true,
								MinItems: 1,
								MaxItems: 16,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							names.AttrPrincipal: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
						},
					},
				},
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

const (
	ResNameTheme = "Theme"
)

func resourceThemeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountId = v.(string)
	}
	themeId := d.Get("theme_id").(string)

	d.SetId(createThemeId(awsAccountId, themeId))

	input := &quicksight.CreateThemeInput{
		AwsAccountId: aws.String(awsAccountId),
		ThemeId:      aws.String(themeId),
		Name:         aws.String(d.Get(names.AttrName).(string)),
		BaseThemeId:  aws.String(d.Get("base_theme_id").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("version_description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Configuration = expandThemeConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Permissions = expandResourcePermissions(v.([]interface{}))
	}

	_, err := conn.CreateThemeWithContext(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionCreating, ResNameTheme, d.Get(names.AttrName).(string), err)
	}

	if _, err := waitThemeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForCreation, ResNameTheme, d.Id(), err)
	}

	return append(diags, resourceThemeRead(ctx, d, meta)...)
}

func resourceThemeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, themeId, err := ParseThemeId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := FindThemeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Theme (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionReading, ResNameTheme, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountId)
	d.Set("base_theme_id", out.Version.BaseThemeId)
	d.Set(names.AttrCreatedTime, out.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, out.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrStatus, out.Version.Status)
	d.Set("theme_id", out.ThemeId)
	d.Set("version_description", out.Version.Description)
	d.Set("version_number", out.Version.VersionNumber)

	if err := d.Set(names.AttrConfiguration, flattenThemeConfiguration(out.Version.Configuration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}

	permsResp, err := conn.DescribeThemePermissionsWithContext(ctx, &quicksight.DescribeThemePermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		ThemeId:      aws.String(themeId),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Theme (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, flattenPermissions(permsResp.Permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceThemeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, themeId, err := ParseThemeId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrPermissions, names.AttrTags, names.AttrTagsAll) {
		in := &quicksight.UpdateThemeInput{
			AwsAccountId: aws.String(awsAccountId),
			ThemeId:      aws.String(themeId),
			BaseThemeId:  aws.String(d.Get("base_theme_id").(string)),
			Name:         aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.Configuration = expandThemeConfiguration(v.([]interface{}))
		}

		log.Printf("[DEBUG] Updating QuickSight Theme (%s): %#v", d.Id(), in)
		_, err := conn.UpdateThemeWithContext(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionUpdating, ResNameTheme, d.Id(), err)
		}

		if _, err := waitThemeUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.QuickSight, create.ErrActionWaitingForUpdate, ResNameTheme, d.Id(), err)
		}
	}

	if d.HasChange(names.AttrPermissions) {
		oraw, nraw := d.GetChange(names.AttrPermissions)
		o := oraw.([]interface{})
		n := nraw.([]interface{})

		toGrant, toRevoke := DiffPermissions(o, n)

		params := &quicksight.UpdateThemePermissionsInput{
			AwsAccountId: aws.String(awsAccountId),
			ThemeId:      aws.String(themeId),
		}

		if len(toGrant) > 0 {
			params.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			params.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateThemePermissionsWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Theme (%s) permissions: %s", themeId, err)
		}
	}

	return append(diags, resourceThemeRead(ctx, d, meta)...)
}

func resourceThemeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, themeId, err := ParseThemeId(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &quicksight.DeleteThemeInput{
		AwsAccountId: aws.String(awsAccountId),
		ThemeId:      aws.String(themeId),
	}

	log.Printf("[INFO] Deleting QuickSight Theme %s", d.Id())
	_, err = conn.DeleteThemeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.QuickSight, create.ErrActionDeleting, ResNameTheme, d.Id(), err)
	}

	return diags
}

func FindThemeByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.Theme, error) {
	awsAccountId, themeId, err := ParseThemeId(id)
	if err != nil {
		return nil, err
	}

	descOpts := &quicksight.DescribeThemeInput{
		AwsAccountId: aws.String(awsAccountId),
		ThemeId:      aws.String(themeId),
	}

	out, err := conn.DescribeThemeWithContext(ctx, descOpts)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: descOpts,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Theme == nil {
		return nil, tfresource.NewEmptyResultError(descOpts)
	}

	return out.Theme, nil
}

func ParseThemeId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,TTHEME_ID", id)
	}
	return parts[0], parts[1], nil
}

func createThemeId(awsAccountID, themeId string) string {
	return fmt.Sprintf("%s,%s", awsAccountID, themeId)
}

func expandThemeConfiguration(tfList []interface{}) *quicksight.ThemeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ThemeConfiguration{}

	if v, ok := tfMap["data_color_palette"].([]interface{}); ok && len(v) > 0 {
		config.DataColorPalette = expandDataColorPalette(v)
	}
	if v, ok := tfMap["sheet"].([]interface{}); ok && len(v) > 0 {
		config.Sheet = expandSheetStyle(v)
	}
	if v, ok := tfMap["typography"].([]interface{}); ok && len(v) > 0 {
		config.Typography = expandTypography(v)
	}
	if v, ok := tfMap["ui_color_palette"].([]interface{}); ok && len(v) > 0 {
		config.UIColorPalette = expandUIColorPalette(v)
	}

	return config
}

func expandDataColorPalette(tfList []interface{}) *quicksight.DataColorPalette {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DataColorPalette{}

	if v, ok := tfMap["colors"].([]interface{}); ok {
		config.Colors = flex.ExpandStringList(v)
	}
	if v, ok := tfMap["empty_fill_color"].(string); ok && v != "" {
		config.EmptyFillColor = aws.String(v)
	}
	if v, ok := tfMap["min_max_gradient"].([]interface{}); ok {
		config.MinMaxGradient = flex.ExpandStringList(v)
	}

	return config
}

func expandSheetStyle(tfList []interface{}) *quicksight.SheetStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SheetStyle{}

	if v, ok := tfMap["tile"].([]interface{}); ok && len(v) > 0 {
		config.Tile = expandTileStyle(v)
	}
	if v, ok := tfMap["tile_layout"].([]interface{}); ok && len(v) > 0 {
		config.TileLayout = expandTileLayoutStyle(v)
	}

	return config
}

func expandTileStyle(tfList []interface{}) *quicksight.TileStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.TileStyle{}

	if v, ok := tfMap["border"].([]interface{}); ok && len(v) > 0 {
		config.Border = expandBorderStyle(v)
	}

	return config
}

func expandBorderStyle(tfList []interface{}) *quicksight.BorderStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.BorderStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		config.Show = aws.Bool(v)
	}

	return config
}

func expandTileLayoutStyle(tfList []interface{}) *quicksight.TileLayoutStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.TileLayoutStyle{}

	if v, ok := tfMap["gutter"].([]interface{}); ok && len(v) > 0 {
		config.Gutter = expandGutterStyle(v)
	}
	if v, ok := tfMap["margin"].([]interface{}); ok && len(v) > 0 {
		config.Margin = expandMarginStyle(v)
	}

	return config
}

func expandGutterStyle(tfList []interface{}) *quicksight.GutterStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.GutterStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		config.Show = aws.Bool(v)
	}

	return config
}

func expandMarginStyle(tfList []interface{}) *quicksight.MarginStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.MarginStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		config.Show = aws.Bool(v)
	}

	return config
}

func expandTypography(tfList []interface{}) *quicksight.Typography {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.Typography{}

	if v, ok := tfMap["font_families"].([]interface{}); ok && len(v) > 0 {
		config.FontFamilies = expandFontFamilies(v)
	}
	return config
}

func expandFontFamilies(tfList []interface{}) []*quicksight.Font {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.Font
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		font := expandFont(tfMap)
		if font == nil {
			continue
		}

		configs = append(configs, font)
	}

	return configs
}

func expandFont(tfMap map[string]interface{}) *quicksight.Font {
	if tfMap == nil {
		return nil
	}

	font := &quicksight.Font{}

	if v, ok := tfMap["font_family"].(string); ok && v != "" {
		font.FontFamily = aws.String(v)
	}

	return font
}

func expandUIColorPalette(tfList []interface{}) *quicksight.UIColorPalette {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.UIColorPalette{}

	if v, ok := tfMap["accent"].(string); ok && v != "" {
		config.Accent = aws.String(v)
	}
	if v, ok := tfMap["accent_foreground"].(string); ok && v != "" {
		config.AccentForeground = aws.String(v)
	}
	if v, ok := tfMap["danger"].(string); ok && v != "" {
		config.Danger = aws.String(v)
	}
	if v, ok := tfMap["danger_foreground"].(string); ok && v != "" {
		config.DangerForeground = aws.String(v)
	}
	if v, ok := tfMap["dimension"].(string); ok && v != "" {
		config.Dimension = aws.String(v)
	}
	if v, ok := tfMap["dimension_foreground"].(string); ok && v != "" {
		config.DimensionForeground = aws.String(v)
	}
	if v, ok := tfMap["measure"].(string); ok && v != "" {
		config.Measure = aws.String(v)
	}
	if v, ok := tfMap["measure_foreground"].(string); ok && v != "" {
		config.MeasureForeground = aws.String(v)
	}
	if v, ok := tfMap["primary_background"].(string); ok && v != "" {
		config.PrimaryBackground = aws.String(v)
	}
	if v, ok := tfMap["primary_foreground"].(string); ok && v != "" {
		config.PrimaryForeground = aws.String(v)
	}
	if v, ok := tfMap["secondary_background"].(string); ok && v != "" {
		config.SecondaryBackground = aws.String(v)
	}
	if v, ok := tfMap["secondary_foreground"].(string); ok && v != "" {
		config.SecondaryForeground = aws.String(v)
	}
	if v, ok := tfMap["success"].(string); ok && v != "" {
		config.Success = aws.String(v)
	}
	if v, ok := tfMap["success_foreground"].(string); ok && v != "" {
		config.SuccessForeground = aws.String(v)
	}
	if v, ok := tfMap["warning"].(string); ok && v != "" {
		config.Warning = aws.String(v)
	}
	if v, ok := tfMap["warning_foreground"].(string); ok && v != "" {
		config.WarningForeground = aws.String(v)
	}

	return config
}

func flattenThemeConfiguration(apiObject *quicksight.ThemeConfiguration) []interface{} {
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

func flattenDataColorPalette(apiObject *quicksight.DataColorPalette) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Colors != nil {
		tfMap["colors"] = flex.FlattenStringList(apiObject.Colors)
	}
	if apiObject.EmptyFillColor != nil {
		tfMap["empty_fill_color"] = aws.StringValue(apiObject.EmptyFillColor)
	}
	if apiObject.MinMaxGradient != nil {
		tfMap["min_max_gradient"] = flex.FlattenStringList(apiObject.MinMaxGradient)
	}

	return []interface{}{tfMap}
}

func flattenSheetStyle(apiObject *quicksight.SheetStyle) []interface{} {
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

func flattenTileStyle(apiObject *quicksight.TileStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Border != nil {
		tfMap["border"] = flattenBorderStyle(apiObject.Border)
	}

	return []interface{}{tfMap}
}

func flattenBorderStyle(apiObject *quicksight.BorderStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Show != nil {
		tfMap["show"] = aws.BoolValue(apiObject.Show)
	}

	return []interface{}{tfMap}
}

func flattenTileLayoutStyle(apiObject *quicksight.TileLayoutStyle) []interface{} {
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

func flattenGutterStyle(apiObject *quicksight.GutterStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Show != nil {
		tfMap["show"] = aws.BoolValue(apiObject.Show)
	}

	return []interface{}{tfMap}
}

func flattenMarginStyle(apiObject *quicksight.MarginStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Show != nil {
		tfMap["show"] = aws.BoolValue(apiObject.Show)
	}

	return []interface{}{tfMap}
}

func flattenTypography(apiObject *quicksight.Typography) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FontFamilies != nil {
		tfMap["font_families"] = flattenFonts(apiObject.FontFamilies)
	}

	return []interface{}{tfMap}
}

func flattenFonts(apiObject []*quicksight.Font) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, font := range apiObject {
		if font == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if font.FontFamily != nil {
			tfMap["font_family"] = aws.StringValue(font.FontFamily)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenUIColorPalette(apiObject *quicksight.UIColorPalette) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Accent != nil {
		tfMap["accent"] = aws.StringValue(apiObject.Accent)
	}
	if apiObject.AccentForeground != nil {
		tfMap["accent_foreground"] = aws.StringValue(apiObject.AccentForeground)
	}
	if apiObject.Danger != nil {
		tfMap["danger"] = aws.StringValue(apiObject.Danger)
	}
	if apiObject.DangerForeground != nil {
		tfMap["danger_foreground"] = aws.StringValue(apiObject.DangerForeground)
	}
	if apiObject.Dimension != nil {
		tfMap["dimension"] = aws.StringValue(apiObject.Dimension)
	}
	if apiObject.DimensionForeground != nil {
		tfMap["dimension_foreground"] = aws.StringValue(apiObject.DimensionForeground)
	}
	if apiObject.Measure != nil {
		tfMap["measure"] = aws.StringValue(apiObject.Measure)
	}
	if apiObject.MeasureForeground != nil {
		tfMap["measure_foreground"] = aws.StringValue(apiObject.MeasureForeground)
	}
	if apiObject.PrimaryBackground != nil {
		tfMap["primary_background"] = aws.StringValue(apiObject.PrimaryBackground)
	}
	if apiObject.PrimaryForeground != nil {
		tfMap["primary_foreground"] = aws.StringValue(apiObject.PrimaryForeground)
	}
	if apiObject.SecondaryBackground != nil {
		tfMap["secondary_background"] = aws.StringValue(apiObject.SecondaryBackground)
	}
	if apiObject.SecondaryForeground != nil {
		tfMap["secondary_foreground"] = aws.StringValue(apiObject.SecondaryForeground)
	}
	if apiObject.Success != nil {
		tfMap["success"] = aws.StringValue(apiObject.Success)
	}
	if apiObject.SuccessForeground != nil {
		tfMap["success_foreground"] = aws.StringValue(apiObject.SuccessForeground)
	}
	if apiObject.Warning != nil {
		tfMap["warning"] = aws.StringValue(apiObject.Warning)
	}
	if apiObject.WarningForeground != nil {
		tfMap["warning_foreground"] = aws.StringValue(apiObject.WarningForeground)
	}

	return []interface{}{tfMap}
}
