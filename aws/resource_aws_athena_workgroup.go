package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
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
													ValidateFunc: validateArn,
												},
											},
										},
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
		Configuration: expandAthenaWorkGroupConfiguration(d.Get("configuration").([]interface{})),
		Name:          aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	// Prevent the below error:
	// InvalidRequestException: Tags provided upon WorkGroup creation must not be empty
	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().AthenaTags()
	}

	_, err := conn.CreateWorkGroup(input)

	if err != nil {
		return fmt.Errorf("error creating Athena WorkGroup: %s", err)
	}

	d.SetId(name)

	if v := d.Get("state").(string); v == athena.WorkGroupStateDisabled {
		input := &athena.UpdateWorkGroupInput{
			State:     aws.String(v),
			WorkGroup: aws.String(d.Id()),
		}

		if _, err := conn.UpdateWorkGroup(input); err != nil {
			return fmt.Errorf("error disabling Athena WorkGroup (%s): %s", d.Id(), err)
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

	if tfawserr.ErrMessageContains(err, athena.ErrCodeInvalidRequestException, "is not found") {
		log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Athena WorkGroup (%s): %s", d.Id(), err)
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

	if err := d.Set("configuration", flattenAthenaWorkGroupConfiguration(resp.WorkGroup.Configuration)); err != nil {
		return fmt.Errorf("error setting configuration: %s", err)
	}

	d.Set("name", resp.WorkGroup.Name)
	d.Set("state", resp.WorkGroup.State)

	if v, ok := d.GetOk("force_destroy"); ok {
		d.Set("force_destroy", v.(bool))
	} else {
		d.Set("force_destroy", false)
	}

	tags, err := tftags.AthenaListTags(conn, arn.String())

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

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
		return fmt.Errorf("error deleting Athena WorkGroup (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceWorkGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	workGroupUpdate := false

	input := &athena.UpdateWorkGroupInput{
		WorkGroup: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("configuration") {
		workGroupUpdate = true
		input.ConfigurationUpdates = expandAthenaWorkGroupConfigurationUpdates(d.Get("configuration").([]interface{}))
	}

	if d.HasChange("description") {
		workGroupUpdate = true
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("state") {
		workGroupUpdate = true
		input.State = aws.String(d.Get("state").(string))
	}

	if workGroupUpdate {
		_, err := conn.UpdateWorkGroup(input)

		if err != nil {
			return fmt.Errorf("error updating Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.AthenaUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceWorkGroupRead(d, meta)
}

func expandAthenaWorkGroupConfiguration(l []interface{}) *athena.WorkGroupConfiguration {
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

	if v, ok := m["publish_cloudwatch_metrics_enabled"]; ok {
		configuration.PublishCloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["result_configuration"]; ok {
		configuration.ResultConfiguration = expandAthenaWorkGroupResultConfiguration(v.([]interface{}))
	}

	if v, ok := m["requester_pays_enabled"]; ok {
		configuration.RequesterPaysEnabled = aws.Bool(v.(bool))
	}

	return configuration
}

func expandAthenaWorkGroupConfigurationUpdates(l []interface{}) *athena.WorkGroupConfigurationUpdates {
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

	if v, ok := m["publish_cloudwatch_metrics_enabled"]; ok {
		configurationUpdates.PublishCloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["result_configuration"]; ok {
		configurationUpdates.ResultConfigurationUpdates = expandAthenaWorkGroupResultConfigurationUpdates(v.([]interface{}))
	}

	if v, ok := m["requester_pays_enabled"]; ok {
		configurationUpdates.RequesterPaysEnabled = aws.Bool(v.(bool))
	}

	return configurationUpdates
}

func expandAthenaWorkGroupResultConfiguration(l []interface{}) *athena.ResultConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	resultConfiguration := &athena.ResultConfiguration{}

	if v, ok := m["encryption_configuration"]; ok {
		resultConfiguration.EncryptionConfiguration = expandAthenaWorkGroupEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := m["output_location"]; ok && v.(string) != "" {
		resultConfiguration.OutputLocation = aws.String(v.(string))
	}

	return resultConfiguration
}

func expandAthenaWorkGroupResultConfigurationUpdates(l []interface{}) *athena.ResultConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	resultConfigurationUpdates := &athena.ResultConfigurationUpdates{}

	if v, ok := m["encryption_configuration"]; ok {
		resultConfigurationUpdates.EncryptionConfiguration = expandAthenaWorkGroupEncryptionConfiguration(v.([]interface{}))
	} else {
		resultConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
	}

	if v, ok := m["output_location"]; ok && v.(string) != "" {
		resultConfigurationUpdates.OutputLocation = aws.String(v.(string))
	} else {
		resultConfigurationUpdates.RemoveOutputLocation = aws.Bool(true)
	}

	return resultConfigurationUpdates
}

func expandAthenaWorkGroupEncryptionConfiguration(l []interface{}) *athena.EncryptionConfiguration {
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

func flattenAthenaWorkGroupConfiguration(configuration *athena.WorkGroupConfiguration) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bytes_scanned_cutoff_per_query":     aws.Int64Value(configuration.BytesScannedCutoffPerQuery),
		"enforce_workgroup_configuration":    aws.BoolValue(configuration.EnforceWorkGroupConfiguration),
		"publish_cloudwatch_metrics_enabled": aws.BoolValue(configuration.PublishCloudWatchMetricsEnabled),
		"result_configuration":               flattenAthenaWorkGroupResultConfiguration(configuration.ResultConfiguration),
		"requester_pays_enabled":             aws.BoolValue(configuration.RequesterPaysEnabled),
	}

	return []interface{}{m}
}

func flattenAthenaWorkGroupResultConfiguration(resultConfiguration *athena.ResultConfiguration) []interface{} {
	if resultConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"encryption_configuration": flattenAthenaWorkGroupEncryptionConfiguration(resultConfiguration.EncryptionConfiguration),
		"output_location":          aws.StringValue(resultConfiguration.OutputLocation),
	}

	return []interface{}{m}
}

func flattenAthenaWorkGroupEncryptionConfiguration(encryptionConfiguration *athena.EncryptionConfiguration) []interface{} {
	if encryptionConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"encryption_option": aws.StringValue(encryptionConfiguration.EncryptionOption),
		"kms_key_arn":       aws.StringValue(encryptionConfiguration.KmsKey),
	}

	return []interface{}{m}
}
