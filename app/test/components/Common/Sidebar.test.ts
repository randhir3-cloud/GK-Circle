import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import Sidebar from "~/components/common/Sidebar.vue";
import { useUsersStore } from "~/store/users";
import { createTestingPinia } from "@pinia/testing";
import { mockNuxtImport } from "@nuxt/test-utils/runtime";

mockNuxtImport("useRoute", () => () => ({
  path: "/",
  hash: "",
}));

describe("Sidebar Component test", () => {
  let pinia: ReturnType<typeof createTestingPinia>;
  let usersStore: ReturnType<typeof useUsersStore>;

  beforeEach(() => {
    pinia = createTestingPinia({
      createSpy: vi.fn,
    });
    usersStore = useUsersStore(pinia);
    vi.resetAllMocks();
  });

  const mountComponent = () => {
    return mount(Sidebar, {
      global: {
        plugins: [pinia],
        stubs: {
          NuxtLink: {
            template: "<a><slot /></a>",
          },
          NavigationLink: {
            template: "<a><slot /></a>",
          },
        },
      },
    });
  };

  it("renders role label correctly for super_admin", async () => {
    usersStore.getUserData = vi.fn().mockReturnValue({
      role: "admin-user",
      roles: ["super_admin"],
      firstname: "Super",
      username: "admin",
      avatar: "Sophia",
      canCreatePublicQuiz: true,
    });

    const wrapper = mountComponent();
    // Set mounted to true to trigger render of elements within template v-if="mounted"
    await wrapper.setData({ mounted: true });

    const roleBadge = wrapper.find('[data-testid="sidebar-user-role"]');
    expect(roleBadge.text()).toBe("Super Admin");
  });

  it("renders role label correctly for user", async () => {
    usersStore.getUserData = vi.fn().mockReturnValue({
      role: "guest-user",
      roles: ["user"],
      firstname: "Normal",
      username: "user",
      avatar: "Eden",
      canCreatePublicQuiz: false,
    });

    const wrapper = mountComponent();
    await wrapper.setData({ mounted: true });

    // Since showAdminNav will be false for user, it won't render showAdminNav block with the role label.
    // Instead we assert showAdminNav is false.
    const showAdminNav = wrapper.vm.showAdminNav;
    expect(showAdminNav).toBe(false);
  });
});
