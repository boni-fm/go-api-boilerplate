package models

import "time"

type PingPongResponse struct {
	IsSuccess bool      `json:"is_success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
