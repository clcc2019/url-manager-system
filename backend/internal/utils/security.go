package utils

import (
	"crypto/subtle"
	"fmt"
	"regexp"
	"strings"
	"url-manager-system/backend/internal/db/models"
)

// ValidateImageName 验证镜像名称格式
func ValidateImageName(image string) bool {
	// 检查是否包含大写字母
	if strings.ToLower(image) != image {
		return false
	}
	// 简单的镜像名称验证正则表达式
	// 格式：[registry/]namespace/repository[:tag][@digest]
	imageRegex := regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)*(?::[a-z0-9]+(?:[._-][a-z0-9]+)*)?(?:@sha256:[a-f0-9]{64})?$`)
	return imageRegex.MatchString(image)
}

// ValidateResourceString 验证资源字符串格式（如CPU、内存）
func ValidateResourceString(resource string) bool {
	if resource == "" {
		return false
	}

	// CPU格式：数字+m（毫核心）或纯数字（核心）
	cpuRegex := regexp.MustCompile(`^(\d+(\.\d+)?|\d+m)$`)
	if cpuRegex.MatchString(resource) {
		return true
	}

	// 内存格式：数字+单位（Ki、Mi、Gi、Ti等）
	memRegex := regexp.MustCompile(`^(\d+(\.\d+)?)(Ki|Mi|Gi|Ti|Pi|Ei|k|M|G|T|P|E)?$`)
	return memRegex.MatchString(resource)
}

// SanitizeInput 清理输入字符串
func SanitizeInput(input string) string {
	// 移除前后空格
	input = strings.TrimSpace(input)

	// 移除潜在的危险字符
	dangerousChars := []string{"\x00", "\n", "\r", "\t"}
	for _, char := range dangerousChars {
		input = strings.ReplaceAll(input, char, "")
	}

	return input
}

// SecureCompare 安全字符串比较，防止时序攻击
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ValidateEnvironmentVariableName 验证环境变量名称
func ValidateEnvironmentVariableName(name string) bool {
	// 环境变量名只能包含字母、数字和下划线，且不能以数字开头
	envRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return envRegex.MatchString(name)
}

// IsValidProjectName 验证项目名称
func IsValidProjectName(name string) bool {
	if len(name) == 0 || len(name) > 100 {
		return false
	}

	// 项目名称只能包含字母、数字、连字符和下划线
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return nameRegex.MatchString(name)
}

// SanitizeKubernetesName 将字符串转换为符合Kubernetes RFC 1123规范的名称
func SanitizeKubernetesName(name string) string {
	if name == "" {
		return "unnamed"
	}

	// 移除前后空格
	name = strings.TrimSpace(name)

	// 将所有非字母数字字符替换为连字符
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	name = reg.ReplaceAllString(name, "-")

	// 转换为小写
	name = strings.ToLower(name)

	// 移除开头和结尾的连字符
	name = strings.Trim(name, "-")

	// 确保不为空
	if name == "" {
		return "unnamed"
	}

	// 确保以字母数字字符开头和结尾
	if !regexp.MustCompile(`^[a-z0-9]`).MatchString(name) {
		name = "x" + name
	}
	if !regexp.MustCompile(`[a-z0-9]$`).MatchString(name) {
		name = name + "x"
	}

	// 限制长度（Kubernetes名称限制为63个字符）
	if len(name) > 63 {
		name = name[:63]
		// 确保截断后仍以字母数字结尾
		name = strings.TrimRight(name, "-")
		if name == "" {
			return "unnamed"
		}
	}

	return name
}

// SanitizeKubernetesLabel 将字符串转换为符合Kubernetes标签规范的值
func SanitizeKubernetesLabel(value string) string {
	if value == "" {
		return ""
	}

	// 移除前后空格
	value = strings.TrimSpace(value)

	// 将所有非字母数字、连字符、下划线、点号的字符替换为连字符
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	value = reg.ReplaceAllString(value, "-")

	// 移除开头和结尾的非字母数字字符
	value = regexp.MustCompile(`^[^a-zA-Z0-9]+`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`[^a-zA-Z0-9]+$`).ReplaceAllString(value, "")

	// 确保不为空
	if value == "" {
		return "unnamed"
	}

	// 限制长度（Kubernetes标签值限制为63个字符）
	if len(value) > 63 {
		value = value[:63]
		// 确保截断后仍以字母数字结尾
		value = regexp.MustCompile(`[^a-zA-Z0-9]+$`).ReplaceAllString(value, "")
		if value == "" {
			return "unnamed"
		}
	}

	return value
}

// ValidateContainerConfig 验证容器配置
func ValidateContainerConfig(config models.ContainerConfig) error {
	// 验证容器名称
	if config.ContainerName != "" {
		if len(config.ContainerName) > 63 {
			return fmt.Errorf("容器名称不能超过63个字符")
		}
		// Kubernetes名称格式验证
		nameRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
		if !nameRegex.MatchString(config.ContainerName) {
			return fmt.Errorf("容器名称格式无效，只能包含小写字母、数字和连字符，且必须以字母数字开头和结尾")
		}
	}

	// 验证设备映射
	for i, device := range config.Devices {
		if device.HostPath == "" {
			return fmt.Errorf("设备映射 %d: 主机路径不能为空", i+1)
		}
		if device.ContainerPath == "" {
			return fmt.Errorf("设备映射 %d: 容器路径不能为空", i+1)
		}
		// 验证权限格式
		if device.Permissions != "" {
			permRegex := regexp.MustCompile(`^[rwm]+$`)
			if !permRegex.MatchString(device.Permissions) {
				return fmt.Errorf("设备映射 %d: 权限格式无效，只能包含 r、w、m 字符", i+1)
			}
		}
		// 验证设备映射安全性
		if err := ValidateDeviceMapping(device); err != nil {
			return fmt.Errorf("设备映射 %d: %w", i+1, err)
		}
	}

	// 验证工作目录
	if config.WorkingDir != "" && !strings.HasPrefix(config.WorkingDir, "/") {
		return fmt.Errorf("工作目录必须是绝对路径")
	}

	return nil
}

// ValidateDeviceMapping 验证设备映射安全性
func ValidateDeviceMapping(device models.DeviceMapping) error {
	// 安全检查：不允许映射敏感设备
	sensitiveDevices := []string{
		"/dev/mem",
		"/dev/kmem",
		"/dev/port",
		"/dev/sda",
		"/dev/sdb",
		"/dev/hda",
		"/dev/hdb",
		"/dev/root",
		"/proc",
		"/sys",
	}

	for _, sensitive := range sensitiveDevices {
		if strings.HasPrefix(device.HostPath, sensitive) {
			return fmt.Errorf("不允许映射敏感设备或路径: %s", sensitive)
		}
	}

	// 检查路径是否为绝对路径
	if !strings.HasPrefix(device.HostPath, "/") {
		return fmt.Errorf("主机路径必须是绝对路径")
	}
	if !strings.HasPrefix(device.ContainerPath, "/") {
		return fmt.Errorf("容器路径必须是绝对路径")
	}

	return nil
}
