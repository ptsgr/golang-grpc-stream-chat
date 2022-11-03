package grpc

const (
	ErrGroupExists    = "Group already exists"
	ErrGroupNotFound  = "Connot found group"
	ErrNotGroupMember = "You not a group member"
)

const (
	JoinGroupChatCommand   = "join-group"
	LeftGroupChatCommand   = "left-group"
	CreateGroupChatCommand = "create-group"
)
