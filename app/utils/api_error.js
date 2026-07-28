const errorStatus = (error) =>
  Number(error?.statusCode || error?.status || error?.response?.status || 0);

const errorText = (error) =>
  [error?.data?.data, error?.data?.message, error?.message]
    .filter((value) => typeof value === "string")
    .join(" ")
    .toLowerCase();

export const isAuthenticationError = (error) => {
  const status = errorStatus(error);
  if (status === 401 || status === 403) return true;

  const message = errorText(error);
  return (
    message.includes("authentication required") ||
    message.includes("unauthenticated") ||
    (message.includes("session") && message.includes("cookie")) ||
    message.includes("user identity")
  );
};

export const getSafeAPIErrorMessage = (
  error,
  fallback = "Something went wrong. Please try again."
) => {
  if (isAuthenticationError(error)) return "";
  return fallback;
};
