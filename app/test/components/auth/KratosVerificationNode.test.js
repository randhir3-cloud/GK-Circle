import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import KratosVerificationNode from "~/components/auth/KratosVerificationNode.vue";

describe("KratosVerificationNode", () => {
  it("preserves group and safe input attributes", () => {
    const wrapper = mount(KratosVerificationNode, {
      props: {
        node: {
          type: "input",
          group: "code",
          attributes: {
            name: "code",
            type: "text",
            required: true,
            disabled: false,
            autocomplete: "one-time-code",
          },
          meta: { label: { text: "Verification code" } },
          messages: [{ id: 1, text: "Enter the code" }],
        },
        index: 4,
        modelValue: "",
        currentOrigin: "https://gkcircle.com",
      },
    });

    expect(wrapper.attributes("data-kratos-node-index")).toBe("4");
    expect(wrapper.attributes("data-kratos-node-group")).toBe("code");
    expect(wrapper.get("input").attributes()).toMatchObject({
      name: "code",
      type: "text",
      required: "",
      autocomplete: "one-time-code",
    });
    expect(wrapper.text()).toContain("Enter the code");
  });

  it("retains unsupported nodes without executing them", () => {
    const wrapper = mount(KratosVerificationNode, {
      props: {
        node: {
          type: "script",
          group: "default",
          attributes: { src: "javascript:alert(1)" },
          meta: { label: { text: "Unsupported verification node" } },
          messages: [],
        },
        index: 0,
        currentOrigin: "https://gkcircle.com",
      },
    });
    expect(wrapper.find("script").exists()).toBe(false);
    expect(
      wrapper.get('[data-kratos-node-unsupported="true"]').text()
    ).toContain("Unsupported verification node");
  });
});
