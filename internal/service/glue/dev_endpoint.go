// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_dev_endpoint", name="Dev Endpoint")
// @Tags(identifierAttribute="arn")
func ResourceDevEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDevEndpointCreate,
		ReadWithoutTimeout:   resourceDevEndpointRead,
		UpdateWithoutTimeout: resourceDevEndpointUpdate,
		DeleteWithoutTimeout: resourceDevEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arguments": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"extra_jars_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"extra_python_libs_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"glue_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^\w+\.\w+$`), "must match version pattern X.X"),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"number_of_nodes": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"number_of_workers", "worker_type"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return new == "0"
				},
				ValidateFunc: validation.IntAtLeast(2),
			},
			"number_of_workers": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntAtLeast(2),
				ConflictsWith: []string{"number_of_nodes"},
			},
			names.AttrPublicKey: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"public_keys"},
			},
			"public_keys": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
				ConflictsWith: []string{names.AttrPublicKey},
				MaxItems:      5,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrSecurityGroupIDs: {
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Set:          schema.HashString,
				RequiredWith: []string{names.AttrSubnetID},
			},
			names.AttrSubnetID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{names.AttrSecurityGroupIDs},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"private_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"yarn_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zeppelin_remote_spark_interpreter_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"worker_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.WorkerType](),
				ConflictsWith:    []string{"number_of_nodes"},
				ForceNew:         true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDevEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &glue.CreateDevEndpointInput{
		EndpointName: aws.String(name),
		RoleArn:      aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("arguments"); ok {
		input.Arguments = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("extra_jars_s3_path"); ok {
		input.ExtraJarsS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("extra_python_libs_s3_path"); ok {
		input.ExtraPythonLibsS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("number_of_nodes"); ok {
		input.NumberOfNodes = int32(v.(int))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrPublicKey); ok {
		input.PublicKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_keys"); ok {
		publicKeys := flex.ExpandStringValueSet(v.(*schema.Set))
		input.PublicKeys = publicKeys
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		securityGroupIDs := flex.ExpandStringValueSet(v.(*schema.Set))
		input.SecurityGroupIds = securityGroupIDs
	}

	if v, ok := d.GetOk(names.AttrSubnetID); ok {
		input.SubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = awstypes.WorkerType(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Dev Endpoint: %#v", *input)
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.CreateDevEndpoint(ctx, input)
		if err != nil {
			// Retry for IAM eventual consistency
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "should be given assume role permissions for Glue Service") {
				return retry.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "is not authorized to perform") {
				return retry.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "S3 endpoint and NAT validation has failed for subnetId") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateDevEndpoint(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Dev Endpoint: %s", err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] Waiting for Glue Dev Endpoint (%s) to become available", d.Id())
	if _, err := waitDevEndpointCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "while waiting for Glue Dev Endpoint (%s) to become available: %s", d.Id(), err)
	}

	return append(diags, resourceDevEndpointRead(ctx, d, meta)...)
}

func resourceDevEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	endpoint, err := FindDevEndpointByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Dev Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	endpointARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("devEndpoint/%s", d.Id()),
	}.String()

	if err := d.Set(names.AttrARN, endpointARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arn for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("arguments", endpoint.Arguments); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arguments for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrAvailabilityZone, endpoint.AvailabilityZone); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting availability_zone for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("extra_jars_s3_path", endpoint.ExtraJarsS3Path); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting extra_jars_s3_path for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("extra_python_libs_s3_path", endpoint.ExtraPythonLibsS3Path); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting extra_python_libs_s3_path for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("failure_reason", endpoint.FailureReason); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting failure_reason for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("glue_version", endpoint.GlueVersion); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting glue_version for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrName, endpoint.EndpointName); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("number_of_nodes", endpoint.NumberOfNodes); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting number_of_nodes for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("number_of_workers", endpoint.NumberOfWorkers); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting number_of_workers for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("private_address", endpoint.PrivateAddress); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting private_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_address", endpoint.PublicAddress); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting public_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPublicKey, endpoint.PublicKey); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting public_key for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_keys", flex.FlattenStringValueSet(endpoint.PublicKeys)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting public_keys for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrRoleARN, endpoint.RoleArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting role_arn for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("security_configuration", endpoint.SecurityConfiguration); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_configuration for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrSecurityGroupIDs, flex.FlattenStringValueSet(endpoint.SecurityGroupIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_group_ids for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrStatus, endpoint.Status); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting status for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrSubnetID, endpoint.SubnetId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_id for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrVPCID, endpoint.VpcId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_id for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("worker_type", endpoint.WorkerType); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting worker_type for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("yarn_endpoint_address", endpoint.YarnEndpointAddress); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting yarn_endpoint_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("zeppelin_remote_spark_interpreter_port", endpoint.ZeppelinRemoteSparkInterpreterPort); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting zeppelin_remote_spark_interpreter_port for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceDevEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := &glue.UpdateDevEndpointInput{
		EndpointName: aws.String(d.Get(names.AttrName).(string)),
	}

	hasChanged := false

	customLibs := &awstypes.DevEndpointCustomLibraries{}

	if d.HasChange("arguments") {
		oldRaw, newRaw := d.GetChange("arguments")
		old := oldRaw.(map[string]interface{})
		new := newRaw.(map[string]interface{})
		add, remove, _ := flex.DiffStringValueMaps(old, new)

		removeKeys := make([]string, 0)
		for k := range remove {
			removeKeys = append(removeKeys, k)
		}

		input.AddArguments = add
		input.DeleteArguments = removeKeys

		hasChanged = true
	}

	if d.HasChange("extra_jars_s3_path") {
		customLibs.ExtraJarsS3Path = aws.String(d.Get("extra_jars_s3_path").(string))
		input.CustomLibraries = customLibs
		input.UpdateEtlLibraries = true

		hasChanged = true
	}

	if d.HasChange("extra_python_libs_s3_path") {
		customLibs.ExtraPythonLibsS3Path = aws.String(d.Get("extra_python_libs_s3_path").(string))
		input.CustomLibraries = customLibs
		input.UpdateEtlLibraries = true

		hasChanged = true
	}

	if d.HasChange(names.AttrPublicKey) {
		input.PublicKey = aws.String(d.Get(names.AttrPublicKey).(string))

		hasChanged = true
	}

	if d.HasChange("public_keys") {
		o, n := d.GetChange("public_keys")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := os.Difference(ns)
		add := ns.Difference(os)

		input.AddPublicKeys = flex.ExpandStringValueSet(add)
		log.Printf("[DEBUG] expectedCreate public keys: %v", add)

		input.DeletePublicKeys = flex.ExpandStringValueSet(remove)
		log.Printf("[DEBUG] remove public keys: %v", remove)

		hasChanged = true
	}

	if hasChanged {
		log.Printf("[DEBUG] Updating Glue Dev Endpoint: %+v", input)
		err := retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
			_, err := conn.UpdateDevEndpoint(ctx, input)
			if err != nil {
				if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "another concurrent update operation") {
					return retry.RetryableError(err)
				}

				return retry.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateDevEndpoint(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Dev Endpoint: %s", err)
		}
	}

	return append(diags, resourceDevEndpointRead(ctx, d, meta)...)
}

func resourceDevEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[INFO] Deleting Glue Dev Endpoint: %s", d.Id())
	_, err := conn.DeleteDevEndpoint(ctx, &glue.DeleteDevEndpointInput{
		EndpointName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for Glue Dev Endpoint (%s) to become terminated", d.Id())
	if _, err := waitDevEndpointDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "while waiting for Glue Dev Endpoint (%s) to become terminated: %s", d.Id(), err)
	}

	return diags
}
