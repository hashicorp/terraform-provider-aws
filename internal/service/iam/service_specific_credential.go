package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceServiceSpecificCredential() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceSpecificCredentialCreate,
		ReadWithoutTimeout:   resourceServiceSpecificCredentialRead,
		UpdateWithoutTimeout: resourceServiceSpecificCredentialUpdate,
		DeleteWithoutTimeout: resourceServiceSpecificCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      iam.StatusTypeActive,
				ValidateFunc: validation.StringInSlice(iam.StatusType_Values(), false),
			},
			"service_password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"service_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_specific_credential_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServiceSpecificCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	input := &iam.CreateServiceSpecificCredentialInput{
		ServiceName: aws.String(d.Get("service_name").(string)),
		UserName:    aws.String(d.Get("user_name").(string)),
	}

	out, err := conn.CreateServiceSpecificCredentialWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Service Specific Credential: %s", err)
	}

	cred := out.ServiceSpecificCredential

	d.SetId(fmt.Sprintf("%s:%s:%s", aws.StringValue(cred.ServiceName), aws.StringValue(cred.UserName), aws.StringValue(cred.ServiceSpecificCredentialId)))
	d.Set("service_password", cred.ServicePassword)

	if v, ok := d.GetOk("status"); ok && v.(string) != iam.StatusTypeActive {
		updateInput := &iam.UpdateServiceSpecificCredentialInput{
			ServiceSpecificCredentialId: cred.ServiceSpecificCredentialId,
			UserName:                    cred.UserName,
			Status:                      aws.String(v.(string)),
		}

		_, err := conn.UpdateServiceSpecificCredentialWithContext(ctx, updateInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "settings IAM Service Specific Credential status: %s", err)
		}
	}

	return append(diags, resourceServiceSpecificCredentialRead(ctx, d, meta)...)
}

func resourceServiceSpecificCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	serviceName, userName, credID, err := DecodeServiceSpecificCredentialId(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Service Specific Credential (%s): %s", d.Id(), err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindServiceSpecificCredential(ctx, conn, serviceName, userName, credID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Service Specific Credential (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Service Specific Credential (%s): %s", d.Id(), err)
	}

	cred := outputRaw.(*iam.ServiceSpecificCredentialMetadata)

	d.Set("service_specific_credential_id", cred.ServiceSpecificCredentialId)
	d.Set("service_user_name", cred.ServiceUserName)
	d.Set("service_name", cred.ServiceName)
	d.Set("user_name", cred.UserName)
	d.Set("status", cred.Status)

	return diags
}

func resourceServiceSpecificCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	request := &iam.UpdateServiceSpecificCredentialInput{
		ServiceSpecificCredentialId: aws.String(d.Get("service_specific_credential_id").(string)),
		UserName:                    aws.String(d.Get("user_name").(string)),
		Status:                      aws.String(d.Get("status").(string)),
	}
	_, err := conn.UpdateServiceSpecificCredentialWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Service Specific Credential %s: %s", d.Id(), err)
	}

	return append(diags, resourceServiceSpecificCredentialRead(ctx, d, meta)...)
}

func resourceServiceSpecificCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	request := &iam.DeleteServiceSpecificCredentialInput{
		ServiceSpecificCredentialId: aws.String(d.Get("service_specific_credential_id").(string)),
		UserName:                    aws.String(d.Get("user_name").(string)),
	}

	if _, err := conn.DeleteServiceSpecificCredentialWithContext(ctx, request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting IAM Service Specific Credential %s: %s", d.Id(), err)
	}
	return diags
}

func DecodeServiceSpecificCredentialId(id string) (string, string, string, error) {
	creds := strings.Split(id, ":")
	if len(creds) != 3 {
		return "", "", "", fmt.Errorf("unknown IAM Service Specific Credential ID format")
	}
	serviceName := creds[0]
	userName := creds[1]
	credId := creds[2]

	return serviceName, userName, credId, nil
}
