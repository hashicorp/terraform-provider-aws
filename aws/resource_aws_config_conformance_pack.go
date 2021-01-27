package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsConfigConformancePack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigConformancePackPut,
		Read:   resourceAwsConfigConformancePackRead,
		Update: resourceAwsConfigConformancePackPut,
		Delete: resourceAwsConfigConformancePackDelete,

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
		},
	}
}

func resourceAwsConfigConformancePackPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("name").(string)
	input := configservice.PutConformancePackInput{
		ConformancePackName: aws.String(name),
	}

	if v, ok := d.GetOk("delivery_s3_bucket"); ok {
		input.DeliveryS3Bucket = aws.String(v.(string))
	}
	if v, ok := d.GetOk("delivery_s3_key_prefix"); ok {
		input.DeliveryS3KeyPrefix = aws.String(v.(string))
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

	_, err := conn.PutConformancePack(&input)
	if err != nil {
		return fmt.Errorf("failed to put AWSConfig conformance pack %q: %s", name, err)
	}

	d.SetId(name)
	conf := resource.StateChangeConf{
		Pending: []string{
			configservice.ConformancePackStateCreateInProgress,
		},
		Target: []string{
			configservice.ConformancePackStateCreateComplete,
		},
		Timeout: 30 * time.Minute,
		Refresh: refreshConformancePackStatus(d, conn),
	}
	if _, err := conf.WaitForState(); err != nil {
		return err
	}
	return resourceAwsConfigConformancePackRead(d, meta)
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

func refreshConformancePackStatus(d *schema.ResourceData, conn *configservice.ConfigService) func() (interface{}, string, error) {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeConformancePackStatus(&configservice.DescribeConformancePackStatusInput{
			ConformancePackNames: []*string{aws.String(d.Id())},
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && isAWSErr(awsErr, configservice.ErrCodeNoSuchConformancePackException, "") {
				return 42, "", nil
			}
			return 42, "", fmt.Errorf("failed to describe conformance pack %q: %s", d.Id(), err)
		}
		if len(out.ConformancePackStatusDetails) < 1 {
			return 42, "", nil
		}
		status := out.ConformancePackStatusDetails[0]
		return out, *status.ConformancePackState, nil
	}
}

func resourceAwsConfigConformancePackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	out, err := conn.DescribeConformancePacks(&configservice.DescribeConformancePacksInput{
		ConformancePackNames: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && isAWSErr(err, configservice.ErrCodeNoSuchConformancePackException, "") {
			log.Printf("[WARN]  Conformance Pack %q is gone (%s)", d.Id(), awsErr.Code())
			d.SetId("")
			return nil
		}
		return err
	}

	numberOfPacks := len(out.ConformancePackDetails)
	if numberOfPacks < 1 {
		log.Printf("[WARN]  Conformance Pack %q is gone (no packs found)", d.Id())
		d.SetId("")
		return nil
	}

	if numberOfPacks > 1 {
		return fmt.Errorf("expected exactly 1 conformance pack, received %d: %#v",
			numberOfPacks, out.ConformancePackDetails)
	}

	log.Printf("[DEBUG] AWS Config conformance packs received: %s", out)

	pack := out.ConformancePackDetails[0]
	if err = d.Set("arn", pack.ConformancePackArn); err != nil {
		return err
	}
	if err = d.Set("name", pack.ConformancePackName); err != nil {
		return err
	}
	if err = d.Set("delivery_s3_bucket", pack.DeliveryS3Bucket); err != nil {
		return err
	}
	if err = d.Set("delivery_s3_key_prefix", pack.DeliveryS3KeyPrefix); err != nil {
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

func resourceAwsConfigConformancePackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting AWS Config conformance pack %q", name)
	input := &configservice.DeleteConformancePackInput{
		ConformancePackName: aws.String(name),
	}
	err := resource.Retry(30*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteConformancePack(input)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceInUseException" {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteConformancePack(input)
	}
	if err != nil {
		return fmt.Errorf("deleting conformance pack failed: %s", err)
	}

	conf := resource.StateChangeConf{
		Pending: []string{
			configservice.ConformancePackStateDeleteInProgress,
		},
		Target:  []string{""},
		Timeout: 30 * time.Minute,
		Refresh: refreshConformancePackStatus(d, conn),
	}
	_, err = conf.WaitForState()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] AWS conformance pack %q deleted", name)

	return nil
}
