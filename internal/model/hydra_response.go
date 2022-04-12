package model

type HydraResponse struct {
	Active    bool   `json:"active" example:"true" format:"bool"`
	Expiry    int    `json:"exp" example:":1648235530" format:"int64"`
	Scope     string `json:"scope" example:"offline"`
	ClientId  string `json:"client_id" example:"client-123"`
	Subject   string `json:"sub" example:"user-123"`
	TokenType string `json:"token_type" example:"access_token"`
}
