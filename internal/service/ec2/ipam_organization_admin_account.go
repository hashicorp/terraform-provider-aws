package ec2

// ec2 has no action for Describe() to see if IPAM delegated admin has already been assigned
import ( // nosemgrep: aws-sdk-go-multiple-service-imports
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCIpamOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCIpamOrganizationAdminAccountCreate,
		Read:   resourceVPCIpamOrganizationAdminAccountRead,
		Delete: resourceVPCIpamOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delegated_admin_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_principal": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	Ipam_service_principal = "ipam.amazonaws.com"
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

	d.SetId(adminAccountID)

	return resourceVPCIpamOrganizationAdminAccountRead(d, meta)
}

func resourceVPCIpamOrganizationAdminAccountRead(d *schema.ResourceData, meta interface{}) error {
	org_conn := meta.(*conns.AWSClient).OrganizationsConn

	input := &organizations.ListDelegatedAdministratorsInput{
		ServicePrincipal: aws.String(Ipam_service_principal),
	}

	output, err := org_conn.ListDelegatedAdministrators(input)

	if err != nil {
		return fmt.Errorf("error finding IPAM organization delegated account: (%s): %w", d.Id(), err)
	}

	if output == nil || len(output.DelegatedAdministrators) == 0 || output.DelegatedAdministrators[0] == nil {
		log.Printf("[WARN] VPC Ipam Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	admin_account := output.DelegatedAdministrators[0]

	d.Set("arn", admin_account.Arn)
	d.Set("delegated_admin_account_id", admin_account.Id)
	d.Set("email", admin_account.Email)
	d.Set("name", admin_account.Name)
	d.Set("service_principal", Ipam_service_principal)

	return nil
}

func resourceVPCIpamOrganizationAdminAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DisableIpamOrganizationAdminAccountInput{
		DelegatedAdminAccountId: aws.String(d.Id()),
	}

	output, err := conn.DisableIpamOrganizationAdminAccount(input)

	if err != nil {
		return fmt.Errorf("error disabling IPAM Organization Admin Account (%s): %w", d.Id(), err)
	}
	if !aws.BoolValue(output.Success) {
		return fmt.Errorf("error disabling IPAM Organization Admin Account (%s): %w", d.Id(), err)
	}
	return nil
}
