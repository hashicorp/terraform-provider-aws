// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
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

	for _, address := range ng.NatGatewayAddresses {
		// Length check guarantees the attributes are always set (#30865).
		if len(ng.NatGatewayAddresses) == 1 || aws.BoolValue(address.IsPrimary) {
			d.Set("allocation_id", address.AllocationId)
			d.Set("association_id", address.AssociationId)
			d.Set("network_interface_id", address.NetworkInterfaceId)
			d.Set("private_ip", address.PrivateIp)
			d.Set("public_ip", address.PublicIp)
			break
		}
	}

	d.Set("connectivity_type", ng.ConnectivityType)
	d.Set("subnet_id", ng.SubnetId)

	setTagsOut(ctx, ng.Tags)

	return nil
}

func resourceNATGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
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
