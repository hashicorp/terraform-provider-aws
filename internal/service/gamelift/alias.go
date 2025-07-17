// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_gamelift_alias", name="Alias")
// @Tags(identifierAttribute="arn")
func resourceAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAliasCreate,
		ReadWithoutTimeout:   resourceAliasRead,
		UpdateWithoutTimeout: resourceAliasUpdate,
		DeleteWithoutTimeout: resourceAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"routing_strategy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fleet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrMessage: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RoutingStrategyType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAliasCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &gamelift.CreateAliasInput{
		Name:            aws.String(name),
		RoutingStrategy: expandRoutingStrategy(d.Get("routing_strategy").([]any)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateAlias(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Alias (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Alias.AliasId))

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	alias, err := findAliasByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Alias (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, alias.AliasArn)
	d.Set(names.AttrDescription, alias.Description)
	d.Set(names.AttrName, alias.Name)
	if err := d.Set("routing_strategy", flattenRoutingStrategy(alias.RoutingStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting routing_strategy: %s", err)
	}

	return diags
}

func resourceAliasUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &gamelift.UpdateAliasInput{
			AliasId:         aws.String(d.Id()),
			Name:            aws.String(d.Get(names.AttrName).(string)),
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
			RoutingStrategy: expandRoutingStrategy(d.Get("routing_strategy").([]any)),
		}

		_, err := conn.UpdateAlias(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Alias (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Alias: %s", d.Id())
	_, err := conn.DeleteAlias(ctx, &gamelift.DeleteAliasInput{
		AliasId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Alias (%s): %s", d.Id(), err)
	}

	return diags
}

func findAliasByID(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.Alias, error) {
	input := &gamelift.DescribeAliasInput{
		AliasId: aws.String(id),
	}

	output, err := conn.DescribeAlias(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Alias == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Alias, nil
}

func expandRoutingStrategy(tfList []any) *awstypes.RoutingStrategy {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.RoutingStrategy{
		Type: awstypes.RoutingStrategyType(tfMap[names.AttrType].(string)),
	}

	if v, ok := tfMap["fleet_id"].(string); ok && len(v) > 0 {
		apiObject.FleetId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrMessage].(string); ok && len(v) > 0 {
		apiObject.Message = aws.String(v)
	}

	return apiObject
}

func flattenRoutingStrategy(apiObject *awstypes.RoutingStrategy) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if apiObject.FleetId != nil {
		tfMap["fleet_id"] = aws.ToString(apiObject.FleetId)
	}

	if apiObject.Message != nil {
		tfMap[names.AttrMessage] = aws.ToString(apiObject.Message)
	}

	tfMap[names.AttrType] = apiObject.Type

	return []any{tfMap}
}
