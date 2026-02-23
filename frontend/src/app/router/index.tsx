import { createBrowserRouter, Navigate } from "react-router-dom";
import { DashboardPage } from "@/pages/admin/DashboardPage";
import { ImportsPage } from "@/pages/admin/ImportsPage";
import { NotificationsPage } from "@/pages/admin/NotificationsPage";
import { ReportsPage } from "@/pages/admin/ReportsPage";
import { SafetyPage } from "@/pages/admin/SafetyPage";
import { MedicinesPage } from "@/pages/doctor/MedicinesPage";
import { VisitDetailPage } from "@/pages/doctor/VisitDetailPage";
import { VisitsPage } from "@/pages/doctor/VisitsPage";
import { NotFoundPage } from "@/pages/NotFoundPage";
import { CheckInPage } from "@/pages/student/CheckInPage";
import { MainLayout } from "@/shared/layouts/MainLayout";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <MainLayout />,
    children: [
      { index: true, element: <Navigate to="/doctor/visits" replace /> },
      { path: "/student/checkin", element: <CheckInPage /> },
      { path: "/doctor/visits", element: <VisitsPage /> },
      { path: "/doctor/visit/:id", element: <VisitDetailPage /> },
      { path: "/doctor/medicines", element: <MedicinesPage /> },
      { path: "/admin/dashboard", element: <DashboardPage /> },
      { path: "/admin/imports", element: <ImportsPage /> },
      { path: "/admin/reports", element: <ReportsPage /> },
      { path: "/admin/notifications", element: <NotificationsPage /> },
      { path: "/admin/safety", element: <SafetyPage /> },
    ],
  },
  { path: "*", element: <NotFoundPage /> },
]);

