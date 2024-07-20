package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"main/app/config"
	"main/app/internal/handler"
	"main/app/internal/repository/cache"
	"main/app/internal/repository/database"
	elastic "main/app/internal/repository/elasticsearch"
	"main/app/internal/route"
	"main/app/internal/service"
	"main/app/pkg/kafka/consumer"
	"main/app/pkg/kafka/producer"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	ctx := context.Background()

	conf := config.Load(ctx)

	// initialize database mysql
	mysqlClient := initDB(conf)
	// initialize redis
	redisClient := initRedis(conf)
	// initialize elastic
	elasticClient := initElastic(conf)

	consumer := consumer.New(conf)
	producer := producer.New(conf)

	// repository := repository.New(mysqlClient, redisClient, elasticClient)
	db := database.New(mysqlClient)
	cache := cache.New(redisClient)
	es := elastic.New(elasticClient)
	service := service.New(db, cache, es, *consumer, *producer)
	handler := handler.New(service)
	router := route.New(e, handler)
	router.RegisterRoute()

	fmt.Println("Run consumer in background")
	var wg sync.WaitGroup

	// Start Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Run consumer in background")
		if err := service.ConsumeMessage(ctx); err != nil {
			log.Printf("Error consuming messages: %v", err)
		}
	}()

	// Start Server
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.Logger.Fatal(e.Start(":8090")) // Start the Echo server
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Interrupt is detected")

	// Shutdown Echo server
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}

func initDB(conf *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Secrets.MysqlUser,
		conf.Secrets.MysqlPassword,
		conf.Secrets.MysqlHost,
		conf.Secrets.MysqlPort,
		conf.Secrets.MysqlDBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get sql.DB from gorm.DB")
	}
	defer sqlDB.Close()

	err = sqlDB.Ping()
	if err != nil {
		panic("failed to ping database")
	}
	fmt.Println("Connection to MySQL database successful!")

	return db
}

func initRedis(conf *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         conf.Secrets.RedisAddress,
		Password:     "",
		DB:           0,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	pong, err := rdb.Ping().Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to Redis: %v", err))
	}
	fmt.Println("Connection to Redis successful!", pong)

	return rdb
}

func initElastic(conf *config.Config) *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Addresses: []string{
			conf.Secrets.ElasticHost,
		},
		Username: "foo",
		Password: "bar",
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	elasticClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to Elastic: %v", err))
	}

	fmt.Println("Connection to Elastic successful!")

	return elasticClient
}
