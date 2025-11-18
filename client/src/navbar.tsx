import { useEffect } from 'react'
import { Link, useNavigate, useMatch, useResolvedPath } from 'react-router'

function Navbar() {
  let navigate = useNavigate();

  const handleSignOut = () => {
    localStorage.removeItem("jwt");
    navigate("/auth/login");
  }

  return (
    <nav
      className={`
        flex-none bg-tn-d-black flex
      `}
    >
      <ul
        className={`
          flex-1 flex
          font-bold text-tn-d-fg text-2xl
          gap-8 px-4 items-center
        `}
      >
        <NavLink to="/">Home</NavLink>
        <NavLink to="/profile">Profile</NavLink>
      </ul>
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

function NavLink({ to, children }) {
  let path = useResolvedPath(to);
  let isActive = useMatch({ path: path.pathname, end: true });

  return (
    <li>
      <Link
        to={to}
        className={`
          cursor-pointer
          hover:text-tn-d-blue
          ${isActive ? "text-tn-d-dblue" : ""}
        `}
      >
        {children}
      </Link>
    </li>
  );
}

export default Navbar;
