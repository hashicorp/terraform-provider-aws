# Adding Resource Tagging Support

AWS provides key-value metadata across many services and resources, which can be used for a variety of use cases including billing, ownership, and more. See the [AWS Tagging Strategy page](https://aws.amazon.com/answers/account-management/aws-tagging-strategies/) for more information about tagging at a high level.

The Terraform AWS Provider supports default tags configured on the provider in addition to tags configured on the resource.
Implementing tagging support for Terraform AWS Provider resources requires the following, each with its own section below:

- _Generated Service Tagging Code_: Each service has a `generate.go` file where generator directives live.
  Through these directives and their flags, you can customize code generation for the service.
  You can find the code that the tagging generator generates in a `tags_gen.go` file in a service, such as `internal/service/ec2/tags_gen.go`.
  You should generally _not_ need to edit the generator code itself (i.e., in `internal/generate/tags`).
- _Resource Tagging Code Implementation_: In the resource code (e.g., `internal/service/{service}/{thing}.go`),
  implementation of `tags` and `tags_all` schema attributes,
  along with implementation of `CustomizeDiff` in the resource definition and handling in `Create`, `Read`, and `Update` functions.
- _Resource Tagging Acceptance Testing Implementation_: In the resource acceptance testing (e.g., `internal/service/{service}/{thing}_test.go`),
  implementation of new acceptance test function and configurations to exercise new tagging logic.
- _Resource Tagging Documentation Implementation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`),
  addition of `tags` argument and `tags_all` attributes.

## Generating Tag Code for a Service

This step is generally only necessary for the first implementation and may have been previously completed.

More details about this code generation,
including fixes for potential error messages in this process,
can be found in the [`generate` package documentation](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/generate/tags/README.md).

The generator will create several types of tagging-related code.
All services that support tagging will generate the function `KeyValueTags`, which converts from service-specific structs returned by the AWS SDK into a common format used by the provider,
and the function `Tags`, which converts from the common format back to the service-specific structs.
In addition, many services have separate functions to list or update tags, so the corresponding `ListTags` and `UpdateTags` can be generated.
Optionally, to retrieve a specific tag, you can generate the `GetTag` function.

If the service directory does not contain a `generate.go` file, create one.
This file must only contain generate directives and a package declaration (e.g., `package eks`).
For examples of the `generate.go` file, many service directories contain one, e.g., `internal/service/eks/generate.go`.

If the `generate.go` file does not contain a generate directive for tagging code, i.e., `//go:generate go run ../../generate/tags/main.go`, add it.
Note that without flags, the directive itself will not do anything useful.
You must not include more than one `generate/tags/main.go` directive, as subsequent directives will overwrite previous directives.
To generate multiple types of tag code, use multiple flags with the directive.

### Generating Tagging Types

Determine how the service implements tagging:
Some services will use a simple map style (`map[string]*string` in Go),
while others will have a separate structure, often a `[]service.Tag` struct with `Key` and `Value` fields.

If the service uses the simple map style, pass the flag `-ServiceTagsMap`.

If the service uses a slice of structs, pass the flag `-ServiceTagsSlice`.
If the name of the tag struct is not `Tag`, pass the flag `-TagType=<struct name>`.
Note that the struct name is used without the package name.
For example, the AppMesh service uses the struct `TagRef`, so the flag is `-TagType=TagRef`.
If the key and value fields on the struct are not `Key` and `Value`,
specify the names using the flags `-TagTypeKeyElem` and `-TagTypeValElem` respectively.
For example, the KMS service uses the struct `Tag`, but the key and value fields are `TagKey` and `TagValue`,
so the flags are `-TagTypeKeyElem=TagKey` and `-TagTypeValElem=TagValue`.

Some services, such as EC2 and Auto Scaling, return a different type depending on the API call used to retrieve the tag.
To indicate the additional type, include the flag `-TagType2=<struct name>`.
For example, the Auto Scaling uses the struct `Tag` as part of resource calls, but returns the struct `TagDescription` from the `DescribeTags` API call. The flag used is `-TagType2=TagDescription`.

For more details on flags for generating service keys, see the
[documentation for the tag generator](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/generate/tags/README.md)

### Generating Standalone Tag Listing Functions

If the service API uses a standalone function to retrieve tags instead of including them with the resource (usually a `ListTags` or `ListTagsForResource` API call), pass the flag `-ListTags`.

If the API call is not `ListTagsForResource`, pass the flag `-ListTagsOp=<API call name>`.
Note that this does not include the package name.
For example, the Auto Scaling service uses the API call `DescribeTags`, so the flag is `-ListTagsOp=DescribeTags`.

If the API call uses a field other than `ResourceArn` to identify the resource, pass the flag `-ListTagsInIDElem=<field name>`.
For example, the CloudWatch service uses the field `ResourceARN`, so the flag is `-ListTagsInIDElem=ResourceARN`.
Some API calls take a slice of identifiers instead of a single identifier.
In this case, pass the flag `-ListTagsInIDNeedSlice=yes`.

If the field containing the tags in the result of the API call is not named `Tags`, pass the flag `-ListTagsOutTagsElem=<struct name>`.
For example, the CloudTrail service returns a nested structure, where the resulting flag is `-ListTagsOutTagsElem=ResourceTagList[0].TagsList`.

In some cases, it can be useful to retrieve single tags.
Pass the flag `-GetTag` to generate a function to do so.

For more details on flags for generating tag listing functions, see the
[documentation for the tag generator](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/generate/tags/README.md)

### Generating Standalone Tag Updating Functions

If the service API uses a standalone function to update tags instead of including them when updating the resource (usually a `TagResource` and `UntagResource` API call), pass the flag `-UpdateTags`.

If the API call to add tags is not `TagResource`, pass the flag `-TagOp=<API call name>`.
Note that this does not include the package name.
For example, the ElastiCache service uses the API call `AddTagsToResource`, so the flag is `-TagOp=AddTagsToResource`.

If the API call to add tags uses a field other than `ResourceArn` to identify the resource, pass the flag `-TagInIDElem=<field name>`.
For example, the EC2 service uses the field `Resources`, so the flag is `-TagInIDElem=Resources`.
Some API calls take a slice of identifiers instead of a single identifier.
In this case, pass the flag `-TagInIDNeedSlice=yes`.

If the API call to remove tags is not `UntagResource`, pass the flag `-UntagOp=<API call name>`.
Note that this does not include the package name.
For example, the ElastiCache service uses the API call `RemoveTagsFromResource`, so the flag is `-UntagOp=RemoveTagsFromResource`.

If the API call to remove tags uses a field other than `ResourceArn` to identify the resource, pass the flag `-UntagInTagsElem=<field name>`.
For example, the Route 53 service uses the field `Keys`, so the flag is `-UntagInTagsElem=Keys`.

For more details on flags for generating tag updating functions, see the
[documentation for the tag generator](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/generate/tags/README.md)

### Specifying the AWS SDK for Go version

The vast majority of the Terraform AWS Provider is implemented using [version 1 of the AWS SDK for Go](https://github.com/aws/aws-sdk-go).
For new services, however, we have started to use [version 2 of the SDK](https://github.com/aws/aws-sdk-go-v2).

By default, the generated code uses the AWS SDK for Go v1.
To generate code using the AWS SDK for Go v2, pass the flag `-AwsSdkVersion=2`.

For more information, see the [documentation on AWS SDK versions](./aws-go-sdk-versions.md).

### Running Code generation

Run the command `make gen` to run the code generators for the project.
To ensure that the code compiles, run `make test`.

## Resource Tagging Code Implementation

### Resource Schema

Add the following imports to the resource's Go source file:

```go
imports (
  /* ... other imports ... */
  tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)
```

Add the `tags` parameter and `tags_all` attribute to the schema.
The `tags` parameter contains the tags set directly on the resource.
The `tags_all` attribute contains union of the tags set directly on the resource and default tags configured on the provider.

```go
func ResourceCluster() *schema.Resource {
  return &schema.Resource{
    /* ... other configuration ... */
		Schema: map[string]*schema.Schema{
      /* ... other configuration ... */
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
    },
  }
}
```

The function `verify.SetTagsDiff` handles the combination of tags set on the resource and default tags,
and must be added to the resource's `CustomizeDiff` function.

If the resource has no other `CustomizeDiff` handler functions, set it directly:

```go
func ResourceCluster() *schema.Resource {
  return &schema.Resource{
    /* ... other configuration ... */
    CustomizeDiff: verify.SetTagsDiff,
  }
}
```

Otherwise, if the resource already contains a `CustomizeDiff` function, append the `SetTagsDiff` via the `customdiff.All` method:

```go
func ResourceExample() *schema.Resource {
  return &schema.Resource{
    /* ... other configuration ... */
    CustomizeDiff: customdiff.All(
      resourceExampleCustomizeDiff,
      verify.SetTagsDiff,
    ),
  }
}
```

### Resource Create Operation

When creating a resource, some AWS APIs support passing tags in the Create call
while others require setting the tags after the initial creation.

If the API supports tagging on creation (e.g., the `Input` struct accepts a `Tags` field),
implement the logic to convert the configuration tags into the service tags, e.g., with EKS Clusters:

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

Otherwise, if the API does not support tagging on creation,
implement the logic to convert the configuration tags into the service API call to tag a resource, e.g., with Device Farm device pools:

```go
// Typically declared near conn := /* ... */
defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

/* ... creation steps ... */

if len(tags) > 0 {
  if err := UpdateTags(conn, d.Id(), nil, tags); err != nil {
    return fmt.Errorf("adding DeviceFarm Device Pool (%s) tags: %w", d.Id(), err)
  }
}
```

Some EC2 resources (e.g., [`aws_ec2_fleet`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ec2_fleet)) have a `TagSpecifications` field in the `InputStruct` instead of a `Tags` field.
In these cases the `tagSpecificationsFromKeyValueTags()` helper function should be used.
This example shows using `TagSpecifications`:

  ```go
  // Typically declared near conn := /* ... */
  defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
  tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
  
  input := &ec2.CreateFleetInput{
    /* ... other configuration ... */
    TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeFleet),
  }
  ```

### Resource Read Operation

In the resource `Read` operation, implement the logic to convert the service tags to save them into the Terraform state for drift detection, e.g., with EKS Clusters:

```go
// Typically declared near conn := /* ... */
defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

/* ... other d.Set(...) logic ... */

tags := KeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
  return fmt.Errorf("setting tags: %w", err)
}

if err := d.Set("tags_all", tags.Map()); err != nil {
  return fmt.Errorf("setting tags_all: %w", err)
}
```

If the service API does not return the tags directly from reading the resource and requires a separate API call,
use the generated `ListTags` function, e.g., with Athena Workgroups:

```go
// Typically declared near conn := /* ... */
defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

/* ... other d.Set(...) logic ... */

tags, err := ListTags(conn, arn.String())

if err != nil {
  return fmt.Errorf("listing tags for resource (%s): %w", arn, err)
}

tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
  return fmt.Errorf("setting tags: %w", err)
}

if err := d.Set("tags_all", tags.Map()); err != nil {
  return fmt.Errorf("setting tags_all: %w", err)
}
```

### Resource Update Operation

In the resource `Update` operation, implement the logic to handle tagging updates, e.g., with EKS Clusters:

```go
if d.HasChange("tags_all") {
  o, n := d.GetChange("tags_all")
  if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
    return fmt.Errorf("updating tags: %w", err)
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
      return fmt.Errorf("updating IAM policy (%s): %w", d.Id(), err)
  }
}
```

## Resource Tagging Acceptance Testing Implementation

In the resource testing (e.g., `internal/service/eks/cluster_test.go`), verify that existing resources without tagging are unaffected and do not have tags saved into their Terraform state. This should be done in the `_basic` acceptance test by adding one line similar to `resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),` and one similar to `resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),`

In the resource testing, implement a new test named `_tags` with associated configurations, that verifies creating the resource with tags and updating tags. E.g., EKS Clusters:

```go
func TestAccEKSCluster_tags(t *testing.T) {
  var cluster1, cluster2, cluster3 eks.Cluster
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_eks_cluster.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
    ErrorCheck:               acctest.ErrorCheck(t, eks.EndpointsID),
    ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckClusterDestroy,
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

Verify all acceptance testing passes for the resource (e.g., `make testacc TESTS=TestAccEKSCluster_ PKG=eks`)

## Resource Tagging Documentation Implementation

In the resource documentation (e.g., `website/docs/r/eks_cluster.html.markdown`), add the following to the arguments reference:

```markdown
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
```

In the resource documentation (e.g., `website/docs/r/eks_cluster.html.markdown`), add the following to the attributes reference:

```markdown
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
```
