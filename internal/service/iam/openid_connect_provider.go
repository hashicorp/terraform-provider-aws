// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_openid_connect_provider", name="OIDC Provider")
// @Tags(identifierAttribute="id", resourceType="OIDCProvider")
// @Testing(name="OpenIDConnectProvider")
func resourceOpenIDConnectProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOpenIDConnectProviderCreate,
		ReadWithoutTimeout:   resourceOpenIDConnectProviderRead,
		UpdateWithoutTimeout: resourceOpenIDConnectProviderUpdate,
		DeleteWithoutTimeout: resourceOpenIDConnectProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id_list": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"thumbprint_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(40, 40),
				},
			},
			names.AttrURL: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validOpenIDURL,
				DiffSuppressFunc: suppressOpenIDURL,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOpenIDConnectProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	input := &iam.CreateOpenIDConnectProviderInput{
		ClientIDList:   flex.ExpandStringValueSet(d.Get("client_id_list").(*schema.Set)),
		Tags:           getTagsIn(ctx),
		ThumbprintList: flex.ExpandStringValueList(d.Get("thumbprint_list").([]interface{})),
		Url:            aws.String(d.Get(names.AttrURL).(string)),
	}

	output, err := conn.CreateOpenIDConnectProvider(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	partition := meta.(*conns.AWSClient).Partition
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateOpenIDConnectProvider(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM OIDC Provider: %s", err)
	}

	d.SetId(aws.ToString(output.OpenIDConnectProviderArn))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := openIDConnectProviderCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceOpenIDConnectProviderRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM OIDC Provider (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceOpenIDConnectProviderRead(ctx, d, meta)...)
}

func resourceOpenIDConnectProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	output, err := findOpenIDConnectProviderByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM OIDC Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM OIDC Provider (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, d.Id())
	d.Set("client_id_list", output.ClientIDList)
	d.Set("thumbprint_list", output.ThumbprintList)
	d.Set(names.AttrURL, output.Url)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceOpenIDConnectProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChange("thumbprint_list") {
		input := &iam.UpdateOpenIDConnectProviderThumbprintInput{
			OpenIDConnectProviderArn: aws.String(d.Id()),
			ThumbprintList:           flex.ExpandStringValueList(d.Get("thumbprint_list").([]interface{})),
		}

		_, err := conn.UpdateOpenIDConnectProviderThumbprint(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM OIDC Provider (%s) thumbprint: %s", d.Id(), err)
		}
	}

	if d.HasChange("client_id_list") {
		o, n := d.GetChange("client_id_list")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		for _, v := range ns.Difference(os).List() {
			v := v.(string)
			input := &iam.AddClientIDToOpenIDConnectProviderInput{
				ClientID:                 aws.String(v),
				OpenIDConnectProviderArn: aws.String(d.Id()),
			}

			_, err := conn.AddClientIDToOpenIDConnectProvider(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "adding IAM OIDC Provider (%s) client ID (%s): %s", d.Id(), v, err)
			}
		}

		for _, v := range os.Difference(ns).List() {
			v := v.(string)
			input := &iam.RemoveClientIDFromOpenIDConnectProviderInput{
				ClientID:                 aws.String(v),
				OpenIDConnectProviderArn: aws.String(d.Id()),
			}

			_, err := conn.RemoveClientIDFromOpenIDConnectProvider(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "removing IAM OIDC Provider (%s) client ID (%s): %s", d.Id(), v, err)
			}
		}
	}

	return append(diags, resourceOpenIDConnectProviderRead(ctx, d, meta)...)
}

func resourceOpenIDConnectProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	log.Printf("[INFO] Deleting IAM OIDC Provider: %s", d.Id())
	_, err := conn.DeleteOpenIDConnectProvider(ctx, &iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM OIDC Provider (%s): %s", d.Id(), err)
	}

	return diags
}

func findOpenIDConnectProviderByARN(ctx context.Context, conn *iam.Client, arn string) (*iam.GetOpenIDConnectProviderOutput, error) {
	input := &iam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(arn),
	}

	output, err := conn.GetOpenIDConnectProvider(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
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

func openIDConnectProviderTags(ctx context.Context, conn *iam.Client, identifier string) ([]awstypes.Tag, error) {
	output, err := conn.ListOpenIDConnectProviderTags(ctx, &iam.ListOpenIDConnectProviderTagsInput{
		OpenIDConnectProviderArn: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}

	return output.Tags, nil
}
