package container

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultContainerNamespaceTimeout = 5 * time.Minute
	defaultContainerCronTimeout      = 5 * time.Minute
	defaultContainerTimeout          = 12*time.Minute + 30*time.Second
	defaultContainerDomainTimeout    = 10 * time.Minute
	DefaultContainerRetryInterval    = 5 * time.Second
	defaultTriggerRetryInterval      = 5 * time.Second
)

// newAPIWithRegion returns a new container API and the region.
func newAPIWithRegion(d *schema.ResourceData, m any) (*container.API, scw.Region, error) {
	api := container.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a new container API, region and ID.
func NewAPIWithRegionAndID(m any, id string) (*container.API, scw.Region, string, error) {
	api := container.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

func setCreateContainerRequest(d *schema.ResourceData, region scw.Region) (*container.CreateContainerRequest, error) {
	// required
	nameRaw := d.Get("name")
	namespaceID := d.Get("namespace_id")

	name := types.ExpandOrGenerateString(nameRaw.(string), "co")
	privacyType := d.Get("privacy")
	protocol := d.Get("protocol")
	httpOption := d.Get("http_option")

	req := &container.CreateContainerRequest{
		Region:      region,
		NamespaceID: locality.ExpandID(namespaceID),
		Name:        name,
		Privacy:     container.ContainerPrivacy(privacyType.(string)),
		Protocol:    container.ContainerProtocol(*types.ExpandStringPtr(protocol)),
		HTTPOption:  container.ContainerHTTPOption(httpOption.(string)),
	}

	// optional
	if envVariablesRaw, ok := d.GetOk("environment_variables"); ok {
		req.EnvironmentVariables = types.ExpandMapPtrStringString(envVariablesRaw)
	}

	if secretEnvVariablesRaw, ok := d.GetOk("secret_environment_variables"); ok {
		req.SecretEnvironmentVariables = expandContainerSecrets(secretEnvVariablesRaw)
	}

	if minScale, ok := d.GetOk("min_scale"); ok {
		req.MinScale = scw.Uint32Ptr(uint32(minScale.(int)))
	}

	if maxScale, ok := d.GetOk("max_scale"); ok {
		req.MaxScale = scw.Uint32Ptr(uint32(maxScale.(int)))
	}

	if memoryLimit, ok := d.GetOk("memory_limit"); ok {
		req.MemoryLimit = scw.Uint32Ptr(uint32(memoryLimit.(int)))
	}

	if cpuLimit, ok := d.GetOk("cpu_limit"); ok {
		req.CPULimit = scw.Uint32Ptr(uint32(cpuLimit.(int)))
	}

	if timeout, ok := d.GetOk("timeout"); ok {
		timeInt := timeout.(int)
		req.Timeout = &scw.Duration{Seconds: int64(timeInt)}
	}

	if port, ok := d.GetOk("port"); ok {
		req.Port = scw.Uint32Ptr(uint32(port.(int)))
	}

	if description, ok := d.GetOk("description"); ok {
		req.Description = types.ExpandStringPtr(description)
	}

	if registryImage, ok := d.GetOk("registry_image"); ok {
		req.RegistryImage = types.ExpandStringPtr(registryImage)
	}

	if maxConcurrency, ok := d.GetOk("max_concurrency"); ok {
		req.MaxConcurrency = scw.Uint32Ptr(uint32(maxConcurrency.(int))) //nolint:staticcheck
	}

	if sandbox, ok := d.GetOk("sandbox"); ok {
		req.Sandbox = container.ContainerSandbox(sandbox.(string))
	}

	if healthCheck, ok := d.GetOk("health_check"); ok {
		healthCheckReq, errExpandHealthCheck := expandHealthCheck(healthCheck)
		if errExpandHealthCheck != nil {
			return nil, errExpandHealthCheck
		}

		req.HealthCheck = healthCheckReq
	}

	if scalingOption, ok := d.GetOk("scaling_option"); ok {
		scalingOptionReq, err := expandScalingOptions(scalingOption)
		if err != nil {
			return nil, err
		}

		req.ScalingOption = scalingOptionReq
	}

	if localStorageLimit, ok := d.GetOk("local_storage_limit"); ok {
		req.LocalStorageLimit = scw.Uint32Ptr(uint32(localStorageLimit.(int)))
	}

	if tags, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(tags)
	}

	if command, ok := d.GetOk("command"); ok {
		req.Command = types.ExpandStrings(command)
	}

	if args, ok := d.GetOk("args"); ok {
		req.Args = types.ExpandStrings(args)
	}

	if pnID, ok := d.GetOk("private_network_id"); ok {
		req.PrivateNetworkID = types.ExpandStringPtr(locality.ExpandID(pnID.(string)))
	}

	return req, nil
}

func setUpdateContainerRequest(d *schema.ResourceData, region scw.Region, containerID string) (*container.UpdateContainerRequest, error) {
	req := &container.UpdateContainerRequest{
		Region:      region,
		ContainerID: containerID,
	}

	if d.HasChanges("environment_variables") {
		envVariablesRaw := d.Get("environment_variables")
		req.EnvironmentVariables = types.ExpandMapPtrStringString(envVariablesRaw)
	}

	if d.HasChanges("secret_environment_variables") {
		oldEnv, newEnv := d.GetChange("secret_environment_variables")
		req.SecretEnvironmentVariables = filterSecretEnvsToPatch(expandContainerSecrets(oldEnv), expandContainerSecrets(newEnv))
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChanges("min_scale") {
		req.MinScale = scw.Uint32Ptr(uint32(d.Get("min_scale").(int)))
	}

	if d.HasChanges("max_scale") {
		req.MaxScale = scw.Uint32Ptr(uint32(d.Get("max_scale").(int)))
	}

	if d.HasChanges("memory_limit") {
		req.MemoryLimit = scw.Uint32Ptr(uint32(d.Get("memory_limit").(int)))
	}

	if d.HasChanges("cpu_limit") {
		req.CPULimit = scw.Uint32Ptr(uint32(d.Get("cpu_limit").(int)))
	}

	if d.HasChanges("timeout") {
		req.Timeout = &scw.Duration{Seconds: int64(d.Get("timeout").(int))}
	}

	if d.HasChanges("privacy") {
		req.Privacy = container.ContainerPrivacy(*types.ExpandStringPtr(d.Get("privacy")))
	}

	if d.HasChanges("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}

	if d.HasChanges("registry_image") {
		req.RegistryImage = types.ExpandStringPtr(d.Get("registry_image"))
	}

	if d.HasChanges("max_concurrency") {
		req.MaxConcurrency = scw.Uint32Ptr(uint32(d.Get("max_concurrency").(int))) //nolint:staticcheck
	}

	if d.HasChanges("protocol") {
		req.Protocol = container.ContainerProtocol(*types.ExpandStringPtr(d.Get("protocol")))
	}

	if d.HasChanges("port") {
		req.Port = scw.Uint32Ptr(uint32(d.Get("port").(int)))
	}

	if d.HasChanges("http_option") {
		req.HTTPOption = container.ContainerHTTPOption(d.Get("http_option").(string))
	}

	if d.HasChanges("deploy") {
		req.Redeploy = types.ExpandBoolPtr(d.Get("deploy"))
	}

	if d.HasChanges("sandbox") {
		req.Sandbox = container.ContainerSandbox(d.Get("sandbox").(string))
	}

	if d.HasChanges("health_check") {
		healthCheck := d.Get("health_check")

		healthCheckReq, errExpandHealthCheck := expandHealthCheck(healthCheck)
		if errExpandHealthCheck != nil {
			return nil, errExpandHealthCheck
		}

		req.HealthCheck = healthCheckReq
	}

	if d.HasChanges("scaling_option") {
		scalingOption := d.Get("scaling_option")

		scalingOptionReq, err := expandScalingOptions(scalingOption)
		if err != nil {
			return nil, err
		}

		req.ScalingOption = scalingOptionReq
	}

	imageHasChanged := d.HasChanges("registry_sha256")
	if imageHasChanged {
		req.Redeploy = &imageHasChanged
	}

	if d.HasChanges("local_storage_limit") {
		req.LocalStorageLimit = scw.Uint32Ptr(uint32(d.Get("local_storage_limit").(int)))
	}

	if d.HasChanges("command") {
		req.Command = types.ExpandUpdatedStringsPtr(d.Get("command"))
	}

	if d.HasChanges("args") {
		req.Args = types.ExpandUpdatedStringsPtr(d.Get("args"))
	}

	if d.HasChanges("private_network_id") {
		req.PrivateNetworkID = types.ExpandUpdatedStringPtr(locality.ExpandID(d.Get("private_network_id")))
	}

	return req, nil
}

func expandHealthCheck(healthCheckSchema any) (*container.ContainerHealthCheckSpec, error) {
	healthCheck, ok := healthCheckSchema.(*schema.Set)
	if !ok {
		return &container.ContainerHealthCheckSpec{}, nil
	}

	for _, option := range healthCheck.List() {
		rawOption, isRawOption := option.(map[string]any)
		if !isRawOption {
			continue
		}

		healthCheckSpec := &container.ContainerHealthCheckSpec{}
		if http, ok := rawOption["http"].(*schema.Set); ok {
			healthCheckSpec.HTTP = expendHealthCheckHTTP(http)
		}

		// Failure threshold is a required field and will be checked by TF.
		healthCheckSpec.FailureThreshold = uint32(rawOption["failure_threshold"].(int))

		if interval, ok := rawOption["interval"]; ok {
			duration, err := types.ExpandDuration(interval)
			if err != nil {
				return nil, err
			}

			healthCheckSpec.Interval = scw.NewDurationFromTimeDuration(*duration)
		}

		return healthCheckSpec, nil
	}

	return &container.ContainerHealthCheckSpec{}, nil
}

func expendHealthCheckHTTP(healthCheckHTTPSchema any) *container.ContainerHealthCheckSpecHTTPProbe {
	healthCheckHTTP, ok := healthCheckHTTPSchema.(*schema.Set)
	if !ok {
		return &container.ContainerHealthCheckSpecHTTPProbe{}
	}

	for _, option := range healthCheckHTTP.List() {
		rawOption, isRawOption := option.(map[string]any)
		if !isRawOption {
			continue
		}

		httpProbe := &container.ContainerHealthCheckSpecHTTPProbe{}
		if path, ok := rawOption["path"].(string); ok {
			httpProbe.Path = path
		}

		return httpProbe
	}

	return &container.ContainerHealthCheckSpecHTTPProbe{}
}

func flattenHealthCheck(healthCheck *container.ContainerHealthCheckSpec) any {
	if healthCheck == nil {
		return nil
	}

	var interval *time.Duration
	if healthCheck.Interval != nil {
		interval = healthCheck.Interval.ToTimeDuration()
	}

	flattenedHealthCheck := []map[string]any(nil)
	flattenedHealthCheck = append(flattenedHealthCheck, map[string]any{
		"http":              flattenHealthCheckHTTP(healthCheck.HTTP),
		"failure_threshold": types.FlattenUint32Ptr(&healthCheck.FailureThreshold),
		"interval":          types.FlattenDuration(interval),
	})

	return flattenedHealthCheck
}

func flattenHealthCheckHTTP(healthCheckHTTP *container.ContainerHealthCheckSpecHTTPProbe) any {
	if healthCheckHTTP == nil {
		return nil
	}

	flattenedHealthCheckHTTP := []map[string]any(nil)
	flattenedHealthCheckHTTP = append(flattenedHealthCheckHTTP, map[string]any{
		"path": types.FlattenStringPtr(&healthCheckHTTP.Path),
	})

	return flattenedHealthCheckHTTP
}

func expandScalingOptions(scalingOptionSchema any) (*container.ContainerScalingOption, error) {
	scalingOption, ok := scalingOptionSchema.(*schema.Set)
	if !ok {
		return &container.ContainerScalingOption{}, nil
	}

	for _, option := range scalingOption.List() {
		rawOption, isRawOption := option.(map[string]any)
		if !isRawOption {
			continue
		}

		setFields := 0

		cso := &container.ContainerScalingOption{}
		if concurrentRequestThresold, ok := rawOption["concurrent_requests_threshold"].(int); ok && concurrentRequestThresold != 0 {
			cso.ConcurrentRequestsThreshold = scw.Uint32Ptr(uint32(concurrentRequestThresold))
			setFields++
		}

		if cpuUsageThreshold, ok := rawOption["cpu_usage_threshold"].(int); ok && cpuUsageThreshold != 0 {
			cso.CPUUsageThreshold = scw.Uint32Ptr(uint32(cpuUsageThreshold))
			setFields++
		}

		if memoryUsageThreshold, ok := rawOption["memory_usage_threshold"].(int); ok && memoryUsageThreshold != 0 {
			cso.MemoryUsageThreshold = scw.Uint32Ptr(uint32(memoryUsageThreshold))
			setFields++
		}

		if setFields > 1 {
			return &container.ContainerScalingOption{}, errors.New("a maximum of one scaling option can be set")
		}

		return cso, nil
	}

	return &container.ContainerScalingOption{}, nil
}

func flattenScalingOption(scalingOption *container.ContainerScalingOption) any {
	if scalingOption == nil {
		return nil
	}

	flattenedScalingOption := []map[string]any(nil)
	flattenedScalingOption = append(flattenedScalingOption, map[string]any{
		"concurrent_requests_threshold": types.FlattenUint32Ptr(scalingOption.ConcurrentRequestsThreshold),
		"cpu_usage_threshold":           types.FlattenUint32Ptr(scalingOption.CPUUsageThreshold),
		"memory_usage_threshold":        types.FlattenUint32Ptr(scalingOption.MemoryUsageThreshold),
	})

	return flattenedScalingOption
}

func flattenContainerSecrets(secrets []*container.SecretHashedValue) any {
	if len(secrets) == 0 {
		return nil
	}

	flattenedSecrets := make(map[string]any)
	for _, secret := range secrets {
		flattenedSecrets[secret.Key] = secret.HashedValue
	}

	return flattenedSecrets
}

func expandContainerSecrets(secretsRawMap any) []*container.Secret {
	secretsMap := secretsRawMap.(map[string]any)
	secrets := make([]*container.Secret, 0, len(secretsMap))

	for k, v := range secretsMap {
		secrets = append(secrets, &container.Secret{
			Key:   k,
			Value: types.ExpandStringPtr(v),
		})
	}

	return secrets
}

func isContainerDNSResolveError(err error) bool {
	responseError := &scw.ResponseError{}

	if !errors.As(err, &responseError) {
		return false
	}

	if strings.HasPrefix(responseError.Message, "could not validate domain") {
		return true
	}

	return false
}

func retryCreateContainerDomain(ctx context.Context, containerAPI *container.API, req *container.CreateDomainRequest, timeout time.Duration) (*container.Domain, error) {
	timeoutChannel := time.After(timeout)

	for {
		select {
		case <-time.After(DefaultContainerRetryInterval):
			domain, err := containerAPI.CreateDomain(req, scw.WithContext(ctx))
			if err != nil && isContainerDNSResolveError(err) {
				continue
			}

			return domain, err
		case <-timeoutChannel:
			return containerAPI.CreateDomain(req, scw.WithContext(ctx))
		}
	}
}

func filterSecretEnvsToPatch(oldEnv []*container.Secret, newEnv []*container.Secret) []*container.Secret {
	toPatch := []*container.Secret{}
	// create and update - ignore hashed values
	for _, env := range newEnv {
		if env.Value != nil && strings.HasPrefix(*env.Value, "$argon2id") {
			continue
		}

		toPatch = append(toPatch, env)
	}

	// delete
	for _, env := range oldEnv {
		if !slices.ContainsFunc(newEnv, func(s *container.Secret) bool {
			return s.Key == env.Key
		}) {
			toPatch = append(toPatch, &container.Secret{Key: env.Key, Value: nil})
		}
	}

	return toPatch
}
