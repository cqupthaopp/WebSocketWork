package main

import (
	"time"
)

type Msg struct {
	time    time.Time
	User    string
	Content []byte
}
