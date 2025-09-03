package main

import (
	"strings"

	"github.com/go-portfolio/order-pipeline/internal/config"
	"github.com/go-portfolio/order-pipeline/internal/server"
)

func main() {
	appCfg := config.LoadConfig()

	brokers := strings.Split(appCfg.KafkaBrokers, ",")
	workerServer := server.NewWorkerServer(
		brokers,
		appCfg.KafkaTopic,
		appCfg.DlqTopic,
		appCfg.WorkerGroup,
		appCfg.RedisAddr,
	)

	workerServer.Run()
}
