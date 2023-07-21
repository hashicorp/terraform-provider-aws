package ec2

import (
	"context"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"golang.org/x/exp/slices"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_nat_gateway", name="NAT Gateway")
// @Tags(identifierAttribute="id")
func ResourceNATGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNATGatewayCreate,
		ReadWithoutTimeout:   resourceNATGatewayRead,
		UpdateWithoutTimeout: resourceNATGatewayUpdate,
		DeleteWithoutTimeout: resourceNATGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allocation_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connectivity_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.ConnectivityTypePublic,
				ValidateFunc: validation.StringInSlice(ec2.ConnectivityType_Values(), false),
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"secondary_private_ips": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNATGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateNatGatewayInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeNatgateway),
	}

	if v, ok := d.GetOk("allocation_id"); ok {
		input.AllocationId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("connectivity_type"); ok {
		input.ConnectivityType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("private_ip"); ok {
		input.PrivateIpAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("secondary_private_ips"); ok {
		input.SecondaryPrivateIpAddresses = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateNatGatewayWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating EC2 NAT Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.NatGateway.NatGatewayId))

	if _, err := WaitNATGatewayCreated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for EC2 NAT Gateway (%s) create: %s", d.Id(), err)
	}

	return resourceNATGatewayRead(ctx, d, meta)
}

func resourceNATGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	ng, err := FindNATGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 NAT Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EC2 NAT Gateway (%s): %s", d.Id(), err)
	}

	log.Printf("[FP] reading")
	secondaryPrivateAddresses := schema.NewSet(schema.HashString, nil)
	for _, address := range ng.NatGatewayAddresses {
		// Length check guarantees the attributes are always set (#30865).
		if len(ng.NatGatewayAddresses) == 1 || aws.BoolValue(address.IsPrimary) {
			d.Set("allocation_id", address.AllocationId)
			d.Set("association_id", address.AssociationId)
			d.Set("network_interface_id", address.NetworkInterfaceId)
			d.Set("private_ip", address.PrivateIp)
			d.Set("public_ip", address.PublicIp)
		} else if !aws.BoolValue(address.IsPrimary) &&
			address.PrivateIp != nil &&
			address.Status != nil &&
			!slices.Contains([]string{ec2.NatGatewayAddressStatusFailed, ec2.NatGatewayAddressStatusDisassociating, ec2.NatGatewayAddressStatusUnassigning}, *address.Status) {
			secondaryPrivateAddresses.Add(aws.StringValue(address.PrivateIp))
		}
	}

	d.Set("secondary_private_ips", secondaryPrivateAddresses)
	d.Set("connectivity_type", ng.ConnectivityType)
	d.Set("subnet_id", ng.SubnetId)

	setTagsOut(ctx, ng.Tags)

	return nil
}

func resourceNATGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if d.HasChanges("secondary_private_ips") {
		oRaw, nRaw := d.GetChange("secondary_private_ips")
		o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)
		add := n.Difference(o)
		del := o.Difference(n)

		if add.Len() > 0 {
			_, err := conn.AssignPrivateNatGatewayAddress(&ec2.AssignPrivateNatGatewayAddressInput{
				NatGatewayId:       aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringSet(add),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "adding secondary ip address allocations (%s): %s", d.Id(), err)
			}
		}

		if del.Len() > 0 {
			_, err := conn.UnassignPrivateNatGatewayAddress(&ec2.UnassignPrivateNatGatewayAddressInput{
				NatGatewayId:       aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringSet(del),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "adding secondary ip address allocations (%s): %s", d.Id(), err)
			}
		}
	}

	return resourceNATGatewayRead(ctx, d, meta)
}

func resourceNATGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[INFO] Deleting EC2 NAT Gateway: %s", d.Id())
	_, err := conn.DeleteNatGatewayWithContext(ctx, &ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNatGatewayNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EC2 NAT Gateway (%s): %s", d.Id(), err)
	}

	if _, err := WaitNATGatewayDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for EC2 NAT Gateway (%s) delete: %s", d.Id(), err)
	}

	return nil
}
