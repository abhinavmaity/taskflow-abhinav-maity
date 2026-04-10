import { apiRequest } from "./client";
import type {
  AuthResponse,
  ProjectDetailResponse,
  ProjectListResponse,
  StatsResponse,
  TaskListResponse
} from "./types";

type AuthInput = {
  email: string;
  password: string;
};

type RegisterInput = AuthInput & {
  name: string;
};

type CreateProjectInput = {
  name: string;
  description?: string;
};

type ListProjectsInput = {
  page: number;
  limit: number;
};

type ListTasksInput = {
  page: number;
  limit: number;
  status?: string;
  assignee?: string;
};

export function login(payload: AuthInput) {
  return apiRequest<AuthResponse>("/auth/login", {
    method: "POST",
    body: payload
  });
}

export function register(payload: RegisterInput) {
  return apiRequest<AuthResponse>("/auth/register", {
    method: "POST",
    body: payload
  });
}

export function listProjects(token: string, input: ListProjectsInput) {
  const query = new URLSearchParams({
    page: String(input.page),
    limit: String(input.limit)
  });

  return apiRequest<ProjectListResponse>(`/projects?${query.toString()}`, {
    token
  });
}

export function createProject(token: string, payload: CreateProjectInput) {
  return apiRequest(`/projects`, {
    method: "POST",
    token,
    body: payload
  });
}

export function getProjectDetail(token: string, projectId: string) {
  return apiRequest<ProjectDetailResponse>(`/projects/${projectId}`, { token });
}

export function getProjectStats(token: string, projectId: string) {
  return apiRequest<StatsResponse>(`/projects/${projectId}/stats`, { token });
}

export function listProjectTasks(token: string, projectId: string, input: ListTasksInput) {
  const query = new URLSearchParams({
    page: String(input.page),
    limit: String(input.limit)
  });
  if (input.status) {
    query.set("status", input.status);
  }
  if (input.assignee) {
    query.set("assignee", input.assignee);
  }

  return apiRequest<TaskListResponse>(`/projects/${projectId}/tasks?${query.toString()}`, { token });
}
