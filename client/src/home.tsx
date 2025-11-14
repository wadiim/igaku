import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'
import { isTokenExpired } from './utils/auth'

function Home() {
  const [count, setCount] = useState(0)

  let navigate = useNavigate();

  useEffect(() => {
    let jwt = localStorage.getItem("jwt");
    if (isTokenExpired(jwt)) {
      navigate("/auth/login");
    }
  });

  return (
    <>
      <div
        className={`
          flex items-center justify-center h-screen
        `}
      >
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
          onClick={() => setCount((count) => count + 1)}
        >
          count is {count}
        </button>
      </div>
    </>
  )
}

export default Home;
