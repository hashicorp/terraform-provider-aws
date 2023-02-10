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

func ResourceHostedTransitVirtualInterfaceAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedTransitVirtualInterfaceAccepterCreate,
		ReadWithoutTimeout:   resourceHostedTransitVirtualInterfaceAccepterRead,
		UpdateWithoutTimeout: resourceHostedTransitVirtualInterfaceAccepterUpdate,
		DeleteWithoutTimeout: resourceHostedTransitVirtualInterfaceAccepterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceHostedTransitVirtualInterfaceAccepterImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHostedTransitVirtualInterfaceAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	vifId := d.Get("virtual_interface_id").(string)
	req := &directconnect.ConfirmTransitVirtualInterfaceInput{
		DirectConnectGatewayId: aws.String(d.Get("dx_gateway_id").(string)),
		VirtualInterfaceId:     aws.String(vifId),
	}

	log.Printf("[DEBUG] Accepting Direct Connect hosted transit virtual interface: %s", req)
	_, err := conn.ConfirmTransitVirtualInterfaceWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Direct Connect hosted transit virtual interface (%s): %s", vifId, err)
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

	if err := hostedTransitVirtualInterfaceAccepterWaitUntilAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceHostedTransitVirtualInterfaceAccepterUpdate(ctx, d, meta)...)
}

func resourceHostedTransitVirtualInterfaceAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vif, err := virtualInterfaceRead(ctx, d.Id(), conn)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect transit virtual interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	vifState := aws.StringValue(vif.VirtualInterfaceState)
	if vifState != directconnect.VirtualInterfaceStateAvailable && vifState != directconnect.VirtualInterfaceStateDown {
		log.Printf("[WARN] Direct Connect virtual interface (%s) is '%s', removing from state", vifState, d.Id())
		d.SetId("")
		return diags
	}

	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	d.Set("virtual_interface_id", vif.VirtualInterfaceId)

	arn := d.Get("arn").(string)
	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Direct Connect hosted transit virtual interface (%s): %s", arn, err)
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

func resourceHostedTransitVirtualInterfaceAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, virtualInterfaceUpdate(ctx, d, meta)...)
	if diags.HasError() {
		return diags
	}

	return append(diags, resourceHostedTransitVirtualInterfaceAccepterRead(ctx, d, meta)...)
}

func resourceHostedTransitVirtualInterfaceAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[WARN] Will not delete Direct Connect virtual interface. Terraform will remove this resource from the state file, however resources may remain.")
	return diags
}

func resourceHostedTransitVirtualInterfaceAccepterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	vif, err := virtualInterfaceRead(ctx, d.Id(), conn)
	if err != nil {
		return nil, err
	}
	if vif == nil {
		return nil, fmt.Errorf("virtual interface (%s) not found", d.Id())
	}

	if vifType := aws.StringValue(vif.VirtualInterfaceType); vifType != "transit" {
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

func hostedTransitVirtualInterfaceAccepterWaitUntilAvailable(ctx context.Context, conn *directconnect.DirectConnect, vifId string, timeout time.Duration) error {
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
