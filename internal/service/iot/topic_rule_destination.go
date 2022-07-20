package iot

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTopicRuleDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicRuleDestinationCreate,
		ReadWithoutTimeout:   resourceTopicRuleDestinationRead,
		UpdateWithoutTimeout: resourceTopicRuleDestinationUpdate,
		DeleteWithoutTimeout: resourceTopicRuleDestinationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"vpc_configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
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
	conn := meta.(*conns.AWSClient).IoTConn

	input := &iot.CreateTopicRuleDestinationInput{
		DestinationConfiguration: &iot.TopicRuleDestinationConfiguration{},
	}

	if v, ok := d.GetOk("vpc_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationConfiguration.VpcConfiguration = expandVPCDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[INFO] Creating IoT Topic Rule Destination: %s", input)
	outputRaw, err := tfresource.RetryWhen(propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateTopicRuleDestinationWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, iot.ErrCodeInvalidRequestException, "sts:AssumeRole") ||
				tfawserr.ErrMessageContains(err, iot.ErrCodeInvalidRequestException, "Missing permission") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return diag.Errorf("creating IoT Topic Rule Destination: %s", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*iot.CreateTopicRuleDestinationOutput).TopicRuleDestination.Arn))

	if _, err := waitTopicRuleDestinationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for IoT Topic Rule Destination (%s) create: %s", d.Id(), err)
	}

	if _, ok := d.GetOk("enabled"); !ok {
		_, err := conn.UpdateTopicRuleDestinationWithContext(ctx, &iot.UpdateTopicRuleDestinationInput{
			Arn:    aws.String(d.Id()),
			Status: aws.String(iot.TopicRuleDestinationStatusDisabled),
		})

		if err != nil {
			return diag.Errorf("disabling IoT Topic Rule Destination (%s): %s", d.Id(), err)
		}

		if _, err := waitTopicRuleDestinationDisabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("waiting for IoT Topic Rule Destination (%s) disable: %s", d.Id(), err)
		}
	}

	return resourceTopicRuleDestinationRead(ctx, d, meta)
}

func resourceTopicRuleDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	output, err := FindTopicRuleDestinationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Topic Rule Destination %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading IoT Topic Rule Destination (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("enabled", aws.StringValue(output.Status) == iot.TopicRuleDestinationStatusEnabled)
	if output.VpcProperties != nil {
		if err := d.Set("vpc_configuration", []interface{}{flattenVPCDestinationProperties(output.VpcProperties)}); err != nil {
			return diag.Errorf("setting vpc_configuration: %s", err)
		}
	} else {
		d.Set("vpc_configuration", nil)
	}

	return nil
}

func resourceTopicRuleDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.HasChange("enabled") {
		input := &iot.UpdateTopicRuleDestinationInput{
			Arn:    aws.String(d.Id()),
			Status: aws.String(iot.TopicRuleDestinationStatusEnabled),
		}
		waiter := waitTopicRuleDestinationEnabled

		if _, ok := d.GetOk("enabled"); !ok {
			input.Status = aws.String(iot.TopicRuleDestinationStatusDisabled)
			waiter = waitTopicRuleDestinationDisabled
		}

		_, err := conn.UpdateTopicRuleDestinationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating IoT Topic Rule Destination (%s): %s", d.Id(), err)
		}

		if _, err := waiter(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("waiting for IoT Topic Rule Destination (%s) update: %s", d.Id(), err)
		}
	}

	return resourceTopicRuleDestinationRead(ctx, d, meta)
}

func resourceTopicRuleDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	log.Printf("[INFO] Deleting IoT Topic Rule Destination: %s", d.Id())
	_, err := conn.DeleteTopicRuleDestinationWithContext(ctx, &iot.DeleteTopicRuleDestinationInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting IoT Topic Rule Destination: %s", err)
	}

	if _, err := waitTopicRuleDestinationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for IoT Topic Rule Destination (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func expandVPCDestinationConfiguration(tfMap map[string]interface{}) *iot.VpcDestinationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &iot.VpcDestinationConfiguration{}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["vpc_id"].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenVPCDestinationProperties(apiObject *iot.VpcDestinationProperties) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.SecurityGroups; v != nil {
		tfMap["security_groups"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	return tfMap
}

func statusTopicRuleDestination(ctx context.Context, conn *iot.IoT, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTopicRuleDestinationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitTopicRuleDestinationCreated(ctx context.Context, conn *iot.IoT, arn string, timeout time.Duration) (*iot.TopicRuleDestination, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{iot.TopicRuleDestinationStatusInProgress},
		Target:  []string{iot.TopicRuleDestinationStatusEnabled},
		Refresh: statusTopicRuleDestination(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*iot.TopicRuleDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitTopicRuleDestinationDeleted(ctx context.Context, conn *iot.IoT, arn string, timeout time.Duration) (*iot.TopicRuleDestination, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{iot.TopicRuleDestinationStatusDeleting},
		Target:  []string{},
		Refresh: statusTopicRuleDestination(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*iot.TopicRuleDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitTopicRuleDestinationDisabled(ctx context.Context, conn *iot.IoT, arn string, timeout time.Duration) (*iot.TopicRuleDestination, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{iot.TopicRuleDestinationStatusInProgress},
		Target:  []string{iot.TopicRuleDestinationStatusDisabled},
		Refresh: statusTopicRuleDestination(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*iot.TopicRuleDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitTopicRuleDestinationEnabled(ctx context.Context, conn *iot.IoT, arn string, timeout time.Duration) (*iot.TopicRuleDestination, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{iot.TopicRuleDestinationStatusInProgress},
		Target:  []string{iot.TopicRuleDestinationStatusEnabled},
		Refresh: statusTopicRuleDestination(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*iot.TopicRuleDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))

		return output, err
	}

	return nil, err
}
