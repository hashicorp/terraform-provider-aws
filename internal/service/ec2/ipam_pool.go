package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMPool() *schema.Resource {
	return &schema.Resource{
		Create: ResourceIPAMPoolCreate,
		Read:   ResourceIPAMPoolRead,
		Update: ResourceIPAMPoolUpdate,
		Delete: ResourceIPAMPoolDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.AddressFamily_Values(), false),
			},
			"allocation_default_netmask_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 128),
			},
			"allocation_max_netmask_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 128),
			},
			"allocation_min_netmask_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 128),
			},
			"allocation_resource_tags": tftags.TagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_import": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"aws_service": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.IpamPoolAwsService_Values(), false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_scope_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ipam_scope_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringInSlice([]string{"None"}, false),
					verify.ValidRegionName,
				),
				Default: "None",
			},
			"pool_depth": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"publicly_advertisable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"source_ipam_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func ResourceIPAMPoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	addressFamily := d.Get("address_family").(string)
	input := &ec2.CreateIpamPoolInput{
		AddressFamily:     aws.String(addressFamily),
		ClientToken:       aws.String(resource.UniqueId()),
		IpamScopeId:       aws.String(d.Get("ipam_scope_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeIpamPool),
	}

	if v, ok := d.GetOk("allocation_default_netmask_length"); ok {
		input.AllocationDefaultNetmaskLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("allocation_max_netmask_length"); ok {
		input.AllocationMaxNetmaskLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("allocation_min_netmask_length"); ok {
		input.AllocationMinNetmaskLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("allocation_resource_tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.AllocationResourceTags = ipamResourceTags(tftags.New(v.(map[string]interface{})))
	}

	if v, ok := d.GetOk("auto_import"); ok {
		input.AutoImport = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("locale"); ok && v != "None" {
		input.Locale = aws.String(v.(string))
	}

	if v, ok := d.GetOk("aws_service"); ok {
		input.AwsService = aws.String(v.(string))
	}

	if v := d.Get("publicly_advertisable"); v != "" && addressFamily == ec2.AddressFamilyIpv6 {
		input.PubliclyAdvertisable = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("source_ipam_pool_id"); ok {
		input.SourceIpamPoolId = aws.String(v.(string))
	}

	output, err := conn.CreateIpamPool(input)

	if err != nil {
		return fmt.Errorf("creating IPAM Pool: %w", err)
	}

	d.SetId(aws.StringValue(output.IpamPool.IpamPoolId))

	if _, err := WaitIPAMPoolCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for IPAM Pool (%s) create: %w", d.Id(), err)
	}

	return ResourceIPAMPoolRead(d, meta)
}

func ResourceIPAMPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	pool, err := FindIPAMPoolByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading IPAM Pool (%s): %w", d.Id(), err)
	}

	d.Set("address_family", pool.AddressFamily)
	d.Set("allocation_resource_tags", KeyValueTags(tagsFromIPAMAllocationTags(pool.AllocationResourceTags)).Map())
	d.Set("arn", pool.IpamPoolArn)
	d.Set("auto_import", pool.AutoImport)
	d.Set("aws_service", pool.AwsService)
	d.Set("description", pool.Description)
	scopeID := strings.Split(aws.StringValue(pool.IpamScopeArn), "/")[1]
	d.Set("ipam_scope_id", scopeID)
	d.Set("ipam_scope_type", pool.IpamScopeType)
	d.Set("locale", pool.Locale)
	d.Set("pool_depth", pool.PoolDepth)
	d.Set("publicly_advertisable", pool.PubliclyAdvertisable)
	d.Set("source_ipam_pool_id", pool.SourceIpamPoolId)
	d.Set("state", pool.State)

	tags := KeyValueTags(pool.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func ResourceIPAMPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyIpamPoolInput{
			IpamPoolId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("allocation_default_netmask_length"); ok {
			input.AllocationDefaultNetmaskLength = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("allocation_max_netmask_length"); ok {
			input.AllocationMaxNetmaskLength = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("allocation_min_netmask_length"); ok {
			input.AllocationMinNetmaskLength = aws.Int64(int64(v.(int)))
		}

		if d.HasChange("allocation_resource_tags") {
			o, n := d.GetChange("allocation_resource_tags")
			oldTags := tftags.New(o)
			newTags := tftags.New(n)

			if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
				input.RemoveAllocationResourceTags = ipamResourceTags(removedTags.IgnoreAWS())
			}

			if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
				input.AddAllocationResourceTags = ipamResourceTags(updatedTags.IgnoreAWS())
			}
		}

		if v, ok := d.GetOk("auto_import"); ok {
			input.AutoImport = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.ModifyIpamPool(input)

		if err != nil {
			return fmt.Errorf("updating IPAM Pool (%s): %w", d.Id(), err)
		}

		if _, err := WaitIPAMPoolUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for IPAM Pool (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating IPAM Pool (%s) tags: %w", d.Id(), err)
		}
	}

	return ResourceIPAMPoolRead(d, meta)
}

func ResourceIPAMPoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting IPAM Pool: %s", d.Id())
	_, err := conn.DeleteIpamPool(&ec2.DeleteIpamPoolInput{
		IpamPoolId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IPAM Pool (%s): %w", d.Id(), err)
	}

	if _, err = WaitIPAMPoolDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for IPAM Pool (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func ipamResourceTags(tags tftags.KeyValueTags) []*ec2.RequestIpamResourceTag {
	result := make([]*ec2.RequestIpamResourceTag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &ec2.RequestIpamResourceTag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

func tagsFromIPAMAllocationTags(rts []*ec2.IpamResourceTag) []*ec2.Tag {
	if len(rts) == 0 {
		return nil
	}

	tags := []*ec2.Tag{}
	for _, ts := range rts {
		tags = append(tags, &ec2.Tag{
			Key:   ts.Key,
			Value: ts.Value,
		})
	}

	return tags
}
