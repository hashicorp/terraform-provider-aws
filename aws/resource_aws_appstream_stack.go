package aws

import (
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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appstream/waiter"
)

var (
	flagDiffUserSettings = false
)

func resourceAwsAppStreamStack() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsAppStreamStackCreate,
		ReadWithoutTimeout:   resourceAwsAppStreamStackRead,
		UpdateWithoutTimeout: resourceAwsAppStreamStackUpdate,
		DeleteWithoutTimeout: resourceAwsAppStreamStackDelete,
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
				Type:     schema.TypeList,
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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
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
			},
			"tags":     tagsSchemaForceNew(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsAppStreamStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn
	input := &appstream.CreateStackInput{
		Name: aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))),
	}

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

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
		input.EmbedHostDomains = expandStringList(v.([]interface{}))
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

	if isResourceTimeoutError(err) {
		output, err = conn.CreateStackWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Appstream Stack (%s): %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.Stack.Name))

	return resourceAwsAppStreamStackRead(ctx, d, meta)
}

func resourceAwsAppStreamStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

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
		if err = d.Set("embed_host_domains", flattenStringList(v.EmbedHostDomains)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "user_settings", d.Id(), err))
		}
		d.Set("feedback_url", v.FeedbackURL)
		d.Set("name", v.Name)
		d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(v.Name)))
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

		tags := keyvaluetags.AppstreamKeyValueTags(tg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

		if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "tags", d.Id(), err))
		}

		if err = d.Set("tags_all", tags.Map()); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "tags_all", d.Id(), err))
		}
	}
	return nil
}

func resourceAwsAppStreamStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

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
		if err := keyvaluetags.AppstreamUpdateTags(conn, arn, o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Appstream Stack tags (%s): %w", d.Id(), err))
		}
	}

	return resourceAwsAppStreamStackRead(ctx, d, meta)
}

func resourceAwsAppStreamStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	_, err := conn.DeleteStackWithContext(ctx, &appstream.DeleteStackInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Appstream Stack (%s): %w", d.Id(), err))
	}

	// Will wait to finish to delete because after delete it makes a stack inactive then it deletes
	time.Sleep(waiter.StackSleep)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream Stack (%s): %w", d.Id(), err))
	}

	return nil
}

func expandAccessEndpoints(tfList []interface{}) []*appstream.AccessEndpoint {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*appstream.AccessEndpoint

	for _, v := range tfList {
		v1 := v.(map[string]interface{})

		endpoint := &appstream.AccessEndpoint{
			EndpointType: aws.String(v1["endpoint_type"].(string)),
		}
		if v2, ok := v1["vpce_id"]; ok {
			endpoint.VpceId = aws.String(v2.(string))
		}

		apiObjects = append(apiObjects, endpoint)
	}

	return apiObjects
}

func flattenAccessEndpoints(apiObjects []*appstream.AccessEndpoint) []map[string]interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []map[string]interface{}

	for _, endpoint := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"endpoint_type": aws.StringValue(endpoint.EndpointType),
			"vpce_id":       aws.StringValue(endpoint.VpceId),
		})
	}

	return tfList
}

func expandApplicationSettings(tfList []interface{}) *appstream.ApplicationSettings {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &appstream.ApplicationSettings{}

	attr := tfList[0].(map[string]interface{})
	if v, ok := attr["enabled"]; ok {
		apiObject.Enabled = aws.Bool(v.(bool))
	}
	if v, ok := attr["settings_group"]; ok {
		apiObject.SettingsGroup = aws.String(v.(string))
	}

	return apiObject
}

func flattenApplicationSettings(apiObject *appstream.ApplicationSettingsResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := map[string]interface{}{}
	tfList["enabled"] = aws.BoolValue(apiObject.Enabled)
	tfList["settings_group"] = aws.StringValue(apiObject.SettingsGroup)

	return []interface{}{tfList}
}

func expandStorageConnectors(tfList []interface{}) []*appstream.StorageConnector {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*appstream.StorageConnector

	for _, v := range tfList {
		v1 := v.(map[string]interface{})

		connector := &appstream.StorageConnector{
			ConnectorType: aws.String(v1["connector_type"].(string)),
		}
		if v2, ok := v1["domains"]; ok && len(v2.([]interface{})) > 0 {
			connector.Domains = expandStringList(v2.([]interface{}))
		}
		if v2, ok := v1["resource_identifier"]; ok && v2.(string) != "" {
			connector.ResourceIdentifier = aws.String(v2.(string))
		}

		apiObjects = append(apiObjects, connector)
	}

	return apiObjects
}

func flattenStorageConnectors(apiObjects []*appstream.StorageConnector) []map[string]interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []map[string]interface{}

	for _, connector := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"connector_type":      aws.StringValue(connector.ConnectorType),
			"domains":             aws.StringValueSlice(connector.Domains),
			"resource_identifier": aws.StringValue(connector.ResourceIdentifier),
		})
	}

	return tfList
}

func expandUserSettings(tfList []interface{}) []*appstream.UserSetting {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*appstream.UserSetting

	for _, v := range tfList {
		v1 := v.(map[string]interface{})

		user := &appstream.UserSetting{
			Action:     aws.String(v1["action"].(string)),
			Permission: aws.String(v1["permission"].(string)),
		}

		apiObjects = append(apiObjects, user)
	}

	return apiObjects
}

func flattenUserSettings(apiObjects []*appstream.UserSetting) []map[string]interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []map[string]interface{}

	for _, user := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"action":     aws.StringValue(user.Action),
			"permission": aws.StringValue(user.Permission),
		})
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
