package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

func resourceAwsQuickSightDataSource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsQuickSightDataSourceCreate,
		Read:   resourceAwsQuickSightDataSourceRead,
		Update: resourceAwsQuickSightDataSourceUpdate,
		Delete: resourceAwsQuickSightDataSourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"credentials": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credential_pair": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"username": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								}
							},
						},
					},
				},
			},

			"id": {
				Type:         schema.TypeString,
				Required:     true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"parameters": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amazon_elasticsearch": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								}
							},
						},
						"athena": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"work_group": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								}
							},
						},
						"aurora": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"aurora_postgre_sql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// TODO: Extract common
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"aws_iot_analytics": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								}
							},
						},
						"jira": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"site_base_url": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								}
							},
						},
						"maria_db": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"mysql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"postgresql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"presto": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog": {
										Type:         schema.TypeString,
										Required:     true,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"rds": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"instance_id": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								}
							},
						},
						"redshift": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cluster_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Optional:     true,
									},
									"port": {
										Type:         schema.TypeInt,
										Optional:     true,
									},
								}
							},
						},
						"s3": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"manifest_file_location": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
												},
												"key": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
												},
											}
										}
									}
								}
							},
						},
						"service_now": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"site_base_url": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								}
							},
						},
						"snowflake": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"warehouse": {
										Type:         schema.TypeString,
										Required:     true,
									},
								}
							},
						},
						"spark": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:         schema.TypeString,
										Required:     true,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"sql_server": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"teradata": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								}
							},
						},
						"twitter": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_rows": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"query": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								}
							},
						},
					},
				},
			},

			// TODO
			"permissions": {
			},

			// TODO
			"ssl_properties": {
			},

			// TODO
			"tags": {
			},

			// TODO
			"type": {
			},

			// TODO
			"vpc_connection_properties": {
			},

			// TODO: this is a tough one...
			"creation_stauts": {
				Type:     schema.TypeString,
			},
		},
	}
}

func resourceAwsQuickSightDataSourceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID := meta.(*AWSClient).accountid
	namespace := d.Get("namespace").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountID = v.(string)
	}

	createOpts := &quicksight.CreateDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		DataSourceName:    aws.String(d.Get("group_name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		createOpts.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateDataSource(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating QuickSight DataSource: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountID, namespace, aws.StringValue(resp.DataSource.DataSourceName)))

	return resourceAwsQuickSightDataSourceRead(d, meta)
}

func resourceAwsQuickSightDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	descOpts := &quicksight.DescribeDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		DataSourceName:    aws.String(groupName),
	}

	resp, err := conn.DescribeDataSource(descOpts)
	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight DataSource %s is already gone", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing QuickSight DataSource (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.DataSource.Arn)
	d.Set("aws_account_id", awsAccountID)
	d.Set("group_name", resp.DataSource.DataSourceName)
	d.Set("description", resp.DataSource.Description)
	d.Set("namespace", namespace)

	return nil
}

func resourceAwsQuickSightDataSourceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	updateOpts := &quicksight.UpdateDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		DataSourceName:    aws.String(groupName),
	}

	if v, ok := d.GetOk("description"); ok {
		updateOpts.Description = aws.String(v.(string))
	}

	_, err = conn.UpdateDataSource(updateOpts)
	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight DataSource %s is already gone", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error updating QuickSight DataSource %s: %s", d.Id(), err)
	}

	return resourceAwsQuickSightDataSourceRead(d, meta)
}

func resourceAwsQuickSightDataSourceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	deleteOpts := &quicksight.DeleteDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		DataSourceName:    aws.String(groupName),
	}

	if _, err := conn.DeleteDataSource(deleteOpts); err != nil {
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting QuickSight DataSource %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsQuickSightDataSourceParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/DATA_SOURCE_ID", id)
	}
	return parts[0], parts[1], parts[2], nil
}
