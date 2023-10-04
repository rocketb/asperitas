package paging

import (
	"net/http"
	"strconv"

	"github.com/rocketb/asperitas/pkg/validate"
)

type Response[T any] struct {
	Items       []T `json:"items"`
	Total       int `json:"total"`
	Page        int `json:"page"`
	RowsPerPage int `json:"rows_per_page"`
}

func NewResponse[T any](items []T, total int, page int, rowsPerPage int) Response[T] {
	return Response[T]{
		Items:       items,
		Total:       total,
		Page:        page,
		RowsPerPage: rowsPerPage,
	}
}

type Page struct {
	Number      int
	RowsPerPage int
}

func ParseRequest(r *http.Request) (Page, error) {
	values := r.URL.Query()

	number := 1
	if page := values.Get("page"); page != "" {
		var err error
		number, err = strconv.Atoi(page)
		if err != nil {
			return Page{}, validate.NewFieldsError("page", err)
		}
	}

	rowsPerPage := 10
	if rows := values.Get("rows"); rows != "" {
		var err error
		rowsPerPage, err = strconv.Atoi(rows)
		if err != nil {
			return Page{}, validate.NewFieldsError("rows", err)
		}
	}
	return Page{
		Number:      number,
		RowsPerPage: rowsPerPage,
	}, nil
}
