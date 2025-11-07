package models

import "time"

type NewsRequest struct {
	URL               string    `json:"url" binding:"required"`
	Title             string    `json:"title" binding:"required"`
	Content           string    `json:"content" binding:"required"`
	Abstract          string    `json:"abstract"`
	Author            string    `json:"author"`
	AuthorDescription string    `json:"author_description"`
	MediaOutlet       string    `json:"media_outlet" binding:"required"`
	Country           string    `json:"country" binding:"required"`
	PublishedDate     time.Time `json:"published_date" binding:"required"`
	Multimedia        []string  `json:"multimedia"`
}

type NewsResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
