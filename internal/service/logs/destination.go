package logs

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	_sp.registerSDKResourceFactory("aws_cloudwatch_log_destination", resourceDestination)
}

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDestinationCreate,
		ReadWithoutTimeout:   resourceDestinationRead,
		UpdateWithoutTimeout: resourceDestinationUpdate,
		DeleteWithoutTimeout: resourceDestinationDelete,

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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"target_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	propagationTimeout = 2 * time.Minute
)

func resourceDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &cloudwatchlogs.PutDestinationInput{
		DestinationName: aws.String(name),
		RoleArn:         aws.String(d.Get("role_arn").(string)),
		TargetArn:       aws.String(d.Get("target_arn").(string)),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.PutDestinationWithContext(ctx, input)
	}, cloudwatchlogs.ErrCodeInvalidParameterException)

	if err != nil {
		return diag.Errorf("creating CloudWatch Logs Destination (%s): %s", name, err)
	}

	destination := outputRaw.(*cloudwatchlogs.PutDestinationOutput).Destination
	d.SetId(aws.StringValue(destination.DestinationName))

	// Although PutDestinationInput has a Tags field, specifying tags there results in
	// "InvalidParameterException: Could not deliver test message to specified destination. Check if the destination is valid."
	if len(tags) > 0 {
		if err := UpdateTags(ctx, conn, aws.StringValue(destination.Arn), nil, tags); err != nil {
			return diag.Errorf("adding CloudWatch Logs Destination (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDestinationRead(ctx, d, meta)
}

func resourceDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	destination, err := FindDestinationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Logs Destination (%s): %s", d.Id(), err)
	}

	d.Set("arn", destination.Arn)
	d.Set("name", destination.DestinationName)
	d.Set("role_arn", destination.RoleArn)
	d.Set("target_arn", destination.TargetArn)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("listing tags for CloudWatch Logs Destination (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &cloudwatchlogs.PutDestinationInput{
			DestinationName: aws.String(d.Id()),
			RoleArn:         aws.String(d.Get("role_arn").(string)),
			TargetArn:       aws.String(d.Get("target_arn").(string)),
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
			return conn.PutDestinationWithContext(ctx, input)
		}, cloudwatchlogs.ErrCodeInvalidParameterException)

		if err != nil {
			return diag.Errorf("updating CloudWatch Logs Destination (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			log.Printf("[WARN] failed updating tags for CloudWatch Logs Destination (%s): %s", d.Id(), err)
		}
	}

	return resourceDestinationRead(ctx, d, meta)
}

func resourceDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	log.Printf("[INFO] Deleting CloudWatch Logs Destination: %s", d.Id())
	_, err := conn.DeleteDestinationWithContext(ctx, &cloudwatchlogs.DeleteDestinationInput{
		DestinationName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Logs Destination (%s): %s", d.Id(), err)
	}

	return nil
}

func FindDestinationByName(ctx context.Context, conn *cloudwatchlogs.CloudWatchLogs, name string) (*cloudwatchlogs.Destination, error) {
	input := &cloudwatchlogs.DescribeDestinationsInput{
		DestinationNamePrefix: aws.String(name),
	}
	var output *cloudwatchlogs.Destination

	err := conn.DescribeDestinationsPagesWithContext(ctx, input, func(page *cloudwatchlogs.DescribeDestinationsOutput, lastPage bool) bool {
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
