import { describe, it, expect } from "vitest";
import { mount, RouterLinkStub } from "@vue/test-utils";
import Pagination from "~/components/Pagination.vue";

const mountComponent = (props) =>
  mount(Pagination, {
    props,
    global: {
      stubs: {
        NuxtLink: RouterLinkStub,
      },
    },
  });

describe("Pagination test", () => {
  it("renders the current page correctly", () => {
    const wrapper = mountComponent({ page: 2, numOfRecords: 5 });
    expect(wrapper.text()).toContain("2");
  });

  it("disables the previous control on the first page", () => {
    const wrapper = mountComponent({ page: 1, numOfRecords: 5 });
    const links = wrapper.findAllComponents(RouterLinkStub);
    expect(links).toHaveLength(1);
    expect(links[0].props("to")).toEqual({
      path: "/",
      query: { page: 2 },
    });
    expect(wrapper.html()).toContain("cursor-not-allowed");
  });

  it("disables the next control on the last page", () => {
    const wrapper = mountComponent({ page: 5, numOfRecords: 5 });
    const links = wrapper.findAllComponents(RouterLinkStub);
    expect(links).toHaveLength(1);
    expect(links[0].props("to")).toEqual({
      path: "/",
      query: { page: 4 },
    });
    expect(wrapper.html()).toContain("cursor-not-allowed");
  });

  it("enables both controls on a middle page", () => {
    const wrapper = mountComponent({ page: 3, numOfRecords: 5 });
    const links = wrapper.findAllComponents(RouterLinkStub);
    expect(links).toHaveLength(2);
    expect(links[0].props("to")).toEqual({
      path: "/",
      query: { page: 2 },
    });
    expect(links[1].props("to")).toEqual({
      path: "/",
      query: { page: 4 },
    });
  });

  it("navigates to the correct previous page URL", () => {
    const wrapper = mountComponent({ page: 2, numOfRecords: 5 });
    const prevLink = wrapper.findAllComponents(RouterLinkStub)[0];
    expect(prevLink.props().to).toEqual({
      path: "/",
      query: { page: 1 },
    });
  });

  it("navigates to the correct next page URL", () => {
    const wrapper = mountComponent({ page: 2, numOfRecords: 5 });
    const nextLink = wrapper.findAllComponents(RouterLinkStub)[1];
    expect(nextLink.props("to")).toEqual({
      path: "/",
      query: { page: 3 },
    });
  });
});
