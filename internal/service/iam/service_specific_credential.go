// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credential_age_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 36600),
			},
			"expiration_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_credential_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_credential_secret": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"service_specific_credential_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.StatusTypeActive,
				ValidateDiagFunc: enum.Validate[awstypes.StatusType](),
			},
			names.AttrUserName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceServiceSpecificCredentialCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	serviceName, userName := d.Get(names.AttrServiceName).(string), d.Get(names.AttrUserName).(string)
	input := iam.CreateServiceSpecificCredentialInput{
		ServiceName: aws.String(serviceName),
		UserName:    aws.String(userName),
	}

	if v, ok := d.GetOk("credential_age_days"); ok {
		input.CredentialAgeDays = aws.Int32(int32(v.(int)))
	}

	output, err := conn.CreateServiceSpecificCredential(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Service-Specific Credential: %s", err)
	}

	cred := output.ServiceSpecificCredential
	credID := aws.ToString(cred.ServiceSpecificCredentialId)
	d.SetId(serviceSpecificCredentialCreateResourceID(serviceName, userName, credID))
	d.Set("service_credential_secret", cred.ServiceCredentialSecret)
	d.Set("service_password", cred.ServicePassword)

	if v, ok := d.GetOk(names.AttrStatus); ok && awstypes.StatusType(v.(string)) != awstypes.StatusTypeActive {
		input := iam.UpdateServiceSpecificCredentialInput{
			ServiceSpecificCredentialId: aws.String(credID),
			Status:                      awstypes.StatusType(v.(string)),
			UserName:                    aws.String(userName),
		}

		_, err := conn.UpdateServiceSpecificCredential(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM Service-Specific Credential (%s) status: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceSpecificCredentialRead(ctx, d, meta)...)
}

func resourceServiceSpecificCredentialRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	serviceName, userName, credID, err := serviceSpecificCredentialParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	cred, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func(ctx context.Context) (*awstypes.ServiceSpecificCredentialMetadata, error) {
		return findServiceSpecificCredentialByThreePartKey(ctx, conn, serviceName, userName, credID)
	}, d.IsNewResource())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM Service Specific Credential (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Service Specific Credential (%s): %s", d.Id(), err)
	}

	d.Set("create_date", cred.CreateDate.Format(time.RFC3339))
	if cred.ExpirationDate != nil {
		d.Set("expiration_date", cred.ExpirationDate.Format(time.RFC3339))
	}
	d.Set("service_credential_alias", cred.ServiceCredentialAlias)
	d.Set(names.AttrServiceName, cred.ServiceName)
	d.Set("service_specific_credential_id", cred.ServiceSpecificCredentialId)
	d.Set("service_user_name", cred.ServiceUserName)
	d.Set(names.AttrStatus, cred.Status)
	d.Set(names.AttrUserName, cred.UserName)

	return diags
}

func resourceServiceSpecificCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	_, userName, credID, err := serviceSpecificCredentialParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := iam.UpdateServiceSpecificCredentialInput{
		ServiceSpecificCredentialId: aws.String(credID),
		Status:                      awstypes.StatusType(d.Get(names.AttrStatus).(string)),
		UserName:                    aws.String(userName),
	}
	_, err = conn.UpdateServiceSpecificCredential(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Service-Specific Credential (%s): %s", d.Id(), err)
	}

	return append(diags, resourceServiceSpecificCredentialRead(ctx, d, meta)...)
}

func resourceServiceSpecificCredentialDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	_, userName, credID, err := serviceSpecificCredentialParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting IAM Service-Specific Credential: %s", d.Id())
	input := iam.DeleteServiceSpecificCredentialInput{
		ServiceSpecificCredentialId: aws.String(credID),
		UserName:                    aws.String(userName),
	}
	_, err = conn.DeleteServiceSpecificCredential(ctx, &input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Service-Specific Credential (%s): %s", d.Id(), err)
	}

	return diags
}

const serviceSpecificCredentialResourceIDSeparator = ":"

func serviceSpecificCredentialCreateResourceID(serviceName, userName, serviceSpecificCredentialID string) string {
	parts := []string{serviceName, userName, serviceSpecificCredentialID}
	id := strings.Join(parts, serviceSpecificCredentialResourceIDSeparator)

	return id
}

func serviceSpecificCredentialParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, serviceSpecificCredentialResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SERVICE-NAME%[2]sUSER-NAME%[2]sSERVICE-SPECIFIC-CREDENTIAL-ID", id, serviceSpecificCredentialResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findServiceSpecificCredentialByThreePartKey(ctx context.Context, conn *iam.Client, serviceName, userName, serviceSpecificCredentialID string) (*awstypes.ServiceSpecificCredentialMetadata, error) {
	input := iam.ListServiceSpecificCredentialsInput{
		ServiceName: aws.String(serviceName),
		UserName:    aws.String(userName),
	}

	return findServiceSpecificCredential(ctx, conn, &input, func(v *awstypes.ServiceSpecificCredentialMetadata) bool {
		return aws.ToString(v.ServiceSpecificCredentialId) == serviceSpecificCredentialID
	})
}

func findServiceSpecificCredential(ctx context.Context, conn *iam.Client, input *iam.ListServiceSpecificCredentialsInput, filter tfslices.Predicate[*awstypes.ServiceSpecificCredentialMetadata]) (*awstypes.ServiceSpecificCredentialMetadata, error) {
	output, err := findServiceSpecificCredentials(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findServiceSpecificCredentials(ctx context.Context, conn *iam.Client, input *iam.ListServiceSpecificCredentialsInput, filter tfslices.Predicate[*awstypes.ServiceSpecificCredentialMetadata]) ([]awstypes.ServiceSpecificCredentialMetadata, error) {
	var output []awstypes.ServiceSpecificCredentialMetadata

	err := listServiceSpecificCredentialsPages(ctx, conn, input, func(page *iam.ListServiceSpecificCredentialsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ServiceSpecificCredentials {
			if p := &v; !inttypes.IsZero(p) && filter(p) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
