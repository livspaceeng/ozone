package model

type RelationTuple struct {
	Namespace  string `json:"namespace" example:"canvas"`
	Object     string `json:"object" example:"project-123"`
	Relation   string `json:"relation" example:"read"`
	Subject_Id string `json:"subject_id" example:"user-123"`
}
