package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"net/http"
	"payments/app/models"
	"payments/utils"
)

const ERROR_PAYMENT_ALREADY_EXISTS = "Payment already exists with that ID"
const ERROR_REQUESTED_UUID_INVALID = "Requested UUID is Invalid"
const ERROR_ID_MISMATCH = "Mismatching IDs"

var CreatePayment = func(w http.ResponseWriter, r *http.Request) {

	//user := r.Context().Value("user") . (uint) //Grab the id of the user that send the request

	// read the POSTed payment by decoding it from JSON
	var payment models.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		if response, err := json.Marshal(utils.Response{Errors: []string{utils.ERROR_INVALID_JSON}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	// select the requested payment from the db
	if err := utils.GetDB().Where("ID = ?", payment.ID).First(&payment).Error; !gorm.IsRecordNotFoundError(err) {
		if response, err := json.Marshal(utils.Response{Errors: []string{ERROR_PAYMENT_ALREADY_EXISTS}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	if utils.GetDB().Create(&payment).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Location", fmt.Sprintf("/v1/payments/%s", payment.ID.String()))
	w.WriteHeader(http.StatusCreated)
}

var GetPayments = func(w http.ResponseWriter, r *http.Request) {

	var payments []models.Payment

	if err := utils.GetDB().Set("gorm:auto_preload", true).Find(&payments).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(payments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

var GetPayment = func(w http.ResponseWriter, r *http.Request) {

	// read the ID from the mux vars
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // this should not be possible as muxer will only route requests with an ID
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// parse the supplied UUID
	uuid, err := uuid.FromString(id)
	if err != nil {
		// write an error response indicating the UUID was invalid
		if response, err := json.Marshal(utils.Response{Errors: []string{ERROR_REQUESTED_UUID_INVALID}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	var payment models.Payment

	// select the requested payment from the db
	if err := utils.GetDB().Set("gorm:auto_preload", true).Where("ID = ?", uuid).First(&payment).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			if response, err := json.Marshal(utils.Response{Errors: []string{utils.ERROR_RESOURCE_NOT_FOUND}}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, err = w.Write(response)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// encode the query result
	data, err := json.Marshal(payment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// build the API response
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

var UpdatePayment = func(w http.ResponseWriter, r *http.Request) {

	// grab ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // the muxer should not assign this handler if the id is missing, so internal error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// parse ID as UUID
	uuid, err := uuid.FromString(id)
	if err != nil {
		if response, err := json.Marshal(utils.Response{Errors: []string{ERROR_REQUESTED_UUID_INVALID}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	var payment models.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		if response, err := json.Marshal(utils.Response{Errors: []string{utils.ERROR_INVALID_JSON}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	// ensure the payment being updated matches the one specified in the URL
	if payment.ID.String() != uuid.String() {
		if response, err := json.Marshal(utils.Response{Errors: []string{ERROR_ID_MISMATCH}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	// check the payment exists before editing/replacing it
	var oldPayment models.Payment

	if err := utils.GetDB().Set("gorm:auto_preload", true).Where("ID = ?", uuid).First(&oldPayment).Error; err != nil {
		if response, err := json.Marshal(utils.Response{Errors: []string{utils.ERROR_RESOURCE_NOT_FOUND}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write(response)
		}
		return
	}

	oldPayment = payment

	if err := utils.GetDB().Save(&oldPayment).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write response
	w.Header().Add("Location", fmt.Sprintf("/v1/payments/%s", payment.ID.String()))
	w.WriteHeader(http.StatusNoContent)
}

var DeletePayment = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // the muxer should not assign this handler if the id is missing, so internal error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// parse ID as UUID
	uuid, err := uuid.FromString(id)
	if err != nil {
		if response, err := json.Marshal(utils.Response{Errors: []string{ERROR_REQUESTED_UUID_INVALID}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	// check the payment exists before attempting to delete it
	payment := models.Payment{
		ID: uuid,
	}

	if err := utils.GetDB().Where("ID = ?", uuid).First(&payment).Error; err != nil {
		if response, err := json.Marshal(utils.Response{Errors: []string{utils.ERROR_RESOURCE_NOT_FOUND}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}

	// delete the payment
	if err := utils.GetDB().Delete(&payment).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}