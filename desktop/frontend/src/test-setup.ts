import '@testing-library/jest-dom/vitest'

if (typeof globalThis.Response === 'undefined') {
  // Provide a minimal Response polyfill for Node runtimes without fetch/Web APIs.
  class MinimalResponse {
    status: number
    headers: Headers
    private readonly bodyText: string

    constructor(body?: BodyInit | null, init?: ResponseInit) {
      this.status = init?.status ?? 200
      this.headers = new Headers(init?.headers)
      this.bodyText = typeof body === 'string' ? body : body ? String(body) : ''
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

  Object.defineProperty(globalThis, 'Response', {
    value: MinimalResponse,
    configurable: true,
    writable: true,
  })
}

if (typeof window !== 'undefined' && typeof window.matchMedia === 'undefined') {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      addListener: () => undefined,
      removeListener: () => undefined,
      dispatchEvent: () => false,
    }),
  })
}
