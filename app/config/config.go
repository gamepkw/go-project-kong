package config

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type Config struct {
	Server         Server   `mapstructure:"server"`
	AllowOrigins   string   `mapstructure:"allow_origins"`
	WebportalMysql Database `mapstructure:"webportal_mysql"`
	PlatformMysql  Database `mapstructure:"platform_mysql"`
	Secrets        Secrets  `mapstructure:"secrets"`
}

type Server struct {
	Port int `mapstructure:"port"`
}

type Database struct {
	DBName string `mapstructure:"dbname"`
}

type Secrets struct {
	MysqlHost          string `envconfig:"SECRET_MYSQL_HOST"`
	MysqlPort          int    `envconfig:"SECRET_MYSQL_PORT"`
	MysqlDBName        string `envconfig:"SECRET_MYSQL_DB_NAME"`
	MysqlUser          string `envconfig:"SECRET_MYSQL_USER"`
	MysqlPassword      string `envconfig:"SECRET_MYSQL_PASSWORD"`
	RedisAddress       string `envconfig:"SECRET_REDIS_ADDRESS"`
	ElasticHost        string `envconfig:"SECRET_ELASTIC_HOST"`
	ElasticPort        string `envconfig:"SECRET_ELASTIC_PORT"`
	KafkaBroker        string `envconfig:"SECRET_KAFKA_BROKER"`
	KafkaConsumerGroup string `envconfig:"SECRET_KAFKA_CONSUMER_GROUP"`
	KafkaVersion       string `envconfig:"SECRET_KAFKA_VERSION"`
	KafkaTopic         string `envconfig:"SECRET_KAFKA_TOPIC"`
}

var configOnce sync.Once
var conf *Config

func Load(ctx context.Context) *Config {
	configOnce.Do(func() {
		log.SetOutput(os.Stdout)
		configPath, ok := os.LookupEnv("API_CONFIG_PATH")
		if !ok {
			log.Println(ctx, "API_CONFIG_PATH not found, using default config")
			configPath = "./config"
		}
		configName, ok := os.LookupEnv("API_CONFIG_NAME")
		if !ok {
			log.Println(ctx, "API_CONFIG_NAME not found, using default config")
			configName = "config"
		}
		viper.AddConfigPath(configPath)
		viper.SetConfigName(configName)
		viper.SetConfigType("yaml")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		err := viper.ReadInConfig()
		if err != nil {
			log.Println(ctx, "config file not found. using default/env config: "+err.Error())
		}
		viper.AutomaticEnv()
		if err := viper.MergeConfig(strings.NewReader(viper.GetString("configs"))); err != nil {
			log.Panic(err.Error())
		}
		if err := viper.Unmarshal(&conf); err != nil {
			log.Panic(err, "unmarshal config to struct error")
		}
		for _, value := range os.Environ() {
			pair := strings.SplitN(value, "=", 2)
			if strings.Contains(pair[0], "SECRET_") {
				keys := strings.Replace(pair[0], "SECRET_", "secrets.", -1)
				keys = strings.Replace(keys, "_", ".", -1)
				newKey := strings.Trim(keys, " ")
				newValue := strings.Trim(pair[1], " ")
				viper.Set(newKey, newValue)
			}
		}
		if err := godotenv.Load("./config/secret.env"); err != nil {
			log.Println("can't load ./config/secret.env", err)
		}
		if err := envconfig.Process("SECRET", &conf.Secrets); err != nil {
			log.Println("can't process SECRET", err)
		}
	})
	return conf
}
