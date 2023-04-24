package ec2

import (
  "context"
  "time"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/service/ec2"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
  "github.com/hashicorp/terraform-provider-aws/internal/conns"
  "github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)


func DataSourceTransitGatewayRouteTableRoutes() *schema.Resource {
  return &schema.Resource{
     ReadWithoutTimeout: dataSourceTransitGatewayRouteTableRoutes,
     Timeouts: &schema.ResourceTimeout{
      Read: schema.DefaultTimeout(20 * time.Minute),
    },
    
    Schema: map[string]*schema.Schema{
      "filter": DataSourceFiltersSchema(),
      "transit_gateway_route_table_id": {
        Type:         schema.TypeString,
        Required:     true,
        ValidateFunc: validation.NoZeroValues,
      },
      "routes": {
        Type: schema.TypeList,
        Computed: true,
        Elem:     &schema.Resource{
          Schema: map[string]*schema.Schema{
            "destination_cidr_block": {
              Type:     schema.TypeString,
              Computed: true,
            },
            "prefix_list_id": {
              Type:     schema.TypeString,
              Computed: true,
            },
            "state": {
              Type:     schema.TypeString,
              Computed: true,
            },
            "transit_gateway_route_table_announcement_id": {
              Type:     schema.TypeString,
              Computed: true,
            },
            "type": {
              Type:     schema.TypeString,
              Computed: true,
            },
          },
        },
      }
    },
  }
}


func dataSourceTransitGatewayRouteTableRoutesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
  var diags diag.Diagnostics
  conn := meta.(*conns.AWSClient).EC2Conn()
  input := &ec2.SearchTransitGatewayRoutesInput{}
  if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
    input.TransitGatewayRouteTableId = aws.String(v.(string))
  }
  input.Filters = append(input.Filters, BuildFiltersDataSource(
    d.Get("filter").(*schema.Set),
  )...)

  if len(input.Filters) == 0 {
    input.Filters = nil
  }
  output, err := conn.SearchTransitGatewayRoutesWithContex(ctx, conn, input)
  if err != nil {
    return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Routes: %s", err)
  }

  if output == nil || len(output.Routes) == 0 {
    return nil, tfresource.NewEmptyResultError(input)
  }

  var routes []&ec2.TransitGatewayRoute
  for _, route := range output.Routes {
    if route == nil {
      continue
    } else {
      routes.append(route)
    }
  }

  d.SetId(meta.(*conns.AWSClient).Region)
  d.Set("routes",routes)
  return diags
}
