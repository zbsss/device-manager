package scheduler

import "time"

type TokenLease struct {
	PodId     string
	ExpiresAt time.Time
}

type TokenLeaseRequest struct {
	PodId    string
	Response chan *TokenLease
}

type LeaseHistoryEntry struct {
	PodId      string    `json:"podId"`
	LeasedAt   time.Time `json:"leasedAt"`
	ReturnedAt time.Time `json:"returnedAt"`
}

type PodQuota struct {
	PodId    string
	Requests float64
	Limit    float64
}
