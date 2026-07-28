import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import PageStateCard from "@/components/common/PageStateCard.vue";

describe("PageStateCard", () => {
  it("uses the GK Circle yellow-and-ink treatment for its eyebrow", () => {
    const wrapper = mount(PageStateCard, {
      props: {
        eyebrow: "Your learning dashboard",
        title: "Please sign in",
        description: "Sign in to continue.",
      },
      global: {
        stubs: {
          NuxtLink: true,
        },
      },
    });

    const eyebrow = wrapper.get("p");
    expect(eyebrow.classes()).toContain("bg-jv-yellow-soft");
    expect(eyebrow.classes()).toContain("text-jv-ink");
    expect(eyebrow.classes()).not.toContain("text-jv-coral");
  });
});
