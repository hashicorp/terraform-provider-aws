# Adding Resource Name Generation Support

Terraform AWS Provider resources can use shared logic to support and test name generation, where the operator can choose between an expected naming value, a generated naming value with a prefix, or a fully generated name.

Implementing name generation support for Terraform AWS Provider resources requires the following, each with its own section below:

- _Resource Name Generation Code Implementation_: In the resource code (e.g., `internal/service/{service}/{thing}.go`), implementation of `name_prefix` attribute, along with handling in `Create` function.
- _Resource Name Generation Testing Implementation_: In the resource acceptance testing (e.g., `internal/service/{service}/{thing}_test.go`), implementation of new acceptance test functions and configurations to exercise new naming logic.
- _Resource Name Generation Documentation Implementation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), addition of `name_prefix` argument and update of `name` argument description.

## Resource name generation code implementation

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

## Resource name generation testing implementation

- In the resource testing (e.g., `internal/service/{service}/{thing}_test.go`), add the following Go import: `"github.com/hashicorp/terraform-provider-aws/internal/create"`
- In the resource testing, implement two new tests named `_Name_Generated` and `_NamePrefix` with associated configurations, that verifies creating the resource without `name` and `name_prefix` arguments (for the former) and with only the `name_prefix` argument (for the latter). E.g.

```go
func TestAccServiceThing_nameGenerated(t *testing.T) {
  ctx := acctest.Context(t)
  var thing service.ServiceThing
  resourceName := "aws_service_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, service.EndpointsID),
    ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccThingConfig_nameGenerated(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckThingExists(ctx, resourceName, &thing),
          acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
          resource.TestCheckResourceAttr(resourceName, "name_prefix", id.UniqueIdPrefix),
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
  ctx := acctest.Context(t)
  var thing service.ServiceThing
  resourceName := "aws_service_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, service.EndpointsID),
    ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccThingConfig_namePrefix("tf-acc-test-prefix-"),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckThingExists(ctx, resourceName, &thing),
          acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
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

func testAccThingConfig_nameGenerated() string {
  return fmt.Sprintf(`
resource "aws_service_thing" "test" {
  # ... other configuration ...
}
`)
}

func testAccThingConfig_namePrefix(namePrefix string) string {
  return fmt.Sprintf(`
resource "aws_service_thing" "test" {
  # ... other configuration ...

  name_prefix = %[1]q
}
`, namePrefix)
}
```

## Resource name generation documentation implementation

- In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), add the following to the arguments reference:

```markdown
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
```

- Adjust the existing `name` argument reference to ensure its denoted as `Optional`, includes a mention that it can be generated, and that it conflicts with `name_prefix`:

```markdown
* `name` - (Optional) Name of the thing. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
```

## Resource name generation with suffix

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

There are also functions `acctest.CheckResourceAttrNameWithSuffixGenerated` and `acctest.CheckResourceAttrNameWithSuffixFromPrefix` for use in tests.
