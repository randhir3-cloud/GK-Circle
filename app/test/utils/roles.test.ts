import { describe, it, expect } from "vitest";
import {
  normalizeRoles,
  getHighestRole,
  getRoleLabel,
  hasAnyRole,
} from "../../utils/roles";

describe("Role Normalization and Verification Helpers", () => {
  describe("normalizeRoles", () => {
    it("normalizes a single valid role string", () => {
      expect(normalizeRoles("super_admin")).toEqual(["super_admin"]);
    });

    it("normalizes a comma-separated role string", () => {
      expect(normalizeRoles("user,super_admin")).toEqual([
        "user",
        "super_admin",
      ]);
    });

    it("normalizes a comma-separated role string with spaces and casing", () => {
      expect(normalizeRoles(" USER , SUPER_ADMIN ")).toEqual([
        "user",
        "super_admin",
      ]);
    });

    it("normalizes an array of valid role strings", () => {
      expect(normalizeRoles(["user", "super_admin"])).toEqual([
        "user",
        "super_admin",
      ]);
    });

    it("maps legacy admin-user role to admin", () => {
      expect(normalizeRoles("admin-user")).toEqual(["admin"]);
      expect(normalizeRoles(["admin-user", "user"])).toEqual(["admin", "user"]);
    });

    it("maps legacy guest-user role to user", () => {
      expect(normalizeRoles("guest-user")).toEqual(["user"]);
    });

    it("filters out unknown or empty roles", () => {
      expect(normalizeRoles("unknown,super_admin,,")).toEqual(["super_admin"]);
    });

    it("de-duplicates roles", () => {
      expect(normalizeRoles(["user", "user", "super_admin"])).toEqual([
        "user",
        "super_admin",
      ]);
    });

    it("handles null, undefined, and empty inputs gracefully", () => {
      expect(normalizeRoles(null)).toEqual([]);
      expect(normalizeRoles(undefined)).toEqual([]);
      expect(normalizeRoles("")).toEqual([]);
    });
  });

  describe("getHighestRole", () => {
    it("correctly identifies the most privileged role", () => {
      expect(getHighestRole(["user", "super_admin"])).toBe("super_admin");
      expect(getHighestRole(["admin", "moderator"])).toBe("admin");
      expect(getHighestRole(["moderator", "user"])).toBe("moderator");
      expect(getHighestRole(["user"])).toBe("user");
      expect(getHighestRole([])).toBe("user");
    });
  });

  describe("getRoleLabel", () => {
    it("returns correct user-friendly display labels", () => {
      expect(getRoleLabel(["super_admin"])).toBe("Super Admin");
      expect(getRoleLabel(["admin"])).toBe("Admin");
      expect(getRoleLabel(["moderator"])).toBe("Moderator");
      expect(getRoleLabel(["user"])).toBe("User");
    });
  });

  describe("hasAnyRole", () => {
    it("returns true if user has at least one required role", () => {
      expect(
        hasAnyRole(["user", "super_admin"], ["admin", "super_admin"])
      ).toBe(true);
    });

    it("returns false if user has none of the required roles", () => {
      expect(hasAnyRole(["user", "moderator"], ["admin", "super_admin"])).toBe(
        false
      );
    });
  });
});
