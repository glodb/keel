package paginationmodels

type ChatRetrievalRequest struct {
	BasePagination
	UserId         string `json:"userId"`
	ConversationId string `json:"conversationId" validate:"required" field:"conversationId"`
}

type DoctorChatRetrievalRequest struct {
	BasePagination
	PatientId string `json:"patientId" validate:"required" field:"patientId"`
}

type CallStatusRequest struct {
	ConversationId string `json:"conversationId" validate:"required" field:"conversationId"`
	MessageId      string `json:"messageId" validate:"required" field:"messageId"`
	CallStatus     int    `json:"callStatus" validate:"required" field:"callStatus"`
}
