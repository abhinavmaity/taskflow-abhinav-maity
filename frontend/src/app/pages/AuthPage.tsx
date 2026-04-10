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
import { ApiError, toErrorMessage } from "../../api/client";
import { login, register } from "../../api/taskflowApi";
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
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const content = useMemo(
    () =>
      mode === "login"
        ? {
            title: "Welcome Back",
            subtitle: "Sign in with your TaskFlow account.",
            action: "Sign In",
            altText: "Need an account?",
            altLink: "/register",
            altLinkLabel: "Create one"
          }
        : {
            title: "Create Account",
            subtitle: "Register a new TaskFlow account.",
            action: "Create Account",
            altText: "Already have an account?",
            altLink: "/login",
            altLinkLabel: "Sign in"
          },
    [mode]
  );

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (submitting) {
      return;
    }

    const trimmedEmail = email.trim().toLowerCase();
    const trimmedName = name.trim();
    const trimmedPassword = password.trim();
    setFormError("");
    setFieldErrors({});

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

    setSubmitting(true);
    try {
      const authResponse =
        mode === "login"
          ? await login({
              email: trimmedEmail,
              password: trimmedPassword
            })
          : await register({
              name: trimmedName,
              email: trimmedEmail,
              password: trimmedPassword
            });

      signIn({
        token: authResponse.token,
        user: {
          name: authResponse.user.name,
          email: authResponse.user.email
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
    } catch (error) {
      if (error instanceof ApiError && error.fields) {
        setFieldErrors(error.fields);
      }
      setFormError(toErrorMessage(error));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Box
      sx={{
        alignItems: "center",
        backgroundColor: "#ffffff",
        display: "grid",
        minHeight: "100dvh",
        padding: 2,
        position: "relative"
      }}
    >
      <Box
        aria-hidden
        sx={{
          backgroundColor: "#1d4ed8",
          border: "4px solid #000000",
          height: 72,
          left: { sm: "12%", xs: "5%" },
          position: "absolute",
          top: { sm: "16%", xs: "9%" },
          width: 72
        }}
      />
      <Box
        aria-hidden
        sx={{
          backgroundColor: "#f4c430",
          border: "4px solid #000000",
          bottom: { sm: "14%", xs: "10%" },
          height: 58,
          position: "absolute",
          right: { sm: "16%", xs: "6%" },
          width: 58
        }}
      />
      <Card sx={{ marginX: "auto", maxWidth: 460, width: "100%", zIndex: 1 }}>
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

            {formError ? <Alert severity="error">{formError}</Alert> : null}

            <Box component="form" onSubmit={onSubmit}>
              <Stack spacing={2}>
                {mode === "register" ? (
                  <TextField
                    autoComplete="name"
                    error={Boolean(fieldErrors.name)}
                    helperText={fieldErrors.name}
                    label="Name"
                    onChange={(event) => setName(event.target.value)}
                    required
                    value={name}
                  />
                ) : null}
                <TextField
                  autoComplete="email"
                  error={Boolean(fieldErrors.email)}
                  helperText={fieldErrors.email}
                  label="Email"
                  onChange={(event) => setEmail(event.target.value)}
                  required
                  type="email"
                  value={email}
                />
                <TextField
                  autoComplete={mode === "login" ? "current-password" : "new-password"}
                  error={Boolean(fieldErrors.password)}
                  helperText={fieldErrors.password}
                  label="Password"
                  onChange={(event) => setPassword(event.target.value)}
                  required
                  type="password"
                  value={password}
                />
                <Button disabled={submitting} size="large" type="submit" variant="contained">
                  {content.action}
                </Button>
              </Stack>
            </Box>

            <Typography color="text.secondary" variant="body2">
              {content.altText}{" "}
              <Typography
                component={RouterLink}
                sx={{ color: "secondary.main", fontWeight: 700, textDecoration: "none", textTransform: "uppercase" }}
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
