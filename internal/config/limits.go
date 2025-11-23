package config

import "time"


const (
	MaxStringLength          = 255
	MaxTeamMembers           = 100
	DefaultReviewers         = 2
	ReplacementReviewerCount = 1

	DefaultHTTPAddr = "0.0.0.0"

	ReadTimeout     = 15 * time.Second
	WriteTimeout    = 15 * time.Second
	IdleTimeout     = 60 * time.Second
	ShutdownTimeout = 30 * time.Second
)
