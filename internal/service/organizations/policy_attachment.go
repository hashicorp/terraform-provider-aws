package organizations

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyAttachmentCreate,
		Read:   resourcePolicyAttachmentRead,
		Delete: resourcePolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	policyID := d.Get("policy_id").(string)
	targetID := d.Get("target_id").(string)
	id := fmt.Sprintf("%s:%s", targetID, policyID)
	input := &organizations.AttachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(4*time.Minute, func() (interface{}, error) {
		return conn.AttachPolicy(input)
	}, organizations.ErrCodeFinalizingOrganizationException)

	if err != nil {
		return fmt.Errorf("creating Organizations Policy Attachment (%s): %w", id, err)
	}

	d.SetId(id)

	return resourcePolicyAttachmentRead(d, meta)
}

func resourcePolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	targetID, policyID, err := DecodePolicyAttachmentID(d.Id())
	if err != nil {
		return err
	}

	_, err = FindPolicyAttachmentByTwoPartKey(conn, targetID, policyID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Policy Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Organizations Policy Attachment (%s): %w", d.Id(), err)
	}

	d.Set("policy_id", policyID)
	d.Set("target_id", targetID)

	return nil
}

func resourcePolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	targetID, policyID, err := DecodePolicyAttachmentID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting Organizations Policy Attachment: %s", d.Id())
	_, err = conn.DetachPolicy(&organizations.DetachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeTargetNotFoundException, organizations.ErrCodePolicyNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Organizations Policy Attachment (%s): %w", d.Id(), err)
	}

	return nil
}

func DecodePolicyAttachmentID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format of TARGETID:POLICYID, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
