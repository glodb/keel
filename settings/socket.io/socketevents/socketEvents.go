package socketevents

type ClientEvents string

const (
	LOGIN          = "socketLogin"
	MESSAGE        = "socketMessage"
	SUCCESS        = "success"
	LOGIN_RESPONSE = "loginResponse"
)

const (
	SEND_CHAT = ClientEvents("chatMessage")
)
