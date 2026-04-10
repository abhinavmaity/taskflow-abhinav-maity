package projects

import (
	"strings"
	"time"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/httpx"
)

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Assignee struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TaskSummary struct {
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

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type AssigneeCount struct {
	AssigneeID   string `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
	Count        int    `json:"count"`
}

type StatsResponse struct {
	ByStatus   []StatusCount   `json:"by_status"`
	ByAssignee []AssigneeCount `json:"by_assignee"`
}

type ListProjectsResponse struct {
	Projects   []Project            `json:"projects"`
	Pagination httpx.PaginationMeta `json:"pagination"`
}

type ProjectDetailResponse struct {
	Project            Project       `json:"project"`
	Tasks              []TaskSummary `json:"tasks"`
	AvailableAssignees []Assignee    `json:"available_assignees"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (r CreateProjectRequest) Validate() map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(r.Name) == "" {
		fields["name"] = "is required"
	}
	return fields
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (r UpdateProjectRequest) Validate() map[string]string {
	fields := map[string]string{}
	if r.Name == nil && r.Description == nil {
		fields["body"] = "must include at least one field"
		return fields
	}
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		fields["name"] = "cannot be empty"
	}
	return fields
}
