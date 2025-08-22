// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_directory_config", name="Directory Config")
func resourceDirectoryConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDirectoryConfigCreate,
		ReadWithoutTimeout:   resourceDirectoryConfigRead,
		UpdateWithoutTimeout: resourceDirectoryConfigUpdate,
		DeleteWithoutTimeout: resourceDirectoryConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"organizational_unit_distinguished_names": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 2000),
				},
			},
			"service_account_credentials": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"account_password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
	}
}

func resourceDirectoryConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	directoryName := d.Get("directory_name").(string)
	input := appstream.CreateDirectoryConfigInput{
		DirectoryName:                        aws.String(directoryName),
		OrganizationalUnitDistinguishedNames: flex.ExpandStringValueSet(d.Get("organizational_unit_distinguished_names").(*schema.Set)),
		ServiceAccountCredentials:            expandServiceAccountCredentials(d.Get("service_account_credentials").([]any)),
	}

	output, err := conn.CreateDirectoryConfig(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream Directory Config (%s): %s", directoryName, err)
	}

	d.SetId(aws.ToString(output.DirectoryConfig.DirectoryName))

	return append(diags, resourceDirectoryConfigRead(ctx, d, meta)...)
}

func resourceDirectoryConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	directoryConfig, err := findDirectoryConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppStream Directory Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream Directory Config (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCreatedTime, aws.ToTime(directoryConfig.CreatedTime).Format(time.RFC3339))
	d.Set("directory_name", directoryConfig.DirectoryName)
	d.Set("organizational_unit_distinguished_names", flex.FlattenStringValueSet(directoryConfig.OrganizationalUnitDistinguishedNames))
	if err = d.Set("service_account_credentials", flattenServiceAccountCredentials(directoryConfig.ServiceAccountCredentials, d)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_account_credentials: %s", err)
	}

	return diags
}

func resourceDirectoryConfigUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	input := appstream.UpdateDirectoryConfigInput{
		DirectoryName: aws.String(d.Id()),
	}

	if d.HasChange("organizational_unit_distinguished_names") {
		input.OrganizationalUnitDistinguishedNames = flex.ExpandStringValueSet(d.Get("organizational_unit_distinguished_names").(*schema.Set))
	}

	if d.HasChange("service_account_credentials") {
		input.ServiceAccountCredentials = expandServiceAccountCredentials(d.Get("service_account_credentials").([]any))
	}

	_, err := conn.UpdateDirectoryConfig(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppStream Directory Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceDirectoryConfigRead(ctx, d, meta)...)
}

func resourceDirectoryConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	log.Printf("[DEBUG] Deleting AppStream Directory Config: %s", d.Id())
	input := appstream.DeleteDirectoryConfigInput{
		DirectoryName: aws.String(d.Id()),
	}
	_, err := conn.DeleteDirectoryConfig(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream Directory Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findDirectoryConfigByID(ctx context.Context, conn *appstream.Client, id string) (*awstypes.DirectoryConfig, error) {
	input := appstream.DescribeDirectoryConfigsInput{
		DirectoryNames: []string{id},
	}

	return findDirectoryConfig(ctx, conn, &input)
}

func findDirectoryConfig(ctx context.Context, conn *appstream.Client, input *appstream.DescribeDirectoryConfigsInput) (*awstypes.DirectoryConfig, error) {
	output, err := findDirectoryConfigs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDirectoryConfigs(ctx context.Context, conn *appstream.Client, input *appstream.DescribeDirectoryConfigsInput) ([]awstypes.DirectoryConfig, error) {
	var output []awstypes.DirectoryConfig

	err := describeDirectoryConfigsPages(ctx, conn, input, func(page *appstream.DescribeDirectoryConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.DirectoryConfigs...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func expandServiceAccountCredentials(tfList []any) *awstypes.ServiceAccountCredentials {
	if len(tfList) == 0 {
		return nil
	}

	attr := tfList[0].(map[string]any)

	apiObject := &awstypes.ServiceAccountCredentials{
		AccountName:     aws.String(attr["account_name"].(string)),
		AccountPassword: aws.String(attr["account_password"].(string)),
	}

	return apiObject
}

func flattenServiceAccountCredentials(apiObject *awstypes.ServiceAccountCredentials, d *schema.ResourceData) []any {
	if apiObject == nil {
		return nil
	}

	tfList := map[string]any{}
	tfList["account_name"] = aws.ToString(apiObject.AccountName)
	tfList["account_password"] = d.Get("service_account_credentials.0.account_password").(string)

	return []any{tfList}
}
