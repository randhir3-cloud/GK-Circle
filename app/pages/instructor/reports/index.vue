<template>
  <div class="space-y-6">
    <!-- Header Controls -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">
          Instructor Reports & Analytics Export
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          Download snapshot reports in CSV, XLSX, or PDF formats, or automate
          scheduled reporting.
        </p>
      </div>
      <div class="flex items-center space-x-3">
        <button
          class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg shadow-sm flex items-center"
          @click="isExportDialogOpen = true"
        >
          <span>New One-Time Export</span>
        </button>
      </div>
    </div>

    <!-- Active Job Status Tracker -->
    <ReportJobProgress
      v-if="activeJob"
      :job="activeJob"
      :download-url="
        activeJob.status === 'COMPLETED' ? getDownloadUrl(activeJob.id) : ''
      "
    />

    <!-- Navigation Tabs -->
    <div class="border-b border-slate-200 dark:border-slate-700">
      <nav class="-mb-px flex space-x-8">
        <button
          v-for="tab in ['history', 'schedules', 'audit']"
          :key="tab"
          :class="[
            'pb-4 px-1 border-b-2 font-medium text-sm capitalize transition-colors',
            activeTab === tab
              ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400'
              : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300 dark:text-slate-400',
          ]"
          @click="activeTab = tab"
        >
          {{ tab }}
        </button>
      </nav>
    </div>

    <!-- Tab 1: Export History -->
    <div
      v-if="activeTab === 'history'"
      class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden"
    >
      <table
        class="w-full text-left text-sm text-slate-600 dark:text-slate-300"
      >
        <thead
          class="bg-slate-50 dark:bg-slate-700/50 text-xs font-semibold uppercase text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-700"
        >
          <tr>
            <th class="px-4 py-3">Title</th>
            <th class="px-4 py-3">Format</th>
            <th class="px-4 py-3">Type</th>
            <th class="px-4 py-3">Status</th>
            <th class="px-4 py-3">Requested At</th>
            <th class="px-4 py-3 text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200 dark:divide-slate-700">
          <tr
            v-for="item in history"
            :key="item.id"
            class="hover:bg-slate-50 dark:hover:bg-slate-700/30"
          >
            <td class="px-4 py-3 font-medium text-slate-900 dark:text-white">
              {{ item.title }}
            </td>
            <td class="px-4 py-3">{{ item.export_format }}</td>
            <td class="px-4 py-3 text-xs">{{ item.export_type }}</td>
            <td class="px-4 py-3">
              <span
                :class="[
                  'px-2 py-0.5 text-xs font-medium rounded-full',
                  item.status === 'COMPLETED'
                    ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300'
                    : item.status === 'FAILED'
                    ? 'bg-rose-100 text-rose-800 dark:bg-rose-950/60 dark:text-rose-300'
                    : 'bg-slate-100 text-slate-800 dark:bg-slate-700 dark:text-slate-300',
                ]"
              >
                {{ item.status }}
              </span>
            </td>
            <td class="px-4 py-3 text-xs text-slate-500">
              {{ formatDate(item.queued_at) }}
            </td>
            <td class="px-4 py-3 text-right space-x-2">
              <a
                v-if="item.status === 'COMPLETED' && !item.deleted_at"
                :href="getDownloadUrl(item.id)"
                target="_blank"
                download
                class="text-xs font-medium text-indigo-600 hover:text-indigo-800 dark:text-indigo-400"
              >
                Download
              </a>
              <span
                v-else-if="item.deleted_at"
                class="text-xs text-slate-400 italic"
                >Deleted</span
              >
            </td>
          </tr>
          <tr v-if="history.length === 0">
            <td colspan="6" class="px-4 py-8 text-center text-slate-400">
              No export reports found. Click "New One-Time Export" to start.
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Export Dialog Modal -->
    <ExportDialog
      :is-open="isExportDialogOpen"
      @close="isExportDialogOpen = false"
      @submitted="handleExportSubmitted"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useInstructorReports } from "~/composables/instructor_reports";
import ExportDialog from "~/components/instructor-reports/ExportDialog.vue";
import ReportJobProgress from "~/components/instructor-reports/ReportJobProgress.vue";

const activeTab = ref("history");
const isExportDialogOpen = ref(false);
const history = ref([]);
const activeJob = ref(null);

const { requestExport, getExportStatus, getDownloadUrl, getHistory } =
  useInstructorReports();

const loadHistory = async () => {
  try {
    const res = await getHistory();
    if (res && res.data && res.data.items) {
      history.value = res.data.items;
    }
  } catch (err) {
    console.error("Failed to load report history:", err);
  }
};

const handleExportSubmitted = async (payload) => {
  isExportDialogOpen.value = false;
  try {
    const res = await requestExport(payload);
    if (res && res.data && res.data.id) {
      activeJob.value = { id: res.data.id, status: "QUEUED", ...payload };
      pollJobStatus(res.data.id);
    }
  } catch (err) {
    console.error("Failed to request export:", err);
  }
};

const pollJobStatus = async (jobId) => {
  const timer = setInterval(async () => {
    try {
      const res = await getExportStatus(jobId);
      if (res && res.data) {
        activeJob.value = res.data;
        if (res.data.status === "COMPLETED" || res.data.status === "FAILED") {
          clearInterval(timer);
          loadHistory();
        }
      }
    } catch {
      clearInterval(timer);
    }
  }, 2000);
};

const formatDate = (isoStr) => {
  if (!isoStr) return "";
  return new Date(isoStr).toLocaleString();
};

onMounted(() => {
  loadHistory();
});
</script>
