/**
 * Password validation result
 */
export interface PasswordValidationResult {
  valid: boolean
  errors: string[]
  strength: 'weak' | 'medium' | 'strong'
}

/**
 * Special characters allowed in passwords
 */
export const SPECIAL_CHARS = '!@#$%^&*()_+-=[]{}|;:,.<>?'

/**
 * Validate password strength
 * Requirements:
 * - At least 8 characters
 * - At least one uppercase letter
 * - At least one lowercase letter
 * - At least one digit
 * - At least one special character
 * 
 * @param password Password to validate
 * @returns Validation result with errors and strength indicator
 */
export function validatePassword(password: string): PasswordValidationResult {
  const errors: string[] = []
  let score = 0

  // Check length
  if (password.length < 8) {
    errors.push('密码长度至少8位')
  } else {
    score++
    if (password.length >= 12) score++
  }

  // Check for uppercase
  const hasUpper = /[A-Z]/.test(password)
  if (!hasUpper) {
    errors.push('需要包含大写字母')
  } else {
    score++
  }

  // Check for lowercase
  const hasLower = /[a-z]/.test(password)
  if (!hasLower) {
    errors.push('需要包含小写字母')
  } else {
    score++
  }

  // Check for digit
  const hasDigit = /[0-9]/.test(password)
  if (!hasDigit) {
    errors.push('需要包含数字')
  } else {
    score++
  }

  // Check for special character
  const specialCharsRegex = /[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]/
  const hasSpecial = specialCharsRegex.test(password)
  if (!hasSpecial) {
    errors.push('需要包含特殊字符 (!@#$%^&*()_+-=[]{}|;:,.<>?)')
  } else {
    score++
  }

  // Determine strength
  let strength: 'weak' | 'medium' | 'strong'
  if (score <= 2) {
    strength = 'weak'
  } else if (score <= 4) {
    strength = 'medium'
  } else {
    strength = 'strong'
  }

  return {
    valid: errors.length === 0,
    errors,
    strength,
  }
}

/**
 * Get password strength color for UI display
 * @param strength Password strength level
 * @returns Color string for Element Plus
 */
export function getStrengthColor(strength: 'weak' | 'medium' | 'strong'): string {
  switch (strength) {
    case 'weak':
      return '#F56C6C'
    case 'medium':
      return '#E6A23C'
    case 'strong':
      return '#67C23A'
    default:
      return '#909399'
  }
}

/**
 * Get password strength percentage for progress bar
 * @param strength Password strength level
 * @returns Percentage value (0-100)
 */
export function getStrengthPercentage(strength: 'weak' | 'medium' | 'strong'): number {
  switch (strength) {
    case 'weak':
      return 33
    case 'medium':
      return 66
    case 'strong':
      return 100
    default:
      return 0
  }
}

/**
 * Get password strength label for display
 * @param strength Password strength level
 * @returns Localized strength label
 */
export function getStrengthLabel(strength: 'weak' | 'medium' | 'strong'): string {
  switch (strength) {
    case 'weak':
      return '弱'
    case 'medium':
      return '中'
    case 'strong':
      return '强'
    default:
      return ''
  }
}
