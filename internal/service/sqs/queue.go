// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sqs

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	queueSchema = map[string]*schema.Schema{
		names.AttrARN: {
			Type:     schema.TypeString,
			Computed: true,
		},
		"content_based_deduplication": {
			Type:     schema.TypeBool,
			Default:  false,
			Optional: true,
		},
		"deduplication_scope": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice(deduplicationScope_Values(), false),
		},
		"delay_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      defaultQueueDelaySeconds,
			ValidateFunc: validation.IntBetween(0, 900),
		},
		"fifo_queue": {
			Type:     schema.TypeBool,
			Default:  false,
			ForceNew: true,
			Optional: true,
		},
		"fifo_throughput_limit": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice(fifoThroughputLimit_Values(), false),
		},
		"kms_data_key_reuse_period_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(60, 86_400),
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// Only valid for encrypted queues, not returned by SQS
				return !d.Get("sqs_managed_sse_enabled").(bool) && d.Get("kms_master_key_id").(string) == ""
			},
		},
		"kms_master_key_id": {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"sqs_managed_sse_enabled"},
		},
		"max_message_size": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      defaultQueueMaximumMessageSize,
			ValidateFunc: validation.IntBetween(1024, 1_048_576),
		},
		"message_retention_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      defaultQueueMessageRetentionPeriod,
			ValidateFunc: validation.IntBetween(60, 1_209_600),
		},
		names.AttrName: {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{names.AttrNamePrefix},
		},
		names.AttrNamePrefix: {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{names.AttrName},
		},
		names.AttrPolicy: {
			Type:                  schema.TypeString,
			Optional:              true,
			Computed:              true,
			ValidateFunc:          validation.StringIsJSON,
			DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
			DiffSuppressOnRefresh: true,
			StateFunc: func(v any) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		"receive_wait_time_seconds": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  defaultQueueReceiveMessageWaitTimeSeconds,
		},
		"redrive_allow_policy": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringIsJSON,
			StateFunc: func(v any) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		"redrive_policy": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringIsJSON,
			StateFunc: func(v any) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		"sqs_managed_sse_enabled": {
			Type:          schema.TypeBool,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"kms_master_key_id"},
		},
		names.AttrTags:    tftags.TagsSchema(),
		names.AttrTagsAll: tftags.TagsSchemaComputed(),
		names.AttrURL: {
			Type:     schema.TypeString,
			Computed: true,
		},
		"visibility_timeout_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      defaultQueueVisibilityTimeout,
			ValidateFunc: validation.IntBetween(0, 43_200),
		},
	}

	queueAttributeMap = attrmap.New(map[string]types.QueueAttributeName{
		names.AttrARN:                       types.QueueAttributeNameQueueArn,
		"content_based_deduplication":       types.QueueAttributeNameContentBasedDeduplication,
		"deduplication_scope":               types.QueueAttributeNameDeduplicationScope,
		"delay_seconds":                     types.QueueAttributeNameDelaySeconds,
		"fifo_queue":                        types.QueueAttributeNameFifoQueue,
		"fifo_throughput_limit":             types.QueueAttributeNameFifoThroughputLimit,
		"kms_data_key_reuse_period_seconds": types.QueueAttributeNameKmsDataKeyReusePeriodSeconds,
		"kms_master_key_id":                 types.QueueAttributeNameKmsMasterKeyId,
		"max_message_size":                  types.QueueAttributeNameMaximumMessageSize,
		"message_retention_seconds":         types.QueueAttributeNameMessageRetentionPeriod,
		names.AttrPolicy:                    types.QueueAttributeNamePolicy,
		"receive_wait_time_seconds":         types.QueueAttributeNameReceiveMessageWaitTimeSeconds,
		"redrive_allow_policy":              types.QueueAttributeNameRedriveAllowPolicy,
		"redrive_policy":                    types.QueueAttributeNameRedrivePolicy,
		"sqs_managed_sse_enabled":           types.QueueAttributeNameSqsManagedSseEnabled,
		"visibility_timeout_seconds":        types.QueueAttributeNameVisibilityTimeout,
	}, queueSchema).WithIAMPolicyAttribute(names.AttrPolicy).WithMissingSetToNil("*").WithAlwaysSendConfiguredBooleanValueOnCreate("sqs_managed_sse_enabled")
)

// @SDKResource("aws_sqs_queue", name="Queue")
// @Tags(identifierAttribute="id")
// @IdentityVersion(1)
// @CustomInherentRegionIdentity("url", "parseQueueURL")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sqs/types;awstypes;map[awstypes.QueueAttributeName]string")
// @Testing(preIdentityVersion="v6.9.0")
// @Testing(identityVersion="0;v6.10.0")
// @Testing(identityVersion="1;v6.19.0")
// @Testing(existsTakesT=false, destroyTakesT=false)
func resourceQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueueCreate,
		ReadWithoutTimeout:   resourceQueueRead,
		UpdateWithoutTimeout: resourceQueueUpdate,
		DeleteWithoutTimeout: resourceQueueDelete,

		CustomizeDiff: resourceQueueCustomizeDiff,

		Schema: queueSchema,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
	}
}

func resourceQueueCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SQSClient(ctx)

	name := queueName(ctx, d)
	input := &sqs.CreateQueueInput{
		QueueName: aws.String(name),
		Tags:      getTagsIn(ctx),
	}

	attributes, err := queueAttributeMap.ResourceDataToAPIAttributesCreate(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input.Attributes = flex.ExpandStringyValueMap(attributes)

	// create is 2 phase: 1. create, 2. wait for propagation
	deadline := inttypes.NewDeadline(d.Timeout(schema.TimeoutCreate))

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate)/2, func(ctx context.Context) (any, error) {
		return conn.CreateQueue(ctx, input)
	}, errCodeQueueDeletedRecently)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition(ctx), err) {
		input.Tags = nil

		outputRaw, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate)/2, func(ctx context.Context) (any, error) {
			return conn.CreateQueue(ctx, input)
		}, errCodeQueueDeletedRecently)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SQS Queue (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*sqs.CreateQueueOutput).QueueUrl))

	if err := waitQueueAttributesPropagated(ctx, conn, d.Id(), attributes, deadline.Remaining()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SQS Queue (%s) attributes create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition(ctx), err) {
			return append(diags, resourceQueueRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting SQS Queue (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SQSClient(ctx)

	output, err := tfresource.RetryWhenNotFound(ctx, queueReadTimeout, func(ctx context.Context) (map[types.QueueAttributeName]string, error) {
		return findQueueAttributesByURL(ctx, conn, d.Id())
	})

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SQS Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SQS Queue (%s): %s", d.Id(), err)
	}

	name, err := queueNameFromURL(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = queueAttributeMap.APIAttributesToResourceData(output, d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// Backwards compatibility: https://github.com/hashicorp/terraform-provider-aws/issues/19786.
	if d.Get("kms_data_key_reuse_period_seconds").(int) == 0 {
		d.Set("kms_data_key_reuse_period_seconds", defaultQueueKMSDataKeyReusePeriodSeconds)
	}

	d.Set(names.AttrName, name)
	if d.Get("fifo_queue").(bool) {
		d.Set(names.AttrNamePrefix, create.NamePrefixFromNameWithSuffix(name, fifoQueueNameSuffix))
	} else {
		d.Set(names.AttrNamePrefix, create.NamePrefixFromName(name))
	}
	d.Set(names.AttrURL, d.Id())

	return diags
}

func resourceQueueUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SQSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		attributes, err := queueAttributeMap.ResourceDataToAPIAttributesUpdate(d)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &sqs.SetQueueAttributesInput{
			Attributes: flex.ExpandStringyValueMap(attributes),
			QueueUrl:   aws.String(d.Id()),
		}

		_, err = conn.SetQueueAttributes(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SQS Queue (%s) attributes: %s", d.Id(), err)
		}

		if err := waitQueueAttributesPropagated(ctx, conn, d.Id(), attributes, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SQS Queue (%s) attributes update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SQSClient(ctx)

	log.Printf("[DEBUG] Deleting SQS Queue: %s", d.Id())
	_, err := conn.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeQueueDoesNotExist) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SQS Queue (%s): %s", d.Id(), err)
	}

	if err := waitQueueDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SQS Queue (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceQueueCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
	fifoQueue := diff.Get("fifo_queue").(bool)
	contentBasedDeduplication := diff.Get("content_based_deduplication").(bool)
	sqsManagedSSEEnabled := diff.Get("sqs_managed_sse_enabled").(bool)

	if diff.Id() == "" {
		// Create.
		name := queueName(ctx, diff)
		var re *regexp.Regexp

		if fifoQueue {
			re = regexache.MustCompile(`^[0-9A-Za-z_-]{1,75}\.fifo$`)
		} else {
			re = regexache.MustCompile(`^[0-9A-Za-z_-]{1,80}$`)
		}

		if !re.MatchString(name) {
			return fmt.Errorf("invalid queue name: %s", name)
		}

		if sqsManagedSSEEnabled {
			// KmsDataKeyReusePeriodSeconds is only valid for SqsManagedSseEnabled queues.
			if err := diff.SetNew("kms_data_key_reuse_period_seconds", nil); err != nil {
				return err
			}
		}
	} else {
		// Update.
		if sqsManagedSSEEnabled {
			if err := diff.Clear("kms_data_key_reuse_period_seconds"); err != nil {
				return err
			}
		}
	}

	if !fifoQueue && contentBasedDeduplication {
		return fmt.Errorf("content-based deduplication can only be set for FIFO queue")
	}

	return nil
}

func queueName(ctx context.Context, d sdkv2.ResourceDiffer) string {
	optFns := []create.NameGeneratorOptionsFunc{create.WithConfiguredName(d.Get(names.AttrName).(string)), create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string))}
	if d.Get("fifo_queue").(bool) {
		optFns = append(optFns, create.WithSuffix(fifoQueueNameSuffix))
	}
	return create.NewNameGenerator(optFns...).Generate(ctx)
}

// queueNameFromURL returns the SQS queue name from the specified URL.
func queueNameFromURL(u string) (string, error) {
	v, err := url.Parse(u)

	if err != nil {
		return "", err
	}

	// http://sqs.us-west-2.amazonaws.com/123456789012/queueName
	parts := strings.Split(v.Path, "/")

	if len(parts) != 3 {
		return "", fmt.Errorf("SQS Queue URL (%s) is in the incorrect format", u)
	}

	return parts[2], nil
}

func findQueueAttributesByURL(ctx context.Context, conn *sqs.Client, url string) (map[types.QueueAttributeName]string, error) {
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
		QueueUrl:       aws.String(url),
	}

	output, err := findQueueAttributes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findQueueAttributeByTwoPartKey(ctx context.Context, conn *sqs.Client, url string, attributeName types.QueueAttributeName) (*string, error) {
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: []types.QueueAttributeName{attributeName},
		QueueUrl:       aws.String(url),
	}

	output, err := findQueueAttributes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if v, ok := output[attributeName]; ok && v != "" {
		return &v, nil
	}

	return nil, tfresource.NewEmptyResultError()
}

func findQueueAttributes(ctx context.Context, conn *sqs.Client, input *sqs.GetQueueAttributesInput) (map[types.QueueAttributeName]string, error) {
	output, err := conn.GetQueueAttributes(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeQueueDoesNotExist) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Attributes) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	return tfmaps.ApplyToAllKeys(output.Attributes, func(v string) types.QueueAttributeName {
		return types.QueueAttributeName(v)
	}), nil
}

const (
	// Because accounts vary significantly, customizable timeouts are now used to ensure that users
	// who need to wait longer can do so. The default timeouts are set to 3 minutes.

	// If you delete a queue, you must wait at least 60 seconds before creating a queue with the same name.
	// Reference: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_CreateQueue.html
	queueReadTimeout          = 20 * time.Second
	queueDeletedTimeout       = 3 * time.Minute
	queueAttributeReadTimeout = 20 * time.Second

	queueStateExists = "exists"

	queueAttributeStateNotEqual = "notequal"
	queueAttributeStateEqual    = "equal"
)

func statusQueueState(ctx context.Context, conn *sqs.Client, url string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findQueueAttributesByURL(ctx, conn, url)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, queueStateExists, nil
	}
}

func statusQueueAttributeState(ctx context.Context, conn *sqs.Client, url string, expected map[types.QueueAttributeName]string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		attributesMatch := func(got map[types.QueueAttributeName]string) string {
			for k, e := range expected {
				g, ok := got[k]

				if !ok {
					// Missing attribute equivalent to empty expected value.
					if e == "" {
						continue
					}

					// Backwards compatibility: https://github.com/hashicorp/terraform-provider-aws/issues/19786.
					if k == types.QueueAttributeNameKmsDataKeyReusePeriodSeconds && e == strconv.Itoa(defaultQueueKMSDataKeyReusePeriodSeconds) {
						continue
					}

					sse, sseOK := got[types.QueueAttributeNameSqsManagedSseEnabled]
					kmsMaster, kmsOK := got[types.QueueAttributeNameKmsMasterKeyId]
					if k == types.QueueAttributeNameKmsDataKeyReusePeriodSeconds &&
						((!sseOK || (sseOK && sse == "false")) && (!kmsOK || (kmsOK && kmsMaster == ""))) {
						// API won't set if not encrypted
						continue
					}

					return queueAttributeStateNotEqual
				}

				switch k {
				case types.QueueAttributeNamePolicy:
					equivalent, err := awspolicy.PoliciesAreEquivalent(g, e)

					if err != nil {
						return queueAttributeStateNotEqual
					}

					if !equivalent {
						return queueAttributeStateNotEqual
					}
				case types.QueueAttributeNameRedriveAllowPolicy, types.QueueAttributeNameRedrivePolicy:
					if !verify.JSONStringsEqual(g, e) {
						return queueAttributeStateNotEqual
					}
				default:
					if g != e {
						return queueAttributeStateNotEqual
					}
				}
			}

			return queueAttributeStateEqual
		}

		got, err := findQueueAttributesByURL(ctx, conn, url)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		status := attributesMatch(got)

		return got, status, nil
	}
}

func waitQueueAttributesPropagated(ctx context.Context, conn *sqs.Client, url string, expected map[types.QueueAttributeName]string, timeout time.Duration) error {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   []string{queueAttributeStateNotEqual},
		Target:                    []string{queueAttributeStateEqual},
		Refresh:                   statusQueueAttributeState(ctx, conn, url, expected),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 6,               // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		MinTimeout:                5 * time.Second, // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		NotFoundChecks:            10,              // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitQueueDeleted(ctx context.Context, conn *sqs.Client, url string, timeout time.Duration) error {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   []string{queueStateExists},
		Target:                    []string{},
		Refresh:                   statusQueueState(ctx, conn, url),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 15,              // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		MinTimeout:                3 * time.Second, // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		NotFoundChecks:            5,               // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func parseQueueURL(u string) (result inttypes.BaseIdentity, err error) {
	re := regexache.MustCompile(`^https://sqs\.([a-z0-9-]+)\.[^/]+/([0-9]{12})/.+`)
	match := re.FindStringSubmatch(u)
	if match == nil {
		return inttypes.BaseIdentity{}, fmt.Errorf("could not parse %q as SQS URL", u)
	}
	return inttypes.BaseIdentity{
		AccountID: match[2],
		Region:    match[1],
	}, nil
}
