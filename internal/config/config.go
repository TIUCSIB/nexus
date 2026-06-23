package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App          AppConfig      `yaml:"app"`
	Server       ServerConfig   `yaml:"server"`
	Database     DatabaseConfig `yaml:"database"`
	JWT          JWTConfig      `yaml:"jwt"`
	GRPC         GRPCConfig     `yaml:"grpc"`
	Node         NodeConfig     `yaml:"node"`
	Subscription SubConfig      `yaml:"subscription"`
}

type AppConfig struct {
	Name      string `yaml:"name"`
	Debug     bool   `yaml:"debug"`
	SecretKey string `yaml:"secret_key"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type GRPCConfig struct {
	Listen   string `yaml:"listen"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type NodeConfig struct {
	HeartbeatInterval int `yaml:"heartbeat_interval"`
	OfflineTimeout    int `yaml:"offline_timeout"`
}

type SubConfig struct {
	TrafficResetDays int `yaml:"traffic_reset_days"`
	PlanSort         int `yaml:"plan_sort"`
}

var Global Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &Global)
}
