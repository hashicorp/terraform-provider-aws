package events

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_event_bus", name="Event Bus")
// @Tags(identifierAttribute="arn")
func ResourceBus() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBusCreate,
		ReadWithoutTimeout:   resourceBusRead,
		UpdateWithoutTimeout: resourceBusUpdate,
		DeleteWithoutTimeout: resourceBusDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validCustomEventBusName,
			},
			"event_source_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validSourceName,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBusCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

	eventBusName := d.Get("name").(string)
	input := &eventbridge.CreateEventBusInput{
		Name: aws.String(eventBusName),
		Tags: GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("event_source_name"); ok {
		input.EventSourceName = aws.String(v.(string))
	}

	output, err := conn.CreateEventBusWithContext(ctx, input)

	// Some partitions may not support tag-on-create
	if input.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] EventBridge Bus (%s) create failed (%s) with tags. Trying create without tags.", eventBusName, err)
		input.Tags = nil
		output, err = conn.CreateEventBusWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating EventBridge event bus (%s) failed: %s", eventBusName, err)
	}

	d.SetId(eventBusName)

	log.Printf("[INFO] EventBridge event bus (%s) created", d.Id())

	// Post-create tagging supported in some partitions
	if tags := KeyValueTags(ctx, GetTagsIn(ctx)); input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(ctx, conn, aws.StringValue(output.EventBusArn), nil, tags)

		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] error adding tags after create for EventBridge Bus (%s): %s", d.Id(), err)
			return append(diags, resourceBusRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EventBridge Bus (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBusRead(ctx, d, meta)...)
}

func resourceBusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

	input := &eventbridge.DescribeEventBusInput{
		Name: aws.String(d.Id()),
	}

	output, err := conn.DescribeEventBusWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge event bus (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge event bus: %s", err)
	}

	d.Set("arn", output.Arn)
	d.Set("name", output.Name)

	return diags
}

func resourceBusUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceBusRead(ctx, d, meta)...)
}

func resourceBusDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()
	log.Printf("[INFO] Deleting EventBridge event bus (%s)", d.Id())
	_, err := conn.DeleteEventBusWithContext(ctx, &eventbridge.DeleteEventBusInput{
		Name: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge event bus (%s) not found", d.Id())
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge event bus (%s): %s", d.Id(), err)
	}
	log.Printf("[INFO] EventBridge event bus (%s) deleted", d.Id())

	return diags
}
