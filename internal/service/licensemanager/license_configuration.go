// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_licensemanager_license_configuration", name="License Configuration")
// @Tags(identifierAttribute="id")
func ResourceLicenseConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLicenseConfigurationCreate,
		ReadWithoutTimeout:   resourceLicenseConfigurationRead,
		UpdateWithoutTimeout: resourceLicenseConfigurationUpdate,
		DeleteWithoutTimeout: resourceLicenseConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"license_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"license_count_hard_limit": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"license_counting_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(licensemanager.LicenseCountingType_Values(), false),
			},
			"license_rules": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexache.MustCompile("^#([^=]+)=(.+)$"), "Expected format is #RuleType=RuleValue"),
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLicenseConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &licensemanager.CreateLicenseConfigurationInput{
		LicenseCountingType: aws.String(d.Get("license_counting_type").(string)),
		Name:                aws.String(name),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_count"); ok {
		input.LicenseCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("license_count_hard_limit"); ok {
		input.LicenseCountHardLimit = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("license_rules"); ok && len(v.([]interface{})) > 0 {
		input.LicenseRules = flex.ExpandStringList(v.([]interface{}))
	}

	output, err := conn.CreateLicenseConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating License Manager License Configuration (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.LicenseConfigurationArn))

	return append(diags, resourceLicenseConfigurationRead(ctx, d, meta)...)
}

func resourceLicenseConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	output, err := FindLicenseConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] License Manager License Configuration %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager License Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.LicenseConfigurationArn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("license_count", output.LicenseCount)
	d.Set("license_count_hard_limit", output.LicenseCountHardLimit)
	d.Set("license_counting_type", output.LicenseCountingType)
	d.Set("license_rules", aws.StringValueSlice(output.LicenseRules))
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrOwnerAccountID, output.OwnerAccountId)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceLicenseConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &licensemanager.UpdateLicenseConfigurationInput{
			Description:             aws.String(d.Get(names.AttrDescription).(string)),
			LicenseConfigurationArn: aws.String(d.Id()),
			LicenseCountHardLimit:   aws.Bool(d.Get("license_count_hard_limit").(bool)),
			Name:                    aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk("license_count"); ok {
			input.LicenseCount = aws.Int64(int64(v.(int)))
		}

		_, err := conn.UpdateLicenseConfigurationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating License Manager License Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLicenseConfigurationRead(ctx, d, meta)...)
}

func resourceLicenseConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	log.Printf("[DEBUG] Deleting License Manager License Configuration: %s", d.Id())
	_, err := conn.DeleteLicenseConfigurationWithContext(ctx, &licensemanager.DeleteLicenseConfigurationInput{
		LicenseConfigurationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, licensemanager.ErrCodeInvalidParameterValueException, "Invalid license configuration ARN") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting License Manager License Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func FindLicenseConfigurationByARN(ctx context.Context, conn *licensemanager.LicenseManager, arn string) (*licensemanager.GetLicenseConfigurationOutput, error) {
	input := &licensemanager.GetLicenseConfigurationInput{
		LicenseConfigurationArn: aws.String(arn),
	}

	output, err := conn.GetLicenseConfigurationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, licensemanager.ErrCodeInvalidParameterValueException, "Invalid license configuration ARN") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
