package redshiftserverless

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceWorkgroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkgroupCreate,
		ReadWithoutTimeout:   resourceWorkgroupRead,
		UpdateWithoutTimeout: resourceWorkgroupUpdate,
		DeleteWithoutTimeout: resourceWorkgroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"config_parameter": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameter_key": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"datestyle", "enable_user_activity_logging", "query_group", "search_path", "max_query_execution_time"}, false),
							Required:     true,
						},
						"parameter_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vpc_endpoint": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"vpc_endpoint_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"vpc_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"network_interface": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"availability_zone": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"network_interface_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"private_ip_address": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"subnet_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"enhanced_vpc_routing": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workgroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workgroup_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkgroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := redshiftserverless.CreateWorkgroupInput{
		NamespaceName: aws.String(d.Get("namespace_name").(string)),
		WorkgroupName: aws.String(d.Get("workgroup_name").(string)),
	}

	if v, ok := d.GetOk("base_capacity"); ok {
		input.BaseCapacity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("config_parameter"); ok && len(v.([]interface{})) > 0 {
		input.ConfigParameters = expandConfigParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("enhanced_vpc_routing"); ok {
		input.EnhancedVpcRouting = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("publicly_accessible"); ok {
		input.PubliclyAccessible = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	input.Tags = Tags(tags.IgnoreAWS())

	out, err := conn.CreateWorkgroupWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Workgroup : %s", err)
	}

	d.SetId(aws.StringValue(out.Workgroup.WorkgroupName))

	if _, err := waitWorkgroupAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Workgroup (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceWorkgroupRead(ctx, d, meta)...)
}

func resourceWorkgroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindWorkgroupByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Workgroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Workgroup (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(out.WorkgroupArn)
	d.Set("arn", arn)
	d.Set("base_capacity", out.BaseCapacity)
	d.Set("enhanced_vpc_routing", out.EnhancedVpcRouting)
	d.Set("namespace_name", out.NamespaceName)
	d.Set("publicly_accessible", out.PubliclyAccessible)
	d.Set("security_group_ids", flex.FlattenStringSet(out.SecurityGroupIds))
	d.Set("subnet_ids", flex.FlattenStringSet(out.SubnetIds))
	d.Set("workgroup_id", out.WorkgroupId)
	d.Set("workgroup_name", out.WorkgroupName)
	if err := d.Set("config_parameter", flattenConfigParameters(out.ConfigParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting config_parameter: %s", err)
	}

	if err := d.Set("endpoint", []interface{}{flattenEndpoint(out.Endpoint)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, "UnknownOperationException") {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "listing tags for edshift Serverless Workgroup (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceWorkgroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			WorkgroupName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("base_capacity"); ok {
			input.BaseCapacity = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("config_parameter"); ok && len(v.([]interface{})) > 0 {
			input.ConfigParameters = expandConfigParameters(v.([]interface{}))
		}

		if v, ok := d.GetOk("enhanced_vpc_routing"); ok {
			input.EnhancedVpcRouting = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("publicly_accessible"); ok {
			input.PubliclyAccessible = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		_, err := conn.UpdateWorkgroupWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Workgroup (%s): %s", d.Id(), err)
		}

		if _, err := waitWorkgroupAvailable(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Workgroup (%s) to be updated: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Workgroup (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceWorkgroupRead(ctx, d, meta)...)
}

func resourceWorkgroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn()

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 10*time.Minute,
		func() (interface{}, error) {
			return conn.DeleteWorkgroupWithContext(ctx, &redshiftserverless.DeleteWorkgroupInput{
				WorkgroupName: aws.String(d.Id()),
			})
		},
		// "ConflictException: There is an operation running on the workgroup. Try deleting the workgroup again later."
		redshiftserverless.ErrCodeConflictException, "operation running")

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Workgroup (%s): %s", d.Id(), err)
	}

	if _, err := waitWorkgroupDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Workgroup (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func expandConfigParameter(tfMap map[string]interface{}) *redshiftserverless.ConfigParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshiftserverless.ConfigParameter{}

	if v, ok := tfMap["parameter_key"].(string); ok {
		apiObject.ParameterKey = aws.String(v)
	}

	if v, ok := tfMap["parameter_value"].(string); ok {
		apiObject.ParameterValue = aws.String(v)
	}

	return apiObject
}

func expandConfigParameters(tfList []interface{}) []*redshiftserverless.ConfigParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*redshiftserverless.ConfigParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandConfigParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenConfigParameter(apiObject *redshiftserverless.ConfigParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ParameterKey; v != nil {
		tfMap["parameter_key"] = aws.StringValue(v)
	}

	if v := apiObject.ParameterValue; v != nil {
		tfMap["parameter_value"] = aws.StringValue(v)
	}
	return tfMap
}

func flattenConfigParameters(apiObjects []*redshiftserverless.ConfigParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenConfigParameter(apiObject))
	}

	return tfList
}

func flattenEndpoint(apiObject *redshiftserverless.Endpoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if v := apiObject.Address; v != nil {
		tfMap["address"] = aws.StringValue(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap["port"] = aws.Int64Value(v)
	}

	if v := apiObject.VpcEndpoints; v != nil {
		tfMap["vpc_endpoint"] = flattenVPCEndpoints(v)
	}

	return tfMap
}

func flattenVPCEndpoints(apiObjects []*redshiftserverless.VpcEndpoint) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenVPCEndpoint(apiObject))
	}

	return tfList
}

func flattenVPCEndpoint(apiObject *redshiftserverless.VpcEndpoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.VpcEndpointId; v != nil {
		tfMap["vpc_endpoint_id"] = aws.StringValue(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaces; v != nil {
		tfMap["network_interface"] = flattenNetworkInterfaces(v)
	}
	return tfMap
}

func flattenNetworkInterfaces(apiObjects []*redshiftserverless.NetworkInterface) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterface(apiObject))
	}

	return tfList
}

func flattenNetworkInterface(apiObject *redshiftserverless.NetworkInterface) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap["availability_zone"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	if v := apiObject.PrivateIpAddress; v != nil {
		tfMap["private_ip_address"] = aws.StringValue(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap["subnet_id"] = aws.StringValue(v)
	}
	return tfMap
}
