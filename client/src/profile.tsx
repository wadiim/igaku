import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { isTokenExpired } from './utils/auth'

function Profile() {
  const [userData, setUserData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  let navigate = useNavigate();

  useEffect(() => {
    let jwt = localStorage.getItem("jwt");
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

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
        setError(false);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      })
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
       <div
         className={`
           grid grid-cols-[max-content_1fr]
           text-tn-d-fg text-2xl
           border-2 border-tn-d-fg rounded-2xl p-4 gap-y-2
         `}
       >
         <ProfileItem title="Username" value={userData.username} />
         <ProfileItem title="Email" value={userData.email} />
         <ProfileItem title="Role" value={userData.role} />
       </div>
     </div>
   );
 }

function ProfileItem({ title, value }) {
   return (
     <>
       <span className="font-bold">{ title }:</span>
       <span className="ms-4">{ value }</span>
     </>
   );
}

export default Profile;
