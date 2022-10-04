package fsx

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
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

func ResourceOntapStorageVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Create: resourceOntapStorageVirtualMachineCreate,
		Read:   resourceOntapStorageVirtualMachineRead,
		Update: resourceOntapStorageVirtualMachineUpdate,
		Delete: resourceOntapStorageVirtualMachineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    ResourceOntapStorageVirtualMachineV0().CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceOntapStorageVirtualMachineStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active_directory_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"netbios_name": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							ValidateFunc: validation.StringLenBetween(1, 15),
						},
						"self_managed_active_directory_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_ips": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										MaxItems: 3,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.IsIPAddress,
										},
									},
									"domain_name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"file_system_administrators_group": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"organizational_unit_distinguished_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 2000),
									},
									"password": {
										Type:         schema.TypeString,
										Sensitive:    true,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"username": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
					},
				},
			},
			"endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iscsi": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"management": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"nfs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"smb": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"file_system_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(11, 21),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 47),
			},
			"root_volume_security_style": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.StorageVirtualMachineRootVolumeSecurityStyle_Values(), false),
			},
			"subtype": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"svm_admin_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 50),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOntapStorageVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateStorageVirtualMachineInput{
		FileSystemId: aws.String(d.Get("file_system_id").(string)),
		Name:         aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("active_directory_configuration"); ok {
		input.ActiveDirectoryConfiguration = expandOntapSvmActiveDirectoryConfigurationCreate(v.([]interface{}))
	}

	if v, ok := d.GetOk("root_volume_security_style"); ok {
		input.RootVolumeSecurityStyle = aws.String(v.(string))
	}

	if v, ok := d.GetOk("svm_admin_password"); ok {
		input.SvmAdminPassword = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating FSx ONTAP Storage Virtual Machine: %s", input)
	result, err := conn.CreateStorageVirtualMachine(input)

	if err != nil {
		return fmt.Errorf("error creating FSx Storage Virtual System: %w", err)
	}

	d.SetId(aws.StringValue(result.StorageVirtualMachine.StorageVirtualMachineId))

	if _, err := waitStorageVirtualMachineCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx Storage Virtual Machine (%s) create: %w", d.Id(), err)
	}

	return resourceOntapStorageVirtualMachineRead(d, meta)

}

func resourceOntapStorageVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	storageVirtualMachine, err := FindStorageVirtualMachineByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx ONTAP Storage Virtual Machine (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx ONTAP Storage Virtual Machine (%s): %w", d.Id(), err)
	}

	d.Set("arn", storageVirtualMachine.ResourceARN)
	d.Set("name", storageVirtualMachine.Name)
	d.Set("file_system_id", storageVirtualMachine.FileSystemId)
	//RootVolumeSecurityStyle and SVMAdminPassword are write only properties so they don't get returned from the describe API so we just store the original setting to state
	d.Set("root_volume_security_style", d.Get("root_volume_security_style").(string))
	d.Set("svm_admin_password", d.Get("svm_admin_password").(string))
	d.Set("subtype", storageVirtualMachine.Subtype)
	d.Set("uuid", storageVirtualMachine.UUID)

	if err := d.Set("active_directory_configuration", flattenOntapSvmActiveDirectoryConfiguration(d, storageVirtualMachine.ActiveDirectoryConfiguration)); err != nil {
		return fmt.Errorf("error setting svm_active_directory: %w", err)
	}

	if err := d.Set("endpoints", flattenOntapStorageVirtualMachineEndpoints(storageVirtualMachine.Endpoints)); err != nil {
		return fmt.Errorf("error setting endpoints: %w", err)
	}

	//SVM tags do not get returned with describe call so need to make a separate list tags call
	tags, tagserr := ListTags(conn, *storageVirtualMachine.ResourceARN)

	if tagserr != nil {
		return fmt.Errorf("error reading Tags for FSx ONTAP Storage Virtual Machine (%s): %w", d.Id(), err)
	} else {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceOntapStorageVirtualMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx ONTAP Storage Virtual Machine (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateStorageVirtualMachineInput{
			ClientRequestToken:      aws.String(resource.UniqueId()),
			StorageVirtualMachineId: aws.String(d.Id()),
		}

		if d.HasChange("active_directory_configuration") {
			input.ActiveDirectoryConfiguration = expandOntapSvmActiveDirectoryConfigurationUpdate(d.Get("active_directory_configuration").([]interface{}))
		}

		if d.HasChange("svm_admin_password") {
			input.SvmAdminPassword = aws.String(d.Get("svm_admin_password").(string))
		}

		_, err := conn.UpdateStorageVirtualMachine(input)

		if err != nil {
			return fmt.Errorf("error updating FSx ONTAP Storage Virtual Machine (%s): %w", d.Id(), err)
		}

		if _, err := waitStorageVirtualMachineUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for FSx ONTAP Storage Virtual Machine (%s) update: %w", d.Id(), err)
		}
	}

	return resourceOntapStorageVirtualMachineRead(d, meta)
}

func resourceOntapStorageVirtualMachineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	log.Printf("[DEBUG] Deleting FSx ONTAP Storage Virtual Machine: %s", d.Id())
	_, err := conn.DeleteStorageVirtualMachine(&fsx.DeleteStorageVirtualMachineInput{
		StorageVirtualMachineId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeStorageVirtualMachineNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FSx ONTAP Storage Virtual Machine (%s): %w", d.Id(), err)
	}

	if _, err := waitStorageVirtualMachineDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx ONTAP Storage Virtual Machine (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandOntapSvmActiveDirectoryConfigurationCreate(cfg []interface{}) *fsx.CreateSvmActiveDirectoryConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.CreateSvmActiveDirectoryConfiguration{}

	if v, ok := conf["netbios_name"].(string); ok && len(v) > 0 {
		out.NetBiosName = aws.String(v)
	}

	if v, ok := conf["self_managed_active_directory_configuration"].([]interface{}); ok {
		out.SelfManagedActiveDirectoryConfiguration = expandOntapSvmSelfManagedActiveDirectoryConfiguration(v)
	}

	return &out
}

func expandOntapSvmSelfManagedActiveDirectoryConfiguration(cfg []interface{}) *fsx.SelfManagedActiveDirectoryConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.SelfManagedActiveDirectoryConfiguration{}

	if v, ok := conf["dns_ips"].(*schema.Set); ok {
		out.DnsIps = flex.ExpandStringSet(v)
	}

	if v, ok := conf["domain_name"].(string); ok && len(v) > 0 {
		out.DomainName = aws.String(v)
	}

	if v, ok := conf["file_system_administrators_group"].(string); ok && len(v) > 0 {
		out.FileSystemAdministratorsGroup = aws.String(v)
	}

	if v, ok := conf["organizational_unit_distinguished_name"].(string); ok && len(v) > 0 {
		out.OrganizationalUnitDistinguishedName = aws.String(v)
	}

	if v, ok := conf["password"].(string); ok && len(v) > 0 {
		out.Password = aws.String(v)
	}

	if v, ok := conf["username"].(string); ok && len(v) > 0 {
		out.UserName = aws.String(v)
	}

	return &out
}

func expandOntapSvmActiveDirectoryConfigurationUpdate(cfg []interface{}) *fsx.UpdateSvmActiveDirectoryConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.UpdateSvmActiveDirectoryConfiguration{}

	if v, ok := conf["self_managed_active_directory_configuration"].([]interface{}); ok {
		out.SelfManagedActiveDirectoryConfiguration = expandOntapSvmSelfManagedActiveDirectoryConfigurationUpdate(v)
	}

	return &out
}

func expandOntapSvmSelfManagedActiveDirectoryConfigurationUpdate(cfg []interface{}) *fsx.SelfManagedActiveDirectoryConfigurationUpdates {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.SelfManagedActiveDirectoryConfigurationUpdates{}

	if v, ok := conf["dns_ips"].(*schema.Set); ok {
		out.DnsIps = flex.ExpandStringSet(v)
	}

	if v, ok := conf["password"].(string); ok {
		out.Password = aws.String(v)
	}

	if v, ok := conf["username"].(string); ok {
		out.UserName = aws.String(v)
	}

	return &out
}

func flattenOntapSvmActiveDirectoryConfiguration(d *schema.ResourceData, rs *fsx.SvmActiveDirectoryConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.NetBiosName != nil {
		m["netbios_name"] = rs.NetBiosName
	}

	if rs.SelfManagedActiveDirectoryConfiguration != nil {
		m["self_managed_active_directory_configuration"] = flattenOntapSelfManagedActiveDirectoryConfiguration(d, rs.SelfManagedActiveDirectoryConfiguration)
	}

	return []interface{}{m}
}

func flattenOntapSelfManagedActiveDirectoryConfiguration(d *schema.ResourceData, rs *fsx.SelfManagedActiveDirectoryAttributes) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.DnsIps != nil {
		m["dns_ips"] = aws.StringValueSlice(rs.DnsIps)
	}

	if rs.DomainName != nil {
		m["domain_name"] = aws.StringValue(rs.DomainName)
	}

	if rs.FileSystemAdministratorsGroup != nil {
		m["file_system_administrators_group"] = aws.StringValue(rs.FileSystemAdministratorsGroup)
	}

	if rs.OrganizationalUnitDistinguishedName != nil {
		if _, ok := d.GetOk("active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name"); ok {
			m["organizational_unit_distinguished_name"] = aws.StringValue(rs.OrganizationalUnitDistinguishedName)
		}
	}

	if rs.UserName != nil {
		m["username"] = aws.StringValue(rs.UserName)
	}

	// Since we are in a configuration block and the FSx API does not return
	// the password, we need to set the value if we can or Terraform will
	// show a difference for the argument from empty string to the value.
	// This is not a pattern that should be used normally.
	// See also: flattenEmrKerberosAttributes
	m["password"] = d.Get("active_directory_configuration.0.self_managed_active_directory_configuration.0.password").(string)

	return []interface{}{m}
}

func flattenOntapStorageVirtualMachineEndpoints(rs *fsx.SvmEndpoints) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.Iscsi != nil {
		m["iscsi"] = flattenOntapStorageVirtualMachineEndpoint(rs.Iscsi)
	}
	if rs.Management != nil {
		m["management"] = flattenOntapStorageVirtualMachineEndpoint(rs.Management)
	}
	if rs.Nfs != nil {
		m["nfs"] = flattenOntapStorageVirtualMachineEndpoint(rs.Nfs)
	}
	if rs.Smb != nil {
		m["smb"] = flattenOntapStorageVirtualMachineEndpoint(rs.Smb)
	}
	return []interface{}{m}
}

func flattenOntapStorageVirtualMachineEndpoint(rs *fsx.SvmEndpoint) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.DNSName != nil {
		m["dns_name"] = aws.StringValue(rs.DNSName)
	}
	if rs.IpAddresses != nil {
		m["ip_addresses"] = flex.FlattenStringSet(rs.IpAddresses)
	}

	return []interface{}{m}
}
