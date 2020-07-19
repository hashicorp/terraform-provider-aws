package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsCodeArtifactDomainPermissionsPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeArtifactDomainPermissionsPolicyPut,
		Update: resourceAwsCodeArtifactDomainPermissionsPolicyPut,
		Read:   resourceAwsCodeArtifactDomainPermissionsPolicyRead,
		Delete: resourceAwsCodeArtifactDomainPermissionsPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policy_document": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
			"policy_revision": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodeArtifactDomainPermissionsPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Print("[DEBUG] Creating CodeArtifact Domain Permissions Policy")

	domain := d.Get("domain").(string)

	params := &codeartifact.PutDomainPermissionsPolicyInput{
		Domain:         aws.String(domain),
		PolicyDocument: aws.String(d.Get("policy_document").(string)),
	}

	if v, ok := d.GetOk("policy_revision"); ok {
		params.PolicyRevision = aws.String(v.(string))
	}

	_, err := conn.PutDomainPermissionsPolicy(params)
	if err != nil {
		return fmt.Errorf("error creating CodeArtifact Domain Permissions Policy: %s", err)
	}

	d.SetId(domain)

	return resourceAwsCodeArtifactDomainPermissionsPolicyRead(d, meta)
}

func resourceAwsCodeArtifactDomainPermissionsPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn

	log.Printf("[DEBUG] Reading CodeArtifact Domain Permissions Policy: %s", d.Id())

	dm, err := conn.GetDomainPermissionsPolicy(&codeartifact.GetDomainPermissionsPolicyInput{
		Domain: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] CodeArtifact Domain Permissions Policy %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading CodeArtifact Domain Permissions Policy (%s): %w", d.Id(), err)
	}

	d.Set("domain", d.Id())
	d.Set("resource_arn", dm.Policy.ResourceArn)
	d.Set("policy_document", dm.Policy.Document)
	d.Set("policy_revision", dm.Policy.Revision)

	return nil
}

func resourceAwsCodeArtifactDomainPermissionsPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Printf("[DEBUG] Deleting CodeArtifact Domain Permissions Policy: %s", d.Id())

	input := &codeartifact.DeleteDomainPermissionsPolicyInput{
		Domain: aws.String(d.Id()),
	}

	_, err := conn.DeleteDomainPermissionsPolicy(input)

	if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CodeArtifact Domain Permissions Policy (%s): %w", d.Id(), err)
	}

	return nil
}
