// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serverlessrepo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_serverlessapplicationrepository_application")
func DataSourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceApplicationRead,

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"semantic_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"required_capabilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"source_code_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServerlessRepoClient(ctx)

	applicationID := d.Get(names.AttrApplicationID).(string)
	semanticVersion := d.Get("semantic_version").(string)

	output, err := findApplication(ctx, conn, applicationID, semanticVersion)
	if err != nil {
		descriptor := applicationID
		if semanticVersion != "" {
			descriptor += fmt.Sprintf(", version %s", semanticVersion)
		}
		return sdkdiag.AppendErrorf(diags, "getting Serverless Application Repository application (%s): %s", descriptor, err)
	}

	d.SetId(applicationID)
	d.Set(names.AttrName, output.Name)
	d.Set("semantic_version", output.Version.SemanticVersion)
	d.Set("source_code_url", output.Version.SourceCodeUrl)
	d.Set("template_url", output.Version.TemplateUrl)
	if err = d.Set("required_capabilities", flex.FlattenStringyValueSet(output.Version.RequiredCapabilities)); err != nil {
		return sdkdiag.AppendErrorf(diags, "to set required_capabilities: %s", err)
	}

	return diags
}
