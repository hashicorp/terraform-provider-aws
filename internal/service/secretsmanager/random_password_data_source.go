// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_secretsmanager_random_password")
func DataSourceRandomPassword() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRandomPasswordRead,

		Schema: map[string]*schema.Schema{
			"exclude_characters": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"exclude_lowercase": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"exclude_numbers": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"exclude_punctuation": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"exclude_uppercase": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"include_space": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"password_length": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  32,
			},
			"require_each_included_type": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"random_password": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceRandomPasswordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	var excludeCharacters string
	if v, ok := d.GetOk("exclude_characters"); ok {
		excludeCharacters = v.(string)
	}
	var excludeLowercase bool
	if v, ok := d.GetOk("exclude_lowercase"); ok {
		excludeLowercase = v.(bool)
	}
	var excludeNumbers bool
	if v, ok := d.GetOk("exclude_numbers"); ok {
		excludeNumbers = v.(bool)
	}
	var excludePunctuation bool
	if v, ok := d.GetOk("exclude_punctuation"); ok {
		excludePunctuation = v.(bool)
	}
	var excludeUppercase bool
	if v, ok := d.GetOk("exclude_uppercase"); ok {
		excludeUppercase = v.(bool)
	}
	var includeSpace bool
	if v, ok := d.GetOk("exclude_space"); ok {
		includeSpace = v.(bool)
	}
	var passwordLength int64
	if v, ok := d.GetOk("password_length"); ok {
		passwordLength = int64(v.(int))
	}
	var requireEachIncludedType bool
	if v, ok := d.GetOk("require_each_included_type"); ok {
		requireEachIncludedType = v.(bool)
	}

	input := &secretsmanager.GetRandomPasswordInput{
		ExcludeCharacters:       aws.String(excludeCharacters),
		ExcludeLowercase:        aws.Bool(excludeLowercase),
		ExcludeNumbers:          aws.Bool(excludeNumbers),
		ExcludePunctuation:      aws.Bool(excludePunctuation),
		ExcludeUppercase:        aws.Bool(excludeUppercase),
		IncludeSpace:            aws.Bool(includeSpace),
		PasswordLength:          aws.Int64(passwordLength),
		RequireEachIncludedType: aws.Bool(requireEachIncludedType),
	}

	output, err := conn.GetRandomPasswordWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Get Random Password: %s", err)
	}

	d.SetId(aws.StringValue(output.RandomPassword))
	d.Set("random_password", output.RandomPassword)

	return diags
}
