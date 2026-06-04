package socketevents

type ClientEvents string

const (
	LOGIN          = "socketLogin"
	SUCCESS        = "success"
	LOGIN_RESPONSE = "loginResponse"
)

const (
	SEND_CHAT = ClientEvents("chatMessage")
)
