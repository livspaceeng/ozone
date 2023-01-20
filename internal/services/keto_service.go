package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/livspaceeng/ozone/configs"
	"github.com/livspaceeng/ozone/internal/model"
	"github.com/livspaceeng/ozone/internal/utils"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type KetoService interface {
	ValidatePolicy(ctx context.Context, hydraResponse string, namespace string, relation string, object string) (int, string, error)
	ValidatePolicyWithSet (ctx context.Context, namespace string, relation string, object string, subjectSetNamespace string, subjectSetRelation string, subjectSetObject string) (int, string, error)
	ExpandPolicy (ctx context.Context, namespace string, relation string, object string) (int, map[string]interface{}, error)
}

type ketoService struct {
	httpClient *http.Client
}

func NewKetoService(httpClient *http.Client) KetoService {
	return &ketoService{
		httpClient: httpClient,
	}
}

func (ketoSvc ketoService) ValidatePolicy (ctx context.Context, namespace string, relation string, object string, hydraResponse string) (int, string, error) {
	name := "CallKetoToValidatePolicy"
	childCtx, span := otel.Tracer(name).Start(ctx, "CallKetoToValidatePolicy")
	defer span.End()

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
	q.Add("relation", relation)
	q.Add("object", object)
	q.Add("subject_id", hydraResponse)
	u.RawQuery = q.Encode()
	log.Info(u.String())

	resp, err := httpClient.SendRequest(childCtx, http.MethodGet, u.String(), bytes.NewBuffer(body), headers)
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

func (ketoSvc ketoService) ValidatePolicyWithSet (ctx context.Context, namespace string, relation string, object string, subjectSetNamespace string, subjectSetRelation string, subjectSetObject string) (int, string, error) {
	name := "CallKetoToValidatePolicyWithSet"
	childCtx, span := otel.Tracer(name).Start(ctx, "CallKetoToValidatePolicyWithSet")
	defer span.End()

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

	resp, err := httpClient.SendRequest(childCtx, http.MethodGet, u.String(), bytes.NewBuffer(body), headers)
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

func (ketoSvc ketoService) ExpandPolicy (ctx context.Context, namespace string, relation string, object string) (int, map[string]interface{}, error) {
	name := "CallKetoToExpandPolicy"
	childCtx, span := otel.Tracer(name).Start(ctx, "CallKetoToExpandPolicy")
	defer span.End()

	config := configs.GetConfig()
	httpClient := utils.NewHttpClient(ketoSvc.httpClient)
	var headers = make(map[string]string)
	var ketoResponse map[string]interface{}
	var body []byte

	if namespace=="" || relation=="" || object=="" {
		log.Error("Invalid query params")
		return http.StatusBadRequest, ketoResponse, errors.New("Invalid query params")
	}

	ketoUrl := config.GetString("keto.read.url")
	ketoPath := config.GetString("keto.read.path.expand")
	headers["Accept"] = "application/json"

	u, _ := url.ParseRequestURI(ketoUrl)
	u.Path = ketoPath
	q := u.Query()
	q.Add("namespace", namespace)
	q.Add("relation", relation)
	q.Add("object", object)
	u.RawQuery = q.Encode()
	log.Info(u.String())

	resp, err := httpClient.SendRequest(childCtx, http.MethodGet, u.String(), bytes.NewBuffer(body), headers)
	if err != nil {
		log.Error("Errored when sending request to the server", err.Error())
		return http.StatusFailedDependency, ketoResponse, err
	}

	defer resp.Body.Close()
	encodedBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
		return http.StatusInternalServerError, ketoResponse, err
	}
	
	json.Unmarshal([]byte(string(encodedBody)), &ketoResponse)
	_, errBody := ketoResponse["error"]
	if errBody {
		log.Error("Encountered error: ", ketoResponse["error"])
		return http.StatusBadRequest, ketoResponse, err
	}
	log.Info("Response body : ", ketoResponse)
	return http.StatusOK, ketoResponse, err
}