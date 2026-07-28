import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import LearningItemRenderer from "@/components/learning-items/LearningItemRenderer.vue";

const metadataWith = (blocks) => ({ version: 1, blocks });

const deepFreeze = (value) => {
  if (value && typeof value === "object" && !Object.isFrozen(value)) {
    Object.freeze(value);
    Object.values(value).forEach(deepFreeze);
  }
  return value;
};

describe("LearningItemRenderer", () => {
  it("renders every frozen T14 block type with semantic accessible markup", () => {
    const metadata = metadataWith([
      { id: "heading", type: "HEADING", data: { text: "Overview", level: 3 } },
      { id: "text", type: "TEXT", data: { text: "First line\nSecond line" } },
      {
        id: "image",
        type: "IMAGE",
        data: {
          url: "/images/map.png",
          alt: "Map of India",
          caption: "Administrative map",
        },
      },
      {
        id: "video",
        type: "VIDEO",
        data: {
          url: "https://video.example/embed/1",
          title: "Constitution lesson",
          caption: "Lesson video",
        },
      },
      {
        id: "pdf",
        type: "PDF",
        data: { url: "/files/notes.pdf", title: "Course notes" },
      },
      {
        id: "link",
        type: "LINK",
        data: { url: "https://example.com/reference", label: "Reference" },
      },
      {
        id: "quote",
        type: "QUOTE",
        data: { text: "Knowledge matters.", attribution: "Teacher" },
      },
      {
        id: "callout",
        type: "CALLOUT",
        data: { text: "Remember this." },
      },
      { id: "divider", type: "DIVIDER", data: {} },
    ]);

    const wrapper = mount(LearningItemRenderer, { props: { metadata } });

    expect(wrapper.get('[data-testid="heading-block"]').element.tagName).toBe(
      "H3"
    );
    expect(wrapper.get('[data-testid="text-block"]').text()).toContain(
      "Second line"
    );
    expect(wrapper.get("img").attributes("alt")).toBe("Map of India");
    expect(
      wrapper.get('[data-testid="video-block"] iframe').attributes("title")
    ).toBe("Constitution lesson");
    expect(
      wrapper.get('[data-testid="pdf-block"] iframe').attributes("title")
    ).toBe("Course notes");
    expect(wrapper.get('[data-testid="pdf-block"] a').text()).toContain(
      "Open or download Course notes"
    );
    expect(
      wrapper.get('[data-testid="link-block"]').attributes()
    ).toMatchObject({
      target: "_blank",
      rel: "noopener noreferrer",
    });
    expect(wrapper.get('[data-testid="quote-block"]').element.tagName).toBe(
      "BLOCKQUOTE"
    );
    expect(wrapper.get('[data-testid="callout-block"]').element.tagName).toBe(
      "BLOCKQUOTE"
    );
    expect(wrapper.get('[data-testid="divider-block"]').element.tagName).toBe(
      "HR"
    );
  });

  it("keeps frozen, deliberately non-sorted blocks in exact array order", () => {
    const metadata = deepFreeze(
      metadataWith([
        { id: "z", type: "TEXT", data: { text: "Position 9" }, position: 9 },
        { id: "a", type: "TEXT", data: { text: "Position 1" }, position: 1 },
        { id: "m", type: "TEXT", data: { text: "Position 4" }, position: 4 },
      ])
    );

    const wrapper = mount(LearningItemRenderer, { props: { metadata } });
    const rendered = wrapper.findAll("[data-block-index]");

    expect(
      rendered.map((entry) => entry.attributes("data-block-index"))
    ).toEqual(["0", "1", "2"]);
    expect(rendered.map((entry) => entry.text())).toEqual([
      "Position 9",
      "Position 1",
      "Position 4",
    ]);
  });

  it("retains unsupported and malformed entries in sequence and continues", () => {
    const metadata = metadataWith([
      { id: "first", type: "TEXT", data: { text: "First" } },
      { id: "code", type: "CODE", data: { code: "x" } },
      { id: "broken", type: "IMAGE", data: { url: "/image.png" } },
      { id: "table", type: "TABLE", data: { rows: [] } },
      { id: "future", type: "FUTURE", data: {} },
      { id: "last", type: "TEXT", data: { text: "Last" } },
    ]);

    const wrapper = mount(LearningItemRenderer, { props: { metadata } });
    const rendered = wrapper.findAll("[data-block-index]");

    expect(rendered.map((entry) => entry.text())).toEqual([
      "First",
      "Unsupported content block.",
      "This content block is unavailable.",
      "Unsupported content block.",
      "Unsupported content block.",
      "Last",
    ]);
  });

  it.each([
    [undefined, "Content unavailable."],
    [null, "Content unavailable."],
    [{}, "Content unavailable."],
    [{ blocks: "not-an-array" }, "Content unavailable."],
    [{ version: 1, blocks: [] }, "No content available."],
  ])("renders the exact metadata state for %j", (metadata, message) => {
    const wrapper = mount(LearningItemRenderer, { props: { metadata } });
    expect(wrapper.text()).toBe(message);
  });

  it("does not mutate metadata or block data", () => {
    const metadata = deepFreeze(
      metadataWith([
        {
          id: "text",
          type: "TEXT",
          data: { text: "Immutable", extra: { nested: true } },
        },
      ])
    );
    const before = JSON.parse(JSON.stringify(metadata));

    mount(LearningItemRenderer, { props: { metadata } });

    expect(metadata).toEqual(before);
  });

  it("renders raw HTML-like text literally without injecting elements", () => {
    const raw = '<img src=x onerror="alert(1)"><script>alert(2)</script>';
    const wrapper = mount(LearningItemRenderer, {
      props: {
        metadata: metadataWith([
          { id: "raw", type: "TEXT", data: { text: raw } },
        ]),
      },
    });

    expect(wrapper.get('[data-testid="text-block"]').text()).toBe(raw);
    expect(wrapper.find("img").exists()).toBe(false);
    expect(wrapper.find("script").exists()).toBe(false);
  });

  it("keeps root-relative links in the current tab", () => {
    const wrapper = mount(LearningItemRenderer, {
      props: {
        metadata: metadataWith([
          {
            id: "link",
            type: "LINK",
            data: { url: "/course/next", label: "Next section" },
          },
        ]),
      },
    });
    const link = wrapper.get('[data-testid="link-block"]');

    expect(link.attributes("href")).toBe("/course/next");
    expect(link.attributes("target")).toBeUndefined();
    expect(link.attributes("rel")).toBeUndefined();
  });

  it("rejects unsafe media and link URLs as malformed content", () => {
    const wrapper = mount(LearningItemRenderer, {
      props: {
        metadata: metadataWith([
          {
            id: "image",
            type: "IMAGE",
            data: { url: "data:image/png;base64,abc", alt: "Unsafe" },
          },
          {
            id: "link",
            type: "LINK",
            data: { url: "javascript:alert(1)", label: "Unsafe" },
          },
          {
            id: "video",
            type: "VIDEO",
            data: { url: "//example.com/embed", title: "Unsafe" },
          },
        ]),
      },
    });

    expect(wrapper.findAll('[data-testid="malformed-block"]')).toHaveLength(3);
    expect(wrapper.find("img").exists()).toBe(false);
    expect(wrapper.find("iframe").exists()).toBe(false);
    expect(wrapper.find("a").exists()).toBe(false);
  });

  it("exposes responsive layout contracts without changing content order", () => {
    const wrapper = mount(LearningItemRenderer, {
      props: {
        metadata: metadataWith([
          {
            id: "image",
            type: "IMAGE",
            data: { url: "/image.png", alt: "Responsive image" },
          },
          {
            id: "pdf",
            type: "PDF",
            data: { url: "/document.pdf", title: "Responsive PDF" },
          },
        ]),
      },
    });

    expect(wrapper.get("img").classes()).toContain("w-full");
    expect(wrapper.get('[data-testid="pdf-block"] iframe').classes()).toEqual(
      expect.arrayContaining(["w-full", "sm:h-[38rem]"])
    );
    expect(
      wrapper
        .findAll("[data-block-index]")
        .map((entry) => entry.attributes("data-block-index"))
    ).toEqual(["0", "1"]);
  });
});
