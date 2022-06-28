package codeartifact

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceRepositoryPermissionsPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceRepositoryPermissionsPolicyPut,
		Update: resourceRepositoryPermissionsPolicyPut,
		Read:   resourceRepositoryPermissionsPolicyRead,
		Delete: resourceRepositoryPermissionsPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"policy_document": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
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

func resourceRepositoryPermissionsPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeArtifactConn
	log.Print("[DEBUG] Creating CodeArtifact Repository Permissions Policy")

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	params := &codeartifact.PutRepositoryPermissionsPolicyInput{
		Domain:         aws.String(d.Get("domain").(string)),
		Repository:     aws.String(d.Get("repository").(string)),
		PolicyDocument: aws.String(policy),
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		params.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_revision"); ok {
		params.PolicyRevision = aws.String(v.(string))
	}

	res, err := conn.PutRepositoryPermissionsPolicy(params)
	if err != nil {
		return fmt.Errorf("error creating CodeArtifact Repository Permissions Policy: %w", err)
	}

	d.SetId(aws.StringValue(res.Policy.ResourceArn))

	return resourceRepositoryPermissionsPolicyRead(d, meta)
}

func resourceRepositoryPermissionsPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeArtifactConn
	log.Printf("[DEBUG] Reading CodeArtifact Repository Permissions Policy: %s", d.Id())

	domainOwner, domainName, repoName, err := DecodeRepositoryID(d.Id())
	if err != nil {
		return err
	}

	dm, err := conn.GetRepositoryPermissionsPolicy(&codeartifact.GetRepositoryPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
		Repository:  aws.String(repoName),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CodeArtifact, names.ErrActionReading, ResRepositoryPermissionsPolicy, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CodeArtifact, names.ErrActionReading, ResRepositoryPermissionsPolicy, d.Id(), err)
	}

	d.Set("domain", domainName)
	d.Set("domain_owner", domainOwner)
	d.Set("repository", repoName)
	d.Set("resource_arn", dm.Policy.ResourceArn)
	d.Set("policy_revision", dm.Policy.Revision)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy_document").(string), aws.StringValue(dm.Policy.Document))

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	d.Set("policy_document", policyToSet)

	return nil
}

func resourceRepositoryPermissionsPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeArtifactConn
	log.Printf("[DEBUG] Deleting CodeArtifact Repository Permissions Policy: %s", d.Id())

	domainOwner, domainName, repoName, err := DecodeRepositoryID(d.Id())
	if err != nil {
		return err
	}

	input := &codeartifact.DeleteRepositoryPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
		Repository:  aws.String(repoName),
	}

	_, err = conn.DeleteRepositoryPermissionsPolicy(input)

	if tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CodeArtifact Repository Permissions Policy (%s): %w", d.Id(), err)
	}

	return nil
}
