package codecommit

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceApprovalRuleTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceApprovalRuleTemplateCreate,
		Read:   resourceApprovalRuleTemplateRead,
		Update: resourceApprovalRuleTemplateUpdate,
		Delete: resourceApprovalRuleTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"content": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				ValidateFunc: validation.All(
					validation.StringIsJSON,
					validation.StringLenBetween(1, 3000),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"approval_rule_template_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_user": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_content_sha256": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceApprovalRuleTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeCommitConn

	name := d.Get("name").(string)

	input := &codecommit.CreateApprovalRuleTemplateInput{
		ApprovalRuleTemplateName:    aws.String(name),
		ApprovalRuleTemplateContent: aws.String(d.Get("content").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.ApprovalRuleTemplateDescription = aws.String(v.(string))
	}

	_, err := conn.CreateApprovalRuleTemplate(input)
	if err != nil {
		return fmt.Errorf("error creating CodeCommit Approval Rule Template (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceApprovalRuleTemplateRead(d, meta)
}

func resourceApprovalRuleTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeCommitConn

	input := &codecommit.GetApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(d.Id()),
	}

	resp, err := conn.GetApprovalRuleTemplate(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException) {
		log.Printf("[WARN] CodeCommit Approval Rule Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CodeCommit Approval Rule Template (%s): %w", d.Id(), err)
	}

	if resp == nil || resp.ApprovalRuleTemplate == nil {
		return fmt.Errorf("error reading CodeCommit Approval Rule Template (%s): empty output", d.Id())
	}

	result := resp.ApprovalRuleTemplate

	d.Set("name", result.ApprovalRuleTemplateName)
	d.Set("approval_rule_template_id", result.ApprovalRuleTemplateId)
	d.Set("description", result.ApprovalRuleTemplateDescription)
	d.Set("content", result.ApprovalRuleTemplateContent)
	d.Set("creation_date", result.CreationDate.Format(time.RFC3339))
	d.Set("last_modified_date", result.LastModifiedDate.Format(time.RFC3339))
	d.Set("last_modified_user", result.LastModifiedUser)
	d.Set("rule_content_sha256", result.RuleContentSha256)

	return nil
}

func resourceApprovalRuleTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeCommitConn

	if d.HasChange("description") {
		input := &codecommit.UpdateApprovalRuleTemplateDescriptionInput{
			ApprovalRuleTemplateDescription: aws.String(d.Get("description").(string)),
			ApprovalRuleTemplateName:        aws.String(d.Id()),
		}

		_, err := conn.UpdateApprovalRuleTemplateDescription(input)

		if err != nil {
			return fmt.Errorf("error updating CodeCommit Approval Rule Template (%s) description: %w", d.Id(), err)
		}
	}

	if d.HasChange("content") {
		input := &codecommit.UpdateApprovalRuleTemplateContentInput{
			ApprovalRuleTemplateName:  aws.String(d.Id()),
			ExistingRuleContentSha256: aws.String(d.Get("rule_content_sha256").(string)),
			NewRuleContent:            aws.String(d.Get("content").(string)),
		}

		_, err := conn.UpdateApprovalRuleTemplateContent(input)

		if err != nil {
			return fmt.Errorf("error updating CodeCommit Approval Rule Template (%s) content: %w", d.Id(), err)
		}
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)

		input := &codecommit.UpdateApprovalRuleTemplateNameInput{
			NewApprovalRuleTemplateName: aws.String(newName),
			OldApprovalRuleTemplateName: aws.String(d.Id()),
		}

		_, err := conn.UpdateApprovalRuleTemplateName(input)

		if err != nil {
			return fmt.Errorf("error updating CodeCommit Approval Rule Template (%s) name: %w", d.Id(), err)
		}

		d.SetId(newName)
	}

	return resourceApprovalRuleTemplateRead(d, meta)
}

func resourceApprovalRuleTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeCommitConn

	input := &codecommit.DeleteApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(d.Id()),
	}

	_, err := conn.DeleteApprovalRuleTemplate(input)

	if tfawserr.ErrCodeEquals(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CodeCommit Approval Rule Template (%s): %w", d.Id(), err)
	}

	return nil
}
