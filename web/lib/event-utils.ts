import type { EventResponse, ListEventsResponse } from "@/lib/types";

export type AdminUserEvent = {
  id: string;
  type: string;
  createdAt: string;
  payload: Record<string, unknown> | null;
  ipAddress?: string;
  userAgent?: string;
};

export type AdminUserEventsPage = {
  events: AdminUserEvent[];
  nextCursor?: string;
};

function parseEventData(data: string): Record<string, unknown> | null {
  if (!data) {
    return null;
  }

  try {
    const parsed = JSON.parse(data);
    if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>;
    }
    return { value: parsed };
  } catch {
    return { raw: data };
  }
}

function mapAdminUserEvent(event: EventResponse): AdminUserEvent {
  return {
    id: event.id,
    type: event.type,
    createdAt: event.created_at,
    payload: parseEventData(event.data),
    ipAddress: event.ip_address,
    userAgent: event.user_agent,
  };
}

export function mapAdminUserEvents(
  response: ListEventsResponse | null | undefined,
): AdminUserEventsPage {
  if (!response) {
    return { events: [] };
  }

  return {
    events: response.events.map(mapAdminUserEvent),
    nextCursor: response.next_cursor,
  };
}
