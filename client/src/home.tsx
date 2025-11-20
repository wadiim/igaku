import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'

import { isTokenExpired } from './utils/auth'
import { sendNotification } from './utils/notify'

function Home() {
  const [count, setCount] = useState(0)

  let navigate = useNavigate();

  useEffect(() => {
    let jwt = localStorage.getItem("jwt");
    if (isTokenExpired(jwt)) {
      navigate("/auth/login");
    }
  });

  const handleClick = () => {
    if (count === 68) {
      sendNotification("Nice!");
    }
    setCount(count + 1);
  }

  return (
    <div className={`flex-1 flex flex-col items-center justify-center`}>
      <button
        className={`
          bg-tn-d-black text-tn-d-fg
          px-4 py-2
          rounded-md
          cursor-pointer
          border border-transparent
          hover:border hover:border-tn-d-blue
          transition-border duration-250
        `}
        onClick={handleClick}
      >
        count is {count}
      </button>
    </div>
  )
}

export default Home;
