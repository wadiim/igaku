import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faAngleLeft, faAngleRight, faArrowUpLong, faArrowDownLong, faFloppyDisk,
} from '@fortawesome/free-solid-svg-icons'

import type { UserData } from './utils/user'
import { saveUserListToFile } from './utils/user'
import ProfileCard from './profile-card'
import { getRole, isTokenExpired } from './utils/auth'

let nextId = 0;

function Users() {
  const [usersData, setUsersData] = useState<UserData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [currPage, setCurrPage] = useState(1);
  const [pageCount, setPageCount] = useState(1);
  const [sortBy, setSortBy] = useState("id");
  const [sortOrder, setSortOrder] = useState("asc");

  let navigate = useNavigate();

  const fetchUserList = async (
    jwt: string,
    page: number,
    sortBy: string,
    sortOrder: string,
  ): Promise<any> => {
    let params = `page=${page}&orderBy=${sortBy}&orderMethod=${sortOrder}`;
    fetch(`http://localhost:4000/user/list/?${params}`, {
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
      fetchUserList(jwt, currPage, sortBy, sortOrder);
    }
  });

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

  const handleSortByChange = (value: string) => {
    fetchUserList(getJwt(), currPage, value, sortOrder);
    setSortBy(value);
  }

  const handleSortOrderChange = () => {
    let order = (sortOrder === "asc") ? "desc" : "asc";
    fetchUserList(getJwt(), currPage, sortBy, order);
    setSortOrder(order);
  }

  const handleSaveClick = () => {
    saveUserListToFile(
      1,
      10*pageCount,
      sortBy,
      sortOrder,
      (err: string) => { alert(err); },
    );
  }

  const handlePrevClick = () => {
    if (currPage <= 1) return;

    let jwt = getJwt();
    let prevPage = currPage - 1;
    setCurrPage(prevPage);
    fetchUserList(jwt, prevPage, sortBy, sortOrder);
  }

  const handleNextClick = () => {
    if (currPage >= pageCount) return;

    let jwt = getJwt();
    let nextPage = currPage + 1;
    setCurrPage(nextPage);
    fetchUserList(jwt, nextPage, sortBy, sortOrder);
  }

  return (
    <div
      className={`
        flex flex-col items-center justify-center
      `}
    >
    <div
      className={`
        text-tn-d-fg
        m-4
      `}
    >
      <label
        htmlFor="sortBy"
      >
        Sort by:
      </label>
      <select
        id="sortBy"
        name="sortBy"
        value={sortBy}
        onChange={e => handleSortByChange(e.target.value)}
        className={`
          bg-tn-d-black
          p-2 mx-2
        `}
      >
        <option value="id">ID</option>
        <option value="username">Username</option>
      </select>

      <button
        onClick={handleSortOrderChange}
        className={`
          cursor-pointer
          hover:bg-tn-d-black
          rounded-full
          p-2 ms-2
        `}
      >
        { (sortOrder === "asc")
            && <FontAwesomeIcon icon={faArrowUpLong} />
            || <FontAwesomeIcon icon={faArrowDownLong} />
        }
      </button>

      <button
        onClick={handleSaveClick}
        className={`
          cursor-pointer
          hover:bg-tn-d-black
          rounded-full
          p-2 ms-2
        `}
      >
        <FontAwesomeIcon icon={faFloppyDisk} />
      </button>

    </div>
      <ul>
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
