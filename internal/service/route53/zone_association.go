// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_zone_association", name="Zone Association")
func resourceZoneAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceZoneAssociationCreate,
		ReadWithoutTimeout:   resourceZoneAssociationRead,
		DeleteWithoutTimeout: resourceZoneAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func resourceZoneAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	vpcRegion := meta.(*conns.AWSClient).Region
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
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			// AccessDenied errors likely due to cross-account issue.
			if !tfawserr.ErrCodeEquals(err, errCodeAccessDenied) {
				return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Zone Association (%s) synchronize: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceZoneAssociationRead(ctx, d, meta)...)
}

func resourceZoneAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		vpcRegion = meta.(*conns.AWSClient).Region
	}

	hostedZoneSummary, err := findZoneAssociationByThreePartKey(ctx, conn, zoneID, vpcID, vpcRegion)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Zone Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Zone Association (%s): %s", d.Id(), err)
	}

	d.Set("owning_account", hostedZoneSummary.Owner.OwningAccount)
	d.Set(names.AttrVPCID, vpcID)
	d.Set("vpc_region", vpcRegion)
	d.Set("zone_id", hostedZoneSummary.HostedZoneId)

	return diags
}

func resourceZoneAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		vpcRegion = meta.(*conns.AWSClient).Region
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
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
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
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
