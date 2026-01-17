import React from "react";
import { Refine } from "@refinedev/core";
import { RefineKbar, RefineKbarProvider } from "@refinedev/kbar";
import dataProviderHasura from "@refinedev/hasura";
import { BrowserRouter, Outlet, Route, Routes, Navigate } from "react-router-dom";
import "./index.css";

import { authProvider } from "./providers/auth-provider";
import { dataProvider, liveProvider } from "./providers/data-provider";
import { accessControlProvider } from "./providers/access-control-provider";
import { notificationProvider } from "./providers/notification-provider";

import { Layout } from "./components/layout";
import { Dashboard } from "./pages/dashboard";
import { ApplicationList } from "./pages/applications/list";
import { ApplicationCreate } from "./pages/applications/create";
import { DatabaseList } from "./pages/databases/list";
import { DatabaseCreate } from "./pages/databases/create";
import { ProjectList } from "./pages/projects/list";
import { ProjectCreate } from "./pages/projects/create";
import { SecretList } from "./pages/secrets/list";
import { SecretCreate } from "./pages/secrets/create";
import { DeploymentList } from "./pages/deployments";
import { LogsPage } from "./pages/observability/logs";
import { MetricsPage } from "./pages/observability/metrics";
import { LoginPage } from "./pages/auth/login";
import { PlaceholderPage } from "./pages/placeholder";
import { APIExplorerPage } from "./pages/settings";

function App() {
  return (
    <BrowserRouter>
      <RefineKbarProvider>
        <Refine
          dataProvider={dataProvider}
          liveProvider={liveProvider}
          authProvider={authProvider}
          accessControlProvider={accessControlProvider}
          notificationProvider={notificationProvider}
          resources={[
            {
              name: "dashboard",
              list: "/",
              meta: { label: "Dashboard" },
            },
            {
              name: "applications",
              list: "/applications",
              create: "/applications/create",
              edit: "/applications/:id",
              show: "/applications/:id",
              meta: { label: "Applications" },
            },
            {
              name: "databases",
              list: "/databases",
              create: "/databases/create",
              edit: "/databases/:id",
              show: "/databases/:id",
              meta: { label: "Databases" },
            },
            {
              name: "deployments",
              list: "/deployments",
              meta: { label: "Deployments" },
            },
            {
              name: "projects",
              list: "/projects",
              create: "/projects/create",
              meta: { label: "Projects" },
            },
            {
              name: "secrets",
              list: "/secrets",
              create: "/secrets/create",
              meta: { label: "Secrets" },
            },
            {
              name: "logs",
              list: "/logs",
              meta: { label: "Logs" },
            },
            {
              name: "metrics",
              list: "/metrics",
              meta: { label: "Metrics" },
            },
            {
              name: "settings",
              list: "/settings",
              meta: { label: "Settings" },
            },
          ]}
          options={{
            syncWithLocation: true,
            warnWhenUnsavedChanges: true,
            projectId: "antigravity-platform",
            liveMode: "auto",
          }}
        >
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<LoginPage />} />

            {/* Protected Routes */}
            <Route
              element={
                <Layout>
                  <Outlet />
                </Layout>
              }
            >
              {/* Dashboard */}
              <Route index element={<Dashboard />} />

              {/* Applications */}
              <Route path="/applications">
                <Route index element={<ApplicationList />} />
                <Route path="create" element={<ApplicationCreate />} />
                <Route path=":id" element={<PlaceholderPage />} />
              </Route>

              {/* Databases */}
              <Route path="/databases">
                <Route index element={<DatabaseList />} />
                <Route path="create" element={<DatabaseCreate />} />
                <Route path=":id" element={<PlaceholderPage />} />
              </Route>

              {/* Deployments */}
              <Route path="/deployments" element={<DeploymentList />} />

              {/* Projects */}
              <Route path="/projects">
                <Route index element={<ProjectList />} />
                <Route path="create" element={<ProjectCreate />} />
              </Route>

              {/* Secrets */}
              <Route path="/secrets">
                <Route index element={<SecretList />} />
                <Route path="create" element={<SecretCreate />} />
              </Route>

              {/* Observability */}
              <Route path="/logs" element={<LogsPage />} />
              <Route path="/metrics" element={<MetricsPage />} />

              {/* Settings */}
              <Route path="/settings" element={<APIExplorerPage />} />

              {/* Catch all */}
              <Route path="*" element={<Navigate to="/" />} />
            </Route>
          </Routes>
          <RefineKbar />
        </Refine>
      </RefineKbarProvider>
    </BrowserRouter>
  );
}

export default App;

