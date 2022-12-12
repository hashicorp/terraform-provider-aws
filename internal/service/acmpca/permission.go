package acmpca

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionCreate,
		Read:   resourcePermissionRead,
		Delete: resourcePermissionDelete,

		Schema: map[string]*schema.Schema{
			"actions": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(acmpca.ActionType_Values(), false),
				},
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"principal": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"acm.amazonaws.com",
				}, false),
			},
			"source_account": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourcePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	caARN := d.Get("certificate_authority_arn").(string)
	principal := d.Get("principal").(string)
	sourceAccount := d.Get("source_account").(string)
	id := PermissionCreateResourceID(caARN, principal, sourceAccount)
	input := &acmpca.CreatePermissionInput{
		Actions:                 flex.ExpandStringSet(d.Get("actions").(*schema.Set)),
		CertificateAuthorityArn: aws.String(caARN),
		Principal:               aws.String(principal),
	}

	if sourceAccount != "" {
		input.SourceAccount = aws.String(sourceAccount)
	}

	log.Printf("[DEBUG] Creating ACM PCA Permission: %s", input)
	_, err := conn.CreatePermission(input)

	if err != nil {
		return fmt.Errorf("creating ACM PCA Permission (%s): %w", id, err)
	}

	d.SetId(id)

	return resourcePermissionRead(d, meta)
}

func resourcePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	caARN, principal, sourceAccount, err := PermissionParseResourceID(d.Id())

	if err != nil {
		return err
	}

	permission, err := FindPermission(conn, caARN, principal, sourceAccount)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading ACM PCA Permission (%s): %w", d.Id(), err)
	}

	d.Set("actions", aws.StringValueSlice(permission.Actions))
	d.Set("certificate_authority_arn", permission.CertificateAuthorityArn)
	d.Set("policy", permission.Policy)
	d.Set("principal", permission.Principal)
	d.Set("source_account", permission.SourceAccount)

	return nil
}

func resourcePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	caARN, principal, sourceAccount, err := PermissionParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &acmpca.DeletePermissionInput{
		CertificateAuthorityArn: aws.String(caARN),
		Principal:               aws.String(principal),
	}

	if sourceAccount != "" {
		input.SourceAccount = aws.String(sourceAccount)
	}

	log.Printf("[DEBUG] Deleting ACM PCA Permission: %s", d.Id())
	_, err = conn.DeletePermission(input)

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting ACM PCA Permission: %s", err)
	}

	return nil
}

const permissionIDSeparator = ","

func PermissionCreateResourceID(caARN, principal, sourceAccount string) string {
	parts := []string{caARN, principal, sourceAccount}
	id := strings.Join(parts, permissionIDSeparator)

	return id
}

func PermissionParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, permissionIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected CertificateAuthorityARN%[2]sPrincipal%[2]sSourceAccount", id, permissionIDSeparator)
}
