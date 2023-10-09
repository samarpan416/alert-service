package jobStatusChangeTypes

import (
	"encoding/json"
	"time"
)

type ESResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int           `json:"total"`
		MaxScore float64       `json:"max_score"`
		Hits     []interface{} `json:"hits"`
	} `json:"hits"`
	Aggregations json.RawMessage `json:"aggregations"`
}

type JobStatusChangeAggregation struct {
	Cloud struct {
		DocCountErrorUpperBound int           `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int           `json:"sum_other_doc_count"`
		Buckets                 []CloudBucket `json:"buckets"`
	} `json:"cloud"`
}

type CloudBucket struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
	Tenant   struct {
		DocCountErrorUpperBound int            `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int            `json:"sum_other_doc_count"`
		Buckets                 []TenantBucket `json:"buckets"`
	} `json:"tenant"`
}

type TenantBucket struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
	Channel  struct {
		DocCountErrorUpperBound int             `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int             `json:"sum_other_doc_count"`
		Buckets                 []ChannelBucket `json:"buckets"`
	} `json:"channel"`
}

type ChannelBucket struct {
	Key          string `json:"key"`
	DocCount     int    `json:"doc_count"`
	LastStatuses struct {
		Hits struct {
			Total    int         `json:"total"`
			MaxScore interface{} `json:"max_score"`
			Hits     []Event     `json:"hits"`
		} `json:"hits"`
	} `json:"last_statuses"`
}

type Event struct {
	Index  string      `json:"_index"`
	Type   string      `json:"_type"`
	ID     string      `json:"_id"`
	Score  interface{} `json:"_score"`
	Source struct {
		ErrorMessage string    `json:"error_message"`
		Timestamp    time.Time `json:"@timestamp"`
		Stepname     string    `json:"stepname"`
		Status       string    `json:"status"`
	} `json:"_source"`
	Sort []int64 `json:"sort"`
}
