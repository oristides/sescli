package query

import (
	"sescli/internal/sescapi"
	"sescli/internal/what"
	"sescli/internal/when"
	"sescli/internal/where"
)

type Query struct {
	When  when.Filter
	Where where.Resolved
	What  what.Resolved
	Page  PageOptions
	Out   OutputOptions
}

type PageOptions struct {
	Limit int
	Page  int
}

type OutputOptions struct {
	Format       string
	IncludeRaw   bool
	SummaryChars int
}

func (q Query) EventsQuery() sescapi.EventsQuery {
	limit := q.Page.Limit
	if limit <= 0 {
		limit = 40
	}
	page := q.Page.Page
	if page <= 0 {
		page = 1
	}
	return sescapi.EventsQuery{
		Units:         q.Where.IDs,
		Audience:      q.What.Audience,
		ActivityTypes: q.What.ActivityTypes,
		Language:      q.What.Language,
		From:          q.When.From,
		To:            q.When.To,
		PerPage:       limit,
		Page:          page,
	}
}
