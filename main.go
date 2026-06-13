package main

import (
	"fmt";
	"github.com/HeythisisSud/webhook-engine/server/internal/config";
	"context";
	"github.com/jackc/pgx/v5/pgxpool"

)

godotenv.Load()
cfg:=config.Load()


pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)




