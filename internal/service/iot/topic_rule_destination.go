// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_topic_rule_destination", name="Topic Rule Destination")
func resourceTopicRuleDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicRuleDestinationCreate,
		ReadWithoutTimeout:   resourceTopicRuleDestinationRead,
		UpdateWithoutTimeout: resourceTopicRuleDestinationUpdate,
		DeleteWithoutTimeout: resourceTopicRuleDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrVPCConfiguration: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceTopicRuleDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	input := &iot.CreateTopicRuleDestinationInput{
		DestinationConfiguration: &awstypes.TopicRuleDestinationConfiguration{},
	}

	if v, ok := d.GetOk(names.AttrVPCConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationConfiguration.VpcConfiguration = expandVPCDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateTopicRuleDestination(ctx, input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Topic Rule Destination: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*iot.CreateTopicRuleDestinationOutput).TopicRuleDestination.Arn))

	if _, err := waitTopicRuleDestinationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IoT Topic Rule Destination (%s) create: %s", d.Id(), err)
	}

	if _, ok := d.GetOk(names.AttrEnabled); !ok {
		input := &iot.UpdateTopicRuleDestinationInput{
			Arn:    aws.String(d.Id()),
			Status: awstypes.TopicRuleDestinationStatusDisabled,
		}

		_, err := conn.UpdateTopicRuleDestination(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling IoT Topic Rule Destination (%s): %s", d.Id(), err)
		}

		if _, err := waitTopicRuleDestinationDisabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IoT Topic Rule Destination (%s) disable: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTopicRuleDestinationRead(ctx, d, meta)...)
}

func resourceTopicRuleDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findTopicRuleDestinationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Topic Rule Destination %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Topic Rule Destination (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrEnabled, (output.Status == awstypes.TopicRuleDestinationStatusEnabled))
	if output.VpcProperties != nil {
		if err := d.Set(names.AttrVPCConfiguration, []interface{}{flattenVPCDestinationProperties(output.VpcProperties)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_configuration: %s", err)
		}
	} else {
		d.Set(names.AttrVPCConfiguration, nil)
	}

	return diags
}

func resourceTopicRuleDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChange(names.AttrEnabled) {
		input := &iot.UpdateTopicRuleDestinationInput{
			Arn:    aws.String(d.Id()),
			Status: awstypes.TopicRuleDestinationStatusEnabled,
		}
		waiter := waitTopicRuleDestinationEnabled

		if _, ok := d.GetOk(names.AttrEnabled); !ok {
			input.Status = awstypes.TopicRuleDestinationStatusDisabled
			waiter = waitTopicRuleDestinationDisabled
		}

		_, err := conn.UpdateTopicRuleDestination(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Topic Rule Destination (%s): %s", d.Id(), err)
		}

		if _, err := waiter(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IoT Topic Rule Destination (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTopicRuleDestinationRead(ctx, d, meta)...)
}

func resourceTopicRuleDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	log.Printf("[INFO] Deleting IoT Topic Rule Destination: %s", d.Id())

	// DeleteTopicRuleDestination returns unhelpful errors such as
	// "UnauthorizedException: Access to TopicRuleDestination 'xxx' was denied" when querying for a rule destination that doesn't exist.
	_, err := conn.DeleteTopicRuleDestination(ctx, &iot.DeleteTopicRuleDestinationInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.UnauthorizedException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Topic Rule Destination: %s", err)
	}

	if _, err := waitTopicRuleDestinationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IoT Topic Rule Destination (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findTopicRuleDestinationByARN(ctx context.Context, conn *iot.Client, arn string) (*awstypes.TopicRuleDestination, error) {
	// GetTopicRuleDestination returns unhelpful errors such as
	//	"UnauthorizedException: Access to TopicRuleDestination 'arn:aws:iot:us-west-2:123456789012:ruledestination/vpc/f267138a-7383-4670-9e44-a7fe2f48af5e' was denied"
	// when querying for a rule destination that doesn't exist.
	inputL := &iot.ListTopicRuleDestinationsInput{}
	var destination *awstypes.TopicRuleDestinationSummary

	pages := iot.NewListTopicRuleDestinationsPaginator(conn, inputL)
pageLoop:
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DestinationSummaries {
			v := v
			if aws.ToString(v.Arn) == arn {
				destination = &v
				break pageLoop
			}
		}
	}

	if destination == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	inputG := &iot.GetTopicRuleDestinationInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetTopicRuleDestination(ctx, inputG)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: inputG,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TopicRuleDestination == nil {
		return nil, tfresource.NewEmptyResultError(inputG)
	}

	return output.TopicRuleDestination, nil
}

func statusTopicRuleDestination(ctx context.Context, conn *iot.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTopicRuleDestinationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitTopicRuleDestinationCreated(ctx context.Context, conn *iot.Client, arn string, timeout time.Duration) (*awstypes.TopicRuleDestination, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(string(awstypes.TopicRuleDestinationStatusInProgress)),
		Target:  enum.Slice(string(awstypes.TopicRuleDestinationStatusEnabled)),
		Refresh: statusTopicRuleDestination(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TopicRuleDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitTopicRuleDestinationDeleted(ctx context.Context, conn *iot.Client, arn string, timeout time.Duration) (*awstypes.TopicRuleDestination, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(string(awstypes.TopicRuleDestinationStatusDeleting)),
		Target:  []string{},
		Refresh: statusTopicRuleDestination(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TopicRuleDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitTopicRuleDestinationDisabled(ctx context.Context, conn *iot.Client, arn string, timeout time.Duration) (*awstypes.TopicRuleDestination, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(string(awstypes.TopicRuleDestinationStatusInProgress)),
		Target:  enum.Slice(string(awstypes.TopicRuleDestinationStatusDisabled)),
		Refresh: statusTopicRuleDestination(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TopicRuleDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitTopicRuleDestinationEnabled(ctx context.Context, conn *iot.Client, arn string, timeout time.Duration) (*awstypes.TopicRuleDestination, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(string(awstypes.TopicRuleDestinationStatusInProgress)),
		Target:  enum.Slice(string(awstypes.TopicRuleDestinationStatusEnabled)),
		Refresh: statusTopicRuleDestination(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TopicRuleDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func expandVPCDestinationConfiguration(tfMap map[string]interface{}) *awstypes.VpcDestinationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VpcDestinationConfiguration{}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrVPCID].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenVPCDestinationProperties(apiObject *awstypes.VpcDestinationProperties) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.SecurityGroups; v != nil {
		tfMap[names.AttrSecurityGroups] = aws.StringSlice(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = aws.StringSlice(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap[names.AttrVPCID] = aws.ToString(v)
	}

	return tfMap
}
