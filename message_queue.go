package gossip

import (
	"math/rand"
	"strconv"
	"strings"
)

type preparedMessage struct {
	msg              Message
	distributionList []int
	ttl              int
}

func newPreparedMessage(msg Message, list []int, ttl int) *preparedMessage {
	return &preparedMessage{msg, list, ttl}
}

type messageQueue struct {
	q []*preparedMessage
}

func newMessageQueue() *messageQueue {
	return &messageQueue{q: make([]*preparedMessage, 0, 100)}
}

func (q *messageQueue) String() string {
	intSliceToString := func(values []int) string {
		valuesText := []string{}
		for _, val := range values {
			text := strconv.Itoa(val)
			valuesText = append(valuesText, text)
		}
		return "[ " + strings.Join(valuesText, " ") + " ]"
	}

	res := ""
	for _, val := range q.q {
		res += strconv.Itoa(val.msg.ID) + " "
		res += intSliceToString(val.distributionList)
		res += "\n"
	}
	return res
}

// putMessage emulates enqueue oreration.
//
// msg and its potential recipients are memorized.
// Ttl is inited from package constant.
func (q *messageQueue) putMessage(msg Message, recipients []int) {
	q.q = append(q.q, newPreparedMessage(msg, recipients, TTL))
}

// getMessage emulates dequeue operation.
//
// It gets random message and random recipient from slice.
// NOTE: seed for random has to be inited in specific goroutine.
func (q *messageQueue) getMessage() (msg Message, id int, empty bool) {
	if len(q.q) == 0 {
		return Message{}, 0, true
	}
	r := int(rand.Int31n(int32(len(q.q))))
	message := q.q[r]
	t := int(rand.Int31n(int32(len((*message).distributionList))))
	recipient := (*message).distributionList[t]
	q.q[r].ttl--
	if q.q[r].ttl == 0 {
		q.q = append(q.q[:r], q.q[r+1:]...)
	}
	return message.msg, recipient, false
}
