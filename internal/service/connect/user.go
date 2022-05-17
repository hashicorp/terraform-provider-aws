package connect

import (
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"hierarchy_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"identity_info": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"first_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 100),
						},
						"last_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 100),
						},
					},
				},
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 64),
			},
			"phone_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after_contact_work_time_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"auto_accept": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"desk_phone_number": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validDeskPhoneNumber,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v := d.Get("phone_config.0.phone_type").(string); v == connect.PhoneTypeDeskPhone {
									return false
								}
								return true
							},
						},
						"phone_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(connect.PhoneType_Values(), false),
						},
					},
				},
			},
			"routing_profile_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"security_profile_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
