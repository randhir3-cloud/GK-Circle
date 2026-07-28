import { computed, onBeforeUnmount, ref } from "vue";
import {
  ATTEMPT_STATUS_AUTO_SUBMITTED,
  ATTEMPT_STATUS_SUBMITTED,
  TIMER_CRITICAL_THRESHOLD_SECONDS,
  TIMER_RESYNC_INTERVAL_SECONDS,
  TIMER_WARNING_THRESHOLD_SECONDS,
} from "@/utils/attempt_player_constants";

export const formatTimerDisplay = (seconds) => {
  if (seconds == null || Number.isNaN(Number(seconds))) return "--:--";
  const total = Math.max(0, Math.floor(Number(seconds)));
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const secs = total % 60;

  const pad = (num) => String(num).padStart(2, "0");
  if (hours > 0) {
    return `${hours}:${pad(minutes)}:${pad(secs)}`;
  }
  return `${pad(minutes)}:${pad(secs)}`;
};

export const createAttemptTimer = ({
  expiresAt,
  quizId,
  attemptId,
  api,
  onExpired,
  onTerminalStatus,
}) => {
  if (!expiresAt) {
    return {
      hasDeadline: false,
      remainingSeconds: ref(null),
      formattedTime: computed(() => "Untimed"),
      isExpired: computed(() => false),
      isWarning: computed(() => false),
      isCritical: computed(() => false),
      stop: () => undefined,
      resync: async () => undefined,
    };
  }

  let deadlineTime = new Date(expiresAt).getTime();
  const remainingSeconds = ref(
    Math.max(0, Math.floor((deadlineTime - Date.now()) / 1000))
  );
  let tickTimeoutId = null;
  let resyncIntervalId = null;
  let hasTriggeredExpired = false;
  let isStopped = false;

  const isExpired = computed(
    () => remainingSeconds.value !== null && remainingSeconds.value <= 0
  );
  const isWarning = computed(
    () =>
      remainingSeconds.value !== null &&
      remainingSeconds.value > 0 &&
      remainingSeconds.value <= TIMER_WARNING_THRESHOLD_SECONDS
  );
  const isCritical = computed(
    () =>
      remainingSeconds.value !== null &&
      remainingSeconds.value > 0 &&
      remainingSeconds.value <= TIMER_CRITICAL_THRESHOLD_SECONDS
  );
  const formattedTime = computed(() =>
    formatTimerDisplay(remainingSeconds.value)
  );

  const tick = () => {
    if (isStopped) return;
    const now = Date.now();
    const diff = Math.max(0, Math.floor((deadlineTime - now) / 1000));
    remainingSeconds.value = diff;

    if (diff <= 0) {
      if (!hasTriggeredExpired) {
        hasTriggeredExpired = true;
        onExpired?.();
      }
      return;
    }

    tickTimeoutId = setTimeout(tick, 1000);
  };

  const resync = async () => {
    if (isStopped || !api || !quizId || !attemptId) return;
    try {
      const statusResp = await api.getAttemptStatus(quizId, attemptId);
      if (isStopped || !statusResp) return;

      if (
        statusResp.status === ATTEMPT_STATUS_SUBMITTED ||
        statusResp.status === ATTEMPT_STATUS_AUTO_SUBMITTED
      ) {
        stop();
        onTerminalStatus?.(statusResp.status);
        return;
      }

      if (statusResp.expires_at) {
        deadlineTime = new Date(statusResp.expires_at).getTime();
        tick();
      }
    } catch {
      // Retain last known deadline on network failure
    }
  };

  const handleVisibilityOrFocus = () => {
    if (
      typeof document !== "undefined" &&
      document.visibilityState === "hidden"
    )
      return;
    resync();
  };

  if (typeof window !== "undefined") {
    window.addEventListener("visibilitychange", handleVisibilityOrFocus);
    window.addEventListener("focus", handleVisibilityOrFocus);
    resyncIntervalId = setInterval(
      resync,
      TIMER_RESYNC_INTERVAL_SECONDS * 1000
    );
  }

  const stop = () => {
    if (isStopped) return;
    isStopped = true;
    if (tickTimeoutId) {
      clearTimeout(tickTimeoutId);
      tickTimeoutId = null;
    }
    if (resyncIntervalId) {
      clearInterval(resyncIntervalId);
      resyncIntervalId = null;
    }
    if (typeof window !== "undefined") {
      window.removeEventListener("visibilitychange", handleVisibilityOrFocus);
      window.removeEventListener("focus", handleVisibilityOrFocus);
    }
  };

  try {
    onBeforeUnmount(stop);
  } catch {
    // Utility invoked outside Vue lifecycle setup
  }

  // Start countdown tick
  tick();

  return {
    hasDeadline: true,
    remainingSeconds,
    formattedTime,
    isExpired,
    isWarning,
    isCritical,
    stop,
    resync,
  };
};
