import request from '@/utils/request'
import type { LoginResponse, ChangePasswordRequest, UserInfo } from '@/types'

/**
 * Authentication API service
 * Handles login, logout, password change, and user info retrieval
 */
/**
 * Captcha response
 */
export interface CaptchaResponse {
  captcha_id: string
  image: string // base64 image data URL
}

/**
 * Login request with captcha
 */
export interface LoginRequest {
  username: string
  password: string
  captcha_id?: string
  captcha?: string
}

export const authApi = {
  /**
   * Get captcha image
   * @returns Captcha ID and base64 image
   */
  async getCaptcha(): Promise<CaptchaResponse> {
    const response = await request.get<CaptchaResponse>('/auth/captcha')
    return response.data
  },

  /**
   * Login with username and password and captcha
   * @param data Login credentials
   * @returns Login response with token and password change status
   */
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await request.post<LoginResponse>('/auth/login', data)
    return response.data
  },

  /**
   * Logout current user
   * Clears server-side session
   */
  async logout(): Promise<void> {
    await request.post('/auth/logout')
  },

  /**
   * Change user password
   * @param data Old and new password
   * @returns Response with new token if successful
   */
  async changePassword(data: ChangePasswordRequest): Promise<{ message: string; token: string }> {
    const response = await request.post<{ message: string; token: string }>('/auth/change-password', data)
    return response.data
  },

  /**
   * Get current user information
   * @returns User info including username and password change requirement
   */
  async getCurrentUser(): Promise<UserInfo> {
    const response = await request.get<UserInfo>('/auth/me')
    return response.data
  },
}

export default authApi
