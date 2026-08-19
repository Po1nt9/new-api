import { z } from "zod";

export const invitationSchema = z.object({
  id: z.number(),
  user_id: z.number(),
  name: z.string(),
  key: z.string(),
  status: z.number(), // 1: enabled, 2: disabled, 3: used
  quota: z.number(),
  group: z.string().optional().default(""),
  created_time: z.number(),
  used_time: z.number(),
  expired_time: z.number(), // 0 for never expires
  used_user_id: z.number(),
});

export type Invitation = z.infer<typeof invitationSchema>;

export interface ApiResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
}

export interface GetInvitationsParams {
  p?: number;
  page_size?: number;
}

export interface GetInvitationsResponse {
  success: boolean;
  message?: string;
  data?: {
    items: Invitation[];
    total: number;
    page: number;
    page_size: number;
  };
}

export interface SearchInvitationsParams {
  keyword?: string;
  status?: string;
  p?: number;
  page_size?: number;
}

export interface InvitationFormData {
  id?: number;
  name: string;
  prefix?: string;
  key?: string;
  quota: number;
  group?: string;
  expired_time: number;
  count?: number;
  status?: number;
}

export type InvitationsDialogType = "create" | "update" | "delete" | "view";
