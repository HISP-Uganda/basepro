export type NormalizedAppErrorType =
  | 'validation'
  | 'unauthorized'
  | 'forbidden'
  | 'not_found'
  | 'conflict'
  | 'network'
  | 'timeout'
  | 'server'
  | 'unknown'

export type NormalizedAppError = {
  type: NormalizedAppErrorType
  message: string
  fieldErrors?: Record<string, string[]>
  requestId?: string
}
