package keelmodels

import (
	"encoding/json"
	"fmt"
)

// Chat represents a conversation/chat room
// One Chat object per conversation (event chat, direct message, or group chat)
type Chat struct {
	ChatId                string            `json:"chatId" bson:"chatId"` // Unique chat identifier
	Type                  string            `json:"type" bson:"type"`
	UserAvatars           map[string]string `json:"userAvatars,omitempty" bson:"userAvatars,omitempty"`       // "event" | "direct" | "group"
	EventId               string            `json:"eventId,omitempty" bson:"eventId,omitempty"`               // Only for event chats
	EventStartTime        int64             `json:"eventStartTime,omitempty" bson:"eventStartTime,omitempty"` // Only for event chats
	EventEndTime          int64             `json:"eventEndTime,omitempty" bson:"eventEndTime,omitempty"`
	EventName             string            `json:"eventName,omitempty" bson:"eventName,omitempty"` // Only for event chats
	PalId                 string            `json:"palId,omitempty" bson:"palId,omitempty"`         // Only for event chats (references Pal.Id)
	Name                  string            `json:"name,omitempty" bson:"name,omitempty"`           // Chat name (for group chats)
	Description           string            `json:"description,omitempty" bson:"description,omitempty"`
	CreatedBy             string            `json:"createdBy" bson:"createdBy"` // UserId who created the chat
	EnrollmentClosingTime int64             `json:"enrollmentClosingTime,omitempty" bson:"enrollmentClosingTime,omitempty"`
	LastMessageId         string            `json:"lastMessageId,omitempty" bson:"lastMessageId,omitempty"` // For quick access
	LastMessageAt         int64             `json:"lastMessageAt,omitempty" bson:"lastMessageAt,omitempty"` // For sorting chats
	IsActive              bool              `json:"isActive" bson:"isActive"`                               // Whether chat is active
	CreatedAt             int64             `json:"createdAt" bson:"createdAt"`
	UpdatedAt             int64             `json:"updatedAt" bson:"updatedAt"`
}

// ChatParticipant represents a user's participation in a chat
// One entry per user per chat - tracks unread count and participation status
type ChatParticipant struct {
	ParticipantId     string `json:"participantId" bson:"participantId"`                             // Unique: chatId + userId
	ChatId            string `json:"chatId" bson:"chatId"`                                           // Reference to Chat
	UserId            string `json:"userId" bson:"userId"`                                           // User participating
	UnreadCount       int    `json:"unreadCount" bson:"unreadCount"`                                 // Number of unread messages
	LastReadAt        int64  `json:"lastReadAt" bson:"lastReadAt"`                                   // Timestamp of last read message
	LastReadMessageId string `json:"lastReadMessageId,omitempty" bson:"lastReadMessageId,omitempty"` // Last message user read
	IsMuted           bool   `json:"isMuted" bson:"isMuted"`                                         // Whether user muted this chat
	IsArchived        bool   `json:"isArchived" bson:"isArchived"`                                   // Whether user archived this chat
	JoinedAt          int64  `json:"joinedAt" bson:"joinedAt"`                                       // When user joined
	LeftAt            int64  `json:"leftAt,omitempty" bson:"leftAt,omitempty"`                       // When user left (0 if still active)
	// Cached participant info (for direct chats - shows the other participant)
	// For direct chats, this stores the OTHER participant's info for quick access
	OtherParticipantId     string `json:"otherParticipantId,omitempty" bson:"otherParticipantId,omitempty"`         // Other user's userId (for direct chats)
	OtherParticipantName   string `json:"otherParticipantName,omitempty" bson:"otherParticipantName,omitempty"`     // Other user's name (cached)
	OtherParticipantAvatar string `json:"otherParticipantAvatar,omitempty" bson:"otherParticipantAvatar,omitempty"` // Other user's avatar (cached)
	CreatedAt              int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt              int64  `json:"updatedAt" bson:"updatedAt"`
}

// ChatMessage represents a single message in a chat
type ChatMessage struct {
	MessageId    string                 `json:"messageId" bson:"messageId"`   // Unique message identifier
	ChatId       string                 `json:"chatId" bson:"chatId"`         // Reference to Chat
	Type         string                 `json:"type" bson:"type"`             // "invitation" | "text" | "system" | "image" | etc.
	SenderId     string                 `json:"senderId" bson:"senderId"`     // UserId who sent the message
	SenderName   string                 `json:"senderName" bson:"senderName"` // Cached sender name for quick display
	SenderAvatar string                 `json:"senderAvatar,omitempty" bson:"senderAvatar,omitempty"`
	Content      string                 `json:"content" bson:"content"`                       // Message content
	Metadata     map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"` // For invitation data, attachments, etc.
	Images       []string               `json:"images,omitempty" bson:"images,omitempty"`     // Images in the message
	// Targeted message fields (for invitations visible only to specific users)
	IsTargeted bool     `json:"isTargeted" bson:"isTargeted"`                     // If true, only recipients can see this message
	Recipients []string `json:"recipients,omitempty" bson:"recipients,omitempty"` // UserIds who should see this message (empty = all participants)
	IsRead     bool     `json:"isRead" bson:"isRead"`                             // Deprecated - use ChatParticipant instead
	IsDeleted  bool     `json:"isDeleted" bson:"isDeleted"`                       // Soft delete
	DeletedAt  int64    `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	CreatedAt  int64    `json:"createdAt" bson:"createdAt"`
	UpdatedAt  int64    `json:"updatedAt" bson:"updatedAt"`
}

type ChatUserData struct {
	ChatMessage          ChatMessage `json:"chatMessage" bson:"chatMessage"`
	TotalUnreadCount     int         `json:"totalUnreadCount" bson:"totalUnreadCount"`
	MainTotalUnreadCount int         `json:"mainTotalUnreadCount" bson:"mainTotalUnreadCount"`
}

// EventInvitation represents an invitation sent to a user for an event
// Used for verification and tracking invitation status
type EventInvitation struct {
	InvitationId    string `json:"invitationId" bson:"invitationId"`                 // Unique invitation identifier
	EventId         string `json:"eventId" bson:"eventId"`                           // Reference to Event
	PalId           string `json:"palId" bson:"palId"`                               // Reference to Pal (specific event instance)
	ChatId          string `json:"chatId" bson:"chatId"`                             // Reference to Chat
	MessageId       string `json:"messageId" bson:"messageId"`                       // Reference to ChatMessage (invitation message)
	InvitedUserId   string `json:"invitedUserId" bson:"invitedUserId"`               // User who was invited
	InvitedBy       string `json:"invitedBy" bson:"invitedBy"`                       // UserId who sent the invitation (organizer)
	Status          string `json:"status" bson:"status"`                             // "pending" | "accepted" | "declined" | "expired" | "cancelled"
	InvitationToken string `json:"invitationToken" bson:"invitationToken"`           // Token for verification
	ExpiresAt       int64  `json:"expiresAt,omitempty" bson:"expiresAt,omitempty"`   // When invitation expires
	AcceptedAt      int64  `json:"acceptedAt,omitempty" bson:"acceptedAt,omitempty"` // When user accepted
	DeclinedAt      int64  `json:"declinedAt,omitempty" bson:"declinedAt,omitempty"` // When user declined
	CreatedAt       int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt       int64  `json:"updatedAt" bson:"updatedAt"`
}

type ChatSupportMessage struct {
	UserId         string `json:"userId" bson:"userId"`
	MonthStart     int64  `json:"monthStart" bson:"monthStart"`
	SupportCredits int    `json:"supportCredits" bson:"supportCredits"`
	CreatedAt      int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt" bson:"updatedAt"`
}

// GetQuery returns the query for finding a Chat record
func (c *Chat) GetQuery() map[string]interface{} {
	return map[string]interface{}{"chatId": c.ChatId}
}

// GetQuery returns the query for finding a ChatParticipant record
func (cp *ChatParticipant) GetQuery() map[string]interface{} {
	return map[string]interface{}{"participantId": cp.ParticipantId}
}

// GetQuery returns the query for finding a ChatMessage record
func (cm *ChatMessage) GetQuery() map[string]interface{} {
	return map[string]interface{}{"messageId": cm.MessageId}
}

func (cm *ChatMessage) GetMapData() map[string]interface{} {
	// FCM requires all data values to be strings
	// Convert arrays, booleans, and numbers to string representations

	// Convert images array to JSON string
	imagesJSON := "[]"
	if len(cm.Images) > 0 {
		if bytes, err := json.Marshal(cm.Images); err == nil {
			imagesJSON = string(bytes)
		}
	}

	// Convert recipients array to JSON string
	recipientsJSON := "[]"
	if len(cm.Recipients) > 0 {
		if bytes, err := json.Marshal(cm.Recipients); err == nil {
			recipientsJSON = string(bytes)
		}
	}

	// Convert metadata to JSON string
	metadataJSON := "{}"
	if cm.Metadata != nil {
		if bytes, err := json.Marshal(cm.Metadata); err == nil {
			metadataJSON = string(bytes)
		}
	}

	return map[string]interface{}{
		"messageId":    cm.MessageId,
		"chatId":       cm.ChatId,
		"type":         cm.Type,
		"senderId":     cm.SenderId,
		"senderName":   cm.SenderName,
		"senderAvatar": cm.SenderAvatar,
		"content":      cm.Content,
		"metadata":     metadataJSON,                     // JSON string
		"images":       imagesJSON,                       // JSON string
		"isTargeted":   fmt.Sprintf("%t", cm.IsTargeted), // "true" or "false"
		"recipients":   recipientsJSON,                   // JSON string
		"isRead":       fmt.Sprintf("%t", cm.IsRead),     // "true" or "false"
		"isDeleted":    fmt.Sprintf("%t", cm.IsDeleted),  // "true" or "false"
		"deletedAt":    fmt.Sprintf("%d", cm.DeletedAt),  // number as string
		"createdAt":    fmt.Sprintf("%d", cm.CreatedAt),  // number as string
		"updatedAt":    fmt.Sprintf("%d", cm.UpdatedAt),  // number as string
	}
}

func (cud *ChatUserData) GetMapData() map[string]interface{} {
	// Get the chat message data (already flattened and stringified)
	chatData := cud.ChatMessage.GetMapData()

	// Add the unread counts as strings
	chatData["totalUnreadCount"] = fmt.Sprintf("%d", cud.TotalUnreadCount)
	chatData["mainTotalUnreadCount"] = fmt.Sprintf("%d", cud.MainTotalUnreadCount)

	return chatData
}

// GetQuery returns the query for finding an EventInvitation record
func (ei *EventInvitation) GetQuery() map[string]interface{} {
	return map[string]interface{}{"invitationId": ei.InvitationId}
}
