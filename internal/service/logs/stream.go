package logs

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamCreate,
		Read:   resourceStreamRead,
		Delete: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			State: resourceStreamImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ValidStreamName,
			},

			"log_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStreamCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	log.Printf("[DEBUG] Creating CloudWatch Log Stream: %s", d.Get("name").(string))
	_, err := conn.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(d.Get("log_group_name").(string)),
		LogStreamName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return fmt.Errorf("Creating CloudWatch Log Stream failed: %s", err)
	}

	d.SetId(d.Get("name").(string))

	return resourceStreamRead(d, meta)
}

func resourceStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	group := d.Get("log_group_name").(string)

	var ls *cloudwatchlogs.LogStream
	var exists bool

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		ls, exists, err = LookupStream(conn, d.Id(), group, nil)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if d.IsNewResource() && !exists {
			return resource.RetryableError(&resource.NotFoundError{})
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		ls, exists, err = LookupStream(conn, d.Id(), group, nil)
	}

	if err != nil {
		if !tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
			return err
		}

		log.Printf("[DEBUG] container CloudWatch group %q Not Found.", group)
		exists = false
	}

	if !exists {
		log.Printf("[DEBUG] CloudWatch Stream %q Not Found. Removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", ls.Arn)
	d.Set("name", ls.LogStreamName)

	return nil
}

func resourceStreamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	log.Printf("[INFO] Deleting CloudWatch Log Stream: %s", d.Id())
	params := &cloudwatchlogs.DeleteLogStreamInput{
		LogGroupName:  aws.String(d.Get("log_group_name").(string)),
		LogStreamName: aws.String(d.Id()),
	}

	_, err := conn.DeleteLogStream(params)

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudWatch Log Stream (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceStreamImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'log-group-name:log-stream-name'", d.Id())
	}

	logGroupName := parts[0]
	logStreamName := parts[1]

	d.SetId(logStreamName)
	d.Set("log_group_name", logGroupName)

	return []*schema.ResourceData{d}, nil
}

func LookupStream(conn *cloudwatchlogs.CloudWatchLogs,
	name string, logStreamName string, nextToken *string) (*cloudwatchlogs.LogStream, bool, error) {
	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogStreamNamePrefix: aws.String(name),
		LogGroupName:        aws.String(logStreamName),
		NextToken:           nextToken,
	}
	resp, err := conn.DescribeLogStreams(input)
	if err != nil {
		return nil, true, err
	}

	for _, ls := range resp.LogStreams {
		if aws.StringValue(ls.LogStreamName) == name {
			return ls, true, nil
		}
	}

	if resp.NextToken != nil {
		return LookupStream(conn, name, logStreamName, resp.NextToken)
	}

	return nil, false, nil
}

func ValidStreamName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if regexp.MustCompile(`:`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"colons not allowed in %q:", k))
	}
	if len(value) < 1 || len(value) > 512 {
		errors = append(errors, fmt.Errorf(
			"%q must be between 1 and 512 characters: %q", k, value))
	}

	return

}
