// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_codecommit_repository", name="Repository")
func dataSourceRepository() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRepositoryRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"clone_url_http": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"clone_url_ssh": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRepositoryName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
		},
	}
}

func dataSourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	name := d.Get(names.AttrRepositoryName).(string)
	repository, err := findRepositoryByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Repository (%s): %s", name, err)
	}

	d.SetId(aws.ToString(repository.RepositoryName))
	d.Set(names.AttrARN, repository.Arn)
	d.Set("clone_url_http", repository.CloneUrlHttp)
	d.Set("clone_url_ssh", repository.CloneUrlSsh)
	d.Set(names.AttrKMSKeyID, repository.KmsKeyId)
	d.Set("repository_id", repository.RepositoryId)
	d.Set(names.AttrRepositoryName, repository.RepositoryName)

	return diags
}
