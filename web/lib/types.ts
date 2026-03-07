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
