package config

import "os"


type Config struct{
	DatabaseURL string
	RedisAddr string
	Port string


}

func Load() Config{
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
		Port: os.Getenv("PORT"),
	}
}