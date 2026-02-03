package cmd

import (
	"ewallet-topup/external"
	"ewallet-topup/helpers"
	"ewallet-topup/internal/api"
	"ewallet-topup/internal/interfaces"
	"ewallet-topup/internal/repository"
	"ewallet-topup/internal/services"
	"log"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

func ServeHTTP(temporal client.Client) {
	d := dependencyInject(temporal)

	r := gin.Default()

	r.GET("/health", d.HealthcheckAPI.HealthcheckHandlerHTTP)

	transactionV1 := r.Group("/transaction/v1")
	transactionV1.POST("/create", d.MiddlewareValidateToken, d.TransactionAPI.CreateTransaction)
	transactionV1.PUT("/update-status/:reference", d.MiddlewareValidateToken, d.TransactionAPI.UpdateStatusTransaction)
	//transactionV1.GET("/", d.MiddlewareValidateToken, d.TransactionAPI.GetTransaction)
	//transactionV1.GET("/:reference", d.MiddlewareValidateToken, d.TransactionAPI.GetTransactionDetail)
	//transactionV1.POST("/refund", d.MiddlewareValidateToken, d.TransactionAPI.RefundTransaction)

	err := r.Run(":" + helpers.GetEnv("APP_PORT", ""))
	if err != nil {
		log.Fatal(err)
	}
}

type Dependency struct {
	HealthcheckAPI interfaces.IHealthcheckAPI
	External       interfaces.IExternal
	TransactionAPI interfaces.ITransactionAPI
}

func dependencyInject(temporal client.Client) Dependency {
	healthcheckSvc := &services.Healthcheck{}
	healthcheckAPI := &api.Healthcheck{
		HealthcheckServices: healthcheckSvc,
	}

	notifClient, err := external.NewNotificationClient(helpers.GetEnv("NOTIFICATION_HOST", ""))
	if err != nil {
		log.Fatal("failed to init notification client")
	}

	ext := &external.External{
		NotificationClient: notifClient,
	}

	trxRepo := &repository.TransactionRepo{
		DB: helpers.DB,
	}
	trxService := &services.TransactionService{
		TransactionRepo: trxRepo,
		External:        ext,
	}
	trxAPI := &api.TransactionAPI{
		TransactionService: trxService,
		Temporal:           temporal,
	}

	return Dependency{
		HealthcheckAPI: healthcheckAPI,
		External:       ext,
		TransactionAPI: trxAPI,
	}
}
