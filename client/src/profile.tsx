import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { isTokenExpired } from './utils/auth'

import ProfileCard from './profile-card'
import type { UserData } from './utils/user'

function Profile() {
  const [userData, setUserData] = useState<UserData>({
    username: "",
    email: "",
    role: "",
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  let navigate = useNavigate();

  useEffect(() => {
    let jwt = localStorage.getItem("jwt");
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if (jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch('http://localhost:4000/user/self/', {
        method: 'GET',
        headers: {
          'accept': 'application/json',
          'Authorization': jwt,
        }
      })
      .then((res) => {
        if (!res.ok) {
          throw new Error("Failed to load user data");
        }
        return res.json();
      })
      .then((data) => {
        setUserData(data);
        localStorage.setItem("userData", JSON.stringify(data));
        setError(null);
        setLoading(false);
      })
      .catch((err) => {
        if (err instanceof TypeError && err.message === "Failed to fetch") {
          // No network connection.
          // NOTE: This detection mechanism does not work in Firefox.
          const catchedUserData = localStorage.getItem("userData");
          if (catchedUserData) {
            setUserData(JSON.parse(catchedUserData));
            setError(null);
          } else {
            setError("Failed to load user data");
          }
          setLoading(false);
        } else {
          setError(err.message);
          setLoading(false);
        }
      })
    }
  }, [navigate]);

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
     <div className={`flex-1 flex flex-col items-center justify-center`}>
      <ProfileCard userData={userData} />
     </div>
   );
}

export default Profile;
