import { http } from "@/shared/api/http";
import type { ApiResponse, Paginated } from "@/shared/types/api";

export type Medicine = {
  id: string;
  name: string;
  specification: string;
  stock: number;
  safe_stock: number;
  expiry_date: string;
  warnings: string[];
  is_low_stock: boolean;
  is_expiring_soon: boolean;
  created_at: string;
  updated_at: string;
};

type StockChangePayload = {
  medicine_id: string;
  quantity: number;
};

export async function listMedicines(params: { page?: number; pageSize?: number }) {
  const response = await http.get<ApiResponse<Paginated<Medicine>>>("/medicines", {
    params: {
      page: params.page ?? 1,
      page_size: params.pageSize ?? 20,
    },
  });
  return response.data.data;
}

export async function inboundMedicine(payload: StockChangePayload) {
  const response = await http.post<ApiResponse<Medicine>>("/medicines/inbound", payload);
  return response.data.data;
}

export async function outboundMedicine(payload: StockChangePayload) {
  const response = await http.post<ApiResponse<Medicine>>("/medicines/outbound", payload);
  return response.data.data;
}
