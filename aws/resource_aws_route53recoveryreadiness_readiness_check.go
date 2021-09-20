package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsRoute53RecoveryReadinessReadinessCheck() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53RecoveryReadinessReadinessCheckCreate,
		Read:   resourceAwsRoute53RecoveryReadinessReadinessCheckRead,
		Update: resourceAwsRoute53RecoveryReadinessReadinessCheckUpdate,
		Delete: resourceAwsRoute53RecoveryReadinessReadinessCheckDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"readiness_check_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_set_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsRoute53RecoveryReadinessReadinessCheckCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoveryreadinessconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &route53recoveryreadiness.CreateReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Get("readiness_check_name").(string)),
		ResourceSetName:    aws.String(d.Get("resource_set_name").(string)),
	}

	resp, err := conn.CreateReadinessCheck(input)
	if err != nil {
		return fmt.Errorf("error creating Route53 Recovery Readiness ReadinessCheck: %w", err)
	}

	d.SetId(aws.StringValue(resp.ReadinessCheckName))

	if len(tags) > 0 {
		arn := aws.StringValue(resp.ReadinessCheckArn)
		if err := keyvaluetags.Route53recoveryreadinessUpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error adding Route53 Recovery Readiness ReadinessCheck (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsRoute53RecoveryReadinessReadinessCheckRead(d, meta)
}

func resourceAwsRoute53RecoveryReadinessReadinessCheckRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoveryreadinessconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &route53recoveryreadiness.GetReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	}

	resp, err := conn.GetReadinessCheck(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53RecoveryReadiness Readiness Check (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Route53 Recovery Readiness ReadinessCheck: %s", err)
	}

	d.Set("arn", resp.ReadinessCheckArn)
	d.Set("readiness_check_name", resp.ReadinessCheckName)
	d.Set("resource_set_name", resp.ResourceSet)

	tags, err := keyvaluetags.Route53recoveryreadinessListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Recovery Readiness ReadinessCheck (%s): %w", d.Id(), err)
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

func resourceAwsRoute53RecoveryReadinessReadinessCheckUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoveryreadinessconn

	input := &route53recoveryreadiness.UpdateReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Get("readiness_check_name").(string)),
		ResourceSetName:    aws.String(d.Get("resource_set_name").(string)),
	}

	_, err := conn.UpdateReadinessCheck(input)
	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Readiness ReadinessCheck: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)
		if err := keyvaluetags.Route53recoveryreadinessUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Route53 Recovery Readiness ReadinessCheck (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsRoute53RecoveryReadinessReadinessCheckRead(d, meta)
}

func resourceAwsRoute53RecoveryReadinessReadinessCheckDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoveryreadinessconn

	input := &route53recoveryreadiness.DeleteReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	}
	_, err := conn.DeleteReadinessCheck(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, route53recoveryreadiness.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Route53 Recovery Readiness ReadinessCheck: %s", err)
	}

	gcinput := &route53recoveryreadiness.GetReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	}
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.GetReadinessCheck(gcinput)
		if err != nil {
			if tfawserr.ErrMessageContains(err, route53recoveryreadiness.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Route 53 Recovery Readiness ReadinessCheck (%s) still exists", d.Id()))
	})

	if tfresource.TimedOut(err) {
		_, err = conn.GetReadinessCheck(gcinput)
	}

	if err != nil {
		return fmt.Errorf("error waiting for Route 53 Recovery Readiness ReadinessCheck (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
