package main

import (
	"fmt"
	"task_mng/cmd/web/config"
	"task_mng/interfaces/http/server"
	"task_mng/pkg/jwt"
	"task_mng/pkg/postgres"
	"task_mng/pkg/redis"

	userR "task_mng/domain/user"
	userE "task_mng/domain/user/entity"
	userS "task_mng/services/user"
)

func main() {
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err.Error())
		return
	}

	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Printf("Failed to validate config: %v\n", err.Error())
		return
	}

	jwtManager := initializeJWTManager()
	if jwtManager == nil {
		fmt.Println("Failed to initialize JWT manager, exiting...")
		return
	}

	postgres := initializePostgres()
	if postgres == nil {
		fmt.Println("Failed to initialize Postgres, exiting...")
		return
	}

	redis := initializeRedis()
	if redis == nil {
		fmt.Println("Failed to initialize Redis, exiting...")
		return
	}

	migrateDatabase(postgres)

	srv := server.New(&cfg, jwtManager, postgres, redis)

	if err := srv.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
}

// Initialize JWT manager
func initializeJWTManager() *jwt.Manager {
	config, err := jwt.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("Failed to load JWT config: %v\n", err.Error())
		return nil
	}

	if err := jwt.ValidateConfig(config); err != nil {
		fmt.Printf("Failed to validate JWT config: %v\n", err.Error())
		return nil
	}

	fmt.Println("JWT manager initialized")

	return jwt.NewManager(config)
}

// Initialize Postgres
func initializePostgres() *postgres.Database {
	config, err := postgres.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("Failed to load Postgres config: %v\n", err.Error())
		return nil
	}

	if err := postgres.ValidateConfig(config); err != nil {
		fmt.Printf("Failed to validate Postgres config: %v\n", err.Error())
		return nil
	}

	db, err := postgres.New(config)
	if err != nil {
		fmt.Printf("Failed to initialize Postgres: %v\n", err.Error())
		return nil
	}

	fmt.Println("Postgres initialized")

	return db
}

// Initialize Redis
func initializeRedis() *redis.Redis {
	config, err := redis.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("Failed to load Redis config: %v\n", err.Error())
		return nil
	}

	if err := redis.ValidateConfig(config); err != nil {
		fmt.Printf("Failed to validate Redis config: %v\n", err.Error())
		return nil
	}

	redisClient, err := redis.New(config)
	if err != nil {
		fmt.Printf("Failed to initialize Redis: %v\n", err.Error())
		return nil
	}

	fmt.Println("Redis initialized")

	return redisClient
}

func migrateDatabase(postgres *postgres.Database) {
	if err := postgres.DB.AutoMigrate(&userE.User{}); err != nil {
		fmt.Printf("Failed to migrate tables: %v\n", err)
		return
	}

	// insert default user
	userRepo := userR.New(postgres)
	userService := userS.New(userRepo, nil)

	userService.Create(&userS.CreateRequest{
		Username: "admin",
		FullName: "Admin",
		Email:    "admin@xdr.com",
		Password: "Admin!123",
	})

	fmt.Println("Migrating tables completed")
}
