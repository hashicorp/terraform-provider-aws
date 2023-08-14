// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codeartifact_domain", name="Domain")
// @Tags(identifierAttribute="arn")
func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		DeleteWithoutTimeout: resourceDomainDelete,
		UpdateWithoutTimeout: resourceDomainUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encryption_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"repository_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)

	input := &codeartifact.CreateDomainInput{
		Domain: aws.String(d.Get("domain").(string)),
		Tags:   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("encryption_key"); ok {
		input.EncryptionKey = aws.String(v.(string))
	}

	v, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 2*time.Minute, func() (any, error) {
		return conn.CreateDomainWithContext(ctx, input)
	}, "ValidationException", "KMS key not found")
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeArtifact Domain: %s", err)
	}
	domain := v.(*codeartifact.CreateDomainOutput)

	d.SetId(aws.StringValue(domain.Domain.Arn))

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)

	domainOwner, domainName, err := DecodeDomainID(d.Id())
	if err != nil {
		return create.DiagError(names.CodeArtifact, create.ErrActionReading, ResNameDomain, d.Id(), err)
	}

	sm, err := conn.DescribeDomainWithContext(ctx, &codeartifact.DescribeDomainInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CodeArtifact, create.ErrActionReading, ResNameDomain, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CodeArtifact, create.ErrActionReading, ResNameDomain, d.Id(), err)
	}

	arn := aws.StringValue(sm.Domain.Arn)
	d.Set("domain", sm.Domain.Name)
	d.Set("arn", arn)
	d.Set("encryption_key", sm.Domain.EncryptionKey)
	d.Set("owner", sm.Domain.Owner)
	d.Set("asset_size_bytes", sm.Domain.AssetSizeBytes)
	d.Set("repository_count", sm.Domain.RepositoryCount)
	d.Set("created_time", sm.Domain.CreatedTime.Format(time.RFC3339))

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Printf("[DEBUG] Deleting CodeArtifact Domain: %s", d.Id())

	domainOwner, domainName, err := DecodeDomainID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Domain (%s): %s", d.Id(), err)
	}

	input := &codeartifact.DeleteDomainInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	}

	_, err = conn.DeleteDomainWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Domain (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeDomainID(id string) (string, string, error) {
	repoArn, err := arn.Parse(id)
	if err != nil {
		return "", "", err
	}

	domainName := strings.TrimPrefix(repoArn.Resource, "domain/")
	return repoArn.AccountID, domainName, nil
}
