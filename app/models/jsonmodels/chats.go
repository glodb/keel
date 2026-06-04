package jsonmodels

type SendMessageRequest struct {
	ChatId  string   `json:"chatId" validate:"required"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

type SendDMRequest struct {
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

type AcceptInvitationRequest struct {
	InvitationToken string `json:"invitationToken,omitempty"`
	MessageId       string `json:"messageId,omitempty"`
	PaymentMethodId string `json:"paymentMethodId,omitempty"`
}

type ChatNameAndAvatar struct {
	UserName string `json:"userName"`
	Avatar   string `json:"avatar"`
}
