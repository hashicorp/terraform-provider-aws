// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_bot_association", name="Bot Association")
func resourceBotAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBotAssociationCreate,
		ReadWithoutTimeout:   resourceBotAssociationRead,
		DeleteWithoutTimeout: resourceBotAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lex_bot": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lex_region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(2, 50),
						},
					},
				},
			},
		},
	}
}

func resourceBotAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.AssociateBotInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("lex_bot"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.LexBot = expandLexBot(v.([]any)[0].(map[string]any))
		if input.LexBot.LexRegion == nil {
			input.LexBot.LexRegion = aws.String(meta.(*conns.AWSClient).Region(ctx))
		}
	}

	id := botAssociationCreateResourceID(instanceID, aws.ToString(input.LexBot.Name), aws.ToString(input.LexBot.LexRegion))

	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, timeout, func() (any, error) {
		return conn.AssociateBot(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Bot Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceBotAssociationRead(ctx, d, meta)...)
}

func resourceBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, name, region, err := botAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	lexBot, err := findBotAssociationByThreePartKey(ctx, conn, instanceID, name, region)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Bot Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Bot Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrInstanceID, instanceID)
	if err := d.Set("lex_bot", flattenLexBot(lexBot)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lex_bot: %s", err)
	}

	return diags
}

func resourceBotAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, name, region, err := botAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Bot Association: %s", d.Id())
	input := connect.DisassociateBotInput{
		InstanceId: aws.String(instanceID),
		LexBot: &awstypes.LexBot{
			Name:      aws.String(name),
			LexRegion: aws.String(region),
		},
	}
	_, err = conn.DisassociateBot(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Bot Association (%s): %s", d.Id(), err)
	}

	return diags
}

const botAssociationResourceIDSeparator = ":"

func botAssociationCreateResourceID(instanceID, botName, region string) string {
	parts := []string{instanceID, botName, region}
	id := strings.Join(parts, botAssociationResourceIDSeparator)

	return id
}

func botAssociationParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, botAssociationResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected instanceID%[2]sname%[2]sregion", id, botAssociationResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findBotAssociationByThreePartKey(ctx context.Context, conn *connect.Client, instanceID, name, region string) (*awstypes.LexBot, error) {
	const maxResults = 25
	input := &connect.ListBotsInput{
		InstanceId: aws.String(instanceID),
		LexVersion: awstypes.LexVersionV1,
		MaxResults: aws.Int32(maxResults),
	}

	return findLexBot(ctx, conn, input, func(v *awstypes.LexBotConfig) bool {
		return aws.ToString(v.LexBot.Name) == name && aws.ToString(v.LexBot.LexRegion) == region
	})
}

func findLexBot(ctx context.Context, conn *connect.Client, input *connect.ListBotsInput, filter tfslices.Predicate[*awstypes.LexBotConfig]) (*awstypes.LexBot, error) {
	output, err := findBot(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	if output.LexBot == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output.LexBot, nil
}

func findBot(ctx context.Context, conn *connect.Client, input *connect.ListBotsInput, filter tfslices.Predicate[*awstypes.LexBotConfig]) (*awstypes.LexBotConfig, error) {
	output, err := findBots(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findBots(ctx context.Context, conn *connect.Client, input *connect.ListBotsInput, filter tfslices.Predicate[*awstypes.LexBotConfig]) ([]awstypes.LexBotConfig, error) {
	var output []awstypes.LexBotConfig

	pages := connect.NewListBotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.LexBots {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func expandLexBot(tfMap map[string]any) *awstypes.LexBot {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LexBot{
		Name: aws.String(tfMap[names.AttrName].(string)),
	}

	if v, ok := tfMap["lex_region"].(string); ok && v != "" {
		apiObject.LexRegion = aws.String(v)
	}

	return apiObject
}

func flattenLexBot(apiObject *awstypes.LexBot) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"lex_region":   aws.ToString(apiObject.LexRegion),
		names.AttrName: aws.ToString(apiObject.Name),
	}

	return []any{tfMap}
}
