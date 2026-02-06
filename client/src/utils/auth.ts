import { jwtDecode } from 'jwt-decode'
import type { JwtPayload } from 'jwt-decode'

interface CustomJwtPayload extends JwtPayload {
  role: string;
}

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

export function getRole(token: string): string {
    const decoded = jwtDecode<CustomJwtPayload>(token);
    return decoded.role;
}
