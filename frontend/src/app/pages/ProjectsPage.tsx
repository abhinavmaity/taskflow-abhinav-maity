import { Button, Card, CardContent, Chip, Stack, Typography } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";

const demoProjectId = "33333333-3333-3333-3333-333333333333";

export function ProjectsPage() {
  return (
    <Stack spacing={3}>
      <Stack direction={{ sm: "row", xs: "column" }} justifyContent="space-between" spacing={2}>
        <div>
          <Typography sx={{ fontWeight: 700 }} variant="h4">
            Projects
          </Typography>
          <Typography color="text.secondary">
            Protected route is enabled. Data fetching lands in the next task batch.
          </Typography>
        </div>
        <Button disabled variant="contained">
          New Project (Next)
        </Button>
      </Stack>

      <Card>
        <CardContent>
          <Stack spacing={1.5}>
            <Stack direction="row" justifyContent="space-between" spacing={2}>
              <Typography sx={{ fontWeight: 600 }} variant="h6">
                TaskFlow Demo Project
              </Typography>
              <Chip color="primary" label="Seeded" size="small" />
            </Stack>
            <Typography color="text.secondary">
              Foundation placeholder until API list/create integration is completed.
            </Typography>
            <Stack direction="row" spacing={1}>
              <Button component={RouterLink} size="small" to={`/projects/${demoProjectId}`} variant="outlined">
                Open Project
              </Button>
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}
