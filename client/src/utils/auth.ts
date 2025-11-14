import { jwtDecode } from 'jwt-decode'

export function isTokenExpired(token: string | null): boolean {
  if (!token) return true;
  try {
    const decoded = jwtDecode(token);
    const expDate = decoded.exp;
    if (expDate === undefined) {
      throw new Error("No expiration date");
    } else {
      return expDate*1000 < Date.now();
    }
  } catch (err) {
    console.log("Error decoding token: ", err);
    return true;
  }
}
