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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePlacementGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlacementGroupCreate,
		ReadWithoutTimeout:   resourcePlacementGroupRead,
		UpdateWithoutTimeout: resourcePlacementGroupUpdate,
		DeleteWithoutTimeout: resourcePlacementGroupDelete,
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
			},
			"partition_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
				// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#placement-groups-limitations-partition.
				ValidateFunc: validation.IntBetween(0, 7),
			},
			"placement_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"spread_level": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.SpreadLevel_Values(), false),
			},
			"strategy": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.PlacementStrategy_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.All(
			resourcePlacementGroupCustomizeDiff,
			verify.SetTagsDiff,
		),
	}
}

func resourcePlacementGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &ec2.CreatePlacementGroupInput{
		GroupName:         aws.String(name),
		Strategy:          aws.String(d.Get("strategy").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypePlacementGroup),
	}

	if v, ok := d.GetOk("partition_count"); ok {
		input.PartitionCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("spread_level"); ok {
		input.SpreadLevel = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Placement Group: %s", input)
	_, err := conn.CreatePlacementGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Placement Group (%s): %s", name, err)
	}

	d.SetId(name)

	_, err = WaitPlacementGroupCreated(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Placement Group (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePlacementGroupRead(ctx, d, meta)...)
}

func resourcePlacementGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	pg, err := FindPlacementGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Placement Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Placement Group (%s): %s", d.Id(), err)
	}

	d.Set("name", pg.GroupName)
	d.Set("partition_count", pg.PartitionCount)
	d.Set("placement_group_id", pg.GroupId)
	d.Set("spread_level", pg.SpreadLevel)
	d.Set("strategy", pg.Strategy)

	tags := KeyValueTags(pg.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("placement-group/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return diags
}

func resourcePlacementGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("placement_group_id").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Placement Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourcePlacementGroupRead(ctx, d, meta)...)
}

func resourcePlacementGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[DEBUG] Deleting EC2 Placement Group: %s", d.Id())
	_, err := conn.DeletePlacementGroupWithContext(ctx, &ec2.DeletePlacementGroupInput{
		GroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPlacementGroupUnknown) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Placement Group (%s): %s", d.Id(), err)
	}

	_, err = WaitPlacementGroupDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Placement Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourcePlacementGroupCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" {
		if partitionCount, strategy := diff.Get("partition_count").(int), diff.Get("strategy").(string); partitionCount > 0 && strategy != ec2.PlacementGroupStrategyPartition {
			return fmt.Errorf("partition_count must not be set when strategy = %q", strategy)
		}
	}

	if diff.Id() == "" {
		if spreadLevel, strategy := diff.Get("spread_level").(string), diff.Get("strategy").(string); spreadLevel != "" && strategy != ec2.PlacementGroupStrategySpread {
			return fmt.Errorf("spread_level must not be set when strategy = %q", strategy)
		}
	}

	return nil
}
