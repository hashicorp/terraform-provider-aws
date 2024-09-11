// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_api_key", name="API Key")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;apigateway.GetApiKeyOutput")
func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPIKeyCreate,
		ReadWithoutTimeout:   resourceAPIKeyRead,
		UpdateWithoutTimeout: resourceAPIKeyUpdate,
		DeleteWithoutTimeout: resourceAPIKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrLastUpdatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrValue: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(20, 128),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &apigateway.CreateApiKeyInput{
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		Enabled:     d.Get(names.AttrEnabled).(bool),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
		Value:       aws.String(d.Get(names.AttrValue).(string)),
	}

	if v, ok := d.GetOk("customer_id"); ok {
		input.CustomerId = aws.String(v.(string))
	}

	apiKey, err := conn.CreateApiKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway API Key (%s): %s", name, err)
	}

	d.SetId(aws.ToString(apiKey.Id))

	return append(diags, resourceAPIKeyRead(ctx, d, meta)...)
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiKey, err := findAPIKeyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway API Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway API Key (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, apiKey.Tags)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/apikeys/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, apiKey.CreatedDate.Format(time.RFC3339))
	d.Set("customer_id", apiKey.CustomerId)
	d.Set(names.AttrDescription, apiKey.Description)
	d.Set(names.AttrEnabled, apiKey.Enabled)
	d.Set(names.AttrLastUpdatedDate, apiKey.LastUpdatedDate.Format(time.RFC3339))
	d.Set(names.AttrName, apiKey.Name)
	d.Set(names.AttrValue, apiKey.Value)

	return diags
}

func resourceAPIKeyUpdateOperations(d *schema.ResourceData) []types.PatchOperation {
	operations := make([]types.PatchOperation, 0)

	if d.HasChange(names.AttrEnabled) {
		isEnabled := "false"
		if d.Get(names.AttrEnabled).(bool) {
			isEnabled = "true"
		}
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/enabled"),
			Value: aws.String(isEnabled),
		})
	}

	if d.HasChange(names.AttrDescription) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/description"),
			Value: aws.String(d.Get(names.AttrDescription).(string)),
		})
	}

	if d.HasChange(names.AttrName) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/name"),
			Value: aws.String(d.Get(names.AttrName).(string)),
		})
	}

	if d.HasChange("customer_id") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/customerId"),
			Value: aws.String(d.Get("customer_id").(string)),
		})
	}

	return operations
}

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		_, err := conn.UpdateApiKey(ctx, &apigateway.UpdateApiKeyInput{
			ApiKey:          aws.String(d.Id()),
			PatchOperations: resourceAPIKeyUpdateOperations(d),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway API Key (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAPIKeyRead(ctx, d, meta)...)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway API Key: %s", d.Id())
	_, err := conn.DeleteApiKey(ctx, &apigateway.DeleteApiKeyInput{
		ApiKey: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway API Key (%s): %s", d.Id(), err)
	}

	return diags
}

func findAPIKeyByID(ctx context.Context, conn *apigateway.Client, id string) (*apigateway.GetApiKeyOutput, error) {
	input := &apigateway.GetApiKeyInput{
		ApiKey:       aws.String(id),
		IncludeValue: aws.Bool(true),
	}

	output, err := conn.GetApiKey(ctx, input)

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

	return output, nil
}
