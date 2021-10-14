package storagegateway

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTapePool() *schema.Resource {
	return &schema.Resource{
		Create: resourceTapePoolCreate,
		Read:   resourceTapePoolRead,
		Update: resourceTapePoolUpdate,
		Delete: resourceTapePoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"storage_class": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(storagegateway.TapeStorageClass_Values(), false),
			},
			"retention_lock_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      storagegateway.RetentionLockTypeNone,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(storagegateway.RetentionLockType_Values(), false),
			},
			"retention_lock_time_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 36500),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTapePoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &storagegateway.CreateTapePoolInput{
		PoolName:                aws.String(d.Get("pool_name").(string)),
		StorageClass:            aws.String(d.Get("storage_class").(string)),
		RetentionLockType:       aws.String(d.Get("retention_lock_type").(string)),
		RetentionLockTimeInDays: aws.Int64(int64(d.Get("retention_lock_time_in_days").(int))),
		Tags:                    Tags(tags.IgnoreAws()),
	}

	log.Printf("[DEBUG] Creating Storage Gateway Tape Pool: %s", input)
	output, err := conn.CreateTapePool(input)
	if err != nil {
		return fmt.Errorf("error creating Storage Gateway Tape Pool: %w", err)
	}

	d.SetId(aws.StringValue(output.PoolARN))

	return resourceTapePoolRead(d, meta)
}

func resourceTapePoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceTapePoolRead(d, meta)
}

func resourceTapePoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &storagegateway.ListTapePoolsInput{
		PoolARNs: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading Storage Gateway Tape Pool: %s", input)
	output, err := conn.ListTapePools(input)

	if err != nil {
		return fmt.Errorf("[ERROR] Listing Storage Gateway Tape Pools %w", err)
	}

	if output == nil || len(output.PoolInfos) == 0 || output.PoolInfos[0] == nil || aws.StringValue(output.PoolInfos[0].PoolARN) != d.Id() {
		log.Printf("[WARN] Storage Gateway Tape Pool %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	pool := output.PoolInfos[0]

	poolArn := aws.StringValue(pool.PoolARN)
	d.Set("arn", poolArn)
	d.Set("pool_name", pool.PoolName)
	d.Set("retention_lock_time_in_days", pool.RetentionLockTimeInDays)
	d.Set("retention_lock_type", pool.RetentionLockType)
	d.Set("storage_class", pool.StorageClass)

	tags, err := ListTags(conn, poolArn)
	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %w", poolArn, err)
	}
	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceTapePoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn

	input := &storagegateway.DeleteTapePoolInput{
		PoolARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway Tape Pool: %s", input)
	_, err := conn.DeleteTapePool(input)
	if err != nil {
		return fmt.Errorf("error deleting Storage Gateway Tape Pool %q: %w", d.Id(), err)
	}

	return nil
}
