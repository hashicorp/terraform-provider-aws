package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMPoolCIDR() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMPoolCIDRCreate,
		ReadWithoutTimeout:   resourceIPAMPoolCIDRRead,
		DeleteWithoutTimeout: resourceIPAMPoolCIDRDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			// Allocations release are eventually consistent with a max time of 20m.
			Delete: schema.DefaultTimeout(32 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			ResourceIPAMPoolCIDRCustomizeDiff,
		),

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
			// This resource's ID is a concatenated id of `<cidr>_<poolid>`
			// ipam_pool_cidr_id was not part of the initial feature release
			"ipam_pool_cidr_id": {
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
				ValidateFunc:  validation.IntBetween(0, 128),
				ConflictsWith: []string{"cidr"},
				// NetmaskLength is not outputted by GetIpamPoolCidrsOutput
				DiffSuppressFunc: func(k, o, n string, d *schema.ResourceData) bool {
					if o != "0" && n == "0" {
						return true
					}
					return false
				},
			},
		},
	}
}

func resourceIPAMPoolCIDRCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	poolID := d.Get("ipam_pool_id").(string)
	input := &ec2.ProvisionIpamPoolCidrInput{
		IpamPoolId: aws.String(poolID),
	}

	if v, ok := d.GetOk("cidr"); ok {
		input.Cidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("netmask_length"); ok {
		input.NetmaskLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("cidr_authorization_context"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CidrAuthorizationContext = expandIPAMCIDRAuthorizationContext(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.ProvisionIpamPoolCidrWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Pool (%s) CIDR: %s", poolID, err)
	}

	// its possible that cidr is computed based on netmask_length
	cidrBlock := aws.StringValue(output.IpamPoolCidr.Cidr)
	poolCidrId := aws.StringValue(output.IpamPoolCidr.IpamPoolCidrId)

	ipamPoolCidr, err := WaitIPAMPoolCIDRIdCreated(ctx, conn, poolCidrId, poolID, cidrBlock, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool CIDR with ID (%s) create: %s", poolCidrId, err)
	}

	// This resource's ID is a concatenated id of `<cidr>_<poolid>`
	// ipam_pool_cidr_id was not part of the initial feature release
	d.SetId(IPAMPoolCIDRCreateResourceID(aws.StringValue(ipamPoolCidr.Cidr), poolID))

	return append(diags, resourceIPAMPoolCIDRRead(ctx, d, meta)...)
}

func resourceIPAMPoolCIDRRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	cidrBlock, poolID, err := IPAMPoolCIDRParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pool CIDR (%s): %s", d.Id(), err)
	}

	output, err := FindIPAMPoolCIDRByTwoPartKey(ctx, conn, cidrBlock, poolID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Pool CIDR (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pool CIDR (%s): %s", d.Id(), err)
	}

	d.Set("cidr", output.Cidr)
	d.Set("ipam_pool_cidr_id", output.IpamPoolCidrId)
	d.Set("ipam_pool_id", poolID)

	return diags
}

func resourceIPAMPoolCIDRDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	cidrBlock, poolID, err := IPAMPoolCIDRParseResourceID(d.Id())
	poolCidrId := d.Get("ipam_pool_cidr_id").(string)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Pool CIDR (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting IPAM Pool CIDR: %s", d.Id())
	_, err = conn.DeprovisionIpamPoolCidrWithContext(ctx, &ec2.DeprovisionIpamPoolCidrInput{
		Cidr:       aws.String(cidrBlock),
		IpamPoolId: aws.String(poolID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
		return diags
	}

	// IncorrectState error can mean: State = "deprovisioned" || State = "pending-deprovision".
	if err != nil && !tfawserr.ErrCodeEquals(err, errCodeIncorrectState) {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Pool CIDR (%s): %s", d.Id(), err)
	}

	if _, err := WaitIPAMPoolCIDRDeleted(ctx, conn, cidrBlock, poolID, poolCidrId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool CIDR (%s) delete: %s", d.Id(), err)
	}

	return diags
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

func ResourceIPAMPoolCIDRCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// cidr can be set by a value returned from IPAM or explicitly in config.
	if diff.Id() != "" && diff.HasChange("cidr") {
		// If netmask is set then cidr is derived from IPAM, ignore changes.
		if diff.Get("netmask_length") != 0 {
			return diff.Clear("cidr")
		}
		return diff.ForceNew("cidr")
	}

	return nil
}
