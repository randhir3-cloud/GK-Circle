FROM mcr.microsoft.com/playwright:v1.61.1-noble

WORKDIR /app

# Copy lock and package manifest
COPY app/package*.json ./
RUN npm ci

# Copy test files, helpers, and configurations
COPY app/tests ./tests
COPY app/playwright.config.ts ./

# Run Playwright tests headlessly
CMD ["npx", "playwright", "test", "tests/e2e/runtime-verification.spec.ts", "--project=chromium-observation", "--headed=false"]
