package rpcreplymodels

type RegistrationReply struct {
	BaseReply
	SSOId string `json:"ssoId,omitempty"`
	Token string `json:"token,omitempty"`
}

type UpdateUserReply struct {
	BaseReply
}

type GetUserReply struct {
	BaseReply
	UserId    string `json:"userId,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	AvatarUrl string `json:"avatarUrl,omitempty"`
	UserName  string `json:"userName,omitempty"`
}

type UserCountReply struct {
	BaseReply
	Count int64 `json:"count,omitempty"`
}
