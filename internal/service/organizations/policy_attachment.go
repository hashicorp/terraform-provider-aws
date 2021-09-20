package organizations

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	input := &organizations.AttachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	}

	log.Printf("[DEBUG] Creating Organizations Policy Attachment: %s", input)

	err := resource.Retry(4*time.Minute, func() *resource.RetryError {
		_, err := conn.AttachPolicy(input)

		if err != nil {
			if tfawserr.ErrMessageContains(err, organizations.ErrCodeFinalizingOrganizationException, "") {
				log.Printf("[DEBUG] Trying to create policy attachment again: %q", err.Error())
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AttachPolicy(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Organizations Policy Attachment: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", targetID, policyID))

	return resourcePolicyAttachmentRead(d, meta)
}

func resourcePolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	targetID, policyID, err := decodeAwsOrganizationsPolicyAttachmentID(d.Id())
	if err != nil {
		return err
	}

	input := &organizations.ListTargetsForPolicyInput{
		PolicyId: aws.String(policyID),
	}

	log.Printf("[DEBUG] Listing Organizations Policies for Target: %s", input)
	var output *organizations.PolicyTargetSummary

	err = conn.ListTargetsForPolicyPages(input, func(page *organizations.ListTargetsForPolicyOutput, lastPage bool) bool {
		for _, policySummary := range page.Targets {
			if aws.StringValue(policySummary.TargetId) == targetID {
				output = policySummary
				return true
			}
		}
		return !lastPage
	})

	if err != nil {
		if tfawserr.ErrMessageContains(err, organizations.ErrCodeTargetNotFoundException, "") {
			log.Printf("[WARN] Target does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if output == nil {
		log.Printf("[WARN] Attachment does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("policy_id", policyID)
	d.Set("target_id", targetID)
	return nil
}

func resourcePolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	targetID, policyID, err := decodeAwsOrganizationsPolicyAttachmentID(d.Id())
	if err != nil {
		return err
	}

	input := &organizations.DetachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	}

	log.Printf("[DEBUG] Detaching Organizations Policy %q from %q", policyID, targetID)
	_, err = conn.DetachPolicy(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, organizations.ErrCodePolicyNotFoundException, "") {
			return nil
		}
		if tfawserr.ErrMessageContains(err, organizations.ErrCodeTargetNotFoundException, "") {
			return nil
		}
		return err
	}
	return nil
}

func decodeAwsOrganizationsPolicyAttachmentID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format of TARGETID:POLICYID, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
