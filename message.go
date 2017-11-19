package gossip

import "fmt"

// Message is the representation of a simple JSON message for nodes communications.
type Message struct {
	ID 		int		`json:"id"`
	MsgType string	`json:"type"`
	Sender 	int		`json:"sender"`
	Origin 	int		`json:"origin"`
	Data 	string	`json:"data"`
}

// NewMessage creates new message from input parameters.
func NewMessage(id int, msgType string, sender int, origin int, data string) Message {
	return Message{ id, msgType, sender, origin, data}	
}

func (m Message) String() string {
	return fmt.Sprintf("{ ID: %d MsgType: %s Sender: %d Origin: %d Data: %s }",  
		m.ID, m.MsgType, m.Sender, m.Origin, m.Data)
}
