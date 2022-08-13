package model

type KetoResponse struct {
	Allowed bool   `json:"allowed" example:"true" format:"bool"`
	Code    int    `json:"code,omitempty" example:"403" format:"int64"`
	Message string `json:"message,omitempty" example:"Access Forbidden"`
	Reason  string `json:"reason,omitempty" example:"Subject does not have access"`
	Request string `json:"request,omitempty" example:"xyz"`
	Status  string `json:"status,omitempty" example:"403"`
}
