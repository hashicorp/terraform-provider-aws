// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_domain_name_access_association", name="Domain Name Access Association")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;types.DomainNameAccessAssociation")
// @Testing(generator="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.RandomSubdomain()")
// @Testing(tlsKey=true, tlsKeyDomain="rName")
func resourceDomainNameAccessAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainNameAccessAssociationCreate,
		ReadWithoutTimeout:   resourceDomainNameAccessAssociationRead,
		UpdateWithoutTimeout: resourceDomainNameAccessAssociationUpdate,
		DeleteWithoutTimeout: resourceDomainNameAccessAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_association_source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"access_association_source_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AccessAssociationSourceType](),
				ForceNew:         true,
			},
			"domain_name_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
				ForceNew:     true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainNameAccessAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.CreateDomainNameAccessAssociationInput{
		AccessAssociationSource:     aws.String(d.Get("access_association_source").(string)),
		AccessAssociationSourceType: awstypes.AccessAssociationSourceType(d.Get("access_association_source_type").(string)),
		DomainNameArn:               aws.String(d.Get("domain_name_arn").(string)),
		Tags:                        getTagsIn(ctx),
	}

	out, err := conn.CreateDomainNameAccessAssociation(ctx, input)

	id := aws.ToString(out.DomainNameAccessAssociationArn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Domain Name Access Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceDomainNameAccessAssociationRead(ctx, d, meta)...)
}

func resourceDomainNameAccessAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	out, err := findDomainNameAccessAssociationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Domain Name Access Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Domain Name Access Association (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, out.Tags)

	d.Set(names.AttrARN, out.DomainNameAccessAssociationArn)
	d.Set("access_association_source", out.AccessAssociationSource)
	d.Set("access_association_source_type", out.AccessAssociationSourceType)
	d.Set("domain_name_arn", out.DomainNameArn)

	return diags
}

func resourceDomainNameAccessAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Tags only.
	return append(diags, resourceDomainNameAccessAssociationRead(ctx, d, meta)...)
}

func resourceDomainNameAccessAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Domain Name Access Association: %s", d.Id())
	_, err := conn.DeleteDomainNameAccessAssociation(ctx, &apigateway.DeleteDomainNameAccessAssociationInput{
		DomainNameAccessAssociationArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Domain Name Access Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findDomainNameAccessAssociationByARN(ctx context.Context, conn *apigateway.Client, arn string) (*awstypes.DomainNameAccessAssociation, error) {
	input := &apigateway.GetDomainNameAccessAssociationsInput{
		ResourceOwner: awstypes.ResourceOwnerSelf,
	}

	return findDomainNameAccessAssociation(ctx, conn, input, func(v *awstypes.DomainNameAccessAssociation) bool {
		return aws.ToString(v.DomainNameAccessAssociationArn) == arn
	})
}

func findDomainNameAccessAssociation(ctx context.Context, conn *apigateway.Client, input *apigateway.GetDomainNameAccessAssociationsInput, filter tfslices.Predicate[*awstypes.DomainNameAccessAssociation]) (*awstypes.DomainNameAccessAssociation, error) {
	output, err := findDomainNameAccessAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDomainNameAccessAssociations(ctx context.Context, conn *apigateway.Client, input *apigateway.GetDomainNameAccessAssociationsInput, filter tfslices.Predicate[*awstypes.DomainNameAccessAssociation]) ([]awstypes.DomainNameAccessAssociation, error) {
	var output []awstypes.DomainNameAccessAssociation

	err := getDomainNameAccessAssociationsPages(ctx, conn, input, func(page *apigateway.GetDomainNameAccessAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
