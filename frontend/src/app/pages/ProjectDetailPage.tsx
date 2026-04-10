import { Card, CardContent, Grid, Stack, Typography } from "@mui/material";
import { useParams } from "react-router-dom";

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();

  return (
    <Stack spacing={3}>
      <div>
        <Typography sx={{ fontWeight: 700 }} variant="h4">
          Project Detail
        </Typography>
        <Typography color="text.secondary">Project ID: {id}</Typography>
      </div>

      <Grid container spacing={2}>
        <Grid item md={8} xs={12}>
          <Card>
            <CardContent>
              <Typography sx={{ fontWeight: 600 }} variant="h6">
                Tasks
              </Typography>
              <Typography color="text.secondary" sx={{ marginTop: 1 }}>
                Protected project detail shell is active. Task list, filters, and modal CRUD come next.
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item md={4} xs={12}>
          <Card>
            <CardContent>
              <Typography sx={{ fontWeight: 600 }} variant="h6">
                Stats
              </Typography>
              <Typography color="text.secondary" sx={{ marginTop: 1 }}>
                Stats panel foundation is ready for API data wiring.
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Stack>
  );
}
