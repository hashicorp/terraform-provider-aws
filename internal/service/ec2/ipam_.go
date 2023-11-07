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

// @SDKResource("aws_vpc_ipam", name="IPAM")
// @Tags(identifierAttribute="id")
func ResourceIPAM() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMCreate,
		ReadWithoutTimeout:   resourceIPAMRead,
		UpdateWithoutTimeout: resourceIPAMUpdate,
		DeleteWithoutTimeout: resourceIPAMDelete,

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
			"cascade": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"default_resource_discovery_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_resource_discovery_id": {
				Type:     schema.TypeString,
				Computed: true,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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

func resourceIPAMCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateIpamInput{
		ClientToken:       aws.String(id.UniqueId()),
		OperatingRegions:  expandIPAMOperatingRegions(d.Get("operating_regions").(*schema.Set).List()),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeIpam),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIpamWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM: %s", err)
	}

	d.SetId(aws.StringValue(output.Ipam.IpamId))

	if _, err := WaitIPAMCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM (%s) created: %s", d.Id(), err)
	}

	return append(diags, resourceIPAMRead(ctx, d, meta)...)
}

func resourceIPAMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	ipam, err := FindIPAMByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM (%s): %s", d.Id(), err)
	}

	d.Set("arn", ipam.IpamArn)
	d.Set("default_resource_discovery_association_id", ipam.DefaultResourceDiscoveryAssociationId)
	d.Set("default_resource_discovery_id", ipam.DefaultResourceDiscoveryId)
	d.Set("description", ipam.Description)
	if err := d.Set("operating_regions", flattenIPAMOperatingRegions(ipam.OperatingRegions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting operating_regions: %s", err)
	}
	d.Set("public_default_scope_id", ipam.PublicDefaultScopeId)
	d.Set("private_default_scope_id", ipam.PrivateDefaultScopeId)
	d.Set("scope_count", ipam.ScopeCount)

	setTagsOut(ctx, ipam.Tags)

	return diags
}

func resourceIPAMUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

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

		_, err := conn.ModifyIpamWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IPAM (%s): %s", d.Id(), err)
		}

		if _, err := WaitIPAMUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IPAM (%s) update: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceIPAMDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DeleteIpamInput{
		IpamId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("cascade"); ok {
		input.Cascade = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Deleting IPAM: %s", d.Id())
	_, err := conn.DeleteIpamWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM: (%s): %s", d.Id(), err)
	}

	if _, err := WaitIPAMDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM (%s) delete: %s", d.Id(), err)
	}

	return diags
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
