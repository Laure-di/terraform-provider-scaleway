package inference

//import (
//	"context"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//	"github.com/scaleway/scaleway-sdk-go/api/inference/v1"
//)
//
//func ResourceEndpoint() *schema.Resource {
//	return &schema.Resource{
//		CreateContext: ResourceEndpointCreate,
//		ReadContext:   ResourceEndpointRead,
//		UpdateContext: ResourceEndpointUpdate,
//		DeleteContext: ResourceEndpointDelete,
//		Importer: &schema.ResourceImporter{
//			StateContext: schema.ImportStatePassthroughContext,
//		},
//		SchemaVersion: 0,
//		Schema: map[string]*schema.Schema{
//			"deployment_id": {
//				Type:     schema.TypeString,
//				Required: true,
//				ForceNew: true,
//			},
//			"private_endpoint": {
//				Type:         schema.TypeList,
//				Optional:     true,
//				MaxItems:     1,
//				AtLeastOneOf: []string{"public_endpoint"},
//				Description:  "List of endpoints",
//				Elem: &schema.Resource{
//					Schema: map[string]*schema.Schema{
//						"id": {
//							Type:        schema.TypeString,
//							Description: "The id of the private endpoint",
//							Computed:    true,
//						},
//						"private_network_id": {
//							Type:        schema.TypeString,
//							Description: "The id of the private network",
//							Optional:    true,
//						},
//						"disable_auth": {
//							Type:        schema.TypeBool,
//							Description: "Disable the authentication on the endpoint.",
//							Optional:    true,
//							Default:     false,
//						},
//						"url": {
//							Type:        schema.TypeString,
//							Description: "The URL of the endpoint.",
//							Computed:    true,
//						},
//					},
//				},
//			},
//
//			"public_endpoint": {
//				Type:         schema.TypeList,
//				Optional:     true,
//				AtLeastOneOf: []string{"private_endpoint"},
//				Description:  "Public endpoints",
//				MaxItems:     1,
//				Elem: &schema.Resource{
//					Schema: map[string]*schema.Schema{
//						"id": {
//							Type:        schema.TypeString,
//							Description: "The id of the public endpoint",
//							Computed:    true,
//						},
//						"is_enabled": {
//							Type:        schema.TypeBool,
//							Description: "Enable or disable public endpoint",
//							Optional:    true,
//						},
//						"disable_auth": {
//							Type:        schema.TypeBool,
//							Description: "Disable the authentication on the endpoint.",
//							Optional:    true,
//							Default:     false,
//						},
//						"url": {
//							Type:        schema.TypeString,
//							Description: "The URL of the endpoint.",
//							Computed:    true,
//						},
//					},
//				},
//			},
//		},
//	}
//}
//
//func ResourceEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	api, region, err := NewAPIWithRegion(d, m)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	req := &inference.CreateEndpointRequest{
//		DeploymentID: d.Get("deployment_id").(string),
//	}
//
//	var endpoints []*inference.EndpointSpec
//
//	if privateEndpoint, ok := d.GetOk("private_endpoint"); ok {
//		privateEndpointMap := privateEndpoint.([]interface{})[0].(map[string]interface{})
//		endpoints = append(endpoints, &inference.EndpointSpec{
//			PublicNetwork: nil,
//			DisableAuth:   privateEndpointMap["disable_auth"].(bool),
//			PrivateNetwork:
//		})
//	}
//
//	return diag.Errorf("not yet implemented")
//}
//
//func ResourceEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	return diag.Errorf("not yet implemented")
//}
//
//func ResourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	return diag.Errorf("not yet implemented")
//}
//
//func ResourceEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	return diag.Errorf("not yet implemented")
//}
