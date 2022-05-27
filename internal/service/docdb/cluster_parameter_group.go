package docdb

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const clusterParameterGroupMaxParamsBulkEdit = 20

func ResourceClusterParameterGroup() *schema.Resource {

	return &schema.Resource{
		Create: resourceClusterParameterGroupCreate,
		Read:   resourceClusterParameterGroupRead,
		Update: resourceClusterParameterGroupUpdate,
		Delete: resourceClusterParameterGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				ValidateFunc:  validParamGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validParamGroupNamePrefix,
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
						"apply_method": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  docdb.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice([]string{
								docdb.ApplyMethodImmediate,
								docdb.ApplyMethodPendingReboot,
							}, false),
						},
					},
				},
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}

}

func resourceClusterParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DocDBConn
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

	createOpts := docdb.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Create DocDB Cluster Parameter Group: %#v", createOpts)

	resp, err := conn.CreateDBClusterParameterGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating DocDB Cluster Parameter Group: %s", err)
	}

	d.SetId(aws.StringValue(createOpts.DBClusterParameterGroupName))

	d.Set("arn", resp.DBClusterParameterGroup.DBClusterParameterGroupArn)

	return resourceClusterParameterGroupUpdate(d, meta)
}

func resourceClusterParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DocDBConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeOpts := &docdb.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeDBClusterParameterGroups(describeOpts)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
			log.Printf("[WARN] DocDB Cluster Parameter Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading DocDB Cluster Parameter Group (%s): %w", d.Id(), err)
	}

	if len(describeResp.DBClusterParameterGroups) != 1 ||
		aws.StringValue(describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupName) != d.Id() {
		return fmt.Errorf("Unable to find Cluster Parameter Group: %#v", describeResp.DBClusterParameterGroups)
	}

	arn := aws.StringValue(describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupArn)
	d.Set("arn", arn)
	d.Set("description", describeResp.DBClusterParameterGroups[0].Description)
	d.Set("family", describeResp.DBClusterParameterGroups[0].DBParameterGroupFamily)
	d.Set("name", describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupName)

	describeParametersOpts := &docdb.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	describeParametersResp, err := conn.DescribeDBClusterParameters(describeParametersOpts)
	if err != nil {
		return fmt.Errorf("error reading DocDB Cluster Parameter Group (%s) parameters: %w", d.Id(), err)
	}

	if err := d.Set("parameter", flattenParameters(describeParametersResp.Parameters, d.Get("parameter").(*schema.Set).List())); err != nil {
		return fmt.Errorf("error setting docdb cluster parameter: %w", err)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for DocumentDB Cluster Parameter Group (%s): %s", d.Get("arn").(string), err)
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

func resourceClusterParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DocDBConn

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		parameters := expandParameters(ns.Difference(os).List())
		if len(parameters) > 0 {
			// We can only modify 20 parameters at a time, so walk them until
			// we've got them all.
			for parameters != nil {
				var paramsToModify []*docdb.Parameter
				if len(parameters) <= clusterParameterGroupMaxParamsBulkEdit {
					paramsToModify, parameters = parameters[:], nil
				} else {
					paramsToModify, parameters = parameters[:clusterParameterGroupMaxParamsBulkEdit], parameters[clusterParameterGroupMaxParamsBulkEdit:]
				}
				parameterGroupName := d.Id()
				modifyOpts := docdb.ModifyDBClusterParameterGroupInput{
					DBClusterParameterGroupName: aws.String(parameterGroupName),
					Parameters:                  paramsToModify,
				}

				_, err := conn.ModifyDBClusterParameterGroup(&modifyOpts)
				if err != nil {
					return fmt.Errorf("Error modifying DocDB Cluster Parameter Group: %w", err)
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DocumentDB Cluster Parameter Group (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceClusterParameterGroupRead(d, meta)
}

func resourceClusterParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DocDBConn

	deleteOpts := &docdb.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteDBClusterParameterGroup(deleteOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
			return nil
		}
		return err
	}

	return WaitForClusterParameterGroupDeletion(conn, d.Id())
}

func WaitForClusterParameterGroupDeletion(conn *docdb.DocDB, name string) error {
	params := &docdb.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(name),
	}

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeDBClusterParameterGroups(params)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("DocDB Parameter Group (%s) still exists", name))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeDBClusterParameterGroups(params)
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error deleting DocDB cluster parameter group: %s", err)
	}
	return nil
}
