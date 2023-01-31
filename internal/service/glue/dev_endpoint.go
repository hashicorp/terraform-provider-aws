package glue

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
				Elem:     schema.TypeString,
			},
			"arn": {
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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\w+\.\w+$`), "must match version pattern X.X"),
			},
			"name": {
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
			"public_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"public_keys"},
			},
			"public_keys": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
				ConflictsWith: []string{"public_key"},
				MaxItems:      5,
			},
			"role_arn": {
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
			"security_group_ids": {
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Set:          schema.HashString,
				RequiredWith: []string{"subnet_id"},
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"security_group_ids"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(glue.WorkerType_Values(), false),
				ConflictsWith: []string{"number_of_nodes"},
				ForceNew:      true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
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
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &glue.CreateDevEndpointInput{
		EndpointName: aws.String(name),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		Tags:         Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("arguments"); ok {
		input.Arguments = flex.ExpandStringMap(v.(map[string]interface{}))
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
		input.NumberOfNodes = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("public_key"); ok {
		input.PublicKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_keys"); ok {
		publicKeys := flex.ExpandStringSet(v.(*schema.Set))
		input.PublicKeys = publicKeys
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		securityGroupIDs := flex.ExpandStringSet(v.(*schema.Set))
		input.SecurityGroupIds = securityGroupIDs
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Dev Endpoint: %#v", *input)
	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		_, err := conn.CreateDevEndpointWithContext(ctx, input)
		if err != nil {
			// Retry for IAM eventual consistency
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "should be given assume role permissions for Glue Service") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "is not authorized to perform") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "S3 endpoint and NAT validation has failed for subnetId") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateDevEndpointWithContext(ctx, input)
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
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	if err := d.Set("arn", endpointARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arn for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("arguments", aws.StringValueMap(endpoint.Arguments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arguments for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("availability_zone", endpoint.AvailabilityZone); err != nil {
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

	if err := d.Set("name", endpoint.EndpointName); err != nil {
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

	if err := d.Set("public_key", endpoint.PublicKey); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting public_key for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_keys", flex.FlattenStringSet(endpoint.PublicKeys)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting public_keys for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("role_arn", endpoint.RoleArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting role_arn for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("security_configuration", endpoint.SecurityConfiguration); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_configuration for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("security_group_ids", flex.FlattenStringSet(endpoint.SecurityGroupIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_group_ids for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("status", endpoint.Status); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting status for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("subnet_id", endpoint.SubnetId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_id for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("vpc_id", endpoint.VpcId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_id for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("worker_type", endpoint.WorkerType); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting worker_type for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, endpointARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Glue Dev Endpoint (%s): %s", endpointARN, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
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
	conn := meta.(*conns.AWSClient).GlueConn()

	input := &glue.UpdateDevEndpointInput{
		EndpointName: aws.String(d.Get("name").(string)),
	}

	hasChanged := false

	customLibs := &glue.DevEndpointCustomLibraries{}

	if d.HasChange("arguments") {
		oldRaw, newRaw := d.GetChange("arguments")
		old := oldRaw.(map[string]interface{})
		new := newRaw.(map[string]interface{})
		add, remove, _ := verify.DiffStringMaps(old, new)

		removeKeys := make([]*string, 0)
		for k := range remove {
			removeKeys = append(removeKeys, aws.String(k))
		}

		input.AddArguments = add
		input.DeleteArguments = removeKeys

		hasChanged = true
	}

	if d.HasChange("extra_jars_s3_path") {
		customLibs.ExtraJarsS3Path = aws.String(d.Get("extra_jars_s3_path").(string))
		input.CustomLibraries = customLibs
		input.UpdateEtlLibraries = aws.Bool(true)

		hasChanged = true
	}

	if d.HasChange("extra_python_libs_s3_path") {
		customLibs.ExtraPythonLibsS3Path = aws.String(d.Get("extra_python_libs_s3_path").(string))
		input.CustomLibraries = customLibs
		input.UpdateEtlLibraries = aws.Bool(true)

		hasChanged = true
	}

	if d.HasChange("public_key") {
		input.PublicKey = aws.String(d.Get("public_key").(string))

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

		input.AddPublicKeys = flex.ExpandStringSet(add)
		log.Printf("[DEBUG] expectedCreate public keys: %v", add)

		input.DeletePublicKeys = flex.ExpandStringSet(remove)
		log.Printf("[DEBUG] remove public keys: %v", remove)

		hasChanged = true
	}

	if hasChanged {
		log.Printf("[DEBUG] Updating Glue Dev Endpoint: %s", input)
		err := resource.RetryContext(ctx, 5*time.Minute, func() *resource.RetryError {
			_, err := conn.UpdateDevEndpointWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "another concurrent update operation") {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateDevEndpointWithContext(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Dev Endpoint: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceDevEndpointRead(ctx, d, meta)...)
}

func resourceDevEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	log.Printf("[INFO] Deleting Glue Dev Endpoint: %s", d.Id())
	_, err := conn.DeleteDevEndpointWithContext(ctx, &glue.DeleteDevEndpointInput{
		EndpointName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
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
