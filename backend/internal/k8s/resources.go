package k8s

import (
	"context"
	"fmt"
	"strings"
	"url-manager-system/backend/internal/db/models"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ResourceManager Kubernetes资源管理器
type ResourceManager struct {
	client    *Client
	namespace string
}

// NewResourceManager 创建资源管理器
func NewResourceManager(client *Client, namespace string) *ResourceManager {
	return &ResourceManager{
		client:    client,
		namespace: namespace,
	}
}

// CreateDeployment 创建Deployment
func (rm *ResourceManager) CreateDeployment(ctx context.Context, url *models.EphemeralURL) error {
	if rm.client == nil {
		return fmt.Errorf("Kubernetes client not available")
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *url.K8sDeploymentName,
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
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(80),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(80),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
							},
						},
					},
				},
			},
		},
	}

	_, err := rm.client.GetClientset().AppsV1().Deployments(rm.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

// CreateService 创建Service
func (rm *ResourceManager) CreateService(ctx context.Context, url *models.EphemeralURL) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *url.K8sServiceName,
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

	_, err := rm.client.GetClientset().CoreV1().Services(rm.namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
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
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/",
					Port: intstr.FromInt(80),
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/",
					Port: intstr.FromInt(80),
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
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
