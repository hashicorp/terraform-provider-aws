// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_key_pair", name="Key Pair")
// @Tags
// @Testing(tagsTest=false)
func dataSourceKeyPair() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceKeyPairRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"include_public_key": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key_pair_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeKeyPairsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = newCustomFilterList(v.(*schema.Set))
	}

	if v, ok := d.GetOk("key_name"); ok {
		input.KeyNames = []string{v.(string)}
	}

	if v, ok := d.GetOk("key_pair_id"); ok {
		input.KeyPairIds = []string{v.(string)}
	}

	if v, ok := d.GetOk("include_public_key"); ok {
		input.IncludePublicKey = aws.Bool(v.(bool))
	}

	keyPair, err := findKeyPair(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Key Pair", err))
	}

	d.SetId(aws.ToString(keyPair.KeyPairId))
	keyName := aws.ToString(keyPair.KeyName)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ec2",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "key-pair/" + keyName,
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreateTime, aws.ToTime(keyPair.CreateTime).Format(time.RFC3339))
	d.Set("fingerprint", keyPair.KeyFingerprint)
	d.Set("include_public_key", input.IncludePublicKey)
	d.Set("key_name", keyName)
	d.Set("key_pair_id", keyPair.KeyPairId)
	d.Set("key_type", keyPair.KeyType)
	d.Set(names.AttrPublicKey, keyPair.PublicKey)

	setTagsOut(ctx, keyPair.Tags)

	return diags
}
