import { Link, Route, Routes } from "react-router-dom";

function PlaceholderPage({ title }: { title: string }) {
  return (
    <main style={{ padding: "1.5rem", fontFamily: "system-ui, sans-serif" }}>
      <h1>{title}</h1>
      <p>Milestone 0 scaffold route.</p>
      <nav style={{ display: "flex", gap: "1rem", marginTop: "1rem" }}>
        <Link to="/login">Login</Link>
        <Link to="/register">Register</Link>
        <Link to="/projects">Projects</Link>
      </nav>
    </main>
  );
}

export function App() {
  return (
    <Routes>
      <Route path="/" element={<PlaceholderPage title="TaskFlow" />} />
      <Route path="/login" element={<PlaceholderPage title="Login" />} />
      <Route path="/register" element={<PlaceholderPage title="Register" />} />
      <Route path="/projects" element={<PlaceholderPage title="Projects" />} />
      <Route path="/projects/:id" element={<PlaceholderPage title="Project Detail" />} />
    </Routes>
  );
}

