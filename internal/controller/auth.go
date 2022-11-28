package controller

import (
	// "context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/livspaceeng/ozone/configs"
	"github.com/livspaceeng/ozone/internal/model"
	// client "github.com/ory/keto-client-go"
	log "github.com/sirupsen/logrus"
)

type AuthController struct{}

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
func (a AuthController) Check(c *gin.Context) {
	httpClient := &http.Client{}
	config := configs.GetConfig()

	//Hydra
	headers := c.Request.Header
	hydraClient := c.Query("hydra")
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
	bearer := headers.Get("Authorization")
	if len(bearer) <= 0 {
		log.Error("Bearer token absent!")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// log.Info("Token: ", bearer)
	validBearer := strings.Contains(bearer, "Bearer") || strings.Contains(bearer, "bearer")
	if !validBearer {
		log.Error("Authorization header format is not valid!")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	token := strings.Split(bearer, " ")[1]
	data := url.Values{}
	data.Set("token", token)
	hydraRequest, _ := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	hydraRequest.Header.Add("Authorization", bearer)
	hydraRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(hydraRequest)
	if err != nil {
		log.Error("Errored when sending request to the server", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	var hydraResponse model.HydraResponse
	err = json.NewDecoder(resp.Body).Decode(&hydraResponse)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	log.Info("Subject: ", hydraResponse.Subject)
	if hydraResponse.Subject == "" {
		log.Error("Subject is nil!")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//Keto
	ketoUrl := config.GetString("keto.read.url")
	ketoPath := config.GetString("keto.read.path.check")
	ketoRequest, _ := http.NewRequest(http.MethodGet, ketoUrl+ketoPath, nil)
	ketoRequest.Header.Add("Accept", "application/json")
	q := ketoRequest.URL.Query()
	q.Add("namespace", c.Query("namespace"))
	q.Add("subject_id", hydraResponse.Subject)
	q.Add("relation", c.Query("relation"))
	q.Add("object", c.Query("object"))
	ketoRequest.URL.RawQuery = q.Encode()
	log.Info(ketoRequest)
	resp, err = httpClient.Do(ketoRequest)

	if err != nil {
		log.Error("Errored when sending request to the server", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	defer resp.Body.Close()
	var ketoResponse model.KetoResponse
	err = json.NewDecoder(resp.Body).Decode(&ketoResponse)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	if !ketoResponse.Allowed {
		log.Info("Policy is not created for subject:", hydraResponse.Subject)
		log.Info("Namespace:", c.Query("namespace"))
		log.Info("Relation:", c.Query("relation"))
		log.Info("Object:", c.Query("object"))
		c.JSON(http.StatusForbidden, hydraResponse.Subject)
		return
	}
	c.JSON(http.StatusOK, hydraResponse.Subject)
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
func (a AuthController) Query(c *gin.Context) {
	httpClient := &http.Client{}
	config := configs.GetConfig()
	ketoUrl := config.GetString("keto.read.url")

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

	ketoPath := config.GetString("keto.read.path.check")
	ketoRequest, _ := http.NewRequest(http.MethodGet, ketoUrl+ketoPath, nil)
	ketoRequest.Header.Add("Accept", "application/json")

	q := ketoRequest.URL.Query()
	q.Add("namespace", c.Query("namespace"))
	q.Add("object", c.Query("object"))
	q.Add("relation", c.Query("relation"))

	if len(c.Query("subject_id")) > 0 {
		q.Add("subject_id", c.Query("subject_id"))
	} else {
		q.Add("subject_set.namespace", c.Query("subject_set.namespace"))
		q.Add("subject_set.object", c.Query("subject_set.object"))
		q.Add("subject_set.relation", c.Query("subject_set.relation"))
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
func (a AuthController) Expand(c *gin.Context) {
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
