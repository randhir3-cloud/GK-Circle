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
            <!-- Loop through input nodes that are not submit or hidden -->
            <div
              v-for="node in inputFields"
              :key="node.attributes.name"
              class="flex flex-col gap-1.5"
            >
              <label
                :for="node.attributes.name"
                class="px-0.5 font-body text-xs font-bold uppercase tracking-wide text-jv-ink sm:text-[13px]"
              >
                {{ node.meta?.label?.text || node.attributes.name }}
              </label>

              <div
                class="jv-card flex items-center gap-2.5 border-2 border-jv-ink bg-jv-white px-3 py-2.5 shadow-brutal-sm transition-all focus-within:translate-x-[1px] focus-within:translate-y-[1px] focus-within:shadow-none"
              >
                <Mail
                  v-if="node.attributes.name === 'email'"
                  class="size-[18px] shrink-0 text-jv-ink/70"
                  :stroke-width="2.2"
                />
                <ShieldCheck
                  v-else-if="node.attributes.name === 'code'"
                  class="size-[18px] shrink-0 text-jv-ink/70"
                  :stroke-width="2.2"
                />
                <input
                  :id="node.attributes.name"
                  v-model="formValues[node.attributes.name]"
                  :name="node.attributes.name"
                  :type="node.attributes.type"
                  :required="node.attributes.required"
                  :disabled="node.attributes.disabled"
                  :autocomplete="node.attributes.autocomplete"
                  :maxlength="node.attributes.name === 'code' ? 6 : undefined"
                  class="min-w-0 flex-1 border-0 bg-transparent font-body text-sm text-jv-ink outline-none placeholder:text-jv-ink/40 sm:text-base"
                  :class="{
                    'tracking-[0.2em]': node.attributes.name === 'code',
                  }"
                  :placeholder="
                    node.attributes.name === 'code'
                      ? '123456'
                      : 'you@example.com'
                  "
                />
              </div>

              <!-- Node specific messages -->
              <p
                v-for="(msg, index) in node.messages"
                :key="index"
                class="font-body text-jv-accent-red text-xs px-0.5 m-0 flex items-center gap-1"
              >
                <AlertCircle class="size-3.5 shrink-0" :stroke-width="2.2" />
                {{ msg.text }}
              </p>
            </div>

            <!-- Submit buttons rendered dynamically -->
            <button
              v-for="node in submitFields"
              :key="node.attributes.name + node.attributes.value"
              type="submit"
              :disabled="node.attributes.disabled || isLoading"
              class="jv-card mt-2 inline-flex h-12 items-center justify-center gap-2 border-2 border-jv-ink bg-jv-coral font-headings text-base text-white shadow-brutal-sm transition-transform hover:rotate-[1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none sm:text-lg disabled:opacity-70 disabled:cursor-not-allowed"
              @click="
                setSubmitAction(node.attributes.name, node.attributes.value)
              "
            >
              {{ node.meta?.label?.text || "Submit" }}
              <ArrowRight class="size-5" :stroke-width="2.4" />
            </button>
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
              @click="initiateVerificationFlow"
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
import { ShieldCheck, ArrowRight, Mail, AlertCircle } from "lucide-vue-next";
import { ref, computed, onMounted } from "vue";

definePageMeta({
  layout: "auth",
});

useSeoMeta({
  title: "Verify Account - GK Circle",
  description:
    "Verify your GK Circle account email to activate quiz hosting and account features.",
  robots: "noindex, nofollow",
});

const { kratosUrl } = useRuntimeConfig().public;
const route = useRoute("");
const toast = usePush();

const flowData = ref(null);
const formValues = ref({});
const isLoading = ref(false);
const flowActive = ref(false);
const resending = ref(false);
const resendSuccess = ref(false);
const rateLimitMsg = ref("");

// Track clicked submit node name & value
const activeSubmitName = ref("");
const activeSubmitValue = ref("");

const setSubmitAction = (name, value) => {
  activeSubmitName.value = name;
  activeSubmitValue.value = value;
};

// Filter out standard fields for rendering
const inputFields = computed(() => {
  if (!flowData.value?.ui?.nodes) return [];
  return flowData.value.ui.nodes.filter(
    (node) =>
      node.type === "input" &&
      node.attributes.type !== "submit" &&
      node.attributes.type !== "hidden"
  );
});

// Filter hidden fields
const hiddenFields = computed(() => {
  if (!flowData.value?.ui?.nodes) return [];
  return flowData.value.ui.nodes.filter(
    (node) => node.type === "input" && node.attributes.type === "hidden"
  );
});

// Filter submit fields
const submitFields = computed(() => {
  if (!flowData.value?.ui?.nodes) return [];
  return flowData.value.ui.nodes.filter(
    (node) => node.type === "input" && node.attributes.type === "submit"
  );
});

onMounted(async () => {
  if (route.query.flow) {
    await fetchFlowIdAndCsrfToken();
  }
});

const fetchFlowIdAndCsrfToken = async () => {
  isLoading.value = true;
  flowActive.value = false;
  rateLimitMsg.value = "";
  try {
    const response = await fetch(
      `${kratosUrl}/self-service/verification/flows?id=${route.query.flow}`,
      {
        method: "GET",
        headers: {
          Accept: "application/json",
        },
        credentials: "include",
      }
    );

    if (!response.ok) {
      throw new Error(
        `Failed to fetch verification status: ${response.statusText}`
      );
    }

    const data = await response.json();
    flowData.value = data;
    flowActive.value = true;

    // Prepopulate form values
    data?.ui?.nodes?.forEach((node) => {
      if (node.attributes.name && node.attributes.value !== undefined) {
        formValues.value[node.attributes.name] = node.attributes.value;
      }
    });

    // Inspect flow messages
    data?.ui?.messages?.forEach((element) => {
      if (element.type === "error") {
        toast.error(element.text);
        if (
          element.id === 4000007 ||
          element.text.toLowerCase().includes("rate limit") ||
          element.text.toLowerCase().includes("please wait")
        ) {
          rateLimitMsg.value = element.text;
        }
      } else {
        toast.success(element.text);
        if (
          data.state === "sent_email" ||
          element.text.toLowerCase().includes("sent")
        ) {
          resendSuccess.value = true;
        }
      }
    });

    if (data.state === "passed_challenge") {
      toast.success("Verification successful! Redirecting to login...");
      setTimeout(() => {
        navigateTo("/account/login");
      }, 1000);
    }
  } catch (error) {
    console.error("Error fetching flow ID and CSRF token:", error.message);
    flowActive.value = false;
  } finally {
    isLoading.value = false;
  }
};

const initiateVerificationFlow = () => {
  resending.value = true;
  window.location.href = `${kratosUrl}/self-service/verification/browser`;
};

const handleSubmit = async () => {
  if (!flowData.value) return;
  isLoading.value = true;
  rateLimitMsg.value = "";

  // Build request body with all form values including hidden csrf token
  const body = { ...formValues.value };

  // Inject hidden nodes value specifically if missing
  hiddenFields.value.forEach((node) => {
    body[node.attributes.name] = node.attributes.value;
  });

  if (activeSubmitName.value) {
    body[activeSubmitName.value] = activeSubmitValue.value;
  }

  try {
    const actionUrl = flowData.value.ui.action;
    const method = flowData.value.ui.method || "POST";

    const response = await fetch(actionUrl, {
      method: method,
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      credentials: "include",
      body: JSON.stringify(body),
    });

    if (
      response.status === 400 ||
      response.status === 422 ||
      response.status === 401
    ) {
      // Flow error or update containing refreshed UI nodes/messages
      const updatedFlow = await response.json();
      flowData.value = updatedFlow;

      // Update values
      updatedFlow?.ui?.nodes?.forEach((node) => {
        if (node.attributes.name && node.attributes.value !== undefined) {
          formValues.value[node.attributes.name] = node.attributes.value;
        }
      });

      // Handle general flow level messages
      updatedFlow?.ui?.messages?.forEach((msg) => {
        if (msg.type === "error") {
          toast.error(msg.text);
          if (msg.text.toLowerCase().includes("rate limit")) {
            rateLimitMsg.value = msg.text;
          }
        } else {
          toast.success(msg.text);
          if (
            updatedFlow.state === "sent_email" ||
            msg.text.toLowerCase().includes("sent")
          ) {
            resendSuccess.value = true;
          }
        }
      });

      // Handle node-specific messages
      updatedFlow?.ui?.nodes?.forEach((node) => {
        node.messages?.forEach((msg) => {
          if (msg.type === "error") {
            toast.error(
              `${node.meta?.label?.text || node.attributes.name}: ${msg.text}`
            );
          }
        });
      });
      return;
    }

    if (!response.ok) {
      throw new Error(`Submission failed: HTTP ${response.status}`);
    }

    // Process successful flow completion or transition
    const result = await response.json();
    flowData.value = result;

    if (result.state === "passed_challenge" || response.status === 200) {
      toast.success("Verification successful! Redirecting...");
      setTimeout(() => {
        navigateTo("/account/login");
      }, 1000);
    } else {
      resendSuccess.value = true;
      result?.ui?.messages?.forEach((msg) => {
        if (msg.type === "success") {
          toast.success(msg.text);
        }
      });
    }
  } catch (error) {
    console.error("Error submitting verification form:", error);
    toast.error("Failed to complete verification step.");
  } finally {
    isLoading.value = false;
  }
};
</script>
