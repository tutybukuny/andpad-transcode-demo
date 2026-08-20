package entity

type IEntity interface {
	TableName() string
	IDField() string
}
