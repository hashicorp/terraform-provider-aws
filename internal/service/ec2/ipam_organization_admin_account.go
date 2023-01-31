package ec2

// ec2 has no action for Describe() to see if IPAM delegated admin has already been assigned
import ( // nosemgrep:ci.aws-sdk-go-multiple-service-imports
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMOrganizationAdminAccountCreate,
		ReadWithoutTimeout:   resourceIPAMOrganizationAdminAccountRead,
		DeleteWithoutTimeout: resourceIPAMOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	IPAMServicePrincipal = "ipam.amazonaws.com"
)

func resourceIPAMOrganizationAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	adminAccountID := d.Get("delegated_admin_account_id").(string)

	input := &ec2.EnableIpamOrganizationAdminAccountInput{
		DelegatedAdminAccountId: aws.String(adminAccountID),
	}

	output, err := conn.EnableIpamOrganizationAdminAccountWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling IPAM Organization Admin Account (%s): %s", adminAccountID, err)
	}
	if !aws.BoolValue(output.Success) {
		return sdkdiag.AppendErrorf(diags, "enabling IPAM Organization Admin Account (%s): %s", adminAccountID, err)
	}

	d.SetId(adminAccountID)

	return append(diags, resourceIPAMOrganizationAdminAccountRead(ctx, d, meta)...)
}

func resourceIPAMOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	org_conn := meta.(*conns.AWSClient).OrganizationsConn()

	input := &organizations.ListDelegatedAdministratorsInput{
		ServicePrincipal: aws.String(IPAMServicePrincipal),
	}

	output, err := org_conn.ListDelegatedAdministratorsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding IPAM organization delegated account: (%s): %s", d.Id(), err)
	}

	if output == nil || len(output.DelegatedAdministrators) == 0 || output.DelegatedAdministrators[0] == nil {
		log.Printf("[WARN] VPC Ipam Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	admin_account := output.DelegatedAdministrators[0]

	d.Set("arn", admin_account.Arn)
	d.Set("delegated_admin_account_id", admin_account.Id)
	d.Set("email", admin_account.Email)
	d.Set("name", admin_account.Name)
	d.Set("service_principal", IPAMServicePrincipal)

	return diags
}

func resourceIPAMOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.DisableIpamOrganizationAdminAccountInput{
		DelegatedAdminAccountId: aws.String(d.Id()),
	}

	output, err := conn.DisableIpamOrganizationAdminAccountWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling IPAM Organization Admin Account (%s): %s", d.Id(), err)
	}
	if !aws.BoolValue(output.Success) {
		return sdkdiag.AppendErrorf(diags, "disabling IPAM Organization Admin Account (%s): %s", d.Id(), err)
	}
	return diags
}
