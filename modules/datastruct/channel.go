package datastruct

import (
	"fmt"

	"go_code/myselfgo/define"
	"go_code/myselfgo/utils"
)

type SafeStrBuffer struct {
	Cap int
	Ch  chan string
}

func NewStrBuf(cap int) *SafeStrBuffer {
	return &SafeStrBuffer{
		Cap: cap,
		Ch:  make(chan string, cap),
	}
}

func (sb *SafeStrBuffer) Push(msg string) error {
	msgCnt := len(sb.Ch)
	bufCap := sb.Cap
	if msgCnt >= bufCap {
		return fmt.Errorf("buffer is full, len: %d, cap: %d", msgCnt, bufCap)
	}
	sb.Ch <- msg
	return nil
}

func (sb *SafeStrBuffer) Pop() string {
	var msgs []string
	msgCnt := len(sb.Ch)
	for i := 0; i < msgCnt; i++ {
		msgs = append(msgs, <-sb.Ch)
	}

	if len(msgs) == 0 {
		return ""
	}

	return utils.JoinStrWithSep(define.SepDoubleEnter, msgs...)
}
