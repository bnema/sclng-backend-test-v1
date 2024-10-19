package models

import "time"

type Language struct {
	Bytes int `json:"bytes"`
}

type CustomRepository struct {
	FullName   string              `json:"full_name"`
	Owner      string              `json:"owner"`
	Repository string              `json:"repository"`
	Languages  map[string]Language `json:"languages"`
	License    string              `json:"license"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	PushedAt   time.Time           `json:"pushed_at"`
	Stars      int                 `json:"stars"`
	Forks      int                 `json:"forks"`
	Issues     int                 `json:"issues"`
}
