package schema

import (
	"entgo.io/ent"
	"zeus-backend-layout/internal/data/ent/mixin"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return nil
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CreateTimeMixin{},
		mixin.UpdateTimeMixin{},
		mixin.DeleteTimeMixin{},
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
