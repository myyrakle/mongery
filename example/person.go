package models

// @Entity
type Person struct {
	ID   string `bson:"_id,omitempty"`
	Name string `bson:"name"` // 이름
}
