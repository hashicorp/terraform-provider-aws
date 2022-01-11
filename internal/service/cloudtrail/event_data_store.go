package cloudtrail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEventDataStore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEventDataStoreCreate,
		ReadContext:   resourceEventDataStoreRead,
		UpdateContext: resourceEventDataStoreUpdate,
		DeleteContext: resourceEventDataStoreDelete,

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
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
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

func resourceEventDataStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("error creating CloudTrail Event Data Store (%s): %s", name, err)
	}

	if err := waitEventDataStoreAvailable(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for CloudTrail Event Data Store (%s) to be created: %s", name, err)
	}

	d.SetId(name)

	return resourceEventDataStoreRead(ctx, d, meta)
}

func resourceEventDataStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudTrailConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating CloudTrail Event Data Store (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceEventDataStoreRead(ctx, d, meta)
}

func resourceEventDataStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudTrailConn

	log.Printf("[DEBUG] Deleting Event Data Store: (%s)", d.Id())
	_, err := conn.DeleteEventDataStoreWithContext(ctx, &cloudtrail.DeleteEventDataStoreInput{
		EventDataStore: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudtrail.ErrCodeEventDataStoreNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Event Data Store (%s): %s", d.Id(), err)
	}

	if err := waitEventDataStoreDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for CloudTrail Event Data Store (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}
