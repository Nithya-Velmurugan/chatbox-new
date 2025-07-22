// Package service provides an optimized concurrent chat service.
package service

import (
	"context"
	"errors"
	"sync"
	"time"

	errcom "chatbox/error"
	"chatbox/model"

	"golang.org/x/time/rate"
)

type Client struct {
	ID          string
	Ch          chan string
	LastSeen    time.Time
	RateLimiter *rate.Limiter
}

type ChatService interface {
	Join(ctx context.Context, req model.JoinRequest) (*model.JoinResponse, error)
	SendMessage(ctx context.Context, req model.SendMessageRequest) (*model.SendMessageResponse, error)
	Leave(ctx context.Context, req model.LeaveRequest) (*model.LeaveResponse, error)
	GetMessage(ctx context.Context, req model.MessageRequest) (*model.MessageResponse, error)
}

type chatService struct {
	mu      sync.RWMutex
	streams map[string]*Client
}

func NewChatService() ChatService {
	s := &chatService{
		streams: make(map[string]*Client),
	}
	s.startCleanupLoop()
	return s
}

// Background cleanup: remove users idle for 5 minutes
func (s *chatService) startCleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			s.mu.Lock()
			for id, client := range s.streams {
				if time.Since(client.LastSeen) > 5*time.Minute {
					close(client.Ch)
					delete(s.streams, id)
				}
			}
			s.mu.Unlock()
		}
	}()
}

func (s *chatService) Join(ctx context.Context, req model.JoinRequest) (*model.JoinResponse, error) {
	if req.ID == "" {
		return nil, errcom.NewCustomError("ERR_MISSING_USER_ID", errors.New("user ID is required"))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.streams[req.ID]; exists {
		return nil, errcom.NewCustomError("ERR_ALREADY_JOINED", errors.New("user already joined"))
	}

	s.streams[req.ID] = &Client{
		ID:          req.ID,
		Ch:          make(chan string, 10),
		LastSeen:    time.Now(),
		RateLimiter: rate.NewLimiter(1, 5),
	}

	return &model.JoinResponse{
		Success: true,
		Message: "User joined successfully",
	}, nil
}

func (s *chatService) SendMessage(ctx context.Context, req model.SendMessageRequest) (*model.SendMessageResponse, error) {
	if req.From == "" || req.Message == "" {
		return nil, errcom.NewCustomError("ERR_MISSING_FIELD", errors.New("From and Message are required"))
	}

	if len(req.Message) > 500 {
		return nil, errcom.NewCustomError("ERR_MESSAGE_TOO_LONG", errors.New("message must be under 500 characters"))
	}

	s.mu.RLock()
	sender, exists := s.streams[req.From]
	if !exists {
		s.mu.RUnlock()
		return nil, errcom.NewCustomError("ERR_SENDER_NOT_FOUND", errors.New("sender not connected"))
	}
	if !sender.RateLimiter.Allow() {
		s.mu.RUnlock()
		return nil, errcom.NewCustomError("ERR_RATE_LIMIT", errors.New("too many messages"))
	}

	message := req.From + ": " + req.Message
	sentCount := 0
	for id, client := range s.streams {
		if id == req.From {
			continue
		}
		go func(c *Client) {
			select {
			case c.Ch <- message:
				// sent
			default:
				// drop if channel full
			}
		}(client)
		sentCount++
	}
	s.mu.RUnlock()

	if sentCount == 0 {
		return nil, errcom.NewCustomError("ERR_NO_RECEIVERS", errors.New("no clients received the message"))
	}

	return &model.SendMessageResponse{
		Success: true,
		Message: "Message broadcasted to clients",
	}, nil
}

func (s *chatService) Leave(ctx context.Context, req model.LeaveRequest) (*model.LeaveResponse, error) {
	if req.ID == "" {
		return nil, errcom.NewCustomError("ERR_MISSING_USER_ID", errors.New("user ID is required"))
	}

	s.mu.Lock()
	client, exists := s.streams[req.ID]
	if !exists {
		s.mu.Unlock()
		return nil, errcom.NewCustomError("ERR_USER_NOT_FOUND", errors.New("user not connected"))
	}
	delete(s.streams, req.ID)
	s.mu.Unlock()

	close(client.Ch)

	return &model.LeaveResponse{
		Success: true,
		Message: "User disconnected successfully",
	}, nil
}

func (s *chatService) GetMessage(ctx context.Context, req model.MessageRequest) (*model.MessageResponse, error) {
	if req.ID == "" {
		return nil, errcom.NewCustomError("ERR_MISSING_USER_ID", errors.New("user ID is required"))
	}

	s.mu.RLock()
	client, exists := s.streams[req.ID]
	s.mu.RUnlock()

	if !exists {
		return nil, errcom.NewCustomError("ERR_USER_NOT_FOUND", errors.New("user not connected"))
	}

	client.LastSeen = time.Now()

	select {
	case msg, ok := <-client.Ch:
		if !ok {
			return nil, errcom.NewCustomError("ERR_USER_DISCONNECTED", errors.New("user stream closed"))
		}
		return &model.MessageResponse{Message: msg}, nil
	case <-time.After(10 * time.Second):
		return nil, errcom.NewCustomError("ERR_NO_MESSAGES", errors.New("no messages received"))
	}
}
