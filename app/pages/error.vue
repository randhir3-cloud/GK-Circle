<script setup>
import ErrorLayout from "@/error.vue";

definePageMeta({
  layout: "empty",
});

const route = useRoute();

const errorObj = computed(() => {
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
