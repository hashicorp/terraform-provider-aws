package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHost() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostCreate,
		ReadWithoutTimeout:   resourceHostRead,
		UpdateWithoutTimeout: resourceHostUpdate,
		DeleteWithoutTimeout: resourceHostDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_placement": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.AutoPlacementOn,
				ValidateFunc: validation.StringInSlice(ec2.AutoPlacement_Values(), false),
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host_recovery": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.HostRecoveryOff,
				ValidateFunc: validation.StringInSlice(ec2.HostRecovery_Values(), false),
			},
			"instance_family": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"instance_family", "instance_type"},
			},
			"instance_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"instance_family", "instance_type"},
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceHostCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.AllocateHostsInput{
		AutoPlacement:    aws.String(d.Get("auto_placement").(string)),
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		ClientToken:      aws.String(resource.UniqueId()),
		HostRecovery:     aws.String(d.Get("host_recovery").(string)),
		Quantity:         aws.Int64(1),
	}

	if v, ok := d.GetOk("instance_family"); ok {
		input.InstanceFamily = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_type"); ok {
		input.InstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.TagSpecifications = tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeDedicatedHost)
	}

	output, err := conn.AllocateHostsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "allocating EC2 Host: %s", err)
	}

	d.SetId(aws.StringValue(output.HostIds[0]))

	if _, err := WaitHostCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Host (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceHostRead(ctx, d, meta)...)
}

func resourceHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	host, err := FindHostByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Host %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Host (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(host.OwnerId),
		Resource:  fmt.Sprintf("dedicated-host/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("auto_placement", host.AutoPlacement)
	d.Set("availability_zone", host.AvailabilityZone)
	d.Set("host_recovery", host.HostRecovery)
	d.Set("instance_family", host.HostProperties.InstanceFamily)
	d.Set("instance_type", host.HostProperties.InstanceType)
	d.Set("outpost_arn", host.OutpostArn)
	d.Set("owner_id", host.OwnerId)

	tags := KeyValueTags(host.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceHostUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyHostsInput{
			HostIds: aws.StringSlice([]string{d.Id()}),
		}

		if d.HasChange("auto_placement") {
			input.AutoPlacement = aws.String(d.Get("auto_placement").(string))
		}

		if d.HasChange("host_recovery") {
			input.HostRecovery = aws.String(d.Get("host_recovery").(string))
		}

		if hasChange, v := d.HasChange("instance_family"), d.Get("instance_family").(string); hasChange && v != "" {
			input.InstanceFamily = aws.String(v)
		}

		if hasChange, v := d.HasChange("instance_type"), d.Get("instance_type").(string); hasChange && v != "" {
			input.InstanceType = aws.String(v)
		}

		output, err := conn.ModifyHostsWithContext(ctx, input)

		if err == nil && output != nil {
			err = UnsuccessfulItemsError(output.Unsuccessful)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Host (%s): %s", d.Id(), err)
		}

		if _, err := WaitHostUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Host (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Host (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceHostRead(ctx, d, meta)...)
}

func resourceHostDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 Host: %s", d.Id())
	output, err := conn.ReleaseHostsWithContext(ctx, &ec2.ReleaseHostsInput{
		HostIds: aws.StringSlice([]string{d.Id()}),
	})

	if err == nil && output != nil {
		err = UnsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeClientInvalidHostIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "releasing EC2 Host (%s): %s", d.Id(), err)
	}

	if _, err := WaitHostDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Host (%s) delete: %s", d.Id(), err)
	}

	return diags
}
