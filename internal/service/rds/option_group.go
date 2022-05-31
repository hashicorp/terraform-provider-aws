package rds

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOptionGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceOptionGroupCreate,
		Read:   resourceOptionGroupRead,
		Update: resourceOptionGroupUpdate,
		Delete: resourceOptionGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validOptionGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validOptionGroupNamePrefix,
			},
			"engine_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"major_engine_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"option_group_description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},

			"option": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"option_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"option_settings": {
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
						},
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"db_security_group_memberships": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"vpc_security_group_memberships": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceOptionHash,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOptionGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}

	createOpts := &rds.CreateOptionGroupInput{
		EngineName:             aws.String(d.Get("engine_name").(string)),
		MajorEngineVersion:     aws.String(d.Get("major_engine_version").(string)),
		OptionGroupDescription: aws.String(d.Get("option_group_description").(string)),
		OptionGroupName:        aws.String(groupName),
		Tags:                   Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Create DB Option Group: %#v", createOpts)
	output, err := conn.CreateOptionGroup(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating DB Option Group: %s", err)
	}

	d.SetId(strings.ToLower(groupName))
	log.Printf("[INFO] DB Option Group ID: %s", d.Id())

	// Set for update
	d.Set("arn", output.OptionGroup.OptionGroupArn)

	return resourceOptionGroupUpdate(d, meta)
}

func resourceOptionGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &rds.DescribeOptionGroupsInput{
		OptionGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Describe DB Option Group: %#v", params)
	options, err := conn.DescribeOptionGroups(params)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeOptionGroupNotFoundFault) {
		log.Printf("[WARN] RDS Option Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error Describing DB Option Group: %s", err)
	}

	var option *rds.OptionGroup
	for _, ogl := range options.OptionGroupsList {
		if aws.StringValue(ogl.OptionGroupName) == d.Id() {
			option = ogl
			break
		}
	}

	if option == nil {
		log.Printf("[WARN] RDS Option Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", option.OptionGroupArn)
	d.Set("name", option.OptionGroupName)
	d.Set("major_engine_version", option.MajorEngineVersion)
	d.Set("engine_name", option.EngineName)
	d.Set("option_group_description", option.OptionGroupDescription)

	if err := d.Set("option", flattenOptions(option.Options, expandOptionConfiguration(d.Get("option").(*schema.Set).List()))); err != nil {
		return fmt.Errorf("error setting option: %s", err)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for RDS Option Group (%s): %s", d.Get("arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func optionInList(optionName string, list []*string) bool {
	for _, opt := range list {
		if aws.StringValue(opt) == optionName {
			return true
		}
	}
	return false
}

func resourceOptionGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	if d.HasChange("option") {
		o, n := d.GetChange("option")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		optionsToInclude := expandOptionConfiguration(ns.Difference(os).List())
		optionsToIncludeNames := flattenOptionNames(ns.Difference(os).List())
		optionsToRemove := []*string{}
		optionsToRemoveNames := flattenOptionNames(os.Difference(ns).List())

		for _, optionToRemoveName := range optionsToRemoveNames {
			if optionInList(*optionToRemoveName, optionsToIncludeNames) {
				continue
			}
			optionsToRemove = append(optionsToRemove, optionToRemoveName)
		}

		// Ensure there is actually something to update
		// InvalidParameterValue: At least one option must be added, modified, or removed.
		if len(optionsToInclude) > 0 || len(optionsToRemove) > 0 {
			modifyOpts := &rds.ModifyOptionGroupInput{
				OptionGroupName:  aws.String(d.Id()),
				ApplyImmediately: aws.Bool(true),
			}

			if len(optionsToInclude) > 0 {
				modifyOpts.OptionsToInclude = optionsToInclude
			}

			if len(optionsToRemove) > 0 {
				modifyOpts.OptionsToRemove = optionsToRemove
			}

			log.Printf("[DEBUG] Modify DB Option Group: %s", modifyOpts)

			err := resource.Retry(propagationTimeout, func() *resource.RetryError {
				_, err := conn.ModifyOptionGroup(modifyOpts)
				if err != nil {
					// InvalidParameterValue: IAM role ARN value is invalid or does not include the required permissions for: SQLSERVER_BACKUP_RESTORE
					if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if tfresource.TimedOut(err) {
				_, err = conn.ModifyOptionGroup(modifyOpts)
			}
			if err != nil {
				return fmt.Errorf("Error modifying DB Option Group: %s", err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating RDS Option Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceOptionGroupRead(d, meta)
}

func resourceOptionGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	deleteOpts := &rds.DeleteOptionGroupInput{
		OptionGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Delete DB Option Group: %#v", deleteOpts)
	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteOptionGroup(deleteOpts)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, rds.ErrCodeInvalidOptionGroupStateFault) {
				log.Printf(`[DEBUG] AWS believes the RDS Option Group is still in use, this could be because of a internal snapshot create by AWS, see github issue #4597 for more info. retrying...`)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteOptionGroup(deleteOpts)
	}
	if err != nil {
		return fmt.Errorf("Error Deleting DB Option Group: %s", err)
	}
	return nil
}

func flattenOptionNames(configured []interface{}) []*string {
	var optionNames []*string
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})
		optionNames = append(optionNames, aws.String(data["option_name"].(string)))
	}

	return optionNames
}

func resourceOptionHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["option_name"].(string)))
	if _, ok := m["port"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", m["port"].(int)))
	}

	for _, oRaw := range m["option_settings"].(*schema.Set).List() {
		o := oRaw.(map[string]interface{})
		buf.WriteString(fmt.Sprintf("%s-", o["name"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", o["value"].(string)))
	}

	for _, vpcRaw := range m["vpc_security_group_memberships"].(*schema.Set).List() {
		buf.WriteString(fmt.Sprintf("%s-", vpcRaw.(string)))
	}

	for _, sgRaw := range m["db_security_group_memberships"].(*schema.Set).List() {
		buf.WriteString(fmt.Sprintf("%s-", sgRaw.(string)))
	}

	if v, ok := m["version"]; ok && v.(string) != "" {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}
