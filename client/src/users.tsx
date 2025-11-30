import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'

import type { UserData } from './utils/user-data'
import ProfileCard from './profile-card'
import { getRole, isTokenExpired } from './utils/auth'

let nextId = 0;

function Users() {
  const [usersData, setUsersData] = useState<UserData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  let navigate = useNavigate();

  useEffect(() => {
    let jwt = localStorage.getItem("jwt");
    if (isTokenExpired(jwt)) {
      navigate("/auth/login");
    }
    if (jwt !== null) {
      let role = getRole(jwt);
      if (role !== "admin") {
        navigate("/unauthorized");
      }
    }
  })

  useEffect(() => {
    let jwt = localStorage.getItem("jwt");
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if (jwt === null) {
      throw new Error("Authentication failed");
    } else if (usersData.length === 0) {
      fetch('http://localhost:4000/user/list/', {
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
      .then((data) => {
        let users = [];
        for (let user of data.data) {
          users.push(
            {
              id: nextId++,
              username: user.username,
              email: user.email,
              role: user.role,
            }
          );
        }
        setUsersData(users);
        setError(null);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      })
    }
  }, [navigate, nextId]);

  if (loading) {
    return <></>;
  }

  if (error) {
    return (
      <div
        className={`
          flex-1 flex flex-col items-center justify-center
          text-xl font-bold text-tn-d-red
        `}
      >
        {error}
      </div>
    );
  }

  return (
    <div
      className={`
        flex flex-col items-center justify-center h-screen
      `}
    >
      <ul>
        <UserList usersData={usersData} />
      </ul>
    </div>
  );
}

function UserList({ usersData }: { usersData: UserData[] }) {
  const list = usersData.map(user =>
    <li>
      <ProfileCard userData={user} />
    </li>
  );

  return list;
}

export default Users;
