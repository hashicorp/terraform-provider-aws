package elasticache

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceParameterGroupCreate,
		Read:   resourceParameterGroupRead,
		Update: resourceParameterGroupUpdate,
		Delete: resourceParameterGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: ParameterHash,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	createOpts := elasticache.CreateCacheParameterGroupInput{
		CacheParameterGroupName:   aws.String(d.Get("name").(string)),
		CacheParameterGroupFamily: aws.String(d.Get("family").(string)),
		Description:               aws.String(d.Get("description").(string)),
	}

	if len(tags) > 0 {
		createOpts.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Create ElastiCache Parameter Group: %#v", createOpts)
	resp, err := conn.CreateCacheParameterGroup(&createOpts)

	if createOpts.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating ElastiCache Parameter Group with tags: %s. Trying create without tags.", err)

		createOpts.Tags = nil
		resp, err = conn.CreateCacheParameterGroup(&createOpts)
	}

	if err != nil {
		return fmt.Errorf("error creating ElastiCache Parameter Group: %w", err)
	}

	d.SetId(aws.StringValue(resp.CacheParameterGroup.CacheParameterGroupName))
	d.Set("arn", resp.CacheParameterGroup.ARN)
	log.Printf("[INFO] ElastiCache Parameter Group ID: %s", d.Id())

	return resourceParameterGroupUpdate(d, meta)
}

func resourceParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	parameterGroup, err := FindParameterGroupByName(conn, d.Id())
	if err != nil {
		return fmt.Errorf("unable to find ElastiCache Parameter Group (%s): %w", d.Id(), err)
	}

	d.Set("name", parameterGroup.CacheParameterGroupName)
	d.Set("family", parameterGroup.CacheParameterGroupFamily)
	d.Set("description", parameterGroup.Description)
	d.Set("arn", parameterGroup.ARN)

	// Only include user customized parameters as there's hundreds of system/default ones
	describeParametersOpts := elasticache.DescribeCacheParametersInput{
		CacheParameterGroupName: aws.String(d.Id()),
		Source:                  aws.String("user"),
	}

	describeParametersResp, err := conn.DescribeCacheParameters(&describeParametersOpts)
	if err != nil {
		return err
	}

	d.Set("parameter", FlattenParameters(describeParametersResp.Parameters))

	tags, err := ListTags(conn, aws.StringValue(parameterGroup.ARN))

	if err != nil && !verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		return fmt.Errorf("error listing tags for ElastiCache Parameter Group (%s): %w", d.Id(), err)
	}

	if err != nil {
		log.Printf("[WARN] failed listing tags for Elasticache Parameter Group (%s): %s", d.Id(), err)
	}

	if tags != nil {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return fmt.Errorf("error setting tags_all: %w", err)
		}
	}

	return nil
}

func resourceParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
		toRemove, toAdd := ParameterChanges(o, n)

		log.Printf("[DEBUG] Parameters to remove: %#v", toRemove)
		log.Printf("[DEBUG] Parameters to add or update: %#v", toAdd)

		// We can only modify 20 parameters at a time, so walk them until
		// we've got them all.
		const maxParams = 20

		for len(toRemove) > 0 {
			var paramsToModify []*elasticache.ParameterNameValue
			if len(toRemove) <= maxParams {
				paramsToModify, toRemove = toRemove[:], nil
			} else {
				paramsToModify, toRemove = toRemove[:maxParams], toRemove[maxParams:]
			}

			err := resourceResetParameterGroup(conn, d.Get("name").(string), paramsToModify)

			// When attempting to reset the reserved-memory parameter, the API
			// can return two types of error.
			//
			// In the commercial partition, it will return a 400 error with:
			//   InvalidParameterValue: Parameter reserved-memory doesn't exist
			//
			// In the GovCloud partition it will return the below 500 error,
			// which causes the AWS Go SDK to automatically retry and timeout:
			//   InternalFailure: An internal error has occurred. Please try your query again at a later time.
			//
			// Instead of hardcoding the reserved-memory parameter removal
			// above, which may become out of date, here we add logic to
			// workaround this API behavior

			if tfresource.TimedOut(err) || tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "Parameter reserved-memory doesn't exist") {
				for i, paramToModify := range paramsToModify {
					if aws.StringValue(paramToModify.ParameterName) != "reserved-memory" {
						continue
					}

					// Always reset the top level error and remove the reset for reserved-memory
					err = nil
					paramsToModify = append(paramsToModify[:i], paramsToModify[i+1:]...)

					// If we are only trying to remove reserved-memory and not perform
					// an update to reserved-memory or reserved-memory-percent, we
					// can attempt to workaround the API issue by switching it to
					// reserved-memory-percent first then reset that temporary parameter.

					tryReservedMemoryPercentageWorkaround := true
					for _, configuredParameter := range toAdd {
						if aws.StringValue(configuredParameter.ParameterName) == "reserved-memory-percent" {
							tryReservedMemoryPercentageWorkaround = false
							break
						}
					}

					if !tryReservedMemoryPercentageWorkaround {
						break
					}

					// The reserved-memory-percent parameter does not exist in redis2.6 and redis2.8
					family := d.Get("family").(string)
					if family == "redis2.6" || family == "redis2.8" {
						log.Printf("[WARN] Cannot reset ElastiCache Parameter Group (%s) reserved-memory parameter with %s family", d.Id(), family)
						break
					}

					workaroundParams := []*elasticache.ParameterNameValue{
						{
							ParameterName:  aws.String("reserved-memory-percent"),
							ParameterValue: aws.String("0"),
						},
					}
					err = resourceModifyParameterGroup(conn, d.Get("name").(string), paramsToModify)
					if err != nil {
						log.Printf("[WARN] Error attempting reserved-memory workaround to switch to reserved-memory-percent: %s", err)
						break
					}

					err = resourceResetParameterGroup(conn, d.Get("name").(string), workaroundParams)
					if err != nil {
						log.Printf("[WARN] Error attempting reserved-memory workaround to reset reserved-memory-percent: %s", err)
					}

					break
				}

				// Retry any remaining parameter resets with reserved-memory potentially removed
				if len(paramsToModify) > 0 {
					err = resourceResetParameterGroup(conn, d.Get("name").(string), paramsToModify)
				}
			}

			if err != nil {
				return fmt.Errorf("error resetting ElastiCache Parameter Group: %w", err)
			}
		}

		for len(toAdd) > 0 {
			var paramsToModify []*elasticache.ParameterNameValue
			if len(toAdd) <= maxParams {
				paramsToModify, toAdd = toAdd[:], nil
			} else {
				paramsToModify, toAdd = toAdd[:maxParams], toAdd[maxParams:]
			}

			err := resourceModifyParameterGroup(conn, d.Get("name").(string), paramsToModify)
			if err != nil {
				return fmt.Errorf("error modifying ElastiCache Parameter Group: %w", err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, d.Get("arn").(string), o, n)

		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
				// explicitly setting tags or not an iso-unsupported error
				return fmt.Errorf("failed updating ElastiCache Parameter Group (%s) tags: %w", d.Get("arn").(string), err)
			}

			log.Printf("[WARN] failed updating tags for ElastiCache Parameter Group (%s): %s", d.Get("arn").(string), err)
		}
	}

	return resourceParameterGroupRead(d, meta)
}

func resourceParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	err := deleteParameterGroup(conn, d.Id())
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting ElastiCache Parameter Group (%s): %w", d.Id(), err)
	}
	return nil
}

func deleteParameterGroup(conn *elasticache.ElastiCache, name string) error {
	deleteOpts := elasticache.DeleteCacheParameterGroupInput{
		CacheParameterGroupName: aws.String(name),
	}
	err := resource.Retry(3*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteCacheParameterGroup(&deleteOpts)
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == "CacheParameterGroupNotFoundFault" {
				return nil
			}
			if ok && awsErr.Code() == "InvalidCacheParameterGroupState" {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteCacheParameterGroup(&deleteOpts)
	}

	return err
}

func ParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["value"].(string)))

	return create.StringHashcode(buf.String())
}

func ParameterChanges(o, n interface{}) (remove, addOrUpdate []*elasticache.ParameterNameValue) {
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}

	os := o.(*schema.Set)
	ns := n.(*schema.Set)

	om := make(map[string]*elasticache.ParameterNameValue, os.Len())
	for _, raw := range os.List() {
		param := raw.(map[string]interface{})
		om[param["name"].(string)] = expandParameter(param)
	}
	nm := make(map[string]*elasticache.ParameterNameValue, len(addOrUpdate))
	for _, raw := range ns.List() {
		param := raw.(map[string]interface{})
		nm[param["name"].(string)] = expandParameter(param)
	}

	// Remove: key is in old, but not in new
	remove = make([]*elasticache.ParameterNameValue, 0, os.Len())
	for k := range om {
		if _, ok := nm[k]; !ok {
			remove = append(remove, om[k])
		}
	}

	// Add or Update: key is in new, but not in old or has changed value
	addOrUpdate = make([]*elasticache.ParameterNameValue, 0, ns.Len())
	for k, nv := range nm {
		ov, ok := om[k]
		if !ok || ok && (aws.StringValue(nv.ParameterValue) != aws.StringValue(ov.ParameterValue)) {
			addOrUpdate = append(addOrUpdate, nm[k])
		}
	}

	return remove, addOrUpdate
}

func resourceResetParameterGroup(conn *elasticache.ElastiCache, name string, parameters []*elasticache.ParameterNameValue) error {
	input := elasticache.ResetCacheParameterGroupInput{
		CacheParameterGroupName: aws.String(name),
		ParameterNameValues:     parameters,
	}
	return resource.Retry(30*time.Second, func() *resource.RetryError {
		_, err := conn.ResetCacheParameterGroup(&input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidCacheParameterGroupStateFault, " has pending changes") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
}

func resourceModifyParameterGroup(conn *elasticache.ElastiCache, name string, parameters []*elasticache.ParameterNameValue) error {
	input := elasticache.ModifyCacheParameterGroupInput{
		CacheParameterGroupName: aws.String(name),
		ParameterNameValues:     parameters,
	}
	_, err := conn.ModifyCacheParameterGroup(&input)
	return err
}

// Flattens an array of Parameters into a []map[string]interface{}
func FlattenParameters(list []*elasticache.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		if i.ParameterValue != nil {
			result = append(result, map[string]interface{}{
				"name":  strings.ToLower(aws.StringValue(i.ParameterName)),
				"value": aws.StringValue(i.ParameterValue),
			})
		}
	}
	return result
}

// Takes the result of flatmap.Expand for an array of parameters and
// returns Parameter API compatible objects
func ExpandParameters(configured []interface{}) []*elasticache.ParameterNameValue {
	parameters := make([]*elasticache.ParameterNameValue, len(configured))

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for i, pRaw := range configured {
		parameters[i] = expandParameter(pRaw.(map[string]interface{}))
	}

	return parameters
}

func expandParameter(param map[string]interface{}) *elasticache.ParameterNameValue {
	return &elasticache.ParameterNameValue{
		ParameterName:  aws.String(param["name"].(string)),
		ParameterValue: aws.String(param["value"].(string)),
	}
}
