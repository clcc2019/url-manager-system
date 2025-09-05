package k8s

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client Kubernetes客户端包装器
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewClient 创建新的Kubernetes客户端
func NewClient() (*Client, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// GetClientset 获取Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetConfig 获取Kubernetes配置
func (c *Client) GetConfig() *rest.Config {
	return c.config
}

// getKubeConfig 获取Kubernetes配置
func getKubeConfig() (*rest.Config, error) {
	// 首先尝试集群内配置
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// 如果不在集群内，尝试使用kubeconfig文件
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}
