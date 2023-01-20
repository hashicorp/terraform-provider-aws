package licensemanager

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
					ValidateFunc: validation.StringMatch(regexp.MustCompile("^#([^=]+)=(.+)$"), "Expected format is #RuleType=RuleValue"),
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLicenseConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LicenseManagerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &licensemanager.CreateLicenseConfigurationInput{
		LicenseCountingType: aws.String(d.Get("license_counting_type").(string)),
		Name:                aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
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

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating License Manager License Configuration: %s", input)
	resp, err := conn.CreateLicenseConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating License Manager License Configuration (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(resp.LicenseConfigurationArn))

	return resourceLicenseConfigurationRead(ctx, d, meta)
}

func resourceLicenseConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LicenseManagerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindLicenseConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] License Manager License Configuration %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading License Manager License Configuration (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.LicenseConfigurationArn)
	d.Set("description", output.Description)
	d.Set("license_count", output.LicenseCount)
	d.Set("license_count_hard_limit", output.LicenseCountHardLimit)
	d.Set("license_counting_type", output.LicenseCountingType)
	d.Set("license_rules", aws.StringValueSlice(output.LicenseRules))
	d.Set("name", output.Name)
	d.Set("owner_account_id", output.OwnerAccountId)

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceLicenseConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LicenseManagerConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &licensemanager.UpdateLicenseConfigurationInput{
			Description:             aws.String(d.Get("description").(string)),
			LicenseConfigurationArn: aws.String(d.Id()),
			LicenseCountHardLimit:   aws.Bool(d.Get("license_count_hard_limit").(bool)),
			Name:                    aws.String(d.Get("name").(string)),
		}

		if v, ok := d.GetOk("license_count"); ok {
			input.LicenseCount = aws.Int64(int64(v.(int)))
		}

		log.Printf("[DEBUG] Updating License Manager License Configuration: %s", input)
		_, err := conn.UpdateLicenseConfigurationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating License Manager License Configuration (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating License Manager License Configuration (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceLicenseConfigurationRead(ctx, d, meta)
}

func resourceLicenseConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LicenseManagerConn()

	log.Printf("[DEBUG] Deleting License Manager License Configuration: %s", d.Id())
	_, err := conn.DeleteLicenseConfigurationWithContext(ctx, &licensemanager.DeleteLicenseConfigurationInput{
		LicenseConfigurationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, licensemanager.ErrCodeInvalidParameterValueException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting License Manager License Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func FindLicenseConfigurationByARN(ctx context.Context, conn *licensemanager.LicenseManager, arn string) (*licensemanager.GetLicenseConfigurationOutput, error) {
	input := &licensemanager.GetLicenseConfigurationInput{
		LicenseConfigurationArn: aws.String(arn),
	}

	output, err := conn.GetLicenseConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, licensemanager.ErrCodeInvalidParameterValueException) {
		return nil, &resource.NotFoundError{
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
