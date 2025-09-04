package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"url-manager-system/backend/internal/db/models"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// IngressManager Ingress管理器
type IngressManager struct {
	client       *Client
	namespace    string
	ingressClass string
	domain       string
}

// NewIngressManager 创建Ingress管理器
func NewIngressManager(client *Client, namespace, ingressClass, domain string) *IngressManager {
	return &IngressManager{
		client:       client,
		namespace:    namespace,
		ingressClass: ingressClass,
		domain:       domain,
	}
}

// AddPath 向项目的Ingress添加路径
func (im *IngressManager) AddPath(ctx context.Context, url *models.EphemeralURL, projectName string) error {
	ingressName := fmt.Sprintf("project-%s-ingress", projectName)

	// 首先尝试获取现有的Ingress
	ingress, err := im.client.GetClientset().NetworkingV1().Ingresses(im.namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		// 如果Ingress不存在，创建新的
		if errors.IsNotFound(err) {
			return im.createProjectIngress(ctx, ingressName, url, projectName)
		}
		return err
	}

	// 如果Ingress存在，添加新路径
	return im.addPathToExistingIngress(ctx, ingress, url)
}

// RemovePath 从项目的Ingress中移除路径
func (im *IngressManager) RemovePath(ctx context.Context, projectName, path string) error {
	ingressName := fmt.Sprintf("project-%s-ingress", projectName)

	ingress, err := im.client.GetClientset().NetworkingV1().Ingresses(im.namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	return im.removePathFromIngress(ctx, ingress, path)
}

// createProjectIngress 创建项目的Ingress
func (im *IngressManager) createProjectIngress(ctx context.Context, ingressName string, url *models.EphemeralURL, projectName string) error {
	pathType := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: im.namespace,
			Labels: map[string]string{
				"app":        "url-manager-system",
				"project":    projectName,
				"managed-by": "url-manager-system",
			},
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                    im.ingressClass,
				"nginx.ingress.kubernetes.io/rewrite-target":     "/",
				"nginx.ingress.kubernetes.io/ssl-redirect":       "false",
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "false",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: im.domain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     url.Path,
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: *url.K8sServiceName,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := im.client.GetClientset().NetworkingV1().Ingresses(im.namespace).Create(ctx, ingress, metav1.CreateOptions{})
	return err
}

// addPathToExistingIngress 向现有Ingress添加路径
func (im *IngressManager) addPathToExistingIngress(ctx context.Context, ingress *networkingv1.Ingress, url *models.EphemeralURL) error {
	pathType := networkingv1.PathTypePrefix

	newPath := networkingv1.HTTPIngressPath{
		Path:     url.Path,
		PathType: &pathType,
		Backend: networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: *url.K8sServiceName,
				Port: networkingv1.ServiceBackendPort{
					Number: 80,
				},
			},
		},
	}

	// 构建JSON Patch
	patch := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/rules/0/http/paths/-",
			"value": newPath,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	_, err = im.client.GetClientset().NetworkingV1().Ingresses(im.namespace).Patch(
		ctx,
		ingress.Name,
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	)

	return err
}

// removePathFromIngress 从Ingress中移除路径
func (im *IngressManager) removePathFromIngress(ctx context.Context, ingress *networkingv1.Ingress, path string) error {
	if len(ingress.Spec.Rules) == 0 || ingress.Spec.Rules[0].HTTP == nil {
		return nil
	}

	paths := ingress.Spec.Rules[0].HTTP.Paths
	pathIndex := -1

	// 找到要删除的路径索引
	for i, p := range paths {
		if p.Path == path {
			pathIndex = i
			break
		}
	}

	if pathIndex == -1 {
		return nil // 路径不存在
	}

	// 构建JSON Patch来删除路径
	patch := []map[string]interface{}{
		{
			"op":   "remove",
			"path": fmt.Sprintf("/spec/rules/0/http/paths/%d", pathIndex),
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	_, err = im.client.GetClientset().NetworkingV1().Ingresses(im.namespace).Patch(
		ctx,
		ingress.Name,
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	)

	return err
}
