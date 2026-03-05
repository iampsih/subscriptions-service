package domain

import "time"

type Subscription struct {
	ID          string
	ServiceName string
	Price       int
	UserID      string
	StartMonth  time.Time
	EndMonth    *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
