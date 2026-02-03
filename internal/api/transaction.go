package api

import (
	"context"
	constants "ewallet-topup/constant"
	"ewallet-topup/helpers"
	"ewallet-topup/internal/interfaces"
	"ewallet-topup/internal/models"
	"ewallet-topup/internal/workflow"
	"ewallet-topup/internal/workflow/transaction"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

type TransactionAPI struct {
	TransactionService interfaces.ITransactionService
	Temporal           client.Client
}

func (api *TransactionAPI) CreateTransaction(c *gin.Context) {
	var (
		log = helpers.Logger
		req models.CreateTransactionRequest
	)

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to parse request: ", err)
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	token, ok := c.Get("token")
	if !ok {
		log.Error("failed to get token data")
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	tokenData := token.(models.TokenData)
	req.UserID = tokenData.UserID

	req.Token = c.GetHeader("Authorization")

	workflowOptions := client.StartWorkflowOptions{
		ID:        "trx_" + req.Referance,
		TaskQueue: workflow.TransactionTaskQueue,
	}

	we, err := api.Temporal.ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		transaction.TransactionWorkflow,
		req,
	)
	if err != nil {
		log.Error(err)
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	helpers.SendResponseHTTP(c, http.StatusAccepted, constants.SuccessMessage, gin.H{
		"reference":   req.Referance,
		"workflow_id": we.GetID(),
		"run_id":      we.GetRunID(),
		"status":      "PROCESSING",
	})
}

func (api *TransactionAPI) UpdateStatusTransaction(c *gin.Context) {
	ref := c.Param("reference")
	status := c.Query("status")

	var signal string
	switch status {
	case "CONFIRM":
		signal = transaction.SignalTransactionConfirm
	case "CANCEL":
		signal = transaction.SignalTransactionCancel
	default:
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	err := api.Temporal.SignalWorkflow(
		context.Background(),
		"trx_"+ref,
		"",
		signal,
		nil,
	)
	if err != nil {
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	helpers.SendResponseHTTP(c, http.StatusOK, constants.SuccessMessage, nil)
}

//
//func (api *TransactionAPI) GetTransaction(c *gin.Context) {
//	var (
//		log = helpers.Logger
//	)
//
//	token, ok := c.Get("token")
//	if !ok {
//		log.Error("failed to get token data")
//		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
//		return
//	}
//
//	tokenData, ok := token.(models.TokenData)
//	if !ok {
//		log.Error("failed to parse token data")
//		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
//		return
//	}
//
//	resp, err := api.TransactionService.GetTransaction(c.Request.Context(), int(tokenData.UserID))
//	if err != nil {
//		log.Error("failed to create transaction: ", err)
//		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
//		return
//	}
//
//	helpers.SendResponseHTTP(c, http.StatusOK, constants.SuccessMessage, resp)
//}
//
//func (api *TransactionAPI) GetTransactionDetail(c *gin.Context) {
//	var (
//		log = helpers.Logger
//	)
//
//	reference := c.Param("reference")
//	if reference == "" {
//		log.Error("failed to get reference")
//		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
//		return
//	}
//
//	token, ok := c.Get("token")
//	if !ok {
//		log.Error("failed to get token data")
//		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
//		return
//	}
//
//	_, ok = token.(models.TokenData)
//	if !ok {
//		log.Error("failed to parse token data")
//		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
//		return
//	}
//
//	resp, err := api.TransactionService.GetTransactionDetail(c.Request.Context(), reference)
//	if err != nil {
//		log.Error("failed to create transaction: ", err)
//		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
//		return
//	}
//
//	helpers.SendResponseHTTP(c, http.StatusOK, constants.SuccessMessage, resp)
//}
