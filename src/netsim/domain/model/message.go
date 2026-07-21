package model

import "time"

type Message struct {
	SenderID   string
	Payload    string
	From       time.Time
	To         time.Time
	TxNs       uint64
	Overlapped bool
	Retries    uint32
}

type DelayedMessage struct {
	Msg        Message
	RetryAfter time.Time
}
