package acmpca

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionCreate,
		Read:   resourcePermissionRead,
		Delete: resourcePermissionDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},

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

	ca_arn := d.Get("certificate_authority_arn").(string)
	principal := d.Get("principal").(string)

	input := &acmpca.CreatePermissionInput{
		Actions:                 flex.ExpandStringSet(d.Get("actions").(*schema.Set)),
		CertificateAuthorityArn: aws.String(ca_arn),
		Principal:               aws.String(principal),
	}

	source_account := d.Get("source_account").(string)
	if source_account != "" {
		input.SetSourceAccount(source_account)
	}

	log.Printf("[DEBUG] Creating ACMPCA Permission: %s", input)

	_, err := conn.CreatePermission(input)

	if err != nil {
		return fmt.Errorf("error creating ACMPCA Permission: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s", ca_arn, principal))

	return resourcePermissionRead(d, meta)
}

func describePermissions(conn *acmpca.ACMPCA, certificateAuthorityArn string, principal string, sourceAccount string) (*acmpca.Permission, error) {

	out, err := conn.ListPermissions(&acmpca.ListPermissionsInput{
		CertificateAuthorityArn: &certificateAuthorityArn,
	})

	if err != nil {
		log.Printf("[WARN] Error retrieving ACMPCA Permissions (%s) when waiting: %s", certificateAuthorityArn, err)
		return nil, err
	}

	var permission *acmpca.Permission

	for _, p := range out.Permissions {
		if aws.StringValue(p.CertificateAuthorityArn) == certificateAuthorityArn && aws.StringValue(p.Principal) == principal && (sourceAccount == "" || aws.StringValue(p.SourceAccount) == sourceAccount) {
			permission = p
			break
		}
	}
	return permission, nil
}

func resourcePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	permission, err := describePermissions(conn, d.Get("certificate_authority_arn").(string), d.Get("principal").(string), d.Get("source_account").(string))

	if permission == nil {
		log.Printf("[WARN] ACMPCA Permission (%s) not found", d.Get("certificate_authority_arn"))
		d.SetId("")
		return err
	}

	d.Set("source_account", permission.SourceAccount)
	d.Set("policy", permission.Policy)

	return nil
}

func resourcePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	input := &acmpca.DeletePermissionInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		Principal:               aws.String(d.Get("principal").(string)),
		SourceAccount:           aws.String(d.Get("source_account").(string)),
	}

	log.Printf("[DEBUG] Deleting ACMPCA Permission: %s", input)
	_, err := conn.DeletePermission(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting ACMPCA Permission: %s", err)
	}

	return nil
}
