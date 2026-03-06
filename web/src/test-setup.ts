import '@testing-library/jest-dom/vitest'

type HeaderInit = Record<string, string> | Array<[string, string]>

class TestHeaders {
  private readonly values = new Map<string, string>()

  constructor(init?: HeaderInit) {
    if (!init) {
      return
    }
    if (Array.isArray(init)) {
      for (const [key, value] of init) {
        this.set(key, value)
      }
      return
    }
    for (const [key, value] of Object.entries(init)) {
      this.set(key, value)
    }
  }

  set(key: string, value: string) {
    this.values.set(key.toLowerCase(), value)
  }

  get(key: string) {
    return this.values.get(key.toLowerCase()) ?? null
  }

  has(key: string) {
    return this.values.has(key.toLowerCase())
  }

  forEach(callback: (value: string, key: string) => void) {
    for (const [key, value] of this.values.entries()) {
      callback(value, key)
    }
  }
}

class TestResponse {
  status: number
  headers: TestHeaders
  private readonly bodyText: string

  constructor(body?: string | null, init?: { status?: number; headers?: HeaderInit }) {
    this.status = init?.status ?? 200
    this.headers = new TestHeaders(init?.headers)
    this.bodyText = body ?? ''
  }

  get ok() {
    return this.status >= 200 && this.status < 300
  }

  async json() {
    if (!this.bodyText) {
      return {}
    }
    return JSON.parse(this.bodyText)
  }

  async text() {
    return this.bodyText
  }
}

if (typeof globalThis.Response === 'undefined') {
  // TanStack Router and API tests rely on functional fetch primitives.
  Object.defineProperty(globalThis, 'Headers', {
    value: TestHeaders,
    configurable: true,
    writable: true,
  })
  Object.defineProperty(globalThis, 'Response', {
    value: TestResponse,
    configurable: true,
    writable: true,
  })
}

if (typeof window !== 'undefined' && typeof window.matchMedia !== 'function') {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
      matches: query === '(prefers-color-scheme: dark)' ? false : false,
      media: query,
      onchange: null,
      addListener: () => undefined,
      removeListener: () => undefined,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      dispatchEvent: () => false,
    }),
  })
}
