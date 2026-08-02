import { flushPromises } from "@vue/test-utils";
import { mountSuspended, mockNuxtImport } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Sidebar from "@/components/common/Sidebar.vue";

const sidebarState = vi.hoisted(() => ({
  user: {
    role: "admin-user",
    canCreatePublicQuiz: true,
    firstname: "Admin",
  },
}));

vi.mock("~~/store/users", () => ({
  useUsersStore: () => ({
    userData: sidebarState.user,
    getUserData: () => sidebarState.user,
  }),
}));

vi.mock("@/composables/auth", () => ({
  setUserDataStore: vi.fn(),
  handleLogout: vi.fn(),
}));

mockNuxtImport("useRoute", () => () => ({
  path: "/admin/courses/learning-items",
}));

const NavigationLinkStub = {
  props: ["url", "urlName"],
  template:
    '<a v-if="url" :href="url">{{ urlName }}<slot /></a><button v-else><slot /></button>',
};

beforeEach(() => {
  sidebarState.user = {
    role: "admin-user",
    roles: ["admin"],
    canCreatePublicQuiz: true,
    firstname: "Admin",
  };
  window.matchMedia = vi.fn(() => ({
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }));
});

describe("Sidebar Course Content navigation", () => {
  it("shows the link for users with the existing public-admin capability", async () => {
    const wrapper = await mountSuspended(Sidebar, {
      global: { stubs: { NavigationLink: NavigationLinkStub } },
    });
    await flushPromises();

    expect(
      wrapper.find('a[href="/admin/courses/learning-items"]').exists()
    ).toBe(true);
    expect(wrapper.find('a[href="/admin/courses/list"]').exists()).toBe(true);
  });

  it("hides the link when the capability is absent", async () => {
    sidebarState.user = {
      role: "guest-user",
      roles: ["user"],
      canCreatePublicQuiz: false,
      firstname: "Admin",
    };
    const wrapper = await mountSuspended(Sidebar, {
      global: { stubs: { NavigationLink: NavigationLinkStub } },
    });
    await flushPromises();

    expect(
      wrapper.find('a[href="/admin/courses/learning-items"]').exists()
    ).toBe(false);
    expect(wrapper.find('a[href="/admin/courses/list"]').exists()).toBe(false);
  });

  it("displays the correct role badge for super_admin", async () => {
    sidebarState.user = {
      role: "admin-user",
      roles: ["super_admin"],
      firstname: "Randhir",
    };
    const wrapper = await mountSuspended(Sidebar, {
      global: { stubs: { NavigationLink: NavigationLinkStub } },
    });
    await flushPromises();

    const roleBadge = wrapper.find('[data-testid="sidebar-user-role"]');
    expect(roleBadge.exists()).toBe(true);
    expect(roleBadge.text()).toBe("Super Admin");
  });
});
