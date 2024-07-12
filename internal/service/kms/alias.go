// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kms_alias", name="Alias")
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
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validNameForResource,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validNameForResource,
			},
			"target_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_key_id": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentKeyARNOrID,
			},
		},
	}
}

func resourceAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	namePrefix := d.Get(names.AttrNamePrefix).(string)
	if namePrefix == "" {
		namePrefix = aliasNamePrefix
	}
	name := create.Name(d.Get(names.AttrName).(string), namePrefix)
	input := &kms.CreateAliasInput{
		AliasName:   aws.String(name),
		TargetKeyId: aws.String(d.Get("target_key_id").(string)),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.NotFoundException](ctx, keyRotationUpdatedTimeout, func() (interface{}, error) {
		return conn.CreateAlias(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Alias (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findAliasByName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Alias (%s): %s", d.Id(), err)
	}

	alias := outputRaw.(*awstypes.AliasListEntry)
	aliasARN := aws.ToString(alias.AliasArn)
	targetKeyID := aws.ToString(alias.TargetKeyId)
	targetKeyARN, err := aliasARNToKeyARN(aliasARN, targetKeyID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Alias (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, aliasARN)
	d.Set(names.AttrName, alias.AliasName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(alias.AliasName)))
	d.Set("target_key_arn", targetKeyARN)
	d.Set("target_key_id", targetKeyID)

	return diags
}

func resourceAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	if d.HasChange("target_key_id") {
		input := &kms.UpdateAliasInput{
			AliasName:   aws.String(d.Id()),
			TargetKeyId: aws.String(d.Get("target_key_id").(string)),
		}

		_, err := conn.UpdateAlias(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating KMS Alias (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	log.Printf("[DEBUG] Deleting KMS Alias: %s", d.Id())
	_, err := conn.DeleteAlias(ctx, &kms.DeleteAliasInput{
		AliasName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting KMS Alias (%s): %s", d.Id(), err)
	}

	return diags
}

func findAliasByName(ctx context.Context, conn *kms.Client, name string) (*awstypes.AliasListEntry, error) {
	input := &kms.ListAliasesInput{}

	return findAlias(ctx, conn, input, func(v *awstypes.AliasListEntry) bool {
		return aws.ToString(v.AliasName) == name
	})
}

func findAlias(ctx context.Context, conn *kms.Client, input *kms.ListAliasesInput, filter tfslices.Predicate[*awstypes.AliasListEntry]) (*awstypes.AliasListEntry, error) {
	output, err := findAliases(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAliases(ctx context.Context, conn *kms.Client, input *kms.ListAliasesInput, filter tfslices.Predicate[*awstypes.AliasListEntry]) ([]awstypes.AliasListEntry, error) {
	var output []awstypes.AliasListEntry

	pages := kms.NewListAliasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return output, err
		}

		for _, v := range page.Aliases {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func suppressEquivalentKeyARNOrID(k, old, new string, d *schema.ResourceData) bool {
	return keyARNOrIDEqual(old, new)
}
