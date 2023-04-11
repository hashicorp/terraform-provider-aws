package vpclattice

import (
	"context"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_target_group")
// @Tags(identifierAttribute="arn")
func ResourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTargetGroupCreate,
		ReadWithoutTimeout:   resourceTargetGroupRead,
		UpdateWithoutTimeout: resourceTargetGroupUpdate,
		DeleteWithoutTimeout: resourceTargetGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"health_check": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"interval": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  30,
									},
									"timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"healthy_threshold": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  5,
									},
									"unhealthy_threshold": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  2,
									},
									"matcher": {
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/",
									},
									"port": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IsPortNumber,
									},
									"protocol": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"protocol_version": {
										Type:     schema.TypeString,
										Optional: true,
										StateFunc: func(v interface{}) string {
											return strings.ToUpper(v.(string))
										},
										ValidateFunc: validation.StringInSlice([]string{
											"GRPC",
											"HTTP1",
											"HTTP2",
										}, true),
									},
								},
							},
						},
						"ip_address_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"protocol_version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							StateFunc: func(v interface{}) string {
								return strings.ToUpper(v.(string))
							},
							ValidateFunc: validation.StringInSlice([]string{
								"GRPC",
								"HTTP1",
								"HTTP2",
							}, true),
						},
						"vpc_identifier": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameTargetGroup = "Target Group"
)

func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	in := &vpclattice.CreateTargetGroupInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(d.Get("name").(string)),
		Type:        types.TargetGroupType(d.Get("type").(string)),
		Tags:        GetTagsIn(ctx),
	}

	if d.Get("type") != string(types.TargetGroupTypeLambda) {
		if v, ok := d.GetOk("config"); ok && len(v.([]interface{})) > 0 {
			config := expandConfigAttributes(v.([]interface{})[0].(map[string]interface{}))
			in.Config = &types.TargetGroupConfig{
				Port:            config.Port,
				Protocol:        config.Protocol,
				VpcIdentifier:   config.VpcIdentifier,
				IpAddressType:   config.IpAddressType,
				ProtocolVersion: config.ProtocolVersion,
				HealthCheck:     config.HealthCheck,
			}
		}
	}

	out, err := conn.CreateTargetGroup(ctx, in)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameService, d.Get("name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameService, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Id))

	if _, err := waitTargetGroupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionWaitingForCreation, ResNameTargetGroup, d.Id(), err)
	}

	return resourceTargetGroupRead(ctx, d, meta)
}

func resourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	out, err := FindTargetGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VpcLattice TargetGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, ResNameTargetGroup, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	d.Set("status", out.Status)
	d.Set("type", out.Type)
	if out.Config != nil {
		if err := d.Set("config", flattenTargetGroupConfig(out.Config)); err != nil {
			return create.DiagError(names.VPCLattice, create.ErrActionSetting, ResNameTargetGroup, d.Id(), err)
		}
	} else {
		d.Set("config", nil)
	}

	return nil
}

func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	if d.HasChangesExcept("tags", "tags_all") {
		in := &vpclattice.UpdateTargetGroupInput{
			TargetGroupIdentifier: aws.String(d.Id()),
		}

		if d.HasChange("config") {
			oldConfig, newConfig := d.GetChange("config")
			oldConfigMap := expandConfigAttributes(oldConfig.([]interface{})[0].(map[string]interface{}))
			newConfigMap := expandConfigAttributes(newConfig.([]interface{})[0].(map[string]interface{}))

			if !reflect.DeepEqual(oldConfigMap.HealthCheck, newConfigMap.HealthCheck) {
				in.HealthCheck = newConfigMap.HealthCheck
			}
		}

		if in.HealthCheck == nil {
			return nil
		}

		log.Printf("[DEBUG] Updating VpcLattice TargetGroup (%s): %#v", d.Id(), in)
		out, err := conn.UpdateTargetGroup(ctx, in)
		if err != nil {
			return create.DiagError(names.VPCLattice, create.ErrActionUpdating, ResNameTargetGroup, d.Id(), err)
		}

		if _, err := waitTargetGroupUpdated(ctx, conn, aws.ToString(out.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.VPCLattice, create.ErrActionWaitingForUpdate, ResNameTargetGroup, d.Id(), err)
		}
	}

	return resourceTargetGroupRead(ctx, d, meta)
}

func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	log.Printf("[INFO] Deleting VpcLattice TargetGroup %s", d.Id())

	_, err := conn.DeleteTargetGroup(ctx, &vpclattice.DeleteTargetGroupInput{
		TargetGroupIdentifier: aws.String(d.Id()),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.VPCLattice, create.ErrActionDeleting, ResNameTargetGroup, d.Id(), err)
	}

	if _, err := waitTargetGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameTargetGroup, d.Id(), err)
	}

	return nil
}

func waitTargetGroupCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.CreateTargetGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.TargetGroupStatusCreateInProgress),
		Target:                    enum.Slice(types.TargetGroupStatusActive),
		Refresh:                   statusTargetGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.CreateTargetGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTargetGroupUpdated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.UpdateTargetGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.TargetGroupStatusCreateInProgress),
		Target:                    enum.Slice(types.TargetGroupStatusActive),
		Refresh:                   statusTargetGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.UpdateTargetGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTargetGroupDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.DeleteTargetGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TargetGroupStatusDeleteInProgress, types.TargetGroupStatusActive),
		Target:  []string{},
		Refresh: statusTargetGroup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.DeleteTargetGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func statusTargetGroup(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindTargetGroupByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindTargetGroupByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetTargetGroupOutput, error) {
	in := &vpclattice.GetTargetGroupInput{
		TargetGroupIdentifier: aws.String(id),
	}
	out, err := conn.GetTargetGroup(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenTargetGroupConfig(apiObject *types.TargetGroupConfig) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"port":           aws.Int32(*apiObject.Port),
		"protocol":       string(apiObject.Protocol),
		"vpc_identifier": aws.String(*apiObject.VpcIdentifier),
	}

	if apiObject.IpAddressType != "" {
		m["ip_address_type"] = string(apiObject.IpAddressType)
	}

	if apiObject.ProtocolVersion != "" {
		m["protocol_version"] = string(apiObject.ProtocolVersion)
	}

	if apiObject.HealthCheck != nil {
		port := apiObject.Port
		if apiObject.HealthCheck.Port != nil {
			port = apiObject.HealthCheck.Port
		}
		m["health_check"] = []map[string]interface{}{flattenHealthCheckConfig(apiObject.HealthCheck, *port)}
	}

	return []map[string]interface{}{m}
}

func flattenHealthCheckConfig(apiObject *types.HealthCheckConfig, port int32) map[string]interface{} {
	m := map[string]interface{}{
		"enabled":             aws.Bool(*apiObject.Enabled),
		"interval":            aws.Int32(*apiObject.HealthCheckIntervalSeconds),
		"timeout":             aws.Int32(*apiObject.HealthCheckTimeoutSeconds),
		"healthy_threshold":   aws.Int32(*apiObject.HealthyThresholdCount),
		"unhealthy_threshold": aws.Int32(*apiObject.UnhealthyThresholdCount),
		"path":                aws.String(*apiObject.Path),
		"port":                aws.Int32(port),
		"protocol":            string(apiObject.Protocol),
		"protocol_version":    string(apiObject.ProtocolVersion),
	}

	if matcher, ok := apiObject.Matcher.(*types.MatcherMemberHttpCode); ok {
		m["matcher"] = matcher.Value
	}

	return m
}

func expandConfigAttributes(tfMap map[string]interface{}) *types.TargetGroupConfig {
	if tfMap == nil {
		return nil
	}
	apiObject := &types.TargetGroupConfig{}

	if v, ok := tfMap["port"].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["protocol"].(string); ok {
		protocol := types.TargetGroupProtocol(v)
		apiObject.Protocol = protocol
	}

	if v, ok := tfMap["vpc_identifier"].(string); ok {
		apiObject.VpcIdentifier = aws.String(v)
	}
	if v, ok := tfMap["health_check"].([]interface{}); ok && len(v) > 0 {
		apiObject.HealthCheck = expandHealthCheckConfigAttributes(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["ip_address_type"].(string); ok {
		ipAddressType := types.IpAddressType(v)
		apiObject.IpAddressType = ipAddressType
	}

	if v, ok := tfMap["protocol_version"].(string); ok {
		protocolVersion := types.TargetGroupProtocolVersion(v)
		apiObject.ProtocolVersion = protocolVersion
	}

	return apiObject
}

func expandHealthCheckConfigAttributes(tfMap map[string]interface{}) *types.HealthCheckConfig {
	apiObject := &types.HealthCheckConfig{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["interval"].(int); ok {
		apiObject.HealthCheckIntervalSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["timeout"].(int); ok {
		apiObject.HealthCheckTimeoutSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["healthy_threshold"].(int); ok {
		apiObject.HealthyThresholdCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["unhealthy_threshold"].(int); ok {
		apiObject.UnhealthyThresholdCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["path"].(string); ok {
		apiObject.Path = aws.String(v)
	}

	if v, ok := tfMap["port"].(int); ok {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap["protocol"].(string); ok {
		apiObject.Protocol = types.TargetGroupProtocol(v)
	}

	if v, ok := tfMap["protocol_version"].(string); ok {
		apiObject.ProtocolVersion = types.HealthCheckProtocolVersion(v)
	}

	if v, ok := tfMap["matcher"].(string); ok {
		apiObject.Matcher = &types.MatcherMemberHttpCode{Value: v}
	}

	return apiObject
}
