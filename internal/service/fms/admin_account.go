package fms

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAdminAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdminAccountCreate,
		Read:   resourceAdminAccountRead,
		Delete: resourceAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceAdminAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FMSConn

	// Ensure there is not an existing FMS Admin Account
	output, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})

	if err != nil && !tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		return fmt.Errorf("error getting FMS Admin Account: %w", err)
	}

	if output != nil && output.AdminAccount != nil && aws.StringValue(output.RoleStatus) == fms.AccountRoleStatusReady {
		return fmt.Errorf("FMS Admin Account (%s) already associated: import this Terraform resource to manage", aws.StringValue(output.AdminAccount))
	}

	accountID := meta.(*conns.AWSClient).AccountID

	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			fms.AccountRoleStatusDeleted, // Recreating association can return this status
			fms.AccountRoleStatusCreating,
		},
		Target:  []string{fms.AccountRoleStatusReady},
		Refresh: associateAdminAccountRefreshFunc(conn, accountID),
		Timeout: 30 * time.Minute,
		Delay:   10 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for FMS Admin Account (%s) association: %w", accountID, err)
	}

	d.SetId(accountID)

	return resourceAdminAccountRead(d, meta)
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
			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
				return nil, "", nil
			}
			if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
				return nil, "", nil
			}
			return nil, "", err
		}

		if aws.StringValue(res.AdminAccount) != accountId {
			return nil, "", nil
		}

		return res, aws.StringValue(res.RoleStatus), err
	}
}

func resourceAdminAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FMSConn

	output, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] FMS Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting FMS Admin Account (%s): %w", d.Id(), err)
	}

	if aws.StringValue(output.RoleStatus) == fms.AccountRoleStatusDeleted {
		if d.IsNewResource() {
			return fmt.Errorf("error getting FMS Admin Account (%s): %s after creation", d.Id(), aws.StringValue(output.RoleStatus))
		}

		log.Printf("[WARN] FMS Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("account_id", output.AdminAccount)

	return nil
}

func resourceAdminAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FMSConn

	_, err := conn.DisassociateAdminAccount(&fms.DisassociateAdminAccountInput{})

	if err != nil {
		return fmt.Errorf("error disassociating FMS Admin Account (%s): %w", d.Id(), err)
	}

	if err := waitForAdminAccountDeletion(conn); err != nil {
		return fmt.Errorf("error waiting for FMS Admin Account (%s) disassociation: %w", d.Id(), err)
	}

	return nil
}

func waitForAdminAccountDeletion(conn *fms.FMS) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			fms.AccountRoleStatusDeleting,
			fms.AccountRoleStatusPendingDeletion,
			fms.AccountRoleStatusReady,
		},
		Target: []string{fms.AccountRoleStatusDeleted},
		Refresh: func() (interface{}, string, error) {
			output, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})

			if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
				return output, fms.AccountRoleStatusDeleted, nil
			}

			if err != nil {
				return nil, "", err
			}

			return output, aws.StringValue(output.RoleStatus), nil
		},
		Timeout: 10 * time.Minute,
		Delay:   10 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
