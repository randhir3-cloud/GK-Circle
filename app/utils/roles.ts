export type SystemRole = "user" | "moderator" | "admin" | "super_admin";

export function normalizeRoles(
  value: readonly string[] | string | null | undefined
): SystemRole[] {
  const rawRoles = Array.isArray(value)
    ? value
    : typeof value === "string"
    ? value.split(",")
    : [];

  const validRoles = new Set<SystemRole>([
    "user",
    "moderator",
    "admin",
    "super_admin",
  ]);

  const roles = rawRoles
    .map((role) => {
      const trimmed = role.trim().toLowerCase();
      if (trimmed === "admin-user") return "admin";
      if (trimmed === "guest-user") return "user";
      return trimmed;
    })
    .filter((role): role is SystemRole => validRoles.has(role as SystemRole));

  return [...new Set(roles)];
}

export function getHighestRole(roles: readonly SystemRole[]): SystemRole {
  if (roles.includes("super_admin")) return "super_admin";
  if (roles.includes("admin")) return "admin";
  if (roles.includes("moderator")) return "moderator";
  return "user";
}

export function getRoleLabel(roles: readonly SystemRole[]): string {
  const role = getHighestRole(roles);
  const labels: Record<SystemRole, string> = {
    super_admin: "Super Admin",
    admin: "Admin",
    moderator: "Moderator",
    user: "User",
  };
  return labels[role];
}

export function hasAnyRole(
  roles: readonly SystemRole[] | null | undefined,
  requiredRoles: readonly SystemRole[]
): boolean {
  if (!roles) {
    return false;
  }
  return requiredRoles.some((role) => roles.includes(role));
}
