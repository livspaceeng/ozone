package unit_tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livspaceeng/ozone/internal/server"
	"github.com/livspaceeng/ozone/internal/utils"
	"github.com/livspaceeng/ozone/configs"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type input struct{
	namespace string
	relation string
	object string
	subjectId string
	subjectSetNamespace string
	subjectSetObject string
	subjectSetRelation string
	authPrefix string
	maxDepth string
}

type expectation struct{
	status int
	out string
	err error
}

func TestAuthController_Check(t *testing.T) {
	configs.Init()
	utils.Init()
	config := configs.GetConfig()
	r := server.NewRouter()

	tests := map[string]struct {
		in       input
		expected expectation
	}{
		"ValidPolicy": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "get",
				object: "com.livspace.auth;bouncer;users",
				authPrefix: "Bearer ",
			},
			expected: expectation{
				status: 200,
				out: "\"com.livspace.auth;bouncer;users;9338\"",
				err: nil,
				},
			},
		"InvalidPolicy": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "post",
				object: "com.livspace.auth;bouncer;users",
				authPrefix: "Bearer ",
			},
			expected: expectation{
				status: 403,
				out: "\"com.livspace.auth;bouncer;users;9338\"",
				err: nil,
				},
			},
		"InvalidQueryParams": {
			in: input{
				namespace: "",
				relation: "",
				object: "",
				authPrefix: "Bearer ",
			},
			expected: expectation{
				status: 400,
				out: "\"Invalid query params\"",
				err: nil,
				},
			},
		"UnauthorizedRequest": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "get",
				object: "com.livspace.auth;bouncer;users",
				authPrefix: "",
			},
			expected: expectation{
				status: 401,
				out: "\"Authorization header format is not valid\"",
				err: nil,
				},
			},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/auth/check?namespace="+tt.in.namespace+"&relation="+tt.in.relation+"&object="+tt.in.object, nil)
			req.Header.Add("Authorization",tt.in.authPrefix+config.GetString("hydra.bouncer.access_token"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			log.Info("Response: ", w.Body)
			assert.Equal(t, tt.expected.out, w.Body.String())
			assert.Equal(t, tt.expected.status, w.Code)
		})
	}
}

func TestAuthController_Query(t *testing.T) {
	configs.Init()
	r := server.NewRouter()

	tests := map[string]struct {
		in       input
		expected expectation
	}{
		"QueryPolicy": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "member",
				object: "com.livspace.auth;bouncer;roles;BOUNCER_VIEWER",
				subjectId: "com.livspace.auth;bouncer;users;9338",
			},
			expected: expectation{
				status: 200,
				out: "\"com.livspace.auth;bouncer;users;9338\"",
				err: nil,
				},
			},
		"QueryInvalidPolicy": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "member",
				object: "com.livspace.auth;bouncer;roles;BOUNCER_VIEWER",
				subjectId: "com.livspace.auth;bouncer;users;6178",
			},
			expected: expectation{
				status: 403,
				out: "\"com.livspace.auth;bouncer;users;6178\"",
				err: nil,
				},
			},
		"QueryPolicyWithSet": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "get",
				object: "com.livspace.auth;bouncer;users",
				subjectSetNamespace: "com.livspace.auth",
				subjectSetRelation: "member",
				subjectSetObject: "com.livspace.auth;bouncer;roles;BOUNCER_VIEWER",
			},
			expected: expectation{
				status: 200,
				out: "\"Policy exists\"",
				err: nil,
				},
			},
		"QueryInvalidPolicyWithSet": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "get",
				object: "com.livspace.auth;bouncer;user",
				subjectSetNamespace: "com.livspace.auth",
				subjectSetRelation: "member",
				subjectSetObject: "com.livspace.auth;bouncer;roles;BOUNCER_VIEWER",
			},
			expected: expectation{
				status: 403,
				out: "\"Policy does not exist\"",
				err: nil,
				},
			},
		"InvalidQueryParams": {
			in: input{
				namespace: "",
				relation: "",
				object: "",
				subjectId: "com",
			},
			expected: expectation{
				status: 400,
				out: "\"Invalid query params\"",
				err: nil,
				},
			},
		"InvalidQueryParamsWithSet": {
			in: input{
				namespace: "",
				relation: "",
				object: "",
				subjectSetNamespace: "",
				subjectSetRelation: "",
				subjectSetObject: "",
			},
			expected: expectation{
				status: 400,
				out: "\"Invalid query params\"",
				err: nil,
				},
			},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/auth/relation_tuples?namespace="+tt.in.namespace+"&relation="+tt.in.relation+"&object="+tt.in.object+"&subject_id="+tt.in.subjectId, nil)
			if len(tt.in.subjectId) == 0 {
				req, _ = http.NewRequest("GET", "/api/v1/auth/relation_tuples?namespace="+tt.in.namespace+"&relation="+tt.in.relation+"&object="+tt.in.object+"&subject_set.namespace="+tt.in.subjectSetNamespace+"&subject_set.relation="+tt.in.subjectSetRelation+"&subject_set.object="+tt.in.subjectSetObject, nil)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			log.Info("Response: ", w.Body)
			assert.Equal(t, tt.expected.out, w.Body.String())
			assert.Equal(t, tt.expected.status, w.Code)
		})
	}
}

func TestAuthController_Expand(t *testing.T) {
	configs.Init()
	r := server.NewRouter()

	tests := map[string]struct {
		in       input
		expected expectation
	}{
		"ExpandPolicy": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "get",
				object: "com.livspace.auth;bouncer;users",
				maxDepth: "2",
			},
			expected: expectation{
				status: 200,
				err: nil,
				},
			},
		"ExpandInvalidPolicy": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "get",
				object: "com.livspace.auth;bouncer;user",
				maxDepth: "1",
			},
			expected: expectation{
				status: 404,
				err: nil,
				},
			},
		"InvalidMaxDepth": {
			in: input{
				namespace: "com.livspace.auth",
				relation: "get",
				object: "com.livspace.auth;bouncer;users",
				maxDepth: "abc",
			},
			expected: expectation{
				status: 400,
				err: nil,
				},
			},
		"InvalidQueryParams": {
			in: input{
				namespace: "",
				relation: "",
				object: "",
				maxDepth: "0",
			},
			expected: expectation{
				status: 400,
				err: nil,
				},
			},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/auth/expand?namespace="+tt.in.namespace+"&relation="+tt.in.relation+"&object="+tt.in.object+"&max-depth="+tt.in.maxDepth, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.expected.status, w.Code)
		})
	}
}