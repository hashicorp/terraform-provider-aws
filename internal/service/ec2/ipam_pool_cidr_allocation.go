package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMPoolCIDRAllocation() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPAMPoolCIDRAllocationCreate,
		Read:   resourceIPAMPoolCIDRAllocationRead,
		Delete: resourceIPAMPoolCIDRAllocationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"netmask_length"},
				ValidateFunc: validation.Any(
					verify.ValidIPv4CIDRNetworkAddress,
					validation.IsCIDRNetwork(0, 32),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"disallowed_cidrs": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.Any(
						verify.ValidIPv4CIDRNetworkAddress,
						// Follow the numbers used for netmask_length
						validation.IsCIDRNetwork(0, 32),
					),
				},
			},
			"ipam_pool_allocation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"netmask_length": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntBetween(0, 32),
				ConflictsWith: []string{"cidr"},
			},
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	IPAMPoolAllocationNotFound = "InvalidIpamPoolCidrAllocationId.NotFound"
)

func resourceIPAMPoolCIDRAllocationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	pool_id := d.Get("ipam_pool_id").(string)

	input := &ec2.AllocateIpamPoolCidrInput{
		ClientToken: aws.String(resource.UniqueId()),
		IpamPoolId:  aws.String(pool_id),
	}

	if v, ok := d.GetOk("cidr"); ok {
		input.Cidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disallowed_cidrs"); ok && v.(*schema.Set).Len() > 0 {
		input.DisallowedCidrs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v := d.Get("netmask_length"); v != 0 {
		input.NetmaskLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating IPAM Pool Allocation: %s", input)
	output, err := conn.AllocateIpamPoolCidr(input)
	if err != nil {
		return fmt.Errorf("Error allocating cidr from IPAM pool (%s): %w", d.Get("ipam_pool_id").(string), err)
	}
	d.SetId(encodeIPAMPoolCIDRAllocationID(aws.StringValue(output.IpamPoolAllocation.IpamPoolAllocationId), pool_id))

	return resourceIPAMPoolCIDRAllocationRead(d, meta)
}

func resourceIPAMPoolCIDRAllocationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidr_allocation, pool_id, err := FindIPAMPoolCIDRAllocation(conn, d.Id())

	if err != nil {
		return err
	}

	if !d.IsNewResource() && cidr_allocation == nil {
		log.Printf("[WARN] IPAM Pool Cidr Allocation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("ipam_pool_allocation_id", cidr_allocation.IpamPoolAllocationId)
	d.Set("ipam_pool_id", pool_id)

	d.Set("cidr", cidr_allocation.Cidr)

	if cidr_allocation.ResourceId != nil {
		d.Set("resource_id", cidr_allocation.ResourceId)
	}
	if cidr_allocation.ResourceOwner != nil {
		d.Set("resource_owner", cidr_allocation.ResourceOwner)
	}
	if cidr_allocation.ResourceType != nil {
		d.Set("resource_type", cidr_allocation.ResourceType)
	}

	return nil
}

func resourceIPAMPoolCIDRAllocationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.ReleaseIpamPoolAllocationInput{
		IpamPoolAllocationId: aws.String(d.Get("ipam_pool_allocation_id").(string)),
		IpamPoolId:           aws.String(d.Get("ipam_pool_id").(string)),
		Cidr:                 aws.String(d.Get("cidr").(string)),
	}

	log.Printf("[DEBUG] Releasing IPAM Pool CIDR Allocation: %s", input)
	output, err := conn.ReleaseIpamPoolAllocation(input)
	if err != nil || !aws.BoolValue(output.Success) {
		if tfawserr.ErrCodeEquals(err, InvalidIPAMPoolIDNotFound) {
			return nil
		}
		return fmt.Errorf("error releasing IPAM CIDR Allocation: (%s): %w", d.Id(), err)
	}

	return nil
}

func FindIPAMPoolCIDRAllocation(conn *ec2.EC2, id string) (*ec2.IpamPoolAllocation, string, error) {

	allocation_id, pool_id, err := DecodeIPAMPoolCIDRAllocationID(id)
	if err != nil {
		return nil, "", fmt.Errorf("error decoding ID (%s): %w", allocation_id, err)
	}

	input := &ec2.GetIpamPoolAllocationsInput{
		IpamPoolId: aws.String(pool_id),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("ipam-pool-allocation-id"),
				Values: aws.StringSlice([]string{allocation_id}),
			},
		},
	}

	output, err := conn.GetIpamPoolAllocations(input)

	if err != nil {
		return nil, "", err
	}

	if output == nil || len(output.IpamPoolAllocations) == 0 || output.IpamPoolAllocations[0] == nil {
		return nil, "", nil
	}

	return output.IpamPoolAllocations[0], pool_id, nil
}

func encodeIPAMPoolCIDRAllocationID(allocation_id, pool_id string) string {
	return fmt.Sprintf("%s_%s", allocation_id, pool_id)
}

func DecodeIPAMPoolCIDRAllocationID(id string) (string, string, error) {
	idParts := strings.Split(id, "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form of allocationId_poolId, given: %q", id)
	}
	return idParts[0], idParts[1], nil
}
