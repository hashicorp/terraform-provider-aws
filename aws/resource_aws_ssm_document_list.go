package aws

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSsmDocumentList() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmDocumentListCreate,
		Read:   resourceAwsSsmDocumentListRead,
		Update: resourceAwsSsmDocumentListUpdate,
		Delete: resourceAwsSsmDocumentListDelete,

		Schema: map[string]*schema.Schema{
			"documents_hash": {
				Type:     schema.TypeString,
				Required: true,
			},
			"documents_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsSSMDocumentType,
			},
			"documents_permissions": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"account_ids": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"documents_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, o, n string, d *schema.ResourceData) bool {
								return true // id is a hash of name and content, so showing it as a diff is meaningless
							},
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"content": {
							Type:     schema.TypeString,
							Optional: true,
							StateFunc: func(v interface{}) string {
								switch v.(type) {
								case string:
									return contentHashSum(v.(string))
								default:
									return ""
								}
							},
						},
						"schema_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"default_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hash": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hash_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"latest_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"platform_types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"parameter": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"default_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"description": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func contentHashSum(content string) string {
	hash := sha1.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

func GetDocumentListItemKey(index int, name string) string {
	return fmt.Sprintf("documents_list.%d.%s", index, name)
}

func GetDocumentListItemValue(d *schema.ResourceData, index int, name string) (interface{}, error) {
	key := GetDocumentListItemKey(index, name)

	if v, ok := d.GetOk(key); ok && v != nil {
		return v, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Item %q not found", key))
	}
}

func resourceAwsSsmDocumentListCreate(d *schema.ResourceData, meta interface{}) error {

	// Potentially creating a large number of documents, so use partial state
	id := d.Get("documents_hash").(string)
	d.SetId(id)
	d.Partial(true) // make sure we record the id even if the rest of this gets interrupted
	d.Set("id", id)
	d.SetPartial("id")
	d.SetPartial("documents_hash")
	d.SetPartial("documents_type")

	documents := d.Get("documents_list").([]interface{})
	partialDocList := make([]interface{}, 0)

	for i, v := range documents {
		log.Printf("[INFO] SSM Document Number: %d", i)

		document := v.(map[string]interface{})

		name, err := GetDocumentListItemValue(d, i, "name")
		if err != nil {
			return errwrap.Wrapf("[ERROR] Error retrieving name for SSM document: {{err}}", err)
		}

		content, err := GetDocumentListItemValue(d, i, "content")
		if err != nil {
			return errwrap.Wrapf("[ERROR] Error retrieving content for SSM document: {{err}}", err)
		}

		if err := CreateSSMDocument(name.(string), content.(string), d.Get("documents_type").(string), meta); err != nil {
			return err
		}

		// Store the hash of the content
		document["content"] = contentHashSum(document["content"].(string))

		// Document was created successfully, so save partial state
		partialDocList = append(partialDocList, document)
		d.Set("documents_list", partialDocList)
		d.SetPartial("documents_list")

		// if v, ok := d.GetOk("documents_permissions"); ok && v != nil {
		// 	if err := setDocumentPermissions(d, meta); err != nil {
		// 		return err
		// 	}
		// } else {
		// 	log.Printf("[DEBUG] Not setting permissions for %q", d.Id())
		// }

	}
	d.Partial(false)

	return resourceAwsSsmDocumentListRead(d, meta)
}

func resourceAwsSsmDocumentListRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	docsFromState := d.Get("documents_list").([]interface{})
	docsFromAWS := make([]interface{}, 0)

	for _, v := range docsFromState {

		document := v.(map[string]interface{})

		log.Printf("[INFO] Reading SSM Document: %s", document["name"].(string))

		docInput := &ssm.DescribeDocumentInput{
			Name: aws.String(document["name"].(string)),
		}

		resp, err := ssmconn.DescribeDocument(docInput)
		if err != nil {
			if ssmErr, ok := err.(awserr.Error); ok && ssmErr.Code() == "InvalidDocument" {
				log.Printf("[WARN] SSM Document not found so removing from state")
			} else {
				return errwrap.Wrapf("[ERROR] Error describing SSM document: {{err}}", err)
			}

		} else {
			doc := resp.Document

			document["created_date"] = (*doc.CreatedDate).String()

			document["default_version"] = *doc.DefaultVersion
			document["description"] = *doc.Description
			document["schema_version"] = *doc.SchemaVersion

			document["hash"] = *doc.Hash
			document["hash_type"] = *doc.HashType
			document["latest_version"] = *doc.LatestVersion
			document["name"] = *doc.Name

			document["owner"] = *doc.Owner
			document["platform_types"] = flattenStringList(doc.PlatformTypes)
			document["arn"] = flattenAwsSsmDocumentArn(meta, doc.Name)

			document["status"] = *doc.Status

			// gp, err := getDocumentPermissions(d, meta)

			// if err != nil {
			// 	return errwrap.Wrapf("[ERROR] Error reading SSM document permissions: {{err}}", err)
			// }

			// d.Set("documents_permissions", gp)

			// params := make([]map[string]interface{}, 0)
			// for i := 0; i < len(doc.Parameters); i++ {

			// 	dp := doc.Parameters[i]
			// 	param := make(map[string]interface{})

			// 	if dp.DefaultValue != nil {
			// 		param["default_value"] = *dp.DefaultValue
			// 	}
			// 	if dp.Description != nil {
			// 		param["description"] = *dp.Description
			// 	}
			// 	if dp.Name != nil {
			// 		param["name"] = *dp.Name
			// 	}
			// 	if dp.Type != nil {
			// 		param["type"] = *dp.Type
			// 	}
			// 	params = append(params, param)
			// }

			// if len(params) == 0 {
			// 	params = make([]map[string]interface{}, 1)
			// }

			// if err := d.Set("parameter", params); err != nil {
			// 	return err
			// }

			docsFromAWS = append(docsFromAWS, document)
		}
	}

	if err := d.Set("documents_list", docsFromAWS); err != nil {
		return err
	}

	return nil
}

func resourceAwsSsmDocumentListUpdate(d *schema.ResourceData, meta interface{}) error {

	id := d.Get("documents_hash").(string)
	d.SetId(id)

	d.Partial(true) // make sure we record the id even if the rest of this gets interrupted
	d.Set("id", id)
	d.SetPartial("id")
	d.SetPartial("documents_hash")
	d.SetPartial("documents_type")

	if d.HasChange("documents_list") {
		log.Printf("[INFO] Has changes")

		o, n := d.GetChange("documents_list")

		os := o.([]interface{})
		ns := n.([]interface{})

		oContents := make(map[string]string)
		oDocuments := make(map[string]map[string]interface{})
		partialDocList := make(map[string]map[string]interface{})
		for _, v := range os {
			document := v.(map[string]interface{})
			name := document["name"].(string)
			log.Printf("[INFO] Old Docs: %s", name)
			oContents[name] = document["content"].(string)
			oDocuments[name] = document
			partialDocList[name] = document
		}

		nContents := make(map[string]string)
		nDocuments := make(map[string]map[string]interface{})
		for _, v := range ns {
			document := v.(map[string]interface{})
			name := document["name"].(string)
			log.Printf("[INFO] New Docs: %s", name)
			nContents[name] = document["content"].(string)
			nDocuments[name] = document
		}

		// Handle deleted documents
		for name := range oDocuments {
			_, ok := nDocuments[name]
			if !ok {
				if err := DeleteSSMDocument(name, meta); err != nil {
					return err
				}
				delete(partialDocList, name)
				d.Set("documents_list", ConvertPartialDocListMapToSlice(partialDocList))
				d.SetPartial("documents_list")
			}
		}

		// Handle new and updated documents
		for name := range nDocuments {
			log.Printf("[INFO] Processing SSM Document: %s", name)

			_, ok := oDocuments[name]
			if !ok {
				if err := CreateSSMDocument(name, nContents[name], d.Get("documents_type").(string), meta); err != nil {
					return err
				}
				nDocuments[name]["content"] = contentHashSum(nDocuments[name]["content"].(string))
			} else {
				if oContents[name] != nContents[name] {

					doc := oDocuments[name]                      // Get the existing doc from state. This brings across the default version etc
					doc["content"] = nDocuments[name]["content"] // Then add the new content
					if err := UpdateSSMDocument(doc, meta); err != nil {
						return err
					}
					nDocuments[name]["content"] = contentHashSum(nDocuments[name]["content"].(string))
				}
			}
			partialDocList[name] = nDocuments[name]
			d.Set("documents_list", ConvertPartialDocListMapToSlice(partialDocList))
			d.SetPartial("documents_list")
		}

	} else {
		log.Printf("[INFO] No changes")
	}

	// 	if _, ok := d.GetOk("permissions"); ok {
	// 	if err := setDocumentPermissions(d, meta); err != nil {
	// 		return err
	// 	}
	// } else {
	// 	log.Printf("[DEBUG] Not setting document permissions on %q", d.Id())
	// }

	//	return nil
	d.Partial(false)
	d.SetId(d.Get("documents_hash").(string))

	return resourceAwsSsmDocumentListRead(d, meta)
}

func resourceAwsSsmDocumentListDelete(d *schema.ResourceData, meta interface{}) error {

	documents := d.Get("documents_list").([]interface{})

	for i, _ := range documents {
		log.Printf("[INFO] SSM Document Number: %d", i)

		name, err := GetDocumentListItemValue(d, i, "name")
		if err != nil {
			return errwrap.Wrapf("[ERROR] Error retrieving name for SSM document: {{err}}", err)
		}

		// if err := deleteDocumentPermissions(d, meta); err != nil {
		// 	return err
		// }

		if err := DeleteSSMDocument(name.(string), meta); err != nil {
			return err
		}

	}
	d.SetId("")
	return nil
}

func ConvertPartialDocListMapToSlice(m map[string]map[string]interface{}) []map[string]interface{} {

	v := make([]map[string]interface{}, 0, len(m))

	// Sort the map to prevent constant diffs due to the order of docs in state
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		log.Printf("[INFO] Partial Docs: %s", k)
		v = append(v, m[k])
	}
	return v
}

func CreateSSMDocument(name, content, docType string, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Creating SSM Document: %q", name)

	input := &ssm.CreateDocumentInput{
		Name:         aws.String(name),
		Content:      aws.String(content),
		DocumentType: aws.String(docType),
	}

	log.Printf("[DEBUG] Waiting for SSM Document %q to be created", name)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := ssmconn.CreateDocument(input)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return errwrap.Wrapf("[ERROR] Error creating SSM document: {{err}}", err)
	}
	return nil
}

func UpdateSSMDocument(document map[string]interface{}, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	name := document["name"].(string)

	schemaVersion := document["schema_version"].(string)
	schemaNumber, _ := strconv.ParseFloat(schemaVersion, 64)

	if schemaNumber < MINIMUM_VERSIONED_SCHEMA {
		log.Printf("[DEBUG] Skipping document update because schema version is not 2.0 %q", name)
		return nil
	}

	log.Printf("[INFO] Updating SSM Document: %q", name)

	newDefaultVersion := document["default_version"].(string)

	updateDocInput := &ssm.UpdateDocumentInput{
		Name:            aws.String(name),
		Content:         aws.String(document["content"].(string)),
		DocumentVersion: aws.String(newDefaultVersion),
	}

	updated, err := ssmconn.UpdateDocument(updateDocInput)

	if isAWSErr(err, "DuplicateDocumentContent", "") {
		log.Printf("[DEBUG] Content is a duplicate of the latest version so update is not necessary: %s", name)
		log.Printf("[INFO] Updating the default version to the latest version %s: %s", newDefaultVersion, name)

		newDefaultVersion = document["latest_version"].(string)
	} else if err != nil {
		return errwrap.Wrapf("Error updating SSM document: {{err}}", err)
	} else {
		log.Printf("[INFO] Updating the default version to the new version %s: %s", newDefaultVersion, name)
		newDefaultVersion = *updated.DocumentDescription.DocumentVersion
	}

	updateDefaultInput := &ssm.UpdateDocumentDefaultVersionInput{
		Name:            aws.String(name),
		DocumentVersion: aws.String(newDefaultVersion),
	}

	_, err = ssmconn.UpdateDocumentDefaultVersion(updateDefaultInput)

	if err != nil {
		return errwrap.Wrapf("Error updating the default document version to that of the updated document: {{err}}", err)
	}
	return nil
}

func DeleteSSMDocument(name string, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Deleting SSM Document: %q", name)

	input := &ssm.DeleteDocumentInput{
		Name: aws.String(name),
	}

	_, err := ssmconn.DeleteDocument(input)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for SSM Document %q to be deleted", name)
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := ssmconn.DescribeDocument(&ssm.DescribeDocumentInput{
			Name: aws.String(name),
		})

		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if !ok {
				return resource.NonRetryableError(err)
			}

			if awsErr.Code() == "InvalidDocument" {
				return nil
			}

			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for the document to be deleted", input.Name))
	})
	if err != nil {
		return err
	}
	return nil
}
