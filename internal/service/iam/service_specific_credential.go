// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_service_specific_credential", name="Service Specific Credential")
func resourceServiceSpecificCredential() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceSpecificCredentialCreate,
		ReadWithoutTimeout:   resourceServiceSpecificCredentialRead,
		UpdateWithoutTimeout: resourceServiceSpecificCredentialUpdate,
		DeleteWithoutTimeout: resourceServiceSpecificCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrUserName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.StatusTypeActive,
				ValidateDiagFunc: enum.Validate[awstypes.StatusType](),
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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	input := &iam.CreateServiceSpecificCredentialInput{
		ServiceName: aws.String(d.Get(names.AttrServiceName).(string)),
		UserName:    aws.String(d.Get(names.AttrUserName).(string)),
	}

	out, err := conn.CreateServiceSpecificCredential(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Service Specific Credential: %s", err)
	}

	cred := out.ServiceSpecificCredential

	d.SetId(fmt.Sprintf("%s:%s:%s", aws.ToString(cred.ServiceName), aws.ToString(cred.UserName), aws.ToString(cred.ServiceSpecificCredentialId)))
	d.Set("service_password", cred.ServicePassword)

	if v, ok := d.GetOk(names.AttrStatus); ok && v.(string) != string(awstypes.StatusTypeActive) {
		updateInput := &iam.UpdateServiceSpecificCredentialInput{
			ServiceSpecificCredentialId: cred.ServiceSpecificCredentialId,
			UserName:                    cred.UserName,
			Status:                      awstypes.StatusType(v.(string)),
		}

		_, err := conn.UpdateServiceSpecificCredential(ctx, updateInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "settings IAM Service Specific Credential status: %s", err)
		}
	}

	return append(diags, resourceServiceSpecificCredentialRead(ctx, d, meta)...)
}

func resourceServiceSpecificCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

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

	cred := outputRaw.(*awstypes.ServiceSpecificCredentialMetadata)

	d.Set("service_specific_credential_id", cred.ServiceSpecificCredentialId)
	d.Set("service_user_name", cred.ServiceUserName)
	d.Set(names.AttrServiceName, cred.ServiceName)
	d.Set(names.AttrUserName, cred.UserName)
	d.Set(names.AttrStatus, cred.Status)

	return diags
}

func resourceServiceSpecificCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	request := &iam.UpdateServiceSpecificCredentialInput{
		ServiceSpecificCredentialId: aws.String(d.Get("service_specific_credential_id").(string)),
		UserName:                    aws.String(d.Get(names.AttrUserName).(string)),
		Status:                      awstypes.StatusType(d.Get(names.AttrStatus).(string)),
	}
	_, err := conn.UpdateServiceSpecificCredential(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Service Specific Credential %s: %s", d.Id(), err)
	}

	return append(diags, resourceServiceSpecificCredentialRead(ctx, d, meta)...)
}

func resourceServiceSpecificCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	request := &iam.DeleteServiceSpecificCredentialInput{
		ServiceSpecificCredentialId: aws.String(d.Get("service_specific_credential_id").(string)),
		UserName:                    aws.String(d.Get(names.AttrUserName).(string)),
	}

	if _, err := conn.DeleteServiceSpecificCredential(ctx, request); err != nil {
		if errs.IsA[*awstypes.NoSuchEntityException](err) {
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
