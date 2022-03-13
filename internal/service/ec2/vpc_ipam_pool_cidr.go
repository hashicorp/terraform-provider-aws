package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCIpamPoolCidr() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCIpamPoolCidrCreate,
		Read:   resourceVPCIpamPoolCidrRead,
		Delete: resourceVPCIpamPoolCidrDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.Any(
					verify.ValidIPv4CIDRNetworkAddress,
					verify.ValidIPv6CIDRNetworkAddress,
				),
			},
			"cidr_authorization_context": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"signature": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	IpamPoolCidrCreateTimeout = 10 * time.Minute
	// allocations releases are eventually consistent with a max time of 20m
	IpamPoolCidrDeleteTimeout  = 32 * time.Minute
	IpamPoolCidrAvailableDelay = 5 * time.Second
	IpamPoolCidrDeleteDelay    = 5 * time.Second
)

func resourceVPCIpamPoolCidrCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	pool_id := d.Get("ipam_pool_id").(string)

	input := &ec2.ProvisionIpamPoolCidrInput{
		IpamPoolId: aws.String(pool_id),
	}

	if v, ok := d.GetOk("cidr_authorization_context"); ok {
		input.CidrAuthorizationContext = expandVPCIpamPoolCidrCidrAuthorizationContext(v.([]interface{}))
	}

	if v, ok := d.GetOk("cidr"); ok {
		input.Cidr = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Provisioning IPAM Pool Cidr: %s", input)
	output, err := conn.ProvisionIpamPoolCidr(input)
	if err != nil {
		return fmt.Errorf("Error provisioning ipam pool cidr in ipam pool (%s): %w", d.Get("ipam_pool_id").(string), err)
	}

	cidr := aws.StringValue(output.IpamPoolCidr.Cidr)
	id := encodeIpamPoolCidrId(cidr, pool_id)

	if _, err = WaitIpamPoolCidrAvailable(conn, id, IpamPoolCidrCreateTimeout); err != nil {
		return fmt.Errorf("error waiting for IPAM Pool Cidr (%s) to be provision: %w", id, err)
	}

	d.SetId(id)
	return resourceVPCIpamPoolCidrRead(d, meta)
}

func resourceVPCIpamPoolCidrRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	cidr, pool_id, err := FindIpamPoolCidr(conn, d.Id())

	if err != nil {
		return err
	}

	if !d.IsNewResource() && cidr == nil {
		log.Printf("[WARN] IPAM Pool Cidr (%s) not found or was deprovisioned, removing from state", cidr)
		d.SetId("")
		return nil
	}

	if aws.StringValue(cidr.State) == ec2.IpamPoolCidrStateDeprovisioned {
		log.Printf("[WARN] IPAM Pool Cidr (%s) not found or was deprovisioned, removing from state", cidr)
		d.SetId("")
		return nil
	}

	d.Set("cidr", cidr.Cidr)
	// pool id is not returned in describe, adding from concatenated id
	d.Set("ipam_pool_id", pool_id)

	return nil
}

func resourceVPCIpamPoolCidrDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	cidr, pool_id, err := DecodeIpamPoolCidrID(d.Id())

	input := &ec2.DeprovisionIpamPoolCidrInput{
		Cidr:       aws.String(cidr),
		IpamPoolId: aws.String(pool_id),
	}
	return resource.Retry(IpamPoolCidrDeleteTimeout, func() *resource.RetryError {
		log.Printf("[DEBUG] Deprovisioning IPAM Pool Cidr: %s", input)
		// releasing allocations is eventually consistent and can cause deprovisioning to fail
		_, err = conn.DeprovisionIpamPoolCidr(input)

		if err != nil {
			// IncorrectState err can mean: State = "deprovisioned" || State = "pending-deprovision"
			if tfawserr.ErrCodeEquals(err, "IncorrectState") {
				output, err := WaitIpamPoolCidrDeleted(conn, d.Id(), IpamPoolCidrDeleteTimeout)
				if err != nil {
					// State = failed-deprovision
					return resource.RetryableError(fmt.Errorf("Expected cidr to be deprovisioned but was in state %s", aws.StringValue(output.State)))
				}
				// State = deprovisioned
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error deprovisioning ipam pool cidr: (%s): %w", cidr, err))
		}

		output, err := WaitIpamPoolCidrDeleted(conn, d.Id(), IpamPoolCidrDeleteTimeout)
		if err != nil {
			// State = failed-deprovision
			return resource.RetryableError(fmt.Errorf("Expected cidr to be deprovisioned but was in state %s", aws.StringValue(output.State)))
		}
		// State = deprovisioned
		return nil
	})
}

func FindIpamPoolCidr(conn *ec2.EC2, id string) (*ec2.IpamPoolCidr, string, error) {
	cidr_block, pool_id, err := DecodeIpamPoolCidrID(id)
	if err != nil {
		return nil, "", fmt.Errorf("error decoding ID (%s): %w", cidr_block, err)
	}
	input := &ec2.GetIpamPoolCidrsInput{
		IpamPoolId: aws.String(pool_id),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("cidr"),
				Values: aws.StringSlice([]string{cidr_block}),
			},
		},
	}

	output, err := conn.GetIpamPoolCidrs(input)

	if err != nil {
		return nil, "", err
	}

	if output == nil || len(output.IpamPoolCidrs) == 0 || output.IpamPoolCidrs[0] == nil {
		return nil, "", nil
	}

	return output.IpamPoolCidrs[0], pool_id, nil
}

func WaitIpamPoolCidrAvailable(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.IpamPoolCidr, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.IpamPoolCidrStatePendingProvision},
		Target:  []string{ec2.IpamPoolCidrStateProvisioned},
		Refresh: statusIpamPoolCidrStatus(conn, id),
		Timeout: timeout,
		Delay:   IpamPoolCidrAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.IpamPoolCidr); ok {
		return output, err
	}

	return nil, err
}

func WaitIpamPoolCidrDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.IpamPoolCidr, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.IpamPoolCidrStatePendingDeprovision, ec2.IpamPoolCidrStateProvisioned},
		Target:  []string{ec2.IpamPoolCidrStateDeprovisioned},
		Refresh: statusIpamPoolCidrStatus(conn, id),
		Timeout: timeout,
		Delay:   IpamPoolCidrDeleteDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.IpamPoolCidr); ok {
		return output, err
	}

	return nil, err
}

func statusIpamPoolCidrStatus(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		output, _, err := FindIpamPoolCidr(conn, id)

		// there was an unhandled error in the Finder
		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func encodeIpamPoolCidrId(cidr, pool_id string) string {
	return fmt.Sprintf("%s_%s", cidr, pool_id)
}

func DecodeIpamPoolCidrID(id string) (string, string, error) {
	idParts := strings.Split(id, "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form of cidr_poolId, given: %q", id)
	}
	return idParts[0], idParts[1], nil
}

func expandVPCIpamPoolCidrCidrAuthorizationContext(l []interface{}) *ec2.IpamCidrAuthorizationContext {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	cac := &ec2.IpamCidrAuthorizationContext{
		Message:   aws.String(m["message"].(string)),
		Signature: aws.String(m["signature"].(string)),
	}

	return cac
}
