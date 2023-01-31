package route53

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceVPCAssociationAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCAssociationAuthorizationCreate,
		ReadWithoutTimeout:   resourceVPCAssociationAuthorizationRead,
		DeleteWithoutTimeout: resourceVPCAssociationAuthorizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCAssociationAuthorizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn()

	req := &route53.CreateVPCAssociationAuthorizationInput{
		HostedZoneId: aws.String(d.Get("zone_id").(string)),
		VPC: &route53.VPC{
			VPCId:     aws.String(d.Get("vpc_id").(string)),
			VPCRegion: aws.String(meta.(*conns.AWSClient).Region),
		},
	}

	if v, ok := d.GetOk("vpc_region"); ok {
		req.VPC.VPCRegion = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route53 VPC Association Authorization for hosted zone %s with VPC %s and region %s", *req.HostedZoneId, *req.VPC.VPCId, *req.VPC.VPCRegion)
	_, err := conn.CreateVPCAssociationAuthorizationWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 VPC Association Authorization: %s", err)
	}

	// Store association id
	d.SetId(fmt.Sprintf("%s:%s", *req.HostedZoneId, *req.VPC.VPCId))

	return append(diags, resourceVPCAssociationAuthorizationRead(ctx, d, meta)...)
}

func resourceVPCAssociationAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn()

	zone_id, vpc_id, err := VPCAssociationAuthorizationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 VPC Association Authorization (%s): %s", d.Id(), err)
	}

	req := route53.ListVPCAssociationAuthorizationsInput{
		HostedZoneId: aws.String(zone_id),
	}

	for {
		log.Printf("[DEBUG] Listing Route53 VPC Association Authorizations for hosted zone %s", zone_id)
		res, err := conn.ListVPCAssociationAuthorizationsWithContext(ctx, &req)

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
			log.Printf("[WARN] Route53 VPC Association Authorization (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Route53 VPC Association Authorizations: %s", err)
		}

		for _, vpc := range res.VPCs {
			if vpc_id == aws.StringValue(vpc.VPCId) {
				d.Set("vpc_id", vpc.VPCId)
				d.Set("vpc_region", vpc.VPCRegion)
				d.Set("zone_id", zone_id)
				return diags
			}
		}

		// Loop till we find our authorization or we reach the end
		if res.NextToken != nil {
			req.NextToken = res.NextToken
		} else {
			break
		}
	}

	// no association found
	log.Printf("[WARN] Route53 VPC Association Authorization (%s) not found, removing from state", d.Id())
	d.SetId("")
	return diags
}

func resourceVPCAssociationAuthorizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn()

	zone_id, vpc_id, err := VPCAssociationAuthorizationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 VPC Association Authorization (%s): %s", d.Id(), err)
	}

	req := route53.DeleteVPCAssociationAuthorizationInput{
		HostedZoneId: aws.String(zone_id),
		VPC: &route53.VPC{
			VPCId:     aws.String(vpc_id),
			VPCRegion: aws.String(d.Get("vpc_region").(string)),
		},
	}

	_, err = conn.DeleteVPCAssociationAuthorizationWithContext(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 VPC Association Authorization (%s): %s", d.Id(), err)
	}

	return diags
}

func VPCAssociationAuthorizationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected ZONEID:VPCID", id)
	}

	return parts[0], parts[1], nil
}
