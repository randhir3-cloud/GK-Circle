import { describe, it, expect, vi, beforeEach } from "vitest";
import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import authorizationMiddleware from "../../middleware/authorization";

const { storeState, navigateToSpy, mockStore } = vi.hoisted(() => ({
  storeState: { user: null as null | { role: string; roles: string[] } },
  navigateToSpy: vi.fn((url) => url),
  mockStore: {
    getUserData: () => storeState.user,
    fetchAuthenticatedUser: vi.fn(async () => {
      // Simulate session hydration
      storeState.user = {
        role: "admin-user",
        roles: ["admin"],
      };
    }),
  },
}));

vi.mock("~~/store/users", () => ({
  useUsersStore: () => mockStore,
}));

mockNuxtImport("navigateTo", () => navigateToSpy);

describe("Authorization Route Middleware", () => {
  beforeEach(() => {
    storeState.user = null;
    vi.clearAllMocks();
  });

  it("skips checks if route does not require roles", async () => {
    const to = { meta: {} };
    const result = await authorizationMiddleware(to);
    expect(result).toBeUndefined();
  });

  it("allows access if user has required roles", async () => {
    storeState.user = {
      role: "admin-user",
      roles: ["admin"],
    };
    const to = { meta: { requiredRoles: ["admin"] } };
    const result = await authorizationMiddleware(to);
    expect(result).toBeUndefined();
  });

  it("redirects to login if user is not logged in and hydration fails", async () => {
    const to = { meta: { requiredRoles: ["admin"] } };
    // Force hydration failure
    mockStore.fetchAuthenticatedUser.mockRejectedValueOnce(
      new Error("Network Error")
    );

    await authorizationMiddleware(to);
    expect(navigateToSpy).toHaveBeenCalledWith("/account/login");
  });

  it("denies access and redirects to list page if user lacks roles", async () => {
    storeState.user = {
      role: "guest-user",
      roles: ["user"],
    };
    const to = { meta: { requiredRoles: ["admin"] } };
    await authorizationMiddleware(to);
    expect(navigateToSpy).toHaveBeenCalledWith("/admin/quiz/list-quiz");
  });
});
