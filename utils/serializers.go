package utils

import "hyperpage/models"

func SerializeChatRoomWithDetails(room models.ChatRoom) map[string]interface{} {
	roomMap := map[string]interface{}{
		"id":           room.ID,
		"name":         room.Name,
		"version":      room.Version,
		"created_at":   room.CreatedAt,
		"bumped_at":    room.BumpedAt,
		"member_count": len(room.Members),
	}

	if room.LastMessage != nil {
		roomMap["last_message"] = SerializeChatMessage(*room.LastMessage)
	}

	membersSerialized := make([]map[string]interface{}, 0, len(room.Members))
	for _, member := range room.Members {
		membersSerialized = append(membersSerialized, SerializeChatRoomMember(member))
	}
	roomMap["members"] = membersSerialized

	return roomMap
}

func SerializeChatMessage(message models.ChatMessage) map[string]interface{} {
	return map[string]interface{}{
		"id":         message.ID,
		"content":    message.Content,
		"user_id":    message.UserID.String(),
		"room_id":    message.RoomID,
		"is_edited":  message.IsEdited,
		"created_at": message.CreatedAt,
		"is_deleted": message.IsDeleted,
	}
}

func SerializeChatRoomMember(member models.ChatRoomMember) map[string]interface{} {
	return map[string]interface{}{
		"id":            member.ID,
		"room_id":       member.RoomID,
		"user_id":       member.UserID.String(),
		"is_subscribed": member.IsSubscribed,
		"is_new":        member.IsNew,
		"joined_at":     member.JoinedAt,
	}
}
