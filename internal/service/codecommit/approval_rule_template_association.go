package codecommit

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceApprovalRuleTemplateAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceApprovalRuleTemplateAssociationCreate,
		Read:   resourceApprovalRuleTemplateAssociationRead,
		Delete: resourceApprovalRuleTemplateAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"approval_rule_template_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`[\w\.-]+`), ""),
				),
			},
		},
	}
}

func resourceApprovalRuleTemplateAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeCommitConn

	approvalRuleTemplateName := d.Get("approval_rule_template_name").(string)
	repositoryName := d.Get("repository_name").(string)

	input := &codecommit.AssociateApprovalRuleTemplateWithRepositoryInput{
		ApprovalRuleTemplateName: aws.String(approvalRuleTemplateName),
		RepositoryName:           aws.String(repositoryName),
	}

	_, err := conn.AssociateApprovalRuleTemplateWithRepository(input)

	if err != nil {
		return fmt.Errorf("error associating CodeCommit Approval Rule Template (%s) with repository (%s): %w", approvalRuleTemplateName, repositoryName, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", approvalRuleTemplateName, repositoryName))

	return resourceApprovalRuleTemplateAssociationRead(d, meta)
}

func resourceApprovalRuleTemplateAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeCommitConn

	approvalRuleTemplateName, repositoryName, err := ApprovalRuleTemplateAssociationParseID(d.Id())

	if err != nil {
		return err
	}

	err = FindApprovalRuleTemplateAssociation(conn, approvalRuleTemplateName, repositoryName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeCommit Approval Rule Template Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CodeCommit Approval Rule Template Association (%s): %w", d.Id(), err)
	}

	d.Set("approval_rule_template_name", approvalRuleTemplateName)
	d.Set("repository_name", repositoryName)

	return nil
}

func resourceApprovalRuleTemplateAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeCommitConn

	approvalRuleTemplateName, repositoryName, err := ApprovalRuleTemplateAssociationParseID(d.Id())

	if err != nil {
		return err
	}

	input := &codecommit.DisassociateApprovalRuleTemplateFromRepositoryInput{
		ApprovalRuleTemplateName: aws.String(approvalRuleTemplateName),
		RepositoryName:           aws.String(repositoryName),
	}

	_, err = conn.DisassociateApprovalRuleTemplateFromRepository(input)

	if tfawserr.ErrCodeEquals(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException, codecommit.ErrCodeRepositoryDoesNotExistException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating CodeCommit Approval Rule Template (%s) from repository (%s): %w", approvalRuleTemplateName, repositoryName, err)
	}

	return nil
}

func ApprovalRuleTemplateAssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected APPROVAL_RULE_TEMPLATE_NAME,REPOSITORY_NAME", id)
	}

	return parts[0], parts[1], nil
}
