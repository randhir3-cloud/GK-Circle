export const isSafeContentUrl = (value) => {
  if (
    typeof value !== "string" ||
    value.length === 0 ||
    value !== value.trim()
  ) {
    return false;
  }

  if (value.startsWith("//")) return false;
  if (value.startsWith("/")) return true;

  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
};

export const isExternalContentUrl = (value) =>
  isSafeContentUrl(value) &&
  (value.toLowerCase().startsWith("http://") ||
    value.toLowerCase().startsWith("https://"));
