export const API_BASE_URL = (
  process.env.NEXT_PUBLIC_API_BASE_URL || "/api"
).replace(/\/$/, "");
const API_REQUEST_BASE_URL = /^https?:\/\//.test(API_BASE_URL) ? "/api-proxy" : API_BASE_URL;

const TOKEN_KEY = "matchlab_token";
const AUTH_EVENT = "matchlab-auth-change";

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  if (typeof window !== "undefined") {
    window.localStorage.setItem(TOKEN_KEY, token);
    window.dispatchEvent(new Event(AUTH_EVENT));
  }
}

export function clearToken(): void {
  if (typeof window !== "undefined") {
    window.localStorage.removeItem(TOKEN_KEY);
    window.dispatchEvent(new Event(AUTH_EVENT));
  }
}

export function subscribeAuth(listener: () => void): () => void {
  if (typeof window === "undefined") return () => undefined;
  window.addEventListener(AUTH_EVENT, listener);
  window.addEventListener("storage", listener);
  return () => {
    window.removeEventListener(AUTH_EVENT, listener);
    window.removeEventListener("storage", listener);
  };
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  if (options.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  if (token) headers.set("Authorization", `Bearer ${token}`);

  let response: Response;
  try {
    response = await fetch(`${API_REQUEST_BASE_URL}${path.startsWith("/") ? path : `/${path}`}`, {
      ...options,
      headers,
    });
  } catch {
    throw new ApiError("无法连接服务器，请检查网络后重试", 0, "network_error");
  }

  const isJSON = response.headers.get("content-type")?.includes("application/json");
  let payload: { data?: unknown; error?: string; message?: string } | null = null;
  if (isJSON) {
    try {
      payload = await response.json();
    } catch {
      throw new ApiError("服务器响应格式错误，请稍后重试", response.status, "invalid_response");
    }
  }
  if (!response.ok) {
    if (response.status === 401 && token) clearToken();
    throw new ApiError(payload?.message || "请求失败，请稍后重试", response.status, payload?.error);
  }
  return (payload?.data ?? payload) as T;
}

export function postJSON<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: "POST", body: JSON.stringify(body) });
}
