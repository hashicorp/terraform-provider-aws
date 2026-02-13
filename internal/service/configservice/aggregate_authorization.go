// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package configservice

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_aggregate_authorization", name="Aggregate Authorization")
// @Tags(identifierAttribute="arn")
// @Region(overrideEnabled=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/configservice/types;awstypes;awstypes.AggregationAuthorization")
// @Testing(generator=false)
// @Testing(serialize=true)
func resourceAggregateAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAggregateAuthorizationCreate,
		ReadWithoutTimeout:   resourceAggregateAuthorizationRead,
		UpdateWithoutTimeout: resourceAggregateAuthorizationUpdate,
		DeleteWithoutTimeout: resourceAggregateAuthorizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorized_aws_region": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"authorized_aws_region", names.AttrRegion},
			},
			names.AttrRegion: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"authorized_aws_region", names.AttrRegion},
				Deprecated:   "region is deprecated. Use authorized_aws_region instead.",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAggregateAuthorizationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	var authorizedRegion string
	if v, ok := d.GetOk("authorized_aws_region"); ok {
		authorizedRegion = v.(string)
	} else if v, ok := d.GetOk(names.AttrRegion); ok {
		authorizedRegion = v.(string)
	}
	id := aggregateAuthorizationCreateResourceID(accountID, authorizedRegion)
	input := &configservice.PutAggregationAuthorizationInput{
		AuthorizedAccountId: aws.String(accountID),
		AuthorizedAwsRegion: aws.String(authorizedRegion),
		Tags:                getTagsIn(ctx),
	}

	_, err := conn.PutAggregationAuthorization(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ConfigService Aggregate Authorization (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceAggregateAuthorizationRead(ctx, d, meta)...)
}

func resourceAggregateAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	accountID, authorizedRegion, err := aggregateAuthorizationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	aggregationAuthorization, err := findAggregateAuthorizationByTwoPartKey(ctx, conn, accountID, authorizedRegion)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ConfigService Aggregate Authorization (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Aggregate Authorization (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, aggregationAuthorization.AuthorizedAccountId)
	d.Set(names.AttrARN, aggregationAuthorization.AggregationAuthorizationArn)
	if _, ok := d.GetOk(names.AttrRegion); ok {
		d.Set(names.AttrRegion, aggregationAuthorization.AuthorizedAwsRegion)
	} else {
		d.Set("authorized_aws_region", aggregationAuthorization.AuthorizedAwsRegion)
	}

	return diags
}

func resourceAggregateAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceAggregateAuthorizationRead(ctx, d, meta)...)
}

func resourceAggregateAuthorizationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	accountID, authorizedRegion, err := aggregateAuthorizationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting ConfigService Aggregate Authorization: %s", d.Id())
	input := configservice.DeleteAggregationAuthorizationInput{
		AuthorizedAccountId: aws.String(accountID),
		AuthorizedAwsRegion: aws.String(authorizedRegion),
	}
	_, err = conn.DeleteAggregationAuthorization(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Aggregate Authorization (%s): %s", d.Id(), err)
	}

	return diags
}

const aggregateAuthorizationResourceIDSeparator = ":"

func aggregateAuthorizationCreateResourceID(accountID, authorizedRegion string) string {
	parts := []string{accountID, authorizedRegion}
	id := strings.Join(parts, aggregateAuthorizationResourceIDSeparator)

	return id
}

func aggregateAuthorizationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, aggregateAuthorizationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected account_id%[2]sauthorized_aws_region", id, aggregateAuthorizationResourceIDSeparator)
}

func findAggregateAuthorizationByTwoPartKey(ctx context.Context, conn *configservice.Client, accountID, authorizedRegion string) (*types.AggregationAuthorization, error) {
	input := &configservice.DescribeAggregationAuthorizationsInput{}

	return findAggregateAuthorization(ctx, conn, input, func(v *types.AggregationAuthorization) bool {
		return aws.ToString(v.AuthorizedAccountId) == accountID && aws.ToString(v.AuthorizedAwsRegion) == authorizedRegion
	})
}

func findAggregateAuthorization(ctx context.Context, conn *configservice.Client, input *configservice.DescribeAggregationAuthorizationsInput, filter tfslices.Predicate[*types.AggregationAuthorization]) (*types.AggregationAuthorization, error) {
	output, err := findAggregateAuthorizations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAggregateAuthorizations(ctx context.Context, conn *configservice.Client, input *configservice.DescribeAggregationAuthorizationsInput, filter tfslices.Predicate[*types.AggregationAuthorization]) ([]types.AggregationAuthorization, error) {
	var output []types.AggregationAuthorization

	pages := configservice.NewDescribeAggregationAuthorizationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AggregationAuthorizations {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
