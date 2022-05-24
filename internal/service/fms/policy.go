package fms

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyCreate,
		Read:   resourcePolicyRead,
		Update: resourcePolicyUpdate,
		Delete: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delete_all_policy_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"delete_unused_fm_managed_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"exclude_resource_tags": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"exclude_map": {
				Type:             schema.TypeList,
				MaxItems:         1,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"orgunit": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"include_map": {
				Type:             schema.TypeList,
				MaxItems:         1,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"orgunit": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy_update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"remediation_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"resource_tags": tftags.TagsSchema(),
			"resource_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^([\p{L}\p{Z}\p{N}_.:/=+\-@]*)$`), "must match a supported resource type, such as AWS::EC2::VPC, see also: https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html"),
				ConflictsWith: []string{"resource_type_list"},
			},
			"resource_type_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([\p{L}\p{Z}\p{N}_.:/=+\-@]*)$`), "must match a supported resource type, such as AWS::EC2::VPC, see also: https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html"),
				},
				ConflictsWith: []string{"resource_type"},
			},
			"security_service_policy_data": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"managed_service_data": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fms.PutPolicyInput{
		Policy:  resourcePolicyExpandPolicy(d),
		TagList: Tags(tags.IgnoreAWS()),
	}

	output, err := conn.PutPolicy(input)

	if err != nil {
		return fmt.Errorf("error creating FMS Policy: %w", err)
	}

	d.SetId(aws.StringValue(output.Policy.PolicyId))

	return resourcePolicyRead(d, meta)
}

func resourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindPolicyByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FMS Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FMS Policy (%s): %w", d.Id(), err)
	}

	if err := resourcePolicyFlattenPolicy(d, output); err != nil {
		return err
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for FMS Policy (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourcePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FMSConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &fms.PutPolicyInput{
			Policy: resourcePolicyExpandPolicy(d),
		}

		_, err := conn.PutPolicy(input)

		if err != nil {
			return fmt.Errorf("error updating FMS Policy (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FMS Policy (%s) tags: %w", d.Id(), err)
		}
	}

	return resourcePolicyRead(d, meta)
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FMSConn

	log.Printf("[DEBUG] Deleting FMS Policy: %s", d.Id())
	_, err := conn.DeletePolicy(&fms.DeletePolicyInput{
		PolicyId:                 aws.String(d.Id()),
		DeleteAllPolicyResources: aws.Bool(d.Get("delete_all_policy_resources").(bool)),
	})

	if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FMS Policy (%s): %w", d.Id(), err)
	}

	return nil
}

func FindPolicyByID(conn *fms.FMS, id string) (*fms.GetPolicyOutput, error) {
	input := &fms.GetPolicyInput{
		PolicyId: aws.String(id),
	}

	output, err := conn.GetPolicy(input)

	if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func resourcePolicyFlattenPolicy(d *schema.ResourceData, resp *fms.GetPolicyOutput) error {
	d.Set("arn", resp.PolicyArn)

	d.Set("name", resp.Policy.PolicyName)
	d.Set("exclude_resource_tags", resp.Policy.ExcludeResourceTags)
	if err := d.Set("exclude_map", flattenPolicyMap(resp.Policy.ExcludeMap)); err != nil {
		return err
	}
	if err := d.Set("include_map", flattenPolicyMap(resp.Policy.IncludeMap)); err != nil {
		return err
	}
	d.Set("remediation_enabled", resp.Policy.RemediationEnabled)
	if err := d.Set("resource_type_list", resp.Policy.ResourceTypeList); err != nil {
		return err
	}
	d.Set("delete_unused_fm_managed_resources", resp.Policy.DeleteUnusedFMManagedResources)
	d.Set("resource_type", resp.Policy.ResourceType)
	d.Set("policy_update_token", resp.Policy.PolicyUpdateToken)
	if err := d.Set("resource_tags", flattenResourceTags(resp.Policy.ResourceTags)); err != nil {
		return err
	}

	securityServicePolicy := []map[string]string{{
		"type":                 *resp.Policy.SecurityServicePolicyData.Type,
		"managed_service_data": *resp.Policy.SecurityServicePolicyData.ManagedServiceData,
	}}
	if err := d.Set("security_service_policy_data", securityServicePolicy); err != nil {
		return err
	}

	return nil
}

func resourcePolicyExpandPolicy(d *schema.ResourceData) *fms.Policy {
	resourceType := aws.String("ResourceTypeList")
	resourceTypeList := flex.ExpandStringSet(d.Get("resource_type_list").(*schema.Set))
	if t, ok := d.GetOk("resource_type"); ok {
		resourceType = aws.String(t.(string))
	}

	fmsPolicy := &fms.Policy{
		PolicyName:                     aws.String(d.Get("name").(string)),
		RemediationEnabled:             aws.Bool(d.Get("remediation_enabled").(bool)),
		ResourceType:                   resourceType,
		ResourceTypeList:               resourceTypeList,
		ExcludeResourceTags:            aws.Bool(d.Get("exclude_resource_tags").(bool)),
		DeleteUnusedFMManagedResources: aws.Bool(d.Get("delete_unused_fm_managed_resources").(bool)),
	}

	if d.Id() != "" {
		fmsPolicy.PolicyId = aws.String(d.Id())
		fmsPolicy.PolicyUpdateToken = aws.String(d.Get("policy_update_token").(string))
	}

	fmsPolicy.ExcludeMap = expandPolicyMap(d.Get("exclude_map").([]interface{}))

	fmsPolicy.IncludeMap = expandPolicyMap(d.Get("include_map").([]interface{}))

	fmsPolicy.ResourceTags = constructResourceTags(d.Get("resource_tags"))

	securityServicePolicy := d.Get("security_service_policy_data").([]interface{})[0].(map[string]interface{})
	fmsPolicy.SecurityServicePolicyData = &fms.SecurityServicePolicyData{
		ManagedServiceData: aws.String(securityServicePolicy["managed_service_data"].(string)),
		Type:               aws.String(securityServicePolicy["type"].(string)),
	}

	return fmsPolicy
}

func expandPolicyMap(set []interface{}) map[string][]*string {
	fmsPolicyMap := map[string][]*string{}
	if len(set) > 0 {
		if _, ok := set[0].(map[string]interface{}); !ok {
			return fmsPolicyMap
		}
		for key, listValue := range set[0].(map[string]interface{}) {
			var flatKey string
			switch key {
			case "account":
				flatKey = "ACCOUNT"
			case "orgunit":
				flatKey = "ORG_UNIT"
			}

			for _, value := range listValue.(*schema.Set).List() {
				fmsPolicyMap[flatKey] = append(fmsPolicyMap[flatKey], aws.String(value.(string)))
			}
		}
	}
	return fmsPolicyMap
}

func flattenPolicyMap(fmsPolicyMap map[string][]*string) []interface{} {
	flatPolicyMap := map[string]interface{}{}

	for key, value := range fmsPolicyMap {
		switch key {
		case "ACCOUNT":
			flatPolicyMap["account"] = value
		case "ORG_UNIT":
			flatPolicyMap["orgunit"] = value
		default:
			log.Printf("[WARNING] Unexpected key (%q) found in FMS policy", key)
		}
	}

	return []interface{}{flatPolicyMap}
}

func flattenResourceTags(resourceTags []*fms.ResourceTag) map[string]interface{} {
	resTags := map[string]interface{}{}

	for _, v := range resourceTags {
		resTags[*v.Key] = v.Value
	}
	return resTags
}

func constructResourceTags(rTags interface{}) []*fms.ResourceTag {
	var rTagList []*fms.ResourceTag

	tags := rTags.(map[string]interface{})
	for k, v := range tags {
		rTagList = append(rTagList, &fms.ResourceTag{Key: aws.String(k), Value: aws.String(v.(string))})
	}

	return rTagList
}
