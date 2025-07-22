package model

type JoinRequest struct {
	ID string `json:"id"`
}

type SendMessageRequest struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

type LeaveRequest struct {
	ID string `json:"id"`
}

type MessageRequest struct {
	ID string `json:"id"`
}

type JoinResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LeaveResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
