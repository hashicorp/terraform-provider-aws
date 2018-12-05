package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSecurityHubInvitationAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubInvitationAccepterCreate,
		Read:   resourceAwsSecurityHubInvitationAccepterRead,
		Delete: resourceAwsSecurityHubInvitationAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"invitation_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"master_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSecurityHubInvitationAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Creating Security Hub member %s", d.Get("rest_api_id").(string))

	_, err := conn.AcceptInvitation(&securityhub.AcceptInvitationInput{
		InvitationId: aws.String(d.Get("invitation_id").(string)),
		MasterId:     aws.String(d.Get("master_id").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error accepting Security Hub invitation: %s", err)
	}

	d.SetId("securityhub-invitation-accepter")

	return resourceAwsSecurityHubInvitationAccepterRead(d, meta)
}

func resourceAwsSecurityHubInvitationAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Disassociating from Security Hub master account")

	resp, err := conn.GetMasterAccount(&securityhub.GetMasterAccountInput{})

	if err != nil {
		if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
			log.Print("[WARN] Security Hub master account not found")
			return nil
		}
		return err
	}

	master := resp.Master

	d.Set("invitation_id", master.InvitationId)
	d.Set("master_id", master.AccountId)

	return nil
}

func resourceAwsSecurityHubInvitationAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Disassociating from Security Hub master account")

	_, err := conn.DisassociateFromMasterAccount(&securityhub.DisassociateFromMasterAccountInput{})

	if err != nil {
		if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
			log.Print("[WARN] Security Hub invitation not found")
			return nil
		}
		return err
	}

	return nil
}
