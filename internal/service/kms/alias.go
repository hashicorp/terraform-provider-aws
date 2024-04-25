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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_kms_alias")
func ResourceAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAliasCreate,
		ReadWithoutTimeout:   resourceAliasRead,
		UpdateWithoutTimeout: resourceAliasUpdate,
		DeleteWithoutTimeout: resourceAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validNameForResource,
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
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

	namePrefix := d.Get("name_prefix").(string)
	if namePrefix == "" {
		namePrefix = AliasNamePrefix
	}
	name := create.Name(d.Get("name").(string), namePrefix)

	input := &kms.CreateAliasInput{
		AliasName:   aws.String(name),
		TargetKeyId: aws.String(d.Get("target_key_id").(string)),
	}

	// KMS is eventually consistent.
	log.Printf("[DEBUG] Creating KMS Alias: %v", input)

	var NotFoundException = &awstypes.NotFoundException{}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, KeyRotationUpdatedTimeout, func() (interface{}, error) {
		return conn.CreateAlias(ctx, input)
	}, NotFoundException.ErrorCode())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Alias (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		return FindAliasByName(ctx, conn, d.Id())
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
	targetKeyARN, err := AliasARNToKeyARN(aliasARN, targetKeyID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Alias (%s): %s", d.Id(), err)
	}

	d.Set("arn", aliasARN)
	d.Set("name", alias.AliasName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.ToString(alias.AliasName)))
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

		log.Printf("[DEBUG] Updating KMS Alias: %v", input)
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

	log.Printf("[DEBUG] Deleting KMS Alias: (%s)", d.Id())
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

func suppressEquivalentKeyARNOrID(k, old, new string, d *schema.ResourceData) bool {
	return KeyARNOrIDEqual(old, new)
}
