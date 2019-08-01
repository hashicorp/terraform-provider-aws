package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsFmsAdminAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFmsAdminAccountPut,
		Read:   resourceAwsFmsAdminAccountRead,
		Delete: resourceAwsFmsAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
		},
	}
}

func resourceAwsFmsAdminAccountPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	accountId := meta.(*AWSClient).accountid
	if v, ok := d.GetOk("account_id"); ok && v != "" {
		accountId = v.(string)
	}

	stateConf := &resource.StateChangeConf{
		Target:     []string{accountId},
		Refresh:    associateAdminAccountRefreshFunc(conn, accountId),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for firewall manager admin account association: %v", accountId)
	_, sterr := stateConf.WaitForState()
	if sterr != nil {
		return fmt.Errorf("Error waiting for firewall manager admin account association (%s): %s", accountId, sterr)
	}

	d.SetId(accountId)
	return nil
}

func associateAdminAccountRefreshFunc(conn *fms.FMS, accountId string) resource.StateRefreshFunc {
	// This is all wrapped in a refresh func since AssociateAdminAccount returns
	// success even though it failed if called too quickly after creating an organization
	return func() (interface{}, string, error) {
		req := &fms.AssociateAdminAccountInput{
			AdminAccount: aws.String(accountId),
		}

		_, aserr := conn.AssociateAdminAccount(req)
		if aserr != nil {
			return nil, "", aserr
		}

		res, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})
		if err != nil {
			// FMS returns an AccessDeniedException if no account is associated,
			// but does not define this in its error codes
			if isAWSErr(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
				return nil, "", nil
			}
			return nil, "", err
		}
		return *res, *res.AdminAccount, err
	}
}

func resourceAwsFmsAdminAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	res, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})
	if err != nil {
		// FMS returns an AccessDeniedException if no account is associated,
		// but does not define this in its error codes
		if isAWSErr(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
			log.Printf("[WARN] No associated firewall manager admin account found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if d.Id() != aws.StringValue(res.AdminAccount) {
		log.Printf("[WARN] FMS Admin Account does not match, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("account_id", res.AdminAccount)
	return nil
}

func resourceAwsFmsAdminAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	_, err := conn.DisassociateAdminAccount(&fms.DisassociateAdminAccountInput{})
	if err != nil {
		// FMS returns an AccessDeniedException if no account is associated,
		// but does not define this in its error codes
		if isAWSErr(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
			log.Printf("[WARN] No associated firewall manager admin account found, removing from state: %s", d.Id())
			return nil
		}
		return fmt.Errorf("Error disassociating firewall manager admin account: %s", err)
	}

	return nil
}
