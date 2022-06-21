package logs

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSubscriptionFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubscriptionFilterCreate,
		Read:   resourceSubscriptionFilterRead,
		Update: resourceSubscriptionFilterUpdate,
		Delete: resourceSubscriptionFilterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSubscriptionFilterImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"destination_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"filter_pattern": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"log_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"distribution": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cloudwatchlogs.DistributionByLogStream,
				ValidateFunc: validation.StringInSlice(cloudwatchlogs.Distribution_Values(), false),
			},
		},
	}
}

func resourceSubscriptionFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn
	params := getSubscriptionFilterInput(d)
	log.Printf("[DEBUG] Creating SubscriptionFilter %#v", params)

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.PutSubscriptionFilter(&params)

		if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeInvalidParameterException, "Could not deliver test message to specified") {
			return resource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeInvalidParameterException, "Could not execute the lambda function") {
			return resource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeOperationAbortedException, "Please try again") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutSubscriptionFilter(&params)
	}

	if err != nil {
		return fmt.Errorf("Error creating Cloudwatch log subscription filter: %s", err)
	}

	d.SetId(subscriptionFilterID(d.Get("log_group_name").(string)))
	log.Printf("[DEBUG] Cloudwatch logs subscription %q created", d.Id())
	return nil
}

func resourceSubscriptionFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	params := getSubscriptionFilterInput(d)

	log.Printf("[DEBUG] Update SubscriptionFilter %#v", params)

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.PutSubscriptionFilter(&params)

		if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeInvalidParameterException, "Could not deliver test message to specified") {
			return resource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeInvalidParameterException, "Could not execute the lambda function") {
			return resource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeOperationAbortedException, "Please try again") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutSubscriptionFilter(&params)
	}

	if err != nil {
		return fmt.Errorf("error updating CloudWatch Log Subscription Filter (%s): %w", d.Get("log_group_name").(string), err)
	}

	d.SetId(subscriptionFilterID(d.Get("log_group_name").(string)))
	return resourceSubscriptionFilterRead(d, meta)
}

func getSubscriptionFilterInput(d *schema.ResourceData) cloudwatchlogs.PutSubscriptionFilterInput {
	params := cloudwatchlogs.PutSubscriptionFilterInput{
		FilterName:     aws.String(d.Get("name").(string)),
		DestinationArn: aws.String(d.Get("destination_arn").(string)),
		FilterPattern:  aws.String(d.Get("filter_pattern").(string)),
		LogGroupName:   aws.String(d.Get("log_group_name").(string)),
	}

	if _, ok := d.GetOk("role_arn"); ok {
		params.RoleArn = aws.String(d.Get("role_arn").(string))
	}

	if _, ok := d.GetOk("distribution"); ok {
		params.Distribution = aws.String(d.Get("distribution").(string))
	}

	return params
}

func resourceSubscriptionFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	log_group_name := d.Get("log_group_name").(string)
	name := d.Get("name").(string) // "name" is a required field in the schema

	subscriptionFilter, err := FindSubscriptionFilter(conn, log_group_name, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cloudwatch Logs Subscription Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Cloudwatch Logs Subscription Filter (%s): %w", d.Id(), err)
	}

	d.Set("destination_arn", subscriptionFilter.DestinationArn)
	d.Set("distribution", subscriptionFilter.Distribution)
	d.Set("filter_pattern", subscriptionFilter.FilterPattern)
	d.Set("log_group_name", subscriptionFilter.LogGroupName)
	d.Set("role_arn", subscriptionFilter.RoleArn)

	return nil
}

func resourceSubscriptionFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn
	log.Printf("[INFO] Deleting CloudWatch Log Group Subscription: %s", d.Id())
	log_group_name := d.Get("log_group_name").(string)
	name := d.Get("name").(string)

	params := &cloudwatchlogs.DeleteSubscriptionFilterInput{
		FilterName:   aws.String(name),           // Required
		LogGroupName: aws.String(log_group_name), // Required
	}
	_, err := conn.DeleteSubscriptionFilter(params)
	if err != nil {
		if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeResourceNotFoundException, "The specified log group does not exist") {
			return nil
		}
		return fmt.Errorf("error deleting Subscription Filter from log group: %s with name filter name %s: %w", log_group_name, name, err)
	}

	return nil
}

func resourceSubscriptionFilterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "|")
	if len(idParts) < 2 {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <log-group-name>|<filter-name>", d.Id())
	}

	logGroupName := idParts[0]
	filterNamePrefix := idParts[1]

	d.Set("log_group_name", logGroupName)
	d.Set("name", filterNamePrefix)
	d.SetId(subscriptionFilterID(filterNamePrefix))

	return []*schema.ResourceData{d}, nil
}

func subscriptionFilterID(log_group_name string) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s-", log_group_name)) // only one filter allowed per log_group_name at the moment

	return fmt.Sprintf("cwlsf-%d", create.StringHashcode(buf.String()))
}
