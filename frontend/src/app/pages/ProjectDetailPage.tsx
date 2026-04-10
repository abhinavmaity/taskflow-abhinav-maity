import {
  Alert,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Pagination,
  Select,
  Stack,
  TextField,
  Typography
} from "@mui/material";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { ApiError, toErrorMessage } from "../../api/client";
import {
  createTask,
  getProjectDetail,
  getProjectStats,
  listProjectTasks,
  updateTask,
  type UpdateTaskInput,
  type UpsertTaskInput
} from "../../api/taskflowApi";
import type { TaskListResponse, TaskPriority, TaskStatus, TaskSummary } from "../../api/types";
import { useAuth } from "../../auth/AuthProvider";

const TASK_LIMIT = 8;

type TaskDialogState = {
  open: boolean;
  mode: "create" | "edit";
  taskId: string | null;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  assigneeId: string;
  dueDate: string;
  error: string;
};

const defaultTaskDialogState: TaskDialogState = {
  open: false,
  mode: "create",
  taskId: null,
  title: "",
  description: "",
  status: "todo",
  priority: "medium",
  assigneeId: "",
  dueDate: "",
  error: ""
};

export function ProjectDetailPage() {
  const queryClient = useQueryClient();
  const { id } = useParams<{ id: string }>();
  const { session, signOut } = useAuth();
  const token = session?.token ?? "";

  const [statusFilter, setStatusFilter] = useState("");
  const [assigneeFilter, setAssigneeFilter] = useState("");
  const [page, setPage] = useState(1);
  const [taskDialog, setTaskDialog] = useState<TaskDialogState>(defaultTaskDialogState);

  const detailQuery = useQuery({
    queryKey: ["project-detail", id],
    queryFn: () => getProjectDetail(token, id ?? ""),
    enabled: Boolean(token && id),
    retry: false
  });

  const statsQuery = useQuery({
    queryKey: ["project-stats", id],
    queryFn: () => getProjectStats(token, id ?? ""),
    enabled: Boolean(token && id),
    retry: false
  });

  const tasksQueryKey = ["project-tasks", id, page, statusFilter, assigneeFilter] as const;
  const tasksQuery = useQuery({
    queryKey: tasksQueryKey,
    queryFn: () =>
      listProjectTasks(token, id ?? "", {
        page,
        limit: TASK_LIMIT,
        status: statusFilter || undefined,
        assignee: assigneeFilter || undefined
      }),
    enabled: Boolean(token && id),
    retry: false
  });

  const statusMutation = useMutation({
    mutationFn: ({ taskId, nextStatus }: { taskId: string; nextStatus: TaskStatus }) =>
      updateTask(token, taskId, { status: nextStatus }),
    onMutate: async ({ taskId, nextStatus }) => {
      await queryClient.cancelQueries({ queryKey: ["project-tasks", id] });
      const previous = queryClient.getQueryData<TaskListResponse>(tasksQueryKey);

      if (previous) {
        let nextTasks = previous.tasks.map((task) =>
          task.id === taskId ? { ...task, status: nextStatus, updated_at: new Date().toISOString() } : task
        );

        if (statusFilter && statusFilter !== nextStatus) {
          nextTasks = nextTasks.filter((task) => task.id !== taskId);
        }

        queryClient.setQueryData<TaskListResponse>(tasksQueryKey, {
          ...previous,
          tasks: nextTasks
        });
      }

      return { previous };
    },
    onError: (error, _variables, context) => {
      if (error instanceof ApiError && error.status === 401) {
        signOut();
        return;
      }
      if (context?.previous) {
        queryClient.setQueryData(tasksQueryKey, context.previous);
      }
    },
    onSettled: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["project-tasks", id] }),
        queryClient.invalidateQueries({ queryKey: ["project-detail", id] }),
        queryClient.invalidateQueries({ queryKey: ["project-stats", id] })
      ]);
    }
  });

  const taskDialogMutation = useMutation({
    mutationFn: async () => {
      if (!id) {
        throw new Error("Missing project id");
      }

      const basePayload = {
        title: taskDialog.title.trim(),
        description: taskDialog.description.trim() || undefined,
        status: taskDialog.status,
        priority: taskDialog.priority,
        due_date: taskDialog.dueDate || undefined
      };

      if (taskDialog.mode === "create") {
        const payload: UpsertTaskInput = {
          ...basePayload,
          assignee_id: taskDialog.assigneeId || undefined
        };
        return createTask(token, id, payload);
      }

      if (!taskDialog.taskId) {
        throw new Error("Missing task id for update");
      }

      const payload: UpdateTaskInput = {
        ...basePayload,
        assignee_id: taskDialog.assigneeId
      };
      return updateTask(token, taskDialog.taskId, payload);
    },
    onSuccess: async () => {
      setTaskDialog(defaultTaskDialogState);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["project-tasks", id] }),
        queryClient.invalidateQueries({ queryKey: ["project-detail", id] }),
        queryClient.invalidateQueries({ queryKey: ["project-stats", id] })
      ]);
    },
    onError: (error) => {
      if (error instanceof ApiError && error.status === 401) {
        signOut();
        return;
      }
      setTaskDialog((prev) => ({ ...prev, error: toErrorMessage(error) }));
    }
  });

  useEffect(() => {
    if (detailQuery.error instanceof ApiError && detailQuery.error.status === 401) {
      signOut();
    }
  }, [detailQuery.error, signOut]);

  useEffect(() => {
    if (statsQuery.error instanceof ApiError && statsQuery.error.status === 401) {
      signOut();
    }
  }, [statsQuery.error, signOut]);

  useEffect(() => {
    if (tasksQuery.error instanceof ApiError && tasksQuery.error.status === 401) {
      signOut();
    }
  }, [tasksQuery.error, signOut]);

  useEffect(() => {
    setPage(1);
  }, [statusFilter, assigneeFilter]);

  const assignees = useMemo(() => detailQuery.data?.available_assignees ?? [], [detailQuery.data]);
  const taskItems = tasksQuery.data?.tasks ?? [];
  const totalPages = tasksQuery.data?.pagination.total_pages ?? 0;
  const statsColors: Record<string, string> = {
    done: "#1d4ed8",
    in_progress: "#f4c430",
    todo: "#e63946"
  };

  if (!id) {
    return <Alert severity="error">Invalid project id.</Alert>;
  }

  if (detailQuery.isLoading) {
    return (
      <Card>
        <CardContent>
          <Stack alignItems="center" direction="row" spacing={1.5}>
            <CircularProgress size={18} />
            <Typography color="text.secondary">Loading project...</Typography>
          </Stack>
        </CardContent>
      </Card>
    );
  }

  if (detailQuery.isError) {
    return <Alert severity="error">{toErrorMessage(detailQuery.error)}</Alert>;
  }

  const project = detailQuery.data?.project;
  if (!project) {
    return <Alert severity="error">Project could not be loaded.</Alert>;
  }

  const openCreateTaskDialog = () => {
    setTaskDialog({
      ...defaultTaskDialogState,
      open: true,
      mode: "create"
    });
  };

  const openEditTaskDialog = (task: TaskSummary) => {
    setTaskDialog({
      open: true,
      mode: "edit",
      taskId: task.id,
      title: task.title,
      description: task.description ?? "",
      status: task.status,
      priority: task.priority,
      assigneeId: task.assignee_id ?? "",
      dueDate: task.due_date ?? "",
      error: ""
    });
  };

  return (
    <Stack spacing={3}>
      <div>
        <Typography className="tf-hero-label">Project</Typography>
        <Typography sx={{ fontWeight: 700, marginTop: 1 }} variant="h3">
          {project.name}
        </Typography>
        <Typography color="text.secondary">{project.description?.trim() || "No description provided."}</Typography>
      </div>

      <Grid container spacing={2}>
        <Grid item md={8} xs={12}>
          <Card>
            <CardContent>
              <Stack spacing={2}>
                <Stack direction={{ sm: "row", xs: "column" }} spacing={2}>
                  <FormControl fullWidth size="small">
                    <InputLabel id="status-filter-label">Status</InputLabel>
                    <Select
                      label="Status"
                      labelId="status-filter-label"
                      onChange={(event) => setStatusFilter(event.target.value)}
                      value={statusFilter}
                    >
                      <MenuItem value="">All</MenuItem>
                      <MenuItem value="todo">Todo</MenuItem>
                      <MenuItem value="in_progress">In Progress</MenuItem>
                      <MenuItem value="done">Done</MenuItem>
                    </Select>
                  </FormControl>

                  <FormControl fullWidth size="small">
                    <InputLabel id="assignee-filter-label">Assignee</InputLabel>
                    <Select
                      label="Assignee"
                      labelId="assignee-filter-label"
                      onChange={(event) => setAssigneeFilter(event.target.value)}
                      value={assigneeFilter}
                    >
                      <MenuItem value="">All</MenuItem>
                      {assignees.map((assignee) => (
                        <MenuItem key={assignee.id} value={assignee.id}>
                          {assignee.name}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>

                  <Button onClick={openCreateTaskDialog} variant="contained">
                    New Task
                  </Button>
                </Stack>

                {tasksQuery.isLoading ? (
                  <Stack alignItems="center" direction="row" spacing={1.5}>
                    <CircularProgress size={16} />
                    <Typography color="text.secondary">Loading tasks...</Typography>
                  </Stack>
                ) : null}

                {tasksQuery.isError ? <Alert severity="error">{toErrorMessage(tasksQuery.error)}</Alert> : null}

                {!tasksQuery.isLoading && !tasksQuery.isError && taskItems.length === 0 ? (
                  <Typography color="text.secondary">No tasks match the current filters.</Typography>
                ) : null}

                {taskItems.map((task) => (
                  <Card key={task.id} variant="outlined">
                    <CardContent>
                      <Stack spacing={1.5}>
                        <Stack direction="row" justifyContent="space-between" spacing={2}>
                          <Typography sx={{ fontWeight: 600 }}>{task.title}</Typography>
                          <Stack direction="row" spacing={1}>
                            <Chip color="default" label={task.priority} size="small" />
                            <Chip
                              color={task.status === "done" ? "secondary" : task.status === "in_progress" ? "warning" : "primary"}
                              label={task.status}
                              size="small"
                            />
                          </Stack>
                        </Stack>

                        <Typography color="text.secondary" variant="body2">
                          {task.description?.trim() || "No description."}
                        </Typography>

                        <Stack direction={{ sm: "row", xs: "column" }} spacing={1.5}>
                          <FormControl size="small" sx={{ minWidth: 170 }}>
                            <InputLabel id={`task-status-${task.id}`}>Status</InputLabel>
                            <Select
                              label="Status"
                              labelId={`task-status-${task.id}`}
                              onChange={(event) =>
                                statusMutation.mutate({
                                  taskId: task.id,
                                  nextStatus: event.target.value as TaskStatus
                                })
                              }
                              size="small"
                              value={task.status}
                            >
                              <MenuItem value="todo">Todo</MenuItem>
                              <MenuItem value="in_progress">In Progress</MenuItem>
                              <MenuItem value="done">Done</MenuItem>
                            </Select>
                          </FormControl>

                          <Button onClick={() => openEditTaskDialog(task)} size="small" variant="outlined">
                            Edit
                          </Button>
                        </Stack>
                      </Stack>
                    </CardContent>
                  </Card>
                ))}

                {totalPages > 1 ? (
                  <Stack alignItems="center" sx={{ paddingTop: 1 }}>
                    <Pagination
                      count={totalPages}
                      onChange={(_event, nextPage) => setPage(nextPage)}
                      page={page}
                      shape="rounded"
                    />
                  </Stack>
                ) : null}
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid item md={4} xs={12}>
          <Card>
            <CardContent>
              <Typography sx={{ fontWeight: 600 }} variant="h6">
                Stats
              </Typography>
              <Stack spacing={1.25} sx={{ marginTop: 1.5 }}>
                {statsQuery.isLoading ? (
                  <Stack alignItems="center" direction="row" spacing={1.5}>
                    <CircularProgress size={16} />
                    <Typography color="text.secondary">Loading stats...</Typography>
                  </Stack>
                ) : null}

                {statsQuery.isError ? <Alert severity="error">{toErrorMessage(statsQuery.error)}</Alert> : null}

                {statsQuery.data?.by_status.map((entry) => (
                  <Stack
                    direction="row"
                    justifyContent="space-between"
                    key={entry.status}
                    sx={{
                      backgroundColor: statsColors[entry.status] ?? "#eeeeee",
                      border: "2px solid #000000",
                      paddingX: 1.2,
                      paddingY: 0.9
                    }}
                  >
                    <Typography
                      sx={{
                        color: entry.status === "in_progress" ? "#000000" : "#ffffff",
                        textTransform: "capitalize"
                      }}
                      variant="body2"
                    >
                      {entry.status.replace("_", " ")}
                    </Typography>
                    <Typography
                      sx={{
                        color: entry.status === "in_progress" ? "#000000" : "#ffffff",
                        fontWeight: 700
                      }}
                      variant="body2"
                    >
                      {entry.count}
                    </Typography>
                  </Stack>
                ))}
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => setTaskDialog(defaultTaskDialogState)}
        open={taskDialog.open}
      >
        <DialogTitle>{taskDialog.mode === "create" ? "Create Task" : "Edit Task"}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ marginTop: 1 }}>
            {taskDialog.error ? <Alert severity="error">{taskDialog.error}</Alert> : null}

            <TextField
              autoFocus
              label="Title"
              onChange={(event) => setTaskDialog((prev) => ({ ...prev, title: event.target.value }))}
              required
              value={taskDialog.title}
            />

            <TextField
              label="Description"
              minRows={3}
              multiline
              onChange={(event) => setTaskDialog((prev) => ({ ...prev, description: event.target.value }))}
              value={taskDialog.description}
            />

            <Grid container spacing={2}>
              <Grid item sm={6} xs={12}>
                <FormControl fullWidth>
                  <InputLabel id="task-dialog-status-label">Status</InputLabel>
                  <Select
                    label="Status"
                    labelId="task-dialog-status-label"
                    onChange={(event) =>
                      setTaskDialog((prev) => ({ ...prev, status: event.target.value as TaskStatus }))
                    }
                    value={taskDialog.status}
                  >
                    <MenuItem value="todo">Todo</MenuItem>
                    <MenuItem value="in_progress">In Progress</MenuItem>
                    <MenuItem value="done">Done</MenuItem>
                  </Select>
                </FormControl>
              </Grid>

              <Grid item sm={6} xs={12}>
                <FormControl fullWidth>
                  <InputLabel id="task-dialog-priority-label">Priority</InputLabel>
                  <Select
                    label="Priority"
                    labelId="task-dialog-priority-label"
                    onChange={(event) =>
                      setTaskDialog((prev) => ({ ...prev, priority: event.target.value as TaskPriority }))
                    }
                    value={taskDialog.priority}
                  >
                    <MenuItem value="low">Low</MenuItem>
                    <MenuItem value="medium">Medium</MenuItem>
                    <MenuItem value="high">High</MenuItem>
                  </Select>
                </FormControl>
              </Grid>

              <Grid item sm={6} xs={12}>
                <FormControl fullWidth>
                  <InputLabel id="task-dialog-assignee-label">Assignee</InputLabel>
                  <Select
                    label="Assignee"
                    labelId="task-dialog-assignee-label"
                    onChange={(event) => setTaskDialog((prev) => ({ ...prev, assigneeId: event.target.value }))}
                    value={taskDialog.assigneeId}
                  >
                    <MenuItem value="">Unassigned</MenuItem>
                    {assignees.map((assignee) => (
                      <MenuItem key={assignee.id} value={assignee.id}>
                        {assignee.name}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>

              <Grid item sm={6} xs={12}>
                <TextField
                  fullWidth
                  InputLabelProps={{ shrink: true }}
                  label="Due Date"
                  onChange={(event) => setTaskDialog((prev) => ({ ...prev, dueDate: event.target.value }))}
                  type="date"
                  value={taskDialog.dueDate}
                />
              </Grid>
            </Grid>
          </Stack>
        </DialogContent>
        <DialogActions sx={{ padding: 2 }}>
          <Button onClick={() => setTaskDialog(defaultTaskDialogState)} variant="text">
            Cancel
          </Button>
          <Button
            disabled={taskDialogMutation.isPending || !taskDialog.title.trim()}
            onClick={() => {
              setTaskDialog((prev) => ({ ...prev, error: "" }));
              taskDialogMutation.mutate();
            }}
            variant="contained"
          >
            {taskDialog.mode === "create" ? "Create" : "Save"}
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
