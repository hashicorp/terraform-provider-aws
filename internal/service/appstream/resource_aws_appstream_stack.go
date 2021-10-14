package aws

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/appstream/waiter"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var (
	flagDiffUserSettings = false
)

func ResourceStack() *schema.Resource {
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
						"endpoint_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appstream.AccessEndpointType_Values(), false),
						},
						"vpce_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Set: accessEndpointsHash,
			},
			"application_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"settings_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
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
				Set: schema.HashString,
			},
			"feedback_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"redirect_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"storage_connectors": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appstream.StorageConnectorType_Values(), false),
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
				Set: storageConnectorsHash,
			},
			"user_settings": {
				Type:             schema.TypeSet,
				Optional:         true,
				Computed:         true,
				MinItems:         1,
				DiffSuppressFunc: suppressAppsStreamStackUserSettings,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appstream.Action_Values(), false),
						},
						"permission": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appstream.Permission_Values(), false),
						},
					},
				},
				Set: userSettingsHash,
			},
			"tags":     tftags.TagsSchemaForceNew(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn
	input := &appstream.CreateStackInput{
		Name: aws.String(d.Get("name").(string)),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if v, ok := d.GetOk("access_endpoints"); ok {
		input.AccessEndpoints = expandAccessEndpoints(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("application_settings"); ok {
		input.ApplicationSettings = expandApplicationSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("embed_host_domains"); ok {
		input.EmbedHostDomains = flex.ExpandStringList(v.([]interface{}))
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

	if v, ok := d.GetOk("user_settings"); ok {
		input.UserSettings = expandUserSettings(v.(*schema.Set).List())
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().AppstreamTags()
	}

	var err error
	var output *appstream.CreateStackOutput
	err = resource.RetryContext(ctx, waiter.StackOperationTimeout, func() *resource.RetryError {
		output, err = conn.CreateStackWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateStackWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Appstream Stack (%s): %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.Stack.Name))

	return resourceStackRead(ctx, d, meta)
}

func resourceStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeStacksWithContext(ctx, &appstream.DescribeStacksInput{Names: []*string{aws.String(d.Id())}})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appstream Stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream Stack (%s): %w", d.Id(), err))
	}
	for _, v := range resp.Stacks {

		if err = d.Set("access_endpoints", flattenAccessEndpoints(v.AccessEndpoints)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "access_endpoints", d.Id(), err))
		}
		if err = d.Set("application_settings", flattenApplicationSettings(v.ApplicationSettings)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "user_settings", d.Id(), err))
		}
		d.Set("arn", v.Arn)
		d.Set("created_time", aws.TimeValue(v.CreatedTime).Format(time.RFC3339))
		d.Set("description", v.Description)
		d.Set("display_name", v.DisplayName)
		if err = d.Set("embed_host_domains", flex.FlattenStringList(v.EmbedHostDomains)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "user_settings", d.Id(), err))
		}
		d.Set("feedback_url", v.FeedbackURL)
		d.Set("name", v.Name)
		d.Set("redirect_url", v.RedirectURL)
		if err = d.Set("storage_connectors", flattenStorageConnectors(v.StorageConnectors)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "storage_connectors", d.Id(), err))
		}
		if err = d.Set("user_settings", flattenUserSettings(v.UserSettings)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "user_settings", d.Id(), err))
		}

		tg, err := conn.ListTagsForResource(&appstream.ListTagsForResourceInput{
			ResourceArn: v.Arn,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing stack tags for AppStream Stack (%s): %w", d.Id(), err))
		}

		tags := tftags.AppstreamKeyValueTags(tg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

		if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "tags", d.Id(), err))
		}

		if err = d.Set("tags_all", tags.Map()); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "tags_all", d.Id(), err))
		}
	}
	return nil
}

func resourceStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	input := &appstream.UpdateStackInput{
		Name: aws.String(d.Id()),
	}

	if d.HasChange("access_endpoints") {
		input.AccessEndpoints = expandAccessEndpoints(d.Get("access_endpoints").(*schema.Set).List())
	}

	if d.HasChange("application_settings") {
		input.ApplicationSettings = expandApplicationSettings(d.Get("application_settings").(*schema.Set).List())
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("display_name") {
		input.DisplayName = aws.String(d.Get("display_name").(string))
	}

	if d.HasChange("feedback_url") {
		input.FeedbackURL = aws.String(d.Get("feedback_url").(string))
	}

	if d.HasChange("redirect_url") {
		input.RedirectURL = aws.String(d.Get("redirect_url").(string))
	}

	if d.HasChange("user_settings") {
		input.UserSettings = expandUserSettings(d.Get("user_settings").([]interface{}))
	}

	resp, err := conn.UpdateStack(input)

	if err != nil {
		diag.FromErr(fmt.Errorf("error updating Appstream Stack (%s): %w", d.Id(), err))
	}

	if d.HasChange("tags") {
		arn := aws.StringValue(resp.Stack.Arn)

		o, n := d.GetChange("tags")
		if err := tftags.AppstreamUpdateTags(conn, arn, o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Appstream Stack tags (%s): %w", d.Id(), err))
		}
	}

	return resourceStackRead(ctx, d, meta)
}

func resourceStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	_, err := conn.DeleteStackWithContext(ctx, &appstream.DeleteStackInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Appstream Stack (%s): %w", d.Id(), err))
	}

	if _, err = waiter.StackStateDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}

		return diag.FromErr(fmt.Errorf("error waiting for Appstream Stack (%s) to be deleted: %w", d.Id(), err))
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream Stack (%s): %w", d.Id(), err))
	}

	return nil
}

func expandAccessEndpoint(tfMap map[string]interface{}) *appstream.AccessEndpoint {
	if tfMap == nil {
		return nil
	}

	apiObject := &appstream.AccessEndpoint{
		EndpointType: aws.String(tfMap["endpoint_type"].(string)),
	}
	if v, ok := tfMap["vpce_id"]; ok {
		apiObject.VpceId = aws.String(v.(string))
	}

	return apiObject
}

func expandAccessEndpoints(tfList []interface{}) []*appstream.AccessEndpoint {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*appstream.AccessEndpoint

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandAccessEndpoint(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenAccessEndpoint(apiObject *appstream.AccessEndpoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["endpoint_type"] = aws.StringValue(apiObject.EndpointType)
	tfMap["vpce_id"] = aws.StringValue(apiObject.VpceId)

	return tfMap
}

func flattenAccessEndpoints(apiObjects []*appstream.AccessEndpoint) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenAccessEndpoint(apiObject))
	}

	return tfList
}

func expandApplicationSetting(tfMap map[string]interface{}) *appstream.ApplicationSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &appstream.ApplicationSettings{}

	if v, ok := tfMap["enabled"]; ok {
		apiObject.Enabled = aws.Bool(v.(bool))
	}
	if v, ok := tfMap["settings_group"]; ok {
		apiObject.SettingsGroup = aws.String(v.(string))
	}

	return apiObject
}

func expandApplicationSettings(tfList []interface{}) *appstream.ApplicationSettings {
	if len(tfList) == 0 {
		return nil
	}

	var apiObject *appstream.ApplicationSettings

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject = expandApplicationSetting(tfMap)
	}

	return apiObject
}

func flattenApplicationSetting(apiObject *appstream.ApplicationSettingsResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["enabled"] = aws.BoolValue(apiObject.Enabled)
	tfMap["settings_group"] = aws.StringValue(apiObject.SettingsGroup)

	return tfMap
}

func flattenApplicationSettings(apiObject *appstream.ApplicationSettingsResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}

	tfList = append(tfList, flattenApplicationSetting(apiObject))

	return tfList
}

func expandStorageConnector(tfMap map[string]interface{}) *appstream.StorageConnector {
	if tfMap == nil {
		return nil
	}

	apiObject := &appstream.StorageConnector{
		ConnectorType: aws.String(tfMap["connector_type"].(string)),
	}
	if v, ok := tfMap["domains"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Domains = flex.ExpandStringList(v.([]interface{}))
	}
	if v, ok := tfMap["resource_identifier"]; ok && v.(string) != "" {
		apiObject.ResourceIdentifier = aws.String(v.(string))
	}

	return apiObject
}

func expandStorageConnectors(tfList []interface{}) []*appstream.StorageConnector {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*appstream.StorageConnector

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandStorageConnector(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenStorageConnector(apiObject *appstream.StorageConnector) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["connector_type"] = aws.StringValue(apiObject.ConnectorType)
	tfMap["domains"] = aws.StringValueSlice(apiObject.Domains)
	tfMap["resource_identifier"] = aws.StringValue(apiObject.ResourceIdentifier)

	return tfMap
}

func flattenStorageConnectors(apiObjects []*appstream.StorageConnector) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenStorageConnector(apiObject))
	}

	return tfList
}

func expandUserSetting(tfMap map[string]interface{}) *appstream.UserSetting {
	if tfMap == nil {
		return nil
	}

	apiObject := &appstream.UserSetting{
		Action:     aws.String(tfMap["action"].(string)),
		Permission: aws.String(tfMap["permission"].(string)),
	}

	return apiObject
}

func expandUserSettings(tfList []interface{}) []*appstream.UserSetting {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*appstream.UserSetting

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandUserSetting(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenUserSetting(apiObject *appstream.UserSetting) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["action"] = aws.StringValue(apiObject.Action)
	tfMap["permission"] = aws.StringValue(apiObject.Permission)

	return tfMap
}

func flattenUserSettings(apiObjects []*appstream.UserSetting) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}
		tfList = append(tfList, flattenUserSetting(apiObject))
	}

	return tfList
}

func suppressAppsStreamStackUserSettings(k, old, new string, d *schema.ResourceData) bool {
	count := len(d.Get("user_settings").(*schema.Set).List())
	defaultCount := len(appstream.Action_Values())

	if count == defaultCount {
		flagDiffUserSettings = false
	}

	if count != defaultCount && (fmt.Sprintf("%d", count) == new && fmt.Sprintf("%d", defaultCount) == old) {
		flagDiffUserSettings = true
	}

	return flagDiffUserSettings
}

func accessEndpointsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["endpoint_type"].(string))
	buf.WriteString(m["vpce_id"].(string))
	return create.StringHashcode(buf.String())
}

func storageConnectorsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["connector_type"].(string))
	buf.WriteString(fmt.Sprintf("%+v", m["domains"].([]interface{})))
	buf.WriteString(m["resource_identifier"].(string))
	return create.StringHashcode(buf.String())
}

func userSettingsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["action"].(string))
	buf.WriteString(m["permission"].(string))
	return create.StringHashcode(buf.String())
}
