package auth

import "strings"

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r RegisterRequest) Validate() map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(r.Name) == "" {
		fields["name"] = "is required"
	}

	if strings.TrimSpace(r.Email) == "" {
		fields["email"] = "is required"
	} else if !isValidEmail(r.Email) {
		fields["email"] = "must be a valid email"
	}

	if strings.TrimSpace(r.Password) == "" {
		fields["password"] = "is required"
	} else if len(r.Password) < 8 {
		fields["password"] = "must be at least 8 characters"
	}

	return fields
}

func (r LoginRequest) Validate() map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(r.Email) == "" {
		fields["email"] = "is required"
	}

	if strings.TrimSpace(r.Password) == "" {
		fields["password"] = "is required"
	}

	return fields
}

func isValidEmail(value string) bool {
	value = strings.TrimSpace(value)
	return strings.Contains(value, "@") && !strings.HasPrefix(value, "@") && !strings.HasSuffix(value, "@")
}
