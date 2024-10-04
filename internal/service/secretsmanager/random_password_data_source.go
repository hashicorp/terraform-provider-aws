// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_secretsmanager_random_password", name="Random Password")
func dataSourceRandomPassword() *schema.Resource {
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
				Computed: true,
			},
		},
	}
}

func dataSourceRandomPasswordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	input := &secretsmanager.GetRandomPasswordInput{
		ExcludeLowercase:        aws.Bool(d.Get("exclude_lowercase").(bool)),
		ExcludeNumbers:          aws.Bool(d.Get("exclude_numbers").(bool)),
		ExcludePunctuation:      aws.Bool(d.Get("exclude_punctuation").(bool)),
		ExcludeUppercase:        aws.Bool(d.Get("exclude_uppercase").(bool)),
		IncludeSpace:            aws.Bool(d.Get("include_space").(bool)),
		PasswordLength:          aws.Int64(int64(d.Get("password_length").(int))),
		RequireEachIncludedType: aws.Bool(d.Get("require_each_included_type").(bool)),
	}

	if v, ok := d.GetOk("exclude_characters"); ok {
		input.ExcludeCharacters = aws.String(v.(string))
	}

	output, err := conn.GetRandomPassword(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Random Password: %s", err)
	}

	password := aws.ToString(output.RandomPassword)
	d.SetId(password)
	d.Set("random_password", password)

	return diags
}
