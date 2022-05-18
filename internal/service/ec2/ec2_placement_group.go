package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePlacementGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourcePlacementGroupCreate,
		Read:   resourcePlacementGroupRead,
		Update: resourcePlacementGroupUpdate,
		Delete: resourcePlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourcePlacementGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &ec2.CreatePlacementGroupInput{
		GroupName:         aws.String(name),
		Strategy:          aws.String(d.Get("strategy").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypePlacementGroup),
	}

	if v, ok := d.GetOk("partition_count"); ok {
		input.PartitionCount = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating EC2 Placement Group: %s", input)
	_, err := conn.CreatePlacementGroup(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Placement Group (%s): %w", name, err)
	}

	d.SetId(name)

	_, err = WaitPlacementGroupCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Placement Group (%s) create: %w", d.Id(), err)
	}

	return resourcePlacementGroupRead(d, meta)
}

func resourcePlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	pg, err := FindPlacementGroupByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Placement Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Placement Group (%s): %w", d.Id(), err)
	}

	d.Set("name", pg.GroupName)
	d.Set("partition_count", pg.PartitionCount)
	d.Set("placement_group_id", pg.GroupId)
	d.Set("strategy", pg.Strategy)

	tags := KeyValueTags(pg.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("placement-group/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourcePlacementGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("placement_group_id").(string), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Placement Group (%s) tags: %w", d.Id(), err)
		}
	}

	return resourcePlacementGroupRead(d, meta)
}

func resourcePlacementGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Placement Group: %s", d.Id())
	_, err := conn.DeletePlacementGroup(&ec2.DeletePlacementGroupInput{
		GroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPlacementGroupUnknown) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Placement Group (%s): %w", d.Id(), err)
	}

	_, err = WaitPlacementGroupDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Placement Group (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func resourcePlacementGroupCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" {
		if partitionCount, strategy := diff.Get("partition_count").(int), diff.Get("strategy").(string); partitionCount > 0 && strategy != ec2.PlacementGroupStrategyPartition {
			return fmt.Errorf("partition_count must not be set when strategy = %q", strategy)
		}
	}

	return nil
}
