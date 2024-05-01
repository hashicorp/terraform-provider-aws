// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_resource_discovery", name="IPAM Resource Discovery")
// @Tags(identifierAttribute="id")
func ResourceIPAMResourceDiscovery() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMResourceDiscoveryCreate,
		ReadWithoutTimeout:   resourceIPAMResourceDiscoveryRead,
		UpdateWithoutTimeout: resourceIPAMResourceDiscoveryUpdate,
		DeleteWithoutTimeout: resourceIPAMResourceDiscoveryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_resource_discovery_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
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
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			// user must define authn region within `operating_regions {}`
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

func resourceIPAMResourceDiscoveryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateIpamResourceDiscoveryInput{
		ClientToken:       aws.String(id.UniqueId()),
		OperatingRegions:  expandIPAMOperatingRegions(d.Get("operating_regions").(*schema.Set).List()),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeIpamResourceDiscovery),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIpamResourceDiscoveryWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Resource Discovery: %s", err)
	}

	d.SetId(aws.StringValue(output.IpamResourceDiscovery.IpamResourceDiscoveryId))

	if _, err := WaitIPAMResourceDiscoveryAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Resource Discovery (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceIPAMResourceDiscoveryRead(ctx, d, meta)...)
}

func resourceIPAMResourceDiscoveryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	rd, err := FindIPAMResourceDiscoveryByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Resource Discovery (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Resource Discovery (%s): %s", d.Id(), err)
	}

	d.Set("arn", rd.IpamResourceDiscoveryArn)
	d.Set("description", rd.Description)
	d.Set("ipam_resource_discovery_region", rd.IpamResourceDiscoveryRegion)
	d.Set("is_default", rd.IsDefault)
	if err := d.Set("operating_regions", flattenIPAMResourceDiscoveryOperatingRegions(rd.OperatingRegions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting operating_regions: %s", err)
	}
	d.Set("owner_id", rd.OwnerId)

	setTagsOut(ctx, rd.Tags)

	return diags
}

func resourceIPAMResourceDiscoveryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyIpamResourceDiscoveryInput{
			IpamResourceDiscoveryId: aws.String(d.Id()),
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
			operatingRegionUpdateAdd := expandIPAMResourceDiscoveryOperatingRegionsUpdateAddRegions(ns.Difference(os).List())
			operatingRegionUpdateRemove := expandIPAMResourceDiscoveryOperatingRegionsUpdateDeleteRegions(os.Difference(ns).List())

			if len(operatingRegionUpdateAdd) != 0 {
				input.AddOperatingRegions = operatingRegionUpdateAdd
			}

			if len(operatingRegionUpdateRemove) != 0 {
				input.RemoveOperatingRegions = operatingRegionUpdateRemove
			}
		}

		_, err := conn.ModifyIpamResourceDiscoveryWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying IPAM Resource Discovery (%s): %s", d.Id(), err)
		}

		if _, err := WaitIPAMResourceDiscoveryUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IPAM Resource Discovery (%s) update: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceIPAMResourceDiscoveryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting IPAM Resource Discovery: %s", d.Id())
	_, err := conn.DeleteIpamResourceDiscoveryWithContext(ctx, &ec2.DeleteIpamResourceDiscoveryInput{
		IpamResourceDiscoveryId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMResourceDiscoveryIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Resource Discovery: (%s): %s", d.Id(), err)
	}

	if _, err := WaitIPAMResourceDiscoveryDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Resource Discovery (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func flattenIPAMResourceDiscoveryOperatingRegions(operatingRegions []*ec2.IpamOperatingRegion) []interface{} {
	regions := []interface{}{}
	for _, operatingRegion := range operatingRegions {
		regions = append(regions, flattenIPAMResourceDiscoveryOperatingRegion(operatingRegion))
	}
	return regions
}

func flattenIPAMResourceDiscoveryOperatingRegion(operatingRegion *ec2.IpamOperatingRegion) map[string]interface{} {
	region := make(map[string]interface{})
	region["region_name"] = aws.StringValue(operatingRegion.RegionName)
	return region
}

func expandIPAMResourceDiscoveryOperatingRegionsUpdateAddRegions(operatingRegions []interface{}) []*ec2.AddIpamOperatingRegion {
	regionUpdates := make([]*ec2.AddIpamOperatingRegion, 0, len(operatingRegions))
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		regionUpdates = append(regionUpdates, expandIPAMResourceDiscoveryOperatingRegionsUpdateAddRegion(region))
	}
	return regionUpdates
}

func expandIPAMResourceDiscoveryOperatingRegionsUpdateAddRegion(operatingRegion map[string]interface{}) *ec2.AddIpamOperatingRegion {
	regionUpdate := &ec2.AddIpamOperatingRegion{
		RegionName: aws.String(operatingRegion["region_name"].(string)),
	}
	return regionUpdate
}

func expandIPAMResourceDiscoveryOperatingRegionsUpdateDeleteRegions(operatingRegions []interface{}) []*ec2.RemoveIpamOperatingRegion {
	regionUpdates := make([]*ec2.RemoveIpamOperatingRegion, 0, len(operatingRegions))
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		regionUpdates = append(regionUpdates, expandIPAMResourceDiscoveryOperatingRegionsUpdateDeleteRegion(region))
	}
	return regionUpdates
}

func expandIPAMResourceDiscoveryOperatingRegionsUpdateDeleteRegion(operatingRegion map[string]interface{}) *ec2.RemoveIpamOperatingRegion {
	regionUpdate := &ec2.RemoveIpamOperatingRegion{
		RegionName: aws.String(operatingRegion["region_name"].(string)),
	}
	return regionUpdate
}
