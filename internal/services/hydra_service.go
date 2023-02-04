package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/livspaceeng/ozone/configs"
	"github.com/livspaceeng/ozone/internal/model"
	"github.com/livspaceeng/ozone/internal/utils"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type HydraService interface {
	GetSubjectByToken(ctx context.Context, hydraClient string, bearer string) (int, string, error)
}

type hydraService struct {
	httpClient *http.Client
	cacheClient *cache.Cache
}

func NewHydraService(httpClient *http.Client, cacheClient *cache.Cache) HydraService {
	return &hydraService{
		httpClient: httpClient,
		cacheClient: cacheClient,
	}
}

func (hydraSvc hydraService) GetSubjectByToken(ctx context.Context, hydraClient string, bearer string) (int, string, error) {
	name := "CallHydraToFetchSubject"
	childCtx, span := otel.Tracer(name).Start(ctx, "CallHydraToFetchSubject")
	defer span.End()
	
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
		log.Error("Bearer token absent")
		return http.StatusUnauthorized, "", errors.New("Bearer token absent")
	}

	validBearer := strings.Contains(bearer, "Bearer ") || strings.Contains(bearer, "bearer ")
	if !validBearer {
		log.Error("Authorization header format is not valid - ", bearer)
		return http.StatusUnauthorized, "", errors.New("Authorization header format is not valid")
	}
	token := strings.Split(bearer, " ")[1]
	
	//Cache Read
	subject, found := hydraSvc.cacheClient.Get(token)
	if found || subject != nil{
		log.Info("Subject found in cache")
		return http.StatusOK, subject.(string), nil
	}

	data := url.Values{}
	data.Set("token", token)
	headers["Authorization"] = bearer
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	log.Info(u.String())

	resp, err := httpClient.SendRequest(childCtx, http.MethodPost, u.String(), strings.NewReader(data.Encode()), headers)
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
	tokenValidity := hydraResponse.Expiry-int(time.Now().Unix())-1
	hydraSvc.cacheClient.Set(token, hydraResponse.Subject, time.Duration(tokenValidity)*time.Second)

	return http.StatusOK, hydraResponse.Subject, err
}
