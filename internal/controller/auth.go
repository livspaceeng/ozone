package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/livspaceeng/ozone/configs"
	"github.com/livspaceeng/ozone/internal/model"
	acl "github.com/ory/keto/proto/ory/keto/acl/v1alpha1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type AuthController struct{}

// AuthController godoc
// @Summary      auth check
// @Schemes      http
// @Description  check token and policy
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        namespace   path      string  true  "namespace"
// @Param        subject_id  path      string  true  "subject"
// @Param        object      path      string  true  "resource"
// @Param        relation    path      string  true  "access-type"
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
	hydraUrl := config.GetString("hydra.url")
	hydraPath := config.GetString("hydra.path.introspect")
	u, _ := url.ParseRequestURI(hydraUrl)
	u.Path = hydraPath
	headers := c.Request.Header
	bearer := headers.Get("Authorization")
	if len(bearer) <= 0 {
		log.Error("Bearer token absent!")
		c.AbortWithError(http.StatusUnauthorized, nil)
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
		log.Info("Not Allowed!")
		c.AbortWithStatus(http.StatusForbidden)
	}

}

// AuthController godoc
// @Summary      create relation tuple
// @Schemes      http
// @Description  create relation tuple
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        relation_tuple  body      model.RelationTuple  true  "Relation Data"
// @Success      200         {string}  model.KetoResponse
// @Failure      400         {object}  model.KetoResponse
// @Failure      401         {object}  model.KetoResponse
// @Failure      403         {object}  model.KetoResponse
// @Failure      500         {object}  model.KetoResponse
// @Router       /auth/relation_tuples [put]
func (a AuthController) Create(c *gin.Context) {
	config := configs.GetConfig()
	var relation model.RelationTuple
	err := json.NewDecoder(c.Request.Body).Decode(&relation)
	if err != nil {
		log.Error("Decoding error: ", err.Error())
		c.AbortWithError(http.StatusBadRequest, err)
	}
	log.Info("Relation: ", relation)

	conn, err := grpc.Dial(config.GetString("keto.write.url"), grpc.WithInsecure())
	if err != nil {
		log.Error("Encountered error: ", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	client := acl.NewWriteServiceClient(conn)

	if len(relation.Subject_Id) > 0 {
		_, err = client.TransactRelationTuples(context.Background(), &acl.TransactRelationTuplesRequest{
			RelationTupleDeltas: []*acl.RelationTupleDelta{
				{
					Action: acl.RelationTupleDelta_INSERT,
					RelationTuple: &acl.RelationTuple{
						Namespace: relation.Namespace,
						Object:    relation.Object,
						Relation:  relation.Relation,
						Subject:   &acl.Subject{Ref: &acl.Subject_Id{Id: relation.Subject_Id}},
					},
				},
			},
		})
		if err != nil {
			log.Error("Encountered error: ", err.Error())
			c.AbortWithError(http.StatusBadRequest, err)
			return 
		}
	} else {
		_, err = client.TransactRelationTuples(context.Background(), &acl.TransactRelationTuplesRequest{
			RelationTupleDeltas: []*acl.RelationTupleDelta{
				{
					Action: acl.RelationTupleDelta_INSERT,
					RelationTuple: &acl.RelationTuple{
						Namespace: relation.Namespace,
						Object:    relation.Object,
						Relation:  relation.Relation,
						Subject:   &acl.Subject{Ref: &acl.Subject_Set{Set: &acl.SubjectSet{
							Namespace: relation.Subject_Set.Namespace,
							Object:    relation.Subject_Set.Object,
							Relation:  relation.Subject_Set.Relation,
						}}},
					},
				},
			},
		})
		if err != nil {
			log.Error("Encountered error: ", err.Error())
			c.AbortWithError(http.StatusBadRequest, err)
			return 
		}
	}

	log.Info("Successfully created tuple!")

}

// AuthController godoc
// @Summary      delete relation tuple
// @Schemes      http
// @Description  delete relation tuple
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        relation_tuple  body      model.RelationTuple  true  "Relation Data"
// @Success      200         {string}  model.KetoResponse
// @Failure      400         {object}  model.KetoResponse
// @Failure      401         {object}  model.KetoResponse
// @Failure      403         {object}  model.KetoResponse
// @Failure      500         {object}  model.KetoResponse
// @Router       /auth/relation_tuples [delete]
func (a AuthController) Delete(c *gin.Context) {
}

// AuthController godoc
// @Summary      query relation tuple
// @Schemes      http
// @Description  query relation tuple
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        namespace   path      string  true  "namespace"
// @Param        subject_id  path      string  true  "subject"
// @Param        object      path      string  true  "resource"
// @Param        relation    path      string  true  "access-type"
// @Success      200             {string}  model.KetoResponse
// @Failure      400             {object}  model.KetoResponse
// @Failure      401             {object}  model.KetoResponse
// @Failure      403             {object}  model.KetoResponse
// @Failure      500             {object}  model.KetoResponse
// @Router       /auth/relation_tuples [get]
func (a AuthController) Query(c *gin.Context) {
}

// AuthController godoc
// @Summary      expand relation tuple
// @Schemes      http
// @Description  expand relation tuple
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        namespace   path      string  true  "namespace"
// @Param        subject_id  path      string  true  "subject"
// @Param        object      path      string  true  "resource"
// @Param        relation    path      string  true  "access-type"
// @Success      200             {string}  model.KetoResponse
// @Failure      400             {object}  model.KetoResponse
// @Failure      401             {object}  model.KetoResponse
// @Failure      403             {object}  model.KetoResponse
// @Failure      500             {object}  model.KetoResponse
// @Router       /auth/expand [get]
func (a AuthController) Expand(c *gin.Context) {
}
