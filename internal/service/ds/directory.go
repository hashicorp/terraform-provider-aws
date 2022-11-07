package ds

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	directoryApplicationDeauthorizedPropagationTimeout = 2 * time.Minute
)

func ResourceDirectory() *schema.Resource {
	return &schema.Resource{
		Create: resourceDirectoryCreate,
		Read:   resourceDirectoryRead,
		Update: resourceDirectoryUpdate,
		Delete: resourceDirectoryDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"connect_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"connect_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"customer_dns_ips": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPAddress,
							},
						},
						"customer_username": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"desired_number_of_domain_controllers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(2),
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"edition": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(directoryservice.DirectoryEdition_Values(), false),
			},
			"enable_sso": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"size": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(directoryservice.DirectorySize_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      directoryservice.DirectoryTypeSimpleAd,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(directoryservice.DirectoryType_Values(), false),
			},
			"vpc_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	switch directoryType := d.Get("type").(string); directoryType {
	case directoryservice.DirectoryTypeAdconnector:
		input := &directoryservice.ConnectDirectoryInput{
			Name:     aws.String(name),
			Password: aws.String(d.Get("password").(string)),
			Tags:     Tags(tags.IgnoreAWS()),
		}

		if v, ok := d.GetOk("connect_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ConnectSettings = expandDirectoryConnectSettings(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("size"); ok {
			input.Size = aws.String(v.(string))
		} else {
			// Matching previous behavior of Default: "Large" for Size attribute.
			input.Size = aws.String(directoryservice.DirectorySizeLarge)
		}

		if v, ok := d.GetOk("short_name"); ok {
			input.ShortName = aws.String(v.(string))
		}

		output, err := conn.ConnectDirectory(input)

		if err != nil {
			return fmt.Errorf("creating Directory Service %s Directory (%s): %w", directoryType, name, err)
		}

		d.SetId(aws.StringValue(output.DirectoryId))

	case directoryservice.DirectoryTypeMicrosoftAd:
		input := &directoryservice.CreateMicrosoftADInput{
			Name:     aws.String(name),
			Password: aws.String(d.Get("password").(string)),
			Tags:     Tags(tags.IgnoreAWS()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("edition"); ok {
			input.Edition = aws.String(v.(string))
		}

		if v, ok := d.GetOk("short_name"); ok {
			input.ShortName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("vpc_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.VpcSettings = expandDirectoryVpcSettings(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.CreateMicrosoftAD(input)

		if err != nil {
			return fmt.Errorf("creating Directory Service %s Directory (%s): %w", directoryType, name, err)
		}

		d.SetId(aws.StringValue(output.DirectoryId))

	case directoryservice.DirectoryTypeSimpleAd:
		input := &directoryservice.CreateDirectoryInput{
			Name:     aws.String(name),
			Password: aws.String(d.Get("password").(string)),
			Tags:     Tags(tags.IgnoreAWS()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("size"); ok {
			input.Size = aws.String(v.(string))
		} else {
			// Matching previous behavior of Default: "Large" for Size attribute.
			input.Size = aws.String(directoryservice.DirectorySizeLarge)
		}

		if v, ok := d.GetOk("short_name"); ok {
			input.ShortName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("vpc_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.VpcSettings = expandDirectoryVpcSettings(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.CreateDirectory(input)

		if err != nil {
			return fmt.Errorf("creating Directory Service %s Directory (%s): %w", directoryType, name, err)
		}

		d.SetId(aws.StringValue(output.DirectoryId))
	}

	if _, err := waitDirectoryCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for Directory Service Directory (%s) create: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("alias"); ok {
		if err := createAlias(conn, d.Id(), v.(string)); err != nil {
			return err
		}
	}

	if v, ok := d.GetOk("desired_number_of_domain_controllers"); ok {
		if err := updateNumberOfDomainControllers(conn, d.Id(), v.(int), d.Timeout(schema.TimeoutCreate)); err != nil {
			return err
		}
	}

	if _, ok := d.GetOk("enable_sso"); ok {
		if err := enableSSO(conn, d.Id()); err != nil {
			return err
		}
	}

	return resourceDirectoryRead(d, meta)
}

func resourceDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dir, err := FindDirectoryByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Directory Service Directory (%s): %w", d.Id(), err)
	}

	d.Set("access_url", dir.AccessUrl)
	d.Set("alias", dir.Alias)
	if dir.ConnectSettings != nil {
		if err := d.Set("connect_settings", []interface{}{flattenDirectoryConnectSettingsDescription(dir.ConnectSettings, dir.DnsIpAddrs)}); err != nil {
			return fmt.Errorf("setting connect_settings: %w", err)
		}
	} else {
		d.Set("connect_settings", nil)
	}
	d.Set("description", dir.Description)
	d.Set("desired_number_of_domain_controllers", dir.DesiredNumberOfDomainControllers)
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("dns_ip_addresses", aws.StringValueSlice(dir.ConnectSettings.ConnectIps))
	} else {
		d.Set("dns_ip_addresses", aws.StringValueSlice(dir.DnsIpAddrs))
	}
	d.Set("edition", dir.Edition)
	d.Set("enable_sso", dir.SsoEnabled)
	d.Set("name", dir.Name)
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("security_group_id", dir.ConnectSettings.SecurityGroupId)
	} else {
		d.Set("security_group_id", dir.VpcSettings.SecurityGroupId)
	}
	d.Set("short_name", dir.ShortName)
	d.Set("size", dir.Size)
	d.Set("type", dir.Type)
	if dir.VpcSettings != nil {
		if err := d.Set("vpc_settings", []interface{}{flattenDirectoryVpcSettingsDescription(dir.VpcSettings)}); err != nil {
			return fmt.Errorf("setting vpc_settings: %w", err)
		}
	} else {
		d.Set("vpc_settings", nil)
	}

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("listing tags for Directory Service Directory (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn

	if d.HasChange("desired_number_of_domain_controllers") {
		if err := updateNumberOfDomainControllers(conn, d.Id(), d.Get("desired_number_of_domain_controllers").(int), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return err
		}
	}

	if d.HasChange("enable_sso") {
		if _, ok := d.GetOk("enable_sso"); ok {
			if err := enableSSO(conn, d.Id()); err != nil {
				return err
			}
		} else {
			if err := disableSSO(conn, d.Id()); err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating Directory Service Directory (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceDirectoryRead(d, meta)
}

func resourceDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn

	log.Printf("[DEBUG] Deleting Directory Service Directory: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(directoryApplicationDeauthorizedPropagationTimeout, func() (interface{}, error) {
		return conn.DeleteDirectory(&directoryservice.DeleteDirectoryInput{
			DirectoryId: aws.String(d.Id()),
		})
	}, directoryservice.ErrCodeClientException, "authorized applications")

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Directory Service Directory (%s): %w", d.Id(), err)
	}

	if _, err := waitDirectoryDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for Directory Service Directory (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func createAlias(conn *directoryservice.DirectoryService, directoryID, alias string) error {
	input := &directoryservice.CreateAliasInput{
		Alias:       aws.String(alias),
		DirectoryId: aws.String(directoryID),
	}

	_, err := conn.CreateAlias(input)

	if err != nil {
		return fmt.Errorf("creating Directory Service Directory (%s) alias (%s): %w", directoryID, alias, err)
	}

	return nil
}

func disableSSO(conn *directoryservice.DirectoryService, directoryID string) error {
	input := &directoryservice.DisableSsoInput{
		DirectoryId: aws.String(directoryID),
	}

	_, err := conn.DisableSso(input)

	if err != nil {
		return fmt.Errorf("disabling Directory Service Directory (%s) SSO: %w", directoryID, err)
	}

	return nil
}

func enableSSO(conn *directoryservice.DirectoryService, directoryID string) error {
	input := &directoryservice.EnableSsoInput{
		DirectoryId: aws.String(directoryID),
	}

	_, err := conn.EnableSso(input)

	if err != nil {
		return fmt.Errorf("enabling Directory Service Directory (%s) SSO: %w", directoryID, err)
	}

	return nil
}

func updateNumberOfDomainControllers(conn *directoryservice.DirectoryService, directoryID string, desiredNumber int, timeout time.Duration) error {
	oldDomainControllers, err := FindDomainControllers(conn, &directoryservice.DescribeDomainControllersInput{
		DirectoryId: aws.String(directoryID),
	})

	if err != nil {
		return fmt.Errorf("reading Directory Service Directory (%s) domain controllers: %w", directoryID, err)
	}

	input := &directoryservice.UpdateNumberOfDomainControllersInput{
		DesiredNumber: aws.Int64(int64(desiredNumber)),
		DirectoryId:   aws.String(directoryID),
	}

	_, err = conn.UpdateNumberOfDomainControllers(input)

	if err != nil {
		return fmt.Errorf("updating Directory Service Directory (%s) number of domain controllers (%d): %w", directoryID, desiredNumber, err)
	}

	newDomainControllers, err := FindDomainControllers(conn, &directoryservice.DescribeDomainControllersInput{
		DirectoryId: aws.String(directoryID),
	})

	if err != nil {
		return fmt.Errorf("reading Directory Service Directory (%s) domain controllers: %w", directoryID, err)
	}

	var wait []string

	for _, v := range newDomainControllers {
		domainControllerID := aws.StringValue(v.DomainControllerId)
		isNew := true

		for _, v := range oldDomainControllers {
			if aws.StringValue(v.DomainControllerId) == domainControllerID {
				isNew = false

				if aws.StringValue(v.Status) != directoryservice.DomainControllerStatusActive {
					wait = append(wait, domainControllerID)
				}
			}
		}

		if isNew {
			wait = append(wait, domainControllerID)
		}
	}

	for _, v := range wait {
		if len(newDomainControllers) > len(oldDomainControllers) {
			if _, err = waitDomainControllerCreated(conn, directoryID, v, timeout); err != nil {
				return fmt.Errorf("waiting for Directory Service Directory (%s) Domain Controller (%s) create: %w", directoryID, v, err)
			}
		} else {
			if _, err := waitDomainControllerDeleted(conn, directoryID, v, timeout); err != nil {
				return fmt.Errorf("waiting for Directory Service Directory (%s) Domain Controller (%s) delete: %w", directoryID, v, err)
			}
		}
	}

	return nil
}

func expandDirectoryConnectSettings(tfMap map[string]interface{}) *directoryservice.DirectoryConnectSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &directoryservice.DirectoryConnectSettings{}

	if v, ok := tfMap["customer_dns_ips"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CustomerDnsIps = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["customer_username"].(string); ok && v != "" {
		apiObject.CustomerUserName = aws.String(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["vpc_id"].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenDirectoryConnectSettingsDescription(apiObject *directoryservice.DirectoryConnectSettingsDescription, dnsIpAddrs []*string) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZones; v != nil {
		tfMap["availability_zones"] = aws.StringValueSlice(v)
	}

	if v := apiObject.ConnectIps; v != nil {
		tfMap["connect_ips"] = aws.StringValueSlice(v)
	}

	if dnsIpAddrs != nil {
		tfMap["customer_dns_ips"] = aws.StringValueSlice(dnsIpAddrs)
	}

	if v := apiObject.CustomerUserName; v != nil {
		tfMap["customer_username"] = aws.StringValue(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	return tfMap
}

func expandDirectoryVpcSettings(tfMap map[string]interface{}) *directoryservice.DirectoryVpcSettings { // nosemgrep:ci.caps5-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &directoryservice.DirectoryVpcSettings{}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["vpc_id"].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenDirectoryVpcSettings(apiObject *directoryservice.DirectoryVpcSettings) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDirectoryVpcSettingsDescription(apiObject *directoryservice.DirectoryVpcSettingsDescription) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZones; v != nil {
		tfMap["availability_zones"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	return tfMap
}
