import { afterEach, describe, expect, it, vi } from "vitest";
import { API_BASE_URL, ApiError, getToken, request, setToken, subscribeAuth } from "./api";

afterEach(() => {
  localStorage.clear();
  vi.unstubAllGlobals();
});

describe("API client", () => {
  it("defaults to the same-origin API path", () => {
    expect(API_BASE_URL).toBe("/api");
  });

  it("stores and reads the access token", () => {
    setToken("token-123");
    expect(getToken()).toBe("token-123");
  });

  it("notifies subscribers when authentication changes", () => {
    const listener = vi.fn();
    const unsubscribe = subscribeAuth(listener);
    setToken("token-123");
    expect(listener).toHaveBeenCalledOnce();
    unsubscribe();
  });

  it("adds authorization and unwraps the data envelope", async () => {
    setToken("token-123");
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ data: { users: [] } }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(request<{ users: unknown[] }>("/admin/users")).resolves.toEqual({ users: [] });
    expect(fetchMock.mock.calls[0][0]).toBe("/api/admin/users");
    expect((fetchMock.mock.calls[0][1].headers as Headers).get("Authorization")).toBe(
      "Bearer token-123",
    );
  });

  it("uses the backend message for failed responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: "forbidden", message: "无权限访问" }), {
          status: 403,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );

    await expect(request("/admin/stats")).rejects.toEqual(
      expect.objectContaining<ApiError>({ status: 403, message: "无权限访问" }),
    );
  });

  it("clears stale authentication after a protected 401", async () => {
    const listener = vi.fn();
    setToken("expired-token");
    const unsubscribe = subscribeAuth(listener);
    listener.mockClear();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: "invalid_token", message: "登录凭证无效或已过期" }), {
          status: 401,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );

    await expect(request("/me")).rejects.toEqual(
      expect.objectContaining<ApiError>({ status: 401, code: "invalid_token" }),
    );
    expect(getToken()).toBeNull();
    expect(listener).toHaveBeenCalledOnce();
    unsubscribe();
  });

  it("turns malformed JSON into a friendly API error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response("{broken", { status: 200, headers: { "Content-Type": "application/json" } }),
      ),
    );

    await expect(request("/activities")).rejects.toEqual(
      expect.objectContaining<ApiError>({ status: 200, code: "invalid_response" }),
    );
  });
});
