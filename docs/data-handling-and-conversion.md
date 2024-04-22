<!-- markdownlint-configure-file { "code-block-style": false } -->
# Data Handling and Conversion

The Terraform AWS Provider codebase bridges the implementation of a [Terraform Plugin](https://www.terraform.io/plugin/how-terraform-works) and an AWS API client to support AWS operations and data types as Terraform Resources. Data handling and conversion is a large portion of resource implementation given the domain-specific implementations of each side of the provider. The first is where Terraform is a generic infrastructure as code tool with a generic data model and the other is where the details are driven by AWS API data modeling concepts. This guide is intended to explain and show preferred Terraform AWS Provider code implementations required to successfully translate data between these two systems.

At the bottom of this documentation is a [Glossary section](#glossary), which may be a helpful reference while reading the other sections.

## Data Conversions in Terraform Providers

Before getting into highly specific documentation about the Terraform AWS Provider handling of data, it may be helpful to briefly highlight how Terraform Plugins (Terraform Providers in this case) interact with Terraform CLI and the Terraform State in general and where this documentation fits into the whole process.

There are two primary data flows that are typically handled by resources within a Terraform Provider. Data is either being converted from a planned new Terraform State into making a remote system request or a remote system response is being converted into an applied new Terraform State. The semantics of how the data of the planned new Terraform State is surfaced to the resource implementation is determined by where a resource is in its lifecycle and is mainly handled by Terraform CLI. This concept can be explored further in the [Terraform Resource Instance Change Lifecycle documentation](https://github.com/hashicorp/terraform/blob/main/docs/resource-instance-change-lifecycle.md), with the caveat that some additional behaviors occur within the Terraform Plugin SDK as well (if the Terraform Plugin uses that implementation detail).

As a generic walkthrough, the following data handling occurs when creating a Terraform Resource:

- An operator creates a Terraform configuration with a new resource defined and runs `terraform apply`
- Terraform CLI merges an empty prior state for the resource, along with the given configuration state, to create a planned new state for the resource
- Terraform CLI sends a Terraform Plugin Protocol request to create the new resource with its planned new state data
- If the Terraform Plugin is using a higher-level library, such as the Terraform Plugin Framework, that library receives the request and translates the Terraform Plugin Protocol data types into the expected library types
- Terraform Plugin invokes the resource creation function with the planned new state data
    - **The planned new state data is converted into a remote system request (e.g., API creation request) that is invoked**
    - **The remote system response is received and the data is converted into an applied new state**
- If the Terraform Plugin is using a higher-level library, such as the Terraform Plugin Framework, that library translates the library types back into Terraform Plugin Protocol data types
- Terraform Plugin responds to Terraform Plugin Protocol request with the new state data
- Terraform CLI verifies and stores the new state

The lines in bold above are the focus of this page.

### Implicit State Passthrough

An important behavior to note with Terraform State handling is if the value of a particular root attribute or block is not refreshed during plan or apply operations, then the prior Terraform State is implicitly deep copied to the new Terraform State for that attribute or block.

Given a resource with a writeable root attribute named `not_set_attr` that never explicitly writes a value, the following happens:

- If the Terraform configuration contains `not_set_attr = "anything"` on resource creation, the Terraform State contains `not_set_attr` equal to `"anything"` after apply.
- If the Terraform configuration is updated to `not_set_attr = "updated"`, the Terraform State contains `not_set_attr` equal to `"updated"` after apply.
- If the attribute was meant to be associated with a remote system value, it will never update the Terraform State on plan or apply with the remote value. Effectively, it cannot perform drift detection with the remote value.

This however does _not_ apply to nested attributes and blocks if the parent block is refreshed.
Given a resource with a root block named `parent`, with nested child attributes `set_attr` and `not_set_attr`, a read operation which updates the value of `parent` (and the nested `set_attr` attribute) will not copy the Terraform State for the nested `not_set_attr` attribute.

There are valid use cases for passthrough attribute values such as these (see the [Virtual Attributes section](#virtual-attributes)), however the behavior can be confusing or incorrect for operators if the drift detection is expected.
Typically these types of drift detection issues can be discovered by implementing resource import testing with state verification.

## Terraform Plugin Framework versus Plugin SDK V2

Perhaps the most distinct difference between [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) and [Terraform Plugin SDKv2](https://developer.hashicorp.com/terraform/plugin/sdkv2) is data handling.
With Terraform Plugin Framework state data is strongly typed, while Plugin SDK V2 based resources represent state data generically (each attribute is an `interface{}`) and types must be asserted at runtime.
Strongly typed data eliminates an entire class of runtime bugs and crashes, but does require compile type declarations and a slightly different approach to reading and writing data.
The sections below contain examples for both plugin libraries, but Terraform Plugin Framework should be preferred whenever possible.

## Data Conversions in the Terraform AWS Provider

To expand on the data handling that occurs specifically within the Terraform AWS Provider resource implementations, the above resource creation items become the below in practice given our current usage of the Terraform Plugin SDK:

=== "Terraform Plugin Framework (Preferred)"
    - The `Create` method of a resource is invoked with `resource.CreateRequest` containing the planned new state data (`req.Plan`) and an AWS API client (stored in the `Meta()` method of the resource struct).
        - Before reaching this point, the `Plan` data was already translated from the Terraform Plugin Protocol data types by the Terraform Plugin Framework so values can be read by invoking `req.Plan.Get(ctx, &plan)`, where `plan` is an instance of the struct representing the resources data.
    - An AWS Go SDK operation input type (e.g., `*ec2.CreateVpcInput`) is initialized
    - For each necessary field to configure in the operation input type, the data is read from the `plan` struct and converted into the AWS Go SDK type for the field (e.g., `*string`)
    - The AWS Go SDK operation is invoked and the output type (e.g., `*ec2.CreateVpcOutput`) is initialized
    - For each necessary Attribute, Block, or resource identifier to be saved in the state, the data is read from the AWS Go SDK type for the field (`*string`), if necessary converted into the equivalent Plugin Framework compatible type, and saved into a mutated data struct
    - Function is returned

=== "Terraform Plugin SDK V2"
    - The `CreateWithoutTimeout` function of a `schema.Resource` is invoked with `*schema.ResourceData` containing the planned new state data (conventionally named `d`) and an AWS API client (conventionally named `meta`).
        - Before reaching this point, the `ResourceData` was already translated from the Terraform Plugin Protocol data types by the Terraform Plugin SDK so values can be read by invoking `d.Get()` and `d.GetOk()` receiver methods with Attribute and Block names from the `Schema` of the `schema.Resource`.
    - An AWS Go SDK operation input type (e.g., `*ec2.CreateVpcInput`) is initialized
    - For each necessary field to configure in the operation input type, the data is read from the `ResourceData` (e.g., `d.Get()`, `d.GetOk()`) and converted into the AWS Go SDK type for the field (e.g., `*string`)
    - The AWS Go SDK operation is invoked and the output type (e.g., `*ec2.CreateVpcOutput`) is initialized
    - For each necessary Attribute, Block, or resource identifier to be saved in the state, the data is read from the AWS Go SDK type for the field (`*string`), if necessary converted into a `ResourceData` compatible type, and saved into a mutated `ResourceData` (e.g., `d.Set()`, `d.SetId()`)
    - Function is returned

### Type Mapping

To further understand the necessary data conversions used throughout the Terraform AWS Provider codebase between AWS Go SDK types and the Terraform Plugin SDK, the following table can be referenced for most scenarios:

=== "Terraform Plugin Framework (Preferred)"
    <!-- markdownlint-disable no-inline-html --->

    | AWS API Model | AWS Go SDK V2 | Terraform Plugin Framework | Terraform Language/State |
    |---------------|---------------|----------------------------|--------------------------|
    | `boolean` | `bool` | `types.Bool` | `bool` |
    | `float` | `*float64` | `types.Float64` | `number` |
    | `integer` | `*int64` | `types.Int64` | `number` |
    | `list` | `[]*T` | `types.List` <br/>`types.Set` | `list(any)`<br/>`set(any)` |
    | `map` | `map[T1]*T2` | `types.Map` | `map(any)` |
    | `string` | `*string` | `types.String` | `string` |
    | `structure` | `struct` | `types.List` with `MaxItems: 1` | `list(object(any))` |
    | `timestamp` | `*time.Time` | `types.String` (typically RFC3339 formatted) | `string` |

    <!-- markdownlint-enable no-inline-html --->

    [Types](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types) are built into the Terraform Plugin Framework library and handle null and unknown values in accordance with the [Terraform type system](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#type-system).
    This eliminates the need for any special handling of zero values and provides better change detection on unset values.

=== "Terraform Plugin SDK V2"
    <!-- markdownlint-disable no-inline-html --->

    | AWS API Model | AWS Go SDK | Terraform Plugin SDK | Terraform Language/State |
    |---------------|------------|----------------------|--------------------------|
    | `boolean` | `*bool` | `TypeBool` (`bool`) | `bool` |
    | `float` | `*float64` | `TypeFloat` (`float64`) | `number` |
    | `integer` | `*int64` | `TypeInt` (`int`) | `number` |
    | `list` | `[]*T` | `TypeList` (`[]interface{}` of `T`)<br/>`TypeSet` (`*schema.Set` of `T`) | `list(any)`<br/>`set(any)` |
    | `map` | `map[T1]*T2` | `TypeMap` (`map[string]interface{}`) | `map(any)` |
    | `string` | `*string` | `TypeString` (`string`) | `string` |
    | `structure` | `struct` | `TypeList` (`[]interface{}` of `map[string]interface{}`) with `MaxItems: 1` | `list(object(any))` |
    | `timestamp` | `*time.Time` | `TypeString` (typically RFC3339 formatted) | `string` |

    <!-- markdownlint-enable no-inline-html --->

    You may notice there are type encoding differences between the AWS Go SDK and Terraform Plugin SDK:

    - AWS Go SDK types are all Go pointer types, while Terraform Plugin SDK types are not.
    - AWS Go SDK structures are the Go `struct` type, while there is no semantically equivalent Terraform Plugin SDK type. Instead they are represented as a slice of interfaces with an underlying map of interfaces.
    - AWS Go SDK types are all Go concrete types, while the Terraform Plugin SDK types for collections and maps are interfaces.
    - AWS Go SDK whole numeric type is always 64-bit, while the Terraform Plugin SDK type is implementation-specific.

    Conceptually, the first and second items above are the most problematic in the Terraform AWS Provider codebase. The first item because non-pointer types in Go cannot implement the concept of no value (`nil`). The [Zero Value Mapping section](#zero-value-mapping) will go into more detail about the implications of this limitation. The second item because it can be confusing to always handle a structure ("object") type as a list.

### Zero Value Mapping

!!! note
    This section only applies to Plugin SDK V2 based resources. Terraform Plugin Framework based resources will handle null and unknown values distinctly from zero values.

As mentioned in the [Type Mapping section](#type-mapping) for Terraform Plugin SDK V2, there is a discrepancy between how the Terraform Plugin SDK represents values and the reality that a Terraform State may not configure an Attribute.
These values will default to the matching underlying Go type "zero value" if not set:

| Terraform Plugin SDK | Go Type | Zero Value |
|----------------------|---------|------------|
| `TypeBool` | `bool` | `false` |
| `TypeFloat` | `float64` | `0.0` |
| `TypeInt` | `int` | `0` |
| `TypeString` | `string` | `""` |

For Terraform resource logic this means that these special values must always be accounted for in implementation.
The semantics of the API and its meaning of the zero value will determine whether:

- If it is not used/needed, then generally the zero value can safely be used to store an "unset" value and should be ignored when sending to the API.
- If it is used/needed, whether:
    - A value can always be set and it is safe to always send to the API. Generally, boolean values fall into this category.
    - A different default/sentinel value must be used as the "unset" value so it can either match the default of the API or be ignored when sending to the API.
    - A special type implementation is required within the schema to work around the limitation.

The maintainers can provide guidance on appropriate solutions for cases not mentioned in the [Recommended Implementation section](#recommended-implementations).

### Root Attributes Versus Block Attributes

=== "Terraform Plugin Framework (Preferred)"
    All Attributes and Blocks at the top level of a resource structs `Schema` method are considered "root" attributes.
    These will always be handled with the `Plan` and `State` fields from the request and response pointers passed in as arguments the the CRUD methods on the resource struct.
    Values are read from and written to the underlying data structure during CRUD operations, and finally written to state in the response object with a call like `resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)`.

    By convention in the codebase, each level of Block handling beyond root attributes should be separated into "expand" functions that convert Terraform Plugin SDK data into the equivalent AWS Go SDK type (typically named `expand{Service}{Type}`) and "flatten" functions that convert an AWS Go SDK type into the equivalent Terraform Plugin SDK data (typically named `flatten{Service}{Type}`).
    The [Recommended Implementations section](#recommended-implementations) will go into those details.

=== "Terraform Plugin SDK V2"
    All Attributes and Blocks at the top level of `schema.Resource` `Schema` are considered "root" attributes.
    These will always be handled with receiver methods on `ResourceData`, such as reading with `d.Get()`, `d.GetOk()`, etc. and writing with `d.Set()`.
    Any nested Attributes and Blocks inside those root Blocks will then be handled with standard Go types according to the table in the [Type Mapping section](#type-mapping).

    By convention in the codebase, each level of Block handling beyond root attributes should be separated into "expand" functions that convert Terraform Plugin SDK data into the equivalent AWS Go SDK type (typically named `expand{Service}{Type}`) and "flatten" functions that convert an AWS Go SDK type into the equivalent Terraform Plugin SDK data (typically named `flatten{Service}{Type}`).
    The [Recommended Implementations section](#recommended-implementations) will go into those details.
    
    !!! warning
        While it is possible in certain type scenarios to deeply read and write ResourceData information for a Block Attribute, this practice is discouraged in preference of only handling root Attributes and Blocks.

## Recommended Implementations

Given the complexities around conversions between AWS and Terraform Plugin type systems, this section contains recommended implementations for Terraform AWS Provider resources.

!!! tip
    Some of these coding patterns may not be well represented in the codebase, as refactoring the many older styles over years of community development is a large task.
    However this is meant to represent the preferred implementations today.
    These will continue to evolve as this codebase and the Terraform Plugin ecosystem changes.

### Where to Define Flex Functions

Define FLatten and EXpand (i.e., flex) functions at the _most local level_ possible. This table provides guidance on the preferred place to define flex functions based on usage.

| Where Used | Where to Define | Include Service in Name |
|---------------|------------|--------|
| One resource (e.g., `aws_instance`) | Resource file (e.g., `internal/service/ec2/instance.go`) | No |
| Few resources in one service (e.g., `EC2`) | Resource file or service flex file (e.g., `internal/service/ec2/flex.go`) | No |
| Widely used in one service (e.g., `EC2`) | Service flex file (e.g., `internal/service/ec2/flex.go`) | No |
| Two services (e.g., `EC2` and `EKS`) | Define a copy in each service | If helpful |
| 3+ services | `internal/flex/flex.go` | Yes |

### Expand Functions for Blocks

=== "Terraform Plugin Framework (Preferred)"
    ```go
    func expandStructure(tfList []structureData) *service.Structure {
        if len(tfList) == 0 {
            return nil
        }

        tfObj := tfList[0]
        apiObject := &service.Structure{}

        // ... nested attribute handling ...
        
        return apiObject
    }

    func expandStructures(tfList []structureData) []*service.Structure {
        if len(tfList) == 0 {
            return nil
        }

        var apiObjects []*service.Structure
        for _, tfObj := range tfList {
            apiObject := &service.Structure{}

            // ... nested attribute handling ...

            apiObjects = append(apiObjects, apiObject)
        }

        return apiObjects
    }
    ```

=== "Terraform Plugin SDK V2"
    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        if tfMap == nil {
            return nil
        }

        apiObject := &service.Structure{}

        // ... nested attribute handling ...

        return apiObject
    }

    func expandStructures(tfList []interface{}) []*service.Structure {
        if len(tfList) == 0 {
            return nil
        }

        var apiObjects []*service.Structure

        for _, tfMapRaw := range tfList {
            tfMap, ok := tfMapRaw.(map[string]interface{})

            if !ok {
                continue
            }

            apiObject := expandStructure(tfMap)

            if apiObject == nil {
                continue
            }

            apiObjects = append(apiObjects, apiObject)
        }

        return apiObjects
    }
    ```

### Flatten Functions for Blocks

=== "Terraform Plugin Framework (Preferred)"
    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        var diags diag.Diagnostics
        elemType := types.ObjectType{AttrTypes: structureAttrTypes}

        if apiObject == nil {
            return types.ListNull(elemType), diags
        }
        
        obj := map[string]attr.Value{
            // ... nested attribute handling ...
        }
        objVal, d := types.ObjectValue(structureAttrTypes, obj)
        diags.Append(d...)
        
        listVal, d := types.ListValue(elemType, []attr.Value{objVal})
        diags.Append(d...)
        
        return listVal, diags
    }
    
    func flattenStructures(ctx context.Context, apiObjects []*service.Structure) (types.List, diag.Diagnostics) {
        var diags diag.Diagnostics
        elemType := types.ObjectType{AttrTypes: structureAttrTypes}
        
        if len(apiObjects) == 0 {
            return types.ListNull(elemType), diags
        }
        
        elems := []attr.Value{}
        for _, apiObject := range apiObjects {
            if apiObject == nil {
                continue
            }

            obj := map[string]attr.Value{
                // ... nested attribute handling ...
            }
            objVal, d := types.ObjectValue(structureAttrTypes, obj)
            diags.Append(d...)

            elems = append(elems, objVal)
        }
        
        listVal, d := types.ListValue(elemType, elems)
        diags.Append(d...)
        
        return listVal, diags
    }
    ```

=== "Terraform Plugin SDK V2"
    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        if apiObject == nil {
            return nil
        }

        tfMap := map[string]interface{}{}
    
        // ... nested attribute handling ...
    
        return tfMap
    }
    
    func flattenStructures(apiObjects []*service.Structure) []interface{} {
        if len(apiObjects) == 0 {
            return nil
        }
    
        var tfList []interface{}
    
        for _, apiObject := range apiObjects {
            if apiObject == nil {
                continue
            }
    
            tfList = append(tfList, flattenStructure(apiObject))
        }
    
        return tfList
    }
    ```

### Root Bool and AWS Boolean

=== "Terraform Plugin Framework (Preferred)"
    To read, if always sending the attribute value is correct:

    ```go
    input := service.ExampleOperationInput{
        AttributeName: aws.String(plan.AttributeName.ValueBool())
    }
    ```
    
    Alternatively, if only sending the attribute value when `true`:
    
    ```go
    input := service.ExampleOperationInput{}
    
    if v := plan.AttributeName.ValueBool(); v {
        input.AttributeName = aws.Bool(v)
    }
    ```

    Or, if only sending the attribute value when it is known and not null:
    
    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsUnknown() && !plan.AttributeName.IsNull() {
        input.AttributeName = aws.Bool(plan.AttributeName.ValueBool())
    }
    ```
    
    To write:
    
    ```go
    plan.AttributeName = flex.BoolToFramework(output.Thing.AttributeName)
    ```

=== "Terraform Plugin SDK V2"
    To read, if always sending the attribute value is correct:

    ```go
    input := service.ExampleOperationInput{
        AttributeName: aws.String(d.Get("attribute_name").(bool))
    }
    ```
    
    Otherwise to read, if only sending the attribute value when `true` is preferred (`!ok` for opposite):
    
    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok {
        input.AttributeName = aws.Bool(v.(bool))
    }
    ```
    
    To write:
    
    ```go
    d.Set("attribute_name", output.Thing.AttributeName)
    ```

### Root Float and AWS Float

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        input.AttributeName = aws.Float64(plan.AttributeName.ValueFloat64())
    }
    ```
    
    To write:
    
    ```go
    plan.AttributeName = flex.Float64ToFramework(output.Thing.AttributeName)
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok {
        input.AttributeName = aws.Float64(v.(float64))
    }
    ```
    
    To write:
    
    ```go
    d.Set("attribute_name", output.Thing.AttributeName)
    ```

### Root Int and AWS Integer

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        input.AttributeName = aws.Int64(plan.AttributeName.ValueInt64())
    }
    ```
    
    To write:
    
    ```go
    plan.AttributeName = flex.Int64ToFramework(output.Thing.AttributeName)
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok {
        input.AttributeName = aws.Int64(int64(v.(int)))
    }
    ```
    
    To write:
    
    ```go
    d.Set("attribute_name", output.Thing.AttributeName)
    ```

### Root List of Resource and AWS List of Structure

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        var tfList []attributeNameData
        resp.Diagnostics.Append(plan.AttributeName.ElementsAs(ctx, &tfList, false)...)
        if resp.Diagnostics.HasError() {
            return
        }

        input.AttributeName = expandStructures(tfList)
    }
    ```
    
    To write:
    
    ```go
    attributeName, d := flattenStructures(ctx, output.Thing.AttributeName))
    resp.Diagnostics.Append(d...)
    state.AttributeName = attributeName
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok && len(v.([]interface{})) > 0 {
        input.AttributeName = expandStructures(v.([]interface{}))
    }
    ```
    
    To write:
    
    ```go
    if err := d.Set("attribute_name", flattenStructures(output.Thing.AttributeName)); err != nil {
        return sdkdiag.AppendErrorf(diags, "setting attribute_name: %s", err)
    }
    ```

### Root List of Resource and AWS Structure

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        var tfList []attributeNameData
        resp.Diagnostics.Append(plan.AttributeName.ElementsAs(ctx, &tfList, false)...)
        if resp.Diagnostics.HasError() {
            return
        }

        // expander handles translating list with 1 item to a single AWS object
        input.AttributeName = expandStructure(tfList)
    }
    ```
    
    To write:
    
    ```go
    // flattener handles nil output, returning the equivalent null Terraform type
    attributeName, d := flattenStructures(ctx, output.Thing.AttributeName))
    resp.Diagnostics.Append(d...)
    state.AttributeName = attributeName
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
        input.AttributeName = expandStructure(v.([]interface{})[0].(map[string]interface{}))
    }
    ```
    
    To write (_likely to have helper function introduced soon_):
    
    ```go
    if output.Thing.AttributeName != nil {
        if err := d.Set("attribute_name", []interface{}{flattenStructure(output.Thing.AttributeName)}); err != nil {
            return sdkdiag.AppendErrorf(diags, "setting attribute_name: %s", err)
        }
    } else {
        d.Set("attribute_name", nil)
    }
    ```

### Root List of String and AWS List of String

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        input.AttributeName = flex.ExpandFrameworkStringValueList(ctx, plan.AttributeName)
    }
    ```
    
    To write:
    
    ```go
    plan.AttributeName = flex.FlattenFrameworkStringValueList(output.Thing.AttributeName)
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok && len(v.([]interface{})) > 0 {
        input.AttributeName = flex.ExpandStringList(v.([]interface{}))
    }
    ```
    
    To write:
    
    ```go
    d.Set("attribute_name", aws.StringValueSlice(output.Thing.AttributeName))
    ```

### Root Map of String and AWS Map of String

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        input.AttributeName = flex.ExpandFrameworkStringValueMap(ctx, plan.AttributeName)
    }
    ```
    
    To write:
    
    ```go
    plan.AttributeName = flex.FlattenFrameworkStringValueMap(output.Thing.AttributeName)
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok && len(v.(map[string]interface{})) > 0 {
        input.AttributeName = flex.ExpandStringMap(v.(map[string]interface{}))
    }
    ```

    To write:

    ```go
    d.Set("attribute_name", aws.StringValueMap(output.Thing.AttributeName))
    ```

### Root Set of Resource and AWS List of Structure

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        var tfList []attributeNameData
        resp.Diagnostics.Append(plan.AttributeName.ElementsAs(ctx, &tfList, false)...)
        if resp.Diagnostics.HasError() {
            return
        }

        input.AttributeName = expandStructure(tfList)
    }
    ```
    
    To write:
    
    ```go
    // flattener handles nil output, returning the equivalent null Terraform type
    attributeName, d := flattenStructures(ctx, output.Thing.AttributeName))
    resp.Diagnostics.Append(d...)
    state.AttributeName = attributeName
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok && v.(*schema.Set).Len() > 0 {
        input.AttributeName = expandStructures(v.(*schema.Set).List())
    }
    ```

    To write:

    ```go
    if err := d.Set("attribute_name", flattenStructures(output.Thing.AttributeNames)); err != nil {
        return sdkdiag.AppendErrorf(diags, "setting attribute_name: %s", err)
    }
    ```

### Root Set of String and AWS List of String

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        input.AttributeName = flex.ExpandFrameworkStringValueSet(ctx, plan.AttributeName)
    }
    ```
    
    To write:
    
    ```go
    plan.AttributeName = flex.FlattenFrameworkStringValueSet(output.Thing.AttributeName)
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok && v.(*schema.Set).Len() > 0 {
        input.AttributeName = flex.ExpandStringSet(v.(*schema.Set))
    }
    ```

    To write:

    ```go
    d.Set("attribute_name", aws.StringValueSlice(output.Thing.AttributeName))
    ```

### Root String and AWS String

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        input.AttributeName = aws.String(plan.AttributeName.ValueString())
    }
    ```
    
    To write:
    
    ```go
    plan.AttributeName = flex.StringToFramework(output.Thing.AttributeName)
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok {
        input.AttributeName = aws.String(v.(string))
    }
    ```

    To write:

    ```go
    d.Set("attribute_name", output.Thing.AttributeName)
    ```

### Root String and AWS Timestamp

=== "Terraform Plugin Framework (Preferred)"
    To ensure that parsing the read string value does not fail, use the [RFC3339 timetype](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework-timetypes@v0.3.0/timetypes#RFC3339).

    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if !plan.AttributeName.IsNull() {
        attributeName, d := plan.AttributeName.ValueRFC3339Time()
        resp.Diagnostics.Append(d...)
        input.AttributeName = aws.Time(attributeName)
    }
    ```
    
    To write:
    
    ```go
    plan.AttributeName = timetypes.NewRFC3339ValueMust(aws.ToTime(output.Thing.AttributeName).Format(time.RFC3339))
    ```

=== "Terraform Plugin SDK V2"
    To ensure that parsing the read string value does not fail, define `attribute_name`'s `schema.Schema` with an appropriate [`ValidateFunc`](https://www.terraform.io/plugin/sdkv2/schemas/schema-behaviors#validatefunc):

    ```go
    "attribute_name": {
        Type:         schema.TypeString,
        // ...
        ValidateFunc: validation.IsRFC3339Time,
    },
    ```

    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := d.GetOk("attribute_name"); ok {
        v, _ := time.Parse(time.RFC3339, v.(string))
    
        input.AttributeName = aws.Time(v)
    }
    ```

    To write:

    ```go
    if output.Thing.AttributeName != nil {
        d.Set("attribute_name", aws.TimeValue(output.Thing.AttributeName).Format(time.RFC3339))
    } else {
        d.Set("attribute_name", nil)
    }
    ```

### Nested Bool and AWS Boolean

=== "Terraform Plugin Framework (Preferred)"
    To read, if always sending the attribute value is correct:

    ```go
    func expandStructure(tfList []structureData) *service.Structure {
        // ...

        apiObject.NestedAttributeName = aws.Bool(tfObj.NestedAttributeName.ValueBool())

        // ...
    }
    ```

    To read, if only sending the attribute value when known and not nil:

    ```go
    func expandStructure(tfList []structureData) *service.Structure {
        // ...

        if !tfObj.NestedAttributeName.IsUnknown() && !tfObj.NestedAttributeName.IsNull() {
            apiObject.NestedAttributeName = aws.Bool(tfObj.NestedAttributeName.ValueBool())
        }

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flex will handle setting null when appropriate
        obj["nested_attribute_name"] = flex.BoolToFramework(ctx, apiObject.NestedAttributeName)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read, if always sending the attribute value is correct:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].(bool); ok {
            apiObject.NestedAttributeName = aws.Bool(v)
        }
    
        // ...
    }
    ```

    To read, if only sending the attribute value when `true` is preferred (`!v` for opposite):

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].(bool); ok && v {
            apiObject.NestedAttributeName = aws.Bool(v)
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = aws.BoolValue(v)
        }
    
        // ...
    }
    ```

### Nested Float and AWS Float

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(tfList []structureData) *service.Structure {
        // ...

        if !tfObj.NestedAttributeName.IsUnknown() && !tfObj.NestedAttributeName.IsNull() {
            apiObject.NestedAttributeName = aws.Float64(tfObj.NestedAttributeName.ValueFloat64())
        }

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flex will handle setting null when appropriate
        obj["nested_attribute_name"] = flex.Float64ToFramework(ctx, apiObject.NestedAttributeName)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].(float64); ok && v != 0.0 {
            apiObject.NestedAttributeName = aws.Float64(v)
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = aws.Float64Value(v)
        }
    
        // ...
    }
    ```

### Nested Int and AWS Integer

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(tfList []structureData) *service.Structure {
        // ...

        if !tfObj.NestedAttributeName.IsUnknown() && !tfObj.NestedAttributeName.IsNull() {
            apiObject.NestedAttributeName = aws.Int64(tfObj.NestedAttributeName.ValueInt64())
        }

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flex will handle setting null when appropriate
        obj["nested_attribute_name"] = flex.Int64ToFramework(ctx, apiObject.NestedAttributeName)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].(int); ok && v != 0 {
            apiObject.NestedAttributeName = aws.Int64(int64(v))
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = aws.Int64Value(v)
        }
    
        // ...
    }
    ```

### Nested List of Resource and AWS List of Structure

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(ctx context.Context, tfList []structureData) (*service.Structure, diag.Diagnostics) {
        // ...

        var nested []nestedAttributeNameData
        diags.Append(tfObj.NestedAttributeName.ElementsAs(ctx, &nested, false)...)

        // expand will handle null when appropriate
        apiObject.NestedAttributeName = expandNestedAttributeName(nested)

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flatten will handle setting null when appropriate
        obj["nested_attribute_name"] = flattenNestedAttributeName(ctx, v)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].([]interface{}); ok && len(v) > 0 {
            apiObject.NestedAttributeName = expandStructures(v)
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = flattenNestedStructures(v)
        }
    
        // ...
    }
    ```

### Nested List of Resource and AWS Structure

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(ctx context.Context, tfList []structureData) (*service.Structure, diag.Diagnostics) {
        // ...

        var nested []nestedAttributeNameData
        diags.Append(tfObj.NestedAttributeName.ElementsAs(ctx, &nested, false)...)

        // expand will handle null when appropriate
        apiObject.NestedAttributeName = expandNestedAttributeName(nested)

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flatten will handle setting null when appropriate
        obj["nested_attribute_name"] = flattenNestedAttributeName(ctx, v)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
            apiObject.NestedAttributeName = expandStructure(v[0].(map[string]interface{}))
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = []interface{}{flattenNestedStructure(v)}
        }
    
        // ...
    }
    ```

### Nested List of TypeString and AWS List of String

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(ctx context.Context, tfList []structureData) (*service.Structure, diag.Diagnostics) {
        // ...

        if !tfObj.NestedAttributeName.IsUnknown() && !tfObj.NestedAttributeName.IsNull() {
            apiObject.NestedAttributeName = flex.ExpandFrameworkStringList(ctx, tfObj.NestedAttributeName)
        }

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flex will handle setting null when appropriate
        obj["nested_attribute_name"] = flex.FlattenFrameworkStringList(ctx, v)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].([]interface{}); ok && len(v) > 0 {
            apiObject.NestedAttributeName = flex.ExpandStringList(v)
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = aws.StringValueSlice(v)
        }
    
        // ...
    }
    ```

### Nested Map of String and AWS Map of String

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(ctx context.Context, tfList []structureData) (*service.Structure, diag.Diagnostics) {
        // ...

        if !tfObj.NestedAttributeName.IsUnknown() && !tfObj.NestedAttributeName.IsNull() {
            apiObject.NestedAttributeName = flex.ExpandFrameworkStringMap(ctx, tfObj.NestedAttributeName)
        }

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flex will handle setting null when appropriate
        obj["nested_attribute_name"] = flex.FlattenFrameworkStringMap(ctx, v)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    input := service.ExampleOperationInput{}
    
    if v, ok := tfMap["nested_attribute_name"].(map[string]interface{}); ok && len(v) > 0 {
        apiObject.NestedAttributeName = flex.ExpandStringMap(v)
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = aws.StringValueMap(v)
        }
    
        // ...
    }
    ```

### Nested Set of Resource and AWS List of Structure

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(ctx context.Context, tfList []structureData) (*service.Structure, diag.Diagnostics) {
        // ...

        var nested []nestedAttributeNameData
        diags.Append(tfObj.NestedAttributeName.ElementsAs(ctx, &nested, false)...)

        // expand will handle null when appropriate
        apiObject.NestedAttributeName = expandNestedAttributeName(nested)

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flatten will handle setting null when appropriate
        obj["nested_attribute_name"] = flattenNestedAttributeName(ctx, v)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].(*schema.Set); ok && v.Len() > 0 {
            apiObject.NestedAttributeName = expandStructures(v.List())
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = flattenNestedStructures(v)
        }
    
        // ...
    }
    ```

### Nested Set of TypeString and AWS List of String

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(ctx context.Context, tfList []structureData) (*service.Structure, diag.Diagnostics) {
        // ...

        if !tfObj.NestedAttributeName.IsUnknown() && !tfObj.NestedAttributeName.IsNull() {
            apiObject.NestedAttributeName = flex.ExpandFrameworkStringSet(ctx, tfObj.NestedAttributeName)
        }

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flex will handle setting null when appropriate
        obj["nested_attribute_name"] = flex.FlattenFrameworkStringSet(ctx, v)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].(*schema.Set); ok && v.Len() > 0 {
            apiObject.NestedAttributeName = flex.ExpandStringSet(v)
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = aws.StringValueSlice(v)
        }
    
        // ...
    }
    ```

### Nested TypeString and AWS String

=== "Terraform Plugin Framework (Preferred)"
    To read:

    ```go
    func expandStructure(tfList []structureData) *service.Structure {
        // ...

        if !tfObj.NestedAttributeName.IsUnknown() && !tfObj.NestedAttributeName.IsNull() {
            apiObject.NestedAttributeName = aws.String(tfObj.NestedAttributeName.ValueString())
        }

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        // flex will handle setting null when appropriate
        obj["nested_attribute_name"] = flex.StringToFramework(ctx, apiObject.NestedAttributeName)
    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].(string); ok && v != "" {
            apiObject.NestedAttributeName = aws.String(v)
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = aws.StringValue(v)
        }
    
        // ...
    }
    ```

### Nested String and AWS Timestamp

=== "Terraform Plugin Framework (Preferred)"
    To ensure that parsing the read string value does not fail, use the [RFC3339 timetype](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework-timetypes@v0.3.0/timetypes#RFC3339).

    To read:

    ```go
    func expandStructure(tfList []structureData) (*service.Structure, diag.Diagnostics) {
        // ...

        if !tfObj.NestedAttributeName.IsUnknown() && !tfObj.NestedAttributeName.IsNull() {
            nested := tfObj.NestedAttributeName.ValueRFC3339Time()
            diags.Append(tfObj.NestedAttributeName.ElementsAs(ctx, &nested, false)...)

            apiObject.NestedAttributeName = aws.Time(nested)
        }

        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(ctx context.Context, apiObject *service.Structure) (types.List, diag.Diagnostics) {
        // ...
    
        obj["nested_attribute_name"] = timetypes.NewRFC3339ValueMust(aws.ToTime(apiObject.NestedAttributeName).Format(time.RFC3339))

    
        // ...
    }
    ```

=== "Terraform Plugin SDK V2"
    To ensure that parsing the read string value does not fail, define `nested_attribute_name`'s `schema.Schema` with an appropriate [`ValidateFunc`](https://www.terraform.io/plugin/sdkv2/schemas/schema-behaviors#validatefunc):

    ```go
    "nested_attribute_name": {
        Type:         schema.TypeString,
        // ...
        ValidateFunc: validation.IsRFC3339Time,
    },
    ```

    To read:

    ```go
    func expandStructure(tfMap map[string]interface{}) *service.Structure {
        // ...
    
        if v, ok := tfMap["nested_attribute_name"].(string); ok && v != "" {
            v, _ := time.Parse(time.RFC3339, v)
    
            apiObject.NestedAttributeName = aws.Time(v)
        }
    
        // ...
    }
    ```

    To write:

    ```go
    func flattenStructure(apiObject *service.Structure) map[string]interface{} {
        // ...
    
        if v := apiObject.NestedAttributeName; v != nil {
            tfMap["nested_attribute_name"] = aws.TimeValue(v).Format(time.RFC3339)
        }
    
        // ...
    }
    ```

## Further Guidelines

This section includes additional topics related to data design and decision making from the Terraform AWS Provider maintainers.

### Binary Values

Certain resources may need to interact with binary (non UTF-8) data while the Terraform State only supports UTF-8 data. Configurations attempting to pass binary data to an attribute will receive an error from Terraform CLI. These attributes should expect and store the value as a Base64 string while performing any necessary encoding or decoding in the resource logic.

### Destroy State Values

During resource destroy operations, _only_ previously applied Terraform State values are available to resource logic. Even if the configuration is updated in a manner where both the resource destroy is triggered (e.g., setting the resource meta-argument `count = 0`) and an attribute value is updated, the resource logic will only have the previously applied data values.

Any usage of attribute values during destroy should explicitly note in the resource documentation that the desired value must be applied into the Terraform State before any apply to destroy the resource.

### Hashed Values

Attribute values may be very lengthy or potentially contain [Sensitive Values](#sensitive-values). A potential solution might be to use a hashing algorithm, such as MD5 or SHA256, to convert the value before saving in the Terraform State to reduce its relative size or attempt to obfuscate the value. However, there are a few reasons not to do so:

- Terraform expects any planned values to match applied values. Ensuring proper handling during the various Terraform operations such as difference planning and Terraform State storage can be a burden.
- Hashed values are generally unusable in downstream attribute references. If a value is hashed, it cannot be successfully used in another resource or provider configuration that expects the real value.
- Terraform plan differences are meant to be human-readable. If a value is hashed, operators will only see the relatively unhelpful hash differences `abc123 -> def456` in plans.

Any value hashing implementation will not be accepted. An exception to this guidance is if the remote system explicitly provides a separate hash value in responses, in which a resource can provide a separate attribute with that hashed value.

### Sensitive Values

Marking an Attribute in the Terraform Plugin Framework Schema with `Sensitive` has the following real-world implications:

- All occurrences of the Attribute will have the value hidden in plan difference output. In the context of an Attribute within a Block, all Blocks will hide all values of the Attribute.
- In Terraform CLI 0.14 (with the `provider_sensitive_attrs` experiment enabled) and later, any downstream references to the value in other configurations will hide the value in plan difference output.

The value is either always hidden or not as the Terraform Plugin Framework does not currently implement conditional support for this functionality. Since Terraform Configurations have no control over the behavior, hiding values from the plan difference can incur a potentially undesirable user experience cost for operators.

Given that and especially with the improvements in Terraform CLI 0.14, the Terraform AWS Provider maintainers guiding principles for determining whether an Attribute should be marked as `Sensitive` is if an Attribute value:

- Objectively will always contain a credential, password, or other secret material. Operators can have differing opinions on what constitutes secret material and the maintainers will make best-effort determinations, if necessary consulting with the HashiCorp Security team.
- If the Attribute is within a Block, all occurrences of the Attribute value will objectively contain secret material. Some APIs (and therefore the Terraform AWS Provider resources) implement generic "setting" and "value" structures which likely will contain a mixture of secret and non-secret material. These will generally not be accepted for marking as `Sensitive`.

If you are unsatisfied with sensitive value handling, the maintainers can recommend ensuring there is a covering issue in the Terraform CLI and/or Terraform Plugin Framework projects explaining the use case. Ultimately, Terraform Plugins including the Terraform AWS Provider cannot implement their own sensitive value abilities if the upstream projects do not implement the appropriate functionality.

### Virtual Attributes

Attributes which only exist within Terraform and not the remote system are typically referred to as virtual attributes. Especially in the case of [Destroy State Values](#destroy-state-values), these attributes rely on the [Implicit State Passthrough](#implicit-state-passthrough) behavior of values in Terraform to be available in resource logic. A fictitious example of one of these may be a resource attribute such as a `skip_waiting` flag, which is used only in the resource logic to skip the typical behavior of waiting for operations to complete.

If a virtual attribute has a default value that does not match the [Zero Value Mapping](#zero-value-mapping) for the type, it is recommended to explicitly call `d.Set()` with the default value in the `schema.Resource` `Importer` `State` function, for example:

=== "Teraform Plugin Framework (Preferred)"
    <!-- markdownlint-disable no-space-in-emphasis -->
    ```go
    (r *ThingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
        // ... Other import activity

        resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("skip_waiting"), true)...)
    }
    ```
    <!-- markdownlint-enable no-space-in-emphasis -->

=== "Teraform Plugin SDK V2"
    ```go
    &schema.Resource{
        // ... other fields ...
        Importer: &schema.ResourceImporter{
            State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
                d.Set("skip_waiting", true)

    			return []*schema.ResourceData{d}, nil
    		},
    	},
    }
    ```

This helps prevent an immediate plan difference after resource import unless the configuration has a non-default value.

## Glossary

Below is a listing of relevant terms and descriptions for data handling and conversion in the Terraform AWS Provider to establish common conventions throughout this documentation.
This list is not exhaustive of all concepts of Terraform Plugins, the Terraform AWS Provider, or the data handling that occurs during Terraform runs, but these should generally provide enough context about the topics discussed here.

- **AWS Go SDK**: Library that converts Go code into AWS Service API compatible operations and data types. See [AWS SDK For Go Versions](aws-go-sdk-versions.md) for information on which version to use.
- **AWS Go SDK Model**: AWS Go SDK compatible format of AWS Service API Model.
- **AWS Go SDK Service**: AWS Service API Go code generated from the AWS Go SDK Model. Generated by the AWS Go SDK code.
- **AWS Service API**: Logical boundary of an AWS service by API endpoint. Some large AWS services may be marketed with many different product names under the same service API (e.g., VPC functionality is part of the EC2 API) and vice-versa where some services may be marketed with one product name but are split into multiple service APIs (e.g., Single Sign-On functionality is split into the Identity Store and SSO Admin APIs).
- **AWS Service API Model**: Declarative description of the AWS Service API operations and data types. Generated by the AWS service teams. Used to operate the API and generate API clients such as the various AWS Software Development Kits (SDKs).
- **Terraform Language** ("Configuration"): Configuration syntax interpreted by the Terraform CLI. An implementation of [HCL](https://github.com/hashicorp/hcl). [Full Documentation](https://www.terraform.io/language).
- **Terraform Plugin Protocol**: Description of Terraform Plugin operations and data types. Currently based on the Remote Procedure Call (RPC) library [`gRPC`](https://grpc.io/).
- **Terraform Plugin Go**: Low-level library that converts Go code into Terraform Plugin Protocol compatible operations and data types. Not currently implemented in the Terraform AWS Provider. [Project](https://github.com/hashicorp/terraform-plugin-go).
- **Terraform Plugin Framework**: High-level library that converts Go code into Terraform Plugin Protocol compatible operations and data types. This library replaces Plugin SDK V2. See [Terraform Plugin Development Packages](terraform-plugin-development-packages.md) for more information. [Project](https://github.com/hashicorp/terraform-plugin-framework).
- **Terraform Plugin SDK V2**: High-level library that converts Go code into Terraform Plugin Protocol compatible operations and data types. This library is replaced by Plugin Framework. See [Terraform Plugin Development Packages](terraform-plugin-development-packages.md) for more information. [Project](https://github.com/hashicorp/terraform-plugin-sdk).
- **Terraform Plugin Schema**: Declarative description of types and domain-specific behaviors for a Terraform provider, including resources and attributes. [Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas). [SDK V2 Documentation](https://www.terraform.io/plugin/sdkv2/schemas).
- **Terraform State**: Bindings between objects in a remote system (e.g., an EC2 VPC) and a Terraform configuration (e.g., an `aws_vpc` resource configuration). [Full Documentation](https://www.terraform.io/language/state).

AWS Service API Models use specific terminology to describe data and types:

- **Enumeration**: Collection of valid values for a Shape.
- **Operation**: An API call. Includes information about input, output, and error Shapes.
- **Shape**: Type description.
    - **boolean**: Boolean value.
    - **float**: Fractional numeric value. May contain value validation such as maximum or minimum.
    - **integer**: Whole numeric value. May contain value validation such as maximum or minimum.
    - **list**: Collection that contains member Shapes. May contain value validation such as maximum or minimum keys.
    - **map**: Grouping of key Shape to value Shape. May contain value validation such as maximum or minimum keys.
    - **string**: Sequence of characters. May contain value validation such as an enumeration, regular expression pattern, maximum length, or minimum length.
    - **structure**: Object that contains member Shapes. May represent an error.
    - **timestamp**: Date and time value.

The Terraform Language uses the following terminology to describe data and types:

- **Attribute** ("Argument"): Assigns a name to a data value.
- **Block** ("Configuration Block"): Container type for Attributes or Blocks.
- **null**: Virtual value equivalent to the Attribute not being set.
- **Types**: [Full Documentation](https://www.terraform.io/language/expressions/types).
    - **any**: Virtual type representing any concrete type in type declarations.
    - **bool**: Boolean value.
    - **list** ("tuple"): Ordered collection of values.
    - **map** ("object"): Grouping of string keys to values.
    - **number**: Numeric value. Can be either whole or fractional numbers.
    - **set**: Unordered collection of values.
    - **string**: Sequence of characters.

Terraform Plugin Framework Schemas use the following terminology to describe data and types:

- **Resource Schema**: Grouping of Schema that represents a Terraform Resource.
- **Schema**: Represents an Attribute or Block. Has a Type and Behavior(s).
- **Types**: [Full Documentation](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types).
    - **Bool**: Boolean value.
    - **Float64**: Fractional numeric value.
    - **Int64**: Whole numeric value.
    - **List**: An ordered collection of values or Blocks.
    - **Map**: Grouping of key Type to value Type.
    - **Set**: Unordered collection of values or Blocks.
    - **String**: Sequence of characters value.

Terraform Plugin SDK Schemas use the following terminology to describe data and types:

- **Behaviors**: [Full Documentation](https://www.terraform.io/plugin/sdkv2/schemas/schema-behaviors).
    - **Sensitive**: Whether the value should be hidden from user interface output.
    - **StateFunc**: Conversion function between the value set by the Terraform Plugin and the value seen by Terraform Plugin SDK (and ultimately the Terraform State).
- **Element**: Underlying value type for a collection or grouping Schema.
- **Resource Data**: Data representation of a Resource Schema. Translation layer between the Schema and Go code of a Terraform Plugin. In the Terraform Plugin SDK, the `ResourceData` Go type.
- **Resource Schema**: Grouping of Schema that represents a Terraform Resource.
- **Schema**: Represents an Attribute or Block. Has a Type and Behavior(s).
- **Types**: [Full Documentation](https://www.terraform.io/plugin/sdkv2/schemas/schema-types).
    - **TypeBool**: Boolean value.
    - **TypeFloat**: Fractional numeric value.
    - **TypeInt**: Whole numeric value.
    - **TypeList**: Ordered collection of values or Blocks.
    - **TypeMap**: Grouping of key Type to value Type.
    - **TypeSet**: Unordered collection of values or Blocks.
    - **TypeString**: Sequence of characters value.

Some other terms that may be used:

- **Block Attribute** ("Child Attribute", "Nested Attribute"): Block level Attribute.
- **Expand Function**: Function that converts Terraform Plugin SDK data into the equivalent AWS Go SDK type.
- **Flatten Function**: Function that converts an AWS Go SDK type into the equivalent Terraform Plugin SDK data.
- **NullableTypeBool**: (SDK V2 Only) Workaround "schema type" created to accept a boolean value that is not configured in addition to true and false. Not implemented in the Terraform Plugin SDK, but uses `TypeString` (where `""` represents not configured) and additional validation.
- **NullableTypeFloat**: (SDK V2 Only) Workaround "schema type" created to accept a fractional numeric value that is not configured in addition to `0.0`. Not implemented in the Terraform Plugin SDK, but uses `TypeString` (where `""` represents not configured) and additional validation.
- **NullableTypeInt**: (SDK V2 Only) Workaround "schema type" created to accept a whole numeric value that is not configured in addition to `0`. Not implemented in the Terraform Plugin SDK, but uses `TypeString` (where `""` represents not configured) and additional validation.
- **Root Attribute: Resource top-level Attribute or Block.

For additional reference, the Terraform documentation also includes a [full glossary of terminology](https://www.terraform.io/docs/glossary.html).
