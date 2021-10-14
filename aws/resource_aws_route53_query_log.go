package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceQueryLog() *schema.Resource {
	return &schema.Resource{
		Create: resourceQueryLogCreate,
		Read:   resourceQueryLogRead,
		Delete: resourceQueryLogDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudwatch_log_group_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceQueryLogCreate(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*conns.AWSClient).Route53Conn

	input := &route53.CreateQueryLoggingConfigInput{
		CloudWatchLogsLogGroupArn: aws.String(d.Get("cloudwatch_log_group_arn").(string)),
		HostedZoneId:              aws.String(d.Get("zone_id").(string)),
	}

	log.Printf("[DEBUG] Creating Route53 query logging configuration: %#v", input)
	out, err := r53.CreateQueryLoggingConfig(input)
	if err != nil {
		return fmt.Errorf("Error creating Route53 query logging configuration: %s", err)
	}
	log.Printf("[DEBUG] Route53 query logging configuration created: %#v", out)

	d.SetId(aws.StringValue(out.QueryLoggingConfig.Id))

	return resourceQueryLogRead(d, meta)
}

func resourceQueryLogRead(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*conns.AWSClient).Route53Conn

	input := &route53.GetQueryLoggingConfigInput{
		Id: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Route53 query logging configuration: %#v", input)
	out, err := r53.GetQueryLoggingConfig(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, route53.ErrCodeNoSuchQueryLoggingConfig, "") || tfawserr.ErrMessageContains(err, route53.ErrCodeNoSuchHostedZone, "") {
			log.Printf("[WARN] Route53 Query Logging Config (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Route53 query logging configuration: %s", err)
	}
	log.Printf("[DEBUG] Route53 query logging configuration received: %#v", out)

	d.Set("cloudwatch_log_group_arn", out.QueryLoggingConfig.CloudWatchLogsLogGroupArn)
	d.Set("zone_id", out.QueryLoggingConfig.HostedZoneId)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("queryloggingconfig/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceQueryLogDelete(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*conns.AWSClient).Route53Conn

	input := &route53.DeleteQueryLoggingConfigInput{
		Id: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Route53 query logging configuration: %#v", input)
	_, err := r53.DeleteQueryLoggingConfig(input)
	if tfawserr.ErrMessageContains(err, route53.ErrCodeNoSuchQueryLoggingConfig, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route53 query logging configuration (%s): %w", d.Id(), err)
	}

	return nil
}
