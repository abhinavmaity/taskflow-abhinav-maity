import { createTheme } from "@mui/material";

export const appTheme = createTheme({
  palette: {
    background: {
      default: "#ffffff",
      paper: "#ffffff"
    },
    primary: {
      main: "#e63946",
      contrastText: "#ffffff"
    },
    secondary: {
      main: "#1d4ed8",
      contrastText: "#ffffff"
    },
    warning: {
      main: "#f4c430",
      contrastText: "#000000"
    },
    success: {
      main: "#f1faee",
      contrastText: "#000000"
    },
    text: {
      primary: "#000000",
      secondary: "#2f2f2f"
    }
  },
  shape: {
    borderRadius: 0
  },
  typography: {
    fontFamily: "\"Work Sans\", \"Segoe UI\", sans-serif",
    h3: {
      fontFamily: "\"Space Grotesk\", \"Work Sans\", sans-serif",
      fontSize: "2rem",
      fontWeight: 700,
      letterSpacing: "-0.03em",
      textTransform: "uppercase"
    },
    h4: {
      fontFamily: "\"Space Grotesk\", \"Work Sans\", sans-serif",
      fontWeight: 700,
      letterSpacing: "-0.02em",
      textTransform: "uppercase"
    },
    h6: {
      fontFamily: "\"Space Grotesk\", \"Work Sans\", sans-serif",
      fontWeight: 700,
      letterSpacing: "-0.01em",
      textTransform: "uppercase"
    },
    caption: {
      fontSize: "0.68rem",
      fontWeight: 600,
      letterSpacing: "0.22em",
      textTransform: "uppercase"
    }
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          color: "#000000"
        }
      }
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          border: "4px solid #000000",
          borderRadius: 0,
          boxShadow: "8px 8px 0 #000000"
        }
      }
    },
    MuiCard: {
      styleOverrides: {
        root: {
          border: "4px solid #000000",
          borderRadius: 0,
          boxShadow: "8px 8px 0 #000000"
        }
      }
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          borderBottom: "4px solid #000000",
          boxShadow: "none"
        }
      }
    },
    MuiButton: {
      styleOverrides: {
        root: {
          border: "2px solid #000000",
          borderRadius: 0,
          boxShadow: "4px 4px 0 #000000",
          fontWeight: 700,
          letterSpacing: "0.1em",
          textTransform: "uppercase",
          transition: "transform 120ms ease, box-shadow 120ms ease, background-color 120ms ease",
          "&:active": {
            boxShadow: "none",
            transform: "translate(2px, 2px)"
          }
        },
        outlined: {
          backgroundColor: "#ffffff"
        }
      }
    },
    MuiChip: {
      styleOverrides: {
        root: {
          border: "2px solid #000000",
          borderRadius: 0,
          fontWeight: 700,
          letterSpacing: "0.08em",
          textTransform: "uppercase"
        }
      }
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          borderRadius: 0,
          backgroundColor: "#ffffff",
          "& .MuiOutlinedInput-notchedOutline": {
            borderColor: "#000000",
            borderWidth: 2
          }
        }
      }
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          border: "4px solid #000000",
          borderRadius: 0,
          boxShadow: "12px 12px 0 #000000"
        }
      }
    },
    MuiPaginationItem: {
      styleOverrides: {
        root: {
          border: "2px solid #000000",
          borderRadius: 0,
          fontWeight: 700
        }
      }
    }
  }
});
