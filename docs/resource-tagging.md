# Adding Resource Tagging Support

AWS provides key-value metadata across many services and resources, which can be used for a variety of use cases including billing, ownership, and more. See the [AWS Tagging Strategy page](https://aws.amazon.com/answers/account-management/aws-tagging-strategies/) for more information about tagging at a high level.

As of version 3.38.0 of the Terraform AWS Provider, resources that previously implemented tagging support via the argument `tags`, now support provider-wide default tagging.

Thus, for in-flight and future contributions, implementing tagging support for Terraform AWS Provider resources requires the following, each with its own section below:

- _Generated Service Tagging Code_: Each service has a `generate.go` file where generator directives live. Through these directives and their flags, you can customize code generation for the service. You can find the code that the tagging generator generates in a `tags_gen.go` file in a service, such as `internal/service/ec2/tags_gen.go`. Unlike previously, you should generally _not_ need to edit the generator code (i.e., in `internal/generate/tags`).
- _Resource Tagging Code Implementation_: In the resource code (e.g., `internal/service/{service}/{thing}.go`), implementation of `tags` and `tags_all` schema attributes, along with implementation of `CustomizeDiff` in the resource definition and handling in `Create`, `Read`, and `Update` functions.
- _Resource Tagging Acceptance Testing Implementation_: In the resource acceptance testing (e.g., `internal/service/{service}/{thing}_test.go`), implementation of new acceptance test function and configurations to exercise new tagging logic.
- _Resource Tagging Documentation Implementation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), addition of `tags` argument and `tags_all` attribute.

## Generating Tag Code for a Service

This step is only necessary for the first implementation and may have been previously completed. If so, move on to the next section.

More details about this code generation, including fixes for potential error messages in this process, can be found in the [generate documentation](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/generate/tags/README.md).

- Open the AWS Go SDK documentation for the service, e.g., for [`service/eks`](https://docs.aws.amazon.com/sdk-for-go/api/service/eks/). Note: there can be a delay between the AWS announcement and the updated AWS Go SDK documentation.
- Use the AWS Go SDK to determine which types of tagging code to generate. There are three main types of tagging code you can generate: service tags, list tags, and update tags. These are not mutually exclusive and some services use more than one.
- Determine if a service already has a `generate.go` file (e.g., `internal/service/eks/generate.go`). If none exists, follow the example of other `generate.go` files in many other services. This is a very simple file, perhaps 3-5 lines long, and must _only_ contain generate directives at the very top of the file and a package declaration (e.g., `package eks`) -- _nothing else_.
- Check for a tagging code directive: `//go:generate go run ../../generate/tags/main.go`. If one does not exist, add it. Note that without flags, the directive itself will not do anything useful. **WARNING:** You must never have more than one `generate/tags/main.go` directive in a `generate.go` file. Even if you want to generate all three types of tag code, you will use multiple flags but only one `generate/tags/main.go` directive! Including more than one directive will cause the generator to overwrite one set of generated code with whatever is specified in the next directive.
- If the service supports service tags, determine the service's "type" of tagging implementation. Some services will use a simple map style (`map[string]*string` in Go) while others will have a separate structure (`[]service.Tag` `struct` with `Key` and `Value` fields).

    - If the type is a map, add a new flag to the tagging directive (see above): `-ServiceTagsMap`. If the type is `struct`, add a  `-ServiceTagsSlice` flag.
    - If you use the `-ServiceTagsSlice` flag and if the `struct` name is not exactly `Tag`, you must include the `-TagType` flag with the name of the `struct` (e.g., `-TagType=S3Tag`). If the key and value elements of the `struct` are not exactly `Key` and `Value` respectively, you must include the `-TagTypeKeyElem` and/or `-TagTypeValElem` flags with the correct names.
    - In summary, you may need to include one or more of the following flags with `-ServiceTagsSlice` in order to properly customize the generated code: `-TagKeyType`, `TagPackage`, `TagResTypeElem`, `TagType`, `TagType2`, `TagTypeAddBoolElem`, `TagTypeAddBoolElemSnake`, `TagTypeIDElem`, `TagTypeKeyElem`, and `TagTypeValElem`.


- If the service supports listing tags (usually a `ListTags` or `ListTagsForResource` API call), follow these guidelines.

    - Add a new flag to the tagging directive (see above): `-ListTags`.
    - If the API list operation is not exactly `ListTagsForResource`, include the `-ListTagsOp` flag with the name of the operation (e.g., `-ListTagsOp=DescribeTags`).
    - If the API list tags operation identifying element is not exactly `ResourceArn`, include the `-ListTagsInIDElem` flag with the name of the element (e.g., `-ListTagsInIDElem=ResourceARN`).
    - If the API list tags operation identifying element needs a slice, include the `-ListTagsInIDNeedSlice` flag with a `yes` value (e.g., `-ListTagsInIDNeedSlice=yes`).
    - If the API list tags operation output element is not exactly `Tags`, include the `-ListTagsOutTagsElem` flag with the name of the element (e.g., `-ListTagsOutTagsElem=TagList`).
    - In summary, you may need to include one or more of the following flags with `-ListTags` in order to properly customize the generated code: `ListTagsInFiltIDName`, `ListTagsInIDElem`, `ListTagsInIDNeedSlice`, `ListTagsOp`, `ListTagsOutTagsElem`, `TagPackage`, `TagResTypeElem`, and `TagTypeIDElem`.

- If the service API supports updating tags (usually `TagResource` and `UntagResource` API calls), follow these guidelines.

    - Add a new flag to the tagging directive (see above): `-UpdateTags`.
    - If the API tag operation is not exactly `TagResource`, include the `-TagOp` flag with the name of the operation (e.g., `-TagOp=AddTags`).
    - If the API untag operation is not exactly `UntagResource`, include the `-UntagOp` flag with the name of the operation (e.g., `-UntagOp=RemoveTags`).
    - If the API operation identifying element is not exactly `ResourceArn`, include the `-TagInIDElem` flag with the name of the element (e.g., `-TagInIDElem=ResourceARN`).
    - If the API untag operation tags input element is not exactly `TagKeys`, include the `-UntagInTagsElem` flag with the name of the element (e.g., `-UntagInTagsElem=Keys`).
    - In summary, you may need to include one or more of the following flags with `-UpdateTags` in order to properly customize the generated code: `TagInCustomVal`, `TagInIDElem`, `TagInIDNeedSlice`, `TagInTagsElem`, `TagOp`, `TagOpBatchSize`, `TagPackage`, `TagResTypeElem`, `TagTypeAddBoolElem`, `TagTypeIDElem`, `UntagInCustomVal`, `UntagInNeedTagKeyType`, `UntagInNeedTagType`, `UntagInTagsElem`, and `UntagOp`.

- Run `make gen` (`go generate ./...`) and ensure there are no errors via `make test` (`go test ./...`)

## Resource Tagging Code Implementation

- In the resource Go file (e.g., `internal/service/eks/cluster.go`), add the following Go import: `tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"`
- In the resource schema, add `"tags": tagsSchema(),` and `"tags_all": tagsSchemaComputed(),`
- In the `schema.Resource` struct definition, add the `CustomizeDiff: SetTagsDiff` handling essential to resource support for default tags:

  ```go
  func ResourceCluster() *schema.Resource {
    return &schema.Resource{
      /* ... other configuration ... */
      CustomizeDiff: verify.SetTagsDiff,
    }
  }
  ```

  If the resource already contains a `CustomizeDiff` function, append the `SetTagsDiff` via the `customdiff.Sequence` method:

  ```go
  func ResourceExample() *schema.Resource {
    return &schema.Resource{
      /* ... other configuration ... */
      CustomizeDiff: customdiff.Sequence(
        resourceExampleCustomizeDiff,
        verify.SetTagsDiff,
      ),
    }
  }
  ```

- If the API supports tagging on creation (the `Input` struct accepts a `Tags` field), in the resource `Create` function, implement the logic to convert the configuration tags into the service tags, e.g., with EKS Clusters:

  ```go
  // Typically declared near conn := /* ... */
  defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
  tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
  
  input := &eks.CreateClusterInput{
    /* ... other configuration ... */
    Tags: Tags(tags.IgnoreAWS()),
  }
  ```

  If the service API does not allow passing an empty list, the logic can be adjusted similar to:

  ```go
  // Typically declared near conn := /* ... */
  defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
  tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
  
  input := &eks.CreateClusterInput{
    /* ... other configuration ... */
  }

  if len(tags) > 0 {
    input.Tags = Tags(tags.IgnoreAWS())
  }
  ```

- Otherwise if the API does not support tagging on creation (the `Input` struct does not accept a `Tags` field), in the resource `Create` function, implement the logic to convert the configuration tags into the service API call to tag a resource, e.g., with Elasticsearch Domain:

  ```go
  // Typically declared near conn := /* ... */
  defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
  tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
  
  if len(tags) > 0 {
    if err := UpdateTags(conn, d.Id(), nil, tags); err != nil {
      return fmt.Errorf("error adding Elasticsearch Cluster (%s) tags: %w", d.Id(), err)
    }
  }
  ```

- Some EC2 resources (e.g., [`aws_ec2_fleet`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ec2_fleet)) have a `TagSpecifications` field in the `InputStruct` instead of a `Tags` field. In these cases the `ec2TagSpecificationsFromKeyValueTags()` helper function should be used. This example shows using `TagSpecifications`:

  ```go
  // Typically declared near conn := /* ... */
  defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
  tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
  
  input := &ec2.CreateFleetInput{
    /* ... other configuration ... */
    TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeFleet),
  }
  ```

- In the resource `Read` function, implement the logic to convert the service tags to save them into the Terraform state for drift detection, e.g., with EKS Clusters (which had the tags available in the DescribeCluster API call):

  ```go
  // Typically declared near conn := /* ... */
  defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
  ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
  
  /* ... other d.Set(...) logic ... */

  tags := keyvaluetags.EksKeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
  
  if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
    return fmt.Errorf("error setting tags: %w", err)
  }
  
  if err := d.Set("tags_all", tags.Map()); err != nil {
    return fmt.Errorf("error setting tags_all: %w", err)
  }
  ```

  If the service API does not return the tags directly from reading the resource and requires a separate API call, its possible to use the `keyvaluetags` functionality like the following, e.g., with Athena Workgroups:

  ```go
  // Typically declared near conn := /* ... */
  defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
  ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

  /* ... other d.Set(...) logic ... */
  
  tags, err := keyvaluetags.AthenaListTags(conn, arn.String())

  if err != nil {
    return fmt.Errorf("error listing tags for resource (%s): %w", arn, err)
  }

  tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
  
  if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
    return fmt.Errorf("error setting tags: %w", err)
  }
  
  if err := d.Set("tags_all", tags.Map()); err != nil {
    return fmt.Errorf("error setting tags_all: %w", err)
  }
  ```

- In the resource `Update` function (this may be the first functionality requiring the creation of the `Update` function), implement the logic to handle tagging updates, e.g., with EKS Clusters:

  ```go
  if d.HasChange("tags_all") {
    o, n := d.GetChange("tags_all")
    if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
      return fmt.Errorf("error updating tags: %w", err)
    }
  }
  ```

  If the resource `Update` function applies specific updates to attributes regardless of changes to tags, implement the following e.g., with IAM Policy:

  ```go
  if d.HasChangesExcept("tags", "tags_all") {
    /* ... other logic ...*/
    request := &iam.CreatePolicyVersionInput{
      PolicyArn:      aws.String(d.Id()),
      PolicyDocument: aws.String(d.Get("policy").(string)),
      SetAsDefault:   aws.Bool(true),
    }

    if _, err := conn.CreatePolicyVersion(request); err != nil {
        return fmt.Errorf("error updating IAM policy (%s): %w", d.Id(), err)
    }
  }
  ```

## Resource Tagging Acceptance Testing Implementation

- In the resource testing (e.g., `internal/service/eks/cluster_test.go`), verify that existing resources without tagging are unaffected and do not have tags saved into their Terraform state. This should be done in the `_basic` acceptance test by adding one line similar to `resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),` and one similar to `resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),`
- In the resource testing, implement a new test named `_tags` with associated configurations, that verifies creating the resource with tags and updating tags. E.g., EKS Clusters:

  ```go
  func TestAccEKSCluster_tags(t *testing.T) {
    var cluster1, cluster2, cluster3 eks.Cluster
    rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
    resourceName := "aws_eks_cluster.test"

    resource.ParallelTest(t, resource.TestCase{
      PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
      ErrorCheck:   acctest.ErrorCheck(t, eks.EndpointsID),
		  ProviderFactories: acctest.ProviderFactories,
      CheckDestroy: testAccCheckClusterDestroy,
      Steps: []resource.TestStep{
        {
          Config: testAccClusterConfigTags1(rName, "key1", "value1"),
          Check: resource.ComposeTestCheckFunc(
            testAccCheckClusterExists(resourceName, &cluster1),
            resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
            resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
          ),
        },
        {
          ResourceName:      resourceName,
          ImportState:       true,
          ImportStateVerify: true,
        },
        {
          Config: testAccClusterConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
          Check: resource.ComposeTestCheckFunc(
            testAccCheckClusterExists(resourceName, &cluster2),
            resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
            resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
            resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
          ),
        },
        {
          Config: testAccClusterConfigTags1(rName, "key2", "value2"),
          Check: resource.ComposeTestCheckFunc(
            testAccCheckClusterExists(resourceName, &cluster3),
            resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
            resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
          ),
        },
      },
    })
  }

  func testAccClusterConfigTags1(rName, tagKey1, tagValue1 string) string {
    return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
  resource "aws_eks_cluster" "test" {
    name     = %[1]q
    role_arn = aws_iam_role.test.arn

    tags = {
      %[2]q = %[3]q
    }

    vpc_config {
      subnet_ids = aws_subnet.test[*].id
    }

    depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
  }
  `, rName, tagKey1, tagValue1))
  }

  func testAccClusterConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
    return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
  resource "aws_eks_cluster" "test" {
    name     = %[1]q
    role_arn = aws_iam_role.test.arn

    tags = {
      %[2]q = %[3]q
      %[4]q = %[5]q
    }

    vpc_config {
      subnet_ids = aws_subnet.test[*].id
    }

    depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
  }
  `, rName, tagKey1, tagValue1, tagKey2, tagValue2))
  }
  ```

- Verify all acceptance testing passes for the resource (e.g., `make testacc TESTS=TestAccEKSCluster_ PKG=eks`)

## Resource Tagging Documentation Implementation

- In the resource documentation (e.g., `website/docs/r/eks_cluster.html.markdown`), add the following to the arguments reference:

  ```markdown
  * `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
  ```

- In the resource documentation (e.g., `website/docs/r/eks_cluster.html.markdown`), add the following to the attributes reference:

  ```markdown
  * `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
  ```
  
