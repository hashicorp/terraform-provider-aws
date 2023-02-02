package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHostedPrivateVirtualInterfaceAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedPrivateVirtualInterfaceAccepterCreate,
		ReadWithoutTimeout:   resourceHostedPrivateVirtualInterfaceAccepterRead,
		UpdateWithoutTimeout: resourceHostedPrivateVirtualInterfaceAccepterUpdate,
		DeleteWithoutTimeout: resourceHostedPrivateVirtualInterfaceAccepterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceHostedPrivateVirtualInterfaceAccepterImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"vpn_gateway_id"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"dx_gateway_id"},
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHostedPrivateVirtualInterfaceAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	vgwIdRaw, vgwOk := d.GetOk("vpn_gateway_id")
	dxgwIdRaw, dxgwOk := d.GetOk("dx_gateway_id")
	if vgwOk == dxgwOk {
		return sdkdiag.AppendErrorf(diags, "One of ['vpn_gateway_id', 'dx_gateway_id'] must be set to create a Direct Connect private virtual interface accepter")
	}

	vifId := d.Get("virtual_interface_id").(string)
	req := &directconnect.ConfirmPrivateVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(vifId),
	}
	if dxgwOk && dxgwIdRaw.(string) != "" {
		req.DirectConnectGatewayId = aws.String(dxgwIdRaw.(string))
	}
	if vgwOk && vgwIdRaw.(string) != "" {
		req.VirtualGatewayId = aws.String(vgwIdRaw.(string))
	}

	log.Printf("[DEBUG] Accepting Direct Connect hosted private virtual interface: %s", req)
	_, err := conn.ConfirmPrivateVirtualInterfaceWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Direct Connect hosted private virtual interface: %s", err)
	}

	d.SetId(vifId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "directconnect",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	if err := hostedPrivateVirtualInterfaceAccepterWaitUntilAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceHostedPrivateVirtualInterfaceAccepterUpdate(ctx, d, meta)...)
}

func resourceHostedPrivateVirtualInterfaceAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vif, err := virtualInterfaceRead(ctx, d.Id(), conn)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect hosted private virtual interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	vifState := aws.StringValue(vif.VirtualInterfaceState)
	if vifState != directconnect.VirtualInterfaceStateAvailable &&
		vifState != directconnect.VirtualInterfaceStateDown {
		log.Printf("[WARN] Direct Connect hosted private virtual interface (%s) is '%s', removing from state", vifState, d.Id())
		d.SetId("")
		return diags
	}

	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	d.Set("virtual_interface_id", vif.VirtualInterfaceId)
	d.Set("vpn_gateway_id", vif.VirtualGatewayId)

	arn := d.Get("arn").(string)
	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Direct Connect hosted private virtual interface (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceHostedPrivateVirtualInterfaceAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, virtualInterfaceUpdate(ctx, d, meta)...)
	if diags.HasError() {
		return diags
	}

	return append(diags, resourceHostedPrivateVirtualInterfaceAccepterRead(ctx, d, meta)...)
}

func resourceHostedPrivateVirtualInterfaceAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[WARN] Will not delete Direct Connect virtual interface. Terraform will remove this resource from the state file, however resources may remain.")
	return diags
}

func resourceHostedPrivateVirtualInterfaceAccepterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	vif, err := virtualInterfaceRead(ctx, d.Id(), conn)
	if err != nil {
		return nil, err
	}
	if vif == nil {
		return nil, fmt.Errorf("virtual interface (%s) not found", d.Id())
	}

	if vifType := aws.StringValue(vif.VirtualInterfaceType); vifType != "private" {
		return nil, fmt.Errorf("virtual interface (%s) has incorrect type: %s", d.Id(), vifType)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "directconnect",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return []*schema.ResourceData{d}, nil
}

func hostedPrivateVirtualInterfaceAccepterWaitUntilAvailable(ctx context.Context, conn *directconnect.DirectConnect, vifId string, timeout time.Duration) error {
	return virtualInterfaceWaitUntilAvailable(ctx, conn,
		vifId,
		timeout,
		[]string{
			directconnect.VirtualInterfaceStateConfirming,
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateDown,
		})
}
