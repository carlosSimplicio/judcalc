package domain

import "context"

type AreaRepository interface {
	ListAreas(context.Context, ListOptions) (ListResult[Area], error)
}

type ServiceRepository interface {
	ListServices(context.Context, ListOptions) (ListResult[Service], error)
}
