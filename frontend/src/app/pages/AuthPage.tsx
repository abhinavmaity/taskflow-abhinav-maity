import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Stack,
  TextField,
  Typography
} from "@mui/material";
import { FormEvent, useMemo, useState } from "react";
import { Link as RouterLink, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../auth/AuthProvider";

type AuthPageMode = "login" | "register";

type AuthPageProps = {
  mode: AuthPageMode;
};

export function AuthPage({ mode }: AuthPageProps) {
  const { signIn } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [formError, setFormError] = useState("");

  const content = useMemo(
    () =>
      mode === "login"
        ? {
            title: "Welcome Back",
            subtitle: "Use local session auth for now. API auth wiring is next.",
            action: "Sign In",
            altText: "Need an account?",
            altLink: "/register",
            altLinkLabel: "Create one"
          }
        : {
            title: "Create Account",
            subtitle: "This foundation stores auth locally until API integration is added.",
            action: "Create Account",
            altText: "Already have an account?",
            altLink: "/login",
            altLinkLabel: "Sign in"
          },
    [mode]
  );

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedEmail = email.trim().toLowerCase();
    const trimmedName = name.trim();
    const trimmedPassword = password.trim();

    if (!trimmedEmail || !trimmedPassword || (mode === "register" && !trimmedName)) {
      setFormError("Please complete all required fields.");
      return;
    }

    if (!trimmedEmail.includes("@")) {
      setFormError("Please enter a valid email address.");
      return;
    }

    if (trimmedPassword.length < 8) {
      setFormError("Password must be at least 8 characters.");
      return;
    }

    signIn({
      token: `local-session-${Date.now()}`,
      user: {
        name: trimmedName || trimmedEmail.split("@")[0],
        email: trimmedEmail
      }
    });

    const redirectTo =
      typeof location.state === "object" &&
      location.state &&
      "from" in location.state &&
      typeof location.state.from === "string"
        ? location.state.from
        : "/projects";

    navigate(redirectTo, { replace: true });
  };

  return (
    <Box
      sx={{
        alignItems: "center",
        background: "linear-gradient(160deg, #f5f7ff 0%, #f3f7f2 100%)",
        display: "grid",
        minHeight: "100dvh",
        padding: 2
      }}
    >
      <Card sx={{ marginX: "auto", maxWidth: 460, width: "100%" }}>
        <CardContent sx={{ padding: { sm: 5, xs: 3 } }}>
          <Stack spacing={3}>
            <Box>
              <Typography sx={{ fontWeight: 700 }} variant="h4">
                {content.title}
              </Typography>
              <Typography color="text.secondary" sx={{ marginTop: 1 }}>
                {content.subtitle}
              </Typography>
            </Box>

            <Alert severity="info">
              Milestone 6 foundation mode: route guards and local auth persistence are enabled.
            </Alert>

            {formError ? <Alert severity="error">{formError}</Alert> : null}

            <Box component="form" onSubmit={onSubmit}>
              <Stack spacing={2}>
                {mode === "register" ? (
                  <TextField
                    autoComplete="name"
                    label="Name"
                    onChange={(event) => setName(event.target.value)}
                    required
                    value={name}
                  />
                ) : null}
                <TextField
                  autoComplete="email"
                  label="Email"
                  onChange={(event) => setEmail(event.target.value)}
                  required
                  type="email"
                  value={email}
                />
                <TextField
                  autoComplete={mode === "login" ? "current-password" : "new-password"}
                  label="Password"
                  onChange={(event) => setPassword(event.target.value)}
                  required
                  type="password"
                  value={password}
                />
                <Button size="large" type="submit" variant="contained">
                  {content.action}
                </Button>
              </Stack>
            </Box>

            <Typography color="text.secondary" variant="body2">
              {content.altText}{" "}
              <Typography
                component={RouterLink}
                sx={{ color: "primary.main", fontWeight: 600, textDecoration: "none" }}
                to={content.altLink}
                variant="inherit"
              >
                {content.altLinkLabel}
              </Typography>
            </Typography>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
}
