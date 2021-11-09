package appstream

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDirectoryConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDirectoryConfigCreate,
		ReadWithoutTimeout:   resourceDirectoryConfigRead,
		UpdateWithoutTimeout: resourceDirectoryConfigUpdate,
		DeleteWithoutTimeout: resourceDirectoryConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"created_time": {
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
				Set: schema.HashString,
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

func resourceDirectoryConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn
	input := &appstream.CreateDirectoryConfigInput{
		DirectoryName:                        aws.String(d.Get("directory_name").(string)),
		OrganizationalUnitDistinguishedNames: flex.ExpandStringSet(d.Get("organizational_unit_distinguished_names").(*schema.Set)),
		ServiceAccountCredentials:            expandServiceAccountCredentials(d.Get("service_account_credentials").([]interface{})),
	}

	var output *appstream.CreateDirectoryConfigOutput
	err := resource.RetryContext(ctx, directoryConfigTimeout, func() *resource.RetryError {
		out, err := conn.CreateDirectoryConfigWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		output = out

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateDirectoryConfigWithContext(ctx, input)
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Appstream DirectoryConfig (%s): %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.DirectoryConfig.DirectoryName))

	return resourceDirectoryConfigRead(ctx, d, meta)
}

func resourceDirectoryConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	resp, err := conn.DescribeDirectoryConfigsWithContext(ctx, &appstream.DescribeDirectoryConfigsInput{DirectoryNames: []*string{aws.String(d.Id())}})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appstream DirectoryConfig (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream DirectoryConfig (%s): %w", d.Id(), err))
	}

	if len(resp.DirectoryConfigs) == 0 {
		return diag.FromErr(fmt.Errorf("error reading Appstream DirectoryConfig (%s): %s", d.Id(), "empty response"))
	}

	if len(resp.DirectoryConfigs) > 1 {
		return diag.FromErr(fmt.Errorf("error reading Appstream DirectoryConfig (%s): %s", d.Id(), "multiple directories config found"))
	}

	directoryConfig := resp.DirectoryConfigs[0]

	d.Set("created_time", aws.TimeValue(directoryConfig.CreatedTime).Format(time.RFC3339))
	d.Set("directory_name", directoryConfig.DirectoryName)
	d.Set("organizational_unit_distinguished_names", flex.FlattenStringSet(directoryConfig.OrganizationalUnitDistinguishedNames))

	if err = d.Set("service_account_credentials", flattenServiceAccountCredentials(directoryConfig.ServiceAccountCredentials, d)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream DirectoryConfig (%s): %w", "service_account_credentials", d.Id(), err))
	}

	return nil
}

func resourceDirectoryConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn
	input := &appstream.UpdateDirectoryConfigInput{
		DirectoryName: aws.String(d.Id()),
	}

	if d.HasChange("organizational_unit_distinguished_names") {
		input.OrganizationalUnitDistinguishedNames = flex.ExpandStringSet(d.Get("organizational_unit_distinguished_names").(*schema.Set))
	}

	if d.HasChange("service_account_credentials") {
		input.ServiceAccountCredentials = expandServiceAccountCredentials(d.Get("service_account_credentials").([]interface{}))
	}

	_, err := conn.UpdateDirectoryConfigWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Appstream DirectoryConfig (%s): %w", d.Id(), err))
	}

	return resourceDirectoryConfigRead(ctx, d, meta)
}

func resourceDirectoryConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppStreamConn

	_, err := conn.DeleteDirectoryConfigWithContext(ctx, &appstream.DeleteDirectoryConfigInput{
		DirectoryName: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Appstream DirectoryConfig (%s): %w", d.Id(), err))
	}
	return nil
}

func expandServiceAccountCredentials(tfList []interface{}) *appstream.ServiceAccountCredentials {
	if len(tfList) == 0 {
		return nil
	}

	attr := tfList[0].(map[string]interface{})

	apiObject := &appstream.ServiceAccountCredentials{
		AccountName:     aws.String(attr["account_name"].(string)),
		AccountPassword: aws.String(attr["account_password"].(string)),
	}

	return apiObject
}

func flattenServiceAccountCredentials(apiObject *appstream.ServiceAccountCredentials, d *schema.ResourceData) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := map[string]interface{}{}
	tfList["account_name"] = aws.StringValue(apiObject.AccountName)
	tfList["account_password"] = d.Get("service_account_credentials.0.account_password").(string)

	return []interface{}{tfList}
}
