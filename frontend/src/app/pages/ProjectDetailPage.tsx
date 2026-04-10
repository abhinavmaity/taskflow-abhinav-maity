import {
  Alert,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Pagination,
  Select,
  Stack,
  Typography
} from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { ApiError, toErrorMessage } from "../../api/client";
import { getProjectDetail, getProjectStats, listProjectTasks } from "../../api/taskflowApi";
import { useAuth } from "../../auth/AuthProvider";

const TASK_LIMIT = 8;

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { session, signOut } = useAuth();
  const token = session?.token ?? "";

  const [statusFilter, setStatusFilter] = useState("");
  const [assigneeFilter, setAssigneeFilter] = useState("");
  const [page, setPage] = useState(1);

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

  const tasksQuery = useQuery({
    queryKey: ["project-tasks", id, page, statusFilter, assigneeFilter],
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

  return (
    <Stack spacing={3}>
      <div>
        <Typography sx={{ fontWeight: 700 }} variant="h4">
          {project.name}
        </Typography>
        <Typography color="text.secondary">
          {project.description?.trim() || "No description provided."}
        </Typography>
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
                      <Stack spacing={1}>
                        <Stack direction="row" justifyContent="space-between" spacing={2}>
                          <Typography sx={{ fontWeight: 600 }}>{task.title}</Typography>
                          <Stack direction="row" spacing={1}>
                            <Chip color="default" label={task.priority} size="small" />
                            <Chip
                              color={task.status === "done" ? "success" : task.status === "in_progress" ? "warning" : "default"}
                              label={task.status}
                              size="small"
                            />
                          </Stack>
                        </Stack>
                        <Typography color="text.secondary" variant="body2">
                          {task.description?.trim() || "No description."}
                        </Typography>
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
                  <Stack direction="row" justifyContent="space-between" key={entry.status}>
                    <Typography color="text.secondary" sx={{ textTransform: "capitalize" }} variant="body2">
                      {entry.status.replace("_", " ")}
                    </Typography>
                    <Typography sx={{ fontWeight: 600 }} variant="body2">
                      {entry.count}
                    </Typography>
                  </Stack>
                ))}
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Stack>
  );
}
