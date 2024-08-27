// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codeartifact_domain", name="Domain")
// @Tags(identifierAttribute="arn")
func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		DeleteWithoutTimeout: resourceDomainDelete,
		UpdateWithoutTimeout: resourceDomainUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_size_bytes": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
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
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"s3_bucket_arn": {
				Type:     schema.TypeString,
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
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	domain := d.Get(names.AttrDomain).(string)
	input := &codeartifact.CreateDomainInput{
		Domain: aws.String(domain),
		Tags:   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("encryption_key"); ok {
		input.EncryptionKey = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.ValidationException](ctx, propagationTimeout, func() (any, error) {
		return conn.CreateDomain(ctx, input)
	}, "KMS key not found")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeArtifact Domain (%s): %s", domain, err)
	}

	d.SetId(aws.ToString(outputRaw.(*codeartifact.CreateDomainOutput).Domain.Arn))

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, err := parseDomainARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	domain, err := findDomainByTwoPartKey(ctx, conn, owner, domainName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeArtifact Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeArtifact Domain (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, domain.Arn)
	d.Set("asset_size_bytes", strconv.FormatInt(domain.AssetSizeBytes, 10))
	d.Set(names.AttrCreatedTime, domain.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrDomain, domain.Name)
	d.Set("encryption_key", domain.EncryptionKey)
	d.Set(names.AttrOwner, domain.Owner)
	d.Set("repository_count", domain.RepositoryCount)
	d.Set("s3_bucket_arn", domain.S3BucketArn)

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, err := parseDomainARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &codeartifact.DeleteDomainInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(owner),
	}

	log.Printf("[DEBUG] Deleting CodeArtifact Domain: %s", d.Id())
	_, err = conn.DeleteDomain(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Domain (%s): %s", d.Id(), err)
	}

	return diags
}

func parseDomainARN(v string) (string, string, error) {
	// arn:${Partition}:codeartifact:${Region}:${Account}:domain/${DomainName}
	arn, err := arn.Parse(v)
	if err != nil {
		return "", "", err
	}

	return arn.AccountID, strings.TrimPrefix(arn.Resource, "domain/"), nil
}

func findDomainByTwoPartKey(ctx context.Context, conn *codeartifact.Client, owner, domainName string) (*types.DomainDescription, error) {
	input := &codeartifact.DescribeDomainInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(owner),
	}

	output, err := conn.DescribeDomain(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Domain == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Domain, nil
}
