import { useEffect } from 'react'
import { useNavigate } from 'react-router'

function Navbar() {
  let navigate = useNavigate();

  const handleSignOut = () => {
    localStorage.removeItem("jwt");
    navigate("/auth/login");
  }

  return (
    <nav
      className={`
        flex-none bg-tn-d-black flex flex-row-reverse
      `}
    >
      <button
        className={`
          cursor-pointer
          px-2 py-1 m-2
          rounded-md
          bg-tn-d-dblue
          hover:bg-tn-d-blue
          text-sm/6 font-semibold
          text-tn-d-fg
          focus-visible:outline-2 focus-visible:outline-offset-2
          focus-visible:outline-tn-d-blue
        `}
        onClick={handleSignOut}
      >
        Sign out
      </button>
    </nav>
  );
}

export default Navbar;
