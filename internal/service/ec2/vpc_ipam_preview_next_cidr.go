package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	// "github.com/hashicorp/terraform-provider-aws/internal/flex"
	// "github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCIpamPreviewNextCidr() *schema.Resource {
	return &schema.Resource{
		Create: ResourceVPCIpamPreviewNextCidrCreate,
		Read:   ResourceVPCIpamPreviewNextCidrRead,
		Delete: schema.Noop,
		// Having imports results in errors
		// 	error running import: exit status 1
		// 	Error: Error allocating cidr from IPAM pool (): InvalidParameterValue: The allocation size is too big for the pool.
		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },
		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// // temp comment out till bug is resolved
			// "disallowed_cidrs": {
			// 	Type:     schema.TypeSet,
			// 	Optional: true,
			// 	ForceNew: true,
			// 	Elem: &schema.Schema{
			// 		Type: schema.TypeString,
			// 		ValidateFunc: validation.Any(
			// 			verify.ValidIPv4CIDRNetworkAddress,
			// 			// Follow the numbers used for netmask_length
			// 			validation.IsCIDRNetwork(0, 32),
			// 		),
			// 	},
			// },
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"netmask_length": {
				// Possible netmask lengths for IPv4 addresses are 0 - 32.
				// AllocateIpamPoolCidr API
				//   - If there is no DefaultNetmaskLength allocation rule set on the pool,
				//   you must specify either the NetmaskLength or the CIDR.
				//   - If the DefaultNetmaskLength allocation rule is set on the pool,
				//   you can specify either the NetmaskLength or the CIDR and the
				//   DefaultNetmaskLength allocation rule will be ignored.
				// since there is no attribute to check if the rule is set, this attribute is required
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 32),
			},
		},
	}
}

func ResourceVPCIpamPreviewNextCidrCreate(d *schema.ResourceData, meta interface{}) error {
	poolId := d.Get("ipam_pool_id").(string)
	uniqueValue := resource.UniqueId()

	// preview will not produce an IpamPoolAllocationId. Hence use the uniqueValue instead
	d.SetId(encodeVPCIpamPreviewNextCidrID(uniqueValue, poolId))

	return ResourceVPCIpamPreviewNextCidrRead(d, meta)
}

func ResourceVPCIpamPreviewNextCidrRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	uniqueValue, poolId, err := decodeVPCIpamPreviewNextCidrID(d.Id())

	if err != nil {
		return err
	}

	input := &ec2.AllocateIpamPoolCidrInput{
		ClientToken:     aws.String(uniqueValue),
		IpamPoolId:      aws.String(poolId),
		PreviewNextCidr: aws.Bool(true),
	}

	// if v, ok := d.GetOk("disallowed_cidrs"); ok && v.(*schema.Set).Len() > 0 {
	// 	input.DisallowedCidrs = flex.ExpandStringSet(v.(*schema.Set))
	// }

	if v, ok := d.GetOk("netmask_length"); ok {
		input.NetmaskLength = aws.Int64(int64(v.(int)))
	}

	output, err := conn.AllocateIpamPoolCidr(input)

	if err != nil {
		return fmt.Errorf("Error allocating cidr from IPAM pool (%s): %w", d.Get("ipam_pool_id").(string), err)
	}

	if output == nil || output.IpamPoolAllocation == nil {
		return fmt.Errorf("error allocating from ipam pool (%s): empty response", poolId)
	}

	d.Set("cidr", output.IpamPoolAllocation.Cidr)

	return nil
}

func encodeVPCIpamPreviewNextCidrID(uniqueValue, poolId string) string {
	return fmt.Sprintf("%s_%s", uniqueValue, poolId)
}

func decodeVPCIpamPreviewNextCidrID(id string) (string, string, error) {
	idParts := strings.Split(id, "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form of uniqueValue_poolId, given: %q", id)
	}
	return idParts[0], idParts[1], nil
}
