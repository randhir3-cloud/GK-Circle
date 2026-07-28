import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import LearningItemEditorDialog from "@/components/course/LearningItemEditorDialog.vue";

const ModalStub = {
  props: ["modelValue", "title"],
  emits: ["update:modelValue"],
  template: '<div v-if="modelValue"><slot /></div>',
};

const mountDialog = (props = {}) =>
  mount(LearningItemEditorDialog, {
    props: { modelValue: true, mode: "create", ...props },
    global: {
      stubs: {
        Modal: ModalStub,
        NuxtLink: {
          props: ["to"],
          template: '<a :href="to"><slot /></a>',
        },
      },
    },
  });

describe("LearningItemEditorDialog", () => {
  it("offers every Learning Item type and starts with a required Article block", () => {
    const wrapper = mountDialog();

    expect(
      wrapper
        .get('select[name="item_type"]')
        .findAll("option")
        .map((option) => option.attributes("value"))
    ).toEqual(["ARTICLE", "VIDEO", "PDF", "LINK", "QUIZ_REFERENCE"]);
    expect(wrapper.get('[data-testid="content-block-editor"]').exists()).toBe(
      true
    );
    expect(
      wrapper.get('textarea[name="block_0_text"]').attributes("required")
    ).toBeDefined();
    expect(
      wrapper.get('button[type="submit"]').attributes("disabled")
    ).toBeDefined();
  });

  it("creates an Article with persisted content metadata", async () => {
    const wrapper = mountDialog();

    await wrapper.get('input[name="title"]').setValue("Constitution");
    await wrapper
      .get('textarea[name="block_0_text"]')
      .setValue("India is a union of states.");
    await wrapper.get("form").trigger("submit");

    expect(wrapper.emitted("save")[0][0]).toEqual({
      title: "Constitution",
      item_type: "ARTICLE",
      publish_state: "DRAFT",
      quiz_id: null,
      metadata: {
        version: 1,
        blocks: [
          {
            id: expect.any(String),
            type: "TEXT",
            data: { text: "India is a union of states." },
            visibility: { mode: "ALL" },
          },
        ],
      },
    });
  });

  it("switches a blank new item to type-specific Video and PDF fields", async () => {
    const wrapper = mountDialog();

    await wrapper.get('select[name="item_type"]').setValue("VIDEO");
    expect(wrapper.find('textarea[name="block_0_text"]').exists()).toBe(false);
    expect(wrapper.get('input[name="block_0_url"]').exists()).toBe(true);
    expect(wrapper.get('input[name="block_0_title"]').exists()).toBe(true);

    await wrapper.get('select[name="item_type"]').setValue("PDF");
    expect(
      wrapper.get('input[name="block_0_url"]').attributes("placeholder")
    ).toContain("document.pdf");
  });

  it("updates existing block content without losing its ID or visibility", async () => {
    const wrapper = mountDialog({
      mode: "update",
      item: {
        title: "Introduction",
        item_type: "ARTICLE",
        description: "Old",
        publish_state: "DRAFT",
        metadata: {
          version: 1,
          blocks: [
            {
              id: "intro-text",
              type: "TEXT",
              data: { text: "Old content" },
              visibility: { mode: "AUTHENTICATED" },
            },
          ],
        },
      },
    });

    await wrapper
      .get('textarea[name="block_0_text"]')
      .setValue("Updated content");
    await wrapper.get('textarea[name="description"]').setValue("");
    await wrapper.get("form").trigger("submit");

    expect(wrapper.emitted("save")[0][0]).toMatchObject({
      title: "Introduction",
      description: null,
      metadata: {
        version: 1,
        blocks: [
          {
            id: "intro-text",
            type: "TEXT",
            data: { text: "Updated content" },
            visibility: { mode: "AUTHENTICATED" },
          },
        ],
      },
    });
  });

  it("requires and serializes a validated Test selection", async () => {
    const wrapper = mountDialog({
      quizzes: [{ id: "quiz-1", title: "Indian Polity Test" }],
    });

    await wrapper.get('input[name="title"]').setValue("Topic test");
    await wrapper.get('select[name="item_type"]').setValue("QUIZ_REFERENCE");
    expect(wrapper.get('[data-testid="quiz-reference-editor"]').exists()).toBe(
      true
    );
    expect(
      wrapper.get('button[type="submit"]').attributes("disabled")
    ).toBeDefined();

    await wrapper.get('select[name="quiz_id"]').setValue("quiz-1");
    await wrapper.get("form").trigger("submit");

    expect(wrapper.emitted("save")[0][0]).toMatchObject({
      item_type: "QUIZ_REFERENCE",
      quiz_id: "quiz-1",
      metadata: {
        version: 1,
        blocks: [
          {
            type: "LINK",
            data: {
              quiz_id: "quiz-1",
              label: "Indian Polity Test",
            },
          },
        ],
      },
    });
  });
});
