package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"time"
)

type UpdateTimeMixin struct {
	ent.Schema
}

func (UpdateTimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("update_time").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}
