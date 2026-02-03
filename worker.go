package main

import (
	"ewallet-topup/external"
	"ewallet-topup/helpers"
	"ewallet-topup/internal/repository"
	"ewallet-topup/internal/services"

	"ewallet-topup/internal/workflow"
	"ewallet-topup/internal/workflow/transaction"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	helpers.SetupConfig()
	helpers.SetupLogger()

	c, err := client.Dial(client.Options{
		HostPort: helpers.GetEnv("TEMPORAL_HOST", "127.0.0.1:7233"),
	})
	if err != nil {
		log.Fatal("unable to connect to temporal client", err)
	}
	defer c.Close()

	db, err := helpers.SetupMySQL()
	if err != nil {
		log.Fatal("failed to connect database", err)
	}
	Ext, errExt := external.NewExternal(
		helpers.GetEnv("NOTIFICATION_HOST", ""),
	)
	if errExt != nil {
		log.Fatal("failed to connect to external service", errExt)
	}
	trxRepo := &repository.TransactionRepo{
		DB: db,
	}
	trxSvc := services.NewTransactionService(trxRepo, Ext)

	activities := &transaction.TransactionActivities{
		Service:  trxSvc,
		External: Ext,
	}

	w := worker.New(
		c,
		workflow.TransactionTaskQueue,
		worker.Options{
			MaxConcurrentActivityExecutionSize:     50,
			MaxConcurrentWorkflowTaskExecutionSize: 20,
		},
	)

	w.RegisterWorkflow(transaction.TransactionWorkflow)

	w.RegisterActivity(activities)

	log.Println("temporal worker started")

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("unable to start worker", err)
	}

}
