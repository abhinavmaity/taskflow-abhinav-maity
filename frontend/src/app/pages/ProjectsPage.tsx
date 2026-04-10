import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Pagination,
  Stack,
  TextField,
  Typography
} from "@mui/material";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import { ApiError, toErrorMessage } from "../../api/client";
import { createProject, listProjects } from "../../api/taskflowApi";
import { useAuth } from "../../auth/AuthProvider";

export function ProjectsPage() {
  const queryClient = useQueryClient();
  const { session, signOut } = useAuth();
  const token = session?.token ?? "";
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [projectDescription, setProjectDescription] = useState("");
  const [createError, setCreateError] = useState("");
  const [createFieldErrors, setCreateFieldErrors] = useState<Record<string, string>>({});

  const projectsQuery = useQuery({
    queryKey: ["projects", page],
    queryFn: () => listProjects(token, { page, limit: 8 }),
    enabled: Boolean(token),
    retry: false
  });

  const createMutation = useMutation({
    mutationFn: () =>
      createProject(token, {
        name: projectName.trim(),
        description: projectDescription.trim() || undefined
      }),
    onSuccess: async () => {
      setCreateOpen(false);
      setProjectName("");
      setProjectDescription("");
      setCreateError("");
      setCreateFieldErrors({});
      await queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
    onError: (error) => {
      if (error instanceof ApiError) {
        if (error.status === 401) {
          signOut();
          return;
        }
        setCreateError(error.message);
        setCreateFieldErrors(error.fields ?? {});
        return;
      }
      setCreateError(toErrorMessage(error));
    }
  });

  const onCreateSubmit = () => {
    setCreateError("");
    setCreateFieldErrors({});
    if (!projectName.trim()) {
      setCreateFieldErrors({ name: "is required" });
      return;
    }
    createMutation.mutate();
  };

  const projects = projectsQuery.data?.projects ?? [];
  const totalPages = useMemo(() => projectsQuery.data?.pagination.total_pages ?? 0, [projectsQuery.data]);

  if (projectsQuery.error instanceof ApiError && projectsQuery.error.status === 401) {
    signOut();
  }

  return (
    <Stack spacing={3}>
      <Stack direction={{ sm: "row", xs: "column" }} justifyContent="space-between" spacing={2}>
        <div>
          <Typography className="tf-hero-label">Workspace</Typography>
          <Typography sx={{ fontWeight: 700, marginTop: 1 }} variant="h3">
            Projects
          </Typography>
          <Typography color="text.secondary">
            Browse accessible projects and open details.
          </Typography>
        </div>
        <Button onClick={() => setCreateOpen(true)} variant="contained">
          New Project
        </Button>
      </Stack>

      {projectsQuery.isLoading ? (
        <Card>
          <CardContent>
            <Stack alignItems="center" direction="row" spacing={1.5}>
              <CircularProgress size={18} />
              <Typography color="text.secondary">Loading projects...</Typography>
            </Stack>
          </CardContent>
        </Card>
      ) : null}

      {projectsQuery.isError ? (
        <Alert severity="error">{toErrorMessage(projectsQuery.error)}</Alert>
      ) : null}

      {!projectsQuery.isLoading && !projectsQuery.isError && projects.length === 0 ? (
        <Card>
          <CardContent>
            <Typography sx={{ fontWeight: 600 }} variant="h6">
              No projects yet
            </Typography>
            <Typography color="text.secondary" sx={{ marginTop: 1 }}>
              Create your first project to get started.
            </Typography>
          </CardContent>
        </Card>
      ) : null}

      {projects.map((project) => (
        <Card key={project.id} sx={{ backgroundColor: "#ffffff" }}>
          <CardContent>
            <Stack spacing={1.5}>
              <Stack direction="row" justifyContent="space-between" spacing={2}>
                <Typography sx={{ fontWeight: 600 }} variant="h6">
                  {project.name}
                </Typography>
                <Chip color="secondary" label="Project" size="small" />
              </Stack>
              <Typography color="text.secondary">
                {project.description?.trim() || "No description provided."}
              </Typography>
              <Stack direction="row" spacing={1}>
                <Button component={RouterLink} size="small" to={`/projects/${project.id}`} variant="outlined">
                  Open Project
                </Button>
              </Stack>
            </Stack>
          </CardContent>
        </Card>
      ))}

      {totalPages > 1 ? (
        <Box sx={{ display: "flex", justifyContent: "center", paddingTop: 1 }}>
          <Pagination
            count={totalPages}
            onChange={(_event, nextPage) => setPage(nextPage)}
            page={page}
            shape="rounded"
          />
        </Box>
      ) : null}

      <Dialog fullWidth maxWidth="sm" onClose={() => setCreateOpen(false)} open={createOpen}>
        <DialogTitle>Create Project</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ marginTop: 1 }}>
            {createError ? <Alert severity="error">{createError}</Alert> : null}
            <TextField
              autoFocus
              error={Boolean(createFieldErrors.name)}
              helperText={createFieldErrors.name}
              label="Project Name"
              onChange={(event) => setProjectName(event.target.value)}
              value={projectName}
            />
            <TextField
              error={Boolean(createFieldErrors.description)}
              helperText={createFieldErrors.description}
              label="Description (optional)"
              minRows={3}
              multiline
              onChange={(event) => setProjectDescription(event.target.value)}
              value={projectDescription}
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ padding: 2 }}>
          <Button onClick={() => setCreateOpen(false)} variant="text">
            Cancel
          </Button>
          <Button disabled={createMutation.isPending} onClick={onCreateSubmit} variant="contained">
            Create
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
