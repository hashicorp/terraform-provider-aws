package sesv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sesv2_email_identity")
// @Tags(identifierAttribute="arn")
func DataSourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEmailIdentityRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_set_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"dkim_signing_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"current_signing_key_length": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_signing_private_key": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"dkim_signing_attributes.0.domain_signing_selector"},
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 20480),
								func(v interface{}, name string) (warns []string, errs []error) {
									s := v.(string)
									if !verify.IsBase64Encoded([]byte(s)) {
										errs = append(errs, fmt.Errorf(
											"%s: must be base64-encoded", name,
										))
									}
									return
								},
							),
						},
						"domain_signing_selector": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"dkim_signing_attributes.0.domain_signing_private_key"},
							ValidateFunc: validation.StringLenBetween(1, 63),
						},
						"last_key_generation_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"next_signing_key_length": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ConflictsWith:    []string{"dkim_signing_attributes.0.domain_signing_private_key", "dkim_signing_attributes.0.domain_signing_selector"},
							ValidateDiagFunc: enum.Validate[types.DkimSigningKeyLength](),
						},
						"signing_attributes_origin": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tokens": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"email_identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"identity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchema(),
			"verified_for_sending_status": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

const (
	DSNameEmailIdentity = "Email Identity Data Source"
)

func dataSourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	name := d.Get("email_identity").(string)

	out, err := FindEmailIdentityByID(ctx, conn, name)
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, DSNameEmailIdentity, name, err)
	}

	arn := emailIdentityNameToARN(meta, name)

	d.SetId(name)
	d.Set("arn", arn)
	d.Set("configuration_set_name", out.ConfigurationSetName)
	d.Set("email_identity", name)

	if out.DkimAttributes != nil {
		tfMap := flattenDKIMAttributes(out.DkimAttributes)
		tfMap["domain_signing_private_key"] = d.Get("dkim_signing_attributes.0.domain_signing_private_key").(string)
		tfMap["domain_signing_selector"] = d.Get("dkim_signing_attributes.0.domain_signing_selector").(string)

		if err := d.Set("dkim_signing_attributes", []interface{}{tfMap}); err != nil {
			return create.DiagError(names.SESV2, create.ErrActionSetting, ResNameEmailIdentity, name, err)
		}
	} else {
		d.Set("dkim_signing_attributes", nil)
	}

	d.Set("identity_type", string(out.IdentityType))
	d.Set("verified_for_sending_status", out.VerifiedForSendingStatus)

	return nil
}
