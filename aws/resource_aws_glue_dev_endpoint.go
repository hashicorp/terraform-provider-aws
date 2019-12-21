package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGlueDevEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueDevEndpointCreate,
		Read:   resourceAwsGlueDevEndpointRead,
		Update: resourceAwsDevEndpointUpdate,
		Delete: resourceAwsDevEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arguments": {
				Type:     schema.TypeMap,
				Optional: true,
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
				ValidateFunc: validateGlueVersion,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"tags": tagsSchema(),

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
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					glue.WorkerTypeG1x,
					glue.WorkerTypeG2x,
					glue.WorkerTypeStandard,
				}, false),
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

func resourceAwsGlueDevEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	input := &glue.CreateDevEndpointInput{
		EndpointName: aws.String(name),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlueTags(),
	}

	if kv, ok := d.GetOk("arguments"); ok {
		arguments := make(map[string]string)
		for k, v := range kv.(map[string]interface{}) {
			arguments[k] = v.(string)
		}
		input.Arguments = aws.StringMap(arguments)
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
		publicKeys := expandStringSet(v.(*schema.Set))
		input.PublicKeys = publicKeys
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		securityGroupIDs := expandStringSet(v.(*schema.Set))
		input.SecurityGroupIds = securityGroupIDs
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Dev Endpoint: %#v", *input)
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateDevEndpoint(input)
		if err != nil {
			// Retry for IAM eventual consistency
			if isAWSErr(err, glue.ErrCodeInvalidInputException, "should be given assume role permissions for Glue Service") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, glue.ErrCodeInvalidInputException, "S3 endpoint and NAT validation has failed for subnetId") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error creating Glue Dev Endpoint: %s", err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] Waiting for Glue Dev Endpoint (%s) to become available", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"PROVISIONING",
		},
		Target:  []string{"READY"},
		Refresh: glueDevEndpointStateRefreshFunc(conn, d.Id()),
		Timeout: 15 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error while waiting for Glue Dev Endpoint (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsGlueDevEndpointRead(d, meta)
}

func resourceAwsGlueDevEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	request := &glue.GetDevEndpointInput{
		EndpointName: aws.String(d.Id()),
	}

	output, err := conn.GetDevEndpoint(request)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Dev Endpoint (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	endpoint := output.DevEndpoint
	if endpoint == nil {
		log.Printf("[WARN] Glue Dev Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	endpointARN := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("devEndpoint/%s", d.Id()),
	}.String()

	if err := d.Set("arn", endpointARN); err != nil {
		return fmt.Errorf("error setting arn for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("arguments", aws.StringValueMap(endpoint.Arguments)); err != nil {
		return fmt.Errorf("error setting arguments for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("availability_zone", endpoint.AvailabilityZone); err != nil {
		return fmt.Errorf("error setting availability_zone for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("extra_jars_s3_path", endpoint.ExtraJarsS3Path); err != nil {
		return fmt.Errorf("error setting extra_jars_s3_path for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("extra_python_libs_s3_path", endpoint.ExtraPythonLibsS3Path); err != nil {
		return fmt.Errorf("error setting extra_python_libs_s3_path for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("failure_reason", endpoint.FailureReason); err != nil {
		return fmt.Errorf("error setting failure_reason for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("glue_version", endpoint.GlueVersion); err != nil {
		return fmt.Errorf("error setting glue_version for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("name", endpoint.EndpointName); err != nil {
		return fmt.Errorf("error setting name for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("number_of_nodes", endpoint.NumberOfNodes); err != nil {
		return fmt.Errorf("error setting number_of_nodes for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("number_of_workers", endpoint.NumberOfWorkers); err != nil {
		return fmt.Errorf("error setting number_of_workers for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("private_address", endpoint.PrivateAddress); err != nil {
		return fmt.Errorf("error setting private_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_address", endpoint.PublicAddress); err != nil {
		return fmt.Errorf("error setting public_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_key", endpoint.PublicKey); err != nil {
		return fmt.Errorf("error setting public_key for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_keys", flattenStringSet(endpoint.PublicKeys)); err != nil {
		return fmt.Errorf("error setting public_keys for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("role_arn", endpoint.RoleArn); err != nil {
		return fmt.Errorf("error setting role_arn for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("security_configuration", endpoint.SecurityConfiguration); err != nil {
		return fmt.Errorf("error setting security_configuration for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("security_group_ids", flattenStringSet(endpoint.SecurityGroupIds)); err != nil {
		return fmt.Errorf("error setting security_group_ids for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("status", endpoint.Status); err != nil {
		return fmt.Errorf("error setting status for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("subnet_id", endpoint.SubnetId); err != nil {
		return fmt.Errorf("error setting subnet_id for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("vpc_id", endpoint.VpcId); err != nil {
		return fmt.Errorf("error setting vpc_id for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("worker_type", endpoint.WorkerType); err != nil {
		return fmt.Errorf("error setting worker_type for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	tags, err := keyvaluetags.GlueListTags(conn, endpointARN)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Dev Endpoint (%s): %s", endpointARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("yarn_endpoint_address", endpoint.YarnEndpointAddress); err != nil {
		return fmt.Errorf("error setting yarn_endpoint_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("zeppelin_remote_spark_interpreter_port", endpoint.ZeppelinRemoteSparkInterpreterPort); err != nil {
		return fmt.Errorf("error setting zeppelin_remote_spark_interpreter_port for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsDevEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.UpdateDevEndpointInput{
		EndpointName: aws.String(d.Get("name").(string)),
	}

	hasChanged := false

	customLibs := &glue.DevEndpointCustomLibraries{}

	if d.HasChange("arguments") {
		oldRaw, newRaw := d.GetChange("arguments")
		old := oldRaw.(map[string]interface{})
		new := newRaw.(map[string]interface{})
		create, remove := diffArguments(old, new)
		input.AddArguments = create
		input.DeleteArguments = remove

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
		oldRaw, newRaw := d.GetChange("public_keys")
		old := oldRaw.(*schema.Set)
		new := newRaw.(*schema.Set)
		create, remove := diffPublicKeys(expandStringSet(old), expandStringSet(new))
		input.AddPublicKeys = create
		log.Printf("[DEBUG] expectedCreate public keys: %v", create)
		input.DeletePublicKeys = remove
		log.Printf("[DEBUG] remove public keys: %v", remove)

		hasChanged = true
	}

	if hasChanged {
		log.Printf("[DEBUG] Updating Glue Dev Endpoint: %s", input)

		_, err := conn.UpdateDevEndpoint(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Dev Endpoint: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GlueUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGlueDevEndpointRead(d, meta)
}

func resourceAwsDevEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	deleteOpts := &glue.DeleteDevEndpointInput{
		EndpointName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting Glue Dev Endpoint: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDevEndpoint(deleteOpts)
		if err == nil {
			return nil
		}

		glueErr, ok := err.(awserr.Error)
		if !ok {
			return resource.NonRetryableError(err)
		}

		if glueErr.Code() == glue.ErrCodeEntityNotFoundException {
			return nil
		}

		return resource.NonRetryableError(fmt.Errorf("error deleting Glue Dev Endpoint: %s", err))
	})
}

func glueDevEndpointStateRefreshFunc(conn *glue.Glue, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		getDevEndpointInput := &glue.GetDevEndpointInput{
			EndpointName: aws.String(name),
		}
		endpoint, err := conn.GetDevEndpoint(getDevEndpointInput)
		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil, "", nil
			}

			return nil, "", err
		}

		if endpoint == nil {
			return nil, "", nil
		}

		return endpoint, *endpoint.DevEndpoint.Status, nil
	}
}

func diffPublicKeys(oldKeys, newKeys []*string) ([]*string, []*string) {
	var create []*string
	var remove []*string

	for _, oldKey := range oldKeys {
		found := false
		for _, newKey := range newKeys {
			if *oldKey == *newKey {
				found = true
				break
			}
		}
		if !found {
			remove = append(remove, oldKey)
		}
	}

	for _, newKey := range newKeys {
		found := false
		for _, oldKey := range oldKeys {
			if *oldKey == *newKey {
				found = true
				break
			}
		}
		if !found {
			create = append(create, newKey)
		}
	}

	return create, remove
}

func diffArguments(oldArgs, newArgs map[string]interface{}) (map[string]*string, []*string) {
	var create = make(map[string]*string)
	var remove []*string

	for oldArgKey, oldArgVal := range oldArgs {
		found := false
		for newArgKey, newArgVal := range newArgs {
			if oldArgKey == newArgKey &&
				oldArgVal.(string) == newArgVal.(string) {
				found = true
				break
			}
		}
		if !found {
			remove = append(remove, &oldArgKey)
		}
	}

	for newArgKey, newArgVal := range newArgs {
		found := false
		for oldArgKey, oldArgVal := range oldArgs {
			if oldArgKey == newArgKey &&
				oldArgVal.(string) == newArgVal.(string) {
				found = true
				break
			}
		}
		if !found {
			create[newArgKey] = aws.String(newArgVal.(string))
		}
	}

	return create, remove
}
