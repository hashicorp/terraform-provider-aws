package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ram"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsRamResourceShareAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRamResourceShareAccepterCreate,
		Read:   resourceAwsRamResourceShareAccepterRead,
		Delete: resourceAwsRamResourceShareAccepterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"invitation_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},

			"share_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},

			"share_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"receiver_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"sender_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsRamResourceShareAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	in := &ram.AcceptResourceShareInvitationInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("invitation_arn"); ok {
		in.ResourceShareInvitationArn = aws.String(v.(string))
	} else if v, ok := d.GetOk("share_arn"); ok {
		// need to find invitation arn
		invitationARN, err := resourceAwsRamResourceShareGetInvitationARN(conn, v.(string))
		if err != nil {
			return err
		}

		in.ResourceShareInvitationArn = aws.String(invitationARN)
	} else {
		return fmt.Errorf("Either an invitation ARN or share ARN are required")
	}

	log.Printf("[DEBUG] Accept RAM resource share invitation request: %s", in)
	out, err := conn.AcceptResourceShareInvitation(in)
	if err != nil {
		return fmt.Errorf("Error accepting RAM resource share invitation: %s", err)
	}

	d.SetId(aws.StringValue(out.ResourceShareInvitation.ResourceShareInvitationArn))

	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareInvitationStatusPending},
		Target:  []string{ram.ResourceShareInvitationStatusAccepted},
		Refresh: resourceAwsRamResourceShareAccepterStateRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for RAM resource share (%s) state: %s", d.Id(), err)
	}

	return resourceAwsRamResourceShareAccepterRead(d, meta)
}

func resourceAwsRamResourceShareAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	request := &ram.GetResourceShareInvitationsInput{
		ResourceShareInvitationArns: []*string{aws.String(d.Id())},
	}

	out, err := conn.GetResourceShareInvitations(request)
	if err != nil {
		if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
			log.Printf("[WARN] No RAM resource share invitation by ARN (%s) found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading RAM resource share invitation %s: %s", d.Id(), err)
	}

	if len(out.ResourceShareInvitations) == 0 {
		log.Printf("[WARN] No RAM resource share invitation by ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	invitation := out.ResourceShareInvitations[0]

	d.Set("status", invitation.Status)
	d.Set("receiver_account_id", invitation.ReceiverAccountId)
	d.Set("sender_account_id", invitation.SenderAccountId)
	d.Set("share_arn", invitation.ResourceShareArn)
	d.Set("invitation_arn", invitation.ResourceShareInvitationArn)
	d.Set("share_id", resourceAwsRamResourceShareGetIDFromARN(aws.StringValue(invitation.ResourceShareArn)))
	d.Set("share_name", invitation.ResourceShareName)

	var nextToken string
	var resourceARNs []*string
	for {
		listInput := &ram.ListResourcesInput{
			MaxResults:        aws.Int64(int64(500)),
			ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
			ResourceShareArns: aws.StringSlice([]string{aws.StringValue(invitation.ResourceShareArn)}),
			Principal:         invitation.SenderAccountId,
		}

		if nextToken != "" {
			listInput.NextToken = aws.String(nextToken)
		}
		out, err := conn.ListResources(listInput)
		if err != nil {
			return fmt.Errorf("could not list share resources: %s", err)
		}
		for _, resource := range out.Resources {
			resourceARNs = append(resourceARNs, resource.Arn)
		}

		if out.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(out.NextToken)
	}

	if err := d.Set("resources", flattenStringList(resourceARNs)); err != nil {
		return fmt.Errorf("unable to set resources: %s", err)
	}

	return nil
}

/*
func resourceAwsRamResourceShareAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	log.Printf("[WARN] Will not delete resource share invitation. Terraform will remove this invitation accepter from the state file. However, resources may remain.")
	return nil
}
*/

func resourceAwsRamResourceShareAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	v, ok := d.GetOk("share_arn")
	if !ok {
		return fmt.Errorf("The share ARN is required to leave a resource share")
	}
	shareARN := v.(string)

	v, ok = d.GetOk("receiver_account_id")
	if !ok {
		return fmt.Errorf("The receiver account ID is required to leave a resource share")
	}
	receiverID := v.(string)

	in := &ram.DisassociateResourceShareInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ResourceShareArn: aws.String(shareARN),
		Principals:       []*string{aws.String(receiverID)},
	}
	log.Printf("[DEBUG] Leaving RAM resource share request: %s", in)
	_, err := conn.DisassociateResourceShare(in)
	if err != nil {
		return fmt.Errorf("Error leaving RAM resource share: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated},
		Refresh: resourceAwsRamResourceShareStateRefreshFunc(conn, shareARN),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	if _, err := stateConf.WaitForState(); err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			switch awserr.Code() {
			case ram.ErrCodeUnknownResourceException:
				// what we want
				d.SetId("")
				return nil
			}
		}
		return fmt.Errorf("Error waiting for RAM resource share (%s) state: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRamResourceShareGetInvitationARN(conn *ram.RAM, resourceShareARN string) (string, error) {
	var nextToken string
	for {
		input := &ram.GetResourceShareInvitationsInput{
			MaxResults:        aws.Int64(int64(500)),
			ResourceShareArns: aws.StringSlice([]string{resourceShareARN}),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}
		out, err := conn.GetResourceShareInvitations(input)
		if err != nil {
			return "", err
		}
		for _, invitation := range out.ResourceShareInvitations {
			if aws.StringValue(invitation.Status) == ram.ResourceShareInvitationStatusPending {
				return aws.StringValue(invitation.ResourceShareInvitationArn), nil
			}
		}

		if out.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(out.NextToken)
	}

	return "", fmt.Errorf("Unable to find a pending invitation for resource share %s", resourceShareARN)
}

func resourceAwsRamResourceShareAccepterStateRefreshFunc(conn *ram.RAM, invitationArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		request := &ram.GetResourceShareInvitationsInput{
			ResourceShareInvitationArns: []*string{aws.String(invitationArn)},
		}

		out, err := conn.GetResourceShareInvitations(request)

		if err != nil {
			return nil, "Unable to get resource share invitations", err
		}

		if len(out.ResourceShareInvitations) == 0 {
			return nil, "Resource share invitation not found", nil
		}

		invitation := out.ResourceShareInvitations[0]

		return invitation, aws.StringValue(invitation.Status), nil
	}
}

func resourceAwsRamResourceShareGetIDFromARN(arn string) string {
	return strings.Replace(arn[strings.LastIndex(arn, ":")+1:], "resource-share/", "rs-", -1)
}
