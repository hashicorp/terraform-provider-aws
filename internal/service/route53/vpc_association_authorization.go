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

// @SDKResource("aws_route53_vpc_association_authorization", name="VPC Association Authorization")
func resourceVPCAssociationAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCAssociationAuthorizationCreate,
		ReadWithoutTimeout:   resourceVPCAssociationAuthorizationRead,
		DeleteWithoutTimeout: resourceVPCAssociationAuthorizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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

func resourceVPCAssociationAuthorizationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	input := &route53.CreateVPCAssociationAuthorizationInput{
		HostedZoneId: aws.String(d.Get("zone_id").(string)),
		VPC: &awstypes.VPC{
			VPCId:     aws.String(d.Get(names.AttrVPCID).(string)),
			VPCRegion: awstypes.VPCRegion(meta.(*conns.AWSClient).Region),
		},
	}

	if v, ok := d.GetOk("vpc_region"); ok {
		input.VPC.VPCRegion = awstypes.VPCRegion(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ConcurrentModification](ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.CreateVPCAssociationAuthorization(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 VPC Association Authorization: %s", err)
	}

	output := outputRaw.(*route53.CreateVPCAssociationAuthorizationOutput)

	d.SetId(vpcAssociationAuthorizationCreateResourceID(aws.ToString(output.HostedZoneId), aws.ToString(output.VPC.VPCId)))

	return append(diags, resourceVPCAssociationAuthorizationRead(ctx, d, meta)...)
}

func resourceVPCAssociationAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	zoneID, vpcID, err := vpcAssociationAuthorizationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	vpc, err := findVPCAssociationAuthorizationByTwoPartKey(ctx, conn, zoneID, vpcID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 VPC Association Authorization %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 VPC Association Authorization (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrVPCID, vpc.VPCId)
	d.Set("vpc_region", vpc.VPCRegion)
	d.Set("zone_id", zoneID)

	return diags
}

func resourceVPCAssociationAuthorizationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	zoneID, vpcID, err := vpcAssociationAuthorizationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Route53 VPC Association Authorization: %s", d.Id())
	_, err = tfresource.RetryWhenIsA[*awstypes.ConcurrentModification](ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.DeleteVPCAssociationAuthorization(ctx, &route53.DeleteVPCAssociationAuthorizationInput{
			HostedZoneId: aws.String(zoneID),
			VPC: &awstypes.VPC{
				VPCId:     aws.String(vpcID),
				VPCRegion: awstypes.VPCRegion(d.Get("vpc_region").(string)),
			},
		})
	})

	if errs.IsA[*awstypes.NoSuchHostedZone](err) || errs.IsA[*awstypes.VPCAssociationAuthorizationNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 VPC Association Authorization (%s): %s", d.Id(), err)
	}

	return diags
}

const vpcAssociationAuthorizationResourceIDSeparator = ":"

func vpcAssociationAuthorizationCreateResourceID(zoneID, vpcID string) string {
	parts := []string{zoneID, vpcID}
	id := strings.Join(parts, vpcAssociationAuthorizationResourceIDSeparator)

	return id
}

func vpcAssociationAuthorizationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, vpcAssociationAuthorizationResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ZONEID%[2]sVPCID", id, vpcAssociationAuthorizationResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findVPCAssociationAuthorizationByTwoPartKey(ctx context.Context, conn *route53.Client, zoneID, vpcID string) (*awstypes.VPC, error) {
	input := &route53.ListVPCAssociationAuthorizationsInput{
		HostedZoneId: aws.String(zoneID),
	}

	return findVPCAssociationAuthorization(ctx, conn, input, func(v *awstypes.VPC) bool {
		return aws.ToString(v.VPCId) == vpcID
	})
}

func findVPCAssociationAuthorization(ctx context.Context, conn *route53.Client, input *route53.ListVPCAssociationAuthorizationsInput, filter tfslices.Predicate[*awstypes.VPC]) (*awstypes.VPC, error) {
	output, err := findVPCAssociationAuthorizations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCAssociationAuthorizations(ctx context.Context, conn *route53.Client, input *route53.ListVPCAssociationAuthorizationsInput, filter tfslices.Predicate[*awstypes.VPC]) ([]awstypes.VPC, error) {
	var output []awstypes.VPC

	err := listVPCAssociationAuthorizationsPages(ctx, conn, input, func(page *route53.ListVPCAssociationAuthorizationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VPCs {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.NoSuchHostedZone](err) {
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
