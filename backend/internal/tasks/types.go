package tasks

import (
	"strings"
	"time"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/httpx"
)

var validStatuses = map[string]struct{}{
	"todo":        {},
	"in_progress": {},
	"done":        {},
}

var validPriorities = map[string]struct{}{
	"low":    {},
	"medium": {},
	"high":   {},
}

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	ProjectID   string    `json:"project_id"`
	AssigneeID  *string   `json:"assignee_id,omitempty"`
	CreatedBy   string    `json:"created_by"`
	DueDate     *string   `json:"due_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskListResponse struct {
	Tasks      []Task               `json:"tasks"`
	Pagination httpx.PaginationMeta `json:"pagination"`
}

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func (r CreateTaskRequest) Validate() map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(r.Title) == "" {
		fields["title"] = "is required"
	}
	if r.Status != "" {
		if _, ok := validStatuses[r.Status]; !ok {
			fields["status"] = "must be one of: todo, in_progress, done"
		}
	}
	if r.Priority != "" {
		if _, ok := validPriorities[r.Priority]; !ok {
			fields["priority"] = "must be one of: low, medium, high"
		}
	}
	return fields
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func (r UpdateTaskRequest) Validate() map[string]string {
	fields := map[string]string{}
	if r.Title == nil && r.Description == nil && r.Status == nil && r.Priority == nil && r.AssigneeID == nil && r.DueDate == nil {
		fields["body"] = "must include at least one field"
		return fields
	}
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		fields["title"] = "cannot be empty"
	}
	if r.Status != nil {
		if _, ok := validStatuses[*r.Status]; !ok {
			fields["status"] = "must be one of: todo, in_progress, done"
		}
	}
	if r.Priority != nil {
		if _, ok := validPriorities[*r.Priority]; !ok {
			fields["priority"] = "must be one of: low, medium, high"
		}
	}
	return fields
}

type TaskFilters struct {
	Status   string
	Assignee string
}

type TaskPermissionContext struct {
	TaskID       string
	ProjectID    string
	CreatedBy    string
	ProjectOwner string
}
