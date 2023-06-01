package controller

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	service "github.com/livspaceeng/ozone/internal/services"
	"github.com/livspaceeng/ozone/internal/utils"
)

type AuthController interface {
	Check(c *gin.Context)
	Query(c *gin.Context)
	Expand(c *gin.Context)
}

type authController struct{
	hydraService service.HydraService
	ketoService service.KetoService
}

func NewAuthController(hydraSvc service.HydraService, ketoSvc service.KetoService) AuthController {
	return &authController{
		hydraService: hydraSvc,
		ketoService: ketoSvc,
	}
}

// AuthController godoc
// @Summary      auth check
// @Schemes      http
// @Description  check token and policy
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        namespace      query      string  true  "namespace"
// @Param        object         query      string  true  "resource"
// @Param        relation       query      string  true  "access-type"
// @Param        hydra          query      string  false "Default value is Bouncer. Use 'accounts' value for Accounts Hydra"
// @Param        Authorization  header     string  true  "Bearer <Bouncer_access_token>" 
// @Success      200         {string}  model.KetoResponse
// @Failure      400         {object}  model.KetoResponse
// @Failure      401         {object}  model.KetoResponse
// @Failure      403         {object}  model.KetoResponse
// @Failure      500         {object}  model.KetoResponse
// @Router       /auth/check [get]
func (a authController) Check(c *gin.Context) {
	//Hydra
	headers := c.Request.Header
	bearer := headers.Get("Authorization")
	var (
		namespace, relation, object, hydraClient string = "", "", "", ""
		hasHydra bool = false
	)
	queries := strings.Split(c.Request.URL.RawQuery, "&")
	for _, query := range queries {
		if strings.HasPrefix(query, utils.NamespaceString) {
			namespace = strings.Split(query, "=")[1]
			namespace, _ = url.QueryUnescape(namespace)
		} else if strings.HasPrefix(query, utils.RelationString) {
			relation = strings.Split(query, "=")[1]
			relation, _ = url.QueryUnescape(relation)
		} else if strings.HasPrefix(query, utils.ObjectString) {
			object = strings.Split(query, "=")[1]
			object, _ = url.QueryUnescape(object)
		} else if strings.HasPrefix(query, "hydra=") {
			hasHydra = true
			hydraClient = strings.Split(query, "=")[1]
		}
	}

	hydraStatus, hydraResponse, err := a.hydraService.GetSubjectByToken(c.Request.Context(), hydraClient, hasHydra, bearer)
	if hydraStatus == http.StatusFailedDependency {
		c.JSON(hydraStatus, err)
		return
	} else if hydraStatus == http.StatusUnauthorized || hydraStatus == http.StatusBadRequest {
		c.JSON(hydraStatus, err.Error())
		return
	}

	//Keto
	ketoStatus, ketoResponse, err := a.ketoService.ValidatePolicy(c.Request.Context(), namespace, relation, object, hydraResponse)

	if ketoStatus == http.StatusOK || ketoStatus == http.StatusForbidden {
		c.JSON(ketoStatus, ketoResponse)
		return
	} else if ketoStatus == http.StatusFailedDependency {
		c.JSON(ketoStatus, err)
		return
	} else {
		c.JSON(ketoStatus, err.Error())
	}
}

// AuthController godoc
// @Summary      query relation tuple
// @Schemes      http
// @Description  query relation tuple
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        namespace               query      string  true  "namespace"
// @Param        subject_id              query      string  true  "subject"
// @Param        object                  query      string  true  "resource"
// @Param        relation                query      string  true  "access-type"
// @Param        subject_set.namespace   query      string  true  "subject_set namespace"
// @Param        subject_set.object      query      string  true  "subject_set object"
// @Param        subject_set.relation    query      string  true  "subject_set relation"
// @Param        Authorization           header     string  true  "Bearer <Bouncer_access_token>" 
// @Success      200             {string}  model.KetoResponse
// @Failure      400             {object}  model.KetoResponse
// @Failure      401             {object}  model.KetoResponse
// @Failure      403             {object}  model.KetoResponse
// @Failure      500             {object}  model.KetoResponse
// @Router       /auth/relation_tuples [get]
func (a authController) Query(c *gin.Context) {
	var namespace, relation, object, subjectId, subjectSetNamespace, subjectSetRelation, subjectSetObject string = "", "", "", "", "", "", ""
	queries := strings.Split(c.Request.URL.RawQuery, "&")
	for _, query := range queries {
		if strings.HasPrefix(query, utils.NamespaceString) {
			namespace = strings.Split(query, "=")[1]
			namespace, _ = url.QueryUnescape(namespace)
		} else if strings.HasPrefix(query, utils.RelationString) {
			relation = strings.Split(query, "=")[1]
			relation, _ = url.QueryUnescape(relation)
		}  else if strings.HasPrefix(query, utils.ObjectString) {
			object = strings.Split(query, "=")[1]
			object, _ = url.QueryUnescape(object)
		} else if strings.HasPrefix(query, "subject-id=") {
			subjectId = strings.Split(query, "=")[1]
			subjectId, _ = url.QueryUnescape(subjectId)
		} else if strings.HasPrefix(query, "subject-set-namespace=") {
			subjectSetNamespace = strings.Split(query, "=")[1]
			subjectSetNamespace, _ = url.QueryUnescape(subjectSetNamespace)
		} else if strings.HasPrefix(query, "subject-set-relation=") {
			subjectSetRelation = strings.Split(query, "=")[1]
			subjectSetRelation, _ = url.QueryUnescape(subjectSetRelation)
		} else if strings.HasPrefix(query, "subject-set-object=") {
			subjectSetObject = strings.Split(query, "=")[1]
			subjectSetObject, _ = url.QueryUnescape(subjectSetObject)
		}
	}

	var (
		ketoStatus int
		ketoResponse string
		err error
	)
	if len(subjectId) > 0 {
		ketoStatus, ketoResponse, err = a.ketoService.ValidatePolicy(c.Request.Context(), namespace, relation, object, subjectId)
	} else {
		ketoStatus, ketoResponse, err = a.ketoService.ValidatePolicyWithSet(c.Request.Context(), namespace, relation, object, subjectSetNamespace, subjectSetRelation, subjectSetObject)
	}

	if ketoStatus == http.StatusOK || ketoStatus == http.StatusForbidden {
		c.JSON(ketoStatus, ketoResponse)
		return
	} else if ketoStatus == http.StatusFailedDependency {
		c.JSON(ketoStatus, err)
		return
	} else {
		c.JSON(ketoStatus, err.Error())
	}
}

// AuthController godoc
// @Summary      expand relation tuple
// @Schemes      http
// @Description  expand relation tuple
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        namespace      query     string   true  "namespace"
// @Param        max-depth      query     integer  true  "max-depth to expand tuple"
// @Param        object         query     string   true  "resource"
// @Param        relation       query     string   true  "access-type"
// @Param        Authorization  header    string   true  "Bearer <Bouncer_access_token>" 
// @Success      200             {string}  model.KetoResponse
// @Failure      400             {object}  model.KetoResponse
// @Failure      401             {object}  model.KetoResponse
// @Failure      403             {object}  model.KetoResponse
// @Failure      500             {object}  model.KetoResponse
// @Router       /auth/expand [get]
func (a authController) Expand(c *gin.Context) {
	var (
		namespace, relation, object, maxDepth string = "", "", "", ""
		hasDepth bool = false
	)
	queries := strings.Split(c.Request.URL.RawQuery, "&")
	for _, query := range queries {
		if strings.HasPrefix(query, utils.ObjectString) {
			object = strings.Split(query, "=")[1]
			object, _ = url.QueryUnescape(object)
		} else if strings.HasPrefix(query, utils.NamespaceString) {
			namespace = strings.Split(query, "=")[1]
			namespace, _ = url.QueryUnescape(namespace)
		} else if strings.HasPrefix(query, "max-depth=") {
			hasDepth = true
			maxDepth = strings.Split(query, "=")[1]
		} else if strings.HasPrefix(query, utils.RelationString) {
			relation = strings.Split(query, "=")[1]
			relation, _ = url.QueryUnescape(relation)
		} 
	}

	ketoStatus, ketoResponse, err := a.ketoService.ExpandPolicy(c.Request.Context(), namespace, relation, object, maxDepth, hasDepth)

	if ketoStatus == http.StatusOK || ketoStatus == http.StatusNotFound {
		c.JSON(ketoStatus, ketoResponse)
		return
	} else if ketoStatus == http.StatusFailedDependency {
		c.JSON(ketoStatus, err)
		return
	} else {
		c.JSON(ketoStatus, err.Error())
	}
}
