package services

import (
	// "bytes"
	// "context"
	"encoding/json"
	// "io"
	"net/http"
	"net/url"
	"strings"
	// "time"

	// "github.com/gin-gonic/gin"
	"github.com/livspaceeng/ozone/configs"
	"github.com/livspaceeng/ozone/internal/model"
	"github.com/livspaceeng/ozone/internal/utils"
	// "github.com/patrickmn/go-cache"
	// client "github.com/ory/keto-client-go"
	log "github.com/sirupsen/logrus"
)

type HydraService interface {
	GetSubjectByToken(hydraClient string, bearer string) (int, string, error)
}

type hydraService struct {
	httpClient *http.Client
}

func NewHydraService(httpClient *http.Client) HydraService {
	return &hydraService{
		httpClient: httpClient,
	}
}

func (hydraSvc hydraService) GetSubjectByToken(hydraClient string, bearer string) (int, string, error) {
	// cacheManager := cache.New(5*time.Minute, 10*time.Minute)
	log.Info("3")
	config := configs.GetConfig()
	httpClient := utils.NewHttpClient(hydraSvc.httpClient)
	var headers = make(map[string]string)
	var hydraUrl, hydraPath string

	if hydraClient == "accounts" {
		hydraUrl = config.GetString("accounts.hydra.url")
		hydraPath = config.GetString("accounts.hydra.path.introspect")
	} else {
		hydraUrl = config.GetString("bouncer.hydra.url")
		hydraPath = config.GetString("bouncer.hydra.path.introspect")
	} 
	u, _ := url.ParseRequestURI(hydraUrl)
	u.Path = hydraPath

	if len(bearer) <= 0 {
		log.Error("Bearer token absent!")
		return http.StatusUnauthorized, "", nil
	}

	validBearer := strings.Contains(bearer, "Bearer ") || strings.Contains(bearer, "bearer ")
	if !validBearer {
		log.Error("Authorization header format is not valid!")
		return http.StatusUnauthorized, "", nil
	}
	token := strings.Split(bearer, " ")[1]
	data := url.Values{}
	data.Set("token", token)
	log.Info("4")
	// hydraRequest, _ := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	// hydraRequest.Header.Add("Authorization", bearer)
	// hydraRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// resp, err := httpClient.Do(hydraRequest)

	
	headers["Authorization"] = bearer
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	resp, err := httpClient.SendRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()), headers)
	if err != nil {
		log.Error("Errored when sending request to the server", err.Error())
		return http.StatusFailedDependency, "", err
	}
	var hydraResponse model.HydraResponse
	err = json.NewDecoder(resp.Body).Decode(&hydraResponse)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
		return http.StatusInternalServerError, "", err
	}
	log.Info("Subject: ", hydraResponse.Subject)
	if hydraResponse.Subject == "" {
		log.Error("Subject is nil!")
		return http.StatusUnauthorized, hydraResponse.Subject, err
	}

	//Cache Store
	// cacheManager.Set(token, hydraResponse.Subject, cache.DefaultExpiration)

	return http.StatusOK, hydraResponse.Subject, err
}
