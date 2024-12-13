// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
				ValidateDiagFunc: enum.Validate[types.AccessAssociationSourceType](),
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
		AccessAssociationSourceType: types.AccessAssociationSourceType(d.Get("access_association_source_type").(string)),
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

	out, err := findDomainNameAccessAssociationByID(ctx, conn, d.Id())

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

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Domain Name Access Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findDomainNameAccessAssociationByID(ctx context.Context, conn *apigateway.Client, id string) (*types.DomainNameAccessAssociation, error) {
	input := &apigateway.GetDomainNameAccessAssociationsInput{
		ResourceOwner: types.ResourceOwnerSelf,
	}

	output, err := conn.GetDomainNameAccessAssociations(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if len(output.Items) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	var domainNameAccessAssociation *types.DomainNameAccessAssociation

	for _, item := range output.Items {
		if aws.ToString(item.DomainNameAccessAssociationArn) == id {
			domainNameAccessAssociation = &item
			break
		}
	}

	if domainNameAccessAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return domainNameAccessAssociation, nil
}
