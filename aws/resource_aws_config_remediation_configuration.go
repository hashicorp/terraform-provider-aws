package aws

import (
	//"bytes"
	"fmt"
	"log"
	"time"

	//"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/configservice"
)

func resourceAwsConfigRemediationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigRemediationConfigurationPut,
		Read:   resourceAwsConfigRemediationConfigurationRead,
		Update: resourceAwsConfigRemediationConfigurationPut,
		Delete: resourceAwsConfigRemediationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"config_rule_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"resource_type": {
				Type:     schema.TypeString,
				Computed: false,
			},
			"target_id": {
				Type:     schema.TypeString,
				Computed: false,
			},
			"target_type": {
				Type:     schema.TypeString,
				Computed: false,
			},
			"target_version": {
				Type:     schema.TypeString,
				Computed: false,
			},
			"parameters": {
				Type:     schema.TypeSet,
				MaxItems: 25,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"static_value": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Computed: false,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: false,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsConfigRemediationConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("config_rule_name").(string)
	remediationConfigurationInput := configservice.RemediationConfiguration{
		ConfigRuleName: aws.String(name),
	}

	if v, ok := d.GetOk("parameters"); ok {
		remediationConfigurationInput.Parameters = expandConfigRemediationConfigurationParameters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("resource_type"); ok && v.(string) != "" {
		remediationConfigurationInput.ResourceType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_id"); ok && v.(string) != "" {
		remediationConfigurationInput.TargetId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_type"); ok && v.(string) != "" {
		remediationConfigurationInput.TargetType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_version"); ok && v.(string) != "" {
		remediationConfigurationInput.TargetVersion = aws.String(v.(string))
	}

	remediationConfigurations := make([]*configservice.RemediationConfiguration, 1)
	remediationConfigurations = append(remediationConfigurations, &remediationConfigurationInput)

	input := configservice.PutRemediationConfigurationsInput{
		RemediationConfigurations: remediationConfigurations,
	}
	log.Printf("[DEBUG] Creating AWSConfig remediation configuration: %s", input)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.PutRemediationConfigurations(&input)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InsufficientPermissionsException" {
					// IAM is eventually consistent
					return resource.RetryableError(err)
				}
			}

			return resource.NonRetryableError(fmt.Errorf("Failed to create AWSConfig remediation configuration: %s", err))
		}

		return nil
	})
	if err != nil {
		return err
	}

	d.SetId(name)

	log.Printf("[DEBUG] AWSConfig config remediation configuration for rule %q created", name)

	return resourceAwsConfigRemediationConfigurationRead(d, meta)
}

func resourceAwsConfigRemediationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn
	out, err := conn.DescribeRemediationConfigurations(&configservice.DescribeRemediationConfigurationsInput{
		ConfigRuleNames: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchConfigRuleException" {
			log.Printf("[WARN] Config Rule %q is gone (NoSuchConfigRuleException)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	numberOfRemediationConfigurations := len(out.RemediationConfigurations)
	if numberOfRemediationConfigurations < 1 {
		log.Printf("[WARN] No Remediation Configuration for Config Rule %q (no remediation configuration found)", d.Id())
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] AWS Config remediation configurations received: %s", out)

	remediationConfigurations := out.RemediationConfigurations
	d.Set("remediationConfigurations", flattenRemediationConfigurations(remediationConfigurations))

	return nil
}

func resourceAwsConfigRemediationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("config_rule_name").(string)

	deleteRemediationConfigurationInput := configservice.DeleteRemediationConfigurationInput{
		ConfigRuleName: aws.String(name),
	}

	if v, ok := d.GetOk("resource_type"); ok && v.(string) != "" {
		deleteRemediationConfigurationInput.ResourceType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting AWS Config remediation configurations for rule %q", name)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteRemediationConfiguration(&deleteRemediationConfigurationInput)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceInUseException" {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Deleting Remediation Configurations failed: %s", err)
	}

	log.Printf("[DEBUG] AWS Config remediation configurations for rule %q deleted", name)

	return nil
}

/*func configRuleSourceDetailsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["message_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["event_source"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["maximum_execution_frequency"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}*/
