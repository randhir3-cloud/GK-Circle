import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, it } from "vitest";
import ConfirmModal from "~/components/utils/ConfirmModal.vue";

const props = {
  modalTitle: "Test Modal",
  modalMessage: "test message",
  modelPositiveMessage: "this is positive message",
  modelNegativeMessage: "this is negative messsage",
};

describe("ConfirmModal test", () => {
  let wrapper;
  let host;

  afterEach(() => {
    wrapper?.unmount();
    host?.remove();
  });

  const mountModal = () => {
    host = document.createElement("div");
    document.body.appendChild(host);
    wrapper = mount(ConfirmModal, {
      props,
      attachTo: host,
    });
    return wrapper;
  };

  it("renders with props", async () => {
    mountModal();
    await wrapper.vm.$nextTick();

    expect(wrapper.props("modalTitle")).toBe("Test Modal");
    expect(wrapper.props("modalMessage")).toBe("test message");

    const dialog = document.body.querySelector('[role="dialog"]');
    expect(dialog).toBeTruthy();
    expect(dialog.getAttribute("aria-modal")).toBe("true");
    expect(dialog.querySelector("#confirmModalLabel")?.textContent).toContain(
      "Test Modal"
    );
    expect(dialog.textContent).toContain("test message");
  });

  it("emits confirmMessage event on positive button click", async () => {
    mountModal();
    await wrapper.vm.$nextTick();

    const buttons = Array.from(document.body.querySelectorAll("button"));
    const positive = buttons.find(
      (button) => button.textContent?.trim() === props.modelPositiveMessage
    );
    expect(positive).toBeTruthy();
    await positive.click();
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted("confirmMessage")).toBeTruthy();
    expect(wrapper.emitted("confirmMessage")[0]).toEqual([true]);
  });
});
