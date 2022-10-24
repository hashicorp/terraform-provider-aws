package route53

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.CreateQueryLoggingConfigInput{
		CloudWatchLogsLogGroupArn: aws.String(d.Get("cloudwatch_log_group_arn").(string)),
		HostedZoneId:              aws.String(d.Get("zone_id").(string)),
	}

	output, err := conn.CreateQueryLoggingConfig(input)

	if err != nil {
		return fmt.Errorf("creating Route53 Query Logging Config: %w", err)
	}

	d.SetId(aws.StringValue(output.QueryLoggingConfig.Id))

	return resourceQueryLogRead(d, meta)
}

func resourceQueryLogRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn

	output, err := FindQueryLoggingConfigByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Query Logging Config %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Route53 Query Logging Config (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("queryloggingconfig/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("cloudwatch_log_group_arn", output.CloudWatchLogsLogGroupArn)
	d.Set("zone_id", output.HostedZoneId)

	return nil
}

func resourceQueryLogDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn

	log.Printf("[DEBUG] Deleting Route53 Query Logging Config: %s", d.Id())
	_, err := conn.DeleteQueryLoggingConfig(&route53.DeleteQueryLoggingConfigInput{
		Id: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchQueryLoggingConfig) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Route53 Query Logging Config (%s): %w", d.Id(), err)
	}

	return nil
}
