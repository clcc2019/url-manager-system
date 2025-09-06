package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"
	"url-manager-system/backend/internal/db/models"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

// ResourceManager Kubernetes资源管理器
type ResourceManager struct {
	client        *Client
	dynamicClient *dynamic.DynamicClient
	namespace     string
}

// NewResourceManager 创建资源管理器
func NewResourceManager(client *Client, namespace string) *ResourceManager {
	dynamicClient, _ := dynamic.NewForConfig(client.GetConfig())
	return &ResourceManager{
		client:        client,
		dynamicClient: dynamicClient,
		namespace:     namespace,
	}
}

// CreateDeployment 创建Deployment
func (rm *ResourceManager) CreateDeployment(ctx context.Context, url *models.EphemeralURL) error {
	if rm.client == nil {
		return fmt.Errorf("Kubernetes client not available")
	}

	deployment := rm.buildDeploymentSpec(url)
	_, err := rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

// UpdateDeployment 更新Deployment
func (rm *ResourceManager) UpdateDeployment(ctx context.Context, url *models.EphemeralURL) error {
	if rm.client == nil {
		return fmt.Errorf("Kubernetes client not available")
	}

	deploymentName := *url.K8sDeploymentName
	logrus.WithField("deployment", deploymentName).Info("Updating deployment")

	// 先获取现有的Deployment以保留ResourceVersion
	existingDeployment, err := rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		logrus.WithError(err).WithField("deployment", deploymentName).Error("Failed to get existing deployment for update")
		return fmt.Errorf("failed to get existing deployment: %w", err)
	}

	// 构建新的Deployment规格
	newDeployment := rm.buildDeploymentSpec(url)

	// 保留必要的元数据
	newDeployment.ResourceVersion = existingDeployment.ResourceVersion
	newDeployment.UID = existingDeployment.UID
	newDeployment.CreationTimestamp = existingDeployment.CreationTimestamp

	logrus.WithFields(logrus.Fields{
		"deployment":       deploymentName,
		"resource_version": existingDeployment.ResourceVersion,
		"new_image":        newDeployment.Spec.Template.Spec.Containers[0].Image,
		"new_replicas":     *newDeployment.Spec.Replicas,
	}).Info("Updating deployment with new configuration")

	// 更新Deployment
	_, err = rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Update(ctx, newDeployment, metav1.UpdateOptions{})
	if err != nil {
		logrus.WithError(err).WithField("deployment", deploymentName).Error("Failed to update deployment")
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	logrus.WithField("deployment", deploymentName).Info("Deployment updated successfully")
	return nil
}

// CreateOrUpdateDeployment 创建或更新Deployment
func (rm *ResourceManager) CreateOrUpdateDeployment(ctx context.Context, url *models.EphemeralURL) error {
	if rm.client == nil {
		return fmt.Errorf("Kubernetes client not available")
	}

	deploymentName := *url.K8sDeploymentName
	logrus.WithFields(logrus.Fields{
		"deployment": deploymentName,
		"namespace":  rm.namespace,
		"url_id":     url.ID.String(),
	}).Info("Attempting to create or update deployment")

	// 先尝试创建，如果已存在则更新
	err := rm.CreateDeployment(ctx, url)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// 已存在，尝试更新
			logrus.WithField("deployment", deploymentName).Info("Deployment already exists, updating...")
			return rm.UpdateDeployment(ctx, url)
		}
		// 其他错误
		logrus.WithError(err).WithField("deployment", deploymentName).Error("Failed to create deployment")
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	// 创建成功
	logrus.WithField("deployment", deploymentName).Info("Deployment created successfully")
	return nil
}

// deploymentNeedsUpdate 检查Deployment是否需要更新
func (rm *ResourceManager) deploymentNeedsUpdate(existing, new *appsv1.Deployment) bool {
	// 检查副本数
	if *existing.Spec.Replicas != *new.Spec.Replicas {
		return true
	}

	// 检查容器镜像
	if len(existing.Spec.Template.Spec.Containers) > 0 && len(new.Spec.Template.Spec.Containers) > 0 {
		if existing.Spec.Template.Spec.Containers[0].Image != new.Spec.Template.Spec.Containers[0].Image {
			return true
		}
	}

	// 检查环境变量数量（简单检查）
	existingEnvCount := len(existing.Spec.Template.Spec.Containers[0].Env)
	newEnvCount := len(new.Spec.Template.Spec.Containers[0].Env)
	if existingEnvCount != newEnvCount {
		return true
	}

	// 检查资源限制
	existingResources := existing.Spec.Template.Spec.Containers[0].Resources
	newResources := new.Spec.Template.Spec.Containers[0].Resources

	if !existingResources.Requests.Cpu().Equal(*newResources.Requests.Cpu()) ||
		!existingResources.Requests.Memory().Equal(*newResources.Requests.Memory()) ||
		!existingResources.Limits.Cpu().Equal(*newResources.Limits.Cpu()) ||
		!existingResources.Limits.Memory().Equal(*newResources.Limits.Memory()) {
		return true
	}

	return false
}

// buildDeploymentSpec 构建Deployment规格
func (rm *ResourceManager) buildDeploymentSpec(url *models.EphemeralURL) *appsv1.Deployment {
	// 确保deployment名称存在，如果不存在则基于URL ID生成
	deploymentName := ""
	if url.K8sDeploymentName != nil {
		deploymentName = *url.K8sDeploymentName
	} else {
		deploymentName = fmt.Sprintf("ephemeral-%s", url.ID.String()[:8])
		logrus.WithFields(logrus.Fields{
			"url_id":          url.ID.String(),
			"deployment_name": deploymentName,
		}).Warn("K8sDeploymentName was nil, generated new name")
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: rm.namespace,
			Labels: map[string]string{
				"app":                   "ephemeral-url",
				"ephemeral-url-id":      url.ID.String(),
				"ephemeral-url-project": url.ProjectID.String(),
				"managed-by":            "url-manager-system",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(url.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":              "ephemeral-url",
					"ephemeral-url-id": url.ID.String(),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                   "ephemeral-url",
						"ephemeral-url-id":      url.ID.String(),
						"ephemeral-url-project": url.ProjectID.String(),
						"managed-by":            "url-manager-system",
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: boolPtr(true),
						RunAsUser:    int64Ptr(1000),
						FSGroup:      int64Ptr(2000),
					},
					Containers: []corev1.Container{
						rm.buildContainerSpec(url),
					},
				},
			},
		},
	}
}

// CreateService 创建Service
func (rm *ResourceManager) CreateService(ctx context.Context, url *models.EphemeralURL) error {
	service := rm.buildServiceSpec(url)
	_, err := rm.client.GetClientset().CoreV1().Services(rm.namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

// UpdateService 更新Service
func (rm *ResourceManager) UpdateService(ctx context.Context, url *models.EphemeralURL) error {
	if rm.client == nil {
		return fmt.Errorf("Kubernetes client not available")
	}

	// 先获取现有的Service以保留ResourceVersion和ClusterIP
	existingService, err := rm.client.GetClientset().CoreV1().Services(rm.namespace).Get(ctx, *url.K8sServiceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get existing service: %w", err)
	}

	// 构建新的Service规格
	newService := rm.buildServiceSpec(url)

	// 保留必要的元数据和ClusterIP
	newService.ResourceVersion = existingService.ResourceVersion
	newService.UID = existingService.UID
	newService.CreationTimestamp = existingService.CreationTimestamp
	newService.Spec.ClusterIP = existingService.Spec.ClusterIP

	// 更新Service
	_, err = rm.client.GetClientset().CoreV1().Services(rm.namespace).Update(ctx, newService, metav1.UpdateOptions{})
	return err
}

// CreateOrUpdateService 创建或更新Service
func (rm *ResourceManager) CreateOrUpdateService(ctx context.Context, url *models.EphemeralURL) error {
	if rm.client == nil {
		return fmt.Errorf("Kubernetes client not available")
	}

	serviceName := *url.K8sServiceName
	logrus.WithField("service", serviceName).Info("Attempting to create or update service")

	// 先尝试创建，如果已存在则更新
	err := rm.CreateService(ctx, url)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// 已存在，尝试更新
			logrus.WithField("service", serviceName).Info("Service already exists, updating...")
			return rm.UpdateService(ctx, url)
		}
		// 其他错误
		logrus.WithError(err).WithField("service", serviceName).Error("Failed to create service")
		return fmt.Errorf("failed to create service: %w", err)
	}

	// 创建成功
	logrus.WithField("service", serviceName).Info("Service created successfully")
	return nil
}

// buildServiceSpec 构建Service规格
func (rm *ResourceManager) buildServiceSpec(url *models.EphemeralURL) *corev1.Service {
	// 确保service名称存在，如果不存在则基于URL ID生成
	serviceName := ""
	if url.K8sServiceName != nil {
		serviceName = *url.K8sServiceName
	} else {
		serviceName = fmt.Sprintf("svc-ephemeral-%s", url.ID.String()[:8])
		logrus.WithFields(logrus.Fields{
			"url_id":       url.ID.String(),
			"service_name": serviceName,
		}).Warn("K8sServiceName was nil, generated new name")
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: rm.namespace,
			Labels: map[string]string{
				"app":                   "ephemeral-url",
				"ephemeral-url-id":      url.ID.String(),
				"ephemeral-url-project": url.ProjectID.String(),
				"managed-by":            "url-manager-system",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":              "ephemeral-url",
				"ephemeral-url-id": url.ID.String(),
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// CreateSecret 创建Secret (如果有敏感环境变量)
func (rm *ResourceManager) CreateSecret(ctx context.Context, url *models.EphemeralURL) error {
	if url.K8sSecretName == nil || len(url.Env) == 0 {
		return nil
	}

	secretData := make(map[string][]byte)
	for _, env := range url.Env {
		secretData[env.Name] = []byte(env.Value)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *url.K8sSecretName,
			Namespace: rm.namespace,
			Labels: map[string]string{
				"app":                   "ephemeral-url",
				"ephemeral-url-id":      url.ID.String(),
				"ephemeral-url-project": url.ProjectID.String(),
				"managed-by":            "url-manager-system",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}

	_, err := rm.client.GetClientset().CoreV1().Secrets(rm.namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

// DeleteDeployment 删除Deployment
func (rm *ResourceManager) DeleteDeployment(ctx context.Context, name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil && errors.IsNotFound(err) {
		// 资源不存在，忽略错误
		return nil
	}
	return err
}

// DeleteService 删除Service
func (rm *ResourceManager) DeleteService(ctx context.Context, name string) error {
	err := rm.client.GetClientset().CoreV1().Services(rm.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && errors.IsNotFound(err) {
		// 资源不存在，忽略错误
		return nil
	}
	return err
}

// DeleteSecret 删除Secret
func (rm *ResourceManager) DeleteSecret(ctx context.Context, name string) error {
	err := rm.client.GetClientset().CoreV1().Secrets(rm.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && errors.IsNotFound(err) {
		// 资源不存在，忽略错误
		return nil
	}
	return err
}

// CheckDeploymentReady 检查Deployment是否就绪
func (rm *ResourceManager) CheckDeploymentReady(ctx context.Context, name string) (bool, error) {
	deployment, err := rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas, nil
}

// buildEnvVars 构建环境变量
func (rm *ResourceManager) buildEnvVars(url *models.EphemeralURL) []corev1.EnvVar {
	var envVars []corev1.EnvVar

	for _, env := range url.Env {
		if url.K8sSecretName != nil {
			// 如果有Secret，从Secret中引用
			envVars = append(envVars, corev1.EnvVar{
				Name: env.Name,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: *url.K8sSecretName,
						},
						Key: env.Name,
					},
				},
			})
		} else {
			// 直接设置值
			envVars = append(envVars, corev1.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			})
		}
	}

	return envVars
}

// 辅助函数
func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func resourceQuantity(value string) resource.Quantity {
	if value == "" {
		return resource.MustParse("0")
	}
	return resource.MustParse(value)
}

// getContainerName 获取容器名称
func (rm *ResourceManager) getContainerName(url *models.EphemeralURL) string {
	if url.ContainerConfig.ContainerName != "" {
		return url.ContainerConfig.ContainerName
	}
	return "app"
}

// buildContainerSpec 构建容器规格
func (rm *ResourceManager) buildContainerSpec(url *models.EphemeralURL) corev1.Container {
	container := corev1.Container{
		Name:            rm.getContainerName(url),
		Image:           url.Image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: 80,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: rm.buildEnvVars(url),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resourceQuantity(url.Resources.Requests.CPU),
				corev1.ResourceMemory: resourceQuantity(url.Resources.Requests.Memory),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resourceQuantity(url.Resources.Limits.CPU),
				corev1.ResourceMemory: resourceQuantity(url.Resources.Limits.Memory),
			},
		},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: boolPtr(false),
			RunAsNonRoot:             boolPtr(true),
			RunAsUser:                int64Ptr(1000),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
	}

	// 设置命令和参数
	if len(url.ContainerConfig.Command) > 0 {
		container.Command = url.ContainerConfig.Command
	}
	if len(url.ContainerConfig.Args) > 0 {
		container.Args = url.ContainerConfig.Args
	}

	// 设置工作目录
	if url.ContainerConfig.WorkingDir != "" {
		container.WorkingDir = url.ContainerConfig.WorkingDir
	}

	// 设置TTY和Stdin
	if url.ContainerConfig.TTY {
		container.TTY = true
	}
	if url.ContainerConfig.Stdin {
		container.Stdin = true
	}

	// 设置设备映射
	if len(url.ContainerConfig.Devices) > 0 {
		container.VolumeDevices = rm.buildVolumeDevices(url.ContainerConfig.Devices)
	}

	return container
}

// buildVolumeDevices 构建设备映射
func (rm *ResourceManager) buildVolumeDevices(devices models.DeviceMappings) []corev1.VolumeDevice {
	var volumeDevices []corev1.VolumeDevice

	for _, device := range devices {
		volumeDevices = append(volumeDevices, corev1.VolumeDevice{
			Name:       rm.sanitizeDeviceName(device.HostPath),
			DevicePath: device.ContainerPath,
		})
	}

	return volumeDevices
}

// sanitizeDeviceName 清理设备名称用于Kubernetes资源名
func (rm *ResourceManager) sanitizeDeviceName(hostPath string) string {
	// 将路径转换为有效的Kubernetes名称
	name := strings.ReplaceAll(hostPath, "/", "-")
	name = strings.Trim(name, "-")
	if name == "" {
		name = "device"
	}
	return name
}

// CreateResourcesFromYAML 从YAML创建Kubernetes资源
func (rm *ResourceManager) CreateResourcesFromYAML(ctx context.Context, yamlSpec string) error {
	if rm.dynamicClient == nil {
		return fmt.Errorf("dynamic client not available")
	}

	// 使用restmapper获取资源映射
	groupResources, err := restmapper.GetAPIGroupResources(rm.client.GetClientset().Discovery())
	if err != nil {
		return fmt.Errorf("failed to get API group resources: %w", err)
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	// 按文档分割YAML（支持多个资源）
	documents := strings.Split(yamlSpec, "\n---")
	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// 解析YAML到通用对象
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(doc), obj); err != nil {
			return fmt.Errorf("failed to unmarshal YAML: %w", err)
		}

		// 如果对象为空，跳过
		if obj.Object == nil {
			continue
		}

		// 获取资源映射
		gvk := obj.GroupVersionKind()
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return fmt.Errorf("failed to get REST mapping for %s: %w", gvk.String(), err)
		}

		// 获取资源接口
		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == "namespace" {
			dr = rm.dynamicClient.Resource(mapping.Resource).Namespace(rm.namespace)
		} else {
			dr = rm.dynamicClient.Resource(mapping.Resource)
		}

		// 创建资源
		_, err = dr.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create resource %s: %w", gvk.String(), err)
		}
	}

	return nil
}

// GetContainerStatus 获取容器状态
func (rm *ResourceManager) GetContainerStatus(ctx context.Context, deploymentName string) ([]*models.ContainerStatus, error) {
	if rm.client == nil {
		return nil, fmt.Errorf("Kubernetes client not available")
	}

	// 获取Deployment
	deployment, err := rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return []*models.ContainerStatus{}, nil
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// 获取Pod列表
	pods, err := rm.client.GetClientset().CoreV1().Pods(rm.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=ephemeral-url,ephemeral-url-id=%s", deployment.Labels["ephemeral-url-id"]),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var statuses []*models.ContainerStatus
	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			status := &models.ContainerStatus{
				Name:         containerStatus.Name,
				Image:        containerStatus.Image,
				Ready:        containerStatus.Ready,
				Started:      containerStatus.Started != nil && *containerStatus.Started,
				RestartCount: containerStatus.RestartCount,
				ContainerID:  containerStatus.ContainerID,
				State: models.ContainerState{
					Waiting:    convertContainerStateWaiting(containerStatus.State.Waiting),
					Running:    convertContainerStateRunning(containerStatus.State.Running),
					Terminated: convertContainerStateTerminated(containerStatus.State.Terminated),
				},
			}
			statuses = append(statuses, status)
		}
	}

	return statuses, nil
}

// GetPodEvents 获取Pod事件
func (rm *ResourceManager) GetPodEvents(ctx context.Context, deploymentName string) ([]*models.PodEvent, error) {
	if rm.client == nil {
		return nil, fmt.Errorf("Kubernetes client not available")
	}

	// 获取Deployment
	deployment, err := rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return []*models.PodEvent{}, nil
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// 获取Pod列表
	pods, err := rm.client.GetClientset().CoreV1().Pods(rm.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=ephemeral-url,ephemeral-url-id=%s", deployment.Labels["ephemeral-url-id"]),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var events []*models.PodEvent
	for _, pod := range pods.Items {
		// 获取Pod相关的事件
		podEvents, err := rm.client.GetClientset().CoreV1().Events(rm.namespace).List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name),
		})
		if err != nil {
			logrus.WithError(err).WithField("pod", pod.Name).Warn("Failed to get pod events")
			continue
		}

		for _, event := range podEvents.Items {
			podEvent := &models.PodEvent{
				Type:            event.Type,
				Reason:          event.Reason,
				Message:         event.Message,
				Count:           event.Count,
				FirstTimestamp:  event.FirstTimestamp.Time,
				LastTimestamp:   event.LastTimestamp.Time,
				SourceComponent: event.Source.Component,
			}
			events = append(events, podEvent)
		}
	}

	return events, nil
}

// GetContainerLogs 获取容器日志
func (rm *ResourceManager) GetContainerLogs(ctx context.Context, deploymentName, containerName string, lines int) ([]*models.ContainerLog, error) {
	if rm.client == nil {
		return nil, fmt.Errorf("Kubernetes client not available")
	}

	// 获取Deployment
	deployment, err := rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return []*models.ContainerLog{}, nil
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// 获取Pod列表
	pods, err := rm.client.GetClientset().CoreV1().Pods(rm.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=ephemeral-url,ephemeral-url-id=%s", deployment.Labels["ephemeral-url-id"]),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var logs []*models.ContainerLog
	for _, pod := range pods.Items {
		// 如果指定了容器名称，只获取该容器的日志
		containers := []string{}
		if containerName != "" {
			containers = []string{containerName}
		} else {
			// 获取所有容器的日志
			for _, container := range pod.Spec.Containers {
				containers = append(containers, container.Name)
			}
		}

		for _, container := range containers {
			// 获取容器日志
			logOptions := &corev1.PodLogOptions{
				Container: container,
				TailLines: int64Ptr(int64(lines)),
			}

			req := rm.client.GetClientset().CoreV1().Pods(rm.namespace).GetLogs(pod.Name, logOptions)
			logStream, err := req.Stream(ctx)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"pod":       pod.Name,
					"container": container,
				}).Warn("Failed to get container logs")
				continue
			}

			// 读取日志内容
			logBytes := make([]byte, 0, 1024*1024) // 1MB buffer
			buffer := make([]byte, 1024)
			for {
				n, err := logStream.Read(buffer)
				if n > 0 {
					logBytes = append(logBytes, buffer[:n]...)
				}
				if err != nil {
					break
				}
			}
			logStream.Close()

			// 解析日志行
			logLines := strings.Split(string(logBytes), "\n")
			for _, line := range logLines {
				if strings.TrimSpace(line) != "" {
					log := &models.ContainerLog{
						Timestamp: time.Now(), // 简化处理，实际应该解析日志时间戳
						Log:       line,
					}
					logs = append(logs, log)
				}
			}
		}
	}

	return logs, nil
}

// 辅助函数：转换容器状态
func convertContainerStateWaiting(waiting *corev1.ContainerStateWaiting) *models.ContainerStateWaiting {
	if waiting == nil {
		return nil
	}
	return &models.ContainerStateWaiting{
		Reason:  waiting.Reason,
		Message: waiting.Message,
	}
}

func convertContainerStateRunning(running *corev1.ContainerStateRunning) *models.ContainerStateRunning {
	if running == nil {
		return nil
	}
	return &models.ContainerStateRunning{
		StartedAt: running.StartedAt.Time,
	}
}

func convertContainerStateTerminated(terminated *corev1.ContainerStateTerminated) *models.ContainerStateTerminated {
	if terminated == nil {
		return nil
	}
	return &models.ContainerStateTerminated{
		ExitCode:   terminated.ExitCode,
		Signal:     terminated.Signal,
		Reason:     terminated.Reason,
		Message:    terminated.Message,
		StartedAt:  terminated.StartedAt.Time,
		FinishedAt: terminated.FinishedAt.Time,
	}
}
