package rum

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMetricsDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceMetricsDestinationPut,
		Read:   resourceMetricsDestinationRead,
		Update: resourceMetricsDestinationPut,
		Delete: resourceMetricsDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"app_monitor_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"destination": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudwatchrum.MetricDestination_Values(), false),
			},
			"destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceMetricsDestinationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RUMConn

	name := d.Get("app_monitor_name").(string)
	input := &cloudwatchrum.PutRumMetricsDestinationInput{
		AppMonitorName: aws.String(name),
		Destination:    aws.String(d.Get("destination").(string)),
	}

	if v, ok := d.GetOk("destination_arn"); ok {
		input.DestinationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	_, err := conn.PutRumMetricsDestination(input)

	if err != nil {
		return fmt.Errorf("error creating CloudWatch RUM Metric Destination %s: %w", name, err)
	}

	d.SetId(name)

	return resourceMetricsDestinationRead(d, meta)
}

func resourceMetricsDestinationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RUMConn

	dest, err := FindMetricsDestinationsByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find CloudWatch RUM Metric Destination (%s); removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudWatch RUM Metric Destination (%s): %w", d.Id(), err)
	}

	d.Set("destination", dest.Destination)
	d.Set("app_monitor_name", d.Id())
	d.Set("destination_arn", dest.DestinationArn)
	d.Set("iam_role_arn", dest.IamRoleArn)

	return nil
}

func resourceMetricsDestinationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RUMConn

	input := &cloudwatchrum.DeleteRumMetricsDestinationInput{
		AppMonitorName: aws.String(d.Id()),
		Destination:    aws.String(d.Get("destination").(string)),
	}

	if v, ok := d.GetOk("destination_arn"); ok {
		input.DestinationArn = aws.String(v.(string))
	}

	if _, err := conn.DeleteRumMetricsDestination(input); err != nil {
		if tfawserr.ErrCodeEquals(err, cloudwatchrum.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting CloudWatch RUM Metric Destination (%s): %w", d.Id(), err)
	}

	return nil
}
