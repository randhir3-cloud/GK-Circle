import { mountSuspended } from "@nuxt/test-utils/runtime";
import { describe, expect, it } from "vitest";
import HeroSection from "@/components/landing/HeroSection.vue";
import EducationalPlatformFeatures from "@/components/landing/EducationalPlatformFeatures.vue";

describe("GK Circle Educational Operating System homepage", () => {
  it("uses the GK Circle learning hero and removes copied deployment copy", async () => {
    const wrapper = await mountSuspended(HeroSection);

    expect(wrapper.text()).toContain("Learn.");
    expect(wrapper.text()).toContain("Practice.");
    expect(wrapper.text()).toContain("Compete.");
    expect(wrapper.text()).toContain("Grow.");
    expect(wrapper.text()).toContain("Explore Courses");
    expect(wrapper.text()).toContain("AI Assistant");
    expect(wrapper.text()).not.toContain("Open-Source Quizzing");
    expect(wrapper.text()).not.toContain("Demo Instance");
    expect(wrapper.text()).not.toContain("100 concurrent");
    expect(wrapper.text()).not.toContain("Deploy Your Own");
  });

  it("presents all six connected educational capabilities", async () => {
    const wrapper = await mountSuspended(EducationalPlatformFeatures);

    for (const title of [
      "Structured Courses",
      "Smart MCQ Practice",
      "Live Quiz & Gamification",
      "Analytics",
      "AI Assistant",
      "Community",
    ]) {
      expect(wrapper.text()).toContain(title);
    }
    expect(wrapper.text()).toContain("Weak topic detection");
    expect(wrapper.text()).toContain("RAG powered");
    expect(wrapper.text()).toContain("Study circles");
  });
});
