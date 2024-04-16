package model

type Mutation string

const (
	CreateOrUpdate Mutation = "create-or-update"
	Delete         Mutation = "delete"
)
