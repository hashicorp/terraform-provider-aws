// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package route53

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_zone_association", name="Zone Association")
// @IdentityAttribute("zone_id")
// @IdentityAttribute("vpc_id")
// @IdentityAttribute("vpc_region", optional="true", testNotNull="true")
// @ImportIDHandler("zoneAssociationImportID")
// @Testing(preIdentityVersion="v6.45.0")
func resourceZoneAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceZoneAssociationCreate,
		ReadWithoutTimeout:   resourceZoneAssociationRead,
		DeleteWithoutTimeout: resourceZoneAssociationDelete,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"owning_account": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrVPCID: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"vpc_region": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.VPCRegion](),
				},
				"zone_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
			}
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceZoneAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	vpcRegion := meta.(*conns.AWSClient).Region(ctx)
	if v, ok := d.GetOk("vpc_region"); ok {
		vpcRegion = v.(string)
	}
	vpcID := d.Get(names.AttrVPCID).(string)
	zoneID := d.Get("zone_id").(string)
	id := zoneAssociationCreateResourceID(zoneID, vpcID, vpcRegion)
	input := &route53.AssociateVPCWithHostedZoneInput{
		Comment:      aws.String("Managed by Terraform"),
		HostedZoneId: aws.String(zoneID),
		VPC: &awstypes.VPC{
			VPCId:     aws.String(vpcID),
			VPCRegion: awstypes.VPCRegion(vpcRegion),
		},
	}

	output, err := conn.AssociateVPCWithHostedZone(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Zone Association (%s): %s", id, err)
	}

	d.SetId(id)

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id), d.Timeout(schema.TimeoutCreate)); err != nil {
			// AccessDenied errors likely due to cross-account issue.
			if !tfawserr.ErrCodeEquals(err, errCodeAccessDenied) {
				return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Zone Association (%s) synchronize: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceZoneAssociationRead(ctx, d, meta)...)
}

func resourceZoneAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	zoneID, vpcID, vpcRegion, err := zoneAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// Continue supporting older resources without VPC Region in ID.
	if vpcRegion == "" {
		vpcRegion = d.Get("vpc_region").(string)
	}
	if vpcRegion == "" {
		vpcRegion = meta.(*conns.AWSClient).Region(ctx)
	}

	hostedZoneSummary, err := findZoneAssociationByThreePartKey(ctx, conn, zoneID, vpcID, vpcRegion)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Route 53 Zone Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Zone Association (%s): %s", d.Id(), err)
	}

	resourceZoneAssociationFlatten(d, hostedZoneSummary, vpcID, vpcRegion)

	return diags
}

func resourceZoneAssociationFlatten(d *schema.ResourceData, summary *awstypes.HostedZoneSummary, vpcID, vpcRegion string) {
	d.Set("owning_account", summary.Owner.OwningAccount)
	d.Set(names.AttrVPCID, vpcID)
	d.Set("vpc_region", vpcRegion)
	d.Set("zone_id", summary.HostedZoneId)
}

func resourceZoneAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	zoneID, vpcID, vpcRegion, err := zoneAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// Continue supporting older resources without VPC Region in ID.
	if vpcRegion == "" {
		vpcRegion = d.Get("vpc_region").(string)
	}
	if vpcRegion == "" {
		vpcRegion = meta.(*conns.AWSClient).Region(ctx)
	}

	log.Printf("[INFO] Deleting Route53 Zone Association: %s", d.Id())
	output, err := conn.DisassociateVPCFromHostedZone(ctx, &route53.DisassociateVPCFromHostedZoneInput{
		Comment:      aws.String("Managed by Terraform"),
		HostedZoneId: aws.String(zoneID),
		VPC: &awstypes.VPC{
			VPCId:     aws.String(vpcID),
			VPCRegion: awstypes.VPCRegion(vpcRegion),
		},
	})

	if errs.IsA[*awstypes.NoSuchHostedZone](err) || errs.IsA[*awstypes.VPCAssociationNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route 53 Zone Association (%s): %s", d.Id(), err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id), d.Timeout(schema.TimeoutDelete)); err != nil {
			// AccessDenied errors likely due to cross-account issue.
			if !tfawserr.ErrCodeEquals(err, errCodeAccessDenied) {
				return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Zone Association (%s) synchronize: %s", d.Id(), err)
			}
		}
	}

	return diags
}

const zoneAssociationResourceIDSeparator = ":"

func zoneAssociationCreateResourceID(zoneID, vpcID, vpcRegion string) string {
	parts := []string{zoneID, vpcID, vpcRegion}
	id := strings.Join(parts, zoneAssociationResourceIDSeparator)

	return id
}

func zoneAssociationParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, zoneAssociationResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ZONEID%[2]sVPCID or ZONEID%[2]sVPCID%[2]sVPCREGION", id, zoneAssociationResourceIDSeparator)
	}

	return parts[0], parts[1], "", nil
}

func findZoneAssociationByThreePartKey(ctx context.Context, conn *route53.Client, zoneID, vpcID, vpcRegion string) (*awstypes.HostedZoneSummary, error) {
	input := &route53.ListHostedZonesByVPCInput{
		VPCId:     aws.String(vpcID),
		VPCRegion: awstypes.VPCRegion(vpcRegion),
	}

	return findZoneAssociation(ctx, conn, input, func(v *awstypes.HostedZoneSummary) bool {
		return aws.ToString(v.HostedZoneId) == zoneID
	})
}

func findZoneAssociation(ctx context.Context, conn *route53.Client, input *route53.ListHostedZonesByVPCInput, filter tfslices.Predicate[*awstypes.HostedZoneSummary]) (*awstypes.HostedZoneSummary, error) {
	output, err := findZoneAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findZoneAssociations(ctx context.Context, conn *route53.Client, input *route53.ListHostedZonesByVPCInput, filter tfslices.Predicate[*awstypes.HostedZoneSummary]) ([]awstypes.HostedZoneSummary, error) {
	var output []awstypes.HostedZoneSummary

	err := listHostedZonesByVPCPages(ctx, conn, input, func(page *route53.ListHostedZonesByVPCOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.HostedZoneSummaries {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, errCodeAccessDenied, "is not owned by you") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

var _ inttypes.SDKv2ImportID = zoneAssociationImportID{}

type zoneAssociationImportID struct{}

func (zoneAssociationImportID) Create(d *schema.ResourceData) string {
	return zoneAssociationCreateResourceID(d.Get("zone_id").(string), d.Get(names.AttrVPCID).(string), d.Get("vpc_region").(string))
}

func (zoneAssociationImportID) Parse(id string) (string, map[string]any, error) {
	zoneID, vpcID, vpcRegion, err := zoneAssociationParseResourceID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"zone_id":       zoneID,
		names.AttrVPCID: vpcID,
	}
	if vpcRegion != "" {
		result["vpc_region"] = vpcRegion
	}

	return id, result, nil
}
