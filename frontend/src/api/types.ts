export type ApiValidationFields = Record<string, string>;

export type ApiErrorResponse = {
  error: string;
  fields?: ApiValidationFields;
};

export type AuthResponse = {
  token: string;
  user: {
    id: string;
    name: string;
    email: string;
    created_at: string;
  };
};

export type PaginationMeta = {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
};

export type Project = {
  id: string;
  name: string;
  description?: string;
  owner_id: string;
  created_at: string;
};

export type TaskSummary = {
  id: string;
  title: string;
  description?: string;
  status: "todo" | "in_progress" | "done";
  priority: "low" | "medium" | "high";
  project_id: string;
  assignee_id?: string;
  created_by: string;
  due_date?: string;
  created_at: string;
  updated_at: string;
};

export type Assignee = {
  id: string;
  name: string;
  email: string;
};

export type ProjectListResponse = {
  projects: Project[];
  pagination: PaginationMeta;
};

export type ProjectDetailResponse = {
  project: Project;
  tasks: TaskSummary[];
  available_assignees: Assignee[];
};

export type TaskListResponse = {
  tasks: TaskSummary[];
  pagination: PaginationMeta;
};

export type StatsResponse = {
  by_status: Array<{
    status: string;
    count: number;
  }>;
  by_assignee: Array<{
    assignee_id: string;
    assignee_name: string;
    count: number;
  }>;
};
