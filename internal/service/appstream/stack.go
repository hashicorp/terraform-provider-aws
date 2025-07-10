// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	stackOperationTimeout = 4 * time.Minute
)

// @SDKResource("aws_appstream_stack", name="Stack")
// @Tags(identifierAttribute="arn")
func resourceStack() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStackCreate,
		ReadWithoutTimeout:   resourceStackRead,
		UpdateWithoutTimeout: resourceStackUpdate,
		DeleteWithoutTimeout: resourceStackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_endpoints": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEndpointType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AccessEndpointType](),
						},
						"vpce_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"application_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"settings_group": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 100),
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			names.AttrDisplayName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"embed_host_domains": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 20,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 128),
				},
			},
			"feedback_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"redirect_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"storage_connectors": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.StorageConnectorType](),
						},
						"domains": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 50,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 64),
							},
						},
						"resource_identifier": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},
			"streaming_experience_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"preferred_protocol": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PreferredProtocol](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_settings": {
				Type:             schema.TypeSet,
				Optional:         true,
				Computed:         true,
				MinItems:         1,
				DiffSuppressFunc: suppressAppsStreamStackUserSettings,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAction: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Action](),
						},
						"permission": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Permission](),
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.All(
			func(_ context.Context, d *schema.ResourceDiff, meta any) error {
				if d.Id() == "" {
					return nil
				}

				rawConfig := d.GetRawConfig()
				configApplicationSettings := rawConfig.GetAttr("application_settings")
				if configApplicationSettings.IsKnown() && !configApplicationSettings.IsNull() && configApplicationSettings.LengthInt() > 0 {
					return nil
				}

				rawState := d.GetRawState()
				stateApplicationSettings := rawState.GetAttr("application_settings")
				if stateApplicationSettings.IsKnown() && !stateApplicationSettings.IsNull() && stateApplicationSettings.LengthInt() > 0 {
					setting := stateApplicationSettings.Index(cty.NumberIntVal(0))
					if setting.IsKnown() && !setting.IsNull() {
						enabled := setting.GetAttr(names.AttrEnabled)
						if enabled.IsKnown() && !enabled.IsNull() && enabled.True() {
							// Trigger a diff
							return d.SetNew("application_settings", []map[string]any{
								{
									names.AttrEnabled: false,
									"settings_group":  "",
								},
							})
						}
					}
				}

				return nil
			},
			func(_ context.Context, d *schema.ResourceDiff, meta any) error {
				_, enabled := d.GetOk("application_settings.0.enabled")
				v, sg := d.GetOk("application_settings.0.settings_group")
				log.Print(v)

				if enabled && !sg {
					return errors.New("application_settings.settings_group must be set when application_settings.enabled is true")
				} else if !enabled && sg {
					return errors.New("application_settings.settings_group must not be set when application_settings.enabled is false")
				}
				return nil
			},
		),
	}
}

func resourceStackCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := appstream.CreateStackInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_endpoints"); ok {
		input.AccessEndpoints = expandAccessEndpoints(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("application_settings"); ok {
		input.ApplicationSettings = expandApplicationSettings(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDisplayName); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("embed_host_domains"); ok {
		input.EmbedHostDomains = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("feedback_url"); ok {
		input.FeedbackURL = aws.String(v.(string))
	}

	if v, ok := d.GetOk("redirect_url"); ok {
		input.RedirectURL = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_connectors"); ok {
		input.StorageConnectors = expandStorageConnectors(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("streaming_experience_settings"); ok {
		input.StreamingExperienceSettings = expandStreamingExperienceSettings(v.([]any))
	}

	if v, ok := d.GetOk("user_settings"); ok {
		input.UserSettings = expandUserSettings(v.(*schema.Set).List())
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ResourceNotFoundException](ctx, stackOperationTimeout, func() (any, error) {
		return conn.CreateStack(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream Stack (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*appstream.CreateStackOutput).Stack.Name))

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	stack, err := findStackByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppStream Stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream Stack (%s): %s", d.Id(), err)
	}

	if err = d.Set("access_endpoints", flattenAccessEndpoints(stack.AccessEndpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_endpoints: %s", err)
	}
	if err = d.Set("application_settings", flattenApplicationSettings(stack.ApplicationSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting application_settings: %s", err)
	}
	d.Set(names.AttrARN, stack.Arn)
	d.Set(names.AttrCreatedTime, aws.ToTime(stack.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, stack.Description)
	d.Set(names.AttrDisplayName, stack.DisplayName)
	d.Set("embed_host_domains", stack.EmbedHostDomains)
	d.Set("feedback_url", stack.FeedbackURL)
	d.Set(names.AttrName, stack.Name)
	d.Set("redirect_url", stack.RedirectURL)
	if err = d.Set("storage_connectors", flattenStorageConnectors(stack.StorageConnectors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_connectors: %s", err)
	}
	if err = d.Set("streaming_experience_settings", flattenStreamingExperienceSettings(stack.StreamingExperienceSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting streaming_experience_settings: %s", err)
	}
	if err = d.Set("user_settings", flattenUserSettings(stack.UserSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting user_settings: %s", err)
	}

	return diags
}

func resourceStackUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := appstream.UpdateStackInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChange("access_endpoints") {
			input.AccessEndpoints = expandAccessEndpoints(d.Get("access_endpoints").(*schema.Set).List())
		}

		if d.HasChange("application_settings") {
			input.ApplicationSettings = expandApplicationSettings(d.Get("application_settings").([]any))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrDisplayName) {
			input.DisplayName = aws.String(d.Get(names.AttrDisplayName).(string))
		}

		if d.HasChange("feedback_url") {
			input.FeedbackURL = aws.String(d.Get("feedback_url").(string))
		}

		if d.HasChange("redirect_url") {
			input.RedirectURL = aws.String(d.Get("redirect_url").(string))
		}

		if d.HasChange("streaming_experience_settings") {
			input.StreamingExperienceSettings = expandStreamingExperienceSettings(d.Get("streaming_experience_settings").([]any))
		}

		if d.HasChange("user_settings") {
			input.UserSettings = expandUserSettings(d.Get("user_settings").(*schema.Set).List())
		}

		_, err := conn.UpdateStack(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppStream Stack (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	log.Printf("[DEBUG] Deleting AppStream Stack: %s", d.Id())
	input := appstream.DeleteStackInput{
		Name: aws.String(d.Id()),
	}
	_, err := conn.DeleteStack(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream Stack (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, stackOperationTimeout, func() (any, error) {
		return findStackByID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppStream Stack (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findStackByID(ctx context.Context, conn *appstream.Client, id string) (*awstypes.Stack, error) {
	input := appstream.DescribeStacksInput{
		Names: []string{id},
	}

	return findStack(ctx, conn, &input)
}

func findStack(ctx context.Context, conn *appstream.Client, input *appstream.DescribeStacksInput) (*awstypes.Stack, error) {
	output, err := findStacks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findStacks(ctx context.Context, conn *appstream.Client, input *appstream.DescribeStacksInput) ([]awstypes.Stack, error) {
	var output []awstypes.Stack

	err := describeStacksPages(ctx, conn, input, func(page *appstream.DescribeStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Stacks...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func expandAccessEndpoint(tfMap map[string]any) awstypes.AccessEndpoint {
	if tfMap == nil {
		return awstypes.AccessEndpoint{}
	}

	apiObject := awstypes.AccessEndpoint{
		EndpointType: awstypes.AccessEndpointType(tfMap[names.AttrEndpointType].(string)),
	}

	if v, ok := tfMap["vpce_id"]; ok {
		apiObject.VpceId = aws.String(v.(string))
	}

	return apiObject
}

func expandAccessEndpoints(tfList []any) []awstypes.AccessEndpoint {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.AccessEndpoint

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandAccessEndpoint(tfMap))
	}

	return apiObjects
}

func flattenAccessEndpoint(apiObject awstypes.AccessEndpoint) map[string]any {
	tfMap := map[string]any{}

	tfMap[names.AttrEndpointType] = apiObject.EndpointType
	tfMap["vpce_id"] = aws.ToString(apiObject.VpceId)

	return tfMap
}

func flattenAccessEndpoints(apiObjects []awstypes.AccessEndpoint) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenAccessEndpoint(apiObject))
	}

	return tfList
}

func expandApplicationSettings(tfList []any) *awstypes.ApplicationSettings {
	if len(tfList) == 0 {
		return &awstypes.ApplicationSettings{
			Enabled: aws.Bool(false),
		}
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.ApplicationSettings{
		Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
	}

	if v, ok := tfMap["settings_group"]; ok {
		apiObject.SettingsGroup = aws.String(v.(string))
	}

	return apiObject
}

func flattenApplicationSetting(apiObject *awstypes.ApplicationSettingsResponse) map[string]any {
	if apiObject == nil {
		return nil
	}

	return map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
		"settings_group":  aws.ToString(apiObject.SettingsGroup),
	}
}

func flattenApplicationSettings(apiObject *awstypes.ApplicationSettingsResponse) []any {
	if apiObject == nil {
		return nil
	}

	var tfList []any

	tfList = append(tfList, flattenApplicationSetting(apiObject))

	return tfList
}

func expandStreamingExperienceSettings(tfList []any) *awstypes.StreamingExperienceSettings {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.StreamingExperienceSettings{
		PreferredProtocol: awstypes.PreferredProtocol(tfMap["preferred_protocol"].(string)),
	}

	return apiObject
}

func flattenStreamingExperienceSetting(apiObject *awstypes.StreamingExperienceSettings) map[string]any {
	if apiObject == nil {
		return nil
	}

	return map[string]any{
		"preferred_protocol": apiObject.PreferredProtocol,
	}
}

func flattenStreamingExperienceSettings(apiObject *awstypes.StreamingExperienceSettings) []any {
	if apiObject == nil {
		return nil
	}

	var tfList []any

	tfList = append(tfList, flattenStreamingExperienceSetting(apiObject))

	return tfList
}

func expandStorageConnector(tfMap map[string]any) awstypes.StorageConnector {
	if tfMap == nil {
		return awstypes.StorageConnector{}
	}

	apiObject := awstypes.StorageConnector{
		ConnectorType: awstypes.StorageConnectorType(tfMap["connector_type"].(string)),
	}

	if v, ok := tfMap["domains"]; ok && len(v.([]any)) > 0 {
		apiObject.Domains = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := tfMap["resource_identifier"]; ok && v.(string) != "" {
		apiObject.ResourceIdentifier = aws.String(v.(string))
	}

	return apiObject
}

func expandStorageConnectors(tfList []any) []awstypes.StorageConnector {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.StorageConnector

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandStorageConnector(tfMap))
	}

	return apiObjects
}

func flattenStorageConnector(apiObject awstypes.StorageConnector) map[string]any {
	tfMap := map[string]any{}

	tfMap["connector_type"] = apiObject.ConnectorType
	tfMap["domains"] = apiObject.Domains
	tfMap["resource_identifier"] = aws.ToString(apiObject.ResourceIdentifier)

	return tfMap
}

func flattenStorageConnectors(apiObjects []awstypes.StorageConnector) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenStorageConnector(apiObject))
	}

	return tfList
}

func expandUserSetting(tfMap map[string]any) awstypes.UserSetting {
	if tfMap == nil {
		return awstypes.UserSetting{}
	}

	apiObject := awstypes.UserSetting{
		Action:     awstypes.Action(tfMap[names.AttrAction].(string)),
		Permission: awstypes.Permission(tfMap["permission"].(string)),
	}

	return apiObject
}

func expandUserSettings(tfList []any) []awstypes.UserSetting {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.UserSetting

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandUserSetting(tfMap))
	}

	return apiObjects
}

func flattenUserSetting(apiObject awstypes.UserSetting) map[string]any {
	tfMap := map[string]any{}

	tfMap[names.AttrAction] = apiObject.Action
	tfMap["permission"] = apiObject.Permission

	return tfMap
}

func flattenUserSettings(apiObjects []awstypes.UserSetting) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenUserSetting(apiObject))
	}

	return tfList
}

func suppressAppsStreamStackUserSettings(k, old, new string, d *schema.ResourceData) bool {
	flagDiffUserSettings := false
	count := len(d.Get("user_settings").(*schema.Set).List())
	defaultCount := len(enum.EnumValues[awstypes.Action]())

	if count == defaultCount {
		flagDiffUserSettings = false
	}

	if count != defaultCount && (strconv.Itoa(count) == new && strconv.Itoa(defaultCount) == old) {
		flagDiffUserSettings = true
	}

	return flagDiffUserSettings
}
