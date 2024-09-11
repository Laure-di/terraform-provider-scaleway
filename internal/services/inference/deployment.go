package inference

import (
	"context"

	inference "github.com/scaleway/scaleway-sdk-go/api/inference/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	_ "time"
)

func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceDeploymentCreate,
		ReadContext:   ResourceDeploymentRead,
		UpdateContext: ResourceDeploymentUpdate,
		DeleteContext: ResourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{ // TODO: remove unused timeouts
			Create:  schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
			Read:    schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
			Update:  schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
			Delete:  schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
			Default: schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The deployment name",
			},
			"region":          regional.Schema(),
			"project_id":      account.ProjectIDSchema(),
			"organization_id": account.OrganizationIDSchema(),
			"node_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The node type to use for the deployment",
			},
			"model_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The model name to use for the deployment",
			},
			"accept_eula": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not the deployment is accepting eula",
			},
			"tags": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "The tags associated with the deployment",
			},
			"min_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum size of the pool",
			},
			"max_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum size of the pool",
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the pool",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the deployment",
			},
			"endpoint_public_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint public URL",
			},
			"endpoint_private_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint private URL",
			},
			"endpoint_public_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint public ID",
			},
			"endpoint_private_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint private ID",
			},
			"endpoints": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of endpoints",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"public_endpoint": {
							Type:        schema.TypeBool,
							Description: "Set the endpoint as public",
							Optional:    true,
						},
						"private_endpoint": {
							Type:        schema.TypeString,
							Description: "The id of the private network",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func ResourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	endpoint := inference.EndpointSpec{
		Public:         nil,
		PrivateNetwork: nil,
		DisableAuth:    false,
	}

	req := &inference.CreateDeploymentRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      d.Get("name").(string),
		NodeType:  d.Get("node_type").(string),
		ModelName: d.Get("model_name").(string),
		Tags:      types.ExpandStrings(d.Get("tags")),
	}

	if _, isEndpoint := d.GetOk("endpoints"); isEndpoint {
		if publicEndpoint := d.Get("endpoints.0.public_endpoint"); publicEndpoint != nil && publicEndpoint.(bool) {
			endpoint.Public = &inference.EndpointSpecPublic{}
		}
		if privateEndpoint := d.Get("endpoints.0.private_endpoint"); privateEndpoint != "" {
			endpoint.PrivateNetwork = &inference.EndpointSpecPrivateNetwork{
				PrivateNetworkID: privateEndpoint.(string),
			}
		}
	}

	if maxSize, ok := d.GetOk("max_size"); ok {
		req.MaxSize = scw.Uint32Ptr(uint32(maxSize.(int)))
	}

	if minSize, ok := d.GetOk("min_size"); ok {
		req.MaxSize = scw.Uint32Ptr(uint32(minSize.(int)))
	}

	req.Endpoints = []*inference.EndpointSpec{&endpoint}

	if isAcceptingEula, ok := d.GetOk("accept_eula"); ok {
		req.AcceptEula = scw.BoolPtr(isAcceptingEula.(bool))
	}

	deployment, err := api.CreateDeployment(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, deployment.ID))

	_, err = waitForDeployment(ctx, api, region, deployment.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceDeploymentRead(ctx, d, m)
}

func ResourceDeploymentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err := waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", deployment.Name)
	_ = d.Set("region", deployment.Region)
	_ = d.Set("project_id", deployment.ProjectID)
	_ = d.Set("node_type", deployment.NodeType)
	_ = d.Set("model_name", deployment.ModelName)
	_ = d.Set("min_size", deployment.MinSize)
	_ = d.Set("max_size", deployment.MaxSize)
	_ = d.Set("size", deployment.Size)
	_ = d.Set("status", deployment.Status)

	return nil
}

func ResourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err := waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	req := &inference.UpdateDeploymentRequest{
		Region:       region,
		DeploymentID: deployment.ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if _, err := api.UpdateDeployment(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceDeploymentRead(ctx, d, m)
}

func ResourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = api.DeleteDeployment(&inference.DeleteDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}