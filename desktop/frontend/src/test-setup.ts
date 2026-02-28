import '@testing-library/jest-dom/vitest'

if (typeof globalThis.Response === 'undefined') {
  // TanStack Router checks for Response existence when handling redirects.
  Object.defineProperty(globalThis, 'Response', {
    value: class {},
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
