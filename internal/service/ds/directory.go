package ds

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
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

		Schema: map[string]*schema.Schema{
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
			"size": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					directoryservice.DirectorySizeLarge,
					directoryservice.DirectorySizeSmall,
				}, false),
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"connect_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connect_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"customer_username": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"customer_dns_ips": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPAddress,
							},
							Set: schema.HashString,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"enable_sso": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  directoryservice.DirectoryTypeSimpleAd,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					directoryservice.DirectoryTypeAdconnector,
					directoryservice.DirectoryTypeMicrosoftAd,
					directoryservice.DirectoryTypeSimpleAd,
					directoryservice.DirectoryTypeSharedMicrosoftAd,
				}, false),
			},
			"edition": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					directoryservice.DirectoryEditionEnterprise,
					directoryservice.DirectoryEditionStandard,
				}, false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func buildVpcSettings(d *schema.ResourceData) (vpcSettings *directoryservice.DirectoryVpcSettings, err error) {
	v, ok := d.GetOk("vpc_settings")
	if !ok {
		return nil, fmt.Errorf("vpc_settings is required for type = SimpleAD or MicrosoftAD")
	}
	settings := v.([]interface{})
	s := settings[0].(map[string]interface{})
	var subnetIds []*string
	for _, id := range s["subnet_ids"].(*schema.Set).List() {
		subnetIds = append(subnetIds, aws.String(id.(string)))
	}

	vpcSettings = &directoryservice.DirectoryVpcSettings{
		SubnetIds: subnetIds,
		VpcId:     aws.String(s["vpc_id"].(string)),
	}

	return vpcSettings, nil
}

func buildConnectSettings(d *schema.ResourceData) (connectSettings *directoryservice.DirectoryConnectSettings, err error) {
	v, ok := d.GetOk("connect_settings")
	if !ok {
		return nil, fmt.Errorf("connect_settings is required for type = ADConnector")
	}
	settings := v.([]interface{})
	s := settings[0].(map[string]interface{})

	var subnetIds []*string
	for _, id := range s["subnet_ids"].(*schema.Set).List() {
		subnetIds = append(subnetIds, aws.String(id.(string)))
	}

	var customerDnsIps []*string
	for _, id := range s["customer_dns_ips"].(*schema.Set).List() {
		customerDnsIps = append(customerDnsIps, aws.String(id.(string)))
	}

	connectSettings = &directoryservice.DirectoryConnectSettings{
		CustomerDnsIps:   customerDnsIps,
		CustomerUserName: aws.String(s["customer_username"].(string)),
		SubnetIds:        subnetIds,
		VpcId:            aws.String(s["vpc_id"].(string)),
	}

	return connectSettings, nil
}

func createDirectoryConnector(conn *directoryservice.DirectoryService, d *schema.ResourceData, meta interface{}) (directoryId string, err error) {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := directoryservice.ConnectDirectoryInput{
		Name:     aws.String(d.Get("name").(string)),
		Password: aws.String(d.Get("password").(string)),
		Tags:     Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("size"); ok {
		input.Size = aws.String(v.(string))
	} else {
		// Matching previous behavior of Default: "Large" for Size attribute
		input.Size = aws.String(directoryservice.DirectorySizeLarge)
	}
	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}

	input.ConnectSettings, err = buildConnectSettings(d)
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] Creating Directory Connector: %s", input)
	out, err := conn.ConnectDirectory(&input)
	if err != nil {
		return "", err
	}
	log.Printf("[DEBUG] Directory Connector created: %s", out)

	return *out.DirectoryId, nil
}

func createSimple(conn *directoryservice.DirectoryService, d *schema.ResourceData, meta interface{}) (directoryId string, err error) {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := directoryservice.CreateDirectoryInput{
		Name:     aws.String(d.Get("name").(string)),
		Password: aws.String(d.Get("password").(string)),
		Tags:     Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("size"); ok {
		input.Size = aws.String(v.(string))
	} else {
		// Matching previous behavior of Default: "Large" for Size attribute
		input.Size = aws.String(directoryservice.DirectorySizeLarge)
	}
	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}

	input.VpcSettings, err = buildVpcSettings(d)
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] Creating Simple Directory Service: %s", input)
	out, err := conn.CreateDirectory(&input)
	if err != nil {
		return "", err
	}
	log.Printf("[DEBUG] Simple Directory Service created: %s", out)

	return *out.DirectoryId, nil
}

func createActive(conn *directoryservice.DirectoryService, d *schema.ResourceData, meta interface{}) (directoryId string, err error) {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := directoryservice.CreateMicrosoftADInput{
		Name:     aws.String(d.Get("name").(string)),
		Password: aws.String(d.Get("password").(string)),
		Tags:     Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("edition"); ok {
		input.Edition = aws.String(v.(string))
	}

	input.VpcSettings, err = buildVpcSettings(d)
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] Creating Microsoft AD Directory Service: %s", input)
	out, err := conn.CreateMicrosoftAD(&input)
	if err != nil {
		return "", err
	}
	log.Printf("[DEBUG] Microsoft AD Directory Service created: %s", out)

	return *out.DirectoryId, nil
}

func enableSSO(conn *directoryservice.DirectoryService, d *schema.ResourceData) error {
	if v, ok := d.GetOk("enable_sso"); ok && v.(bool) {
		log.Printf("[DEBUG] Enabling SSO for DS directory %q", d.Id())
		if _, err := conn.EnableSso(&directoryservice.EnableSsoInput{
			DirectoryId: aws.String(d.Id()),
		}); err != nil {
			return fmt.Errorf("Error Enabling SSO for DS directory %s: %s", d.Id(), err)
		}
	} else {
		log.Printf("[DEBUG] Disabling SSO for DS directory %q", d.Id())
		if _, err := conn.DisableSso(&directoryservice.DisableSsoInput{
			DirectoryId: aws.String(d.Id()),
		}); err != nil {
			return fmt.Errorf("Error Disabling SSO for DS directory %s: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn

	var directoryId string
	var err error
	directoryType := d.Get("type").(string)

	if directoryType == directoryservice.DirectoryTypeAdconnector {
		directoryId, err = createDirectoryConnector(conn, d, meta)
	} else if directoryType == directoryservice.DirectoryTypeMicrosoftAd {
		directoryId, err = createActive(conn, d, meta)
	} else if directoryType == directoryservice.DirectoryTypeSimpleAd {
		directoryId, err = createSimple(conn, d, meta)
	}

	if err != nil {
		return err
	}

	d.SetId(directoryId)

	_, err = waitDirectoryCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Directory Service Directory (%s) to create: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("alias"); ok {
		input := directoryservice.CreateAliasInput{
			DirectoryId: aws.String(d.Id()),
			Alias:       aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Assigning alias %q to DS directory %q",
			v.(string), d.Id())
		out, err := conn.CreateAlias(&input)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Alias %q assigned to DS directory %q",
			*out.Alias, *out.DirectoryId)
	}

	if d.HasChange("enable_sso") {
		if err := enableSSO(conn, d); err != nil {
			return err
		}
	}

	return resourceDirectoryRead(d, meta)
}

func resourceDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn

	if d.HasChange("enable_sso") {
		if err := enableSSO(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Directory Service Directory (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDirectoryRead(d, meta)
}

func resourceDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dir, err := findDirectoryByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Directory Service Directory (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Received DS directory: %s", dir)

	d.Set("access_url", dir.AccessUrl)
	d.Set("alias", dir.Alias)
	d.Set("description", dir.Description)

	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("dns_ip_addresses", flex.FlattenStringSet(dir.ConnectSettings.ConnectIps))
	} else {
		d.Set("dns_ip_addresses", flex.FlattenStringSet(dir.DnsIpAddrs))
	}
	d.Set("name", dir.Name)
	d.Set("short_name", dir.ShortName)
	d.Set("size", dir.Size)
	d.Set("edition", dir.Edition)
	d.Set("type", dir.Type)

	if err := d.Set("vpc_settings", flattenVPCSettings(dir.VpcSettings)); err != nil {
		return fmt.Errorf("error setting VPC settings: %s", err)
	}

	if err := d.Set("connect_settings", flattenConnectSettings(dir.DnsIpAddrs, dir.ConnectSettings)); err != nil {
		return fmt.Errorf("error setting connect settings: %s", err)
	}

	d.Set("enable_sso", dir.SsoEnabled)

	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("security_group_id", dir.ConnectSettings.SecurityGroupId)
	} else {
		d.Set("security_group_id", dir.VpcSettings.SecurityGroupId)
	}

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Directory Service Directory (%s): %s", d.Id(), err)
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

func resourceDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn

	input := &directoryservice.DeleteDirectoryInput{
		DirectoryId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Directory Service Directory: (%s)", d.Id())
	err := resource.Retry(directoryApplicationDeauthorizedPropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteDirectory(input)

		if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
			return nil
		}
		if tfawserr.ErrMessageContains(err, directoryservice.ErrCodeClientException, "authorized applications") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDirectory(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting Directory Service Directory (%s): %w", d.Id(), err)
	}

	_, err = waitDirectoryDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Directory Service Directory (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
