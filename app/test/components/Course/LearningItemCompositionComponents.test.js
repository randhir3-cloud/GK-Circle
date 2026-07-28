import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import CourseNodeSelector from "@/components/course/CourseNodeSelector.vue";
import CourseSelector from "@/components/course/CourseSelector.vue";
import LearningItemDeleteDialog from "@/components/course/LearningItemDeleteDialog.vue";
import LearningItemTable from "@/components/course/LearningItemTable.vue";

const ModalStub = {
  props: ["modelValue", "title", "description", "closeOnBackdrop", "hideClose"],
  emits: ["update:modelValue"],
  template: '<div v-if="modelValue" role="dialog"><slot /></div>',
};

describe("Course LearningItem composition components", () => {
  it("renders Course loading and error states and emits explicit selection", async () => {
    const wrapper = mount(CourseSelector, {
      props: {
        courses: [
          { id: "course-z", title: "Second returned" },
          { id: "course-a", title: "First alphabetically" },
        ],
        loading: true,
        error: "Course request failed",
      },
    });

    const selector = wrapper.get("select");
    expect(selector.attributes("disabled")).toBeDefined();
    expect(selector.text()).toContain("Loading Courses");
    expect(wrapper.get('[role="alert"]').text()).toBe("Course request failed");

    await wrapper.setProps({ loading: false });
    await selector.setValue("course-a");

    expect(
      wrapper
        .findAll("option")
        .slice(1)
        .map((option) => option.text())
    ).toEqual(["Second returned", "First alphabetically"]);
    expect(wrapper.emitted("update:modelValue")).toEqual([["course-a"]]);
  });

  it("renders node levels in API order and emits the selected level and ID", async () => {
    const wrapper = mount(CourseNodeSelector, {
      props: {
        levels: [
          {
            parentId: null,
            selectedId: "root-z",
            nodes: [
              { id: "root-z", title: "Returned first", node_type: "SUBJECT" },
              { id: "root-a", title: "Returned second", node_type: "TOPIC" },
            ],
            loading: false,
          },
          {
            parentId: "root-z",
            selectedId: "",
            nodes: [{ id: "deep-2", title: "Deep node", node_type: "SECTION" }],
            loading: false,
          },
        ],
        error: "Child nodes unavailable",
      },
    });

    const selectors = wrapper.findAll("select");
    expect(selectors).toHaveLength(2);
    expect(
      selectors[0]
        .findAll("option")
        .slice(1)
        .map((option) => option.text())
    ).toEqual(["Returned first · SUBJECT", "Returned second · TOPIC"]);
    expect(wrapper.get(".md\\:grid-cols-2").exists()).toBe(true);
    expect(wrapper.get('[role="alert"]').text()).toBe(
      "Child nodes unavailable"
    );

    await selectors[1].setValue("deep-2");
    expect(wrapper.emitted("select")).toEqual([
      [{ levelIndex: 1, nodeId: "deep-2" }],
    ]);
  });

  it("preserves item order, exposes read-only positions, and forwards row actions", async () => {
    const items = [
      {
        id: "item-z",
        title: "Returned first",
        description: "Description",
        item_type: "VIDEO",
        position: 9,
        publish_state: "PUBLISHED",
      },
      {
        id: "item-a",
        title: "Returned second",
        description: null,
        item_type: "ARTICLE",
        position: 1,
        publish_state: "DRAFT",
      },
    ];
    const wrapper = mount(LearningItemTable, { props: { items } });

    const rows = wrapper.findAll('[data-testid^="learning-item-item-"]');
    expect(rows.map((row) => row.attributes("data-testid"))).toEqual([
      "learning-item-item-z",
      "learning-item-item-a",
    ]);
    expect(rows.map((row) => row.text())).toEqual([
      expect.stringContaining("9"),
      expect.stringContaining("1"),
    ]);
    expect(wrapper.find('input[name="position"]').exists()).toBe(false);
    expect(wrapper.get('[role="table"]').attributes("aria-label")).toBe(
      "Learning Items"
    );

    await wrapper
      .get('button[aria-label="Edit Returned first"]')
      .trigger("click");
    await wrapper
      .get('button[aria-label="Delete Returned second"]')
      .trigger("click");

    expect(wrapper.emitted("edit")).toEqual([[items[0]]]);
    expect(wrapper.emitted("delete")).toEqual([[items[1]]]);
  });

  it("supports accessible delete confirmation, cancellation, errors, and pending state", async () => {
    const wrapper = mount(LearningItemDeleteDialog, {
      props: {
        modelValue: true,
        item: { id: "item-1", title: "Constitution" },
        error: "Delete rejected",
      },
      global: { stubs: { Modal: ModalStub } },
    });

    expect(wrapper.get('[role="dialog"]').exists()).toBe(true);
    expect(wrapper.get('[role="alert"]').text()).toBe("Delete rejected");
    const buttons = wrapper.findAll("button");
    await buttons.find((button) => button.text() === "Cancel").trigger("click");
    await buttons.find((button) => button.text() === "Delete").trigger("click");
    expect(wrapper.emitted("update:modelValue")).toEqual([[false]]);
    expect(wrapper.emitted("confirm")).toHaveLength(1);

    await wrapper.setProps({ deleting: true });
    expect(
      wrapper
        .findAll("button")
        .every((button) => button.attributes("disabled") !== undefined)
    ).toBe(true);
    expect(wrapper.text()).toContain("Deleting");
  });
});
