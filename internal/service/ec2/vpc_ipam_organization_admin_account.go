package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCIpamOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCIpamOrganizationAdminAccountCreate,
		Read:   resourceVPCIpamOrganizationAdminAccountRead,
		// TODO: validate update is possible
		// Update: resourceVPCIpamOrganizationAdminAccountUpdate,
		Delete: resourceVPCIpamOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"delegated_admin_account_id": {
				Type:     schema.TypeString,
				Required: true,
				// ForceNew = true if cannot update in place - see L21
				// ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

const (
	ipam_service_principal = "ipam.amazonaws.com"
)

func resourceVPCIpamOrganizationAdminAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	adminAccountID := d.Get("delegated_admin_account_id").(string)

	input := &ec2.EnableIpamOrganizationAdminAccountInput{
		DelegatedAdminAccountId: aws.String(adminAccountID),
	}

	output, err := conn.EnableIpamOrganizationAdminAccount(input)

	if err != nil {
		return fmt.Errorf("error enabling IPAM Organization Admin Account (%s): %w", adminAccountID, err)
	}
	if !aws.BoolValue(output.Success) {
		return fmt.Errorf("error enabling IPAM Organization Admin Account (%s): %w", adminAccountID, err)
	}

	d.SetId(encodeIpamOrgAdminId(adminAccountID))

	return resourceVPCIpamOrganizationAdminAccountRead(d, meta)
}

func resourceVPCIpamOrganizationAdminAccountRead(d *schema.ResourceData, meta interface{}) error {
	org_conn := meta.(*conns.AWSClient).OrganizationsConn
	// ListDelegatedAdministratorsInput

	// if err != nil {
	// 	return fmt.Errorf("error reading VPCIpam Organization Admin Account (%s): %w", d.Id(), err)
	// }

	// if adminAccount == nil {
	// 	log.Printf("[WARN] VPCIpam Organization Admin Account (%s) not found, removing from state", d.Id())
	// 	d.SetId("")
	// 	return nil
	// }

	return nil
}

func resourceVPCIpamOrganizationAdminAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	// need to check if its possbile to overwrite
	// likely youll just run the same steps from Create()
	return nil
}

func resourceVPCIpamOrganizationAdminAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	account_id, _, err := DecodeIpamPoolCidrID(d.Id())

	input := &ec2.DisableIpamOrganizationAdminAccountInput{
		DelegatedAdminAccountId: aws.String(account_id),
	}

	output, err := conn.DisableIpamOrganizationAdminAccount(input)

	if err != nil {
		return fmt.Errorf("error disabling IPAM Organization Admin Account (%s): %w", account_id, err)
	}
	if !aws.BoolValue(output.Success) {
		return fmt.Errorf("error disabling IPAM Organization Admin Account (%s): %w", account_id, err)
	}
	return nil
}

func encodeIpamOrgAdminId(account_id string) string {
	return fmt.Sprintf("%s_%s", account_id, ipam_service_principal)
}

func DecodeIpamOrgAdminId(id string) (string, string, error) {
	idParts := strings.Split(id, "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form of <account_id>_<service_principal>, given: %q", id)
	}
	return idParts[0], idParts[1], nil
}
