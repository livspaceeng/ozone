package model

type KetoResponse struct {
	Allowed bool   `json:"allowed" example:"true" format:"bool"`
	Code    int    `json:"code" example:"403" format:"int64"`
	Message string `json:"message" example:"Access Forbidden"`
	Reason  string `json:"reason" example:"Subject does not have access"`
	Request string `json:"request" example:"xyz"`
	Status  string `json:"status" example:"403"`
}
