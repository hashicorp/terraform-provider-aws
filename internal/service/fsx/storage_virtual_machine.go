package fsx

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStorageVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFsxStorageVirtualMachineCreate,
		Read:   resourceAwsFsxStorageVirtualMachineRead,
		Update: resourceAwsFsxStorageVirtualMachineUpdate,
		Delete: resourceAwsFsxStorageVirtualMachineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 47),
			},
			"root_volume_security_style": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(fsx.StorageVirtualMachineRootVolumeSecurityStyle_Values(), false),
			},
			"svm_admin_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 50),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAwsFsxStorageVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateStorageVirtualMachineInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemId:       aws.String(d.Get("file_system_id").(string)),
		Name:               aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("svm_admin_password"); ok {
		input.SvmAdminPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("root_volume_security_style"); ok {
		input.RootVolumeSecurityStyle = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating FSx Storage Virtual Machine: %s", input)
	result, err := conn.CreateStorageVirtualMachine(input)

	if err != nil {
		return fmt.Errorf("error creating FSx Storage Virtual Machine: %w", err)
	}

	d.SetId(aws.StringValue(result.StorageVirtualMachine.StorageVirtualMachineId))

	if _, err := waitStorageVirtualMachineCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx Storage Virtual Machine (%s) create: %w", d.Id(), err)
	}

	return resourceAwsFsxStorageVirtualMachineRead(d, meta)
}

func resourceAwsFsxStorageVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	svm, err := FindStorageVirtualMachineByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Storage Virtual Machine (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx Storage Virtual Machine (%s): %w", d.Id(), err)
	}

	d.Set("arn", svm.ResourceARN)
	d.Set("file_system_id", svm.FileSystemId)
	d.Set("name", svm.Name)
	d.Set("root_volume_security_style", svm.RootVolumeSecurityStyle)
	d.Set("svm_admin_password", aws.String(d.Get("svm_admin_password").(string)))

	tags := KeyValueTags(svm.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsFsxStorageVirtualMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Storage Virtual Machine (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateStorageVirtualMachineInput{
			ClientRequestToken:      aws.String(resource.UniqueId()),
			StorageVirtualMachineId: aws.String(d.Id()),
		}

		if d.HasChange("svm_admin_password") {
			input.SvmAdminPassword = aws.String(d.Get("svm_admin_password").(string))
		}

		_, err := conn.UpdateStorageVirtualMachine(input)

		if err != nil {
			return fmt.Errorf("error updating FSx Storage Virtual Machine (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsFsxStorageVirtualMachineRead(d, meta)
}

func resourceAwsFsxStorageVirtualMachineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	log.Printf("[DEBUG] Deleting FSx Storage Virtual Machine: %s", d.Id())
	_, err := conn.DeleteStorageVirtualMachine(&fsx.DeleteStorageVirtualMachineInput{
		StorageVirtualMachineId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeStorageVirtualMachineNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FSx Storage Virtual Machine (%s): %w", d.Id(), err)
	}

	if _, err := waitStorageVirtualMachineDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx Storage Virtual Machine (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func flattenFsxSvmEndpoints(rs *fsx.SvmEndpoints) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.Iscsi != nil {
		m["iscsi"] = flattenFsxSvmEndpoint(rs.Iscsi)
	}

	if rs.Management != nil {
		m["management"] = flattenFsxSvmEndpoint(rs.Management)
	}

	if rs.Nfs != nil {
		m["nfs"] = flattenFsxSvmEndpoint(rs.Nfs)
	}

	if rs.Smb != nil {
		m["smb"] = flattenFsxSvmEndpoint(rs.Smb)
	}

	return []interface{}{m}
}

func flattenFsxSvmEndpoint(rs *fsx.SvmEndpoint) []interface{} {
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
