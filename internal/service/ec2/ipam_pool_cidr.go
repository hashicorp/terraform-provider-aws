package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMPoolCIDR() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPAMPoolCIDRCreate,
		Read:   resourceIPAMPoolCIDRRead,
		Delete: resourceIPAMPoolCIDRDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			// Allocations release are eventually consistent with a max time of 20m.
			Delete: schema.DefaultTimeout(32 * time.Minute),
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

func resourceIPAMPoolCIDRCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	poolID := d.Get("ipam_pool_id").(string)
	input := &ec2.ProvisionIpamPoolCidrInput{
		IpamPoolId: aws.String(poolID),
	}

	if v, ok := d.GetOk("cidr"); ok {
		input.Cidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cidr_authorization_context"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CidrAuthorizationContext = expandIPAMCIDRAuthorizationContext(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.ProvisionIpamPoolCidr(input)

	if err != nil {
		return fmt.Errorf("creating IPAM Pool (%s) CIDR: %w", poolID, err)
	}

	cidrBlock := aws.StringValue(output.IpamPoolCidr.Cidr)
	d.SetId(IPAMPoolCIDRCreateResourceID(cidrBlock, poolID))

	if _, err := WaitIPAMPoolCIDRCreated(conn, cidrBlock, poolID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for IPAM Pool CIDR (%s) create: %w", d.Id(), err)
	}

	return resourceIPAMPoolCIDRRead(d, meta)
}

func resourceIPAMPoolCIDRRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidrBlock, poolID, err := IPAMPoolCIDRParseResourceID(d.Id())

	if err != nil {
		return err
	}

	output, err := FindIPAMPoolCIDRByTwoPartKey(conn, cidrBlock, poolID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Pool CIDR (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading IPAM Pool CIDR (%s): %w", d.Id(), err)
	}

	d.Set("cidr", output.Cidr)
	d.Set("ipam_pool_id", poolID)

	return nil
}

func resourceIPAMPoolCIDRDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidrBlock, poolID, err := IPAMPoolCIDRParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting IPAM Pool CIDR: %s", d.Id())
	_, err = conn.DeprovisionIpamPoolCidr(&ec2.DeprovisionIpamPoolCidrInput{
		Cidr:       aws.String(cidrBlock),
		IpamPoolId: aws.String(poolID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
		return nil
	}

	// IncorrectState error can mean: State = "deprovisioned" || State = "pending-deprovision".
	if err != nil && !tfawserr.ErrCodeEquals(err, errCodeIncorrectState) {
		return fmt.Errorf("deleting IPAM Pool CIDR (%s): %w", d.Id(), err)
	}

	if _, err := WaitIPAMPoolCIDRDeleted(conn, cidrBlock, poolID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for IPAM Pool CIDR (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const ipamPoolCIDRIDSeparator = "_"

func IPAMPoolCIDRCreateResourceID(cidrBlock, poolID string) string {
	parts := []string{cidrBlock, poolID}
	id := strings.Join(parts, ipamPoolCIDRIDSeparator)

	return id
}

func IPAMPoolCIDRParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ipamPoolCIDRIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cidr%[2]spool-id", id, ipamPoolCIDRIDSeparator)
	}

	return parts[0], parts[1], nil
}

func expandIPAMCIDRAuthorizationContext(tfMap map[string]interface{}) *ec2.IpamCidrAuthorizationContext {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.IpamCidrAuthorizationContext{}

	if v, ok := tfMap["message"].(string); ok && v != "" {
		apiObject.Message = aws.String(v)
	}

	if v, ok := tfMap["signature"].(string); ok && v != "" {
		apiObject.Signature = aws.String(v)
	}

	return apiObject
}
