package codestarnotifications

import (
	"context"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for target subscriptions to propagate
	targetSubscriptionTimeout = 30 * time.Second
)

func ResourceNotificationRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNotificationRuleCreate,
		ReadWithoutTimeout:   resourceNotificationRuleRead,
		UpdateWithoutTimeout: resourceNotificationRuleUpdate,
		DeleteWithoutTimeout: resourceNotificationRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"detail_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					codestarnotifications.DetailTypeBasic,
					codestarnotifications.DetailTypeFull,
				}, false),
			},

			"event_type_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
				MaxItems: 200,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9\-_ ]+$`), "must be one or more alphanumeric, hyphen, underscore or space characters"),
				),
			},

			"resource": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  codestarnotifications.NotificationRuleStatusEnabled,
				ValidateFunc: validation.StringInSlice([]string{
					codestarnotifications.NotificationRuleStatusEnabled,
					codestarnotifications.NotificationRuleStatusDisabled,
				}, false),
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"target": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"type": {
							Type:     schema.TypeString,
							Default:  "SNS",
							Optional: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func expandNotificationRuleTargets(targetsData []interface{}) []*codestarnotifications.Target {
	targets := make([]*codestarnotifications.Target, 0, len(targetsData))
	for _, t := range targetsData {
		target := t.(map[string]interface{})
		targets = append(targets, &codestarnotifications.Target{
			TargetAddress: aws.String(target["address"].(string)),
			TargetType:    aws.String(target["type"].(string)),
		})
	}
	return targets
}

func resourceNotificationRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarNotificationsConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	params := &codestarnotifications.CreateNotificationRuleInput{
		DetailType:   aws.String(d.Get("detail_type").(string)),
		EventTypeIds: flex.ExpandStringSet(d.Get("event_type_ids").(*schema.Set)),
		Name:         aws.String(d.Get("name").(string)),
		Resource:     aws.String(d.Get("resource").(string)),
		Status:       aws.String(d.Get("status").(string)),
		Targets:      expandNotificationRuleTargets(d.Get("target").(*schema.Set).List()),
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	res, err := conn.CreateNotificationRuleWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeStar Notification Rule: %s", err)
	}

	d.SetId(aws.StringValue(res.Arn))

	return append(diags, resourceNotificationRuleRead(ctx, d, meta)...)
}

func resourceNotificationRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarNotificationsConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rule, err := conn.DescribeNotificationRuleWithContext(ctx, &codestarnotifications.DescribeNotificationRuleInput{
		Arn: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codestarnotifications.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CodeStarNotifications, create.ErrActionReading, ResNotificationRule, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CodeStarNotifications, create.ErrActionReading, ResNotificationRule, d.Id(), err)
	}

	d.Set("arn", rule.Arn)
	d.Set("detail_type", rule.DetailType)
	eventTypeIds := make([]string, 0, len(rule.EventTypes))
	for _, et := range rule.EventTypes {
		eventTypeIds = append(eventTypeIds, aws.StringValue(et.EventTypeId))
	}
	if err := d.Set("event_type_ids", eventTypeIds); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting event_type_ids: %s", err)
	}
	d.Set("name", rule.Name)
	d.Set("status", rule.Status)
	d.Set("resource", rule.Resource)
	tags := tftags.New(rule.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	targets := make([]map[string]interface{}, 0, len(rule.Targets))
	for _, t := range rule.Targets {
		targets = append(targets, map[string]interface{}{
			"address": aws.StringValue(t.TargetAddress),
			"type":    aws.StringValue(t.TargetType),
			"status":  aws.StringValue(t.TargetStatus),
		})
	}
	if err = d.Set("target", targets); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting CodeStar notification target: %s", err)
	}

	return diags
}

const notificationRuleErrorSubscribed = "The target cannot be deleted because it is subscribed to one or more notification rules."

// cleanupNotificationRuleTargets tries to remove unused notification targets. AWS API does not
// provide expicit way for creating targets, they are created on first subscription. Here we are trying to remove all
// unused targets which were unsubscribed from this notification rule.
func cleanupNotificationRuleTargets(ctx context.Context, conn *codestarnotifications.CodeStarNotifications, oldVal *schema.Set, newVal *schema.Set) error {
	removedTargets := oldVal
	if newVal != nil {
		removedTargets = oldVal.Difference(newVal)
	}

	for _, targetRaw := range removedTargets.List() {
		target, ok := targetRaw.(map[string]interface{})

		if !ok {
			continue
		}

		input := &codestarnotifications.DeleteTargetInput{
			ForceUnsubscribeAll: aws.Bool(false),
			TargetAddress:       aws.String(target["address"].(string)),
		}

		err := resource.RetryContext(ctx, targetSubscriptionTimeout, func() *resource.RetryError {
			_, err := conn.DeleteTargetWithContext(ctx, input)

			if tfawserr.ErrMessageContains(err, codestarnotifications.ErrCodeValidationException, notificationRuleErrorSubscribed) {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.DeleteTargetWithContext(ctx, input)
		}

		// Treat target deletion as best effort
		if tfawserr.ErrMessageContains(err, codestarnotifications.ErrCodeValidationException, notificationRuleErrorSubscribed) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceNotificationRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarNotificationsConn()

	params := &codestarnotifications.UpdateNotificationRuleInput{
		Arn:          aws.String(d.Id()),
		DetailType:   aws.String(d.Get("detail_type").(string)),
		EventTypeIds: flex.ExpandStringSet(d.Get("event_type_ids").(*schema.Set)),
		Name:         aws.String(d.Get("name").(string)),
		Status:       aws.String(d.Get("status").(string)),
		Targets:      expandNotificationRuleTargets(d.Get("target").(*schema.Set).List()),
	}

	if _, err := conn.UpdateNotificationRuleWithContext(ctx, params); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CodeStar Notification Rule (%s): %s", d.Id(), err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeStar Notification Rule (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("target") {
		o, n := d.GetChange("target")
		if err := cleanupNotificationRuleTargets(ctx, conn, o.(*schema.Set), n.(*schema.Set)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeStar Notification Rule (%s): cleaning targets: %s", d.Id(), err)
		}
	}

	return append(diags, resourceNotificationRuleRead(ctx, d, meta)...)
}

func resourceNotificationRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarNotificationsConn()

	_, err := conn.DeleteNotificationRuleWithContext(ctx, &codestarnotifications.DeleteNotificationRuleInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeStar Notification Rule: %s", err)
	}

	if err = cleanupNotificationRuleTargets(ctx, conn, d.Get("target").(*schema.Set), nil); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeStar notification targets: %s", err)
	}

	return diags
}
