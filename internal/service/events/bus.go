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
)

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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBusCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	eventBusName := d.Get("name").(string)
	input := &eventbridge.CreateEventBusInput{
		Name: aws.String(eventBusName),
	}

	if v, ok := d.GetOk("event_source_name"); ok {
		input.EventSourceName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating EventBridge event bus: %v", input)

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
	if input.Tags == nil && len(tags) > 0 {
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
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	tags, err := ListTags(ctx, conn, aws.StringValue(output.Arn))

	// ISO partitions may not support tagging, giving error
	if verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] Unable to list tags for EventBridge Bus %s: %s", d.Id(), err)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for EventBridge event bus (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceBusUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(ctx, conn, arn, o, n)

		if verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] Unable to update tags for EventBridge Bus %s: %s", d.Id(), err)
			return append(diags, resourceBusRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Bus tags: %s", err)
		}
	}

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
