package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"time"
)

type CreateTimeMixin struct {
	ent.Schema
}

func (CreateTimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("create_time").
			Default(time.Now).
			Immutable(),
	}
}
