package main

import (
	"fmt"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
)

type CustomScoreQuery struct {
	wrapped bluge.Query
	custom  func(match *search.DocumentMatch) *search.DocumentMatch
}

func NewCustomScoreQuery(q bluge.Query, custom func(match *search.DocumentMatch) *search.DocumentMatch) *CustomScoreQuery {
	return &CustomScoreQuery{
		wrapped: q,
		custom:  custom,
	}
}

func (c *CustomScoreQuery) Searcher(i search.Reader,
	options search.SearcherOptions) (search.Searcher, error) {
	searcher, err := c.wrapped.Searcher(i, options)
	if err != nil {
		return nil, fmt.Errorf("error wrapping searcher: %w", err)
	}
	return &CustomScoreSearcher{
		wrapped: searcher,
		custom:  c.custom,
	}, nil
}

type CustomScoreSearcher struct {
	wrapped search.Searcher
	custom  func(match *search.DocumentMatch) *search.DocumentMatch
}

func (c *CustomScoreSearcher) Next(ctx *search.Context) (*search.DocumentMatch, error) {
	rv, err := c.wrapped.Next(ctx)
	if err != nil {
		return nil, err
	}
	return c.custom(rv), nil
}

func (c *CustomScoreSearcher) Advance(ctx *search.Context, number uint64) (*search.DocumentMatch, error) {
	rv, err := c.wrapped.Advance(ctx, number)
	if err != nil {
		return nil, err
	}
	return c.custom(rv), nil
}

func (c *CustomScoreSearcher) Close() error {
	return c.wrapped.Close()
}
func (c *CustomScoreSearcher) Count() uint64 {
	return c.wrapped.Count()
}

func (c *CustomScoreSearcher) Min() int {
	return c.wrapped.Min()
}

func (c *CustomScoreSearcher) Size() int {
	return c.wrapped.Size()
}

func (c *CustomScoreSearcher) DocumentMatchPoolSize() int {
	return c.wrapped.DocumentMatchPoolSize()
}
