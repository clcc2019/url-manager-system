package utils

import (
	"fmt"
	"strings"

	"url-manager-system/backend/internal/db/models"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ParseYAMLToTemplateSpec 解析YAML到模板规格
func ParseYAMLToTemplateSpec(yamlContent string) (*models.TemplateSpec, error) {
	spec := &models.TemplateSpec{}

	// 按文档分割YAML
	documents := strings.Split(yamlContent, "\n---")
	var deploymentSpec map[interface{}]interface{}

	// 查找Deployment资源
	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var resource map[interface{}]interface{}
		if err := yaml.Unmarshal([]byte(doc), &resource); err != nil {
			continue
		}

		if resource["kind"] == "Deployment" {
			deploymentSpec = resource
			break
		}
	}

	if deploymentSpec == nil {
		return spec, fmt.Errorf("no Deployment resource found in YAML")
	}

	// 解析spec.template.spec.containers
	if specData, ok := deploymentSpec["spec"].(map[interface{}]interface{}); ok {
		if templateData, ok := specData["template"].(map[interface{}]interface{}); ok {
			if podSpecData, ok := templateData["spec"].(map[interface{}]interface{}); ok {
				if containersData, ok := podSpecData["containers"].([]interface{}); ok && len(containersData) > 0 {
					if containerData, ok := containersData[0].(map[interface{}]interface{}); ok {
						// 解析容器规格
						if err := parseContainerSpec(containerData, spec); err != nil {
							return spec, err
						}
					}
				}
			}
		}
	}

	return spec, nil
}

// parseContainerSpec 解析容器规格
func parseContainerSpec(containerData map[interface{}]interface{}, spec *models.TemplateSpec) error {
	// 解析镜像
	if image, ok := containerData["image"].(string); ok {
		spec.Image = image
	}

	// 解析环境变量
	if envData, ok := containerData["env"].([]interface{}); ok {
		var envVars models.EnvironmentVars
		for _, envItem := range envData {
			if envMap, ok := envItem.(map[interface{}]interface{}); ok {
				if name, ok := envMap["name"].(string); ok {
					var value string
					if val, ok := envMap["value"].(string); ok {
						value = val
					}
					envVars = append(envVars, models.EnvironmentVar{
						Name:  name,
						Value: value,
					})
				}
			}
		}
		spec.Env = envVars
	}

	// 解析命令
	if commandData, ok := containerData["command"].([]interface{}); ok {
		var commands []string
		for _, cmd := range commandData {
			if cmdStr, ok := cmd.(string); ok {
				commands = append(commands, cmdStr)
			}
		}
		spec.Command = commands
	}

	// 解析参数
	if argsData, ok := containerData["args"].([]interface{}); ok {
		var args []string
		for _, arg := range argsData {
			if argStr, ok := arg.(string); ok {
				args = append(args, argStr)
			}
		}
		spec.Args = args
	}

	// 解析端口
	if portsData, ok := containerData["ports"].([]interface{}); ok {
		var ports []models.ContainerPort
		for _, portItem := range portsData {
			if portMap, ok := portItem.(map[interface{}]interface{}); ok {
				port := models.ContainerPort{}
				if containerPort, ok := portMap["containerPort"].(int); ok {
					port.ContainerPort = int32(containerPort)
				}
				if name, ok := portMap["name"].(string); ok {
					port.Name = name
				}
				if protocol, ok := portMap["protocol"].(string); ok {
					port.Protocol = protocol
				}
				ports = append(ports, port)
			}
		}
		spec.Ports = ports
	}

	// 解析资源限制
	if resourcesData, ok := containerData["resources"].(map[interface{}]interface{}); ok {
		resources := models.ResourceLimits{}
		if requestsData, ok := resourcesData["requests"].(map[interface{}]interface{}); ok {
			if cpu, ok := requestsData["cpu"].(string); ok {
				resources.Requests.CPU = cpu
			}
			if memory, ok := requestsData["memory"].(string); ok {
				resources.Requests.Memory = memory
			}
		}
		if limitsData, ok := resourcesData["limits"].(map[interface{}]interface{}); ok {
			if cpu, ok := limitsData["cpu"].(string); ok {
				resources.Limits.CPU = cpu
			}
			if memory, ok := limitsData["memory"].(string); ok {
				resources.Limits.Memory = memory
			}
		}
		spec.Resources = resources
	}

	// 解析卷挂载
	if volumeMountsData, ok := containerData["volumeMounts"].([]interface{}); ok {
		var volumeMounts []models.VolumeMount
		for _, vmItem := range volumeMountsData {
			if vmMap, ok := vmItem.(map[interface{}]interface{}); ok {
				vm := models.VolumeMount{}
				if name, ok := vmMap["name"].(string); ok {
					vm.Name = name
				}
				if mountPath, ok := vmMap["mountPath"].(string); ok {
					vm.MountPath = mountPath
				}
				if subPath, ok := vmMap["subPath"].(string); ok {
					vm.SubPath = subPath
				}
				if readOnly, ok := vmMap["readOnly"].(bool); ok {
					vm.ReadOnly = readOnly
				}
				volumeMounts = append(volumeMounts, vm)
			}
		}
		spec.VolumeMounts = volumeMounts
	}

	// 解析工作目录
	if workingDir, ok := containerData["workingDir"].(string); ok {
		spec.WorkingDir = workingDir
	}

	// 解析存活探针
	if livenessProbeData, ok := containerData["livenessProbe"].(map[interface{}]interface{}); ok {
		probe, err := parseProbe(livenessProbeData)
		if err == nil {
			spec.LivenessProbe = probe
		}
	}

	// 解析就绪探针
	if readinessProbeData, ok := containerData["readinessProbe"].(map[interface{}]interface{}); ok {
		probe, err := parseProbe(readinessProbeData)
		if err == nil {
			spec.ReadinessProbe = probe
		}
	}

	// 解析安全上下文
	if securityContextData, ok := containerData["securityContext"].(map[interface{}]interface{}); ok {
		securityContext := &models.SecurityContext{}
		if runAsUser, ok := securityContextData["runAsUser"].(int); ok {
			runAsUser64 := int64(runAsUser)
			securityContext.RunAsUser = &runAsUser64
		}
		if runAsNonRoot, ok := securityContextData["runAsNonRoot"].(bool); ok {
			securityContext.RunAsNonRoot = &runAsNonRoot
		}
		if readOnlyRootFilesystem, ok := securityContextData["readOnlyRootFilesystem"].(bool); ok {
			securityContext.ReadOnlyRootFilesystem = &readOnlyRootFilesystem
		}
		if allowPrivilegeEscalation, ok := securityContextData["allowPrivilegeEscalation"].(bool); ok {
			securityContext.AllowPrivilegeEscalation = &allowPrivilegeEscalation
		}
		spec.SecurityContext = securityContext
	}

	return nil
}

// parseProbe 解析探针配置
func parseProbe(probeData map[interface{}]interface{}) (*models.Probe, error) {
	probe := &models.Probe{}

	if initialDelaySeconds, ok := probeData["initialDelaySeconds"].(int); ok {
		probe.InitialDelaySeconds = int32(initialDelaySeconds)
	}
	if periodSeconds, ok := probeData["periodSeconds"].(int); ok {
		probe.PeriodSeconds = int32(periodSeconds)
	}
	if timeoutSeconds, ok := probeData["timeoutSeconds"].(int); ok {
		probe.TimeoutSeconds = int32(timeoutSeconds)
	}
	if successThreshold, ok := probeData["successThreshold"].(int); ok {
		probe.SuccessThreshold = int32(successThreshold)
	}
	if failureThreshold, ok := probeData["failureThreshold"].(int); ok {
		probe.FailureThreshold = int32(failureThreshold)
	}

	// 解析HTTP GET探针
	if httpGetData, ok := probeData["httpGet"].(map[interface{}]interface{}); ok {
		httpGet := &models.HTTPGetAction{}
		if path, ok := httpGetData["path"].(string); ok {
			httpGet.Path = path
		}
		if port, ok := httpGetData["port"].(int); ok {
			httpGet.Port = intstr.FromInt(port)
		} else if portStr, ok := httpGetData["port"].(string); ok {
			httpGet.Port = intstr.FromString(portStr)
		}
		if host, ok := httpGetData["host"].(string); ok {
			httpGet.Host = host
		}
		if scheme, ok := httpGetData["scheme"].(string); ok {
			httpGet.Scheme = scheme
		}
		probe.HTTPGet = httpGet
	}

	// 解析TCP Socket探针
	if tcpSocketData, ok := probeData["tcpSocket"].(map[interface{}]interface{}); ok {
		tcpSocket := &models.TCPSocketAction{}
		if port, ok := tcpSocketData["port"].(int); ok {
			tcpSocket.Port = intstr.FromInt(port)
		} else if portStr, ok := tcpSocketData["port"].(string); ok {
			tcpSocket.Port = intstr.FromString(portStr)
		}
		if host, ok := tcpSocketData["host"].(string); ok {
			tcpSocket.Host = host
		}
		probe.TCPSocket = tcpSocket
	}

	// 解析Exec探针
	if execData, ok := probeData["exec"].(map[interface{}]interface{}); ok {
		if commandData, ok := execData["command"].([]interface{}); ok {
			var commands []string
			for _, cmd := range commandData {
				if cmdStr, ok := cmd.(string); ok {
					commands = append(commands, cmdStr)
				}
			}
			probe.Exec = &models.ExecAction{Command: commands}
		}
	}

	return probe, nil
}

// GenerateYAMLFromTemplateSpec 从模板规格重新生成YAML
func GenerateYAMLFromTemplateSpec(spec *models.TemplateSpec) (string, error) {
	// 创建Deployment元数据
	metadata := map[string]interface{}{
		"name": getDeploymentName(spec),
	}

	// 添加命名空间
	if spec.Namespace != "" {
		metadata["namespace"] = spec.Namespace
	}

	// 添加标签
	if len(spec.Labels) > 0 {
		metadata["labels"] = spec.Labels
	}

	// 添加注解
	if len(spec.Annotations) > 0 {
		metadata["annotations"] = spec.Annotations
	}

	// 创建Pod标签（用于selector和template）
	podLabels := map[string]interface{}{
		"app": getAppName(spec),
	}
	if len(spec.PodLabels) > 0 {
		for k, v := range spec.PodLabels {
			podLabels[k] = v
		}
	}

	// 创建Pod模板元数据
	templateMetadata := map[string]interface{}{
		"labels": podLabels,
	}
	if len(spec.PodAnnotations) > 0 {
		templateMetadata["annotations"] = spec.PodAnnotations
	}

	// 创建Pod规格
	podSpec := generatePodSpec(spec)

	// 创建Deployment规格
	deploymentSpec := map[string]interface{}{
		"replicas": getReplicas(spec),
		"selector": map[string]interface{}{
			"matchLabels": podLabels,
		},
		"template": map[string]interface{}{
			"metadata": templateMetadata,
			"spec":     podSpec,
		},
	}

	// 创建完整的Deployment结构
	deployment := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata":   metadata,
		"spec":       deploymentSpec,
	}

	// 序列化为YAML
	yamlData, err := yaml.Marshal(deployment)
	if err != nil {
		return "", fmt.Errorf("failed to marshal deployment to YAML: %w", err)
	}

	return string(yamlData), nil
}

// getDeploymentName 获取Deployment名称
func getDeploymentName(spec *models.TemplateSpec) string {
	if spec.DeploymentName != "" {
		return spec.DeploymentName
	}
	return "app-deployment"
}

// getAppName 获取应用名称
func getAppName(spec *models.TemplateSpec) string {
	if spec.DeploymentName != "" {
		return spec.DeploymentName
	}
	return "my-app"
}

// getReplicas 获取副本数
func getReplicas(spec *models.TemplateSpec) int32 {
	if spec.Replicas > 0 {
		return spec.Replicas
	}
	return 1
}

// generatePodSpec 生成Pod规格
func generatePodSpec(spec *models.TemplateSpec) map[string]interface{} {
	podSpec := map[string]interface{}{
		"containers": []map[string]interface{}{
			generateContainerSpec(spec),
		},
	}

	// 添加重启策略
	if spec.RestartPolicy != "" {
		podSpec["restartPolicy"] = spec.RestartPolicy
	}

	// 添加服务账号
	if spec.ServiceAccount != "" {
		podSpec["serviceAccountName"] = spec.ServiceAccount
	}

	// 添加节点选择器
	if len(spec.NodeSelector) > 0 {
		podSpec["nodeSelector"] = spec.NodeSelector
	}

	// 添加容忍度
	if len(spec.Tolerations) > 0 {
		tolerations := make([]map[string]interface{}, len(spec.Tolerations))
		for i, t := range spec.Tolerations {
			toleration := map[string]interface{}{}
			if t.Key != "" {
				toleration["key"] = t.Key
			}
			if t.Operator != "" {
				toleration["operator"] = t.Operator
			}
			if t.Value != "" {
				toleration["value"] = t.Value
			}
			if t.Effect != "" {
				toleration["effect"] = t.Effect
			}
			if t.TolerationSeconds != nil {
				toleration["tolerationSeconds"] = *t.TolerationSeconds
			}
			tolerations[i] = toleration
		}
		podSpec["tolerations"] = tolerations
	}

	// 添加亲和性
	if spec.Affinity != nil {
		podSpec["affinity"] = generateAffinitySpec(spec.Affinity)
	}

	// 添加主机网络
	if spec.HostNetwork {
		podSpec["hostNetwork"] = true
	}

	// 添加DNS策略
	if spec.DNSPolicy != "" {
		podSpec["dnsPolicy"] = spec.DNSPolicy
	}

	// 添加镜像拉取密钥
	if len(spec.ImagePullSecrets) > 0 {
		imagePullSecrets := make([]map[string]interface{}, len(spec.ImagePullSecrets))
		for i, secret := range spec.ImagePullSecrets {
			imagePullSecrets[i] = map[string]interface{}{
				"name": secret,
			}
		}
		podSpec["imagePullSecrets"] = imagePullSecrets
	}

	// 添加初始化容器
	if len(spec.InitContainers) > 0 {
		initContainers := make([]map[string]interface{}, len(spec.InitContainers))
		for i, container := range spec.InitContainers {
			initContainers[i] = generateInitContainerSpec(&container)
		}
		podSpec["initContainers"] = initContainers
	}

	// 添加卷
	if len(spec.Volumes) > 0 {
		volumes := make([]map[string]interface{}, len(spec.Volumes))
		for i, volume := range spec.Volumes {
			volumes[i] = generateVolumeSpec(&volume)
		}
		podSpec["volumes"] = volumes
	}

	return podSpec
}

// generateAffinitySpec 生成亲和性规格
func generateAffinitySpec(affinity *models.Affinity) map[string]interface{} {
	affinitySpec := make(map[string]interface{})

	if affinity.NodeAffinity != nil {
		affinitySpec["nodeAffinity"] = generateNodeAffinitySpec(affinity.NodeAffinity)
	}

	if affinity.PodAffinity != nil {
		affinitySpec["podAffinity"] = generatePodAffinitySpec(affinity.PodAffinity)
	}

	if affinity.PodAntiAffinity != nil {
		affinitySpec["podAntiAffinity"] = generatePodAntiAffinitySpec(affinity.PodAntiAffinity)
	}

	return affinitySpec
}

// generateNodeAffinitySpec 生成节点亲和性规格
func generateNodeAffinitySpec(nodeAffinity *models.NodeAffinity) map[string]interface{} {
	spec := make(map[string]interface{})

	if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		spec["requiredDuringSchedulingIgnoredDuringExecution"] = generateNodeSelectorSpec(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	}

	if len(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		preferred := make([]map[string]interface{}, len(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution))
		for i, term := range nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			preferred[i] = map[string]interface{}{
				"weight":     term.Weight,
				"preference": generateNodeSelectorTermSpec(&term.Preference),
			}
		}
		spec["preferredDuringSchedulingIgnoredDuringExecution"] = preferred
	}

	return spec
}

// generateNodeSelectorSpec 生成节点选择器规格
func generateNodeSelectorSpec(nodeSelector *models.NodeSelector) map[string]interface{} {
	terms := make([]map[string]interface{}, len(nodeSelector.NodeSelectorTerms))
	for i, term := range nodeSelector.NodeSelectorTerms {
		terms[i] = generateNodeSelectorTermSpec(&term)
	}
	return map[string]interface{}{
		"nodeSelectorTerms": terms,
	}
}

// generateNodeSelectorTermSpec 生成节点选择器条件规格
func generateNodeSelectorTermSpec(term *models.NodeSelectorTerm) map[string]interface{} {
	termSpec := make(map[string]interface{})

	if len(term.MatchExpressions) > 0 {
		expressions := make([]map[string]interface{}, len(term.MatchExpressions))
		for i, expr := range term.MatchExpressions {
			expressions[i] = map[string]interface{}{
				"key":      expr.Key,
				"operator": expr.Operator,
				"values":   expr.Values,
			}
		}
		termSpec["matchExpressions"] = expressions
	}

	if len(term.MatchFields) > 0 {
		fields := make([]map[string]interface{}, len(term.MatchFields))
		for i, field := range term.MatchFields {
			fields[i] = map[string]interface{}{
				"key":      field.Key,
				"operator": field.Operator,
				"values":   field.Values,
			}
		}
		termSpec["matchFields"] = fields
	}

	return termSpec
}

// generatePodAffinitySpec 生成Pod亲和性规格
func generatePodAffinitySpec(podAffinity *models.PodAffinity) map[string]interface{} {
	spec := make(map[string]interface{})

	if len(podAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		required := make([]map[string]interface{}, len(podAffinity.RequiredDuringSchedulingIgnoredDuringExecution))
		for i, term := range podAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			required[i] = generatePodAffinityTermSpec(&term)
		}
		spec["requiredDuringSchedulingIgnoredDuringExecution"] = required
	}

	if len(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		preferred := make([]map[string]interface{}, len(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution))
		for i, term := range podAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			preferred[i] = map[string]interface{}{
				"weight":          term.Weight,
				"podAffinityTerm": generatePodAffinityTermSpec(&term.PodAffinityTerm),
			}
		}
		spec["preferredDuringSchedulingIgnoredDuringExecution"] = preferred
	}

	return spec
}

// generatePodAntiAffinitySpec 生成Pod反亲和性规格
func generatePodAntiAffinitySpec(podAntiAffinity *models.PodAntiAffinity) map[string]interface{} {
	spec := make(map[string]interface{})

	if len(podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
		required := make([]map[string]interface{}, len(podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution))
		for i, term := range podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			required[i] = generatePodAffinityTermSpec(&term)
		}
		spec["requiredDuringSchedulingIgnoredDuringExecution"] = required
	}

	if len(podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		preferred := make([]map[string]interface{}, len(podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution))
		for i, term := range podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
			preferred[i] = map[string]interface{}{
				"weight":          term.Weight,
				"podAffinityTerm": generatePodAffinityTermSpec(&term.PodAffinityTerm),
			}
		}
		spec["preferredDuringSchedulingIgnoredDuringExecution"] = preferred
	}

	return spec
}

// generatePodAffinityTermSpec 生成Pod亲和性条件规格
func generatePodAffinityTermSpec(term *models.PodAffinityTerm) map[string]interface{} {
	termSpec := map[string]interface{}{
		"topologyKey": term.TopologyKey,
	}

	if term.LabelSelector != nil {
		termSpec["labelSelector"] = generateLabelSelectorSpec(term.LabelSelector)
	}

	if len(term.Namespaces) > 0 {
		termSpec["namespaces"] = term.Namespaces
	}

	return termSpec
}

// generateLabelSelectorSpec 生成标签选择器规格
func generateLabelSelectorSpec(selector *models.LabelSelector) map[string]interface{} {
	spec := make(map[string]interface{})

	if len(selector.MatchLabels) > 0 {
		spec["matchLabels"] = selector.MatchLabels
	}

	if len(selector.MatchExpressions) > 0 {
		expressions := make([]map[string]interface{}, len(selector.MatchExpressions))
		for i, expr := range selector.MatchExpressions {
			expressions[i] = map[string]interface{}{
				"key":      expr.Key,
				"operator": expr.Operator,
				"values":   expr.Values,
			}
		}
		spec["matchExpressions"] = expressions
	}

	return spec
}

// generateVolumeSpec 生成卷规格
func generateVolumeSpec(volume *models.Volume) map[string]interface{} {
	volumeSpec := map[string]interface{}{
		"name": volume.Name,
	}

	if volume.VolumeSource != nil {
		if volume.VolumeSource.EmptyDir != nil {
			emptyDir := make(map[string]interface{})
			if volume.VolumeSource.EmptyDir.Medium != "" {
				emptyDir["medium"] = volume.VolumeSource.EmptyDir.Medium
			}
			if volume.VolumeSource.EmptyDir.SizeLimit != "" {
				emptyDir["sizeLimit"] = volume.VolumeSource.EmptyDir.SizeLimit
			}
			volumeSpec["emptyDir"] = emptyDir
		}

		if volume.VolumeSource.HostPath != nil {
			hostPath := map[string]interface{}{
				"path": volume.VolumeSource.HostPath.Path,
			}
			if volume.VolumeSource.HostPath.Type != "" {
				hostPath["type"] = volume.VolumeSource.HostPath.Type
			}
			volumeSpec["hostPath"] = hostPath
		}

		if volume.VolumeSource.ConfigMap != nil {
			configMap := map[string]interface{}{
				"name": volume.VolumeSource.ConfigMap.Name,
			}
			if len(volume.VolumeSource.ConfigMap.Items) > 0 {
				items := make([]map[string]interface{}, len(volume.VolumeSource.ConfigMap.Items))
				for i, item := range volume.VolumeSource.ConfigMap.Items {
					itemSpec := map[string]interface{}{
						"key":  item.Key,
						"path": item.Path,
					}
					if item.Mode != nil {
						itemSpec["mode"] = *item.Mode
					}
					items[i] = itemSpec
				}
				configMap["items"] = items
			}
			if volume.VolumeSource.ConfigMap.DefaultMode != nil {
				configMap["defaultMode"] = *volume.VolumeSource.ConfigMap.DefaultMode
			}
			if volume.VolumeSource.ConfigMap.Optional != nil {
				configMap["optional"] = *volume.VolumeSource.ConfigMap.Optional
			}
			volumeSpec["configMap"] = configMap
		}

		if volume.VolumeSource.Secret != nil {
			secret := map[string]interface{}{
				"secretName": volume.VolumeSource.Secret.SecretName,
			}
			if len(volume.VolumeSource.Secret.Items) > 0 {
				items := make([]map[string]interface{}, len(volume.VolumeSource.Secret.Items))
				for i, item := range volume.VolumeSource.Secret.Items {
					itemSpec := map[string]interface{}{
						"key":  item.Key,
						"path": item.Path,
					}
					if item.Mode != nil {
						itemSpec["mode"] = *item.Mode
					}
					items[i] = itemSpec
				}
				secret["items"] = items
			}
			if volume.VolumeSource.Secret.DefaultMode != nil {
				secret["defaultMode"] = *volume.VolumeSource.Secret.DefaultMode
			}
			if volume.VolumeSource.Secret.Optional != nil {
				secret["optional"] = *volume.VolumeSource.Secret.Optional
			}
			volumeSpec["secret"] = secret
		}

		if volume.VolumeSource.PVC != nil {
			pvc := map[string]interface{}{
				"claimName": volume.VolumeSource.PVC.ClaimName,
			}
			if volume.VolumeSource.PVC.ReadOnly {
				pvc["readOnly"] = true
			}
			volumeSpec["persistentVolumeClaim"] = pvc
		}
	}

	return volumeSpec
}

// generateInitContainerSpec 生成初始化容器规格
func generateInitContainerSpec(container *models.Container) map[string]interface{} {
	containerSpec := map[string]interface{}{
		"name":  container.Name,
		"image": container.Image,
	}

	// 添加镜像拉取策略
	if container.ImagePullPolicy != "" {
		containerSpec["imagePullPolicy"] = container.ImagePullPolicy
	}

	// 添加命令
	if len(container.Command) > 0 {
		containerSpec["command"] = container.Command
	}

	// 添加参数
	if len(container.Args) > 0 {
		containerSpec["args"] = container.Args
	}

	// 添加环境变量
	if len(container.Env) > 0 {
		env := make([]map[string]interface{}, len(container.Env))
		for i, envVar := range container.Env {
			env[i] = map[string]interface{}{
				"name":  envVar.Name,
				"value": envVar.Value,
			}
		}
		containerSpec["env"] = env
	}

	// 添加端口
	if len(container.Ports) > 0 {
		ports := make([]map[string]interface{}, len(container.Ports))
		for i, port := range container.Ports {
			portSpec := map[string]interface{}{
				"containerPort": port.ContainerPort,
			}
			if port.Name != "" {
				portSpec["name"] = port.Name
			}
			if port.Protocol != "" {
				portSpec["protocol"] = port.Protocol
			}
			ports[i] = portSpec
		}
		containerSpec["ports"] = ports
	}

	// 添加资源限制
	if container.Resources.Requests.CPU != "" || container.Resources.Requests.Memory != "" ||
		container.Resources.Limits.CPU != "" || container.Resources.Limits.Memory != "" {
		resources := make(map[string]interface{})

		if container.Resources.Requests.CPU != "" || container.Resources.Requests.Memory != "" {
			requests := make(map[string]interface{})
			if container.Resources.Requests.CPU != "" {
				requests["cpu"] = container.Resources.Requests.CPU
			}
			if container.Resources.Requests.Memory != "" {
				requests["memory"] = container.Resources.Requests.Memory
			}
			resources["requests"] = requests
		}

		if container.Resources.Limits.CPU != "" || container.Resources.Limits.Memory != "" {
			limits := make(map[string]interface{})
			if container.Resources.Limits.CPU != "" {
				limits["cpu"] = container.Resources.Limits.CPU
			}
			if container.Resources.Limits.Memory != "" {
				limits["memory"] = container.Resources.Limits.Memory
			}
			resources["limits"] = limits
		}

		containerSpec["resources"] = resources
	}

	// 添加卷挂载
	if len(container.VolumeMounts) > 0 {
		volumeMounts := make([]map[string]interface{}, len(container.VolumeMounts))
		for i, vm := range container.VolumeMounts {
			vmSpec := map[string]interface{}{
				"name":      vm.Name,
				"mountPath": vm.MountPath,
			}
			if vm.SubPath != "" {
				vmSpec["subPath"] = vm.SubPath
			}
			if vm.ReadOnly {
				vmSpec["readOnly"] = true
			}
			volumeMounts[i] = vmSpec
		}
		containerSpec["volumeMounts"] = volumeMounts
	}

	// 添加工作目录
	if container.WorkingDir != "" {
		containerSpec["workingDir"] = container.WorkingDir
	}

	// 添加探针
	if container.LivenessProbe != nil {
		containerSpec["livenessProbe"] = generateProbeSpec(container.LivenessProbe)
	}

	if container.ReadinessProbe != nil {
		containerSpec["readinessProbe"] = generateProbeSpec(container.ReadinessProbe)
	}

	if container.StartupProbe != nil {
		containerSpec["startupProbe"] = generateProbeSpec(container.StartupProbe)
	}

	// 添加安全上下文
	if container.SecurityContext != nil {
		containerSpec["securityContext"] = generateSecurityContextSpec(container.SecurityContext)
	}

	return containerSpec
}

// generateContainerSpec 生成容器规格
func generateContainerSpec(spec *models.TemplateSpec) map[string]interface{} {
	containerName := "app"
	if spec.ContainerName != "" {
		containerName = spec.ContainerName
	}

	container := map[string]interface{}{
		"name":  containerName,
		"image": spec.Image,
	}

	// 添加镜像拉取策略
	if spec.ImagePullPolicy != "" {
		container["imagePullPolicy"] = spec.ImagePullPolicy
	}

	// 添加环境变量
	if len(spec.Env) > 0 {
		env := make([]map[string]interface{}, len(spec.Env))
		for i, envVar := range spec.Env {
			env[i] = map[string]interface{}{
				"name":  envVar.Name,
				"value": envVar.Value,
			}
		}
		container["env"] = env
	}

	// 添加命令
	if len(spec.Command) > 0 {
		container["command"] = spec.Command
	}

	// 添加参数
	if len(spec.Args) > 0 {
		container["args"] = spec.Args
	}

	// 添加端口
	if len(spec.Ports) > 0 {
		ports := make([]map[string]interface{}, len(spec.Ports))
		for i, port := range spec.Ports {
			portSpec := map[string]interface{}{
				"containerPort": port.ContainerPort,
			}
			if port.Name != "" {
				portSpec["name"] = port.Name
			}
			if port.Protocol != "" {
				portSpec["protocol"] = port.Protocol
			}
			ports[i] = portSpec
		}
		container["ports"] = ports
	}

	// 添加资源限制
	if spec.Resources.Requests.CPU != "" || spec.Resources.Requests.Memory != "" ||
		spec.Resources.Limits.CPU != "" || spec.Resources.Limits.Memory != "" {
		resources := make(map[string]interface{})

		if spec.Resources.Requests.CPU != "" || spec.Resources.Requests.Memory != "" {
			requests := make(map[string]interface{})
			if spec.Resources.Requests.CPU != "" {
				requests["cpu"] = spec.Resources.Requests.CPU
			}
			if spec.Resources.Requests.Memory != "" {
				requests["memory"] = spec.Resources.Requests.Memory
			}
			resources["requests"] = requests
		}

		if spec.Resources.Limits.CPU != "" || spec.Resources.Limits.Memory != "" {
			limits := make(map[string]interface{})
			if spec.Resources.Limits.CPU != "" {
				limits["cpu"] = spec.Resources.Limits.CPU
			}
			if spec.Resources.Limits.Memory != "" {
				limits["memory"] = spec.Resources.Limits.Memory
			}
			resources["limits"] = limits
		}

		container["resources"] = resources
	}

	// 添加工作目录
	if spec.WorkingDir != "" {
		container["workingDir"] = spec.WorkingDir
	}

	// 添加卷挂载
	if len(spec.VolumeMounts) > 0 {
		volumeMounts := make([]map[string]interface{}, len(spec.VolumeMounts))
		for i, vm := range spec.VolumeMounts {
			vmSpec := map[string]interface{}{
				"name":      vm.Name,
				"mountPath": vm.MountPath,
			}
			if vm.SubPath != "" {
				vmSpec["subPath"] = vm.SubPath
			}
			if vm.ReadOnly {
				vmSpec["readOnly"] = true
			}
			volumeMounts[i] = vmSpec
		}
		container["volumeMounts"] = volumeMounts
	}

	// 添加存活探针
	if spec.LivenessProbe != nil {
		container["livenessProbe"] = generateProbeSpec(spec.LivenessProbe)
	}

	// 添加就绪探针
	if spec.ReadinessProbe != nil {
		container["readinessProbe"] = generateProbeSpec(spec.ReadinessProbe)
	}

	// 添加启动探针
	if spec.StartupProbe != nil {
		container["startupProbe"] = generateProbeSpec(spec.StartupProbe)
	}

	// 添加安全上下文
	if spec.SecurityContext != nil {
		container["securityContext"] = generateSecurityContextSpec(spec.SecurityContext)
	}

	return container
}

// generateProbeSpec 生成探针规格
func generateProbeSpec(probe *models.Probe) map[string]interface{} {
	probeSpec := make(map[string]interface{})

	if probe.InitialDelaySeconds != 0 {
		probeSpec["initialDelaySeconds"] = probe.InitialDelaySeconds
	}
	if probe.PeriodSeconds != 0 {
		probeSpec["periodSeconds"] = probe.PeriodSeconds
	}
	if probe.TimeoutSeconds != 0 {
		probeSpec["timeoutSeconds"] = probe.TimeoutSeconds
	}
	if probe.SuccessThreshold != 0 {
		probeSpec["successThreshold"] = probe.SuccessThreshold
	}
	if probe.FailureThreshold != 0 {
		probeSpec["failureThreshold"] = probe.FailureThreshold
	}

	if probe.HTTPGet != nil {
		httpGet := make(map[string]interface{})
		if probe.HTTPGet.Path != "" {
			httpGet["path"] = probe.HTTPGet.Path
		}
		if probe.HTTPGet.Port != (intstr.IntOrString{}) {
			httpGet["port"] = probe.HTTPGet.Port
		}
		if probe.HTTPGet.Host != "" {
			httpGet["host"] = probe.HTTPGet.Host
		}
		if probe.HTTPGet.Scheme != "" {
			httpGet["scheme"] = probe.HTTPGet.Scheme
		}
		probeSpec["httpGet"] = httpGet
	}

	if probe.TCPSocket != nil {
		tcpSocket := make(map[string]interface{})
		if probe.TCPSocket.Port != (intstr.IntOrString{}) {
			tcpSocket["port"] = probe.TCPSocket.Port
		}
		if probe.TCPSocket.Host != "" {
			tcpSocket["host"] = probe.TCPSocket.Host
		}
		probeSpec["tcpSocket"] = tcpSocket
	}

	if probe.Exec != nil && len(probe.Exec.Command) > 0 {
		probeSpec["exec"] = map[string]interface{}{
			"command": probe.Exec.Command,
		}
	}

	return probeSpec
}

// generateSecurityContextSpec 生成安全上下文规格
func generateSecurityContextSpec(sc *models.SecurityContext) map[string]interface{} {
	securityContext := make(map[string]interface{})

	if sc.RunAsUser != nil {
		securityContext["runAsUser"] = *sc.RunAsUser
	}
	if sc.RunAsGroup != nil {
		securityContext["runAsGroup"] = *sc.RunAsGroup
	}
	if sc.RunAsNonRoot != nil {
		securityContext["runAsNonRoot"] = *sc.RunAsNonRoot
	}
	if sc.ReadOnlyRootFilesystem != nil {
		securityContext["readOnlyRootFilesystem"] = *sc.ReadOnlyRootFilesystem
	}
	if sc.AllowPrivilegeEscalation != nil {
		securityContext["allowPrivilegeEscalation"] = *sc.AllowPrivilegeEscalation
	}

	if sc.Capabilities != nil {
		caps := make(map[string]interface{})
		if len(sc.Capabilities.Add) > 0 {
			caps["add"] = sc.Capabilities.Add
		}
		if len(sc.Capabilities.Drop) > 0 {
			caps["drop"] = sc.Capabilities.Drop
		}
		if len(caps) > 0 {
			securityContext["capabilities"] = caps
		}
	}

	return securityContext
}

// ValidateYAML 验证YAML格式
func ValidateYAML(yamlContent string) error {
	var data interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}
	return nil
}
