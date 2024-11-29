package repository

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidPageSize = errors.New("invalid page size")
	ErrInvalidCursor   = errors.New("invalid cursor")
)

const (
	minPageSize = 1
	maxPageSize = 30
)

type Cursor struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

func (c *Cursor) limitPlusOne() string {
	limit := c.limit()
	checkNext := limit + 1 // выбираем на один больше, чтобы проверить есть ли элементы впереди
	return fmt.Sprintf("LIMIT %d OFFSET %d", checkNext, c.offset())
}

func (c *Cursor) limit() int {
	return c.PageSize
}

func (c *Cursor) offset() int {
	return (c.Page - 1) * c.PageSize
}

func (c *Cursor) Validate() error {
	if c.Page < 1 {
		return fmt.Errorf("Page field must be greather than 1")
	}
	return validatePageSize(c.PageSize)
}

func validatePageSize(ps int) error {
	if ps := ps; ps < minPageSize || ps > maxPageSize {
		return fmt.Errorf("PageSize field must be in [%d, %d]", minPageSize, maxPageSize)
	}
	return nil
}

type PaginationInfo struct {
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}
