export type RuntimeEnvironment = Record<string, string | undefined>;

export function isProductionRuntime(env: RuntimeEnvironment): boolean {
  const appEnvironment = env.APP_ENV?.trim().toLowerCase();
  const railwayEnvironment = env.RAILWAY_ENVIRONMENT?.trim().toLowerCase();

  if (appEnvironment) {
    return appEnvironment === "production";
  }

  return (
    railwayEnvironment === "production" ||
    env.NODE_ENV === "production" ||
    env.NUXT_PUBLIC_BASE_URL?.includes("gkcircle.com") === true
  );
}
