package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// acceptance tests for byoip related tests are in vpc_byoip_test.go
func ResourceVPCIPv6CIDRBlockAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCIPv6CIDRBlockAssociationCreate,
		Read:   resourceVPCIPv6CIDRBlockAssociationRead,
		Delete: resourceVPCIPv6CIDRBlockAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
			// ipv6_cidr_block can be set by a value returned from IPAM or explicitly in config
			if diff.Id() != "" && diff.HasChange("ipv6_cidr_block") {
				// if netmask is set then ipv6_cidr_block is derived from ipam, ignore changes
				if diff.Get("ipv6_netmask_length") != 0 {
					return diff.Clear("ipv6_cidr_block")
				}
				return diff.ForceNew("ipv6_cidr_block")
			}
			return nil
		},
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ipv6_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validation.All(
						verify.ValidIPv6CIDRNetworkAddress,
						validation.IsCIDRNetwork(VPCCIDRMaxIPv6, VPCCIDRMaxIPv6)),
				),
			},
			// ipam parameters are not required by the API but other usage mechanisms are not implemented yet. TODO ipv6 options:
			// --amazon-provided-ipv6-cidr-block
			// --ipv6-pool
			"ipv6_ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ipv6_netmask_length": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntInSlice([]int{VPCCIDRMaxIPv6}),
				ConflictsWith: []string{"ipv6_cidr_block"},
				// This RequiredWith setting should be applied once L57 is completed
				// RequiredWith:  []string{"ipv6_ipam_pool_id"},
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceVPCIPv6CIDRBlockAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	req := &ec2.AssociateVpcCidrBlockInput{
		VpcId: aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		req.Ipv6CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_ipam_pool_id"); ok {
		req.Ipv6IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_netmask_length"); ok {
		req.Ipv6NetmaskLength = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating VPC IPv6 CIDR block association: %#v", req)
	resp, err := conn.AssociateVpcCidrBlock(req)
	if err != nil {
		return fmt.Errorf("Error creating VPC IPv6 CIDR block association: %s", err)
	}

	d.SetId(aws.StringValue(resp.Ipv6CidrBlockAssociation.AssociationId))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociating},
		Target:     []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh:    vpcIpv6CidrBlockAssociationStateRefresh(conn, d.Get("vpc_id").(string), d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for IPv6 CIDR block association (%s) to become available: %s", d.Id(), err)
	}

	return resourceVPCIPv6CIDRBlockAssociationRead(d, meta)
}

func resourceVPCIPv6CIDRBlockAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"ipv6-cidr-block-association.association-id": d.Id(),
			},
		),
	}

	log.Printf("[DEBUG] Describing VPCs: %s", input)
	output, err := conn.DescribeVpcs(input)
	if err != nil {
		return fmt.Errorf("error describing VPCs: %s", err)
	}

	if output == nil || len(output.Vpcs) == 0 || output.Vpcs[0] == nil {
		log.Printf("[WARN] IPv6 CIDR block association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	vpc := output.Vpcs[0]

	var vpcIpv6CidrBlockAssociation *ec2.VpcIpv6CidrBlockAssociation
	for _, ipv6CidrBlockAssociation := range vpc.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(ipv6CidrBlockAssociation.AssociationId) == d.Id() {
			vpcIpv6CidrBlockAssociation = ipv6CidrBlockAssociation
			break
		}
	}

	if vpcIpv6CidrBlockAssociation == nil {
		log.Printf("[WARN] IPv6 CIDR block association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("ipv6_cidr_block", vpcIpv6CidrBlockAssociation.Ipv6CidrBlock)
	d.Set("vpc_id", vpc.VpcId)

	return nil
}

func resourceVPCIPv6CIDRBlockAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting VPC IPv6 CIDR block association: %s", d.Id())
	_, err := conn.DisassociateVpcCidrBlock(&ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, "InvalidVpcID.NotFound", "") {
			return nil
		}
		return fmt.Errorf("Error deleting VPC IPv6 CIDR block association: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeDisassociating},
		Target:     []string{ec2.VpcCidrBlockStateCodeDisassociated, VpcCidrBlockStateCodeDeleted},
		Refresh:    vpcIpv6CidrBlockAssociationStateRefresh(conn, d.Get("vpc_id").(string), d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for VPC IPv6 CIDR block association (%s) to be deleted: %s", d.Id(), err.Error())
	}

	return nil
}

func vpcIpv6CidrBlockAssociationStateRefresh(conn *ec2.EC2, vpcId, assocId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpc, err := vpcDescribe(conn, vpcId)
		if err != nil {
			return nil, "", err
		}

		if vpc != nil {
			for _, ipv6CidrAssociation := range vpc.Ipv6CidrBlockAssociationSet {
				if aws.StringValue(ipv6CidrAssociation.AssociationId) == assocId {
					return ipv6CidrAssociation, aws.StringValue(ipv6CidrAssociation.Ipv6CidrBlockState.State), nil
				}
			}
		}

		return "", VpcCidrBlockStateCodeDeleted, nil
	}
}
