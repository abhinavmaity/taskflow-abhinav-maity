import { createTheme } from "@mui/material";

export const appTheme = createTheme({
  palette: {
    background: {
      default: "#f8f9fb",
      paper: "#ffffff"
    },
    primary: {
      main: "#2251d1"
    }
  },
  shape: {
    borderRadius: 10
  },
  typography: {
    fontFamily: "\"Work Sans\", \"Segoe UI\", sans-serif",
    h4: {
      fontWeight: 700,
      letterSpacing: "-0.02em"
    }
  }
});
