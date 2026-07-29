import { useUsersStore } from "~~/store/users";

export default defineNuxtRouteMiddleware(async (to) => {
  const isProtected =
    to.meta.requiresVerifiedEmail === true ||
    to.path.startsWith("/admin") ||
    to.path.startsWith("/instructor");

  if (!isProtected) {
    return;
  }

  const usersStore = useUsersStore();

  // On a fresh browser load, the Pinia store might be empty even if a valid
  // Kratos session cookie is present. Hydrate the session first before redirecting.
  let user = usersStore.getUserData();
  if (!user) {
    try {
      await usersStore.fetchAuthenticatedUser();
      user = usersStore.getUserData();
    } catch (error) {
      console.warn("Middleware failed to fetch authenticated user:", error);
      return navigateTo("/account/login");
    }
  }

  // If still not authenticated, send to login page.
  if (!user) {
    return navigateTo("/account/login");
  }

  // Authenticated but email not verified — fail closed (redirect to verification)
  if (user.emailVerified !== true) {
    return navigateTo("/verification");
  }
});
