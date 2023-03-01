package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/livspaceeng/ozone/internal/utils"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type KetoService interface {
	ValidatePolicy(ctx context.Context, hydraResponse string, namespace string, relation string, object string) (int, string, error)
	ValidatePolicyWithSet (ctx context.Context, namespace string, relation string, object string, subjectSetNamespace string, subjectSetRelation string, subjectSetObject string) (int, string, error)
	ExpandPolicy (ctx context.Context, namespace string, relation string, object string, maxDepth string, hasDepth bool) (int, map[string]interface{}, error)
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

	if namespace=="" || relation=="" || object=="" || hydraResponse=="" {
		log.Error("Invalid query params")
		return http.StatusBadRequest, "", errors.New("Invalid query params")
	}

	ketoResponse, r, err := utils.GetKetoReadClient().PermissionApi.CheckPermission(childCtx).
		Namespace(namespace).
		Relation(relation).
		Object(object).
		SubjectId(hydraResponse).
		Execute()
	if err != nil {
		log.Error("Error when calling `PermissionApi.CheckPermission``:\n", err, " Http Response: ", r)
		return http.StatusFailedDependency, "", err
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

	if namespace=="" || relation=="" || object=="" || subjectSetNamespace=="" || subjectSetRelation=="" || subjectSetObject=="" {
		log.Error("Invalid query params")
		return http.StatusBadRequest, "", errors.New("Invalid query params")
	}

	ketoResponse, r, err := utils.GetKetoReadClient().PermissionApi.CheckPermission(childCtx).
		Namespace(namespace).
		Relation(relation).
		Object(object).
		SubjectSetNamespace(subjectSetNamespace).
		SubjectSetRelation(subjectSetRelation).
		SubjectSetObject(subjectSetObject).
		Execute()
	if err != nil {
		log.Error("Error when calling `PermissionApi.CheckPermission``:\n", err, " Http Response: ", r)
		return http.StatusFailedDependency, "", err
	}

	if !ketoResponse.Allowed {
		log.Info("Policy is not created for subjectSetNamespace: ", subjectSetNamespace, " subjectSetRelation: ", subjectSetRelation, " subjectSetObject: ", subjectSetObject, " Namespace: ", namespace, " Relation: ", relation, " Object: ", object)
		return http.StatusForbidden, "Policy does not exist", err
	}
	return http.StatusOK, "Policy exists", err
}

func (ketoSvc ketoService) ExpandPolicy (ctx context.Context, namespace string, relation string, object string, maxDepth string, hasDepth bool) (int, map[string]interface{}, error) {
	name := "CallKetoToExpandPolicy"
	childCtx, span := otel.Tracer(name).Start(ctx, "CallKetoToExpandPolicy")
	defer span.End()

	var (
		ketoResponse map[string]interface{}
		depth int64
		err error
	)
	if namespace=="" || relation=="" || object=="" || (hasDepth==true && maxDepth==""){
		log.Error("Invalid query params")
		return http.StatusBadRequest, ketoResponse, errors.New("Invalid query params")
	}
	if maxDepth != "" {
		depth, err = strconv.ParseInt(maxDepth, 10, 64)
		if err != nil {
			log.Error("MaxDepth cannot be converted to int: ", depth)
			return http.StatusBadRequest, ketoResponse, errors.New("Invalid query params")
		}
	}

	resp, r, err := utils.GetKetoReadClient().PermissionApi.ExpandPermissions(childCtx).
		Namespace(namespace).
		Relation(relation).
		Object(object).
		MaxDepth(depth).
		Execute()
	if err != nil {
		log.Error("Error when calling `PermissionApi.ExpandPermissions``:\n", err, " Http Response: ", r, "Response Body: ", resp)
		return http.StatusFailedDependency, ketoResponse, err
	}

	defer r.Body.Close()
	encodedBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
		return http.StatusFailedDependency, ketoResponse, err
	}

	json.Unmarshal([]byte(string(encodedBody)), &ketoResponse)
	status, _ := ketoResponse["code"].(float64)
	if status == http.StatusNotFound {
		log.Info("Subject set not found with Namespace: ", namespace, " Relation: ", relation, " Object: ", object)
		return http.StatusNotFound, ketoResponse, err
	}
	return http.StatusOK, ketoResponse, err
}