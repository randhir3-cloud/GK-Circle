import {
  expect,
  test,
  type APIRequestContext,
  type Page,
  type Response,
} from "@playwright/test";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";
import { login } from "./fixtures/authenticated-user.js";

interface Course {
  id: string;
  title: string;
}

interface CourseNode {
  id: string;
  title: string;
}

interface LearningItem {
  id: string;
  title: string;
  publish_state: string;
}

interface Neighbor {
  id: string;
  title: string;
}

interface DetailData {
  learning_item: LearningItem;
  previous: Neighbor | null;
  next: Neighbor | null;
}

interface Diagnostics {
  consoleErrors: string[];
  pageErrors: string[];
  failedRequests: string[];
  serverErrors: string[];
}

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const screenshotsDir = path.resolve(
  __dirname,
  "../../../docs/features/course-system/evidence/course-p2-t17/screenshots"
);

const requiredEnvironment = (name: string): string => {
  const value = process.env[name]?.trim();
  if (!value) throw new Error(`${name} is required for T17 E2E verification`);
  return value;
};

const assertLocalTarget = (name: string, value: string): URL => {
  const url = new URL(value);
  if (!["localhost", "127.0.0.1", "::1"].includes(url.hostname)) {
    throw new Error(
      `${name} must identify an authorised local target before T17 may mutate data`
    );
  }
  return url;
};

const jsendData = async <T>(response: {
  json(): Promise<unknown>;
}): Promise<T> => {
  const body = (await response.json()) as { status?: string; data?: T };
  expect(body.status).toBe("success");
  return body.data as T;
};

const apiPath = (apiBaseUrl: URL, route: string): string =>
  new URL(`/api/v1${route}`, apiBaseUrl).toString();

const createDiagnostics = (): Diagnostics => ({
  consoleErrors: [],
  pageErrors: [],
  failedRequests: [],
  serverErrors: [],
});

const attachDiagnostics = (page: Page, diagnostics: Diagnostics): void => {
  page.on("console", (message) => {
    if (
      message.type() === "error" &&
      !/favicon\.ico|Download the Vue Devtools|\[HMR\]/i.test(message.text())
    ) {
      diagnostics.consoleErrors.push(message.text());
    }
  });
  page.on("pageerror", (error) => diagnostics.pageErrors.push(error.message));
  page.on("requestfailed", (request) => {
    if (!request.failure()?.errorText.includes("ERR_ABORTED")) {
      diagnostics.failedRequests.push(
        `${request.method()} ${request.url()} ${request.failure()?.errorText}`
      );
    }
  });
  page.on("response", (response) => {
    if (response.status() >= 500) {
      diagnostics.serverErrors.push(
        `${response.status()} ${response.request().method()} ${response.url()}`
      );
    }
  });
};

const expectCleanDiagnostics = (diagnostics: Diagnostics): void => {
  expect(diagnostics.consoleErrors, "browser console errors").toEqual([]);
  expect(diagnostics.pageErrors, "uncaught page errors").toEqual([]);
  expect(diagnostics.failedRequests, "unexpected failed requests").toEqual([]);
  expect(diagnostics.serverErrors, "unexpected server errors").toEqual([]);
};

const waitForApiResponse = (
  page: Page,
  method: string,
  pathSuffix: string,
  acceptedStatuses: number[] = [200]
): Promise<Response> =>
  page.waitForResponse(
    (response) =>
      response.request().method() === method &&
      new URL(response.url()).pathname.endsWith(pathSuffix) &&
      acceptedStatuses.includes(response.status())
  );

const getData = async <T>(
  request: APIRequestContext,
  url: string
): Promise<T> => {
  const response = await request.get(url);
  expect(response.status(), `GET ${new URL(url).pathname}`).toBe(200);
  return jsendData<T>(response);
};

const findDeepPath = async (
  request: APIRequestContext,
  apiBaseUrl: URL
): Promise<{ course: Course; nodes: CourseNode[] }> => {
  const courses = await getData<Course[]>(
    request,
    apiPath(apiBaseUrl, "/admin/courses")
  );

  const findNodePath = async (
    courseId: string,
    nodes: CourseNode[],
    ancestors: CourseNode[]
  ): Promise<CourseNode[] | null> => {
    for (const node of nodes) {
      const candidate = [...ancestors, node];
      if (candidate.length >= 4) return candidate;
      const children = await getData<CourseNode[]>(
        request,
        apiPath(
          apiBaseUrl,
          `/admin/courses/${encodeURIComponent(
            courseId
          )}/nodes/${encodeURIComponent(node.id)}/children`
        )
      );
      const result = await findNodePath(
        courseId,
        Array.isArray(children) ? children : [],
        candidate
      );
      if (result) return result;
    }
    return null;
  };

  for (const course of courses) {
    const roots = await getData<CourseNode[]>(
      request,
      apiPath(
        apiBaseUrl,
        `/admin/courses/${encodeURIComponent(course.id)}/nodes`
      )
    );
    const nodes = await findNodePath(
      course.id,
      Array.isArray(roots) ? roots : [],
      []
    );
    if (nodes) return { course, nodes };
  }

  throw new Error(
    "T17 requires an existing CourseNode path at least four levels deep"
  );
};

const selectDeepNode = async (
  page: Page,
  course: Course,
  nodes: CourseNode[]
): Promise<void> => {
  const rootsPath = `/api/v1/admin/courses/${encodeURIComponent(
    course.id
  )}/nodes`;
  await Promise.all([
    waitForApiResponse(page, "GET", rootsPath),
    page.getByTestId("course-selector").selectOption(course.id),
  ]);

  for (let index = 0; index < nodes.length; index += 1) {
    const selector = page.getByTestId(`node-selector-${index}`);
    await expect(selector).toBeVisible();
    const node = nodes[index];
    const itemListPath = `/api/v1/admin/courses/${encodeURIComponent(
      course.id
    )}/nodes/${encodeURIComponent(node.id)}/learning-items`;
    const childrenPath = `/api/v1/admin/courses/${encodeURIComponent(
      course.id
    )}/nodes/${encodeURIComponent(node.id)}/children`;
    await Promise.all([
      waitForApiResponse(page, "GET", itemListPath),
      waitForApiResponse(page, "GET", childrenPath),
      selector.selectOption(node.id),
    ]);
  }
};

const openAdminDeepNode = async (
  page: Page,
  course: Course,
  nodes: CourseNode[]
): Promise<void> => {
  await page.goto("/admin/courses/learning-items", {
    waitUntil: "domcontentloaded",
  });
  await expect(
    page.getByRole("heading", { name: "Course Content" })
  ).toBeVisible();
  await expect(page.getByTestId("course-selector")).toBeEnabled();
  await selectDeepNode(page, course, nodes);
};

const learnerListPath = (courseId: string, nodeId: string): string =>
  `/courses/${encodeURIComponent(courseId)}/nodes/${encodeURIComponent(
    nodeId
  )}/learning-items`;

test.describe.configure({ mode: "serial" });

test("admin creates a deep-node item, publishes it, and learner renders it", async ({
  browser,
  page,
}) => {
  test.setTimeout(120_000);

  const webBaseUrl = assertLocalTarget(
    "PLAYWRIGHT_BASE_URL",
    requiredEnvironment("PLAYWRIGHT_BASE_URL")
  );
  const apiBaseUrl = assertLocalTarget(
    "PLAYWRIGHT_API_BASE_URL",
    requiredEnvironment("PLAYWRIGHT_API_BASE_URL")
  );
  const email = requiredEnvironment("LOCAL_COURSE_ADMIN_EMAIL");
  const password = requiredEnvironment("LOCAL_COURSE_ADMIN_PASSWORD");
  const diagnostics = createDiagnostics();
  attachDiagnostics(page, diagnostics);
  fs.mkdirSync(screenshotsDir, { recursive: true });

  let courseId = "";
  let nodeId = "";
  let itemId = "";
  let enrollmentCreated = false;
  let cleanupError: Error | null = null;

  await login(page, {
    email,
    password,
    firstName: "Local",
    lastName: "CourseAdmin",
  });

  try {
    const { course, nodes } = await findDeepPath(page.request, apiBaseUrl);
    courseId = course.id;
    nodeId = nodes.at(-1)?.id || "";
    expect(nodes).toHaveLength(4);
    expect(nodeId).not.toBe("");

    await page.setViewportSize({ width: 1280, height: 900 });
    await openAdminDeepNode(page, course, nodes);

    const title = `T17 Learning Item ${Date.now()}`;
    await page.getByRole("button", { name: "New Learning Item" }).click();
    const editor = page.getByTestId("learning-item-editor");
    await expect(editor).toBeVisible();
    await editor.getByLabel("Title").fill(title);
    await editor.getByLabel("Item type").selectOption("ARTICLE");
    await editor.getByLabel("Publish state").selectOption("DRAFT");
    await editor
      .getByLabel("Description")
      .fill("Temporary T17 persisted end-to-end verification item.");

    const createPath = `/api/v1/admin/courses/${encodeURIComponent(
      courseId
    )}/nodes/${encodeURIComponent(nodeId)}/learning-items`;
    const createResponsePromise = waitForApiResponse(page, "POST", createPath, [
      201,
    ]);
    await editor.getByRole("button", { name: "Create", exact: true }).click();
    const createResponse = await createResponsePromise;
    const createdItem = await jsendData<LearningItem>(createResponse);
    itemId = createdItem.id;
    expect(createdItem.publish_state).toBe("DRAFT");
    await expect(page.getByTestId(`learning-item-${itemId}`)).toContainText(
      title
    );
    await expect(editor).toBeHidden();
    await page.screenshot({
      path: path.join(screenshotsDir, "desktop-admin.png"),
      fullPage: true,
    });

    await page.reload({ waitUntil: "domcontentloaded" });
    await expect(page.getByTestId("course-selector")).toBeEnabled();
    await selectDeepNode(page, course, nodes);
    await expect(page.getByTestId(`learning-item-${itemId}`)).toContainText(
      title
    );

    const listRoute = learnerListPath(courseId, nodeId);
    const learnerApiPath = `/api/v1/learner/courses/${encodeURIComponent(
      courseId
    )}/nodes/${encodeURIComponent(nodeId)}/learning-items`;
    const initialLearnerResponsePromise = waitForApiResponse(
      page,
      "GET",
      learnerApiPath,
      [200, 403]
    );
    await page.goto(listRoute, { waitUntil: "domcontentloaded" });
    const initialLearnerResponse = await initialLearnerResponsePromise;
    if (initialLearnerResponse.status() === 403) {
      await expect(page.getByTestId("enroll-course")).toBeVisible();
      const enrollmentPath = `/api/v1/learner/courses/${encodeURIComponent(
        courseId
      )}/enrollment`;
      const enrollmentResponsePromise = waitForApiResponse(
        page,
        "POST",
        enrollmentPath
      );
      const reloadedListPromise = waitForApiResponse(
        page,
        "GET",
        learnerApiPath
      );
      await page.getByTestId("enroll-course").click();
      await enrollmentResponsePromise;
      await reloadedListPromise;
      enrollmentCreated = true;
    }
    await expect(page.getByTestId(`learning-item-${itemId}`)).toHaveCount(0);

    await openAdminDeepNode(page, course, nodes);
    await page.getByRole("button", { name: `Edit ${title}` }).click();
    const updateEditor = page.getByTestId("learning-item-editor");
    await updateEditor.getByLabel("Publish state").selectOption("PUBLISHED");
    const updatePath = `${createPath}/${encodeURIComponent(itemId)}`;
    const updateResponsePromise = waitForApiResponse(page, "PATCH", updatePath);
    await updateEditor.getByRole("button", { name: "Save changes" }).click();
    const updatedItem = await jsendData<LearningItem>(
      await updateResponsePromise
    );
    expect(updatedItem.id).toBe(itemId);
    expect(updatedItem.publish_state).toBe("PUBLISHED");
    await expect(page.getByTestId(`learning-item-${itemId}`)).toContainText(
      "PUBLISHED"
    );

    const publishedListResponsePromise = waitForApiResponse(
      page,
      "GET",
      learnerApiPath
    );
    await page.goto(listRoute, { waitUntil: "domcontentloaded" });
    const publishedItems = await jsendData<LearningItem[]>(
      await publishedListResponsePromise
    );
    const renderedIds = await page
      .getByTestId("learning-item-list")
      .getByRole("listitem")
      .evaluateAll((elements) =>
        elements.map((element) =>
          element.getAttribute("data-testid")?.replace("learning-item-", "")
        )
      );
    expect(renderedIds).toEqual(publishedItems.map((item) => item.id));
    await expect(page.getByTestId(`learning-item-${itemId}`)).toContainText(
      title
    );
    await page.screenshot({
      path: path.join(screenshotsDir, "desktop-learner-list.png"),
      fullPage: true,
    });

    const detailApiPath = `${learnerApiPath}/${encodeURIComponent(itemId)}`;
    const detailResponsePromise = waitForApiResponse(
      page,
      "GET",
      detailApiPath
    );
    await page.getByTestId(`learning-item-${itemId}`).getByRole("link").click();
    const detail = await jsendData<DetailData>(await detailResponsePromise);
    await expect(
      page.getByRole("heading", { level: 1, name: title })
    ).toBeVisible();
    await expect(page.getByText("No content available.")).toBeVisible();

    const expectedNeighborHref = (neighbor: Neighbor): string =>
      `${listRoute}/${encodeURIComponent(neighbor.id)}`;
    if (detail.previous) {
      await expect(page.getByTestId("previous-item")).toHaveAttribute(
        "href",
        expectedNeighborHref(detail.previous)
      );
    } else {
      await expect(page.getByTestId("previous-item")).toHaveCount(0);
    }
    if (detail.next) {
      await expect(page.getByTestId("next-item")).toHaveAttribute(
        "href",
        expectedNeighborHref(detail.next)
      );
    } else {
      await expect(page.getByTestId("next-item")).toHaveCount(0);
    }
    await page.screenshot({
      path: path.join(screenshotsDir, "desktop-detail.png"),
      fullPage: true,
    });

    await page.reload({ waitUntil: "domcontentloaded" });
    await expect(
      page.getByRole("heading", { level: 1, name: title })
    ).toBeVisible();

    await page.setViewportSize({ width: 360, height: 800 });
    const mobileListResponsePromise = waitForApiResponse(
      page,
      "GET",
      learnerApiPath
    );
    await page.goto(listRoute, { waitUntil: "domcontentloaded" });
    await mobileListResponsePromise;
    await expect(page.getByTestId(`learning-item-${itemId}`)).toBeVisible();
    await page.screenshot({
      path: path.join(screenshotsDir, "mobile-list.png"),
      fullPage: true,
    });

    const mobileDetailResponsePromise = waitForApiResponse(
      page,
      "GET",
      detailApiPath
    );
    await page.goto(`${listRoute}/${encodeURIComponent(itemId)}`, {
      waitUntil: "domcontentloaded",
    });
    await mobileDetailResponsePromise;
    await expect(
      page.getByRole("heading", { level: 1, name: title })
    ).toBeVisible();
    await page.screenshot({
      path: path.join(screenshotsDir, "mobile-detail.png"),
      fullPage: true,
    });

    const signedOutContext = await browser.newContext({
      baseURL: webBaseUrl.toString(),
      viewport: { width: 360, height: 800 },
    });
    try {
      const signedOutPage = await signedOutContext.newPage();
      const deniedResponsePromise = waitForApiResponse(
        signedOutPage,
        "GET",
        detailApiPath,
        [401]
      );
      await signedOutPage.goto(`${listRoute}/${encodeURIComponent(itemId)}`, {
        waitUntil: "domcontentloaded",
      });
      await deniedResponsePromise;
      const signedOutAlert = signedOutPage.getByTestId("item-error");
      await expect(signedOutAlert).toBeVisible();
      await expect(signedOutAlert).not.toBeEmpty();
      await signedOutPage.screenshot({
        path: path.join(screenshotsDir, "signed-out.png"),
        fullPage: true,
      });
    } finally {
      await signedOutContext.close();
    }

    expectCleanDiagnostics(diagnostics);
  } finally {
    if (itemId && courseId && nodeId) {
      const itemUrl = apiPath(
        apiBaseUrl,
        `/admin/courses/${encodeURIComponent(
          courseId
        )}/nodes/${encodeURIComponent(
          nodeId
        )}/learning-items/${encodeURIComponent(itemId)}`
      );
      try {
        const deleteResponse = await page.request.delete(itemUrl);
        expect(deleteResponse.status(), "temporary item cleanup").toBe(200);
        const missingResponse = await page.request.get(itemUrl);
        expect(
          missingResponse.status(),
          "temporary item must be absent after cleanup"
        ).toBe(404);
      } catch (error) {
        cleanupError =
          error instanceof Error ? error : new Error(String(error));
      }
    }

    if (enrollmentCreated && courseId) {
      try {
        const enrollmentUrl = apiPath(
          apiBaseUrl,
          `/learner/courses/${encodeURIComponent(courseId)}/enrollment`
        );
        const response = await page.request.delete(enrollmentUrl);
        expect(response.status(), "temporary enrollment cleanup").toBe(200);
        const enrollment = await getData<{ enrolled: boolean }>(
          page.request,
          enrollmentUrl
        );
        expect(enrollment.enrolled).toBe(false);
      } catch (error) {
        cleanupError =
          error instanceof Error ? error : new Error(String(error));
      }
    }

    if (cleanupError) throw cleanupError;
  }
});
