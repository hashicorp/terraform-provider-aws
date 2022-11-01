package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAM() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPAMCreate,
		Read:   resourceIPAMRead,
		Update: resourceIPAMUpdate,
		Delete: resourceIPAMDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cascade": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"operating_regions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidRegionName,
						},
					},
				},
			},
			"private_default_scope_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_default_scope_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scope_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				if diff.Id() == "" { // Create.
					currentRegion := meta.(*conns.AWSClient).Region

					for _, v := range diff.Get("operating_regions").(*schema.Set).List() {
						if v.(map[string]interface{})["region_name"].(string) == currentRegion {
							return nil
						}
					}

					return fmt.Errorf("`operating_regions` must include %s", currentRegion)
				}

				return nil
			},
		),
	}
}

func resourceIPAMCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateIpamInput{
		ClientToken:       aws.String(resource.UniqueId()),
		OperatingRegions:  expandIPAMOperatingRegions(d.Get("operating_regions").(*schema.Set).List()),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeIpam),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIpam(input)

	if err != nil {
		return fmt.Errorf("creating IPAM: %w", err)
	}

	d.SetId(aws.StringValue(output.Ipam.IpamId))

	if _, err := WaitIPAMCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for IPAM (%s) created: %w", d.Id(), err)
	}

	return resourceIPAMRead(d, meta)
}

func resourceIPAMRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ipam, err := FindIPAMByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading IPAM (%s): %w", d.Id(), err)
	}

	d.Set("arn", ipam.IpamArn)
	d.Set("description", ipam.Description)
	d.Set("operating_regions", flattenIPAMOperatingRegions(ipam.OperatingRegions))
	d.Set("public_default_scope_id", ipam.PublicDefaultScopeId)
	d.Set("private_default_scope_id", ipam.PrivateDefaultScopeId)
	d.Set("scope_count", ipam.ScopeCount)

	tags := KeyValueTags(ipam.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceIPAMUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyIpamInput{
			IpamId: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("operating_regions") {
			o, n := d.GetChange("operating_regions")
			if o == nil {
				o = new(schema.Set)
			}
			if n == nil {
				n = new(schema.Set)
			}

			os := o.(*schema.Set)
			ns := n.(*schema.Set)
			operatingRegionUpdateAdd := expandIPAMOperatingRegionsUpdateAddRegions(ns.Difference(os).List())
			operatingRegionUpdateRemove := expandIPAMOperatingRegionsUpdateDeleteRegions(os.Difference(ns).List())

			if len(operatingRegionUpdateAdd) != 0 {
				input.AddOperatingRegions = operatingRegionUpdateAdd
			}

			if len(operatingRegionUpdateRemove) != 0 {
				input.RemoveOperatingRegions = operatingRegionUpdateRemove
			}
		}

		_, err := conn.ModifyIpam(input)

		if err != nil {
			return fmt.Errorf("updating IPAM (%s): %w", d.Id(), err)
		}

		if _, err := WaitIPAMUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for IPAM (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating IPAM (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceIPAMDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteIpamInput{
		IpamId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("cascade"); ok {
		input.Cascade = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Deleting IPAM: %s", d.Id())
	_, err := conn.DeleteIpam(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IPAM: (%s): %w", d.Id(), err)
	}

	if _, err := WaitIPAMDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for IPAM (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandIPAMOperatingRegions(operatingRegions []interface{}) []*ec2.AddIpamOperatingRegion {
	regions := make([]*ec2.AddIpamOperatingRegion, 0, len(operatingRegions))
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		regions = append(regions, expandIPAMOperatingRegion(region))
	}

	return regions
}

func expandIPAMOperatingRegion(operatingRegion map[string]interface{}) *ec2.AddIpamOperatingRegion {
	region := &ec2.AddIpamOperatingRegion{
		RegionName: aws.String(operatingRegion["region_name"].(string)),
	}
	return region
}

func flattenIPAMOperatingRegions(operatingRegions []*ec2.IpamOperatingRegion) []interface{} {
	regions := []interface{}{}
	for _, operatingRegion := range operatingRegions {
		regions = append(regions, flattenIPAMOperatingRegion(operatingRegion))
	}
	return regions
}

func flattenIPAMOperatingRegion(operatingRegion *ec2.IpamOperatingRegion) map[string]interface{} {
	region := make(map[string]interface{})
	region["region_name"] = aws.StringValue(operatingRegion.RegionName)
	return region
}

func expandIPAMOperatingRegionsUpdateAddRegions(operatingRegions []interface{}) []*ec2.AddIpamOperatingRegion {
	regionUpdates := make([]*ec2.AddIpamOperatingRegion, 0, len(operatingRegions))
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		regionUpdates = append(regionUpdates, expandIPAMOperatingRegionsUpdateAddRegion(region))
	}
	return regionUpdates
}

func expandIPAMOperatingRegionsUpdateAddRegion(operatingRegion map[string]interface{}) *ec2.AddIpamOperatingRegion {
	regionUpdate := &ec2.AddIpamOperatingRegion{
		RegionName: aws.String(operatingRegion["region_name"].(string)),
	}
	return regionUpdate
}

func expandIPAMOperatingRegionsUpdateDeleteRegions(operatingRegions []interface{}) []*ec2.RemoveIpamOperatingRegion {
	regionUpdates := make([]*ec2.RemoveIpamOperatingRegion, 0, len(operatingRegions))
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		regionUpdates = append(regionUpdates, expandIPAMOperatingRegionsUpdateDeleteRegion(region))
	}
	return regionUpdates
}

func expandIPAMOperatingRegionsUpdateDeleteRegion(operatingRegion map[string]interface{}) *ec2.RemoveIpamOperatingRegion {
	regionUpdate := &ec2.RemoveIpamOperatingRegion{
		RegionName: aws.String(operatingRegion["region_name"].(string)),
	}
	return regionUpdate
}
