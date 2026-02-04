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
	"strings"

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
	authHeader := c.GetHeader("Authorization")
	const prefix = "Bearer "
	if strings.HasPrefix(authHeader, prefix) {
		req.Token = authHeader[len(prefix):]
	} else {
		req.Token = authHeader
	}

	token, ok := c.Get("token")
	if !ok {
		log.Error("failed to get token data")
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	req.UserID = token.(models.TokenData).UserID

	//req.Token = c.GetHeader("Authorization")

	req.Referance = helpers.GenerateReference()
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

	var req models.UpdateTransactionStatus
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	var transactionSignalMap = map[string]string{
		"CONFIRM": transaction.SignalTransactionConfirm,
		"CANCEL":  transaction.SignalTransactionCancel,
	}

	signal, ok := transactionSignalMap[req.Status]
	if !ok {
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}
	payload := transaction.SignalTransaction{
		Reason: req.Reason,
	}
	err := api.Temporal.SignalWorkflow(
		c.Request.Context(),
		"trx_"+ref,
		"",
		signal,
		payload,
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
