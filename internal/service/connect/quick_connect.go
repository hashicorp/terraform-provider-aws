package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceQuickConnect() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceQuickConnectCreate,
		ReadContext:   resourceQuickConnectRead,
		UpdateContext: resourceQuickConnectUpdate,
		DeleteContext: resourceQuickConnectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(connectQuickConnectCreatedTimeout),
			Delete: schema.DefaultTimeout(connectQuickConnectDeletedTimeout),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			"quick_connect_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"quick_connect_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 127),
			},
			"quick_connect_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"phone_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"phone_number": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v := d.Get("quick_connect_config.0.quick_connect_type").(string); v == connect.QuickConnectTypePhoneNumber {
									return false
								}
								return true
							},
						},
						"queue_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_flow_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"queue_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v := d.Get("quick_connect_config.0.quick_connect_type").(string); v == connect.QuickConnectTypeQueue {
									return false
								}
								return true
							},
						},
						"quick_connect_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(connect.QuickConnectType_Values(), false),
						},
						"user_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_flow_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v := d.Get("quick_connect_config.0.quick_connect_type").(string); v == connect.QuickConnectTypeUser {
									return false
								}
								return true
							},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}
