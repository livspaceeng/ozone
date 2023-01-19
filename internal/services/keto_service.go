package services

import (
	"bytes"
	// "context"
	"encoding/json"
	// "io"
	"net/http"
	"net/url"
	// "strings"
	// "time"

	// "github.com/gin-gonic/gin"
	"github.com/livspaceeng/ozone/configs"
	"github.com/livspaceeng/ozone/internal/model"
	"github.com/livspaceeng/ozone/internal/utils"
	// "github.com/patrickmn/go-cache"
	// client "github.com/ory/keto-client-go"
	log "github.com/sirupsen/logrus"
)

type KetoService interface {
	ValidatePolicy(hydraResponse string, namespace string, relation string, object string) (int, string, error)
}

type ketoService struct {
	httpClient *http.Client
}

func NewKetoService(httpClient *http.Client) KetoService {
	return &ketoService{
		httpClient: httpClient,
	}
}

func (ketoSvc ketoService) ValidatePolicy (hydraResponse string, namespace string, relation string, object string) (int, string, error) {
	log.Info("5")
	config := configs.GetConfig()
	httpClient := utils.NewHttpClient(ketoSvc.httpClient)
	var headers = make(map[string]string)
	var body []byte

	ketoUrl := config.GetString("keto.read.url")
	ketoPath := config.GetString("keto.read.path.check")
	// ketoRequest, _ := http.NewRequest(http.MethodGet, ketoUrl+ketoPath, nil)
	// ketoRequest.Header.Add("Accept", "application/json")
	headers["Accept"] = "application/json"
	// q := ketoRequest.URL.Query()
	u, _ := url.ParseRequestURI(ketoUrl)
	u.Path = ketoPath
	q := u.Query()
	q.Add("namespace", namespace)
	// q.Add("subject_id", hydraResponse.Subject)
	q.Add("subject_id", hydraResponse)
	q.Add("relation", relation)
	q.Add("object", object)
	// ketoRequest.URL.RawQuery = q.Encode()
	u.RawQuery = q.Encode()
	// log.Info(ketoRequest)
	// resp, err := httpClient.Do(ketoRequest)
	log.Info(u.String())
	resp, err := httpClient.SendRequest(http.MethodGet, u.String(), bytes.NewBuffer(body), headers)
	if err != nil {
		log.Error("Errored when sending request to the server", err.Error())
		return http.StatusInternalServerError, "", err
	}

	// if err != nil {
	// 	log.Error("Errored when sending request to the server", err.Error())
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// }

	defer resp.Body.Close()
	var ketoResponse model.KetoResponse
	err = json.NewDecoder(resp.Body).Decode(&ketoResponse)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
		return http.StatusInternalServerError, "", err
	}
	if !ketoResponse.Allowed {
		log.Info("Policy is not created for subject: ", hydraResponse, " Namespace: ", namespace, " Relation: ", relation, " Object: ", object)
		return http.StatusForbidden, hydraResponse, err
	}
	return http.StatusOK, hydraResponse, err
}