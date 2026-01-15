// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_resource_discovery", name="IPAM Resource Discovery")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceIPAMResourceDiscovery() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
			"organizational_unit_exclusion": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organizations_entity_path": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			// user must define authn region within `operating_regions {}`
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

func resourceIPAMResourceDiscoveryCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.CreateIpamResourceDiscoveryInput{
		ClientToken:       aws.String(sdkid.UniqueId()),
		OperatingRegions:  expandIPAMOperatingRegions(d.Get("operating_regions").(*schema.Set).List()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeIpamResourceDiscovery),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIpamResourceDiscovery(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Resource Discovery: %s", err)
	}

	d.SetId(aws.ToString(output.IpamResourceDiscovery.IpamResourceDiscoveryId))

	if _, err := waitIPAMResourceDiscoveryCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Resource Discovery (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("organizational_unit_exclusion"); ok && v.(*schema.Set).Len() > 0 {
		input := ec2.ModifyIpamResourceDiscoveryInput{
			AddOrganizationalUnitExclusions: expandAddIPAMOrganizationalUnitExclusions(v.(*schema.Set).List()),
			IpamResourceDiscoveryId:         aws.String(d.Id()),
		}
		if err := updateIPAMResourceDiscovery(ctx, conn, &input, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceIPAMResourceDiscoveryRead(ctx, d, meta)...)
}

func resourceIPAMResourceDiscoveryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	rd, err := findIPAMResourceDiscoveryByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IPAM Resource Discovery (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Resource Discovery (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, rd.IpamResourceDiscoveryArn)
	d.Set(names.AttrDescription, rd.Description)
	d.Set("ipam_resource_discovery_region", rd.IpamResourceDiscoveryRegion)
	d.Set("is_default", rd.IsDefault)
	if err := d.Set("operating_regions", flattenIPAMOperatingRegions(rd.OperatingRegions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting operating_regions: %s", err)
	}
	if len(rd.OrganizationalUnitExclusions) > 0 {
		if err := d.Set("organizational_unit_exclusion", flattenIPAMOrganizationalUnitExclusions(rd.OrganizationalUnitExclusions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting organizational_unit_exclusion: %s", err)
		}
	}
	d.Set(names.AttrOwnerID, rd.OwnerId)

	setTagsOut(ctx, rd.Tags)

	return diags
}

func resourceIPAMResourceDiscoveryUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := ec2.ModifyIpamResourceDiscoveryInput{
			IpamResourceDiscoveryId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
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

		if d.HasChange("organizational_unit_exclusion") {
			o, n := d.GetChange("organizational_unit_exclusion")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			if v := expandAddIPAMOrganizationalUnitExclusions(ns.Difference(os).List()); len(v) != 0 {
				input.AddOrganizationalUnitExclusions = v
			}

			if v := expandRemoveIPAMOrganizationalUnitExclusions(os.Difference(ns).List()); len(v) != 0 {
				input.RemoveOrganizationalUnitExclusions = v
			}
		}

		if err := updateIPAMResourceDiscovery(ctx, conn, &input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceIPAMResourceDiscoveryRead(ctx, d, meta)...)
}

func resourceIPAMResourceDiscoveryDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting IPAM Resource Discovery: %s", d.Id())
	input := ec2.DeleteIpamResourceDiscoveryInput{
		IpamResourceDiscoveryId: aws.String(d.Id()),
	}
	_, err := conn.DeleteIpamResourceDiscovery(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMResourceDiscoveryIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Resource Discovery: (%s): %s", d.Id(), err)
	}

	if _, err := waitIPAMResourceDiscoveryDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Resource Discovery (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func updateIPAMResourceDiscovery(ctx context.Context, conn *ec2.Client, input *ec2.ModifyIpamResourceDiscoveryInput, timeout time.Duration) error {
	id := aws.ToString(input.IpamResourceDiscoveryId)

	// https://docs.aws.amazon.com/vpc/latest/ipam/exclude-ous.html#exclude-ous-create-delete:
	// "It takes time for IPAM to discover recently created organizational units".
	err := tfresource.Retry(ctx, ec2PropagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		_, err := conn.ModifyIpamResourceDiscovery(ctx, input)

		if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "One or more of the organizations entity paths is invalid") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	}, tfresource.WithDelay(1*time.Minute), tfresource.WithPollInterval(20*time.Second))

	if err != nil {
		return fmt.Errorf("modifying IPAM Resource Discovery (%s): %w", id, err)
	}

	if _, err := waitIPAMResourceDiscoveryUpdated(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for IPAM Resource Discovery (%s) update: %w", id, err)
	}

	return nil
}

func flattenIPAMOrganizationalUnitExclusions(apiObjects []awstypes.IpamOrganizationalUnitExclusion) []any {
	tfList := []any{}
	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenIPAMOrganizationalUnitExclusion(apiObject))
	}
	return tfList
}

func flattenIPAMOrganizationalUnitExclusion(apiObject awstypes.IpamOrganizationalUnitExclusion) map[string]any {
	tfMap := make(map[string]any)
	tfMap["organizations_entity_path"] = aws.ToString(apiObject.OrganizationsEntityPath)
	return tfMap
}

func expandAddIPAMOrganizationalUnitExclusions(tfList []any) []awstypes.AddIpamOrganizationalUnitExclusion {
	apiObjects := make([]awstypes.AddIpamOrganizationalUnitExclusion, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, expandAddIPAMOrganizationalUnitExclusion(tfMap))
	}
	return apiObjects
}

func expandAddIPAMOrganizationalUnitExclusion(tfMap map[string]any) awstypes.AddIpamOrganizationalUnitExclusion {
	apiObject := awstypes.AddIpamOrganizationalUnitExclusion{
		OrganizationsEntityPath: aws.String(tfMap["organizations_entity_path"].(string)),
	}
	return apiObject
}

func expandRemoveIPAMOrganizationalUnitExclusions(tfList []any) []awstypes.RemoveIpamOrganizationalUnitExclusion {
	apiObjects := make([]awstypes.RemoveIpamOrganizationalUnitExclusion, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, expandRemoveIPAMOrganizationalUnitExclusion(tfMap))
	}
	return apiObjects
}

func expandRemoveIPAMOrganizationalUnitExclusion(tfMap map[string]any) awstypes.RemoveIpamOrganizationalUnitExclusion {
	apiObject := awstypes.RemoveIpamOrganizationalUnitExclusion{
		OrganizationsEntityPath: aws.String(tfMap["organizations_entity_path"].(string)),
	}
	return apiObject
}
