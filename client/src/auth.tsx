import { Outlet } from "react-router"

function Auth() {
  return (
    <div
      className={`
        flex min-h-screen flex-col justify-center px-6 py-12 lg:px-8
      `}
    >
      <div className={`sm:mx-auto sm:w-full sm:max-w-sm`}>
        <h2
          className={`
            text-tn-d-fg text-center text-4xl/8 font-bold
            tracking-tight
          `}
        >
          Hello~
        </h2>
      </div>
      <Outlet />
    </div>
  );
}

export default Auth;
