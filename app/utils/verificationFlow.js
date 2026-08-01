const SAFE_NODE_PROTOCOLS = new Set(["http:", "https:"]);

export const valuesFromVerificationFlow = (flow) => {
  const values = {};
  for (const node of flow?.ui?.nodes || []) {
    const name = node?.attributes?.name;
    if (name && node.attributes.value !== undefined) {
      values[name] = node.attributes.value;
    }
  }
  return values;
};

export const replacementVerificationState = (nextFlow) => ({
  flow: nextFlow,
  values: valuesFromVerificationFlow(nextFlow),
});

export const findVerificationSubmitNode = (flow, name, value) =>
  (flow?.ui?.nodes || []).find(
    (node) =>
      node?.type === "input" &&
      node.attributes?.type === "submit" &&
      node.attributes?.name === name &&
      node.attributes?.value === value
  );

export const verificationSubmissionBlocked = ({
  isLoading,
  resending,
  resendSubmission,
}) => Boolean(isLoading || (resending && !resendSubmission));

export const buildVerificationBody = (flow, values, submitNode) => {
  const body = new URLSearchParams();

  for (const node of flow?.ui?.nodes || []) {
    if (node?.type !== "input") continue;
    const { name, value, type, disabled } = node.attributes || {};
    if (!name || disabled || type === "submit") continue;

    const currentValue = Object.prototype.hasOwnProperty.call(values, name)
      ? values[name]
      : value;
    if (currentValue !== undefined && currentValue !== null) {
      body.set(name, String(currentValue));
    }
  }

  const submitName = submitNode?.attributes?.name;
  if (submitName) {
    body.set(submitName, String(submitNode.attributes.value ?? ""));
  }

  return body;
};

export const configuredRedirectOrigins = ({
  currentOrigin,
  baseUrl,
  configuredOrigins,
}) => {
  const origins = new Set();
  const candidates = [
    currentOrigin,
    baseUrl,
    ...(configuredOrigins || "").split(","),
  ];

  for (const candidate of candidates) {
    const value = String(candidate || "").trim();
    if (!value) continue;
    try {
      origins.add(new URL(value, currentOrigin).origin);
    } catch {
      // Invalid configured values are excluded rather than trusted.
    }
  }

  return origins;
};

export const resolveAllowedRedirect = (
  location,
  currentUrl,
  allowedOrigins
) => {
  if (!location) return null;
  try {
    const resolved = new URL(location, currentUrl);
    return allowedOrigins.has(resolved.origin) ? resolved.href : null;
  } catch {
    return null;
  }
};

export const safeNodeUrl = (value, currentOrigin) => {
  try {
    const url = new URL(String(value || ""), currentOrigin);
    return SAFE_NODE_PROTOCOLS.has(url.protocol) ? url.href : "";
  } catch {
    return "";
  }
};

export const retryAfterSeconds = (value, now = Date.now()) => {
  if (!value) return null;
  const seconds = Number(value);
  if (Number.isFinite(seconds) && seconds >= 0) return Math.ceil(seconds);

  const retryAt = Date.parse(value);
  if (!Number.isFinite(retryAt)) return null;
  return Math.max(0, Math.ceil((retryAt - now) / 1000));
};
