import {
  AppBar,
  Box,
  Button,
  Container,
  Stack,
  Toolbar,
  Typography
} from "@mui/material";
import { Link as RouterLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../../auth/AuthProvider";

export function AppShell() {
  const { session, signOut } = useAuth();
  const navigate = useNavigate();

  const handleSignOut = () => {
    signOut();
    navigate("/login", { replace: true });
  };

  return (
    <Box sx={{ minHeight: "100dvh", backgroundColor: "#f8f9fb" }}>
      <AppBar color="inherit" position="static" elevation={0}>
        <Toolbar>
          <Typography
            component={RouterLink}
            sx={{
              color: "inherit",
              fontWeight: 700,
              letterSpacing: "0.06em",
              textDecoration: "none",
              textTransform: "uppercase"
            }}
            to="/projects"
            variant="h6"
          >
            TaskFlow
          </Typography>
          <Box sx={{ flexGrow: 1 }} />
          <Stack alignItems="flex-end" spacing={0.25}>
            <Typography variant="caption">{session?.user.name ?? "Unknown User"}</Typography>
            <Typography color="text.secondary" variant="caption">
              {session?.user.email ?? "unknown@example.com"}
            </Typography>
          </Stack>
          <Button
            color="inherit"
            onClick={handleSignOut}
            size="small"
            sx={{ marginLeft: 2 }}
            variant="outlined"
          >
            Logout
          </Button>
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg" sx={{ paddingY: 4 }}>
        <Outlet />
      </Container>
    </Box>
  );
}
