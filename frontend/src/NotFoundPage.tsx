import { Link } from 'react-router-dom'

function NotFoundPage() {
  return (
    <div
      className={`
        flex flex-col items-center justify-center h-screen
      `}
    >
      <h1
        className={`
          text-tn-d-fg text-center text-4xl mb-4
        `}
      >
        Page Not Found
      </h1>
      <Link to={"/"}>
        <button
          className={`
            bg-tn-d-black text-tn-d-fg
            px-4 py-2 mt-4
            rounded-md
            cursor-pointer
            border border-transparent
            hover:border hover:border-tn-d-blue
            transition-border duration-250
          `}
        >
          Go back Home
        </button>
      </Link>
    </div>
  );
}

export default NotFoundPage;
