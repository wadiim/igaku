import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { isTokenExpired } from './utils/auth'

function Register() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  let navigate = useNavigate();

  useEffect(() => {
    let jwt = localStorage.getItem("jwt");
    if (!isTokenExpired(jwt)) {
      navigate("/");
    }
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    fetch('http://localhost:4000/auth/register/', {
      method: 'POST',
      body: JSON.stringify({
        username: username,
        email: email,
        password: password,
      }),
    }).then(res => {
      if (res.status === 409) {
        throw new Error("Username or Email already taken");
      } else if (res.status !== 200) {
        throw new Error("Something went wrong");
      }
      return res.text()
    }).then(data => {
      localStorage.setItem("jwt", data);
      navigate("/");
    }).catch(err => {
      setErrorMessage(err.message);
    });
  }

  return (
    <div className={`mt-10 mb-10 sm:mx-auto sm:w-full sm:max-w-sm`}>
      <form method="POST" className={`space-y-6`} onSubmit={handleSubmit}>
        <div>
          <label
            htmlFor="username"
            className={`block text-sm/6 font-medium text-tn-d-fg`}
          >
            Username
          </label>
          <div className={`mt-2`}>
            <input
              name="username"
              value={username}
              onChange={e => {
                setUsername(e.target.value);
                setErrorMessage("");
              }}
              className={`
                block w-full rounded-md
                px-3 py-1.5
                text-base
                text-tn-d-fg
                bg-tn-d-fg/4
                outline-1 -outline-offset-1 outline-tn-d-fg/32
                focus:outline-2 focus:-outline-offset-2 focus:outline-tn-d-blue
                sm:text-sm/6
              `}
            />
          </div>
        </div>

        <div>
          <label
            htmlFor="email"
            className={`block text-sm/6 font-medium text-tn-d-fg`}
          >
            Email
          </label>
          <div className={`mt-2`}>
            <input
              name="email"
              type="email"
              value={email}
              onChange={e => {
                setEmail(e.target.value);
                setErrorMessage("");
              }}
              className={`
                block w-full rounded-md
                px-3 py-1.5
                text-base
                text-tn-d-fg
                bg-tn-d-fg/4
                outline-1 -outline-offset-1 outline-tn-d-fg/32
                focus:outline-2 focus:-outline-offset-2 focus:outline-tn-d-blue
                sm:text-sm/6
              `}
            />
          </div>
        </div>

        <div>
          <label
            htmlFor="password"
            className={`block text-sm/6 font-medium text-tn-d-fg`}
          >
            Password
          </label>
          <div className={`mt-2`}>
            <input
              name="password"
              type="password"
              value={password}
              onChange={e => {
                setPassword(e.target.value);
                setErrorMessage("");
              }}
              className={`
                block w-full rounded-md
                px-3 py-1.5
                text-base
                text-tn-d-fg
                bg-tn-d-fg/4
                outline-1 -outline-offset-1 outline-tn-d-fg/32
                focus:outline-2 focus:-outline-offset-2 focus:outline-tn-d-blue
                sm:text-sm/6
              `}
            />
          </div>
        </div>

        <div>
          {
            errorMessage &&
              <div className={`text-tn-d-red mb-4`}>{errorMessage}</div>
          }
          <button
            type="submit"
            className={`
              flex w-full justify-center rounded-md
              px-3 py-1.5 mt-8
              bg-tn-d-dblue
              hover:bg-tn-d-blue
              text-sm/6 font-semibold
              text-tn-d-fg
              focus-visible:outline-2 focus-visible:outline-offset-2
              focus-visible:outline-tn-d-blue
            `}
          >
            Sign up
          </button>
        </div>
      </form>

      <p className="mt-10 text-center text-sm/6 text-tn-d-fg">
        Already have an account?{" "}
        <Link
          to="/auth/login"
          className={`font-semibold text-tn-d-dblue hover:text-tn-d-blue`}
        >
          Sign in
        </Link>
      </p>
    </div>
  );
}

export default Register;
