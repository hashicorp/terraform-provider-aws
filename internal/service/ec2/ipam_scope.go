package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMScope() *schema.Resource {
	return &schema.Resource{
		Create:        ResourceIPAMScopeCreate,
		Read:          ResourceIPAMScopeRead,
		Update:        ResourceIPAMScopeUpdate,
		Delete:        ResourceIPAMScopeDelete,
		CustomizeDiff: customdiff.Sequence(verify.SetTagsDiff),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_scope_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pool_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	ipamScopeCreateTimeout = 3 * time.Minute
	ipamScopeCreateDeley   = 5 * time.Second
	IPAMScopeDeleteTimeout = 3 * time.Minute
	ipamScopeDeleteDelay   = 5 * time.Second

	invalidIPAMScopeIDNotFound = "InvalidIpamScopeId.NotFound"
)

func ResourceIPAMScopeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateIpamScopeInput{
		ClientToken:       aws.String(resource.UniqueId()),
		IpamId:            aws.String(d.Get("ipam_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, "ipam-scope"),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating IPAM Scope: %s", input)
	output, err := conn.CreateIpamScope(input)
	if err != nil {
		return fmt.Errorf("Error creating ipam scope in ipam (%s): %w", d.Get("ipam_id").(string), err)
	}
	d.SetId(aws.StringValue(output.IpamScope.IpamScopeId))
	log.Printf("[INFO] IPAM Scope ID: %s", d.Id())

	if _, err = waitIPAMScopeAvailable(conn, d.Id(), ipamScopeCreateTimeout); err != nil {
		return fmt.Errorf("error waiting for IPAM Scope (%s) to be Available: %w", d.Id(), err)
	}

	return ResourceIPAMScopeRead(d, meta)
}

func ResourceIPAMScopeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	scope, err := findIPAMScopeById(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Scope (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IPAM Scope (%s): %w", d.Id(), err)
	}

	ipamId := strings.Split(aws.StringValue(scope.IpamArn), "/")[1]

	d.Set("arn", scope.IpamScopeArn)
	d.Set("description", scope.Description)
	d.Set("ipam_arn", scope.IpamArn)
	d.Set("ipam_id", ipamId)
	d.Set("ipam_scope_type", scope.IpamScopeType)
	d.Set("is_default", scope.IsDefault)
	d.Set("pool_count", scope.PoolCount)

	tags := KeyValueTags(scope.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func ResourceIPAMScopeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	if d.HasChange("description") {
		// moved `ModifyIpamScope` call here due to bug during development, can likely be moved out of if statement scope later
		input := &ec2.ModifyIpamScopeInput{
			IpamScopeId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}
		log.Printf("[DEBUG] Updating IPAM scope: %s", input)
		_, err := conn.ModifyIpamScope(input)
		if err != nil {
			return fmt.Errorf("error updating IPAM Scope (%s): %w", d.Id(), err)
		}
	}

	return ResourceIPAMScopeRead(d, meta)
}

func ResourceIPAMScopeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting IPAM Scope: %s", d.Id())
	_, err := conn.DeleteIpamScope(&ec2.DeleteIpamScopeInput{
		IpamScopeId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting IPAM Scope: (%s): %w", d.Id(), err)
	}

	if _, err = WaitIPAMScopeDeleted(conn, d.Id(), IPAMScopeDeleteTimeout); err != nil {
		if tfresource.NotFound(err) {
			return nil
		}
		return fmt.Errorf("error waiting for IPAM Scope (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func findIPAMScopeById(conn *ec2.EC2, id string) (*ec2.IpamScope, error) {
	input := &ec2.DescribeIpamScopesInput{
		IpamScopeIds: aws.StringSlice([]string{id}),
	}

	var results []*ec2.IpamScope

	err := conn.DescribeIpamScopesPages(input, func(page *ec2.DescribeIpamScopesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, scope := range page.IpamScopes {
			if scope == nil {
				continue
			}

			results = append(results, scope)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, invalidIPAMScopeIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(results); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return results[0], nil
}

func waitIPAMScopeAvailable(conn *ec2.EC2, ipamScopeId string, timeout time.Duration) (*ec2.IpamScope, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.IpamScopeStateCreateInProgress},
		Target:  []string{ec2.IpamScopeStateCreateComplete},
		Refresh: statusIPAMScopeStatus(conn, ipamScopeId),
		Timeout: timeout,
		Delay:   ipamScopeCreateDeley,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.IpamScope); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMScopeDeleted(conn *ec2.EC2, ipamScopeId string, timeout time.Duration) (*ec2.IpamScope, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.IpamScopeStateCreateComplete, ec2.IpamScopeStateModifyComplete},
		Target:  []string{invalidIPAMScopeIDNotFound, ec2.IpamScopeStateDeleteComplete},
		Refresh: statusIPAMScopeStatus(conn, ipamScopeId),
		Timeout: timeout,
		Delay:   ipamScopeDeleteDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.IpamScope); ok {
		return output, err
	}

	return nil, err
}

func statusIPAMScopeStatus(conn *ec2.EC2, ipamScopeId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		output, err := findIPAMScopeById(conn, ipamScopeId)

		if tfresource.NotFound(err) {
			return output, invalidIPAMScopeIDNotFound, nil
		}

		// there was an unhandled error in the Finder
		if err != nil {
			return nil, "", err
		}

		return output, ec2.IpamScopeStateCreateComplete, nil
	}
}
