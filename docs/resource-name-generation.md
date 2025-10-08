<!-- markdownlint-configure-file { "code-block-style": false } -->
# Adding Resource Name Generation Support

Terraform AWS Provider resources can use shared logic to support and test name generation, where the operator can choose between an expected naming value, a generated naming value with a prefix, or a fully generated name.

Implementing name generation requires modifying the following:

- _Resource Code_: In the resource code, add a `name_prefix` attribute, along with handling in the `Create` function.
- _Resource Acceptance Tests_: In the resource acceptance tests, add new acceptance test functions and configurations to exercise the naming logic.
- _Resource Documentation_: In the resource documentation, add the `name_prefix` argument and update the `name` argument description.

## Resource Code

- In the resource file (e.g., `internal/service/{service}/{thing}.go`), add the following import: `"github.com/hashicorp/terraform-provider-aws/internal/create"`.
- Inside the resource schema, add the new `name_prefix` attribute and adjust the `name` attribute to be `Optional`, `Computed`, and conflict with the `name_prefix` attribute. Be sure to keep any existing validation functions already present on the `name`.

=== "Terraform Plugin Framework (Preferred)"
    ```go
    "name": schema.StringAttribute{
        Optional: true
        Computed: true,
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.UseStateForUnknown(),
            stringplanmodifier.RequiresReplace(),
        },
        Validators: append(
            stringvalidator.ExactlyOneOf(
                path.MatchRelative().AtParent().AtName("name"),
                path.MatchRelative().AtParent().AtName("name_prefix"),
            ),
        ),
    },
    "name_prefix": schema.StringAttribute{
        Optional:   true,
        Computed:   true,
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.UseStateForUnknown(),
            stringplanmodifier.RequiresReplace(),
        },
    },
    ```

=== "Terraform Plugin SDK V2"
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

- In the resource `Create` function, make use of the `create.Name()` function.

=== "Terraform Plugin Framework (Preferred)"
    ```go
    name := create.Name(plan.Name.ValueString(), plan.NamePrefix.ValueString())

    // ... in AWS Go SDK V2 Input types, etc. use aws.ToString(name)
    ```

=== "Terraform Plugin SDK V2"
    ```go
    name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

    // ... in AWS Go SDK V2 Input types, etc. use aws.ToString(name)
    ```

- If the resource supports import, set both `name` and `name_prefix` in the resource `Read` function.

=== "Terraform Plugin Framework (Preferred)"
    ```go
    state.Name = flex.StringToFramework(ctx, resp.Name)
    state.NamePrefix = create.NamePrefixFromName(flex.StringToFramework(ctx, resp.Name))
    ```

=== "Terraform Plugin SDK V2"
    ```go
    d.Set("name", resp.Name)
    d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(resp.Name)))
    ```

## Resource Acceptance Tests

- In the resource test file (e.g., `internal/service/{service}/{thing}_test.go`), add the following import: `"github.com/hashicorp/terraform-provider-aws/internal/create"`.
- Implement two new tests named `_nameGenerated` and `_namePrefix` which verify the creation of the resource without `name` and `name_prefix` arguments, and with only the `name_prefix` argument, respectively.

```go
func TestAccServiceThing_nameGenerated(t *testing.T) {
  ctx := acctest.Context(t)
  var thing service.ServiceThing
  resourceName := "aws_service_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, names.ServiceServiceID),
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
    ErrorCheck:               acctest.ErrorCheck(t, names.ServiceServiceID),
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

## Resource Documentation

- In the resource documentation (e.g., `website/docs/r/{service}_{thing}.html.markdown`), add the following to the arguments reference.

```markdown
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
```

- Adjust the existing `name` argument description to ensure it is denoted as `Optional`, mention that it can be generated and that it conflicts with `name_prefix`.

```markdown
* `name` - (Optional) Name of the thing. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
```
