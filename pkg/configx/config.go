package configx

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	Server  ServerConfig  `mapstructure:"server"`
	Log     LogConfig     `mapstructure:"log"`
	Auth    AuthConfig    `mapstructure:"auth"`
	Swagger SwaggerConfig `mapstructure:"swagger"`
	MySQL   MySQLConfig   `mapstructure:"mysql"`
	Redis   RedisConfig   `mapstructure:"redis"`
	MinIO   MinIOConfig   `mapstructure:"minio"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type LogConfig struct {
	Level             string `mapstructure:"level"`
	Format            string `mapstructure:"format"`
	DisableStacktrace bool   `mapstructure:"disable_stacktrace"`
}

type AuthConfig struct {
	Issuer             string        `mapstructure:"issuer"`
	AccessSecret       string        `mapstructure:"access_secret"`
	RefreshSecret      string        `mapstructure:"refresh_secret"`
	AccessTTL          time.Duration `mapstructure:"access_ttl"`
	RefreshTTL         time.Duration `mapstructure:"refresh_ttl"`
	RefreshTokenPrefix string        `mapstructure:"refresh_token_prefix"`
	PermissionCacheTTL time.Duration `mapstructure:"permission_cache_ttl"`
}

type SwaggerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	RoutePrefix string `mapstructure:"route_prefix"`
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
	Version     string `mapstructure:"version"`
	ServerURL   string `mapstructure:"server_url"`
}

type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	LogLevel        string        `mapstructure:"log_level"`
	SlowThreshold   time.Duration `mapstructure:"slow_threshold"`
}

type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type MinIOConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	Endpoint         string        `mapstructure:"endpoint"`
	AccessKey        string        `mapstructure:"access_key"`
	SecretKey        string        `mapstructure:"secret_key"`
	UseSSL           bool          `mapstructure:"use_ssl"`
	Bucket           string        `mapstructure:"bucket"`
	Location         string        `mapstructure:"location"`
	AutoCreateBucket bool          `mapstructure:"auto_create_bucket"`
	BaseURL          string        `mapstructure:"base_url"`
	PresignExpiry    time.Duration `mapstructure:"presign_expiry"`
}

func Load() (*Config, error) {
	v := viper.New()
	setDefaults(v)

	configFile := os.Getenv("CONFIG_FILE")
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("config")
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "mapstructure"
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "gin-rocket")
	v.SetDefault("app.env", "dev")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("server.shutdown_timeout", "10s")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.disable_stacktrace", false)

	v.SetDefault("auth.issuer", "gin-rocket")
	v.SetDefault("auth.access_secret", "please-change-access-secret")
	v.SetDefault("auth.refresh_secret", "please-change-refresh-secret")
	v.SetDefault("auth.access_ttl", "2h")
	v.SetDefault("auth.refresh_ttl", "168h")
	v.SetDefault("auth.refresh_token_prefix", "auth:refresh")
	v.SetDefault("auth.permission_cache_ttl", "5m")

	v.SetDefault("swagger.enabled", true)
	v.SetDefault("swagger.route_prefix", "/swagger")
	v.SetDefault("swagger.title", "Gin Rocket API")
	v.SetDefault("swagger.description", "Gin Rocket backend APIs")
	v.SetDefault("swagger.version", "1.0.0")
	v.SetDefault("swagger.server_url", "")

	v.SetDefault("mysql.charset", "utf8mb4")
	v.SetDefault("mysql.parse_time", true)
	v.SetDefault("mysql.loc", "Local")
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.max_open_conns", 50)
	v.SetDefault("mysql.conn_max_lifetime", "30m")
	v.SetDefault("mysql.conn_max_idle_time", "15m")
	v.SetDefault("mysql.log_level", "warn")
	v.SetDefault("mysql.slow_threshold", "200ms")

	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 20)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")

	v.SetDefault("minio.enabled", false)
	v.SetDefault("minio.endpoint", "127.0.0.1:9000")
	v.SetDefault("minio.access_key", "minioadmin")
	v.SetDefault("minio.secret_key", "minioadmin")
	v.SetDefault("minio.use_ssl", false)
	v.SetDefault("minio.bucket", "gin-rocket")
	v.SetDefault("minio.location", "ap-east-1")
	v.SetDefault("minio.auto_create_bucket", true)
	v.SetDefault("minio.base_url", "")
	v.SetDefault("minio.presign_expiry", "15m")
}

func (c *Config) Validate() error {
	switch {
	case c.Server.Port <= 0:
		return fmt.Errorf("server.port must be greater than 0")
	case c.MySQL.Host == "":
		return fmt.Errorf("mysql.host is required")
	case c.MySQL.Port <= 0:
		return fmt.Errorf("mysql.port must be greater than 0")
	case c.MySQL.User == "":
		return fmt.Errorf("mysql.user is required")
	case c.MySQL.DBName == "":
		return fmt.Errorf("mysql.db_name is required")
	case c.Redis.Addr == "":
		return fmt.Errorf("redis.addr is required")
	case c.Auth.AccessSecret == "":
		return fmt.Errorf("auth.access_secret is required")
	case c.Auth.RefreshSecret == "":
		return fmt.Errorf("auth.refresh_secret is required")
	case c.Auth.AccessTTL <= 0:
		return fmt.Errorf("auth.access_ttl must be greater than 0")
	case c.Auth.RefreshTTL <= 0:
		return fmt.Errorf("auth.refresh_ttl must be greater than 0")
	case c.Swagger.Enabled && c.Swagger.RoutePrefix == "":
		return fmt.Errorf("swagger.route_prefix is required when swagger is enabled")
	case c.MinIO.Enabled && c.MinIO.Endpoint == "":
		return fmt.Errorf("minio.endpoint is required when minio is enabled")
	case c.MinIO.Enabled && c.MinIO.AccessKey == "":
		return fmt.Errorf("minio.access_key is required when minio is enabled")
	case c.MinIO.Enabled && c.MinIO.SecretKey == "":
		return fmt.Errorf("minio.secret_key is required when minio is enabled")
	case c.MinIO.Enabled && c.MinIO.Bucket == "":
		return fmt.Errorf("minio.bucket is required when minio is enabled")
	case c.MinIO.Enabled && c.MinIO.PresignExpiry <= 0:
		return fmt.Errorf("minio.presign_expiry must be greater than 0 when minio is enabled")
	}

	return nil
}
