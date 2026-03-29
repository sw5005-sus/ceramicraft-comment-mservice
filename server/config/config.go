package config

import (
	"os"

	"github.com/spf13/viper"
)

var Config = &Conf{}

type Conf struct {
	GrpcConfig      *GrpcConfig    `mapstructure:"grpc"`
	LogConfig       *LogConfig     `mapstructure:"log"`
	HttpConfig      *HttpConfig    `mapstructure:"http"`
	MySQLConfig     *MySQL         `mapstructure:"mysql"`
	MongoConfig     *MongoDBConfig `mapstructure:"mongo"`
	RedisConfig     *RedisConfig   `mapstructure:"redis"`
	AuditGrpcConfig *GrpcConfig    `mapstructure:"audit_grpc"`
}

type RedisConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type MongoDBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
}

type HttpConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	FilePath string `mapstructure:"file_path"`
}

type GrpcConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	ConnectTimeout int    `mapstructure:"connect_timeout"`
	MaxPoolSize    int    `mapstructure:"max_pool_size"`
}

type MySQL struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	UserName string `mapstructure:"userName"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbName"`
}

var useLocalConfig = false

func Init() {
	workDir, _ := os.Getwd()
	if useLocalConfig {
		viper.SetConfigName("config-local")
	} else {
		viper.SetConfigName("config")
	}
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/resources")
	viper.AddConfigPath(workDir)

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&Config)
	if err != nil {
		panic(err)
	}
}
