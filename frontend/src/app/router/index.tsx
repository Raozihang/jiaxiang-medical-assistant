import { Spin } from "antd";
import { Suspense, lazy, type ReactNode } from "react";
import { createBrowserRouter, Navigate, Outlet } from "react-router-dom";
import {
  getStoredUser,
  hasValidSession,
  resolveHomePath,
  type UserRole,
} from "@/shared/auth/session";
import { MainLayout } from "@/shared/layouts/MainLayout";

const DashboardPage = lazy(() =>
  import("@/pages/admin/DashboardPage").then((module) => ({ default: module.DashboardPage })),
);
const ImportsPage = lazy(() =>
  import("@/pages/admin/ImportsPage").then((module) => ({ default: module.ImportsPage })),
);
const NotificationsPage = lazy(() =>
  import("@/pages/admin/NotificationsPage").then((module) => ({ default: module.NotificationsPage })),
);
const ReportsPage = lazy(() =>
  import("@/pages/admin/ReportsPage").then((module) => ({ default: module.ReportsPage })),
);
const SafetyPage = lazy(() =>
  import("@/pages/admin/SafetyPage").then((module) => ({ default: module.SafetyPage })),
);
const LoginPage = lazy(() =>
  import("@/pages/auth/LoginPage").then((module) => ({ default: module.LoginPage })),
);
const MedicinesPage = lazy(() =>
  import("@/pages/doctor/MedicinesPage").then((module) => ({ default: module.MedicinesPage })),
);
const VisitDetailPage = lazy(() =>
  import("@/pages/doctor/VisitDetailPage").then((module) => ({ default: module.VisitDetailPage })),
);
const VisitsPage = lazy(() =>
  import("@/pages/doctor/VisitsPage").then((module) => ({ default: module.VisitsPage })),
);
const ForbiddenPage = lazy(() =>
  import("@/pages/ForbiddenPage").then((module) => ({ default: module.ForbiddenPage })),
);
const NotFoundPage = lazy(() =>
  import("@/pages/NotFoundPage").then((module) => ({ default: module.NotFoundPage })),
);
const CheckInPage = lazy(() =>
  import("@/pages/student/CheckInPage").then((module) => ({ default: module.CheckInPage })),
);

function RouteLoading() {
  return (
    <div style={{ minHeight: 240, display: "grid", placeItems: "center" }}>
      <Spin size="large" tip="Loading page..." />
    </div>
  );
}

function withSuspense(element: ReactNode) {
  return <Suspense fallback={<RouteLoading />}>{element}</Suspense>;
}

function HomeRedirect() {
  if (!hasValidSession()) {
    return <Navigate to="/login" replace />;
  }

  return <Navigate to={resolveHomePath(getStoredUser()?.role)} replace />;
}

function RequireAuth() {
  if (!hasValidSession()) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}

function RequireRole({ role }: { role: UserRole }) {
  const user = getStoredUser();
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  if (user.role !== role) {
    return <Navigate to="/forbidden" replace />;
  }
  return <Outlet />;
}

function PublicOnly() {
  if (hasValidSession()) {
    return <Navigate to={resolveHomePath(getStoredUser()?.role)} replace />;
  }

  return <Outlet />;
}

export const router = createBrowserRouter([
  {
    element: <PublicOnly />,
    children: [{ path: "/login", element: withSuspense(<LoginPage />) }],
  },
  {
    path: "/",
    element: <MainLayout />,
    children: [
      { index: true, element: <HomeRedirect /> },
      { path: "student/checkin", element: withSuspense(<CheckInPage />) },
      { path: "forbidden", element: withSuspense(<ForbiddenPage />) },
      {
        element: <RequireAuth />,
        children: [
          {
            element: <RequireRole role="doctor" />,
            children: [
              { path: "doctor/visits", element: withSuspense(<VisitsPage />) },
              { path: "doctor/visit/:id", element: withSuspense(<VisitDetailPage />) },
              { path: "doctor/medicines", element: withSuspense(<MedicinesPage />) },
            ],
          },
          {
            element: <RequireRole role="admin" />,
            children: [
              { path: "admin/dashboard", element: withSuspense(<DashboardPage />) },
              { path: "admin/imports", element: withSuspense(<ImportsPage />) },
              { path: "admin/reports", element: withSuspense(<ReportsPage />) },
              { path: "admin/notifications", element: withSuspense(<NotificationsPage />) },
              { path: "admin/safety", element: withSuspense(<SafetyPage />) },
            ],
          },
        ],
      },
    ],
  },
  { path: "*", element: withSuspense(<NotFoundPage />) },
]);

