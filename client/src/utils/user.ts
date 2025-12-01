import { getNewCsvFileHandle } from './file'

export interface UserData {
  id?: number,
  username: string,
  email: string,
  role: string,
}

export async function saveUserListToFile(
  page: number,
  pageSize: number,
  sortBy: string,
  sortOrder: string,
  errCallback: (err: string) => void,
) {
  try {
    const handle = await getNewCsvFileHandle();

    const jwt = localStorage.getItem("jwt");
    if (jwt === null) {
      throw new Error("Failed to authenticate");
    }

    let params =
      + `page=${page}`
      + `&pageSize=${pageSize}`
      + `&orderBy=${sortBy}`
      + `&orderMethod=${sortOrder}`;
    fetch(`http://localhost:4000/user/list/?${params}`, {
      method: 'GET',
      headers: {
        'accept': 'application/json',
        'Authorization': jwt,
      }
    })
    .then((res) => {
      if (!res.ok) {
        throw new Error("Failed to load users data");
      }
      return res.json();
    })
    .then(async (data) => {
      const writable = await handle.createWritable();
      for (let user of data.data) {
        await writable.write(`${user.username},${user.email},${user.role}\n`);
      }
      await writable.close();
    })
    .catch((err) => {
      errCallback(err.message);
    })
  } catch (err) {
    if (typeof err === "string") {
      errCallback(err);
    } else if (err instanceof Error) {
      errCallback(err.message);
    } else {
      errCallback("Failed to save user list");
    }
  }
}
