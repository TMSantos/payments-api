package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"net/http"
	"payments/app/models"
	"payments/utils"
)

const ERROR_PAYMENT_ALREADY_EXISTS = "Payment already exists with that ID"
const ERROR_ID_MISMATCH = "Mismatching IDs"

// CreatePayment handler to create a single payment
// Receives the payment and inserts in database
var CreatePayment = func(w http.ResponseWriter, r *http.Request) {

	//user := r.Context().Value("user") . (uint) //Grab the id of the user that send the request

	// Decode the request body into payment struct and failed if any error occur
	var payment models.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		utils.CreateErrorResponse(w, utils.ERROR_INVALID_JSON, http.StatusBadRequest)
		return
	}

	// Verify if the requested payment already exists in DB
	if err := utils.GetDB().Where("ID = ?", payment.ID).First(&payment).Error; !gorm.IsRecordNotFoundError(err) {
		utils.CreateErrorResponse(w, ERROR_PAYMENT_ALREADY_EXISTS, http.StatusBadRequest)
		return
	}

	// Creates the payment in DB
	if utils.GetDB().Create(&payment).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Location", fmt.Sprintf("/v1/payments/%s", payment.ID.String()))
	w.WriteHeader(http.StatusCreated)
}

// GetPayments handler to get all payments
// Returns all payments from database
var GetPayments = func(w http.ResponseWriter, r *http.Request) {

	var payments []models.Payment

	// Fetch all payments from DB
	if err := utils.GetDB().Set("gorm:auto_preload", true).Find(&payments).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Encode the payments to JSON
	data, err := json.Marshal(payments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create the API response
	response, err := json.Marshal(utils.Response{
		Data: data,
		Links: []utils.Link{{
			Rel:  "self",
			Href: "/v1/payments",
		}}})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write the response
	w.Header().Add("Content-Type", "application/json")

	_, err = w.Write(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// GetPayment handler to get a single payment
// Receives the payment id and returns the payment
var GetPayment = func(w http.ResponseWriter, r *http.Request) {

	// Read the ID from the mux vars
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // this should not be possible as muxer will only route requests with an ID
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Parse the UUID
	uuid, err := utils.ConvertStringToUUID(id)
	if err != nil {
		utils.CreateErrorResponse(w, utils.ERROR_REQUESTED_UUID_INVALID, http.StatusBadRequest)
		return
	}

	var payment models.Payment

	// Fetch the requested payment from the db
	if err := utils.GetDB().Set("gorm:auto_preload", true).Where("ID = ?", uuid).First(&payment).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			utils.CreateErrorResponse(w, utils.ERROR_RESOURCE_NOT_FOUND, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Encode the payment to JSON
	data, err := json.Marshal(payment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create the API response
	response, err := json.Marshal(utils.Response{
		Data: data,
		Links: []utils.Link{{
			Rel:  "self",
			Href: fmt.Sprintf("/v1/payments/%s", payment.ID.String()),
		}}})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write the response
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// UpdatePayment handler update a single payment
// Receives payment id and updates the payment in database
var UpdatePayment = func(w http.ResponseWriter, r *http.Request) {

	// Read the ID from the mux vars
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // the muxer should not assign this handler if the id is missing, so internal error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Parse the UUID
	uuid, err := utils.ConvertStringToUUID(id)
	if err != nil {
		utils.CreateErrorResponse(w, utils.ERROR_REQUESTED_UUID_INVALID, http.StatusBadRequest)
		return
	}

	// Decode the request body into payment struct and failed if any error occur
	var payment models.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		utils.CreateErrorResponse(w, utils.ERROR_INVALID_JSON, http.StatusBadRequest)
		return
	}

	// Ensure the payment being updated matches the one specified in the URL
	if payment.ID.String() != uuid.String() {
		utils.CreateErrorResponse(w, ERROR_ID_MISMATCH, http.StatusBadRequest)
		return
	}

	// Verify if the payment exists before editing/replacing it
	var oldPayment models.Payment

	if err := utils.GetDB().Set("gorm:auto_preload", true).Where("ID = ?", uuid).First(&oldPayment).Error; err != nil {
		utils.CreateErrorResponse(w, utils.ERROR_RESOURCE_NOT_FOUND, http.StatusNotFound)
		return
	}

	oldPayment = payment
	// Update the payment in DB
	if err := utils.GetDB().Save(&oldPayment).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write response
	w.Header().Add("Location", fmt.Sprintf("/v1/payments/%s", payment.ID.String()))
	w.WriteHeader(http.StatusNoContent)
}

// DeletePayment handler to delete a single payment
// Receives the payment id and deletes the payment in database
var DeletePayment = func(w http.ResponseWriter, r *http.Request) {

	// Read the ID from the mux vars
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // the muxer should not assign this handler if the id is missing, so internal error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Parse the UUID
	uuid, err := utils.ConvertStringToUUID(id)
	if err != nil {
		utils.CreateErrorResponse(w, utils.ERROR_REQUESTED_UUID_INVALID, http.StatusBadRequest)
		return
	}

	// Verify if the payment exists before attempting to delete it
	payment := models.Payment{
		ID: uuid,
	}

	if err := utils.GetDB().Where("ID = ?", uuid).First(&payment).Error; err != nil {
		utils.CreateErrorResponse(w, utils.ERROR_RESOURCE_NOT_FOUND, http.StatusNotFound)
		return
	}

	// Delete the payment
	if err := utils.GetDB().Delete(&payment).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write response
	w.WriteHeader(http.StatusNoContent)
}
