import type { ChatMessage, Circle } from "./types";

export const CIRCLE_CATEGORIES = [
  ["all", "全部"], ["study", "学习"], ["sports", "运动"],
  ["interest", "兴趣"], ["career", "职涯"], ["other", "其他"],
] as const;

export function categoryLabel(value: string) {
  return CIRCLE_CATEGORIES.find(([key]) => key === value)?.[1] ?? value;
}

export function filterCircles(items: Circle[], query: string, category: string) {
  const needle = query.trim().toLocaleLowerCase();
  return items.filter((circle) => circle.status === "approved")
    .filter((circle) => category === "all" || circle.category === category)
    .filter((circle) => !needle || `${circle.name} ${circle.description}`.toLocaleLowerCase().includes(needle));
}

export function mergeMessages(current: ChatMessage[], incoming: ChatMessage[]) {
  const unique = new Map(current.map((message) => [message.id, message]));
  incoming.forEach((message) => unique.set(message.id, message));
  return [...unique.values()].sort((a, b) => a.created_at.localeCompare(b.created_at) || a.id.localeCompare(b.id));
}

export function messageCursor(messages: ChatMessage[]) {
  const last = messages.at(-1);
  return last ? { after_time: last.created_at, after_id: last.id } as Record<string, string> : {} as Record<string, string>;
}

export function listFrom<T>(payload: unknown, key: string): T[] {
  if (Array.isArray(payload)) return payload as T[];
  const value = (payload as Record<string, unknown> | null)?.[key];
  return Array.isArray(value) ? value as T[] : [];
}

