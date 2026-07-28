import { describe, it, expect, vi, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import Option from "~/components/Option.vue";

vi.mock("~/components/CodeBlockComponent.vue", () => ({
  default: {
    template: "<pre>{{ code }}</pre>",
    props: ["code"],
  },
}));

const mountComponent = (props) =>
  mount(Option, {
    props,
  });

describe("Option test", () => {
  afterEach(() => {
    // Ensure each case mounts independently.
  });

  it("renders the order correctly", () => {
    const wrapper = mountComponent({
      order: 1,
      optionsMedia: "text",
      option: "Option A",
    });
    expect(wrapper.text()).toContain("A");
    expect(wrapper.text()).toContain("Option A");
  });

  it("renders image media correctly for fetchable URLs", () => {
    const wrapper = mountComponent({
      order: 2,
      optionsMedia: "image",
      option: "https://example.com/path/to/image.jpg",
    });
    const img = wrapper.find("img");
    expect(img.exists()).toBe(true);
    expect(img.attributes("src")).toBe("https://example.com/path/to/image.jpg");
    expect(img.attributes("alt")).toBe("Option 2");
  });

  it("does not render a raw object key as an image", () => {
    const wrapper = mountComponent({
      order: 2,
      optionsMedia: "image",
      option: "/path/to/image.jpg",
    });
    expect(wrapper.find("img").exists()).toBe(false);
    expect(wrapper.text()).toContain("/path/to/image.jpg");
  });

  it("renders text media correctly", () => {
    const wrapper = mountComponent({
      order: 1,
      optionsMedia: "text",
      option: "Option Text",
      isCorrect: true,
    });
    expect(wrapper.get('[aria-label="Correct"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Option Text");
  });

  it("renders code media correctly", () => {
    const wrapper = mountComponent({
      order: 1,
      optionsMedia: "code",
      option: "console.log('Hello, world!');",
    });
    const codeBlock = wrapper.find("pre");
    expect(codeBlock.exists()).toBe(true);
    expect(codeBlock.text()).toBe("console.log('Hello, world!');");
  });

  it("renders admin analysis badge when isAdminAnalysis is true", () => {
    const wrapper = mountComponent({
      order: 1,
      optionsMedia: "text",
      option: "Option A",
      isAdminAnalysis: true,
      selected: 12,
      isCorrect: true,
    });
    expect(wrapper.text()).toContain("12");
  });

  it("does not render admin analysis badge when isAdminAnalysis is false", () => {
    const wrapper = mountComponent({
      order: 1,
      optionsMedia: "text",
      option: "Option A",
      isAdminAnalysis: false,
      selected: 12,
    });
    expect(wrapper.text()).not.toMatch(/\b12\b/);
  });
});
