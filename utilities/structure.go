package utilities

import "github.com/uptrace/bun"

type OrderByDirection struct {
	Column    string
	Direction string
}

type DefaultFilters struct {
	Page  int
	Limit int
	Order []*OrderByDirection
}

func (f *DefaultFilters) ToWhere(query *bun.SelectQuery) *bun.SelectQuery {
	if f.Page > 0 {
		query = query.Offset((f.Page - 1) * f.Limit)
	}
	if f.Limit > 0 {
		query = query.Limit(f.Limit)
	}
	if len(f.Order) > 0 {
		for _, order := range f.Order {
			query = query.Order(order.Column + " " + order.Direction)
		}
	}
	return query.Where("deleted_at IS NULL")
}
