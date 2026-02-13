// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package appsync

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appsync_domain_name", name="Domain Name")
func resourceDomainName() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainNameCreate,
		ReadWithoutTimeout:   resourceDomainNameRead,
		UpdateWithoutTimeout: resourceDomainNameUpdate,
		DeleteWithoutTimeout: resourceDomainNameDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"appsync_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainNameCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	input := &appsync.CreateDomainNameInput{
		CertificateArn: aws.String(d.Get(names.AttrCertificateARN).(string)),
		Description:    aws.String(d.Get(names.AttrDescription).(string)),
		DomainName:     aws.String(domainName),
	}

	output, err := conn.CreateDomainName(ctx, input)

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, domainName)
	}

	d.SetId(aws.ToString(output.DomainNameConfig.DomainName))

	return smerr.AppendEnrich(ctx, diags, resourceDomainNameRead(ctx, d, meta))
}

func resourceDomainNameRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	domainName, err := findDomainNameByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		smerr.AppendOne(ctx, diags, sdkdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	d.Set("appsync_domain_name", domainName.AppsyncDomainName)
	d.Set(names.AttrCertificateARN, domainName.CertificateArn)
	d.Set(names.AttrDescription, domainName.Description)
	d.Set(names.AttrDomainName, domainName.DomainName)
	d.Set(names.AttrHostedZoneID, domainName.HostedZoneId)

	return diags
}

func resourceDomainNameUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	input := &appsync.UpdateDomainNameInput{
		DomainName: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	_, err := conn.UpdateDomainName(ctx, input)

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return smerr.AppendEnrich(ctx, diags, resourceDomainNameRead(ctx, d, meta))
}

func resourceDomainNameDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	log.Printf("[INFO] Deleting Appsync Domain Name: %s", d.Id())
	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConcurrentModificationException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.DeleteDomainName(ctx, &appsync.DeleteDomainNameInput{
			DomainName: aws.String(d.Id()),
		})
	})

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return diags
}

func findDomainNameByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.DomainNameConfig, error) {
	input := &appsync.GetDomainNameInput{
		DomainName: aws.String(id),
	}

	output, err := conn.GetDomainName(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.DomainNameConfig == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.DomainNameConfig, nil
}
