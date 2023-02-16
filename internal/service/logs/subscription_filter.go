package logs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	_sp.registerSDKResourceFactory("aws_cloudwatch_log_subscription_filter", resourceSubscriptionFilter)
}

func resourceSubscriptionFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubscriptionFilterPut,
		ReadWithoutTimeout:   resourceSubscriptionFilterRead,
		UpdateWithoutTimeout: resourceSubscriptionFilterPut,
		DeleteWithoutTimeout: resourceSubscriptionFilterDelete,

		Importer: &schema.ResourceImporter{
			State: resourceSubscriptionFilterImport,
		},

		Schema: map[string]*schema.Schema{
			"destination_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"distribution": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cloudwatchlogs.DistributionByLogStream,
				ValidateFunc: validation.StringInSlice(cloudwatchlogs.Distribution_Values(), false),
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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceSubscriptionFilterPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	logGroupName := d.Get("log_group_name").(string)
	name := d.Get("name").(string)
	input := &cloudwatchlogs.PutSubscriptionFilterInput{
		DestinationArn: aws.String(d.Get("destination_arn").(string)),
		FilterName:     aws.String(name),
		FilterPattern:  aws.String(d.Get("filter_pattern").(string)),
		LogGroupName:   aws.String(logGroupName),
	}

	if v, ok := d.GetOk("distribution"); ok {
		input.Distribution = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhen(ctx, 5*time.Minute,
		func() (interface{}, error) {
			return conn.PutSubscriptionFilterWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeInvalidParameterException, "Could not deliver test message to specified") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeInvalidParameterException, "Could not execute the lambda function") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, cloudwatchlogs.ErrCodeOperationAbortedException, "Please try again") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return diag.Errorf("putting CloudWatch Logs Subscription Filter (%s): %s", name, err)
	}

	d.SetId(subscriptionFilterID(logGroupName))

	return nil
}

func resourceSubscriptionFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	subscriptionFilter, err := FindSubscriptionFilterByTwoPartKey(ctx, conn, d.Get("log_group_name").(string), d.Get("name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Subscription Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Logs Subscription Filter (%s): %s", d.Id(), err)
	}

	d.Set("destination_arn", subscriptionFilter.DestinationArn)
	d.Set("distribution", subscriptionFilter.Distribution)
	d.Set("filter_pattern", subscriptionFilter.FilterPattern)
	d.Set("log_group_name", subscriptionFilter.LogGroupName)
	d.Set("name", subscriptionFilter.FilterName)
	d.Set("role_arn", subscriptionFilter.RoleArn)

	return nil
}

func resourceSubscriptionFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	log.Printf("[INFO] Deleting CloudWatch Logs Subscription Filter: %s", d.Id())
	_, err := conn.DeleteSubscriptionFilterWithContext(ctx, &cloudwatchlogs.DeleteSubscriptionFilterInput{
		FilterName:   aws.String(d.Get("name").(string)),
		LogGroupName: aws.String(d.Get("log_group_name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Logs Subscription Filter (%s): %s", d.Id(), err)
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

func FindSubscriptionFilterByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.CloudWatchLogs, logGroupName, name string) (*cloudwatchlogs.SubscriptionFilter, error) {
	input := &cloudwatchlogs.DescribeSubscriptionFiltersInput{
		FilterNamePrefix: aws.String(name),
		LogGroupName:     aws.String(logGroupName),
	}
	var output *cloudwatchlogs.SubscriptionFilter

	err := conn.DescribeSubscriptionFiltersPagesWithContext(ctx, input, func(page *cloudwatchlogs.DescribeSubscriptionFiltersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SubscriptionFilters {
			if aws.StringValue(v.FilterName) == name {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
