import { useUsersStore } from "~~/store/users";
import { normalizeRoles, hasAnyRole } from "~/utils/roles";

export default defineNuxtRouteMiddleware(async (to) => {
  const requiredRoles = to.meta.requiredRoles as string[] | undefined;
  if (!requiredRoles?.length) {
    return;
  }

  const usersStore = useUsersStore();
  let user = usersStore.getUserData();

  // On a fresh browser load, hydrate the session first before redirecting.
  if (!user) {
    try {
      await usersStore.fetchAuthenticatedUser();
      user = usersStore.getUserData();
    } catch (error) {
      console.warn("Middleware failed to fetch authenticated user:", error);
      return navigateTo("/account/login");
    }
  }

  if (!user) {
    return navigateTo("/account/login");
  }

  const roles = normalizeRoles(user.roles);
  if (!hasAnyRole(roles, requiredRoles)) {
    // If not authorized, redirect to an unauthorized page or dashboard
    return navigateTo("/admin/quiz/list-quiz");
  }
});
