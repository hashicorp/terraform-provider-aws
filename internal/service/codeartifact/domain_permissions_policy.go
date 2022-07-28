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

func ResourceDomainPermissionsPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainPermissionsPolicyPut,
		Update: resourceDomainPermissionsPolicyPut,
		Read:   resourceDomainPermissionsPolicyRead,
		Delete: resourceDomainPermissionsPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
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

func resourceDomainPermissionsPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeArtifactConn
	log.Print("[DEBUG] Creating CodeArtifact Domain Permissions Policy")

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	params := &codeartifact.PutDomainPermissionsPolicyInput{
		Domain:         aws.String(d.Get("domain").(string)),
		PolicyDocument: aws.String(policy),
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		params.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_revision"); ok {
		params.PolicyRevision = aws.String(v.(string))
	}

	res, err := conn.PutDomainPermissionsPolicy(params)
	if err != nil {
		return fmt.Errorf("error creating CodeArtifact Domain Permissions Policy: %w", err)
	}

	d.SetId(aws.StringValue(res.Policy.ResourceArn))

	return resourceDomainPermissionsPolicyRead(d, meta)
}

func resourceDomainPermissionsPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeArtifactConn
	log.Printf("[DEBUG] Reading CodeArtifact Domain Permissions Policy: %s", d.Id())

	domainOwner, domainName, err := DecodeDomainID(d.Id())
	if err != nil {
		return err
	}

	dm, err := conn.GetDomainPermissionsPolicy(&codeartifact.GetDomainPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CodeArtifact, names.ErrActionReading, ResDomainPermissionsPolicy, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CodeArtifact, names.ErrActionReading, ResDomainPermissionsPolicy, d.Id(), err)
	}

	d.Set("domain", domainName)
	d.Set("domain_owner", domainOwner)
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

func resourceDomainPermissionsPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeArtifactConn
	log.Printf("[DEBUG] Deleting CodeArtifact Domain Permissions Policy: %s", d.Id())

	domainOwner, domainName, err := DecodeDomainID(d.Id())
	if err != nil {
		return err
	}

	input := &codeartifact.DeleteDomainPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	}

	_, err = conn.DeleteDomainPermissionsPolicy(input)

	if tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CodeArtifact Domain Permissions Policy (%s): %w", d.Id(), err)
	}

	return nil
}
