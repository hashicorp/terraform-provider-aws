package iam

import ( // nosemgrep: aws-sdk-go-multiple-service-imports

	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSigningCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceSigningCertificateCreate,
		Read:   resourceSigningCertificateRead,
		Update: resourceSigningCertificateUpdate,
		Delete: resourceSigningCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"certificate_body": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressNormalizeCertRemoval,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      iam.StatusTypeActive,
				ValidateFunc: validation.StringInSlice(iam.StatusType_Values(), false),
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSigningCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	createOpts := &iam.UploadSigningCertificateInput{
		CertificateBody: aws.String(d.Get("certificate_body").(string)),
		UserName:        aws.String(d.Get("user_name").(string)),
	}

	log.Printf("[DEBUG] Creating IAM Signing Certificate with opts: %s", createOpts)
	resp, err := conn.UploadSigningCertificate(createOpts)
	if err != nil {
		return fmt.Errorf("error uploading IAM Signing Certificate: %w", err)
	}

	cert := resp.Certificate
	certId := cert.CertificateId
	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(certId), aws.StringValue(cert.UserName)))

	if v, ok := d.GetOk("status"); ok && v.(string) != iam.StatusTypeActive {
		updateInput := &iam.UpdateSigningCertificateInput{
			CertificateId: certId,
			UserName:      aws.String(d.Get("user_name").(string)),
			Status:        aws.String(v.(string)),
		}

		_, err := conn.UpdateSigningCertificate(updateInput)
		if err != nil {
			return fmt.Errorf("error settings IAM Signing Certificate status: %w", err)
		}
	}

	return resourceSigningCertificateRead(d, meta)
}

func resourceSigningCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	certId, userName, err := DecodeSigningCertificateId(d.Id())
	if err != nil {
		return err
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindSigningCertificate(conn, userName, certId)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Signing Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Signing Certificate (%s): %w", d.Id(), err)
	}

	resp := outputRaw.(*iam.SigningCertificate)

	d.Set("certificate_body", resp.CertificateBody)
	d.Set("certificate_id", resp.CertificateId)
	d.Set("user_name", resp.UserName)
	d.Set("status", resp.Status)

	return nil
}

func resourceSigningCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	certId, userName, err := DecodeSigningCertificateId(d.Id())
	if err != nil {
		return err
	}

	updateInput := &iam.UpdateSigningCertificateInput{
		CertificateId: aws.String(certId),
		UserName:      aws.String(userName),
		Status:        aws.String(d.Get("status").(string)),
	}

	_, err = conn.UpdateSigningCertificate(updateInput)
	if err != nil {
		return fmt.Errorf("error updating IAM Signing Certificate: %w", err)
	}

	return resourceSigningCertificateRead(d, meta)
}

func resourceSigningCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	log.Printf("[INFO] Deleting IAM Signing Certificate: %s", d.Id())

	certId, userName, err := DecodeSigningCertificateId(d.Id())
	if err != nil {
		return err
	}

	input := &iam.DeleteSigningCertificateInput{
		CertificateId: aws.String(certId),
		UserName:      aws.String(userName),
	}

	if _, err := conn.DeleteSigningCertificate(input); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}
		return fmt.Errorf("Error deleting IAM Signing Certificate %s: %s", d.Id(), err)
	}

	return nil
}

func DecodeSigningCertificateId(id string) (string, string, error) {
	creds := strings.Split(id, ":")
	if len(creds) != 2 {
		return "", "", fmt.Errorf("unknown IAM Signing Certificate ID format")
	}

	certId := creds[0]
	userName := creds[1]

	return certId, userName, nil
}
