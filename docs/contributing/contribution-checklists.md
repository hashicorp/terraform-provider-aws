# Contribution Types and Checklists

There are several different kinds of contribution, each of which has its own
standards for a speedy review. The following sections describe guidelines for
each type of contribution.

- [Documentation Update](#documentation-update)
- [Enhancement/Bugfix to a Resource](#enhancementbugfix-to-a-resource)
- [Adding Resource Import Support](#adding-resource-import-support)
- [Adding Resource Name Generation Support](#adding-resource-name-generation-support)
    - [Resource Name Generation Code Implementation](#resource-name-generation-code-implementation)
    - [Resource Name Generation Testing Implementation](#resource-name-generation-testing-implementation)
    - [Resource Name Generation Documentation Implementation](#resource-name-generation-documentation-implementation)
    - [Resource Name Generation With Suffix](#resource-name-generation-with-suffix)
- [Adding Resource Policy Support](#adding-resource-policy-support)
- [Adding Resource Tagging Support](#adding-resource-tagging-support)
    - [Generating Tag Code for a Service](#generating-tag-code-for-a-service)
    - [Resource Tagging Code Implementation](#resource-tagging-code-implementation)
    - [Resource Tagging Acceptance Testing Implementation](#resource-tagging-acceptance-testing-implementation)
    - [Resource Tagging Documentation Implementation](#resource-tagging-documentation-implementation)
- [Adding Resource Filtering Support](#adding-resource-filtering-support)
    - [Adding Service to Filter Generating Code](#adding-service-to-filter-generating-code)
    - [Resource Filtering Code Implementation](#resource-filtering-code-implementation)
    - [Resource Filtering Documentation Implementation](#resource-filtering-documentation-implementation)
- [New Resource](#new-resource)
    - [New Tag Resource](#new-tag-resource)
- [New Service](#new-service)
- [New Region](#new-region)

## Documentation Update

The [Terraform AWS Provider's website source](../../website) is in this repository
along with the code and tests. Below are some common items that will get
flagged during documentation reviews:

- [ ] __Reasoning for Change__: Documentation updates should include an explanation for why the update is needed.
- [ ] __Prefer AWS Documentation__: Documentation about AWS service features and valid argument values that are likely to update over time should link to AWS service user guides and API references where possible.
- [ ] __Large Example Configurations__: Example Terraform configuration that includes multiple resource definitions should be added to the repository `examples` directory instead of an individual resource documentation page. Each directory under `examples` should be self-contained to call `terraform apply` without special configuration.
- [ ] __Terraform Configuration Language Features__: Individual resource documentation pages and examples should refrain from highlighting particular Terraform configuration language syntax workarounds or features such as `variable`, `local`, `count`, and built-in functions.

## Enhancement/Bugfix to a Resource

Working on existing resources is a great way to get started as a Terraform
contributor because you can work within existing code and tests to get a feel
for what to do.

In addition to the below checklist, please see the [Common Review
Items](pullrequest-submission-and-lifecycle.md#common-review-items) sections for more specific coding and testing
guidelines.

- [ ] __Acceptance test coverage of new behavior__: Existing resources each
   have a set of [acceptance tests][acctests] covering their functionality.
   These tests should exercise all the behavior of the resource. Whether you are
   adding something or fixing a bug, the idea is to have an acceptance test that
   fails if your code were to be removed. Sometimes it is sufficient to
   "enhance" an existing test by adding an assertion or tweaking the config
   that is used, but it's often better to add a new test. You can copy/paste an
   existing test and follow the conventions you see there, modifying the test
   to exercise the behavior of your code.
- [ ] __Documentation updates__: If your code makes any changes that need to
   be documented, you should include those doc updates in the same PR. This
   includes things like new resource attributes or changes in default values.
- [ ] __Well-formed Code__: Do your best to follow existing conventions you
   see in the codebase, and ensure your code is formatted with `go fmt`.
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
- [ ] __Dependency updates__: Create a separate PR if you are updating dependencies.
   This is to avoid conflicts as version updates tend to be fast-
   moving targets. We will plan to merge the PR with this change first.

## Adding Resource Import Support

Adding import support for Terraform resources will allow existing infrastructure to be managed within Terraform. This type of enhancement generally requires a small to moderate amount of code changes.

Comprehensive code examples and information about resource import support can be found in the [Extending Terraform documentation](https://www.terraform.io/docs/extend/resources/import.html).

In addition to the below checklist and the items noted in the Extending Terraform documentation, please see the [Common Review Items](pullrequest-submission-and-lifecycle.md#common-review-items) sections for more specific coding and testing guidelines.

- [ ] _Resource Code Implementation_: In the resource code (e.g., `internal/service/{service}/{thing}.go`), implementation of `Importer` `State` function
- [ ] _Resource Acceptance Testing Implementation_: In the resource acceptance testing (e.g., `internal/service/{service}/{thing}_test.go`), implementation of `TestStep`s with `ImportState: true`
- [ ] _Resource Documentation Implementation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), addition of `Import` documentation section at the bottom of the page

## Adding Resource Name Generation Support

Terraform AWS Provider resources can use shared logic to support and test name generation, where the operator can choose between an expected naming value, a generated naming value with a prefix, or a fully generated name.

Implementing name generation support for Terraform AWS Provider resources requires the following, each with its own section below:

- [ ] _Resource Name Generation Code Implementation_: In the resource code (e.g., `internal/service/{service}/{thing}.go`), implementation of `name_prefix` attribute, along with handling in `Create` function.
- [ ] _Resource Name Generation Testing Implementation_: In the resource acceptance testing (e.g., `internal/service/{service}/{thing}_test.go`), implementation of new acceptance test functions and configurations to exercise new naming logic.
- [ ] _Resource Name Generation Documentation Implementation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), addition of `name_prefix` argument and update of `name` argument description.

### Resource Name Generation Code Implementation

- In the resource Go file (e.g., `internal/service/{service}/{thing}.go`), add the following Go import: `"github.com/hashicorp/terraform-provider-aws/internal/create"`
- In the resource schema, add the new `name_prefix` attribute and adjust the `name` attribute to be `Optional`, `Computed`, and `ConflictsWith` the `name_prefix` attribute. Ensure to keep any existing schema fields on `name` such as `ValidateFunc`. E.g.

```go
"name": {
  Type:          schema.TypeString,
  Optional:      true,
  Computed:      true,
  ForceNew:      true,
  ConflictsWith: []string{"name_prefix"},
},
"name_prefix": {
  Type:          schema.TypeString,
  Optional:      true,
  Computed:      true,
  ForceNew:      true,
  ConflictsWith: []string{"name"},
},
```

- In the resource `Create` function, switch any calls from `d.Get("name").(string)` to instead use the `create.Name()` function, e.g.

```go
name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

// ... in AWS Go SDK Input types, etc. use aws.String(name)
```

- If the resource supports import, in the resource `Read` function add a call to `d.Set("name_prefix", ...)`, e.g.

```go
d.Set("name", resp.Name)
d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(resp.Name)))
```

### Resource Name Generation Testing Implementation

- In the resource testing (e.g., `internal/service/{service}/{thing}_test.go`), add the following Go import: `"github.com/hashicorp/terraform-provider-aws/internal/create"`
- In the resource testing, implement two new tests named `_Name_Generated` and `_NamePrefix` with associated configurations, that verifies creating the resource without `name` and `name_prefix` arguments (for the former) and with only the `name_prefix` argument (for the latter). E.g.

```go
func TestAccServiceThing_nameGenerated(t *testing.T) {
  var thing service.ServiceThing
  resourceName := "aws_service_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:          func() { acctest.PreCheck(t) },
    ErrorCheck:        acctest.ErrorCheck(t, service.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
    CheckDestroy:      testAccCheckThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccThingNameGeneratedConfig(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckThingExists(resourceName, &thing),
          create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
          resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
        ),
      },
      // If the resource supports import:
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

func TestAccServiceThing_namePrefix(t *testing.T) {
  var thing service.ServiceThing
  resourceName := "aws_service_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:          func() { acctest.PreCheck(t) },
    ErrorCheck:        acctest.ErrorCheck(t, service.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
    CheckDestroy:      testAccCheckThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccThingNamePrefixConfig("tf-acc-test-prefix-"),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckThingExists(resourceName, &thing),
          create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
          resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
        ),
      },
      // If the resource supports import:
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

func testAccThingNameGeneratedConfig() string {
  return fmt.Sprintf(`
resource "aws_service_thing" "test" {
  # ... other configuration ...
}
`)
}

func testAccThingNamePrefixConfig(namePrefix string) string {
  return fmt.Sprintf(`
resource "aws_service_thing" "test" {
  # ... other configuration ...

  name_prefix = %[1]q
}
`, namePrefix)
}
```

### Resource Name Generation Documentation Implementation

- In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), add the following to the arguments reference:

```markdown
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
```

- Adjust the existing `name` argument reference to ensure its denoted as `Optional`, includes a mention that it can be generated, and that it conflicts with `name_prefix`:

```markdown
* `name` - (Optional) Name of the thing. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
```

### Resource Name Generation With Suffix

Some generated resource names require a fixed suffix (for example Amazon SNS FIFO topic names must end in `.fifo`).
In these cases use `create.NameWithSuffix()` in the resource `Create` function and `create.NamePrefixFromNameWithSuffix()` in the resource `Read` function, e.g.

```go
name := create.NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string), ".fifo")
```

and

```go
d.Set("name", resp.Name)
d.Set("name_prefix", create.NamePrefixFromNameWithSuffix(aws.StringValue(resp.Name), ".fifo"))
```

There are also functions `create.TestCheckResourceAttrNameWithSuffixGenerated` and `create.TestCheckResourceAttrNameWithSuffixFromPrefix` for use in tests.

## Adding Resource Policy Support

Some AWS components support [resource-based IAM policies](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_identity-vs-resource.html) to control permissions. When implementing this support in the Terraform AWS Provider, we typically prefer creating a separate resource, `aws_{SERVICE}_{THING}_policy` (e.g., `aws_s3_bucket_policy`). See the [New Resource section](#new-resource) for more information about implementing the separate resource and the [Provider Design page](provider-design.md) for rationale.

## Adding Resource Tagging Support

AWS provides key-value metadata across many services and resources, which can be used for a variety of use cases including billing, ownership, and more. See the [AWS Tagging Strategy page](https://aws.amazon.com/answers/account-management/aws-tagging-strategies/) for more information about tagging at a high level.

As of version 3.38.0 of the Terraform AWS Provider, resources that previously implemented tagging support via the argument `tags`, now support provider-wide default tagging.

Thus, for in-flight and future contributions, implementing tagging support for Terraform AWS Provider resources requires the following, each with its own section below:

- [ ] _Generated Service Tagging Code_: Each service has a `generate.go` file where generator directives live. Through these directives and their flags, you can customize code generation for the service. You can find the code that the tagging generator generates in a `tags_gen.go` file in a service, such as `internal/service/ec2/tags_gen.go`. Unlike previously, you should generally _not_ need to edit the generator code (i.e., in `internal/generate/tags`).
- [ ] _Resource Tagging Code Implementation_: In the resource code (e.g., `internal/service/{service}/{thing}.go`), implementation of `tags` and `tags_all` schema attributes, along with implementation of `CustomizeDiff` in the resource definition and handling in `Create`, `Read`, and `Update` functions.
- [ ] _Resource Tagging Acceptance Testing Implementation_: In the resource acceptance testing (e.g., `internal/service/{service}/{thing}_test.go`), implementation of new acceptance test function and configurations to exercise new tagging logic.
- [ ] _Resource Tagging Documentation Implementation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), addition of `tags` argument and `tags_all` attribute.

### Generating Tag Code for a Service

This step is only necessary for the first implementation and may have been previously completed. If so, move on to the next section.

More details about this code generation, including fixes for potential error messages in this process, can be found in the [generate documentation](../../internal/generate/tags/README.md).

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

### Resource Tagging Code Implementation

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

- Some EC2 resources (e.g., [`aws_ec2_fleet`](https://www.terraform.io/docs/providers/aws/r/ec2_fleet.html)) have a `TagSpecifications` field in the `InputStruct` instead of a `Tags` field. In these cases the `ec2TagSpecificationsFromKeyValueTags()` helper function should be used. This example shows using `TagSpecifications`:

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

### Resource Tagging Acceptance Testing Implementation

- In the resource testing (e.g., `internal/service/eks/cluster_test.go`), verify that existing resources without tagging are unaffected and do not have tags saved into their Terraform state. This should be done in the `_basic` acceptance test by adding one line similar to `resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),` and one similar to `resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),`
- In the resource testing, implement a new test named `_tags` with associated configurations, that verifies creating the resource with tags and updating tags. E.g., EKS Clusters:

  ```go
  func TestAccEKSCluster_tags(t *testing.T) {
    var cluster1, cluster2, cluster3 eks.Cluster
    rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
    resourceName := "aws_eks_cluster.test"

    resource.ParallelTest(t, resource.TestCase{
      PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
      ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		  ProviderFactories: acctest.ProviderFactories,
      CheckDestroy:      testAccCheckClusterDestroy,
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

### Resource Tagging Documentation Implementation

- In the resource documentation (e.g., `website/docs/r/eks_cluster.html.markdown`), add the following to the arguments reference:

  ```markdown
  * `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
  ```

- In the resource documentation (e.g., `website/docs/r/eks_cluster.html.markdown`), add the following to the attributes reference:

  ```markdown
  * `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
  ```

## Adding Resource Filtering Support

AWS provides server-side filtering across many services and resources, which can be used when listing resources of that type, for example in the implementation of a data source.
See the [EC2 Listing and filtering your resources page](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Filtering.html#Filtering_Resources_CLI) for information about how server-side filtering can be used with EC2 resources.

Implementing server-side filtering support for Terraform AWS Provider resources requires the following, each with its own section below:

- [ ] _Generated Service Filtering Code_: In the internal code generators (e.g., `internal/generate/namevaluesfilters`), implementation and customization of how a service handles filtering, which is standardized for the resources.
- [ ] _Resource Filtering Code Implementation_: In the resource's equivalent data source code (e.g., `internal/service/{servicename}/thing_data_source.go`), implementation of `filter` schema attribute, along with handling in the `Read` function.
- [ ] _Resource Filtering Documentation Implementation_: In the resource's equivalent data source documentation (e.g., `website/docs/d/service_thing.html.markdown`), addition of `filter` argument

### Adding Service to Filter Generating Code

This step is only necessary for the first implementation and may have been previously completed. If so, move on to the next section.

More details about this code generation can be found in the [namevaluesfilters documentation](../../internal/generate/namevaluesfilters/README.md).

- Open the AWS Go SDK documentation for the service, e.g., for [`service/rds`](https://docs.aws.amazon.com/sdk-for-go/api/service/rds/). Note: there can be a delay between the AWS announcement and the updated AWS Go SDK documentation.
- Determine if the service API includes functionality for filtering resources (usually a `Filters` argument to a `DescribeThing` API call). If so, add the AWS Go SDK service name (e.g., `rds`) to `sliceServiceNames` in `internal/generate/namevaluesfilters/generators/servicefilters/main.go`.
- Run `make gen` (`go generate ./...`) and ensure there are no errors via `make test` (`go test ./...`)

### Resource Filtering Code Implementation

- In the resource's equivalent data source Go file (e.g., `internal/service/ec2/internet_gateway_data_source.go`), add the following Go import: `"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"`
- In the resource schema, add `"filter": namevaluesfilters.Schema(),`
- Implement the logic to build the list of filters:

```go
input := &ec2.DescribeInternetGatewaysInput{}

// Filters based on attributes.
filters := namevaluesfilters.New(map[string]string{
	"internet-gateway-id": d.Get("internet_gateway_id").(string),
})
// Add filters based on keyvalue tags (N.B. Not applicable to all AWS services that support filtering)
filters.Add(namevaluesfilters.Ec2Tags(keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()))
// Add filters based on the custom filtering "filter" attribute.
filters.Add(d.Get("filter").(*schema.Set))

input.Filters = filters.Ec2Filters()
```

### Resource Filtering Documentation Implementation

- In the resource's equivalent data source documentation (e.g., `website/docs/d/internet_gateway.html.markdown`), add the following to the arguments reference:

```markdown
* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInternetGateways.html).

* `values` - (Required) Set of values that are accepted for the given field.
  An Internet Gateway will be selected if any one of the given values matches.
```

## New Resource

_Before submitting this type of contribution, it is highly recommended to read and understand the other pages of the [Contributing Guide](./README.md)._

Implementing a new resource is a good way to learn more about how Terraform
interacts with upstream APIs. There are plenty of examples to draw from in the
existing resources, but you still get to implement something completely new.

In addition to the below checklist, please see the [Common Review
Items](pullrequest-submission-and-lifecycle.md#common-review-items) sections for more specific coding and testing
guidelines.

- [ ] __Minimal LOC__: It's difficult for both the reviewer and author to go
   through long feedback cycles on a big PR with many resources. We ask you to
   only submit **1 resource at a time**.
- [ ] __Acceptance tests__: New resources should include acceptance tests
   covering their behavior. See [Writing Acceptance
   Tests](./running-and-writing-acceptance-tests.md#writing-an-acceptance-test) for a detailed guide on how to
   approach these.
- [ ] __Resource Naming__: Resources should be named `aws_<service>_<name>`,
   using underscores (`_`) as the separator. Resources are namespaced with the
   service name to allow easier searching of related resources, to align
   the resource naming with the service for [Customizing Endpoints](https://www.terraform.io/docs/providers/aws/guides/custom-service-endpoints.html#available-endpoint-customizations),
   and to prevent future conflicts with new AWS services/resources.
   For reference:

    - `service` is the AWS short service name that matches the key in
     the `serviceData` map in the `conns` package (created via the [New Service](#new-service)
     section)
    - `name` represents the conceptual infrastructure represented by the
     create, read, update, and delete methods of the service API. It should
     be a singular noun. For example, in an API that has methods such as
     `CreateThing`, `DeleteThing`, `DescribeThing`, and `ModifyThing` the name
     of the resource would end in `_thing`.

- [ ] __Arguments_and_Attributes__: The HCL for arguments and attributes should mimic the types and structs presented by the AWS API. API arguments should be converted from `CamelCase` to `camel_case`. The resource logic for handling these should follow the recommended implementations in the [Data Handling and Conversion](data-handling-and-conversion.md) documentation.
- [ ] __Documentation__: Each data source and resource gets a page in the Terraform
   documentation, which lives at `website/docs/d/<service>_<name>.html.markdown` and
   `website/docs/r/<service>_<name>.html.markdown` respectively.
- [ ] __Well-formed Code__: Do your best to follow existing conventions you
   see in the codebase, and ensure your code is formatted with `go fmt`.
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
- [ ] __Dependency updates__: Create a separate PR if you are updating dependencies.
   This is to avoid conflicts as version updates tend to be fast-
   moving targets. We will plan to merge the PR with this change first.

### New Tag Resource

Adding a tag resource, similar to the `aws_ecs_tag` resource, has its own implementation procedure since the resource code and initial acceptance testing functions are automatically generated. The rest of the resource acceptance testing and resource documentation must still be manually created.

- In `internal/generate`: Ensure the service is supported by all generators. Run `make gen` after any modifications.
- In `internal/service/{service}/generate.go`: Add the new `//go:generate` call with the correct generator directives. Run `make gen` after any modifications.
- In `internal/provider/provider.go`: Add the new resource.
- Run `make test` and ensure there are no failures.
- Create `internal/service/{service}/tag_gen_test.go` with initial acceptance testing similar to the following (where the parent resource is simple to provision):

```go

import (
	"fmt"
	"testing"

  "github.com/aws/aws-sdk-go/service/{Service}"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc{Service}Tag_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_{service}_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
    ErrorCheck:        acctest.ErrorCheck(t, {Service}.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheck{Service}TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAcc{Service}TagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{Service}TagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc{Service}Tag_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_{service}_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
    ErrorCheck:        acctest.ErrorCheck(t, {Service}.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheck{Service}TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAcc{Service}TagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{Service}TagExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAws{Service}Tag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAcc{Service}Tag_Value(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_{service}_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
    ErrorCheck:        acctest.ErrorCheck(t, {Service}.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheck{Service}TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAcc{Service}TagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{Service}TagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAcc{Service}TagConfig(rName, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{Service}TagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAcc{Service}TagConfig(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_{service}_{thing}" "test" {
  name = %[1]q

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_{service}_tag" "test" {
  resource_arn = aws_{service}_{thing}.test.arn
  key          = %[2]q
  value        = %[3]q
}
`, rName, key, value)
}
```

- Run `make testacc TESTS=TestAcc{Service}Tags_ PKG={Service}` and ensure there are no failures.
- Create `website/docs/r/{service}_tag.html.markdown` with initial documentation similar to the following:

``````markdown
---
subcategory: "{SERVICE}"
layout: "aws"
page_title: "AWS: aws_{service}_tag"
description: |-
  Manages an individual {SERVICE} resource tag
---

# Resource: aws_{service}_tag

Manages an individual {SERVICE} resource tag. This resource should only be used in cases where {SERVICE} resources are created outside Terraform (e.g., {SERVICE} {THING}s implicitly created by {OTHER SERVICE THING}).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_{service}_{thing}` and `aws_{service}_tag` to manage tags of the same {SERVICE} {THING} will cause a perpetual difference where the `aws_{service}_{thing}` resource will try to remove the tag being added by the `aws_{service}_tag` resource.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
resource "aws_{service}_tag" "example" {
  resource_arn = "..."
  key          = "Name"
  value        = "Hello World"
}
```

## Argument Reference

The following arguments are supported:

* `resource_arn` - (Required) Amazon Resource Name (ARN) of the {SERVICE} resource to tag.
* `key` - (Required) Tag name.
* `value` - (Required) Tag value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - {SERVICE} resource identifier and key, separated by a comma (`,`)

## Import

`aws_{service}_tag` can be imported by using the {SERVICE} resource identifier and key, separated by a comma (`,`), e.g.

```
$ terraform import aws_{service}_tag.example arn:aws:{service}:us-east-1:123456789012:{thing}/example,Name
```
``````

## New Service

Implementing a new AWS service gives Terraform the ability to manage resources in
a whole new API. It's a larger undertaking, but brings major new functionality
into Terraform.

- [ ] __Service Client__: Before new resources are submitted, we request
  a separate pull request containing just the new AWS Go SDK service client.
  Doing so will pull the AWS Go SDK service code into the project at the
  current version. Since the AWS Go SDK is updated frequently, these pull
  requests can easily have merge conflicts or be out of date. The maintainers
  prioritize reviewing and merging these quickly to prevent those situations.

  **We have changed these directions a lot! Please review them carefully!**

  To add an AWS Go SDK service client:

    - Determine the service identifier using the rule described in [the Naming Guide](./naming.md#service-identifier).
    - In `names/names_data.csv`, add a new line with all the requested information for the service following the guidance in the [`names` README](../../names/README.md). **_Be very careful when adding or changing data in `names_data.csv`! The Provider and generators depend on the file being correct._**

    - Run the following then submit the pull request:

  ```sh
  make gen
  make test
  go mod tidy
  ```

- [ ] __Initial Resource__: Some services can be big and it can be
  difficult for both reviewer & author to go through long feedback cycles
  on a big PR with many resources. Often feedback items in one resource
  will also need to be applied in other resources. We prefer you to submit
  the necessary minimum in a single PR, ideally **just the first resource**
  of the service.

The initial resource and changes afterwards should follow the other sections
of this guide as appropriate.

## New Region

While region validation is automatically added with SDK updates, new regions
are generally limited in which services they support. Below are some
manually sourced values from documentation. Amazon employees can code search
previous region values to find new region values in internal packages like
RIPStaticConfig if they are not documented yet.

- [ ] Check [Elastic Load Balancing endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region) and add Route53 Hosted Zone ID if available to [`internal/service/elb/hosted_zone_id_data_source.go`](../../internal/service/elb/hosted_zone_id_data_source.go) and [`internal/service/elbv2/hosted_zone_id_data_source.go`](../../internal/service/elbv2/hosted_zone_id_data_source.go)
- [ ] Check [Amazon Simple Storage Service endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_region) and add Route53 Hosted Zone ID if available to [`internal/service/s3/hosted_zones.go`](../../internal/service/s3/hosted_zones.go)
- [ ] Check [CloudTrail Supported Regions docs](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html#cloudtrail-supported-regions) and add AWS Account ID if available to [`internal/service/cloudtrail/service_account_data_source.go`](../../internal/service/cloudtrail/service_account_data_source.go)
- [ ] Check [Elastic Load Balancing Access Logs docs](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy) and add Elastic Load Balancing Account ID if available to [`internal/service/elb/service_account_data_source.go`](../../internal/service/elb/service_account_data_source.go)
- [ ] Check [Redshift Database Audit Logging docs](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) and add AWS Account ID if available to [`internal/service/redshift/service_account_data_source.go`](../../internal/service/redshift/service_account_data_source.go)
- [ ] Check [AWS Elastic Beanstalk endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elasticbeanstalk.html#elasticbeanstalk_region) and add Route53 Hosted Zone ID if available to [`internal/service/elasticbeanstalk/hosted_zone_data_source.go`](../../internal/service/elasticbeanstalk/hosted_zone_data_source.go)
- [ ] Check [SageMaker docs](https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html) and add AWS Account IDs if available to [`internal/service/sagemaker/prebuilt_ecr_image_data_source.go`](../../internal/service/sagemaker/prebuilt_ecr_image_data_source.go)
