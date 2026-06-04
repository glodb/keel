package socketmodels

type Message struct {
	RoomId  string `json:"roomId"`
	Content string `json:"content"`
}

func (m *Message) GetLength() int {
	return len(m.RoomId) + len(m.Content)
}

func (m *Message) String() string {
	return "{RoomId:" + m.RoomId + ", Content:" + m.Content + "}"
}
