import { SecureStorage } from '@aparajita/capacitor-secure-storage';

const TOKEN_KEY = 'auth_token';

/**
 * Сохранить токен в secure-storage
 * @param {string} token
 */
export async function saveToken(token) {
  try {
    await SecureStorage.set({ key: TOKEN_KEY, value: token });
  } catch (error) {
    console.error('Error saving token:', error);
    localStorage.setItem(TOKEN_KEY, token);
  }
}

/**
 * Достать токен из secure-storage
 * @returns {Promise<string|null>}
 */
export async function getToken() {
  try {
    const { value } = await SecureStorage.get({ key: TOKEN_KEY });
    return value;
  } catch (error) {
    return localStorage.getItem(TOKEN_KEY) || null;
  }
}

/**
 * Удалить токен из secure-storage
 */
export async function removeToken() {
  try {
    await SecureStorage.remove({ key: TOKEN_KEY });
    localStorage.removeItem(TOKEN_KEY);
  } catch (error) {
    console.warn('Error removing token:', error);
    localStorage.removeItem(TOKEN_KEY);
  }
}
