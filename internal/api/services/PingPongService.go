// internal/api/services/PingService.go
package services

type PingService struct {
	// Dependencies would go here
}

func NewPingService() *PingService {
	return &PingService{}
}

func (s *PingService) GetPing() string {
	return "Pong"
}
