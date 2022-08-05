package logs

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationPut,
		Update: resourceDestinationPut,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
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

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDestinationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	name := d.Get("name").(string)
	roleArn := d.Get("role_arn").(string)
	targetArn := d.Get("target_arn").(string)

	params := &cloudwatchlogs.PutDestinationInput{
		DestinationName: aws.String(name),
		RoleArn:         aws.String(roleArn),
		TargetArn:       aws.String(targetArn),
	}

	var err error
	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
		_, err = conn.PutDestination(params)

		if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeInvalidParameterException) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.PutDestination(params)
	}
	if err != nil {
		return fmt.Errorf("Error putting cloudwatch log destination: %s", err)
	}
	d.SetId(name)

	return resourceDestinationRead(d, meta)
}

func resourceDestinationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	destination, exists, err := LookupDestination(conn, d.Id(), nil)
	if err != nil {
		return err
	}

	if !exists {
		d.SetId("")
		return nil
	}

	d.Set("arn", destination.Arn)
	d.Set("role_arn", destination.RoleArn)
	d.Set("target_arn", destination.TargetArn)

	return nil
}

func resourceDestinationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	params := &cloudwatchlogs.DeleteDestinationInput{
		DestinationName: aws.String(d.Id()),
	}
	_, err := conn.DeleteDestination(params)
	if err != nil {
		return fmt.Errorf("Error deleting Destination with name %s", d.Id())
	}

	return nil
}

func LookupDestination(conn *cloudwatchlogs.CloudWatchLogs,
	name string, nextToken *string) (*cloudwatchlogs.Destination, bool, error) {
	input := &cloudwatchlogs.DescribeDestinationsInput{
		DestinationNamePrefix: aws.String(name),
		NextToken:             nextToken,
	}
	resp, err := conn.DescribeDestinations(input)
	if err != nil {
		return nil, true, err
	}

	for _, destination := range resp.Destinations {
		if aws.StringValue(destination.DestinationName) == name {
			return destination, true, nil
		}
	}

	if resp.NextToken != nil {
		return LookupDestination(conn, name, resp.NextToken)
	}

	return nil, false, nil
}
