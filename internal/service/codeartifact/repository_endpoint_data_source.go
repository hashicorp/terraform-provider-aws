// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_codeartifact_repository_endpoint", name="Repository Endpoint")
func dataSourceRepositoryEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRepositoryEndpointRead,

		Schema: map[string]*schema.Schema{
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrFormat: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.PackageFormat](),
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRepositoryEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)
	var domainOwner string
	if v, ok := d.GetOk("domain_owner"); ok {
		domainOwner = v.(string)
	} else {
		domainOwner = meta.(*conns.AWSClient).AccountID
	}
	format := types.PackageFormat(d.Get(names.AttrFormat).(string))
	repositoryName := d.Get("repository").(string)
	input := &codeartifact.GetRepositoryEndpointInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
		Format:      format,
		Repository:  aws.String(repositoryName),
	}

	output, err := conn.GetRepositoryEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeArtifact Repository Endpoint: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s:%s", domainOwner, domainName, repositoryName, format))
	d.Set("domain_owner", domainOwner)
	d.Set("repository_endpoint", output.RepositoryEndpoint)

	return diags
}
