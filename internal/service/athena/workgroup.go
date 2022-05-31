package athena

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceWorkGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkGroupCreate,
		Read:   resourceWorkGroupRead,
		Update: resourceWorkGroupUpdate,
		Delete: resourceWorkGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bytes_scanned_cutoff_per_query": {
							Type:     schema.TypeInt,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.IntAtLeast(10485760),
								validation.IntInSlice([]int{0}),
							),
						},
						"enforce_workgroup_configuration": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"engine_version": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"effective_engine_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"selected_engine_version": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "AUTO",
									},
								},
							},
						},
						"publish_cloudwatch_metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"result_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"acl_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_acl_option": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(athena.S3AclOption_Values(), false),
												},
											},
										},
									},
									"encryption_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_option": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(athena.EncryptionOption_Values(), false),
												},
												"kms_key_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"expected_bucket_owner": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"output_location": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"requester_pays_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._-]+$`), "must contain only alphanumeric characters, periods, underscores, and hyphens"),
				),
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      athena.WorkGroupStateEnabled,
				ValidateFunc: validation.StringInSlice(athena.WorkGroupState_Values(), false),
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &athena.CreateWorkGroupInput{
		Configuration: expandWorkGroupConfiguration(d.Get("configuration").([]interface{})),
		Name:          aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	// Prevent the below error:
	// InvalidRequestException: Tags provided upon WorkGroup creation must not be empty
	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.CreateWorkGroup(input)

	if err != nil {
		return fmt.Errorf("error creating Athena WorkGroup: %w", err)
	}

	d.SetId(name)

	if v := d.Get("state").(string); v == athena.WorkGroupStateDisabled {
		input := &athena.UpdateWorkGroupInput{
			State:     aws.String(v),
			WorkGroup: aws.String(d.Id()),
		}

		if _, err := conn.UpdateWorkGroup(input); err != nil {
			return fmt.Errorf("error disabling Athena WorkGroup (%s): %w", d.Id(), err)
		}
	}

	return resourceWorkGroupRead(d, meta)
}

func resourceWorkGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &athena.GetWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}

	resp, err := conn.GetWorkGroup(input)

	if tfawserr.ErrMessageContains(err, athena.ErrCodeInvalidRequestException, "is not found") && !d.IsNewResource() {
		log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Athena WorkGroup (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "athena",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("workgroup/%s", d.Id()),
	}

	d.Set("arn", arn.String())
	d.Set("description", resp.WorkGroup.Description)

	if err := d.Set("configuration", flattenWorkGroupConfiguration(resp.WorkGroup.Configuration)); err != nil {
		return fmt.Errorf("error setting configuration: %w", err)
	}

	d.Set("name", resp.WorkGroup.Name)
	d.Set("state", resp.WorkGroup.State)

	if v, ok := d.GetOk("force_destroy"); ok {
		d.Set("force_destroy", v.(bool))
	} else {
		d.Set("force_destroy", false)
	}

	tags, err := ListTags(conn, arn.String())

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceWorkGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	input := &athena.DeleteWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("force_destroy"); ok {
		input.RecursiveDeleteOption = aws.Bool(v.(bool))
	}
	_, err := conn.DeleteWorkGroup(input)

	if err != nil {
		return fmt.Errorf("error deleting Athena WorkGroup (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceWorkGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &athena.UpdateWorkGroupInput{
			WorkGroup: aws.String(d.Get("name").(string)),
		}

		if d.HasChange("configuration") {
			input.ConfigurationUpdates = expandWorkGroupConfigurationUpdates(d.Get("configuration").([]interface{}))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("state") {
			input.State = aws.String(d.Get("state").(string))
		}
		_, err := conn.UpdateWorkGroup(input)

		if err != nil {
			return fmt.Errorf("error updating Athena WorkGroup (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceWorkGroupRead(d, meta)
}

func expandWorkGroupConfiguration(l []interface{}) *athena.WorkGroupConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &athena.WorkGroupConfiguration{}

	if v, ok := m["bytes_scanned_cutoff_per_query"]; ok && v.(int) > 0 {
		configuration.BytesScannedCutoffPerQuery = aws.Int64(int64(v.(int)))
	}

	if v, ok := m["enforce_workgroup_configuration"]; ok {
		configuration.EnforceWorkGroupConfiguration = aws.Bool(v.(bool))
	}

	if v, ok := m["engine_version"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		configuration.EngineVersion = expandWorkGroupEngineVersion(v)
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"]; ok {
		configuration.PublishCloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["result_configuration"]; ok {
		configuration.ResultConfiguration = expandWorkGroupResultConfiguration(v.([]interface{}))
	}

	if v, ok := m["requester_pays_enabled"]; ok {
		configuration.RequesterPaysEnabled = aws.Bool(v.(bool))
	}

	return configuration
}

func expandWorkGroupEngineVersion(l []interface{}) *athena.EngineVersion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	engineVersion := &athena.EngineVersion{}

	if v, ok := m["selected_engine_version"].(string); ok && v != "" {
		engineVersion.SelectedEngineVersion = aws.String(v)
	}

	return engineVersion
}

func expandWorkGroupConfigurationUpdates(l []interface{}) *athena.WorkGroupConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configurationUpdates := &athena.WorkGroupConfigurationUpdates{}

	if v, ok := m["bytes_scanned_cutoff_per_query"]; ok && v.(int) > 0 {
		configurationUpdates.BytesScannedCutoffPerQuery = aws.Int64(int64(v.(int)))
	} else {
		configurationUpdates.RemoveBytesScannedCutoffPerQuery = aws.Bool(true)
	}

	if v, ok := m["enforce_workgroup_configuration"]; ok {
		configurationUpdates.EnforceWorkGroupConfiguration = aws.Bool(v.(bool))
	}

	if v, ok := m["engine_version"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		configurationUpdates.EngineVersion = expandWorkGroupEngineVersion(v)
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"]; ok {
		configurationUpdates.PublishCloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["result_configuration"]; ok {
		configurationUpdates.ResultConfigurationUpdates = expandWorkGroupResultConfigurationUpdates(v.([]interface{}))
	}

	if v, ok := m["requester_pays_enabled"]; ok {
		configurationUpdates.RequesterPaysEnabled = aws.Bool(v.(bool))
	}

	return configurationUpdates
}

func expandWorkGroupResultConfiguration(l []interface{}) *athena.ResultConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	resultConfiguration := &athena.ResultConfiguration{}

	if v, ok := m["encryption_configuration"]; ok {
		resultConfiguration.EncryptionConfiguration = expandWorkGroupEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := m["output_location"].(string); ok && v != "" {
		resultConfiguration.OutputLocation = aws.String(v)
	}

	if v, ok := m["expected_bucket_owner"].(string); ok && v != "" {
		resultConfiguration.ExpectedBucketOwner = aws.String(v)
	}

	if v, ok := m["acl_configuration"]; ok {
		resultConfiguration.AclConfiguration = expandResultConfigurationAclConfig(v.([]interface{}))
	}

	return resultConfiguration
}

func expandWorkGroupResultConfigurationUpdates(l []interface{}) *athena.ResultConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	resultConfigurationUpdates := &athena.ResultConfigurationUpdates{}

	if v, ok := m["encryption_configuration"]; ok {
		resultConfigurationUpdates.EncryptionConfiguration = expandWorkGroupEncryptionConfiguration(v.([]interface{}))
	} else {
		resultConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
	}

	if v, ok := m["output_location"].(string); ok && v != "" {
		resultConfigurationUpdates.OutputLocation = aws.String(v)
	} else {
		resultConfigurationUpdates.RemoveOutputLocation = aws.Bool(true)
	}

	if v, ok := m["expected_bucket_owner"].(string); ok && v != "" {
		resultConfigurationUpdates.ExpectedBucketOwner = aws.String(v)
	} else {
		resultConfigurationUpdates.RemoveExpectedBucketOwner = aws.Bool(true)
	}

	if v, ok := m["acl_configuration"]; ok {
		resultConfigurationUpdates.AclConfiguration = expandResultConfigurationAclConfig(v.([]interface{}))
	} else {
		resultConfigurationUpdates.RemoveAclConfiguration = aws.Bool(true)
	}

	return resultConfigurationUpdates
}

func expandWorkGroupEncryptionConfiguration(l []interface{}) *athena.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	encryptionConfiguration := &athena.EncryptionConfiguration{}

	if v, ok := m["encryption_option"]; ok && v.(string) != "" {
		encryptionConfiguration.EncryptionOption = aws.String(v.(string))
	}

	if v, ok := m["kms_key_arn"]; ok && v.(string) != "" {
		encryptionConfiguration.KmsKey = aws.String(v.(string))
	}

	return encryptionConfiguration
}

func flattenWorkGroupConfiguration(configuration *athena.WorkGroupConfiguration) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bytes_scanned_cutoff_per_query":     aws.Int64Value(configuration.BytesScannedCutoffPerQuery),
		"enforce_workgroup_configuration":    aws.BoolValue(configuration.EnforceWorkGroupConfiguration),
		"engine_version":                     flattenWorkGroupEngineVersion(configuration.EngineVersion),
		"publish_cloudwatch_metrics_enabled": aws.BoolValue(configuration.PublishCloudWatchMetricsEnabled),
		"result_configuration":               flattenWorkGroupResultConfiguration(configuration.ResultConfiguration),
		"requester_pays_enabled":             aws.BoolValue(configuration.RequesterPaysEnabled),
	}

	return []interface{}{m}
}

func flattenWorkGroupEngineVersion(engineVersion *athena.EngineVersion) []interface{} {
	if engineVersion == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"effective_engine_version": aws.StringValue(engineVersion.EffectiveEngineVersion),
		"selected_engine_version":  aws.StringValue(engineVersion.SelectedEngineVersion),
	}

	return []interface{}{m}
}

func flattenWorkGroupResultConfiguration(resultConfiguration *athena.ResultConfiguration) []interface{} {
	if resultConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"encryption_configuration": flattenWorkGroupEncryptionConfiguration(resultConfiguration.EncryptionConfiguration),
		"output_location":          aws.StringValue(resultConfiguration.OutputLocation),
	}

	if resultConfiguration.ExpectedBucketOwner != nil {
		m["expected_bucket_owner"] = aws.StringValue(resultConfiguration.ExpectedBucketOwner)
	}

	if resultConfiguration.AclConfiguration != nil {
		m["acl_configuration"] = flattenWorkGroupAclConfiguration(resultConfiguration.AclConfiguration)
	}

	return []interface{}{m}
}

func flattenWorkGroupEncryptionConfiguration(encryptionConfiguration *athena.EncryptionConfiguration) []interface{} {
	if encryptionConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"encryption_option": aws.StringValue(encryptionConfiguration.EncryptionOption),
		"kms_key_arn":       aws.StringValue(encryptionConfiguration.KmsKey),
	}

	return []interface{}{m}
}

func flattenWorkGroupAclConfiguration(aclConfig *athena.AclConfiguration) []interface{} {
	if aclConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"s3_acl_option": aws.StringValue(aclConfig.S3AclOption),
	}

	return []interface{}{m}
}
