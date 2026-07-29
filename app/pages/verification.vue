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
                ? "Enter the verification code sent to your email."
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
            Fetching verification status...
          </p>
        </div>

        <!-- No Active Flow State: "Send Verification Code" Button -->
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

        <!-- Flow Active: Code Entry Form -->
        <div v-else class="flex flex-col gap-4">
          <div
            v-if="resendSuccess"
            class="bg-jv-mint/20 border-2 border-jv-mint/80 rounded px-3 py-2 text-sm font-body text-jv-ink/90 text-center mb-1"
          >
            ✨ Code sent! Please check your inbox.
          </div>

          <form
            method="post"
            :action="verificationUrl"
            enctype="application/json"
            class="flex flex-col gap-4"
          >
            <div class="flex flex-col gap-1.5">
              <label
                for="code"
                class="px-0.5 font-body text-xs font-bold uppercase tracking-wide text-jv-ink sm:text-[13px]"
              >
                Verification Code
              </label>
              <div
                class="jv-card flex items-center gap-2.5 border-2 border-jv-ink bg-jv-white px-3 py-2.5 shadow-brutal-sm transition-all focus-within:translate-x-[1px] focus-within:translate-y-[1px] focus-within:shadow-none"
              >
                <ShieldCheck
                  class="size-[18px] shrink-0 text-jv-ink/70"
                  :stroke-width="2.2"
                />
                <input
                  id="code"
                  v-model="code"
                  name="code"
                  type="text"
                  maxlength="6"
                  class="min-w-0 flex-1 border-0 bg-transparent font-body text-sm tracking-[0.2em] text-jv-ink outline-none placeholder:text-jv-ink/40 sm:text-base"
                  placeholder="123456"
                  required
                />
              </div>
            </div>

            <input type="hidden" name="method" value="code" />
            <input type="hidden" name="csrf_token" :value="csrfToken" />

            <button
              type="submit"
              class="jv-card mt-2 inline-flex h-12 items-center justify-center gap-2 border-2 border-jv-ink bg-jv-coral font-headings text-base text-white shadow-brutal-sm transition-transform hover:rotate-[1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none sm:text-lg"
            >
              Verify
              <ArrowRight class="size-5" :stroke-width="2.4" />
            </button>
          </form>

          <!-- Resend Code Button -->
          <div
            class="mt-4 border-t-2 border-jv-ink/10 pt-4 flex flex-col items-center"
          >
            <p
              v-if="rateLimitMsg"
              class="text-xs font-body text-jv-coral font-bold text-center mb-2"
            >
              ⚠️ {{ rateLimitMsg }}
            </p>
            <button
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
import { ShieldCheck, ArrowRight } from "lucide-vue-next";
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

const code = ref("");
const csrfToken = ref("");
const isLoading = ref(false);
const flowActive = ref(false);
const resending = ref(false);
const resendSuccess = ref(false);
const rateLimitMsg = ref("");

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
      throw new Error(`Failed to fetch CSRF token: ${response.statusText}`);
    }

    const data = await response.json();

    // Find CSRF token
    const tokenNode = data?.ui?.nodes?.find(
      (node) => node.attributes.name === "csrf_token"
    );
    if (tokenNode) {
      csrfToken.value = tokenNode.attributes.value;
      flowActive.value = true;
    }

    // Inspect Kratos system messages for errors / rate limiting
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
  // Use full browser navigation for Kratos flow initiation to preserve cookies and CSRF context correctly
  window.location.href = `${kratosUrl}/self-service/verification/browser`;
};

const verificationUrl = computed(
  () =>
    `${kratosUrl}/self-service/verification?token=${code.value}&flow=${route.query.flow}`
);
</script>
