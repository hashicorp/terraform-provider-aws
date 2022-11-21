package logs

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationPut,
		Read:   resourceDestinationRead,
		Update: resourceDestinationPut,
		Delete: resourceDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringLenBetween(1, 512),
					validation.StringMatch(regexp.MustCompile(`[^:*]*`), ""),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"target_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceDestinationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	name := d.Get("name").(string)
	input := &cloudwatchlogs.PutDestinationInput{
		DestinationName: aws.String(name),
		RoleArn:         aws.String(d.Get("role_arn").(string)),
		TargetArn:       aws.String(d.Get("target_arn").(string)),
	}
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(3*time.Minute, func() (interface{}, error) {
		return conn.PutDestination(input)
	}, cloudwatchlogs.ErrCodeInvalidParameterException)

	if err != nil {
		return fmt.Errorf("putting CloudWatch Logs Destination (%s): %w", name, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.StringValue(outputRaw.(*cloudwatchlogs.PutDestinationOutput).Destination.DestinationName))
	}

	return resourceDestinationRead(d, meta)
}

func resourceDestinationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	destination, err := FindDestinationByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading CloudWatch Logs Destination (%s): %w", d.Id(), err)
	}

	d.Set("arn", destination.Arn)
	d.Set("name", destination.DestinationName)
	d.Set("role_arn", destination.RoleArn)
	d.Set("target_arn", destination.TargetArn)

	return nil
}

func resourceDestinationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	log.Printf("[INFO] Deleting CloudWatch Logs Destination: %s", d.Id())
	_, err := conn.DeleteDestination(&cloudwatchlogs.DeleteDestinationInput{
		DestinationName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting CloudWatch Logs Destination (%s): %w", d.Id(), err)
	}

	return nil
}

func FindDestinationByName(conn *cloudwatchlogs.CloudWatchLogs, name string) (*cloudwatchlogs.Destination, error) {
	input := &cloudwatchlogs.DescribeDestinationsInput{
		DestinationNamePrefix: aws.String(name),
	}
	var output *cloudwatchlogs.Destination

	err := conn.DescribeDestinationsPages(input, func(page *cloudwatchlogs.DescribeDestinationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Destinations {
			if aws.StringValue(v.DestinationName) == name {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
