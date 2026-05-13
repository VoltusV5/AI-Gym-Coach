import { SecureStorage } from '@aparajita/capacitor-secure-storage';

const TOKEN_KEY = 'auth_token';


export async function saveToken(token) {
  try {
    await SecureStorage.set({ key: TOKEN_KEY, value: token });
  } catch (error) {
    console.error('Error saving token:', error);
    localStorage.setItem(TOKEN_KEY, token);
  }
}


export async function getToken() {
  try {
    const { value } = await SecureStorage.get({ key: TOKEN_KEY });
    return value;
  } catch (error) {
    return localStorage.getItem(TOKEN_KEY) || null;
  }
}


export async function removeToken() {
  localStorage.removeItem(TOKEN_KEY);
  try {
    await SecureStorage.remove({ key: TOKEN_KEY });
  } catch (error) {
    console.warn('Error removing token from secure storage:', error);
  }
}