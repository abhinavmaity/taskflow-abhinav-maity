import {
  Box,
  Button,
  Stack,
  Typography
} from "@mui/material";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../../auth/AuthProvider";

export function AppShell() {
  const { session, signOut } = useAuth();
  const navigate = useNavigate();

  const handleSignOut = () => {
    signOut();
    navigate("/login", { replace: true });
  };

  return (
    <Box sx={{ minHeight: "100dvh", padding: { xs: 2, md: 3 } }}>
      <Box
        sx={{
          backgroundColor: "#ffffff",
          border: "4px solid #000000",
          display: "grid",
          gridTemplateColumns: { xs: "1fr", md: "260px minmax(0, 1fr)" },
          minHeight: "calc(100dvh - 32px)"
        }}
      >
        <Stack
          sx={{
            backgroundColor: "#ffffff",
            borderBottom: { xs: "4px solid #000000", md: "none" },
            borderRight: { md: "4px solid #000000" },
            padding: 2.5
          }}
          spacing={2}
        >
          <Box>
            <Typography sx={{ letterSpacing: "0.14em" }} variant="caption">
              Workspace
            </Typography>
            <Typography sx={{ marginTop: 0.5 }} variant="h4">
              TaskFlow
            </Typography>
          </Box>

          <Box
            component={NavLink}
            style={({ isActive }) => ({
              backgroundColor: isActive ? "#f1faee" : "#ffffff",
              border: "2px solid #000000",
              color: "#000000",
              display: "block",
              fontFamily: "\"Space Grotesk\", \"Work Sans\", sans-serif",
              fontWeight: 700,
              letterSpacing: "0.12em",
              padding: "10px 12px",
              textDecoration: "none",
              textTransform: "uppercase"
            })}
            to="/projects"
          >
            Projects
          </Box>

          <Box sx={{ flexGrow: 1 }} />

          <Stack spacing={0.5}>
            <Typography variant="caption">{session?.user.name ?? "Unknown User"}</Typography>
            <Typography color="text.secondary" sx={{ fontSize: "0.78rem" }}>
              {session?.user.email ?? "unknown@example.com"}
            </Typography>
          </Stack>
          <Button color="primary" onClick={handleSignOut} variant="contained">
            Logout
          </Button>
        </Stack>

        <Box
          sx={{
            backgroundColor: "#ffffff",
            padding: { xs: 2, md: 4 }
          }}
        >
          <Outlet />
        </Box>
      </Box>
    </Box>
  );
}
