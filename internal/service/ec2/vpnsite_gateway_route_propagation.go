package ec2

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPNGatewayRoutePropagation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPNGatewayRoutePropagationEnable,
		ReadWithoutTimeout:   resourceVPNGatewayRoutePropagationRead,
		DeleteWithoutTimeout: resourceVPNGatewayRoutePropagationDisable,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPNGatewayRoutePropagationEnable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	gatewayID := d.Get("vpn_gateway_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	err := routeTableEnableVGWRoutePropagation(ctx, conn, routeTableID, gatewayID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(VPNGatewayRoutePropagationCreateID(routeTableID, gatewayID))

	return append(diags, resourceVPNGatewayRoutePropagationRead(ctx, d, meta)...)
}

func resourceVPNGatewayRoutePropagationDisable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	routeTableID, gatewayID, err := VPNGatewayRoutePropagationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = routeTableDisableVGWRoutePropagation(ctx, conn, routeTableID, gatewayID)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return diags
	}

	return sdkdiag.AppendFromErr(diags, err)
}

func resourceVPNGatewayRoutePropagationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	routeTableID, gatewayID, err := VPNGatewayRoutePropagationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = FindVPNGatewayRoutePropagationExists(ctx, conn, routeTableID, gatewayID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table (%s) VPN Gateway (%s) route propagation not found, removing from state", routeTableID, gatewayID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table (%s) VPN Gateway (%s) route propagation: %s", routeTableID, gatewayID, err)
	}

	return diags
}
