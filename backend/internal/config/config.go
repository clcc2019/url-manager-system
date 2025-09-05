package config

import (
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Debug       bool           `mapstructure:"debug"`
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Redis       RedisConfig    `mapstructure:"redis"`
	K8s         K8sConfig      `mapstructure:"k8s"`
	Security    SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type K8sConfig struct {
	Namespace     string `mapstructure:"namespace"`
	InCluster     bool   `mapstructure:"in_cluster"`
	ConfigPath    string `mapstructure:"config_path"`
	DefaultDomain string `mapstructure:"default_domain"`
	IngressClass  string `mapstructure:"ingress_class"`
}

type SecurityConfig struct {
	AllowedImages   []string `mapstructure:"allowed_images"`
	MaxReplicas     int      `mapstructure:"max_replicas"`
	MaxTTLSeconds   int      `mapstructure:"max_ttl_seconds"`
	DefaultCPULimit string   `mapstructure:"default_cpu_limit"`
	DefaultMemLimit string   `mapstructure:"default_mem_limit"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/url-manager")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// 覆盖环境变量配置
	overrideWithEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	// Server配置
	viper.SetDefault("debug", false)
	viper.SetDefault("environment", "development")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")

	// Database配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.database", "url_manager")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", time.Hour)

	// Redis配置
	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// K8s配置
	viper.SetDefault("k8s.namespace", "default")
	viper.SetDefault("k8s.in_cluster", false)
	viper.SetDefault("k8s.config_path", "")
	viper.SetDefault("k8s.default_domain", "example.com")
	viper.SetDefault("k8s.ingress_class", "nginx")

	// Security配置
	viper.SetDefault("security.allowed_images", []string{"nginx:latest", "httpd:latest"})
	viper.SetDefault("security.max_replicas", 3)
	viper.SetDefault("security.max_ttl_seconds", 86400*7) // 7天
	viper.SetDefault("security.default_cpu_limit", "500m")
	viper.SetDefault("security.default_mem_limit", "512Mi")
}

func overrideWithEnv() {
	if val := os.Getenv("DEBUG"); val != "" {
		if debug, err := strconv.ParseBool(val); err == nil {
			viper.Set("debug", debug)
		}
	}

	if val := os.Getenv("SERVER_PORT"); val != "" {
		viper.Set("server.port", val)
	}

	if val := os.Getenv("DATABASE_URL"); val != "" {
		viper.Set("database.url", val)
	}

	if val := os.Getenv("DATABASE_HOST"); val != "" {
		viper.Set("database.host", val)
	}

	if val := os.Getenv("DATABASE_PORT"); val != "" {
		viper.Set("database.port", val)
	}

	if val := os.Getenv("DATABASE_USER"); val != "" {
		viper.Set("database.user", val)
	}

	if val := os.Getenv("DATABASE_PASSWORD"); val != "" {
		viper.Set("database.password", val)
	}

	if val := os.Getenv("DATABASE_NAME"); val != "" {
		viper.Set("database.database", val)
	}

	if val := os.Getenv("REDIS_ADDRESS"); val != "" {
		viper.Set("redis.address", val)
	}

	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		viper.Set("redis.password", val)
	}

	if val := os.Getenv("K8S_NAMESPACE"); val != "" {
		viper.Set("k8s.namespace", val)
	}

	if val := os.Getenv("K8S_IN_CLUSTER"); val != "" {
		if inCluster, err := strconv.ParseBool(val); err == nil {
			viper.Set("k8s.in_cluster", inCluster)
		}
	}

	if val := os.Getenv("K8S_CONFIG_PATH"); val != "" {
		viper.Set("k8s.config_path", val)
	}

	if val := os.Getenv("DEFAULT_DOMAIN"); val != "" {
		viper.Set("k8s.default_domain", val)
	}
}
