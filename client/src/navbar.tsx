import { useEffect, useState } from 'react'
import { Link, useNavigate, useMatch, useResolvedPath } from 'react-router'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faBars } from '@fortawesome/free-solid-svg-icons'

function Navbar() {
  const [mobileMenuHidden, setMobileMenuHidden] = useState(true);

  let navigate = useNavigate();

  const handleSignOut = () => {
    localStorage.removeItem("jwt");
    navigate("/auth/login");
  }

  const toggleMobileMenu = () => {
    setMobileMenuHidden(!mobileMenuHidden);
  }

  // Prevents reopening the mobile menu when the window's width is increased
  // to a size greater than the maximum mobile view width and then decreased
  // back.
  useEffect(() => {
    const hideMobileMenu = () => setMobileMenuHidden(true);
    const mediaQueryList = window.matchMedia('(min-width: 48rem)');
    mediaQueryList.addListener(hideMobileMenu);

    return () => {
      mediaQueryList.removeListener(hideMobileMenu);
    };
  });

  return (
    <>
      <nav
        className={`
          flex-none bg-tn-d-black flex
        `}
      >
        <div
          className={`
            flex-1 md:hidden
            flex items-center
          `}
        >
          <button
            onClick={toggleMobileMenu}
            className={`
              cursor-pointer
              text-tn-d-fg text-left text-xl
              hover:bg-gray-800
              px-1 py-0.5 mx-4
              rounded
            `}
          >
            <FontAwesomeIcon icon={faBars} />
          </button>
        </div>
        <ul
          className={`
            flex-1 hidden md:flex
            font-bold text-tn-d-fg text-2xl
            gap-8 px-4 items-center
          `}
        >
          <NavLink to="/" >Home</NavLink>
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

      <div
        className={`
          bg-tn-d-black/70
          float-start
          md:hidden
          ${mobileMenuHidden ? "hidden" : ""}
        `}
      >
        <ul
          className={`
            flex flex-col items-center gap-y-2 p-2
            text-xl
          `}
        >
          <NavLink to="/" toggle={toggleMobileMenu}>Home</NavLink>
          <NavLink to="/profile" toggle={toggleMobileMenu}>Profile</NavLink>
        </ul>
      </div>
    </>
  );
}

interface NavLinkProps {
  to: string,
  toggle?: () => void,
  children: React.ReactNode,
}

function NavLink({ to, toggle, children }: NavLinkProps) {
  let path = useResolvedPath(to);
  let isActive = useMatch({ path: path.pathname, end: true });

  return (
    <li>
      <Link
        to={to}
        onClick={toggle}
        className={`
          cursor-pointer
          hover:text-tn-d-blue
          ${isActive ? "text-tn-d-dblue" : "text-tn-d-fg"}
        `}
      >
        {children}
      </Link>
    </li>
  );
}

export default Navbar;
