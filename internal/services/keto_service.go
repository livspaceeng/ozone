package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/livspaceeng/ozone/configs"
	"github.com/livspaceeng/ozone/internal/model"
	"github.com/livspaceeng/ozone/internal/utils"
	log "github.com/sirupsen/logrus"
)

type KetoService interface {
	ValidatePolicy(hydraResponse string, namespace string, relation string, object string) (int, string, error)
	ValidatePolicyWithSet (namespace string, relation string, object string, subjectSetNamespace string, subjectSetRelation string, subjectSetObject string) (int, string, error)
}

type ketoService struct {
	httpClient *http.Client
}

func NewKetoService(httpClient *http.Client) KetoService {
	return &ketoService{
		httpClient: httpClient,
	}
}

func (ketoSvc ketoService) ValidatePolicy (namespace string, relation string, object string, hydraResponse string) (int, string, error) {
	config := configs.GetConfig()
	httpClient := utils.NewHttpClient(ketoSvc.httpClient)
	var headers = make(map[string]string)
	var body []byte

	if namespace=="" || relation=="" || object=="" {
		log.Error("Invalid query params")
		return http.StatusBadRequest, "", errors.New("Invalid query params")
	}

	ketoUrl := config.GetString("keto.read.url")
	ketoPath := config.GetString("keto.read.path.check")
	headers["Accept"] = "application/json"

	u, _ := url.ParseRequestURI(ketoUrl)
	u.Path = ketoPath
	q := u.Query()
	q.Add("namespace", namespace)
	q.Add("subject_id", hydraResponse)
	q.Add("relation", relation)
	q.Add("object", object)
	u.RawQuery = q.Encode()
	log.Info(u.String())

	resp, err := httpClient.SendRequest(http.MethodGet, u.String(), bytes.NewBuffer(body), headers)
	if err != nil {
		log.Error("Errored when sending request to the server", err.Error())
		return http.StatusFailedDependency, "", err
	}

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

func (ketoSvc ketoService) ValidatePolicyWithSet (namespace string, relation string, object string, subjectSetNamespace string, subjectSetRelation string, subjectSetObject string) (int, string, error) {
	config := configs.GetConfig()
	httpClient := utils.NewHttpClient(ketoSvc.httpClient)
	var headers = make(map[string]string)
	var body []byte

	if namespace=="" || relation=="" || object=="" || subjectSetNamespace=="" || subjectSetRelation=="" || subjectSetObject=="" {
		log.Error("Invalid query params")
		return http.StatusBadRequest, "", errors.New("Invalid query params")
	}

	ketoUrl := config.GetString("keto.read.url")
	ketoPath := config.GetString("keto.read.path.check")
	headers["Accept"] = "application/json"

	u, _ := url.ParseRequestURI(ketoUrl)
	u.Path = ketoPath
	q := u.Query()
	q.Add("namespace", namespace)
	q.Add("relation", relation)
	q.Add("object", object)
	q.Add("subject_set.namespace", subjectSetNamespace)
	q.Add("subject_set.relation", subjectSetRelation)
	q.Add("subject_set.object", subjectSetObject)
	u.RawQuery = q.Encode()
	log.Info(u.String())

	resp, err := httpClient.SendRequest(http.MethodGet, u.String(), bytes.NewBuffer(body), headers)
	if err != nil {
		log.Error("Errored when sending request to the server", err.Error())
		return http.StatusFailedDependency, "", err
	}

	defer resp.Body.Close()
	var ketoResponse model.KetoResponse
	err = json.NewDecoder(resp.Body).Decode(&ketoResponse)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
		return http.StatusInternalServerError, "", err
	}
	if !ketoResponse.Allowed {
		log.Info("Policy is not created for subjectSetNamespace: ", subjectSetNamespace, " subjectSetRelation: ", subjectSetRelation, " subjectSetObject: ", subjectSetObject, " Namespace: ", namespace, " Relation: ", relation, " Object: ", object)
		return http.StatusForbidden, "Policy does not exist", err
	}
	return http.StatusOK, "Policy exists", err
}