import { getAuthSnapshot } from '../auth/state'

function normalize(value: string) {
  return value.trim().toLowerCase()
}

export function hasRole(role: string) {
  const target = normalize(role)
  if (!target) {
    return false
  }

  const user = getAuthSnapshot().user
  if (!user) {
    return false
  }

  return user.roles.some((candidate) => normalize(candidate) === target)
}

export function hasPermission(permission: string) {
  const target = normalize(permission)
  if (!target) {
    return false
  }

  const user = getAuthSnapshot().user
  if (!user) {
    return false
  }

  return user.permissions.some((candidate) => normalize(candidate) === target)
}
