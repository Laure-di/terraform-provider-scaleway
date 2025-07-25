---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container"
---

# Resource: scaleway_container

The `scaleway_container` resource allows you to create and manage [Serverless Containers](https://www.scaleway.com/en/docs/serverless/containers/).

Refer to the Serverless Containers [product documentation](https://www.scaleway.com/en/docs/serverless/containers/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-containers/) for more information.

For more information on the limitations of Serverless Containers, refer to the [dedicated documentation](https://www.scaleway.com/en/docs/serverless-containers/reference-content/containers-limitations/).

## Example Usage

```terraform
resource "scaleway_container_namespace" "main" {
  name        = "my-ns-test"
  description = "test container"
}

resource "scaleway_container" "main" {
  name            = "my-container-02"
  description     = "environment variables test"
  tags            = ["tag1", "tag2"]
  namespace_id    = scaleway_container_namespace.main.id
  registry_image  = "${scaleway_container_namespace.main.registry_endpoint}/alpine:test"
  port            = 9997
  cpu_limit       = 140
  memory_limit    = 256
  min_scale       = 3
  max_scale       = 5
  timeout         = 600
  max_concurrency = 80
  privacy         = "private"
  protocol        = "http1"
  deploy          = true

  command = [ "bash", "-c", "script.sh" ]
  args =    [ "some", "args" ]

  environment_variables = {
    "foo" = "var"
  }
  secret_environment_variables = {
    "key" = "secret"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The unique name of the container name.

- `namespace_id` - (Required) The Containers namespace ID of the container.

~> **Important** Updating the `name` argument will recreate the container.

- `description` (Optional) The description of the container.

- `tags` - (Optional) The list of tags associated with the container.

- `environment_variables` - (Optional) The [environment variables](https://www.scaleway.com/en/docs/serverless-containers/concepts/#environment-variables) of the container.

- `secret_environment_variables` - (Optional) The [secret environment variables](https://www.scaleway.com/en/docs/serverless-containers/concepts/#secrets) of the container.

- `min_scale` - (Optional) The minimum number of container instances running continuously.

- `max_scale` - (Optional) The maximum number of instances this container can scale to.

- `memory_limit` - (Optional) The memory resources in MB to allocate to each container.

- `cpu_limit` - (Optional) The amount of vCPU computing resources to allocate to each container.

- `timeout` - (Optional) The maximum amount of time in seconds your container can spend processing a request before being stopped. Default to `300` seconds.

- `privacy` - (Optional) The privacy type defines the way to authenticate to your container. Please check our dedicated [section](https://www.scaleway.com/en/developers/api/serverless-containers/#protocol-9dd4c8).

- `registry_image` - (Optional) The registry image address (e.g., `rg.fr-par.scw.cloud/$NAMESPACE/$IMAGE`)

- `registry_sha256` - (Optional) The sha256 of your source registry image, changing it will re-apply the deployment. Can be any string.

- `max_concurrency` - (Deprecated) The maximum number of simultaneous requests your container can handle at the same time. Use `scaling_option.concurrent_requests_threshold` instead.

- `protocol` - (Optional) The communication [protocol](https://www.scaleway.com/en/developers/api/serverless-containers/#path-containers-update-an-existing-container) `http1` or `h2c`. Defaults to `http1`.

- `http_option` - (Optional) Allows both HTTP and HTTPS (`enabled`) or redirect HTTP to HTTPS (`redirected`). Defaults to `enabled`.

- `sandbox` - (Optional) Execution environment of the container.

- `health_check` - (Optional) Health check configuration block of the container.
    - `http` - HTTP health check configuration.
        - `path` - Path to use for the HTTP health check.
    - `failure_threshold` - Number of consecutive health check failures before considering the container unhealthy.
    - `interval`- Period between health checks (in seconds).

- `scaling_option` - (Optional) Configuration block used to decide when to scale up or down. Possible values:
    - `concurrent_requests_threshold` - Scale depending on the number of concurrent requests being processed per container instance.
    - `cpu_usage_threshold` - Scale depending on the CPU usage of a container instance.
    - `memory_usage_threshold`- Scale depending on the memory usage of a container instance.

- `port` - (Optional) The port to expose the container.

- `deploy` - (Optional) Boolean indicating whether the container is in a production environment.

- `local_storage_limit` - (Optional) Local storage limit of the container (in MB)

- `command` - (Optional) Command executed when the container starts. This overrides the default command defined in the container image. This is usually the main executable, or entry point script to run.

- `args` - (Optional) Arguments passed to the command specified in the "command" field. These override the default arguments from the container image, and behave like command-line parameters.

- `private_network_id` (Optional) The ID of the Private Network the container is connected to.

~> **Important** This feature is currently in beta and requires a namespace with VPC integration activated by setting the `activate_vpc_integration` attribute to `true`.

Note that if you want to use your own configuration, you must consult our configuration [restrictions](https://www.scaleway.com/en/docs/serverless-containers/reference-content/containers-limitations/#configuration-restrictions) section.

## Attributes Reference

The `scaleway_container` resource exports certain attributes once the Container is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the container.

~> **Important:** Container IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the container was created.

- `status` - The container status.

- `cron_status` - The cron status of the container.

- `error_message` - The error message of the container.

- `domain_name` - The native domain name of the container

## Import

Containers can be imported using, `{region}/{id}`, as shown below:

```bash
terraform import scaleway_container.main fr-par/11111111-1111-1111-1111-111111111111
```

## Protocols

The following protocols are supported:

* `h2c`: HTTP/2 over TCP.
* `http1`: Hypertext Transfer Protocol.

~> **Important:** Refer to the official [Apache documentation](https://httpd.apache.org/docs/2.4/howto/http2.html) for more information.

## Privacy

By default, creating a container will make it `public`, meaning that anybody knowing the endpoint can execute it.

A container can be made `private` with the privacy parameter.

Refer to the [technical information](https://www.scaleway.com/en/developers/api/serverless-containers/#protocol-9dd4c8) for more information on container authentication.

## Memory and vCPUs configuration

The vCPU represents a portion of the underlying, physical CPU that is assigned to a particular virtual machine (VM).

You can determine the computing resources to allocate to each container.

The `memory_limit` (in MB) must correspond with the right amount of vCPU. Refer to the table below to determine the right memory/vCPU combination.

| Memory (in MB) | vCPU |
|----------------|------|
| 128            | 70m  |
| 256            | 140m |
| 512            | 280m |
| 1024           | 560m |
| 2048           | 1120 |
| 3072           | 1680 |
| 4096           | 2240 |

~>**Important:** Make sure to select the right resources, as you will be billed based on compute usage over time and the number of Containers executions.
Refer to the [Serverless Containers pricing](https://www.scaleway.com/en/docs/faq/serverless-containers/#prices) for more information.

## Health check configuration

Custom health checks can be configured on the container.

It's possible to specify the HTTP path that the probe will listen to and the number of failures before considering the container as unhealthy.
During a deployment, if a newly created container fails to pass the health check, the deployment is aborted.
As a result, lowering this value can help to reduce the time it takes to detect a failed deployment.
The period between health checks is also configurable.

Example:

```terraform
resource "scaleway_container" "main" {
  name         = "my-container-02"
  namespace_id = scaleway_container_namespace.main.id

  health_check {
    http {
      path = "/ping"
    }
    failure_threshold = 40
    interval          = "5s"
  }
}
```

~>**Important:** Another probe type can be set to TCP with the API, but currently the SDK has not been updated with this parameter.
This is why the only probe that can be used here is the HTTP probe.
Refer to the [Serverless Containers pricing](https://www.scaleway.com/en/docs/faq/serverless-containers/#prices) for more information.

## Scaling option configuration

Scaling option block configuration allows you to choose which parameter will scale up/down containers.
Options are number of concurrent requests, CPU or memory usage.
It replaces current `max_concurrency` that has been deprecated.

Example:

```terraform
resource "scaleway_container" "main" {
  name         = "my-container-02"
  namespace_id = scaleway_container_namespace.main.id

  scaling_option {
    concurrent_requests_threshold = 15
  }
}
```

~>**Important**: A maximum of one of these parameters may be set. Also, when `cpu_usage_threshold` or `memory_usage_threshold` are used, `min_scale` can't be set to 0.
Refer to the [API Reference](https://www.scaleway.com/en/developers/api/serverless-containers/#path-containers-create-a-new-container) for more information.
