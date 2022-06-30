package model

type SubjectSet struct {
	Namespace   string `json:"namespace" example:"canvas"`
	Object      string `json:"object" example:"project-123"`
	Relation    string `json:"relation" example:"read"`
}