package rbac

import (
	"context"
	"fmt"
)

// Service handles Role-Based Access Control
type Service struct {
	roles map[string][]string
}

// NewService creates a new RBAC service
func NewService() *Service {
	return &Service{
		roles: make(map[string][]string),
	}
}

// HasPermission checks if a user has the required permission
func (s *Service) HasPermission(ctx context.Context, userRole string, requiredPermission string) bool {
	permissions, exists := s.roles[userRole]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == requiredPermission {
			return true
		}
	}

	return false
}

// AddRole adds a new role with its permissions
func (s *Service) AddRole(role string, permissions []string) error {
	if _, exists := s.roles[role]; exists {
		return fmt.Errorf("role %s already exists", role)
	}

	s.roles[role] = permissions
	return nil
}
