export type User = {
  id: string;
  username: string;
  email: string;
  email_verified: boolean;
  roles: string[];
  created_at: string;
  updated_at: string;
};

export type LoginResponse = {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token: string;
};

export type ListUsersResponse = {
  users: User[];
  next_cursor?: string;
};

export type ListRolesResponse = {
  roles: string[];
};

export type EventResponse = {
  id: string;
  user_id: string;
  type: string;
  data: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
};

export type ListEventsResponse = {
  events: EventResponse[];
  next_cursor?: string;
};

export type RefreshTokenResponse = {
  id: string;
  user_id: string;
  issued_at: string;
  expires_at: string;
  revoked_at?: string | null;
  ip_address: string;
  user_agent: string;
  status: "active" | "expired" | "revoked";
};

export type ListRefreshTokensResponse = {
  refresh_tokens: RefreshTokenResponse[];
  next_cursor?: string;
};

export type Client = {
  id: string;
  user_id: string;
  name: string;
  redirect_uri: string;
  allowed_scopes: string[];
  revoked_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type ListClientsResponse = {
  clients: Client[];
  next_cursor?: string;
};
