package cloudtrail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCloudTrailLake() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudTrailLakeCreate,
		ReadContext:   resourceCloudTrailLakeRead,
		UpdateContext: resourceCloudTrailLakeUpdate,
		DeleteContext: resourceCloudTrailLakeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(snapshotAvailableTimeout),
			Delete: schema.DefaultTimeout(snapshotDeletedTimeout),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"advanced_event_selector": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"event_selector"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_selector": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ends_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"field": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(field_Values(), false),
									},
									"not_ends_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"not_equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"not_starts_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
									"starts_with": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 2048),
										},
									},
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1000),
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"multi_region_enabled": {
				Type:         schema.TypeBool,
				Optional:     true,
				Default:      false,
				ValidateFunc: verify.ValidARN,
			},
			"organization_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"retention_period": {
				Type:     schema.TypeInt,
				Optional: false,
				ValidateFunc: validation.All(
					validation.IntBetween(7, 2555),
				),
			},
			"termination_protection_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCloudTrailLakeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudTrailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &cloudtrail.CreateEventDataStoreInput{
		AdvancedEventSelectors: []*cloudtrail.AdvancedEventSelector{},
		Name:                   aws.String(name),
		RetentionPeriod:        aws.Int64(d.Get("retention_period").(int64)),
	}

	if len(tags) > 0 {
		input.TagsList = Tags(tags.IgnoreAWS())
	}
	if v, ok := d.GetOk("organization_enabled"); ok {
		input.OrganizationEnabled = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("multi_region_enabled"); ok {
		input.MultiRegionEnabled = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("termination_protection_enabled"); ok {
		input.TerminationProtectionEnabled = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Creating Event Data Store: %s", input)
	_, err := conn.CreateEventDataStoreWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Event Data Store (%s): %s", name, err)
	}

	if err := waitEventDataStoreAvailable(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Event Data Store (%s) to be created: %s", name, err)
	}

	d.SetId(name)

	return resouceCloudTrailLakeRead(ctx, d, meta)
}
