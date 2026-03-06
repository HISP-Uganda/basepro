const REDIRECT_KEY = 'basepro.desktop.post_login_redirect'

function sanitizeRedirect(path: string) {
  if (!path.startsWith('/')) {
    return ''
  }
  if (path.startsWith('/login') || path.startsWith('/setup') || path.startsWith('/forgot-password') || path.startsWith('/reset-password')) {
    return ''
  }
  return path
}

export function rememberIntendedDestination(path: string) {
  if (typeof window === 'undefined') {
    return
  }
  const redirectPath = sanitizeRedirect(path)
  if (!redirectPath) {
    return
  }
  window.sessionStorage.setItem(REDIRECT_KEY, redirectPath)
}

export function consumeIntendedDestination(fallback = '/dashboard') {
  if (typeof window === 'undefined') {
    return fallback
  }
  const redirectPath = window.sessionStorage.getItem(REDIRECT_KEY) ?? ''
  window.sessionStorage.removeItem(REDIRECT_KEY)
  return sanitizeRedirect(redirectPath) || fallback
}
