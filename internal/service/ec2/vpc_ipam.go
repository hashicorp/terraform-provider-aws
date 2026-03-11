// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam", name="IPAM")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceIPAM() *schema.Resource {
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
			names.AttrARN: {
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
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_private_gua": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"metered_account": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IpamMeteredAccount](),
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
			"tier": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.IpamTierAdvanced,
				ValidateDiagFunc: enum.Validate[awstypes.IpamTier](),
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
				if diff.Id() == "" { // Create.
					currentRegion := meta.(*conns.AWSClient).Region(ctx)

					for _, v := range diff.Get("operating_regions").(*schema.Set).List() {
						if v.(map[string]any)["region_name"].(string) == currentRegion {
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

func resourceIPAMCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.CreateIpamInput{
		ClientToken:       aws.String(sdkid.UniqueId()),
		OperatingRegions:  expandIPAMOperatingRegions(d.Get("operating_regions").(*schema.Set).List()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeIpam),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_private_gua"); ok {
		input.EnablePrivateGua = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("metered_account"); ok {
		input.MeteredAccount = awstypes.IpamMeteredAccount(v.(string))
	}

	if v, ok := d.GetOk("tier"); ok {
		input.Tier = awstypes.IpamTier(v.(string))
	}

	output, err := conn.CreateIpam(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM: %s", err)
	}

	d.SetId(aws.ToString(output.Ipam.IpamId))

	if _, err := waitIPAMCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM (%s) created: %s", d.Id(), err)
	}

	return append(diags, resourceIPAMRead(ctx, d, meta)...)
}

func resourceIPAMRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	ipam, err := findIPAMByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IPAM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, ipam.IpamArn)
	d.Set("default_resource_discovery_association_id", ipam.DefaultResourceDiscoveryAssociationId)
	d.Set("default_resource_discovery_id", ipam.DefaultResourceDiscoveryId)
	d.Set(names.AttrDescription, ipam.Description)
	d.Set("enable_private_gua", ipam.EnablePrivateGua)
	d.Set("metered_account", ipam.MeteredAccount)
	if err := d.Set("operating_regions", flattenIPAMOperatingRegions(ipam.OperatingRegions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting operating_regions: %s", err)
	}
	d.Set("public_default_scope_id", ipam.PublicDefaultScopeId)
	d.Set("private_default_scope_id", ipam.PrivateDefaultScopeId)
	d.Set("scope_count", ipam.ScopeCount)
	d.Set("tier", ipam.Tier)

	setTagsOut(ctx, ipam.Tags)

	return diags
}

func resourceIPAMUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := ec2.ModifyIpamInput{
			IpamId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("enable_private_gua") {
			input.EnablePrivateGua = aws.Bool(d.Get("enable_private_gua").(bool))
		}

		if d.HasChange("metered_account") {
			input.MeteredAccount = awstypes.IpamMeteredAccount(d.Get("metered_account").(string))
		}

		if d.HasChange("operating_regions") {
			o, n := d.GetChange("operating_regions")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			if v := expandAddIPAMOperatingRegions(ns.Difference(os).List()); len(v) != 0 {
				input.AddOperatingRegions = v
			}

			if v := expandRemoveIPAMOperatingRegions(os.Difference(ns).List()); len(v) != 0 {
				input.RemoveOperatingRegions = v
			}
		}

		if d.HasChange("tier") {
			input.Tier = awstypes.IpamTier(d.Get("tier").(string))
		}

		_, err := conn.ModifyIpam(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IPAM (%s): %s", d.Id(), err)
		}

		if _, err := waitIPAMUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IPAM (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPAMRead(ctx, d, meta)...)
}

func resourceIPAMDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.DeleteIpamInput{
		IpamId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("cascade"); ok {
		input.Cascade = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Deleting IPAM: %s", d.Id())
	_, err := conn.DeleteIpam(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM: (%s): %s", d.Id(), err)
	}

	if _, err := waitIPAMDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandIPAMOperatingRegions(tfList []any) []awstypes.AddIpamOperatingRegion {
	apiObjects := make([]awstypes.AddIpamOperatingRegion, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, expandIPAMOperatingRegion(tfMap))
	}

	return apiObjects
}

func expandIPAMOperatingRegion(tfMap map[string]any) awstypes.AddIpamOperatingRegion {
	apiObject := awstypes.AddIpamOperatingRegion{
		RegionName: aws.String(tfMap["region_name"].(string)),
	}
	return apiObject
}

func flattenIPAMOperatingRegions(apiObjects []awstypes.IpamOperatingRegion) []any {
	tfList := []any{}
	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenIPAMOperatingRegion(apiObject))
	}
	return tfList
}

func flattenIPAMOperatingRegion(apiObject awstypes.IpamOperatingRegion) map[string]any {
	tfMap := make(map[string]any)
	tfMap["region_name"] = aws.ToString(apiObject.RegionName)
	return tfMap
}

func expandAddIPAMOperatingRegions(tfList []any) []awstypes.AddIpamOperatingRegion {
	apiObjects := make([]awstypes.AddIpamOperatingRegion, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, expandAddIPAMOperatingRegion(tfMap))
	}
	return apiObjects
}

func expandAddIPAMOperatingRegion(tfMap map[string]any) awstypes.AddIpamOperatingRegion {
	apiObject := awstypes.AddIpamOperatingRegion{
		RegionName: aws.String(tfMap["region_name"].(string)),
	}
	return apiObject
}

func expandRemoveIPAMOperatingRegions(tfList []any) []awstypes.RemoveIpamOperatingRegion {
	apiObjects := make([]awstypes.RemoveIpamOperatingRegion, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, expandRemoveIPAMOperatingRegion(tfMap))
	}
	return apiObjects
}

func expandRemoveIPAMOperatingRegion(tfMap map[string]any) awstypes.RemoveIpamOperatingRegion {
	apiObject := awstypes.RemoveIpamOperatingRegion{
		RegionName: aws.String(tfMap["region_name"].(string)),
	}
	return apiObject
}
