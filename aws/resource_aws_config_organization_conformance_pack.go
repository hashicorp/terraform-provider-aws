package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"time"
)

func resourceAwsConfigOrganizationConformancePack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigOrganizationConformancePackPut,
		Read:   resourceAwsConfigOrganizationConformancePackRead,
		Update: resourceAwsConfigOrganizationConformancePackPut,
		Delete: resourceAwsConfigOrganizationConformancePackDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 51200),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][-a-zA-Z0-9]*$`), "must be a valid conformance pack name"),
				),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template_s3_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"template_body": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 51200),
					validateStringIsJsonOrYaml),
				StateFunc: func(v interface{}) string {
					template, _ := normalizeJsonOrYamlString(v)
					return template
				},
			},
			"delivery_s3_bucket": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile("awsconfigconforms.+"), "must start with 'awsconfigconforms'"),
				),
			},
			"delivery_s3_key_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"input_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"excluded_accounts": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[0-9]{12}$`), "must be a valid AWS account ID"),
				},
				Optional: true,
			},
		},
	}
}

func resourceAwsConfigOrganizationConformancePackPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("name").(string)
	input := configservice.PutOrganizationConformancePackInput{
		OrganizationConformancePackName: aws.String(name),
	}

	if v, ok := d.GetOk("delivery_s3_bucket"); ok {
		input.DeliveryS3Bucket = aws.String(v.(string))
	}
	if v, ok := d.GetOk("delivery_s3_key_prefix"); ok {
		input.DeliveryS3KeyPrefix = aws.String(v.(string))
	}
	if v, ok := d.GetOk("excluded_accounts"); ok {
		input.ExcludedAccounts = expandConfigConformancePackExcludedAccounts(v.([]interface{}))
	}
	if v, ok := d.GetOk("input_parameters"); ok {
		input.ConformancePackInputParameters = expandConfigConformancePackParameters(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}
	if v, ok := d.GetOk("template_s3_uri"); ok {
		input.TemplateS3Uri = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConformancePack(&input)
	if err != nil {
		return fmt.Errorf("failed to put AWSConfig organization conformance pack %q: %s", name, err)
	}

	d.SetId(name)
	conf := resource.StateChangeConf{
		Pending: []string{
			configservice.OrganizationResourceDetailedStatusCreateInProgress,
			configservice.OrganizationResourceDetailedStatusUpdateInProgress,
		},
		Target: []string{
			configservice.OrganizationResourceDetailedStatusCreateSuccessful,
			configservice.OrganizationResourceDetailedStatusUpdateSuccessful,
		},
		Timeout: 30 * time.Minute,
		Refresh: refreshOrganizationConformancePackStatus(d, conn),
	}
	if _, err := conf.WaitForState(); err != nil {
		return err
	}
	return resourceAwsConfigOrganizationConformancePackRead(d, meta)
}

func expandConfigConformancePackExcludedAccounts(i []interface{}) (ret []*string) {
	for _, v := range i {
		ret = append(ret, aws.String(v.(string)))
	}
	return
}

func expandConfigConformancePackParameters(m map[string]interface{}) (params []*configservice.ConformancePackInputParameter) {
	for k, v := range m {
		params = append(params, &configservice.ConformancePackInputParameter{
			ParameterName:  aws.String(k),
			ParameterValue: aws.String(v.(string)),
		})
	}
	return
}

func refreshOrganizationConformancePackStatus(d *schema.ResourceData, conn *configservice.ConfigService) func() (interface{}, string, error) {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeOrganizationConformancePackStatuses(&configservice.DescribeOrganizationConformancePackStatusesInput{
			OrganizationConformancePackNames: []*string{aws.String(d.Id())},
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && isAWSErr(awsErr, configservice.ErrCodeNoSuchOrganizationConformancePackException, "") {
				return 42, "", nil
			}
			return 42, "", fmt.Errorf("failed to describe organization conformance pack %q: %s", d.Id(), err)
		}
		if len(out.OrganizationConformancePackStatuses) < 1 {
			return 42, "", nil
		}
		status := out.OrganizationConformancePackStatuses[0]
		return out, *status.Status, nil
	}
}

func resourceAwsConfigOrganizationConformancePackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	out, err := conn.DescribeOrganizationConformancePacks(&configservice.DescribeOrganizationConformancePacksInput{
		OrganizationConformancePackNames: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && isAWSErr(err, configservice.ErrCodeNoSuchOrganizationConformancePackException, "") {
			log.Printf("[WARN] Organization Conformance Pack %q is gone (%s)", d.Id(), awsErr.Code())
			d.SetId("")
			return nil
		}
		return err
	}

	numberOfPacks := len(out.OrganizationConformancePacks)
	if numberOfPacks < 1 {
		log.Printf("[WARN] Organization Conformance Pack %q is gone (no packs found)", d.Id())
		d.SetId("")
		return nil
	}

	if numberOfPacks > 1 {
		return fmt.Errorf("expected exactly 1 organization conformance pack, received %d: %#v",
			numberOfPacks, out.OrganizationConformancePacks)
	}

	log.Printf("[DEBUG] AWS Config organization conformance packs received: %s", out)

	pack := out.OrganizationConformancePacks[0]
	if err = d.Set("arn", pack.OrganizationConformancePackArn); err != nil {
		return err
	}
	if err = d.Set("name", pack.OrganizationConformancePackName); err != nil {
		return err
	}
	if err = d.Set("delivery_s3_bucket", pack.DeliveryS3Bucket); err != nil {
		return err
	}
	if err = d.Set("delivery_s3_key_prefix", pack.DeliveryS3KeyPrefix); err != nil {
		return err
	}
	if err = d.Set("excluded_accounts", pack.ExcludedAccounts); err != nil {
		return err
	}

	if pack.ConformancePackInputParameters != nil {
		if err = d.Set("input_parameters", flattenConformancePackInputParameters(pack.ConformancePackInputParameters)); err != nil {
			return err
		}
	}

	return nil
}

func flattenConformancePackInputParameters(parameters []*configservice.ConformancePackInputParameter) (m map[string]string) {
	m = make(map[string]string)
	for _, p := range parameters {
		m[*p.ParameterName] = *p.ParameterValue
	}
	return
}

func resourceAwsConfigOrganizationConformancePackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting AWS Config organization conformance pack %q", name)
	input := &configservice.DeleteOrganizationConformancePackInput{
		OrganizationConformancePackName: aws.String(name),
	}
	err := resource.Retry(30*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteOrganizationConformancePack(input)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceInUseException" {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteOrganizationConformancePack(input)
	}
	if err != nil {
		return fmt.Errorf("deleting organization conformance pack failed: %s", err)
	}

	conf := resource.StateChangeConf{
		Pending: []string{
			configservice.OrganizationResourceDetailedStatusDeleteInProgress,
		},
		Target:  []string{""},
		Timeout: 30 * time.Minute,
		Refresh: refreshOrganizationConformancePackStatus(d, conn),
	}
	_, err = conf.WaitForState()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] AWS organization conformance pack %q deleted", name)

	return nil
}
