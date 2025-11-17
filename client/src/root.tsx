import { Outlet } from "react-router"

import Navbar from "./navbar.tsx"

function Root() {
  return (
    <div className="flex flex-col h-screen">
      <Navbar />
      <Outlet />
    </div>
  );
}

export default Root;
