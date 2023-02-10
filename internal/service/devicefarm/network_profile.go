package devicefarm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNetworkProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkProfileCreate,
		ReadWithoutTimeout:   resourceNetworkProfileRead,
		UpdateWithoutTimeout: resourceNetworkProfileUpdate,
		DeleteWithoutTimeout: resourceNetworkProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"downlink_bandwidth_bits": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      104857600,
				ValidateFunc: validation.IntBetween(0, 104857600),
			},
			"downlink_delay_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"downlink_jitter_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"downlink_loss_percent": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"project_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"uplink_bandwidth_bits": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      104857600,
				ValidateFunc: validation.IntBetween(0, 104857600),
			},
			"uplink_delay_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"uplink_jitter_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"uplink_loss_percent": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      devicefarm.NetworkProfileTypePrivate,
				ValidateFunc: validation.StringInSlice(devicefarm.NetworkProfileType_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNetworkProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &devicefarm.CreateNetworkProfileInput{
		Name:       aws.String(d.Get("name").(string)),
		ProjectArn: aws.String(d.Get("project_arn").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	if v, ok := d.GetOk("downlink_bandwidth_bits"); ok {
		input.DownlinkBandwidthBits = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("downlink_delay_ms"); ok {
		input.DownlinkDelayMs = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("downlink_jitter_ms"); ok {
		input.DownlinkJitterMs = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("downlink_loss_percent"); ok {
		input.DownlinkLossPercent = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("uplink_bandwidth_bits"); ok {
		input.UplinkBandwidthBits = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("uplink_delay_ms"); ok {
		input.UplinkDelayMs = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("uplink_jitter_ms"); ok {
		input.UplinkJitterMs = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("uplink_loss_percent"); ok {
		input.UplinkLossPercent = aws.Int64(int64(v.(int)))
	}

	out, err := conn.CreateNetworkProfileWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating DeviceFarm Network Profile: %s", err)
	}

	arn := aws.StringValue(out.NetworkProfile.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm Network Profile: %s", arn)
	d.SetId(arn)

	if len(tags) > 0 {
		if err := UpdateTags(ctx, conn, arn, nil, tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Network Profile (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceNetworkProfileRead(ctx, d, meta)...)
}

func resourceNetworkProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	project, err := FindNetworkProfileByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Network Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Network Profile (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(project.Arn)
	d.Set("arn", arn)
	d.Set("name", project.Name)
	d.Set("description", project.Description)
	d.Set("downlink_bandwidth_bits", project.DownlinkBandwidthBits)
	d.Set("downlink_delay_ms", project.DownlinkDelayMs)
	d.Set("downlink_jitter_ms", project.DownlinkJitterMs)
	d.Set("downlink_loss_percent", project.DownlinkLossPercent)
	d.Set("uplink_bandwidth_bits", project.UplinkBandwidthBits)
	d.Set("uplink_delay_ms", project.UplinkDelayMs)
	d.Set("uplink_jitter_ms", project.UplinkJitterMs)
	d.Set("uplink_loss_percent", project.UplinkLossPercent)
	d.Set("type", project.Type)

	projectArn, err := decodeProjectARN(arn, "networkprofile", meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding project_arn (%s): %s", arn, err)
	}

	d.Set("project_arn", projectArn)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DeviceFarm Network Profile (%s): %s", arn, err)
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

func resourceNetworkProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateNetworkProfileInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("type") {
			input.Type = aws.String(d.Get("type").(string))
		}

		if d.HasChange("downlink_bandwidth_bits") {
			input.DownlinkBandwidthBits = aws.Int64(int64(d.Get("downlink_bandwidth_bits").(int)))
		}

		if d.HasChange("downlink_delay_ms") {
			input.DownlinkDelayMs = aws.Int64(int64(d.Get("downlink_delay_ms").(int)))
		}

		if d.HasChange("downlink_jitter_ms") {
			input.DownlinkJitterMs = aws.Int64(int64(d.Get("downlink_jitter_ms").(int)))
		}

		if d.HasChange("downlink_loss_percent") {
			input.DownlinkLossPercent = aws.Int64(int64(d.Get("downlink_loss_percent").(int)))
		}

		if d.HasChange("uplink_bandwidth_bits") {
			input.UplinkBandwidthBits = aws.Int64(int64(d.Get("uplink_bandwidth_bits").(int)))
		}

		if d.HasChange("uplink_delay_ms") {
			input.UplinkDelayMs = aws.Int64(int64(d.Get("uplink_delay_ms").(int)))
		}

		if d.HasChange("uplink_jitter_ms") {
			input.UplinkJitterMs = aws.Int64(int64(d.Get("uplink_jitter_ms").(int)))
		}

		if d.HasChange("uplink_loss_percent") {
			input.UplinkLossPercent = aws.Int64(int64(d.Get("uplink_loss_percent").(int)))
		}

		log.Printf("[DEBUG] Updating DeviceFarm Network Profile: %s", d.Id())
		_, err := conn.UpdateNetworkProfileWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Error Updating DeviceFarm Network Profile: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Network Profile (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceNetworkProfileRead(ctx, d, meta)...)
}

func resourceNetworkProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()

	input := &devicefarm.DeleteNetworkProfileInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm Network Profile: %s", d.Id())
	_, err := conn.DeleteNetworkProfileWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "Error deleting DeviceFarm Network Profile: %s", err)
	}

	return diags
}
