import fs from 'fs'
import path from 'path'

export async function loginAs(page, email, password, baseUrl) {
  await page.goto(`${baseUrl}/login`, { waitUntil: 'networkidle' })
  await page.fill('input[type="email"], input[name="identifier"], input[name="email"]', email)
  await page.fill('input[type="password"]', password)
  await Promise.all([
    page.waitForNavigation({ waitUntil: 'networkidle' }).catch(() => {}),
    page.click('button[type="submit"]')
  ])
  // Verify successful authentication (no login form present or dashboard URL reached)
  const isLoginVisible = await page.locator('input[type="password"]').isVisible().catch(() => false)
  if (isLoginVisible) {
    throw new Error(`Authentication failed for ${email}`)
  }
}

export async function logout(page, baseUrl) {
  await page.goto(`${baseUrl}/logout`, { waitUntil: 'networkidle' }).catch(() => {})
  await page.context().clearCookies()
}
