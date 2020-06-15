package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsCodeArtifactDomainPermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeArtifactDomainPermissionsPut,
		Update: resourceAwsCodeArtifactDomainPermissionsPut,
		Read:   resourceAwsCodeArtifactDomainPermissionsRead,
		Delete: resourceAwsCodeArtifactDomainPermissionsDelete,
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

func resourceAwsCodeArtifactDomainPermissionsPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Print("[DEBUG] Creating CodeArtifact Domain Permissions")

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
		return fmt.Errorf("error creating CodeArtifact Domain Permissions: %s", err)
	}

	d.SetId(domain)

	return resourceAwsCodeArtifactDomainPermissionsRead(d, meta)
}

func resourceAwsCodeArtifactDomainPermissionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn

	log.Printf("[DEBUG] Reading CodeArtifact Domain Permissions: %s", d.Id())

	dm, err := conn.GetDomainPermissionsPolicy(&codeartifact.GetDomainPermissionsPolicyInput{
		Domain: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] CodeArtifact Domain Permissions %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("domain", d.Id())
	d.Set("resource_arn", dm.Policy.ResourceArn)
	d.Set("policy_document", dm.Policy.Document)
	d.Set("policy_revision", dm.Policy.Revision)

	return nil
}

func resourceAwsCodeArtifactDomainPermissionsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Printf("[DEBUG] Deleting CodeArtifact Domain Permissions: %s", d.Id())

	input := &codeartifact.DeleteDomainPermissionsPolicyInput{
		Domain: aws.String(d.Id()),
	}

	_, err := conn.DeleteDomainPermissionsPolicy(input)

	if err != nil {
		return fmt.Errorf("error deleting CodeArtifact Domain Permissions: %s", err)
	}

	return nil
}
