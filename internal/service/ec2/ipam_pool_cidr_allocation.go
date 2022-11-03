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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

func resourceIPAMPoolCIDRAllocationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	ipamPoolID := d.Get("ipam_pool_id").(string)
	input := &ec2.AllocateIpamPoolCidrInput{
		ClientToken: aws.String(resource.UniqueId()),
		IpamPoolId:  aws.String(ipamPoolID),
	}

	if v, ok := d.GetOk("cidr"); ok {
		input.Cidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disallowed_cidrs"); ok && v.(*schema.Set).Len() > 0 {
		input.DisallowedCidrs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v := d.Get("netmask_length"); v != 0 {
		input.NetmaskLength = aws.Int64(int64(v.(int)))
	}

	output, err := conn.AllocateIpamPoolCidr(input)

	if err != nil {
		return fmt.Errorf("creating IPAM Pool CIDR Allocation: %w", err)
	}
	d.SetId(IPAMPoolCIDRAllocationCreateResourceID(aws.StringValue(output.IpamPoolAllocation.IpamPoolAllocationId), ipamPoolID))

	return resourceIPAMPoolCIDRAllocationRead(d, meta)
}

func resourceIPAMPoolCIDRAllocationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	allocationID, poolID, err := IPAMPoolCIDRAllocationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	allocation, err := FindIPAMPoolAllocationByTwoPartKey(conn, allocationID, poolID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Pool CIDR Allocation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading IPAM Pool CIDR Allocation (%s): %w", d.Id(), err)
	}

	d.Set("cidr", allocation.Cidr)
	d.Set("ipam_pool_allocation_id", allocation.IpamPoolAllocationId)
	d.Set("ipam_pool_id", poolID)
	if v := allocation.ResourceId; v != nil {
		d.Set("resource_id", v)
	}
	if v := allocation.ResourceOwner; v != nil {
		d.Set("resource_owner", v)
	}
	if v := allocation.ResourceType; v != nil {
		d.Set("resource_type", v)
	}

	return nil
}

func resourceIPAMPoolCIDRAllocationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	allocationID, poolID, err := IPAMPoolCIDRAllocationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting IPAM Pool CIDR Allocation: %s", d.Id())
	_, err = conn.ReleaseIpamPoolAllocation(&ec2.ReleaseIpamPoolAllocationInput{
		Cidr:                 aws.String(d.Get("cidr").(string)),
		IpamPoolAllocationId: aws.String(allocationID),
		IpamPoolId:           aws.String(poolID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterCombination, "No allocation found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IPAM Pool CIDR Allocation (%s): %w", d.Id(), err)
	}

	return nil
}

const ipamPoolCIDRAllocationIDSeparator = "_"

func IPAMPoolCIDRAllocationCreateResourceID(allocationID, poolID string) string {
	parts := []string{allocationID, poolID}
	id := strings.Join(parts, ipamPoolCIDRAllocationIDSeparator)

	return id
}

func IPAMPoolCIDRAllocationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ipamPoolCIDRAllocationIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected allocation-id%[2]spool-id", id, ipamPoolCIDRAllocationIDSeparator)
	}

	return parts[0], parts[1], nil
}
