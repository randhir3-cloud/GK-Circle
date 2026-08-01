<script setup>
import { ref, computed, onMounted } from "vue";
import ErrorLayout from "@/error.vue";

definePageMeta({
  layout: "empty",
});

const route = useRoute();
const config = useRuntimeConfig();
const url = config.public;

const errorDetails = ref(null);
const loading = ref(false);

const errorId = computed(() => route.query.id || route.query.error);

onMounted(async () => {
  if (errorId.value) {
    loading.value = true;
    try {
      const response = await fetch(
        `${url.kratosUrl}/self-service/errors?error=${errorId.value}`,
        {
          method: "GET",
          headers: {
            Accept: "application/json",
          },
          credentials: "include",
        }
      );
      if (response.ok) {
        const data = await response.json();
        const kratosError = data?.errors?.[0] || data?.error;
        if (kratosError) {
          // Strict allowlist mapping
          errorDetails.value = {
            id: data.id || errorId.value,
            code: kratosError.code || "",
            status: kratosError.status || "",
            reason: kratosError.reason || "",
            message: kratosError.message || "",
          };
        }
      }
    } catch (err) {
      console.error("Error fetching Kratos error container details:", err);
    } finally {
      loading.value = false;
    }
  }
});

const errorObj = computed(() => {
  if (loading.value) {
    return {
      statusCode: 500,
      statusMessage: "Loading error details...",
      message: "Retrieving secure diagnostic logs...",
    };
  }

  if (errorId.value) {
    if (errorDetails.value) {
      return {
        statusCode: errorDetails.value.status || 500,
        statusMessage:
          errorDetails.value.reason || "Authentication Service Error",
        message:
          errorDetails.value.message ||
          "An unexpected authentication error occurred.",
      };
    } else {
      return {
        statusCode: 500,
        statusMessage: "Authentication service error",
        message: `Reference: ${errorId.value}`,
      };
    }
  }

  return {
    statusCode: route.query.statusCode || 404,
    statusMessage: route.query.status || "Page or resource not found",
    message: route.query.error || route.query.message || "",
  };
});
</script>

<template>
  <ErrorLayout :error="errorObj" />
</template>
