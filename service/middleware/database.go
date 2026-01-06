package middleware

import (
	"compost-bin/logger"
	"compost-bin/service/middleware/dao"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func GetDb() *gorm.DB {
	return db
}

func init() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "password")
	databaseName := getEnv("DB_NAME", "compost_bin")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		password,
		host,
		port,
		databaseName)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.NewDatabaseLogger(),
	})
	if err != nil {
		logger.Fatalf("Failed to connect to MySQL: %v", err)
	}

	err = autoMigrate()
	if err != nil {
		logger.Fatalf("Auto migration failed: %v", err)
	}
}

func autoMigrate() error {
	models := make(map[string]any, 1)
	models["user"] = &dao.User{}

	for n, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("failed to migrate %s", n)
		}
	}
	return nil
}

func getEnv(variableName, defaultValue string) string {
	if value := os.Getenv(variableName); value != "" {
		return value
	}
	return defaultValue
}

var DatabaseFailure = fmt.Errorf("数据库错误o（T_T）o")
