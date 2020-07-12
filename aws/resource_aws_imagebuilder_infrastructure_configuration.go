package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
)

func resourceAwsImageBuilderInfrastructureConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderInfrastructureConfigurationCreate,
		Read:   resourceAwsImageBuilderInfrastructureConfigurationRead,
		Update: resourceAwsImageBuilderInfrastructureConfigurationUpdate,
		Delete: resourceAwsImageBuilderInfrastructureConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"instance_profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"key_pair": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"logging": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_logs": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_bucket_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"s3_key_prefix": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
										Default:      "/",
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"terminate_instance_on_failure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsImageBuilderInfrastructureConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	securityIdSet := d.Get("security_group_ids").(*schema.Set)
	securityIds := expandStringList(securityIdSet.List())
	instanceIdSet := d.Get("instance_types").(*schema.Set)
	instanceIds := expandStringList(instanceIdSet.List())

	input := &imagebuilder.CreateInfrastructureConfigurationInput{
		ClientToken:         aws.String(resource.UniqueId()),
		Name:                aws.String(d.Get("name").(string)),
		InstanceProfileName: aws.String(d.Get("instance_profile_name").(string)),
		SecurityGroupIds:    securityIds,
		InstanceTypes:       instanceIds,
		Logging:             expandAwsImageBuilderLogsConfig(d),
	}

	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("key_pair"); ok {
		input.SetKeyPair(v.(string))
	}
	if v, ok := d.GetOk("sns_topic_arn"); ok {
		input.SetSnsTopicArn(v.(string))
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		input.SetSubnetId(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags())
	}

	log.Printf("[DEBUG] Creating Infrastructure Configuration: %#v", input)

	resp, err := conn.CreateInfrastructureConfiguration(input)
	if err != nil {
		return fmt.Errorf("error creating Component: %s", err)
	}

	d.SetId(aws.StringValue(resp.InfrastructureConfigurationArn))

	return resourceAwsImageBuilderInfrastructureConfigurationRead(d, meta)
}

func resourceAwsImageBuilderInfrastructureConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetInfrastructureConfiguration(&imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	})

	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Infrastructure Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Infrastructure Configuration (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.InfrastructureConfiguration.Arn)
	d.Set("date_created", resp.InfrastructureConfiguration.DateCreated)
	d.Set("date_updated", resp.InfrastructureConfiguration.DateUpdated)
	d.Set("description", resp.InfrastructureConfiguration.Description)
	d.Set("instance_profile_name", resp.InfrastructureConfiguration.InstanceProfileName)
	d.Set("instance_types", resp.InfrastructureConfiguration.InstanceTypes)
	d.Set("key_pair", resp.InfrastructureConfiguration.KeyPair)
	d.Set("logging", flattenAwsImageBuilderLogsConfig(resp.InfrastructureConfiguration.Logging))
	d.Set("name", resp.InfrastructureConfiguration.Name)
	d.Set("security_group_ids", resp.InfrastructureConfiguration.SecurityGroupIds)
	d.Set("sns_topic_arn", resp.InfrastructureConfiguration.SnsTopicArn)
	d.Set("subnet_id", resp.InfrastructureConfiguration.SubnetId)
	d.Set("terminate_instance_on_failure", resp.InfrastructureConfiguration.TerminateInstanceOnFailure)

	tags, err := keyvaluetags.ImagebuilderListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Infrastructure Configuration (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsImageBuilderInfrastructureConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	// Despite not being required by the API, if it's not sent then certain things get wiped to null
	upd := imagebuilder.UpdateInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
		InstanceProfileName:            aws.String(d.Get("instance_profile_name").(string)),
		Logging:                        expandAwsImageBuilderLogsConfig(d),
	}

	if description, ok := d.GetOk("description"); ok {
		upd.Description = aws.String(description.(string))
	}

	if key_pair, ok := d.GetOk("key_pair"); ok {
		upd.KeyPair = aws.String(key_pair.(string))
	}

	if subnet_id, ok := d.GetOk("subnet_id"); ok {
		upd.SubnetId = aws.String(subnet_id.(string))
	}

	if attr := d.Get("instance_types").(*schema.Set); attr.Len() > 0 {
		upd.InstanceTypes = expandStringList(attr.List())
	}

	if d.HasChange("instance_profile_name") {
		upd.InstanceProfileName = aws.String(d.Get("instance_profile_name").(string))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		if attr := v.(*schema.Set); attr.Len() > 0 {
			upd.SecurityGroupIds = expandStringList(attr.List())
		}
	}

	if sns_topic_arn, ok := d.GetOk("sns_topic_arn"); ok {
		upd.SnsTopicArn = aws.String(sns_topic_arn.(string))
	}

	if terminate_instance_on_failure, ok := d.GetOk("terminate_instance_on_failure"); ok {
		upd.TerminateInstanceOnFailure = aws.Bool(terminate_instance_on_failure.(bool))
	}

	if d.HasChange("security_group_ids") {
		if attr := d.Get("security_group_ids").(*schema.Set); attr.Len() > 0 {
			upd.SecurityGroupIds = expandStringList(attr.List())
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for Infrastructure Configuration (%s): %s", d.Id(), err)
		}
	}

	_, err := conn.UpdateInfrastructureConfiguration(&upd)
	if err != nil {
		return fmt.Errorf("error updating Infrastructure Configuration (%s): %s", d.Id(), err)
	}

	return resourceAwsImageBuilderInfrastructureConfigurationRead(d, meta)
}

func resourceAwsImageBuilderInfrastructureConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	_, err := conn.DeleteInfrastructureConfiguration(&imagebuilder.DeleteInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Infrastructure Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func flattenAwsImageBuilderLogsConfig(logsConfig *imagebuilder.Logging) []interface{} {
	if logsConfig == nil || logsConfig.S3Logs == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	// If more logging options are added, add more ifs!
	values["s3_logs"] = flattenAwsImageBuilderS3Logs(logsConfig.S3Logs)

	return []interface{}{values}
}

func flattenAwsImageBuilderS3Logs(s3LogsConfig *imagebuilder.S3Logs) []interface{} {
	values := map[string]interface{}{}

	values["s3_key_prefix"] = aws.StringValue(s3LogsConfig.S3KeyPrefix)
	values["s3_bucket_name"] = aws.StringValue(s3LogsConfig.S3BucketName)

	return []interface{}{values}
}

func expandAwsImageBuilderLogsConfig(d *schema.ResourceData) *imagebuilder.Logging {
	logsConfig := &imagebuilder.Logging{}

	if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		configList := v.([]interface{})
		data := configList[0].(map[string]interface{})

		if v, ok := data["s3_logs"]; ok {
			logsConfig.S3Logs = expandAwsImageBuilderS3LogsConfig(v.([]interface{}))
		}
	}

	return logsConfig
}

func expandAwsImageBuilderS3LogsConfig(configList []interface{}) *imagebuilder.S3Logs {
	if len(configList) == 0 || configList[0] == nil {
		return nil
	}

	data := configList[0].(map[string]interface{})

	s3LogsConfig := &imagebuilder.S3Logs{}

	if v, ok := data["s3_bucket_name"]; ok {
		s3LogsConfig.S3BucketName = aws.String(v.(string))
	}
	if v, ok := data["s3_key_prefix"]; ok {
		s3LogsConfig.S3KeyPrefix = aws.String(v.(string))
	}

	return s3LogsConfig
}
