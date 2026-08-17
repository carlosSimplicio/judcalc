package domain

type ListOptions struct {
	Page     int64
	PageSize int64
	Query    string
}

func (options ListOptions) Offset() int64 {
	return (options.Page - 1) * options.PageSize
}

type ListResult[T any] struct {
	Items []T
	Total int64
}
