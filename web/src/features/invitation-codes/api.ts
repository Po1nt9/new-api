import { api } from "@/lib/api";
import type {
  Invitation,
  ApiResponse,
  GetInvitationsParams,
  GetInvitationsResponse,
  SearchInvitationsParams,
  InvitationFormData,
} from "./types";

export async function getInvitations(
  params: GetInvitationsParams = {}
): Promise<GetInvitationsResponse> {
  const { p = 1, page_size = 10 } = params;
  const res = await api.get(`/api/invitation/?p=${p}&page_size=${page_size}`);
  return res.data;
}

export async function searchInvitations(
  params: SearchInvitationsParams
): Promise<GetInvitationsResponse> {
  const { keyword = "", status = "", p = 1, page_size = 10 } = params;
  const queryParams = new URLSearchParams();
  queryParams.set("keyword", keyword);
  if (status) queryParams.set("status", status);
  queryParams.set("p", String(p));
  queryParams.set("page_size", String(page_size));
  const res = await api.get(`/api/invitation/search?${queryParams.toString()}`);
  return res.data;
}

export async function getInvitation(
  id: number
): Promise<ApiResponse<Invitation>> {
  const res = await api.get(`/api/invitation/${id}`);
  return res.data;
}

export async function createInvitation(
  data: InvitationFormData
): Promise<ApiResponse<string[]>> {
  const res = await api.post("/api/invitation/", data);
  return res.data;
}

export async function updateInvitation(
  data: InvitationFormData & { id: number }
): Promise<ApiResponse<Invitation>> {
  const res = await api.put("/api/invitation/", data);
  return res.data;
}

export async function updateInvitationStatus(
  id: number,
  status: number
): Promise<ApiResponse<Invitation>> {
  const res = await api.put("/api/invitation/?status_only=true", {
    id,
    status,
  });
  return res.data;
}

export async function deleteInvitation(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/invitation/${id}`);
  return res.data;
}

export async function batchDeleteInvitations(
  ids: number[]
): Promise<ApiResponse<number>> {
  const res = await api.post("/api/invitation/batch_delete", { ids });
  return res.data;
}

export async function deleteInvalidInvitations(): Promise<ApiResponse<number>> {
  const res = await api.delete("/api/invitation/invalid");
  return res.data;
}
