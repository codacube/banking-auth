package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/codacube/banking-lib/logger"
	"github.com/codacube/go-microservice-banking/internal/auth/dto"
	"github.com/codacube/go-microservice-banking/internal/auth/service"
)

type AuthHandler struct {
	service service.AuthService
}

type AuthResponse struct {
	IsAuthorised bool   `json:"isAuthorised"`
	Message      string `json:"message,omitempty"`
}

func (h AuthHandler) NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, "Handle not implemented...")
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		log.Println("Error while decoding login request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, appErr := h.service.Login(loginRequest)
	if appErr != nil {
		writeResponse(w, appErr.Code, appErr.AsMessage())
		return
	}

	writeResponse(w, http.StatusOK, *token)
}

/*
	Sample URL string

http://localhost:8181/auth/verify?token=somevalidtokenstring&routeName=GetCustomer&customer_id=2000&account_id=95470
*/
func (h AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	urlParams := make(map[string]string)

	// Covert from Query to Map
	for k := range r.URL.Query() {
		urlParams[k] = r.URL.Query().Get(k)
	}

	if urlParams["token"] == "" {
		writeResponse(w, http.StatusForbidden, notAuthorisedResponse("missing token"))
		return
	}
	appErr := h.service.Verify(urlParams)
	if appErr != nil {
		writeResponse(w, appErr.Code, notAuthorisedResponse(appErr.Message))
		return
	}

	writeResponse(w, http.StatusOK, authorisedResponse())
}

func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshRequest dto.RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&refreshRequest); err != nil {
		logger.Error("Error while decoding refresh token request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, appErr := h.service.Refresh(refreshRequest)
	if appErr != nil {
		writeResponse(w, appErr.Code, appErr.AsMessage())
		return
	}

	writeResponse(w, http.StatusOK, *token)
}

func notAuthorisedResponse(msg string) AuthResponse {
	return AuthResponse{
		IsAuthorised: false,
		Message:      msg,
	}
}

func authorisedResponse() AuthResponse {
	return AuthResponse{
		IsAuthorised: true,
	}
}

func writeResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic(err)
	}
}
