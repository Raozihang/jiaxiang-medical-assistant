import { createBrowserRouter, Navigate, Outlet } from "react-router-dom";
import { DashboardPage } from "@/pages/admin/DashboardPage";
import { ImportsPage } from "@/pages/admin/ImportsPage";
import { NotificationsPage } from "@/pages/admin/NotificationsPage";
import { ReportsPage } from "@/pages/admin/ReportsPage";
import { SafetyPage } from "@/pages/admin/SafetyPage";
import { LoginPage } from "@/pages/auth/LoginPage";
import { MedicinesPage } from "@/pages/doctor/MedicinesPage";
import { VisitDetailPage } from "@/pages/doctor/VisitDetailPage";
import { VisitsPage } from "@/pages/doctor/VisitsPage";
import { NotFoundPage } from "@/pages/NotFoundPage";
import { CheckInPage } from "@/pages/student/CheckInPage";
import { getStoredUser, hasValidSession, resolveHomePath } from "@/shared/auth/session";
import { MainLayout } from "@/shared/layouts/MainLayout";

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

function PublicOnly() {
  if (hasValidSession()) {
    return <Navigate to={resolveHomePath(getStoredUser()?.role)} replace />;
  }

  return <Outlet />;
}

export const router = createBrowserRouter([
  {
    element: <PublicOnly />,
    children: [{ path: "/login", element: <LoginPage /> }],
  },
  {
    path: "/",
    element: <MainLayout />,
    children: [
      { index: true, element: <HomeRedirect /> },
      { path: "student/checkin", element: <CheckInPage /> },
      {
        element: <RequireAuth />,
        children: [
          { path: "doctor/visits", element: <VisitsPage /> },
          { path: "doctor/visit/:id", element: <VisitDetailPage /> },
          { path: "doctor/medicines", element: <MedicinesPage /> },
          { path: "admin/dashboard", element: <DashboardPage /> },
          { path: "admin/imports", element: <ImportsPage /> },
          { path: "admin/reports", element: <ReportsPage /> },
          { path: "admin/notifications", element: <NotificationsPage /> },
          { path: "admin/safety", element: <SafetyPage /> },
        ],
      },
    ],
  },
  { path: "*", element: <NotFoundPage /> },
]);

