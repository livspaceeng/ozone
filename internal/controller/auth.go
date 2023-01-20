package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/livspaceeng/ozone/configs"
	service "github.com/livspaceeng/ozone/internal/services"
	log "github.com/sirupsen/logrus"
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
	hydraClient := c.Query("hydra")
	bearer := headers.Get("Authorization")

	hydraStatus, hydraResponse, err := a.hydraService.GetSubjectByToken(hydraClient, bearer)
	if hydraStatus != http.StatusOK {
		c.JSON(hydraStatus, err)
		return
	}

	//Keto
	var namespace, relation, object string = "", "", ""
	queries := strings.Split(c.Request.URL.RawQuery, "&")
	for _, query := range queries {
		query = strings.ToLower(query)
		if strings.HasPrefix(query, "namespace") {
			namespace = strings.Split(query, "=")[1]
		} else if strings.HasPrefix(query, "relation") {
			relation = strings.Split(query, "=")[1]
		}  else if strings.HasPrefix(query, "object") {
			object = strings.Split(query, "=")[1]
		}
	}

	ketoStatus, ketoResponse, err := a.ketoService.ValidatePolicy(namespace, relation, object, hydraResponse)

	if ketoStatus == http.StatusOK {
		c.JSON(ketoStatus, ketoResponse)
		return
	} else if ketoStatus == http.StatusForbidden {
		c.JSON(ketoStatus, ketoResponse)
		return
	} else {
		c.JSON(ketoStatus, err)
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
	// httpClient := &http.Client{}
	// config := configs.GetConfig()
	// ketoUrl := config.GetString("keto.read.url")

	// configuration := client.NewConfiguration()
    // configuration.Servers = []client.ServerConfiguration{
    //     {
    //         URL: ketoUrl, 
    //     },
    // }
    // apiClient := client.NewAPIClient(configuration)

	// namespace := c.Query("namespace")
    // object := c.Query("object")
    // relation := c.Query("relation")
	// if len(c.Query("subject_id")) > 0 {
	// 	subjectId := c.Query("subject_id")
	// 	resp, r, err := apiClient.ReadApi.GetCheck(context.Background()).Namespace(namespace).Object(object).Relation(relation).SubjectId(subjectId).Execute()
	// 	if err != nil {
	// 		log.Error("Error when calling `ReadApi.GetCheck``: %v\n", err)
	// 		log.Error("Full HTTP response: %v\n", r)
	// 	}
	// 	log.Info("Response from `ReadApi.GetCheck`: %v\n", resp)
	// 	c.JSON(http.StatusOK, resp)
	// } else {
	// 	subjectSetNamespace := c.Query("subject_set.namespace")
	// 	subjectSetObject := c.Query("subject_set.object")
	// 	subjectSetRelation := c.Query("subject_set.relation")
	// 	resp, r, err := apiClient.ReadApi.GetCheck(context.Background()).Namespace(namespace).Object(object).Relation(relation).SubjectSetNamespace(subjectSetNamespace).SubjectSetObject(subjectSetObject).SubjectSetRelation(subjectSetRelation).Execute()
	// 	if err != nil {
	// 		log.Error("Error when calling `ReadApi.GetCheck``: %v\n", err)
	// 		log.Error("Full HTTP response: %v\n", r)
	// 	}
	// 	log.Info("Response from `ReadApi.GetCheck`: %v\n", resp)
	// 	c.JSON(http.StatusOK, resp)
	// }

	// ketoPath := config.GetString("keto.read.path.check")
	// ketoRequest, _ := http.NewRequest(http.MethodGet, ketoUrl+ketoPath, nil)
	// ketoRequest.Header.Add("Accept", "application/json")

	// q := ketoRequest.URL.Query()
	// q.Add("namespace", c.Query("namespace"))
	// q.Add("object", c.Query("object"))
	// q.Add("relation", c.Query("relation"))

	var namespace, relation, object, subjectId, subjectSetNamespace, subjectSetRelation, subjectSetObject string = "", "", "", "", "", "", ""
	queries := strings.Split(c.Request.URL.RawQuery, "&")
	for _, query := range queries {
		query = strings.ToLower(query)
		if strings.HasPrefix(query, "namespace") {
			namespace = strings.Split(query, "=")[1]
		} else if strings.HasPrefix(query, "relation") {
			relation = strings.Split(query, "=")[1]
		}  else if strings.HasPrefix(query, "object") {
			object = strings.Split(query, "=")[1]
		} else if strings.HasPrefix(query, "subject_id") {
			subjectId = strings.Split(query, "=")[1]
		} else if strings.HasPrefix(query, "subject_set.namespace") {
			subjectSetNamespace = strings.Split(query, "=")[1]
		} else if strings.HasPrefix(query, "subject_set.relation") {
			subjectSetRelation = strings.Split(query, "=")[1]
		} else if strings.HasPrefix(query, "subject_set.object") {
			subjectSetObject = strings.Split(query, "=")[1]
		}
	}

	// if len(c.Query("subject_id")) > 0 {
	// 	q.Add("subject_id", c.Query("subject_id"))
	// } else {
	// 	q.Add("subject_set.namespace", c.Query("subject_set.namespace"))
	// 	q.Add("subject_set.object", c.Query("subject_set.object"))
	// 	q.Add("subject_set.relation", c.Query("subject_set.relation"))
	// }

	var (
		ketoStatus int
		ketoResponse string
		err error
	)
	if len(subjectId) > 0 {
		ketoStatus, ketoResponse, err = a.ketoService.ValidatePolicy(namespace, relation, object, subjectId)
	} else {
		ketoStatus, ketoResponse, err = a.ketoService.ValidatePolicyWithSet(namespace, relation, object, subjectSetNamespace, subjectSetRelation, subjectSetObject)
	}

	if ketoStatus == http.StatusOK {
		c.JSON(ketoStatus, ketoResponse)
		return
	} else if ketoStatus == http.StatusForbidden {
		c.JSON(ketoStatus, ketoResponse)
		return
	} else {
		c.JSON(ketoStatus, err)
	}

	// ketoRequest.URL.RawQuery = q.Encode()
	// log.Info(ketoRequest)
	// resp, err := httpClient.Do(ketoRequest)

	// if err != nil {
	// 	log.Error("Errored when sending request to the server", err.Error())
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }
	// defer resp.Body.Close()
	// encodedBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Error("Decoding error: ", err.Error())
	//   	c.AbortWithError(http.StatusInternalServerError, err)
	// 	  return
	// }
	// var body map[string]interface{}
	// json.Unmarshal([]byte(string(encodedBody)), &body)

	// _, errBody := body["error"]
	// if errBody {
	// 	log.Error("Encountered error: ", body["error"])
	// 	c.JSON(http.StatusBadRequest, body["error"])
	// 	return
	// }
	// log.Info("Response body : ", body)
	// c.JSON(http.StatusOK, body)
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
	httpClient := &http.Client{}
	config := configs.GetConfig()

	ketoUrl := config.GetString("keto.read.url")
	ketoPath := config.GetString("keto.read.path.expand")
	ketoRequest, _ := http.NewRequest(http.MethodGet, ketoUrl+ketoPath, nil)
	ketoRequest.Header.Add("Accept", "application/json")

	q := ketoRequest.URL.Query()
	q.Add("namespace", c.Query("namespace"))
	q.Add("object", c.Query("object"))
	q.Add("relation", c.Query("relation"))
	if len(c.Query("max-depth")) > 0 {
		q.Add("max-depth", c.Query("max-depth"))
	}

	ketoRequest.URL.RawQuery = q.Encode()
	log.Info(ketoRequest)
	resp, err := httpClient.Do(ketoRequest)

	if err != nil {
		log.Error("Errored when sending request to the server", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()
	encodedBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
	  	c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var body map[string]interface{}
	json.Unmarshal([]byte(string(encodedBody)), &body)

	_, errBody := body["error"]
	if errBody {
		log.Error("Encountered error: ", body["error"])
		c.JSON(http.StatusBadRequest, body["error"])
		return
	}
	log.Info("Response body : ", body)
	c.JSON(http.StatusOK, body)
}
