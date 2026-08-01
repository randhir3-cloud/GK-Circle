<template>
  <div class="relative z-10 mx-auto w-full max-w-[440px]">
    <div class="relative rotate-1">
      <span
        class="jv-card absolute -top-[8px] left-1/2 z-20 h-3 w-12 -translate-x-1/2 border-2 border-jv-ink bg-jv-slate shadow-brutal-sm"
        aria-hidden="true"
      ></span>

      <div
        class="jv-card border-2 border-jv-ink bg-jv-white px-6 py-7 shadow-brutal-lg sm:px-8 sm:py-9"
      >
        <header class="mb-6 flex flex-col items-center gap-1.5">
          <div class="relative inline-block">
            <h1
              class="m-0 font-headings text-[28px] leading-none text-jv-ink sm:text-[34px]"
            >
              Verify Your Email
            </h1>
            <svg
              class="absolute -bottom-2 left-1/2 -translate-x-1/2 text-jv-mint"
              width="140"
              height="14"
              viewBox="0 0 140 14"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
              aria-hidden="true"
            >
              <path
                d="M3 9 Q 25 1, 50 7 T 95 6 T 137 4"
                stroke="currentColor"
                stroke-width="2.5"
                stroke-linecap="round"
                fill="none"
              />
            </svg>
          </div>
          <p
            class="m-0 mt-3 text-center font-body text-sm text-jv-ink/70 sm:text-base"
          >
            {{
              flowActive
                ? flowData?.state === "sent_email"
                  ? "Enter the verification code sent to your email."
                  : "Please trigger the verification code request."
                : "You need to verify your email address to continue."
            }}
          </p>
        </header>

        <!-- Loading State -->
        <div
          v-if="isLoading"
          class="flex flex-col items-center justify-center py-6 gap-3"
        >
          <div
            class="animate-spin rounded-full h-8 w-8 border-b-2 border-jv-coral"
          ></div>
          <p class="font-body text-sm text-jv-ink/60">
            Processing verification...
          </p>
        </div>

        <!-- No Active Flow State -->
        <div v-else-if="!flowActive" class="flex flex-col gap-4">
          <p class="text-sm font-body text-jv-ink/80 text-center mb-2">
            Click below to generate and send a new verification code to your
            email address.
          </p>
          <button
            :disabled="resending"
            class="jv-card inline-flex h-12 items-center justify-center gap-2 border-2 border-jv-ink bg-jv-coral font-headings text-base text-white shadow-brutal-sm transition-all hover:rotate-[1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none sm:text-lg disabled:opacity-75 disabled:cursor-not-allowed"
            @click="initiateVerificationFlow"
          >
            <span v-if="resending">Sending...</span>
            <span v-else>Send Verification Code</span>
            <ArrowRight v-if="!resending" class="size-5" :stroke-width="2.4" />
          </button>
        </div>

        <!-- Dynamic Form Rendering -->
        <div v-else class="flex flex-col gap-4">
          <div
            v-if="resendSuccess"
            class="bg-jv-mint/20 border-2 border-jv-mint/80 rounded px-3 py-2 text-sm font-body text-jv-ink/90 text-center mb-1"
          >
            ✨ Verification email sent! Please check your inbox.
          </div>

          <form class="flex flex-col gap-4" @submit.prevent="handleSubmit()">
            <KratosVerificationNode
              v-for="(node, index) in flowData.ui.nodes"
              :key="`${flowData.id}:${index}`"
              :node="node"
              :index="index"
              :model-value="formValues[node.attributes?.name]"
              :loading="isLoading || resending"
              :current-origin="currentOrigin"
              @update-value="updateFormValue"
              @submit-node="setSubmitNode"
            />
          </form>

          <!-- Resend Code section mapped from the choose_method / alternate actions -->
          <div
            class="mt-4 border-t-2 border-jv-ink/10 pt-4 flex flex-col items-center"
          >
            <p
              v-if="rateLimitMsg"
              class="text-xs font-body text-jv-coral font-bold text-center mb-2 animate-pulse"
            >
              ⚠️ {{ rateLimitMsg }}
            </p>
            <button
              v-if="flowData?.state === 'sent_email'"
              :disabled="resending || !!rateLimitMsg"
              class="text-sm font-headings text-jv-ink/80 hover:text-jv-coral transition-colors underline underline-offset-4 disabled:opacity-50 disabled:cursor-not-allowed"
              @click="handleResend"
            >
              <span v-if="resending">Requesting new code...</span>
              <span v-else>Resend verification code</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { usePush } from "notivue";
import { ArrowRight } from "lucide-vue-next";
import { computed, onMounted, ref } from "vue";
import { useUsersStore } from "~~/store/users";
import KratosVerificationNode from "~/components/auth/KratosVerificationNode.vue";
import {
  buildVerificationBody,
  configuredRedirectOrigins,
  findVerificationSubmitNode,
  replacementVerificationState,
  resolveAllowedRedirect,
  retryAfterSeconds,
  verificationSubmissionBlocked,
} from "~/utils/verificationFlow";

definePageMeta({ layout: "auth" });

useSeoMeta({
  title: "Verify Account - GK Circle",
  description:
    "Verify your GK Circle account email to activate quiz hosting and account features.",
  robots: "noindex, nofollow",
});

const runtime = useRuntimeConfig().public;
const route = useRoute();
const toast = usePush();
const usersStore = useUsersStore();

const flowData = ref(null);
const formValues = ref({});
const activeSubmitNode = ref(null);
const isLoading = ref(false);
const resending = ref(false);
const resendSuccess = ref(false);
const rateLimitMsg = ref("");
const currentOrigin = ref(new URL(runtime.baseUrl).origin);
const flowActive = computed(() => Boolean(flowData.value?.ui?.nodes));

const allowedRedirectOrigins = () =>
  configuredRedirectOrigins({
    currentOrigin: window.location.origin,
    baseUrl: runtime.baseUrl,
    configuredOrigins: runtime.redirectAllowedOrigins,
  });

const updateFormValue = (name, value) => {
  if (name) formValues.value = { ...formValues.value, [name]: value };
};

const setSubmitNode = (node) => {
  activeSubmitNode.value = node;
};

const processMessages = (flow) => {
  rateLimitMsg.value = "";
  for (const message of flow?.ui?.messages || []) {
    const text = String(message.text || "");
    if (message.type === "error") toast.error(text);
    else toast.success(text);

    const normalized = text.toLowerCase();
    if (
      message.id === 4000007 ||
      normalized.includes("rate limit") ||
      normalized.includes("please wait")
    ) {
      rateLimitMsg.value = text;
    }
  }
};

const replaceFlow = (nextFlow) => {
  const replacement = replacementVerificationState(nextFlow);
  flowData.value = replacement.flow;
  formValues.value = replacement.values;
  activeSubmitNode.value = null;
  resendSuccess.value = nextFlow?.state === "sent_email";
  processMessages(nextFlow);
};

const safeReturnPath = () => {
  const returnTo = Array.isArray(route.query.return_to)
    ? route.query.return_to[0]
    : route.query.return_to;
  const allowed = resolveAllowedRedirect(
    returnTo,
    window.location.href,
    allowedRedirectOrigins()
  );
  if (!allowed) return "/";
  const resolved = new URL(allowed);
  return `${resolved.pathname}${resolved.search}${resolved.hash}`;
};

const finishVerification = async () => {
  await usersStore.fetchAuthenticatedUser();
  const user = usersStore.getUserData();
  toast.success("Verification successful.");
  await navigateTo(
    user?.emailVerified === true ? safeReturnPath() : "/account/login"
  );
};

const initiateVerificationFlow = () => {
  resending.value = true;
  window.location.assign(
    `${runtime.kratosUrl}/self-service/verification/browser`
  );
};

const fetchFlow = async () => {
  if (!route.query.flow) return;
  isLoading.value = true;
  try {
    const response = await fetch(
      `${
        runtime.kratosUrl
      }/self-service/verification/flows?id=${encodeURIComponent(
        String(route.query.flow)
      )}`,
      { headers: { Accept: "application/json" }, credentials: "include" }
    );
    if (response.status === 410) return initiateVerificationFlow();
    if (!response.ok) throw new Error(`flow_fetch_${response.status}`);
    const nextFlow = await response.json();
    replaceFlow(nextFlow);
    if (nextFlow.state === "passed_challenge") await finishVerification();
  } catch (error) {
    console.warn("Verification flow could not be loaded.", {
      cause: error?.message || "request_failed",
    });
    toast.error("Unable to load the verification flow.");
  } finally {
    isLoading.value = false;
  }
};

const handleSubmit = async (
  submitNode = activeSubmitNode.value,
  resendSubmission = false
) => {
  if (
    !flowData.value ||
    verificationSubmissionBlocked({
      isLoading: isLoading.value,
      resending: resending.value,
      resendSubmission,
    })
  )
    return;
  isLoading.value = true;
  rateLimitMsg.value = "";

  try {
    const action = new URL(flowData.value.ui.action, window.location.href);
    const actionOrigins = new Set(allowedRedirectOrigins());
    actionOrigins.add(new URL(runtime.kratosUrl).origin);
    if (!actionOrigins.has(action.origin))
      throw new Error("action_origin_rejected");

    const response = await fetch(action.href, {
      method: String(flowData.value.ui.method || "POST").toUpperCase(),
      headers: {
        Accept: "application/json",
        "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
      },
      credentials: "include",
      redirect: "manual",
      body: buildVerificationBody(flowData.value, formValues.value, submitNode),
    });

    const retrySeconds = retryAfterSeconds(response.headers.get("Retry-After"));
    if (retrySeconds !== null) {
      rateLimitMsg.value = `Please wait ${retrySeconds} seconds before trying again.`;
    }

    if (response.status === 410) return initiateVerificationFlow();

    const location = response.headers.get("Location");
    if (location || (response.status >= 300 && response.status < 400)) {
      const redirectUrl = resolveAllowedRedirect(
        location,
        window.location.href,
        allowedRedirectOrigins()
      );
      if (!redirectUrl) throw new Error("redirect_origin_rejected");
      window.location.assign(redirectUrl);
      return;
    }

    const contentType = response.headers.get("Content-Type") || "";
    if (!contentType.includes("application/json")) {
      throw new Error(`unexpected_response_${response.status}`);
    }

    const nextFlow = await response.json();
    replaceFlow(nextFlow);
    if (nextFlow.state === "passed_challenge") {
      await finishVerification();
      return;
    }
    if (!response.ok && ![400, 401, 422, 429].includes(response.status)) {
      throw new Error(`submission_${response.status}`);
    }
  } catch (error) {
    console.warn("Verification submission failed.", {
      cause: error?.message || "request_failed",
    });
    toast.error("Failed to complete the verification step.");
  } finally {
    isLoading.value = false;
    resending.value = false;
  }
};

const handleResend = async () => {
  if (resending.value || isLoading.value) return;
  const resendNode = findVerificationSubmitNode(
    flowData.value,
    "method",
    "code"
  );
  if (!resendNode) return initiateVerificationFlow();
  resending.value = true;
  await handleSubmit(resendNode, true);
};

onMounted(async () => {
  currentOrigin.value = window.location.origin;
  // Registration's show-verification hook bypasses the normal API callback.
  // Synchronize the valid Kratos session before protected-route middleware runs.
  await usersStore.fetchAuthenticatedUser().catch(() => null);
  await fetchFlow();
});
</script>
