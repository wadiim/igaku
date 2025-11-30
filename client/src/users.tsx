import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faAngleLeft, faAngleRight } from '@fortawesome/free-solid-svg-icons'

import type { UserData } from './utils/user-data'
import ProfileCard from './profile-card'
import { getRole, isTokenExpired } from './utils/auth'

let nextId = 0;

function Users() {
  const [usersData, setUsersData] = useState<UserData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [currPage, setCurrPage] = useState(1);
  const [pageCount, setPageCount] = useState(1);

  let navigate = useNavigate();
  let params = useParams();

  const fetchUserList = async (page: number, jwt: string): Promise<any> => {
    fetch(`http://localhost:4000/user/list/?page=${page}`, {
      method: 'GET',
      headers: {
        'accept': 'application/json',
        'Authorization': jwt,
      }
    })
    .then((res) => {
      if (res.status === 400) {
        throw new Error("Failed to load users data: Invalid parameters");
      } else if (res.status === 401) {
        throw new Error("Failed to load users data: Unauthenticated");
      } else if (res.status === 403) {
        throw new Error("Failed to load users data: Unauthorized");
      } else if (!res.ok) {
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
      setPageCount(data.total_pages);

      setError(null);
      setLoading(false);
    })
    .catch((err) => {
      setError(err.message);
      setLoading(false);
    })
  }

  const getJwt = (): string => {
    let jwt = localStorage.getItem("jwt");
    if (isTokenExpired(jwt)) {
      navigate("/auth/login");
    }
    return jwt || "";
  }

  useEffect(() => {
    let jwt = getJwt();
    let role = getRole(jwt);
    if (role !== "admin") {
      navigate("/unauthorized");
    }
  })

  useEffect(() => {
    let jwt = getJwt();
    if (usersData.length === 0) {
      let page = 1;
      if (params.page !== undefined) {
        page = +params.page;
      }
      setCurrPage(page);

      fetchUserList(page, jwt);
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

  const handlePrevClick = () => {
    if (currPage <= 1) return;

    let jwt = getJwt();
    let prevPage = currPage - 1;
    setCurrPage(prevPage);
    fetchUserList(prevPage, jwt);
  }

  const handleNextClick = () => {
    if (currPage >= pageCount) return;

    let jwt = getJwt();
    let nextPage = currPage + 1;
    setCurrPage(nextPage);
    fetchUserList(nextPage, jwt);
  }

  return (
    <div
      className={`
        flex flex-col items-center justify-center
      `}
    >
      <ul className="mt-4">
        <UserList usersData={usersData} />
      </ul>
      <div
        className={`
          inline-flex rounded-base m-4
        `}
      >
        <button
          onClick={handlePrevClick}
          className={`
            cursor-pointer
            inline-flex items-center justify-center
            bg-tn-d-black text-tn-d-fg
            rounded-s-lg box-border border border-tn-d-white-medium
            hover:border-tn-d-blue
            w-9 h-9
          `}
        >
          <FontAwesomeIcon icon={faAngleLeft} />
        </button>
        <button
          className={`
            inline-flex items-center justify-center
            bg-tn-d-black text-tn-d-fg
            box-border border border-tn-d-white-medium
            px-3 h-9
          `}
        >
          {currPage} of {pageCount}
        </button>
        <button
          onClick={handleNextClick}
          className={`
            cursor-pointer
            inline-flex items-center justify-center
            bg-tn-d-black text-tn-d-fg
            rounded-e-lg box-border border border-tn-d-white-medium
            hover:border-tn-d-blue
            w-9 h-9
          `}
        >
          <FontAwesomeIcon icon={faAngleRight} />
        </button>
      </div>
    </div>
  );
}

function UserList({ usersData }: { usersData: UserData[] }) {
  const list = usersData.map(user =>
    <li key={user.id}>
      <ProfileCard userData={user} />
    </li>
  );

  return list;
}

export default Users;
